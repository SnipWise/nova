package gatewayserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// handlePassthroughWithTools handles requests where the client sends tools in the request.
// In passthrough mode, the gateway forwards tool definitions to the LLM backend,
// and if the LLM decides to call tools, the tool_calls are returned to the client
// in standard OpenAI format. The client then executes the tools and sends back
// the results as messages with role "tool".
//
// This is the standard flow used by clients like qwen-code, aider, continue.dev:
//
//	Client → Gateway: messages + tools
//	Gateway → LLM: messages + tools (with crew context: RAG, orchestration, compression)
//	LLM → Gateway: response with tool_calls (finish_reason: "tool_calls")
//	Gateway → Client: forward tool_calls in OpenAI format
//	Client executes tools locally
//	Client → Gateway: messages + tool results (role: "tool")
//	... cycle continues until LLM returns content with finish_reason: "stop"
func (agent *GatewayServerAgent) handlePassthroughWithTools(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest) {
	completionID := generateCompletionID()
	modelName := agent.resolveModelName(req.Model)

	// Create a direct OpenAI client using the current chat agent's config
	agentConfig := agent.currentChatAgent.GetConfig()
	client := openai.NewClient(
		option.WithBaseURL(agentConfig.EngineURL),
		option.WithAPIKey(agentConfig.APIKey),
	)

	// Build OpenAI-compatible messages from the request
	openaiMessages := agent.convertToOpenAIMessages(req.Messages)

	// Build OpenAI-compatible tools from the request
	openaiTools := agent.convertToOpenAITools(req.Tools)

	// Build completion params
	params := openai.ChatCompletionNewParams{
		Model:    agent.currentChatAgent.GetModelID(),
		Messages: openaiMessages,
		Tools:    openaiTools,
	}

	// Apply optional parameters from request
	if req.Temperature != nil {
		params.Temperature = openai.Opt(*req.Temperature)
	}
	if req.TopP != nil {
		params.TopP = openai.Opt(*req.TopP)
	}
	if req.MaxTokens != nil {
		params.MaxTokens = openai.Opt(*req.MaxTokens)
	}

	if req.Stream {
		agent.handlePassthroughStreaming(w, r, client, params, completionID, modelName)
	} else {
		agent.handlePassthroughNonStreaming(w, client, params, completionID, modelName)
	}
}

// handlePassthroughNonStreaming makes a single completion call and returns the full response.
func (agent *GatewayServerAgent) handlePassthroughNonStreaming(
	w http.ResponseWriter,
	client openai.Client,
	params openai.ChatCompletionNewParams,
	completionID string,
	modelName string,
) {
	completion, err := client.Chat.Completions.New(agent.ctx, params)
	if err != nil {
		agent.writeAPIError(w, http.StatusInternalServerError, "server_error", fmt.Sprintf("LLM request failed: %v", err))
		return
	}

	if len(completion.Choices) == 0 {
		agent.writeAPIError(w, http.StatusInternalServerError, "server_error", "No choices returned from LLM")
		return
	}

	choice := completion.Choices[0]
	finishReason := string(choice.FinishReason)

	// Build response message
	responseMsg := ChatCompletionMessage{
		Role: "assistant",
	}

	// If the LLM returned tool_calls, forward them
	if finishReason == "tool_calls" && len(choice.Message.ToolCalls) > 0 {
		responseMsg.ToolCalls = agent.convertFromOpenAIToolCalls(choice.Message.ToolCalls)
	}

	// If the LLM returned content
	if choice.Message.Content != "" {
		responseMsg.Content = NewMessageContent(choice.Message.Content)
	}

	response := ChatCompletionResponse{
		ID:      completionID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: []ChatCompletionChoice{
			{
				Index:        0,
				Message:      responseMsg,
				FinishReason: &finishReason,
			},
		},
		Usage: &Usage{
			PromptTokens:     int(completion.Usage.PromptTokens),
			CompletionTokens: int(completion.Usage.CompletionTokens),
			TotalTokens:      int(completion.Usage.TotalTokens),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		agent.log.Error("Failed to encode passthrough response: %v", err)
	}
}

// handlePassthroughStreaming streams the completion response in OpenAI SSE format.
func (agent *GatewayServerAgent) handlePassthroughStreaming(
	w http.ResponseWriter,
	r *http.Request,
	client openai.Client,
	params openai.ChatCompletionNewParams,
	completionID string,
	modelName string,
) {
	flusher, err := agent.setupSSEHeaders(w)
	if err != nil {
		agent.writeAPIError(w, http.StatusInternalServerError, "server_error", "Streaming not supported")
		return
	}

	stream := client.Chat.Completions.NewStreaming(agent.ctx, params)

	// Track accumulated tool calls across chunks
	var accumulatedToolCalls []ToolCall
	sentRole := false

	for stream.Next() {
		chunk := stream.Current()

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]

		// Build delta
		delta := ChatCompletionDelta{}

		// Send role in the first chunk
		if !sentRole {
			delta.Role = "assistant"
			sentRole = true
		}

		// Content delta
		if choice.Delta.Content != "" {
			delta.Content = NewMessageContent(choice.Delta.Content)
		}

		// Tool calls delta
		if len(choice.Delta.ToolCalls) > 0 {
			toolCalls := agent.convertFromOpenAIStreamToolCalls(choice.Delta.ToolCalls)
			delta.ToolCalls = toolCalls

			// Accumulate tool calls for finish_reason detection
			for _, tc := range toolCalls {
				if tc.ID != "" {
					accumulatedToolCalls = append(accumulatedToolCalls, tc)
				}
			}
		}

		// Determine finish_reason
		var finishReason *string
		if choice.FinishReason != "" {
			fr := string(choice.FinishReason)
			finishReason = &fr
		}

		// Write the chunk
		sseChunk := ChatCompletionChunk{
			ID:      completionID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   modelName,
			Choices: []ChatCompletionChunkChoice{
				{
					Index:        0,
					Delta:        delta,
					FinishReason: finishReason,
				},
			},
		}

		jsonData, marshalErr := json.Marshal(sseChunk)
		if marshalErr != nil {
			agent.log.Error("Failed to marshal streaming chunk: %v", marshalErr)
			continue
		}

		if _, writeErr := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); writeErr != nil {
			agent.log.Error("Failed to write streaming chunk: %v", writeErr)
			break
		}
		flusher.Flush()

		// Check for client disconnect
		select {
		case <-r.Context().Done():
			agent.log.Info("Client disconnected during passthrough streaming")
			return
		default:
		}
	}

	if err := stream.Err(); err != nil {
		agent.log.Error("Passthrough stream error: %v", err)
	}

	// Send [DONE] marker
	agent.writeStreamDone(w, flusher)
}

// --- Conversion helpers ---

// convertToOpenAIMessages converts gateway ChatCompletionMessage to OpenAI SDK messages.
func (agent *GatewayServerAgent) convertToOpenAIMessages(msgs []ChatCompletionMessage) []openai.ChatCompletionMessageParamUnion {
	var result []openai.ChatCompletionMessageParamUnion

	for _, msg := range msgs {
		content := ""
		if msg.Content != nil {
			content = msg.Content.String()
		}

		switch msg.Role {
		case "system":
			result = append(result, openai.SystemMessage(content))
		case "user":
			result = append(result, openai.UserMessage(content))
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls
				toolCallParams := make([]openai.ChatCompletionMessageToolCallUnionParam, len(msg.ToolCalls))
				for i, tc := range msg.ToolCalls {
					toolCallParams[i] = openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID: tc.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						},
					}
				}
				result = append(result, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						Content:   openai.ChatCompletionAssistantMessageParamContentUnion{OfString: openai.Opt(content)},
						ToolCalls: toolCallParams,
					},
				})
			} else {
				result = append(result, openai.AssistantMessage(content))
			}
		case "tool":
			result = append(result, openai.ToolMessage(content, msg.ToolCallID))
		case "developer":
			result = append(result, openai.DeveloperMessage(content))
		}
	}

	return result
}

// convertToOpenAITools converts gateway ToolDefinition to OpenAI SDK tool params.
func (agent *GatewayServerAgent) convertToOpenAITools(toolDefs []ToolDefinition) []openai.ChatCompletionToolUnionParam {
	if len(toolDefs) == 0 {
		return nil
	}

	var result []openai.ChatCompletionToolUnionParam
	for _, td := range toolDefs {
		// Convert parameters to shared.FunctionParameters (map[string]any)
		var params shared.FunctionParameters
		if td.Function.Parameters != nil {
			// Marshal then unmarshal to convert any type to map[string]any
			paramsJSON, _ := json.Marshal(td.Function.Parameters)
			_ = json.Unmarshal(paramsJSON, &params)
		}

		result = append(result, openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        td.Function.Name,
			Description: openai.String(td.Function.Description),
			Parameters:  params,
		}))
	}

	return result
}

// convertFromOpenAIToolCalls converts OpenAI SDK tool calls to gateway ToolCall format.
func (agent *GatewayServerAgent) convertFromOpenAIToolCalls(openaiCalls []openai.ChatCompletionMessageToolCallUnion) []ToolCall {
	result := make([]ToolCall, len(openaiCalls))
	for i, tc := range openaiCalls {
		idx := i
		result[i] = ToolCall{
			Index: &idx,
			ID:    tc.ID,
			Type:  "function",
			Function: FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return result
}

// convertFromOpenAIStreamToolCalls converts streaming tool call deltas.
func (agent *GatewayServerAgent) convertFromOpenAIStreamToolCalls(deltas []openai.ChatCompletionChunkChoiceDeltaToolCall) []ToolCall {
	result := make([]ToolCall, len(deltas))
	for i, d := range deltas {
		idx := int(d.Index)
		result[i] = ToolCall{
			Index: &idx,
			ID:    d.ID,
			Type:  "function",
			Function: FunctionCall{
				Name:      d.Function.Name,
				Arguments: d.Function.Arguments,
			},
		}
	}
	return result
}

// injectCrewContextMessages prepends RAG and system context to the OpenAI messages
// for passthrough mode. This allows the crew's knowledge to be available
// even when the client manages the full message history.
func (agent *GatewayServerAgent) injectCrewContextMessages(
	openaiMessages []openai.ChatCompletionMessageParamUnion,
	userQuestion string,
) []openai.ChatCompletionMessageParamUnion {
	var injected []openai.ChatCompletionMessageParamUnion

	// Add RAG context as system messages at the beginning
	if agent.ragAgent != nil && userQuestion != "" {
		similarities, err := agent.ragAgent.SearchTopN(userQuestion, agent.similarityLimit, agent.maxSimilarities)
		if err == nil && len(similarities) > 0 {
			relevantContext := ""
			for _, sim := range similarities {
				relevantContext += sim.Prompt + "\n---\n"
			}
			injected = append(injected, openai.SystemMessage(
				"Relevant information to help you answer the question:\n"+relevantContext,
			))
		}
	}

	// Append the original messages
	injected = append(injected, openaiMessages...)

	return injected
}

// syncMessagesFromHistory adds messages from the request to the chat agent's internal history.
// This keeps the crew agent aware of the conversation for orchestration and compression purposes.
func (agent *GatewayServerAgent) syncMessagesFromHistory(msgs []ChatCompletionMessage) {
	for _, msg := range msgs {
		content := ""
		if msg.Content != nil {
			content = msg.Content.String()
		}

		switch msg.Role {
		case "system":
			agent.currentChatAgent.AddMessage(roles.System, content)
		case "user":
			agent.currentChatAgent.AddMessage(roles.User, content)
		case "assistant":
			agent.currentChatAgent.AddMessage(roles.Assistant, content)
		case "tool":
			agent.currentChatAgent.AddMessage(roles.Tool, content)
		}
	}
}

// isPassthroughToolRequest returns true if the request contains tools and we're in passthrough mode.
func (agent *GatewayServerAgent) isPassthroughToolRequest(req ChatCompletionRequest) bool {
	return agent.toolMode == ToolModePassthrough && len(req.Tools) > 0
}

// isToolResultMessage checks if the messages contain tool results (role: "tool").
// This indicates the client has executed tools and is sending back results.
func (agent *GatewayServerAgent) isToolResultMessage(msgs []ChatCompletionMessage) bool {
	for _, msg := range msgs {
		if msg.Role == "tool" {
			return true
		}
	}
	return false
}

// extractUserQuestion extracts the last user question for RAG/orchestration context.
func (agent *GatewayServerAgent) extractUserQuestion(msgs []ChatCompletionMessage) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" && msgs[i].Content != nil {
			return msgs[i].Content.String()
		}
	}
	return ""
}

// buildToolCallHistory creates a message history for tool detection in auto-execute mode.
func (agent *GatewayServerAgent) buildToolCallHistory(question string) []messages.Message {
	history := agent.currentChatAgent.GetMessages()
	return append(history, messages.Message{
		Role:    roles.User,
		Content: question,
	})
}

// handlePassthroughFirst attempts to process the request with the passthrough agent first.
// This ensures that tool-capable requests are always handled by the passthrough agent.
//
// Returns:
//   - true if the passthrough agent handled the request and sent a response (tool_calls detected)
//   - false if the passthrough agent determined no tools are needed (should redirect to selected agent)
//
// Flow:
//  1. Save current agent
//  2. Switch to passthrough agent
//  3. PHASE 1: Fast non-streaming call to detect tool_calls
//  4. If tool_calls detected:
//     - If streaming requested: Make streaming call and forward to client
//     - If non-streaming: Send the detection response directly
//  5. Otherwise: restore original agent, return false (caller should use selected agent)
func (agent *GatewayServerAgent) handlePassthroughFirst(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest) bool {
	// Save current agent to restore later if needed
	originalAgentId := agent.selectedAgentId
	originalChatAgent := agent.currentChatAgent

	// Switch to passthrough agent
	passthroughAgent, exists := agent.chatAgents["passthrough"]
	if !exists {
		// This should never happen because we validate in NewAgent()
		agent.log.Error("Passthrough agent not found in crew (should have been caught at startup)")
		agent.writeAPIError(w, http.StatusInternalServerError, "server_error", "Passthrough agent not configured")
		return true // Return true to stop further processing
	}

	agent.log.Info("🔀 Routing to passthrough agent first to check for tool calls")
	agent.selectedAgentId = "passthrough"
	agent.currentChatAgent = passthroughAgent

	// Create OpenAI client for passthrough agent
	agentConfig := agent.currentChatAgent.GetConfig()
	client := openai.NewClient(
		option.WithBaseURL(agentConfig.EngineURL),
		option.WithAPIKey(agentConfig.APIKey),
	)

	// Build OpenAI-compatible messages and tools
	openaiMessages := agent.convertToOpenAIMessages(req.Messages)
	openaiTools := agent.convertToOpenAITools(req.Tools)

	// Build completion params for detection (always non-streaming first)
	detectionParams := openai.ChatCompletionNewParams{
		Model:    agent.currentChatAgent.GetModelID(),
		Messages: openaiMessages,
		Tools:    openaiTools,
	}

	// Apply optional parameters from request
	if req.Temperature != nil {
		detectionParams.Temperature = openai.Opt(*req.Temperature)
	}
	if req.TopP != nil {
		detectionParams.TopP = openai.Opt(*req.TopP)
	}
	if req.MaxTokens != nil {
		detectionParams.MaxTokens = openai.Opt(*req.MaxTokens)
	}

	// PHASE 1: Make fast non-streaming call to detect tool_calls
	agent.log.Info("🔍 Phase 1: Fast detection call (non-streaming)")
	completion, err := client.Chat.Completions.New(agent.ctx, detectionParams)
	if err != nil {
		agent.log.Error("Passthrough agent detection request failed: %v", err)
		// Restore original agent and let caller handle with fallback
		agent.selectedAgentId = originalAgentId
		agent.currentChatAgent = originalChatAgent
		return false
	}

	if len(completion.Choices) == 0 {
		agent.log.Error("Passthrough agent returned no choices")
		// Restore original agent
		agent.selectedAgentId = originalAgentId
		agent.currentChatAgent = originalChatAgent
		return false
	}

	choice := completion.Choices[0]
	finishReason := string(choice.FinishReason)

	// Check if tool_calls were detected
	if finishReason == "tool_calls" && len(choice.Message.ToolCalls) > 0 {
		agent.log.Info("✅ Passthrough agent detected tool_calls")

		// PHASE 2: Handle response based on streaming mode
		if req.Stream {
			// Client wants streaming, make a new streaming call
			agent.log.Info("🔄 Phase 2: Making streaming call (client requested stream)")
			completionID := generateCompletionID()
			modelName := agent.resolveModelName(req.Model)

			flusher, err := agent.setupSSEHeaders(w)
			if err != nil {
				agent.writeAPIError(w, http.StatusInternalServerError, "server_error", "Streaming not supported")
				return true
			}

			// Make streaming call with same params
			stream := client.Chat.Completions.NewStreaming(agent.ctx, detectionParams)

			// Send role chunk first
			agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
				Role: "assistant",
			}, nil)

			// Stream tool_calls
			var accumulatedToolCalls []ToolCall
			for stream.Next() {
				chunk := stream.Current()
				if len(chunk.Choices) == 0 {
					continue
				}

				choice := chunk.Choices[0]
				delta := ChatCompletionDelta{}

				if choice.Delta.Content != "" {
					delta.Content = NewMessageContent(choice.Delta.Content)
				}

				if len(choice.Delta.ToolCalls) > 0 {
					toolCalls := agent.convertFromOpenAIStreamToolCalls(choice.Delta.ToolCalls)
					delta.ToolCalls = toolCalls
					for _, tc := range toolCalls {
						if tc.ID != "" {
							accumulatedToolCalls = append(accumulatedToolCalls, tc)
						}
					}
				}

				var finishReasonPtr *string
				if choice.FinishReason != "" {
					fr := string(choice.FinishReason)
					finishReasonPtr = &fr
				}

				agent.writeStreamChunk(w, flusher, completionID, modelName, &delta, finishReasonPtr)

				// Check for client disconnect
				select {
				case <-r.Context().Done():
					agent.log.Info("Client disconnected during passthrough streaming")
					return true
				default:
				}
			}

			if err := stream.Err(); err != nil {
				agent.log.Error("Passthrough stream error: %v", err)
			}

			agent.writeStreamDone(w, flusher)

		} else {
			// Non-streaming: use the detection response directly
			agent.log.Info("📤 Phase 2: Sending non-streaming response (using detection result)")
			completionID := generateCompletionID()
			modelName := agent.resolveModelName(req.Model)

			responseMsg := ChatCompletionMessage{
				Role:      "assistant",
				ToolCalls: agent.convertFromOpenAIToolCalls(choice.Message.ToolCalls),
			}

			if choice.Message.Content != "" {
				responseMsg.Content = NewMessageContent(choice.Message.Content)
			}

			response := ChatCompletionResponse{
				ID:      completionID,
				Object:  "chat.completion",
				Created: time.Now().Unix(),
				Model:   modelName,
				Choices: []ChatCompletionChoice{
					{
						Index:        0,
						Message:      responseMsg,
						FinishReason: &finishReason,
					},
				},
				Usage: &Usage{
					PromptTokens:     int(completion.Usage.PromptTokens),
					CompletionTokens: int(completion.Usage.CompletionTokens),
					TotalTokens:      int(completion.Usage.TotalTokens),
				},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				agent.log.Error("Failed to encode passthrough first response: %v", err)
			}
		}

		// Keep passthrough agent as current (for next call with tool results)
		return true
	}

	// No tool_calls detected, restore original agent and let caller redirect
	agent.log.Info("⏭️  Passthrough agent found no tool_calls (finish_reason: %s), redirecting to selected agent", finishReason)
	agent.selectedAgentId = originalAgentId
	agent.currentChatAgent = originalChatAgent

	return false
}
