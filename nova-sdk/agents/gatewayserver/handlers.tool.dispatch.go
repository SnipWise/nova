package gatewayserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
)

// dispatchToolCalls executes tools via toolsAgent using the configured parallel/sequential mode
// and available gateway callbacks. Returns a wrapped result and any execution error.
func (agent *GatewayServerAgent) dispatchToolCalls(historyMessages []messages.Message) (*toolCallResultWrapper, error) {
	modelConfig := agent.toolsAgent.GetModelConfig()
	isParallel := modelConfig.ParallelToolCalls != nil && *modelConfig.ParallelToolCalls

	var result *tools.ToolCallResult
	var err error

	if isParallel {
		result, err = agent.dispatchParallel(historyMessages)
	} else {
		result, err = agent.dispatchSequential(historyMessages)
	}

	return &toolCallResultWrapper{result: result}, err
}

// dispatchParallel executes tools in parallel mode, selecting the appropriate variant
// based on which callbacks are configured.
func (agent *GatewayServerAgent) dispatchParallel(historyMessages []messages.Message) (*tools.ToolCallResult, error) {
	switch {
	case agent.confirmationFn != nil && agent.executeFn != nil:
		return agent.toolsAgent.DetectParallelToolCallsWithConfirmation(historyMessages, agent.executeFn, agent.confirmationFn)
	case agent.executeFn != nil:
		return agent.toolsAgent.DetectParallelToolCalls(historyMessages, agent.executeFn)
	case agent.confirmationFn != nil:
		return agent.toolsAgent.DetectParallelToolCallsWithConfirmation(historyMessages, agent.confirmationFn)
	default:
		return agent.toolsAgent.DetectParallelToolCalls(historyMessages)
	}
}

// dispatchSequential executes tools in sequential (loop) mode, selecting the appropriate
// variant based on which callbacks are configured.
func (agent *GatewayServerAgent) dispatchSequential(historyMessages []messages.Message) (*tools.ToolCallResult, error) {
	switch {
	case agent.confirmationFn != nil && agent.executeFn != nil:
		return agent.toolsAgent.DetectToolCallsLoopWithConfirmation(historyMessages, agent.executeFn, agent.confirmationFn)
	case agent.executeFn != nil:
		return agent.toolsAgent.DetectToolCallsLoop(historyMessages, agent.executeFn)
	case agent.confirmationFn != nil:
		return agent.toolsAgent.DetectToolCallsLoopWithConfirmation(historyMessages, agent.confirmationFn)
	default:
		return agent.toolsAgent.DetectToolCallsLoop(historyMessages)
	}
}

// sendCompletionOrToolResponse generates a final chat completion or returns the tool result
// directly, depending on whether the agent requires further completion after tool execution.
func (agent *GatewayServerAgent) sendCompletionOrToolResponse(
	w http.ResponseWriter,
	r *http.Request,
	req ChatCompletionRequest,
	toolCallsResult *toolCallResultWrapper,
) {
	if agent.shouldGenerateCompletion() {
		if req.Stream {
			agent.handleStreamingCompletion(w, r, req)
		} else {
			agent.handleNonStreamingCompletion(w, r, req)
		}
		return
	}

	// Tools handled everything; return the tool result as the response
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
	w.Header().Set(handlerContentType, handlerMIMEJSON)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		agent.log.Error("Failed to encode tool result response: %v", err)
	}
}
