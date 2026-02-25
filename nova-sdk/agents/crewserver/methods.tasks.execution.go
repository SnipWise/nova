package crewserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// executePlanHTTP executes a tasks plan via HTTP (handleCompletion).
// For each task in the plan:
//   - "tool" tasks are executed via the tools agent
//   - "completion"/"developer" tasks are sent to the current chat agent
//
// Results from each task are passed as context to the next task.
// Returns true if a plan was executed (caller should skip normal flow).
func (agent *CrewServerAgent) executePlanHTTP(
	question string,
	w http.ResponseWriter,
	flusher http.Flusher,
) (bool, error) {
	if agent.tasksAgentConfig == nil {
		return false, nil
	}

	agent.Log.Info("📋 Tasks agent configured, identifying plan...")

	plan, err := agent.tasksAgentConfig.IdentifyPlanFromText(question)
	if err != nil {
		agent.Log.Error("Error identifying plan: %v", err)
		return false, err
	}

	if plan == nil || len(plan.Tasks) == 0 {
		agent.Log.Info("📋 No tasks identified, falling back to normal completion")
		return false, nil
	}

	agent.Log.Info("📋 Plan identified with %d tasks", len(plan.Tasks))

	planSummary := formatPlanSummary(plan)
	agent.writeSSEChunk(w, flusher, planSummary)

	var accumulatedResults []string

	for _, task := range plan.Tasks {
		agent.Log.Info("▶️  Executing task %s: %s (responsible: %s)", task.ID, task.Description, task.Responsible)

		agent.writeSSEChunk(w, flusher, fmt.Sprintf("\n---\n**Task %s**: %s\n", task.ID, task.Description))

		sseCallback := func(chunk string, finishReason string) error {
			if chunk != "" {
				agent.writeSSEChunk(w, flusher, chunk)
			}
			return nil
		}

		var result string
		var taskErr error

		switch task.Responsible {
		case "tool":
			result, taskErr = agent.executeToolTask(task, accumulatedResults, sseCallback)
		case "completion", "developer":
			result, taskErr = agent.executeCompletionTask(task, accumulatedResults, question, sseCallback)
		default:
			result, taskErr = agent.executeCompletionTask(task, accumulatedResults, question, sseCallback)
		}

		if taskErr != nil {
			agent.Log.Error("Error executing task %s: %v", task.ID, taskErr)
			agent.writeSSEChunk(w, flusher, fmt.Sprintf("\n**Error on task %s**: %s\n", task.ID, taskErr.Error()))
			accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): ERROR - %s", task.ID, task.Description, taskErr.Error()))
			continue
		}

		accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): %s", task.ID, task.Description, result))
		agent.Log.Info("✅ Task %s completed", task.ID)
	}

	agent.writeSSEChunk(w, flusher, "\n---\n**All tasks completed.**\n")
	agent.writeSSEFinish(w, flusher)

	// Preserve conversation history: add the original question and a summary
	// of all task results so the chat agent remembers this exchange.
	agent.currentChatAgent.AddMessage(roles.User, question)
	agent.currentChatAgent.AddMessage(roles.Assistant, buildTaskContext(accumulatedResults))

	return true, nil
}

// executeToolTask executes a "tool" task via the tools agent.
func (agent *CrewServerAgent) executeToolTask(
	task agents.Task,
	previousResults []string,
	callback chat.StreamCallback,
) (string, error) {
	if agent.ToolsAgent == nil {
		return "", fmt.Errorf("tools agent not configured, cannot execute tool task: %s", task.ToolName)
	}

	agent.ToolsAgent.ResetMessages()

	contextMessage := buildTaskContext(previousResults)

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

	// Use confirmation function if available, otherwise use web confirmation
	confirmFn := agent.ConfirmationPromptFn
	if confirmFn == nil {
		confirmFn = agent.webConfirmationPrompt
	}

	toolCallsResult, err := agent.ToolsAgent.DetectToolCallsLoopWithConfirmation(
		toolMessages,
		agent.ExecuteFn,
		confirmFn,
	)
	if err != nil {
		return "", err
	}

	if len(toolCallsResult.Results) > 0 {
		result := toolCallsResult.LastAssistantMessage
		if result == "" {
			result = strings.Join(toolCallsResult.Results, "\n")
		}

		agent.currentChatAgent.AddMessage(roles.System, fmt.Sprintf("Tool execution result for task '%s': %s", task.Description, result))
		callback(fmt.Sprintf("\nTool result: %s\n", result), "tool_task_completed")

		agent.ToolsAgent.ResetLastStateToolCalls()
		agent.ToolsAgent.ResetMessages()

		return result, nil
	}

	agent.ToolsAgent.ResetLastStateToolCalls()
	agent.ToolsAgent.ResetMessages()

	return "No tool execution result", nil
}

// executeCompletionTask executes a "completion" or "developer" task via the current chat agent.
func (agent *CrewServerAgent) executeCompletionTask(
	task agents.Task,
	previousResults []string,
	originalQuestion string,
	callback chat.StreamCallback,
) (string, error) {
	contextMessage := buildTaskContext(previousResults)

	prompt := fmt.Sprintf("Task: %s", task.Description)
	if contextMessage != "" {
		prompt = fmt.Sprintf("Context from previous tasks:\n%s\n\nOriginal request: %s\n\nCurrent task: %s",
			contextMessage, originalQuestion, task.Description)
	}

	if contextMessage != "" {
		agent.currentChatAgent.AddMessage(roles.System, "Context from previous tasks:\n"+contextMessage)
	}

	var fullResponse string
	_, err := agent.currentChatAgent.GenerateStreamCompletion(
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

// buildTaskContext creates a context string from accumulated task results
func buildTaskContext(results []string) string {
	if len(results) == 0 {
		return ""
	}
	return strings.Join(results, "\n---\n")
}

// formatPlanSummary creates a human-readable summary of the plan
func formatPlanSummary(plan *agents.Plan) string {
	var sb strings.Builder
	sb.WriteString("**Plan identified:**\n")
	for _, task := range plan.Tasks {
		dependsOn := ""
		if len(task.DependsOn) > 0 {
			dependsOn = fmt.Sprintf(" (depends on: %s)", strings.Join(task.DependsOn, ", "))
		}
		fmt.Fprintf(&sb, "- **%s.** [%s] %s%s\n", task.ID, task.Responsible, task.Description, dependsOn)
	}
	return sb.String()
}
