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

// toolExecutionResult holds the result of a tool execution
type toolExecutionResult struct {
	Content      string
	ShouldStop   bool
	FinishReason string
}

// executeToolCall executes a single tool call without confirmation
func (agent *BaseAgent) executeToolCall(
	functionName string,
	functionArgs string,
	callID string,
	toolCallBack func(string, string) (string, error),
) (toolExecutionResult, error) {
	agent.Log.Info(fmt.Sprintf("▶️ Executing function: %s with args: %s\n", functionName, functionArgs))

	resultContent, errExec := toolCallBack(functionName, functionArgs)

	if errExec != nil {
		agent.Log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))
		return toolExecutionResult{
			Content:      fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error()),
			ShouldStop:   true,
			FinishReason: "exit_loop",
		}, nil
	}

	if resultContent == "" {
		resultContent = `{"error": "Function execution returned empty result"}`
	}

	agent.Log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

	return toolExecutionResult{
		Content:      resultContent,
		ShouldStop:   false,
		FinishReason: "",
	}, nil
}

// executeToolCallWithConfirmation executes a single tool call with confirmation
func (agent *BaseAgent) executeToolCallWithConfirmation(
	functionName string,
	functionArgs string,
	callID string,
	toolCallBack func(string, string) (string, error),
	confirmationCallBack func(string, string) ConfirmationResponse,
) (toolExecutionResult, error) {
	// Ask for confirmation before executing the tool
	agent.Log.Info(fmt.Sprintf("⁉️ Requesting confirmation for function: %s with args: %s\n", functionName, functionArgs))
	confirmation := confirmationCallBack(functionName, functionArgs)

	switch confirmation {
	case Confirmed:
		return agent.executeToolCall(functionName, functionArgs, callID, toolCallBack)

	case Denied:
		// Skip execution but add a message indicating the tool was denied
		agent.Log.Warn(fmt.Sprintf("⛔ Tool execution denied for function: %s\n", functionName))
		return toolExecutionResult{
			Content:      `{"status": "denied", "message": "Tool execution was denied by user"}`,
			ShouldStop:   false,
			FinishReason: "",
		}, nil

	case Quit:
		// Exit the function immediately
		agent.Log.Warn(fmt.Sprintf("🛑 Quit requested for function: %s\n", functionName))
		return toolExecutionResult{
			Content:      "",
			ShouldStop:   true,
			FinishReason: "user_quit",
		}, nil
	}

	return toolExecutionResult{}, nil
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

		var result toolExecutionResult
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
		if result.ShouldStop && result.FinishReason == "user_quit" {
			return messages, true, result.FinishReason
		}

		// Add result to results list if there's content
		if result.Content != "" {
			*results = append(*results, result.Content)

			// Add the tool call result to the conversation history
			messages = append(messages, openai.ToolMessage(result.Content, toolCall.ID))
		}

		// Handle error case
		if result.ShouldStop {
			return messages, true, result.FinishReason
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
	agent.Log.Info(fmt.Sprintf("🤖 %s\n", content))

	// Add final assistant message to conversation history
	messages = append(messages, openai.AssistantMessage(content))
	return messages, content
}
