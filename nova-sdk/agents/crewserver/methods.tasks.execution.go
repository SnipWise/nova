package crewserver

import (
	"fmt"
	"net/http"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
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

	planSummary := serverbase.FormatPlanSummary(plan)
	agent.WriteSSEChunk(w, flusher, planSummary)

	var accumulatedResults []string

	for _, task := range plan.Tasks {
		agent.Log.Info("▶️  Executing task %s: %s (responsible: %s)", task.ID, task.Description, task.Responsible)

		agent.WriteSSEChunk(w, flusher, fmt.Sprintf("\n---\n**Task %s**: %s\n", task.ID, task.Description))

		sseCallback := func(chunk string, finishReason string) error {
			if chunk != "" {
				agent.WriteSSEChunk(w, flusher, chunk)
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
			agent.WriteSSEChunk(w, flusher, fmt.Sprintf("\n**Error on task %s**: %s\n", task.ID, taskErr.Error()))
			accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): ERROR - %s", task.ID, task.Description, taskErr.Error()))
			continue
		}

		accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): %s", task.ID, task.Description, result))
		agent.Log.Info("✅ Task %s completed", task.ID)
	}

	agent.WriteSSEChunk(w, flusher, "\n---\n**All tasks completed.**\n")
	agent.WriteSSEFinish(w, flusher)

	// Preserve conversation history: add the original question and a summary
	// of all task results so the chat agent remembers this exchange.
	agent.currentChatAgent.AddMessage(roles.User, question)
	agent.currentChatAgent.AddMessage(roles.Assistant, serverbase.BuildTaskContext(accumulatedResults))

	return true, nil
}

// executeToolTask executes a "tool" task via the tools agent.
func (agent *CrewServerAgent) executeToolTask(
	task agents.Task,
	previousResults []string,
	callback chat.StreamCallback,
) (string, error) {
	confirmFn := agent.ConfirmationPromptFn
	if confirmFn == nil {
		confirmFn = agent.WebConfirmationPrompt
	}
	return serverbase.ExecuteToolTask(agent.currentChatAgent, agent.ToolsAgent, task, previousResults, agent.ExecuteFn, confirmFn, callback)
}

// executeCompletionTask executes a "completion" or "developer" task via the current chat agent.
func (agent *CrewServerAgent) executeCompletionTask(
	task agents.Task,
	previousResults []string,
	originalQuestion string,
	callback chat.StreamCallback,
) (string, error) {
	return serverbase.ExecuteCompletionTask(agent.currentChatAgent, task, previousResults, originalQuestion, callback)
}
