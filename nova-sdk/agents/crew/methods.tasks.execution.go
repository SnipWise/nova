package crew

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// executePlanCLI executes a tasks plan via CLI (StreamCompletion).
// For each task in the plan:
//   - "tool" tasks are executed via the tools agent
//   - "completion"/"developer" tasks are sent to the current chat agent
//
// Results from each task are passed as context to the next task.
// Returns true if a plan was executed (caller should skip normal flow).
func (agent *CrewAgent) executePlanCLI(
	question string,
	callback chat.StreamCallback,
) (bool, error) {
	if agent.tasksAgent == nil {
		return false, nil
	}

	agent.log.Info("📋 Tasks agent configured, identifying plan...")

	plan, err := agent.tasksAgent.IdentifyPlanFromText(question)
	if err != nil {
		agent.log.Error("Error identifying plan: %v", err)
		return false, err
	}

	if plan == nil || len(plan.Tasks) == 0 {
		agent.log.Info("📋 No tasks identified, falling back to normal completion")
		return false, nil
	}

	agent.log.Info("📋 Plan identified with %d tasks", len(plan.Tasks))

	planSummary := formatPlanSummary(plan)
	callback(planSummary, "tasks_plan_identified")

	var accumulatedResults []string

	for _, task := range plan.Tasks {
		agent.log.Info("▶️  Executing task %s: %s (responsible: %s)", task.ID, task.Description, task.Responsible)

		callback(fmt.Sprintf("\n---\n**Task %s**: %s\n", task.ID, task.Description), "task_started")

		var result string
		var taskErr error

		switch task.Responsible {
		case "tool":
			result, taskErr = agent.executeToolTask(task, accumulatedResults, callback)
		case "completion", "developer":
			result, taskErr = agent.executeCompletionTask(task, accumulatedResults, question, callback)
		default:
			agent.log.Info("⚠️  Unknown responsible type: %s, treating as completion", task.Responsible)
			result, taskErr = agent.executeCompletionTask(task, accumulatedResults, question, callback)
		}

		if taskErr != nil {
			agent.log.Error("Error executing task %s: %v", task.ID, taskErr)
			callback(fmt.Sprintf("\n**Error on task %s**: %s\n", task.ID, taskErr.Error()), "task_error")
			accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): ERROR - %s", task.ID, task.Description, taskErr.Error()))
			continue
		}

		accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): %s", task.ID, task.Description, result))
		agent.log.Info("✅ Task %s completed", task.ID)
	}

	callback("\n---\n**All tasks completed.**\n", "tasks_completed")

	// Preserve conversation history: add the original question and a summary
	// of all task results so the chat agent remembers this exchange.
	agent.currentChatAgent.AddMessage(roles.User, question)
	agent.currentChatAgent.AddMessage(roles.Assistant, buildTaskContext(accumulatedResults))

	return true, nil
}

// executeToolTask executes a "tool" task via the tools agent.
func (agent *CrewAgent) executeToolTask(
	task agents.Task,
	previousResults []string,
	callback chat.StreamCallback,
) (string, error) {
	if agent.toolsAgent == nil {
		return "", fmt.Errorf("tools agent not configured, cannot execute tool task: %s", task.ToolName)
	}

	agent.toolsAgent.ResetMessages()

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

	toolCallsResult, err := agent.toolsAgent.DetectToolCallsLoopWithConfirmation(
		toolMessages,
		agent.executeFn,
		agent.confirmationPromptFn,
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

		agent.toolsAgent.ResetLastStateToolCalls()
		agent.toolsAgent.ResetMessages()

		return result, nil
	}

	agent.toolsAgent.ResetLastStateToolCalls()
	agent.toolsAgent.ResetMessages()

	return "No tool execution result", nil
}

// executeCompletionTask executes a "completion" or "developer" task via the current chat agent.
func (agent *CrewAgent) executeCompletionTask(
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
