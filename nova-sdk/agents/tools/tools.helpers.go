package tools

import (
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared/constant"
)

// createToolCallParams converts detected tool calls to the proper parameter format
func createToolCallParams(detectedToolCalls []openai.ChatCompletionMessageToolCallUnion) []openai.ChatCompletionMessageToolCallUnionParam {
	toolCallParams := make([]openai.ChatCompletionMessageToolCallUnionParam, len(detectedToolCalls))
	for i, toolCall := range detectedToolCalls {
		toolCallParams[i] = openai.ChatCompletionMessageToolCallUnionParam{
			OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
				ID:   toolCall.ID,
				Type: constant.Function("function"),
				Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments,
				},
			},
		}
	}
	return toolCallParams
}

// createAssistantMessageWithToolCalls creates an assistant message with tool calls
func createAssistantMessageWithToolCalls(toolCallParams []openai.ChatCompletionMessageToolCallUnionParam) openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			ToolCalls: toolCallParams,
		},
	}
}

// ToolExecutionResult holds the result of a tool execution
type ToolExecutionResult struct {
	Content    string
	ShouldStop bool
	// Possible values: "function_executed", "user_denied", "user_quit", "error", "exit_loop"
	ExecFinishReason string
}

// executeToolCall executes a single tool call without confirmation
func (agent *BaseAgent) executeToolCall(
	functionName string,
	functionArgs string,
	callID string,
	toolCallBack func(string, string) (string, error),
) (ToolExecutionResult, error) {
	agent.Log.Info(fmt.Sprintf("▶️ Executing function: %s with args: %s\n", functionName, functionArgs))

	resultContent, errExec := toolCallBack(functionName, functionArgs)

	if errExec != nil {
		agent.Log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))

		toolExecRes := ToolExecutionResult{
			Content:          fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error()),
			ShouldStop:       true,
			ExecFinishReason: "exit_loop",
		}
		// Store the last state of tool calls with confirmation
		agent.lastState = LastToolCallsState{
			Confirmation: Confirmed,
			ExecutionResult: ToolExecutionResult{
				Content:          toolExecRes.Content,
				ShouldStop:       toolExecRes.ShouldStop,
				ExecFinishReason: toolExecRes.ExecFinishReason,
			},
		}

		return toolExecRes, nil
	}

	if resultContent == "" {
		resultContent = `{"error": "Function execution returned empty result"}`
	}

	agent.Log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

	toolExecRes := ToolExecutionResult{
		Content:          resultContent,
		ShouldStop:       false,
		ExecFinishReason: "function_executed",
	}
	// Store the last state of tool calls with confirmation
	agent.lastState = LastToolCallsState{
		Confirmation: Confirmed,
		ExecutionResult: ToolExecutionResult{
			Content:          toolExecRes.Content,
			ShouldStop:       toolExecRes.ShouldStop,
			ExecFinishReason: toolExecRes.ExecFinishReason,
		},
	}
	return toolExecRes, nil
}

// executeToolCallWithConfirmation executes a single tool call with confirmation
func (agent *BaseAgent) executeToolCallWithConfirmation(
	functionName string,
	functionArgs string,
	callID string,
	toolCallBack func(string, string) (string, error),
	confirmationCallBack func(string, string) ConfirmationResponse,
) (ToolExecutionResult, error) {
	// Ask for confirmation before executing the tool
	agent.Log.Info(fmt.Sprintf("⁉️ Requesting confirmation for function: %s with args: %s\n", functionName, functionArgs))
	confirmation := confirmationCallBack(functionName, functionArgs)

	switch confirmation {
	case Confirmed:
		// Proceed with tool execution
		toolExecRes, err := agent.executeToolCall(functionName, functionArgs, callID, toolCallBack)

		// Store the last state of tool calls with confirmation
		agent.lastState = LastToolCallsState{
			Confirmation: Confirmed,
			ExecutionResult: ToolExecutionResult{
				Content:          toolExecRes.Content,
				ShouldStop:       toolExecRes.ShouldStop,
				ExecFinishReason: toolExecRes.ExecFinishReason,
			},
		}

		agent.Log.Info(fmt.Sprintf("✅ Tool execution confirmed for function: %s\n", functionName))
		return toolExecRes, err

	case Denied:
		// Skip execution but add a message indicating the tool was denied (cancel in the vscode extension)
		toolExecRes := ToolExecutionResult{
			Content:          `{"status": "denied", "message": "Tool execution was denied by user"}`,
			ShouldStop:       false,
			ExecFinishReason: "user_denied",
		}

		// Store the last state of tool calls with confirmation
		agent.lastState = LastToolCallsState{
			Confirmation: Confirmed,
			ExecutionResult: ToolExecutionResult{
				Content:          toolExecRes.Content,
				ShouldStop:       toolExecRes.ShouldStop,
				ExecFinishReason: toolExecRes.ExecFinishReason,
			},
		}

		agent.Log.Warn(fmt.Sprintf("⛔ Tool execution denied for function: %s\n", functionName))
		return toolExecRes, nil

	case Quit:
		// Exit the function immediately (reset in the vscode extension)
		toolExecRes := ToolExecutionResult{
			Content:          `{"status": "quit", "message": "Tool execution was quit by user"}`,
			ShouldStop:       true,
			ExecFinishReason: "user_quit",
		}

		// Store the last state of tool calls with confirmation
		agent.lastState = LastToolCallsState{
			Confirmation: Confirmed,
			ExecutionResult: ToolExecutionResult{
				Content:          toolExecRes.Content,
				ShouldStop:       toolExecRes.ShouldStop,
				ExecFinishReason: toolExecRes.ExecFinishReason,
			},
		}

		agent.Log.Warn(fmt.Sprintf("🛑 Quit requested for function: %s\n", functionName))
		return toolExecRes, nil
	}

	return ToolExecutionResult{}, nil
}

// processToolCalls processes all detected tool calls and updates the message history
func (agent *BaseAgent) processToolCalls(
	messages []openai.ChatCompletionMessageParamUnion,
	detectedToolCalls []openai.ChatCompletionMessageToolCallUnion,
	results *[]string,
	toolCallBack func(string, string) (string, error),
	confirmationCallBack func(string, string) ConfirmationResponse,
) ([]openai.ChatCompletionMessageParamUnion, bool, string) {
	agent.Log.Info("🚀 Processing tool calls...")

	// Create tool call params and add assistant message
	toolCallParams := createToolCallParams(detectedToolCalls)
	assistantMessage := createAssistantMessageWithToolCalls(toolCallParams)
	messages = append(messages, assistantMessage)

	// Process each detected tool call
	for _, toolCall := range detectedToolCalls {
		functionName := toolCall.Function.Name
		functionArgs := toolCall.Function.Arguments
		callID := toolCall.ID

		var result ToolExecutionResult
		var err error

		// Execute with or without confirmation
		if confirmationCallBack != nil {
			result, err = agent.executeToolCallWithConfirmation(functionName, functionArgs, callID, toolCallBack, confirmationCallBack)
		} else {
			result, err = agent.executeToolCall(functionName, functionArgs, callID, toolCallBack)
		}

		if err != nil {
			return messages, true, "error"
		}

		// Handle quit case
		if result.ShouldStop && result.ExecFinishReason == "user_quit" {
			return messages, true, result.ExecFinishReason
		}

		// Add result to results list if there's content
		if result.Content != "" {
			*results = append(*results, result.Content)

			// Add the tool call result to the conversation history
			messages = append(messages, openai.ToolMessage(result.Content, toolCall.ID))
		}

		// Handle error case
		if result.ShouldStop {
			return messages, true, result.ExecFinishReason
		}
	}

	return messages, false, ""
}

// handleStopReason processes the 'stop' finish reason
func (agent *BaseAgent) handleStopReason(
	messages []openai.ChatCompletionMessageParamUnion,
	content string,
) ([]openai.ChatCompletionMessageParamUnion, string) {
	agent.Log.Info("✋ Stopping due to 'stop' finish reason.")

	agent.Log.Info(fmt.Sprintf("🤖 [from the tool agent] %s\n", content))

	// Add final assistant message to conversation history
	messages = append(messages, openai.AssistantMessage(content))
	return messages, content
}

// collectStreamResponse runs a streaming completion, accumulates the response
// text via streamCallback, and returns the full response string.
// It returns an error if the callback fails or if stream.Err/Close fail.
func (agent *BaseAgent) collectStreamResponse(
	paramsForCall openai.ChatCompletionNewParams,
	streamCallback func(content string) error,
) (string, error) {
	stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, paramsForCall)
	var response string
	var cbkRes error

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			agent.SaveLastChunkResponse(&chunk)
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			cbkRes = streamCallback(chunk.Choices[0].Delta.Content)
			response += chunk.Choices[0].Delta.Content
		}
		if cbkRes != nil {
			agent.Log.Error("Error in stream callback: %v", cbkRes)
			break
		}
	}

	if cbkRes != nil {
		return "", cbkRes
	}
	if err := stream.Err(); err != nil {
		agent.Log.Error("Stream error: %v", err)
		return "", err
	}
	if err := stream.Close(); err != nil {
		agent.Log.Error("Stream close error: %v", err)
		return "", err
	}
	return response, nil
}

// saveHistoryIfNeeded persists workingMessages to the agent's conversation
// history when KeepConversationHistory is enabled.
func (agent *BaseAgent) saveHistoryIfNeeded(workingMessages []openai.ChatCompletionMessageParamUnion) {
	if agent.Config.KeepConversationHistory {
		agent.ChatCompletionParams.Messages = workingMessages
	}
}
