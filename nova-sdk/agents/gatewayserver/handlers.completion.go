package gatewayserver

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// HandleChatCompletionsForTest exposes handleChatCompletions for testing.
func (agent *GatewayServerAgent) HandleChatCompletionsForTest(w http.ResponseWriter, r *http.Request) {
	agent.handleChatCompletions(w, r)
}

// handleChatCompletions is the main handler for POST /v1/chat/completions.
// It dispatches requests through a configurable chain of agent handlers based on agentExecutionOrder.
// Each handler can either:
//  - Handle the request and return (stopping the chain)
//  - Skip and let the next handler process the request
func (agent *GatewayServerAgent) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		agent.writeAPIError(w, http.StatusBadRequest, "invalid_request_error", fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if len(req.Messages) == 0 {
		agent.writeAPIError(w, http.StatusBadRequest, "invalid_request_error", "messages array is required and must not be empty")
		return
	}

	// Call before completion hook
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Compress context if needed
	agent.compressContextIfNeeded()

	// Add RAG context if available
	lastUserMessage := agent.extractLastUserMessage(req.Messages)
	agent.addRAGContext(lastUserMessage)

	// Sync incoming messages to the current chat agent
	agent.syncMessages(req.Messages)

	// Execute tasks plan if tasksAgent is configured (before the execution chain)
	if agent.executePlanOpenAI(w, r, req) {
		if agent.afterCompletion != nil {
			agent.afterCompletion(agent)
		}
		return
	}

	// Process request through the agent execution order chain
	// Each handler returns true if it handled the request, false otherwise
	for _, executionType := range agent.agentExecutionOrder {
		var handled bool

		switch executionType {
		case AgentExecutionClientSideTools:
			// Client-side tool execution: detect tool calls and return them to client
			handled = agent.handleClientSideToolDetection(w, r, req)

		case AgentExecutionServerSideTools:
			// Server-side tool execution: execute tools and continue completion
			handled = agent.handleServerSideToolExecution(w, r, req)

		case AgentExecutionOrchestrator:
			// Orchestrator: detect topic and route to appropriate agent
			handled = agent.handleOrchestration(lastUserMessage)
		}

		// If this handler processed the request, we're done
		if handled {
			if agent.afterCompletion != nil {
				agent.afterCompletion(agent)
			}
			return
		}
	}

	// No specialized handler processed the request, use default completion
	if req.Stream {
		agent.handleStreamingCompletion(w, r, req)
	} else {
		agent.handleNonStreamingCompletion(w, r, req)
	}

	// Call after completion hook
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}
}

// handleNonStreamingCompletion generates a complete JSON response.
func (agent *GatewayServerAgent) handleNonStreamingCompletion(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest) {
	lastUserMessage := agent.extractLastUserMessage(req.Messages)
	completionID := generateCompletionID()

	var fullResponse string

	result, err := agent.currentChatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: lastUserMessage},
		},
		func(chunk string, finishReason string) error {
			fullResponse += chunk
			return nil
		},
	)

	if err != nil {
		agent.writeAPIError(w, http.StatusInternalServerError, "server_error", fmt.Sprintf("Completion failed: %v", err))
		return
	}

	finishReason := "stop"
	if result != nil && result.FinishReason != "" {
		finishReason = result.FinishReason
	}

	content := fullResponse
	response := ChatCompletionResponse{
		ID:      completionID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   agent.resolveModelName(req.Model),
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChatCompletionMessage{
					Role:    "assistant",
					Content: NewMessageContent(content),
				},
				FinishReason: &finishReason,
			},
		},
		Usage: agent.estimateUsage(req.Messages, fullResponse),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		agent.log.Error("Failed to encode completion response: %v", err)
	}
}

// handleStreamingCompletion generates an SSE streaming response in OpenAI format.
func (agent *GatewayServerAgent) handleStreamingCompletion(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest) {
	lastUserMessage := agent.extractLastUserMessage(req.Messages)
	completionID := generateCompletionID()
	modelName := agent.resolveModelName(req.Model)

	// Setup SSE headers
	flusher, err := agent.setupSSEHeaders(w)
	if err != nil {
		agent.writeAPIError(w, http.StatusInternalServerError, "server_error", "Streaming not supported")
		return
	}

	// Send initial chunk with role
	agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
		Role: "assistant",
	}, nil)

	// Stream content chunks
	stopped := false
	_, errCompletion := agent.currentChatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: lastUserMessage},
		},
		func(chunk string, finishReason string) error {
			// Check for stop signal
			select {
			case <-agent.stopStreamChan:
				stopped = true
				return errors.New("stream stopped by user")
			default:
			}

			// Check if client disconnected
			select {
			case <-r.Context().Done():
				stopped = true
				return errors.New("client disconnected")
			default:
			}

			if chunk != "" {
				agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
					Content: NewMessageContent(chunk),
				}, nil)
			}

			if finishReason == "stop" && !stopped {
				fr := "stop"
				agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{}, &fr)
			}

			return nil
		},
	)

	if errCompletion != nil && !stopped {
		agent.log.Error("Streaming completion error: %v", errCompletion)
	}

	// Send [DONE] marker
	agent.writeStreamDone(w, flusher)
}

// handleServerSideToolExecution processes requests with server-side tool execution.
// Returns false to continue the execution chain (context enrichment mode).
//
// This handler executes server-side tools and adds results to the context,
// then lets subsequent handlers (ClientSideTools, Orchestrator) continue the flow.
func (agent *GatewayServerAgent) handleServerSideToolExecution(
	w http.ResponseWriter,
	r *http.Request,
	req ChatCompletionRequest,
) bool {
	// Check if server-side tools agent is configured
	if agent.toolsAgent == nil {
		return false
	}

	// Check if server has tools configured
	serverTools := agent.toolsAgent.GetTools()
	if len(serverTools) == 0 {
		// No server-side tools configured, let client-side handle it
		return false
	}

	// Execute server-side tools and enrich context
	agent.log.Info("🔧 Executing server-side tools (context enrichment mode)")
	toolsExecuted, err := agent.executeToolsAndEnrichContext(req)
	if err != nil {
		agent.log.Error("Server-side tool execution failed: %v", err)
		return false
	}

	if toolsExecuted {
		// Tools were executed successfully, generate final completion
		agent.log.Info("✅ Server-side tools executed, generating final response")
		if req.Stream {
			agent.handleStreamingCompletion(w, r, req)
		} else {
			agent.handleNonStreamingCompletion(w, r, req)
		}
		return true // Stop the chain, we've handled the request
	}

	// No tools executed, continue the execution chain
	// (ClientSideTools and Orchestrator will process next)
	return false
}

// executeToolsAndEnrichContext executes tools via toolsAgent and adds results to currentChatAgent context
// Returns (toolsExecuted bool, error)
func (agent *GatewayServerAgent) executeToolsAndEnrichContext(req ChatCompletionRequest) (bool, error) {
	lastUserMessage := agent.extractLastUserMessage(req.Messages)

	// Build history for tool detection
	historyMessages := agent.currentChatAgent.GetMessages()
	historyMessages = append(historyMessages, messages.Message{
		Role:    roles.User,
		Content: lastUserMessage,
	})

	agent.toolsAgent.ResetMessages()

	var toolCallsResult *toolCallResultWrapper
	var err error

	modelConfig := agent.toolsAgent.GetModelConfig()
	isParallel := modelConfig.ParallelToolCalls != nil && *modelConfig.ParallelToolCalls

	// Execute tools based on parallel/sequential mode
	// Note: We don't pass nil callbacks - let toolsAgent use its own configured callbacks
	if isParallel {
		if agent.confirmationFn != nil && agent.executeFn != nil {
			result, e := agent.toolsAgent.DetectParallelToolCallsWithConfirmation(historyMessages, agent.executeFn, agent.confirmationFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else if agent.executeFn != nil {
			result, e := agent.toolsAgent.DetectParallelToolCalls(historyMessages, agent.executeFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else if agent.confirmationFn != nil {
			// Only confirmation callback, pass it
			result, e := agent.toolsAgent.DetectParallelToolCallsWithConfirmation(historyMessages, agent.confirmationFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else {
			// No callbacks from gateway, use toolsAgent's configured callbacks
			result, e := agent.toolsAgent.DetectParallelToolCalls(historyMessages)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		}
	} else {
		if agent.confirmationFn != nil && agent.executeFn != nil {
			result, e := agent.toolsAgent.DetectToolCallsLoopWithConfirmation(historyMessages, agent.executeFn, agent.confirmationFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else if agent.executeFn != nil {
			result, e := agent.toolsAgent.DetectToolCallsLoop(historyMessages, agent.executeFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else if agent.confirmationFn != nil {
			// Only confirmation callback, pass it
			result, e := agent.toolsAgent.DetectToolCallsLoopWithConfirmation(historyMessages, agent.confirmationFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else {
			// No callbacks from gateway, use toolsAgent's configured callbacks
			result, e := agent.toolsAgent.DetectToolCallsLoop(historyMessages)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		}
	}

	if err != nil {
		return false, fmt.Errorf("tool execution failed: %w", err)
	}

	// Add tool results to context if tools executed successfully
	if toolCallsResult != nil && toolCallsResult.result != nil {
		state := agent.toolsAgent.GetLastStateToolCalls()
		finishReason := state.ExecutionResult.ExecFinishReason

		if len(toolCallsResult.result.Results) > 0 && finishReason == "function_executed" {
			// Add tool results to currentChatAgent context
			agent.log.Info("✅ Adding tool results to context: %s", toolCallsResult.result.LastAssistantMessage)
			agent.currentChatAgent.AddMessage(roles.System, toolCallsResult.result.LastAssistantMessage)
			// Return true to indicate tools were executed
			return true, nil
		} else {
			agent.log.Info("ℹ️  No tools executed (finish_reason: %s)", finishReason)
		}
	}

	// No tools were executed
	return false, nil
}

// handleOrchestration processes requests through the orchestrator.
// Returns true if orchestration was performed, false otherwise.
func (agent *GatewayServerAgent) handleOrchestration(lastUserMessage string) bool {
	// Check if orchestrator is configured
	if agent.orchestratorAgent == nil {
		return false
	}

	// Route to appropriate agent based on topic detection
	agent.routeToAppropriateAgent(lastUserMessage)

	// Orchestration doesn't stop the chain, it just routes to the right agent
	// Return false to continue to the completion handler
	return false
}

// handleAutoExecuteCompletion handles tool calls with server-side execution,
// then generates the final completion.
func (agent *GatewayServerAgent) handleAutoExecuteCompletion(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest) {
	lastUserMessage := agent.extractLastUserMessage(req.Messages)

	// Build history for tool detection
	historyMessages := agent.currentChatAgent.GetMessages()
	historyMessages = append(historyMessages, messages.Message{
		Role:    roles.User,
		Content: lastUserMessage,
	})

	agent.toolsAgent.ResetMessages()

	var toolCallsResult *toolCallResultWrapper
	var err error

	modelConfig := agent.toolsAgent.GetModelConfig()
	isParallel := modelConfig.ParallelToolCalls != nil && *modelConfig.ParallelToolCalls

	// Execute tools based on parallel/sequential mode
	// Only pass non-nil callbacks from gateway - otherwise let toolsAgent use its own configured callbacks
	if isParallel {
		if agent.confirmationFn != nil && agent.executeFn != nil {
			result, e := agent.toolsAgent.DetectParallelToolCallsWithConfirmation(historyMessages, agent.executeFn, agent.confirmationFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else if agent.executeFn != nil {
			result, e := agent.toolsAgent.DetectParallelToolCalls(historyMessages, agent.executeFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else if agent.confirmationFn != nil {
			// Only confirmation callback, pass it
			result, e := agent.toolsAgent.DetectParallelToolCallsWithConfirmation(historyMessages, agent.confirmationFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else {
			// No callbacks from gateway, use toolsAgent's configured callbacks
			result, e := agent.toolsAgent.DetectParallelToolCalls(historyMessages)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		}
	} else {
		if agent.confirmationFn != nil && agent.executeFn != nil {
			result, e := agent.toolsAgent.DetectToolCallsLoopWithConfirmation(historyMessages, agent.executeFn, agent.confirmationFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else if agent.executeFn != nil {
			result, e := agent.toolsAgent.DetectToolCallsLoop(historyMessages, agent.executeFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else if agent.confirmationFn != nil {
			// Only confirmation callback, pass it
			result, e := agent.toolsAgent.DetectToolCallsLoopWithConfirmation(historyMessages, agent.confirmationFn)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		} else {
			// No callbacks from gateway, use toolsAgent's configured callbacks
			result, e := agent.toolsAgent.DetectToolCallsLoop(historyMessages)
			toolCallsResult = &toolCallResultWrapper{result: result}
			err = e
		}
	}

	if err != nil {
		agent.writeAPIError(w, http.StatusInternalServerError, "server_error", fmt.Sprintf("Tool execution failed: %v", err))
		agent.cleanupToolState()
		return
	}

	// Add tool results to context if tools executed successfully
	if toolCallsResult != nil && toolCallsResult.result != nil {
		state := agent.toolsAgent.GetLastStateToolCalls()
		finishReason := state.ExecutionResult.ExecFinishReason
		if len(toolCallsResult.result.Results) > 0 && finishReason == "function_executed" {
			agent.currentChatAgent.AddMessage(roles.System, toolCallsResult.result.LastAssistantMessage)
		}
	}

	// Check if we should generate a completion
	if agent.shouldGenerateCompletion() {
		if req.Stream {
			agent.handleStreamingCompletion(w, r, req)
		} else {
			agent.handleNonStreamingCompletion(w, r, req)
		}
	} else {
		// Tools handled everything; return empty response
		completionID := generateCompletionID()
		content := ""
		if toolCallsResult != nil && toolCallsResult.result != nil && toolCallsResult.result.LastAssistantMessage != "" {
			content = toolCallsResult.result.LastAssistantMessage
		}
		finishReason := "stop"
		response := ChatCompletionResponse{
			ID:      completionID,
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   agent.resolveModelName(req.Model),
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: NewMessageContent(content),
					},
					FinishReason: &finishReason,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			agent.log.Error("Failed to encode tool result response: %v", err)
		}
	}

	agent.cleanupToolState()
}

// --- Helper types ---

type toolCallResultWrapper struct {
	result *tools.ToolCallResult
}

// --- SSE helpers ---

// setupSSEHeaders configures SSE streaming headers and returns the flusher.
func (agent *GatewayServerAgent) setupSSEHeaders(w http.ResponseWriter) (http.Flusher, error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}
	return flusher, nil
}

// writeStreamChunk writes a single SSE chunk in OpenAI format.
func (agent *GatewayServerAgent) writeStreamChunk(
	w http.ResponseWriter,
	flusher http.Flusher,
	id string,
	model string,
	delta *ChatCompletionDelta,
	finishReason *string,
) {
	chunk := ChatCompletionChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []ChatCompletionChunkChoice{
			{
				Index:        0,
				Delta:        *delta,
				FinishReason: finishReason,
			},
		},
	}

	jsonData, err := json.Marshal(chunk)
	if err != nil {
		agent.log.Error("Failed to marshal stream chunk: %v", err)
		return
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		agent.log.Error("Failed to write stream chunk: %v", err)
		return
	}
	flusher.Flush()
}

// writeStreamToolCalls writes tool_calls chunks in OpenAI streaming format.
func (agent *GatewayServerAgent) writeStreamToolCalls(
	w http.ResponseWriter,
	flusher http.Flusher,
	id string,
	model string,
	toolCalls []ToolCall,
) {
	finishReason := "tool_calls"
	chunk := ChatCompletionChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []ChatCompletionChunkChoice{
			{
				Index: 0,
				Delta: ChatCompletionDelta{
					ToolCalls: toolCalls,
				},
				FinishReason: &finishReason,
			},
		},
	}

	jsonData, err := json.Marshal(chunk)
	if err != nil {
		agent.log.Error("Failed to marshal tool calls chunk: %v", err)
		return
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		agent.log.Error("Failed to write tool calls chunk: %v", err)
		return
	}
	flusher.Flush()
}

// writeStreamDone writes the [DONE] marker to end the SSE stream.
func (agent *GatewayServerAgent) writeStreamDone(w http.ResponseWriter, flusher http.Flusher) {
	if _, err := fmt.Fprintf(w, "data: [DONE]\n\n"); err != nil {
		agent.log.Error("Failed to write [DONE] marker: %v", err)
		return
	}
	flusher.Flush()
}

// --- Message helpers ---

// extractLastUserMessage finds the last user message from the request messages.
func (agent *GatewayServerAgent) extractLastUserMessage(msgs []ChatCompletionMessage) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" && msgs[i].Content != nil {
			return msgs[i].Content.String()
		}
	}
	return ""
}

// syncMessages loads the incoming OpenAI messages into the current chat agent's history.
// It resets the agent's messages and replays them, excluding the last user message
// (which will be passed to GenerateStreamCompletion separately).
func (agent *GatewayServerAgent) syncMessages(msgs []ChatCompletionMessage) {
	agent.currentChatAgent.ResetMessages()

	for i, msg := range msgs {
		// Skip the last user message — it's passed separately to the completion
		if i == len(msgs)-1 && msg.Role == "user" {
			continue
		}

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
		case "developer":
			agent.currentChatAgent.AddMessage(roles.Developer, content)
		case "tool":
			// Tool results are added as system messages with context
			agent.currentChatAgent.AddMessage(roles.Tool, content)
		}
	}
}

// resolveModelName returns the model name to use in responses.
// If the request specified a model, use it; otherwise use the agent's model.
func (agent *GatewayServerAgent) resolveModelName(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	return agent.currentChatAgent.GetModelID()
}

// estimateUsage provides a rough token usage estimate based on character counts.
func (agent *GatewayServerAgent) estimateUsage(reqMessages []ChatCompletionMessage, response string) *Usage {
	promptChars := 0
	for _, msg := range reqMessages {
		if msg.Content != nil {
			promptChars += len(msg.Content.String())
		}
	}

	// Rough estimate: ~4 chars per token
	promptTokens := promptChars / 4
	completionTokens := len(response) / 4

	return &Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
}

// shouldGenerateCompletion determines if a completion should be generated after tool execution.
func (agent *GatewayServerAgent) shouldGenerateCompletion() bool {
	if agent.toolsAgent == nil {
		return true
	}

	state := agent.toolsAgent.GetLastStateToolCalls()
	confirmation := state.Confirmation
	finishReason := state.ExecutionResult.ExecFinishReason

	return confirmation == 0 &&
		(finishReason == "user_quit" ||
			finishReason == "user_denied" ||
			finishReason == "")
}

// cleanupToolState resets tool agent state after completion.
func (agent *GatewayServerAgent) cleanupToolState() {
	if agent.toolsAgent != nil {
		agent.toolsAgent.ResetLastStateToolCalls()
		agent.toolsAgent.ResetMessages()
	}
}

// --- Error helpers ---

// writeAPIError writes an OpenAI-compatible error response.
func (agent *GatewayServerAgent) writeAPIError(w http.ResponseWriter, statusCode int, errorType string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	apiErr := APIError{
		Error: APIErrorDetail{
			Message: message,
			Type:    errorType,
		},
	}
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		agent.log.Error("Failed to encode error response: %v", err)
	}
}

// --- ID generation ---

// generateCompletionID generates a unique ID for a completion (chatcmpl-xxx format).
func generateCompletionID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "chatcmpl-" + hex.EncodeToString(b)
}

// generateToolCallID generates a unique ID for a tool call (call_xxx format).
func generateToolCallID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "call_" + hex.EncodeToString(b)
}
