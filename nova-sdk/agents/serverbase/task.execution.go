package serverbase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// ExecuteToolTask executes a "tool" task via the given tools agent.
// Pass executeFn and/or confirmFn as nil to use the agent's stored callbacks.
func ExecuteToolTask(
	chatAgent *chat.Agent,
	toolsAgent *tools.Agent,
	task agents.Task,
	previousResults []string,
	executeFn tools.ToolCallback,
	confirmFn tools.ConfirmationCallback,
	callback chat.StreamCallback,
) (string, error) {
	if toolsAgent == nil {
		return "", fmt.Errorf("tools agent not configured, cannot execute tool task: %s", task.ToolName)
	}

	toolsAgent.ResetMessages()

	contextMessage := BuildTaskContext(previousResults)

	toolMessage := fmt.Sprintf("Execute the following task: %s", task.Description)
	if task.ToolName != "" {
		argsJSON, _ := json.Marshal(task.Arguments)
		toolMessage = fmt.Sprintf("Call the tool '%s' with arguments: %s\nTask description: %s",
			task.ToolName, string(argsJSON), task.Description)
	}

	var toolMessages []messages.Message
	if contextMessage != "" {
		toolMessages = append(toolMessages, messages.Message{
			Role:    roles.System,
			Content: "Context from previous tasks:\n" + contextMessage,
		})
	}
	toolMessages = append(toolMessages, messages.Message{
		Role:    roles.User,
		Content: toolMessage,
	})

	var toolCallsResult *tools.ToolCallResult
	var err error

	if executeFn != nil && confirmFn != nil {
		toolCallsResult, err = toolsAgent.DetectToolCallsLoopWithConfirmation(toolMessages, executeFn, confirmFn)
	} else if executeFn != nil {
		toolCallsResult, err = toolsAgent.DetectToolCallsLoop(toolMessages, executeFn)
	} else {
		toolCallsResult, err = toolsAgent.DetectToolCallsLoop(toolMessages)
	}

	if err != nil {
		return "", err
	}

	if toolCallsResult != nil && len(toolCallsResult.Results) > 0 {
		result := toolCallsResult.LastAssistantMessage
		if result == "" {
			result = strings.Join(toolCallsResult.Results, "\n")
		}

		chatAgent.AddMessage(roles.System, fmt.Sprintf("Tool execution result for task '%s': %s", task.Description, result))
		callback(fmt.Sprintf("\nTool result: %s\n", result), "tool_task_completed")

		toolsAgent.ResetLastStateToolCalls()
		toolsAgent.ResetMessages()

		return result, nil
	}

	toolsAgent.ResetLastStateToolCalls()
	toolsAgent.ResetMessages()

	return "No tool execution result", nil
}

// ExecuteCompletionTask executes a "completion" or "developer" task via the given chat agent.
func ExecuteCompletionTask(
	chatAgent *chat.Agent,
	task agents.Task,
	previousResults []string,
	originalQuestion string,
	callback chat.StreamCallback,
) (string, error) {
	contextMessage := BuildTaskContext(previousResults)

	prompt := fmt.Sprintf("Task: %s", task.Description)
	if contextMessage != "" {
		prompt = fmt.Sprintf("Context from previous tasks:\n%s\n\nOriginal request: %s\n\nCurrent task: %s",
			contextMessage, originalQuestion, task.Description)
	}

	if contextMessage != "" {
		chatAgent.AddMessage(roles.System, "Context from previous tasks:\n"+contextMessage)
	}

	var fullResponse string
	_, err := chatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: prompt},
		},
		func(chunk string, finishReason string) error {
			fullResponse += chunk
			return callback(chunk, finishReason)
		},
	)
	if err != nil {
		return "", err
	}

	return fullResponse, nil
}
