package gatewayserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// handleClientSideToolDetection processes requests with client-side tool execution.
//
// Flow:
//  1. Client sends request with tools definitions
//  2. Gateway uses clientSideToolsAgent's LLM to detect if tools are needed
//  3. If tool_calls detected: return them to client (client executes tools)
//  4. If no tool_calls: return false (let next handler in chain process)
//
// The client is responsible for:
//  - Sending tool definitions in each request
//  - Executing tools locally when tool_calls are returned
//  - Sending tool results back as messages with role "tool"
//
// This handler returns:
//  - true: handled the request (sent response to client)
//  - false: did not handle (next handler should process)
func (agent *GatewayServerAgent) handleClientSideToolDetection(
	w http.ResponseWriter,
	r *http.Request,
	req ChatCompletionRequest,
) bool {
	// Check if client-side tools agent is configured
	if agent.clientSideToolsAgent == nil {
		return false
	}

	// Check if request contains tools
	if len(req.Tools) == 0 {
		return false
	}

	agent.log.Info("🔀 Processing request with client-side tool detection")

	// Create OpenAI client using the client-side tools agent's config
	agentConfig := agent.clientSideToolsAgent.GetConfig()
	client := openai.NewClient(
		option.WithBaseURL(agentConfig.EngineURL),
		option.WithAPIKey(agentConfig.APIKey),
	)

	// Build OpenAI-compatible messages and tools from the request
	openaiMessages := agent.convertToOpenAIMessages(req.Messages)
	openaiTools := agent.convertToOpenAITools(req.Tools)

	// Build completion params
	params := openai.ChatCompletionNewParams{
		Model:    agent.clientSideToolsAgent.GetModelID(),
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

	// Make a detection call (always non-streaming first for detection)
	agent.log.Info("🔍 Detecting tool calls...")
	completion, err := client.Chat.Completions.New(agent.ctx, params)
	if err != nil {
		agent.log.Error("Client-side tool detection failed: %v", err)
		// Don't fail, just let next handler try
		return false
	}

	if len(completion.Choices) == 0 {
		agent.log.Warn("⚠️  Client-side tool detection returned no choices")
		return false
	}

	choice := completion.Choices[0]
	finishReason := string(choice.FinishReason)

	// Check if tool_calls were detected
	if finishReason == "tool_calls" && len(choice.Message.ToolCalls) > 0 {
		agent.log.Info("✅ Client-side tool calls detected, returning to client")

		// Convert tool calls to gateway format
		toolCalls := agent.convertFromOpenAIToolCalls(choice.Message.ToolCalls)

		// Send response based on streaming mode
		if req.Stream {
			agent.sendClientSideToolCallsStreaming(w, r, req, toolCalls, completion)
		} else {
			agent.sendClientSideToolCallsNonStreaming(w, req, toolCalls, completion)
		}

		return true // We handled the request
	}

	// No tool calls detected, let the next handler process
	agent.log.Info("⏭️  No tool calls detected by client-side agent (finish_reason: %s)", finishReason)
	return false
}

// sendClientSideToolCallsNonStreaming sends tool calls in non-streaming format
func (agent *GatewayServerAgent) sendClientSideToolCallsNonStreaming(
	w http.ResponseWriter,
	req ChatCompletionRequest,
	toolCalls []ToolCall,
	completion *openai.ChatCompletion,
) {
	completionID := generateCompletionID()
	modelName := agent.resolveModelName(req.Model)
	finishReason := "tool_calls"

	responseMsg := ChatCompletionMessage{
		Role:      "assistant",
		ToolCalls: toolCalls,
	}

	// Include content if present
	if completion.Choices[0].Message.Content != "" {
		responseMsg.Content = NewMessageContent(completion.Choices[0].Message.Content)
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
		agent.log.Error("Failed to encode client-side tool calls response: %v", err)
	}
}

// sendClientSideToolCallsStreaming sends tool calls in streaming format
func (agent *GatewayServerAgent) sendClientSideToolCallsStreaming(
	w http.ResponseWriter,
	r *http.Request,
	req ChatCompletionRequest,
	toolCalls []ToolCall,
	detectionCompletion *openai.ChatCompletion,
) {
	flusher, err := agent.setupSSEHeaders(w)
	if err != nil {
		agent.writeAPIError(w, http.StatusInternalServerError, "server_error", "Streaming not supported")
		return
	}

	completionID := generateCompletionID()
	modelName := agent.resolveModelName(req.Model)

	// Send role chunk first
	agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
		Role: "assistant",
	}, nil)

	// Send content if present
	if detectionCompletion.Choices[0].Message.Content != "" {
		agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
			Content: NewMessageContent(detectionCompletion.Choices[0].Message.Content),
		}, nil)
	}

	// Send tool calls chunk
	agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
		ToolCalls: toolCalls,
	}, nil)

	// Send finish chunk
	finishReason := "tool_calls"
	agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{}, &finishReason)

	// Check for client disconnect
	select {
	case <-r.Context().Done():
		agent.log.Info("Client disconnected during client-side tool streaming")
		return
	default:
	}

	// Send [DONE] marker
	agent.writeStreamDone(w, flusher)
}
