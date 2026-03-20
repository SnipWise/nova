package crew

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
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

	planSummary := serverbase.FormatPlanSummary(plan)
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
	agent.currentChatAgent.AddMessage(roles.Assistant, serverbase.BuildTaskContext(accumulatedResults))

	return true, nil
}

// executeToolTask executes a "tool" task via the tools agent.
func (agent *CrewAgent) executeToolTask(
	task agents.Task,
	previousResults []string,
	callback chat.StreamCallback,
) (string, error) {
	return serverbase.ExecuteToolTask(agent.currentChatAgent, agent.toolsAgent, task, previousResults, agent.executeFn, agent.confirmationPromptFn, callback)
}

// executeCompletionTask executes a "completion" or "developer" task via the current chat agent.
func (agent *CrewAgent) executeCompletionTask(
	task agents.Task,
	previousResults []string,
	originalQuestion string,
	callback chat.StreamCallback,
) (string, error) {
	return serverbase.ExecuteCompletionTask(agent.currentChatAgent, task, previousResults, originalQuestion, callback)
}
