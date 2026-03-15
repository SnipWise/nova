package gatewayserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// executePlanOpenAI executes a tasks plan and returns results in OpenAI format.
// For each task in the plan:
//   - "tool" tasks are executed via the tools agent
//   - "completion"/"developer" tasks are sent to the current chat agent
//
// Results from each task are passed as context to the next task.
// Returns true if a plan was executed (caller should skip normal flow).
func (agent *GatewayServerAgent) executePlanOpenAI(
	w http.ResponseWriter,
	r *http.Request,
	req ChatCompletionRequest,
) bool {
	if agent.tasksAgent == nil {
		return false
	}

	lastUserMessage := agent.extractLastUserMessage(req.Messages)
	if lastUserMessage == "" {
		return false
	}

	agent.log.Info("📋 Tasks agent configured, identifying plan...")

	plan, err := agent.tasksAgent.IdentifyPlanFromText(lastUserMessage)
	if err != nil {
		agent.log.Error("Error identifying plan: %v", err)
		return false
	}

	if plan == nil || len(plan.Tasks) == 0 {
		agent.log.Info("📋 No tasks identified, falling back to normal completion")
		return false
	}

	agent.log.Info("📋 Plan identified with %d tasks", len(plan.Tasks))

	if req.Stream {
		agent.executePlanStreaming(w, r, req, plan, lastUserMessage)
	} else {
		agent.executePlanNonStreaming(w, req, plan, lastUserMessage)
	}

	return true
}

// executePlanStreaming executes a plan and streams results in OpenAI SSE format.
func (agent *GatewayServerAgent) executePlanStreaming(
	w http.ResponseWriter,
	r *http.Request,
	req ChatCompletionRequest,
	plan *agents.Plan,
	originalQuestion string,
) {
	completionID := generateCompletionID()
	modelName := agent.resolveModelName(req.Model)

	flusher, err := agent.setupSSEHeaders(w)
	if err != nil {
		agent.writeAPIError(w, http.StatusInternalServerError, "server_error", "Streaming not supported")
		return
	}

	// Send initial chunk with role
	agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
		Role: "assistant",
	}, nil)

	// Stream plan summary
	planSummary := serverbase.FormatPlanSummary(plan)
	agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
		Content: NewMessageContent(planSummary),
	}, nil)

	// Execute each task
	var accumulatedResults []string

	for _, task := range plan.Tasks {
		agent.log.Info("▶️  Executing task %s: %s (responsible: %s)", task.ID, task.Description, task.Responsible)

		// Stream task header
		agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
			Content: NewMessageContent(fmt.Sprintf("\n---\n**Task %s**: %s\n", task.ID, task.Description)),
		}, nil)

		streamCallback := func(chunk string, finishReason string) error {
			if chunk != "" {
				agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
					Content: NewMessageContent(chunk),
				}, nil)
			}
			return nil
		}

		var result string
		var taskErr error

		switch task.Responsible {
		case "tool":
			result, taskErr = agent.executeToolTask(task, accumulatedResults, streamCallback)
		case "completion", "developer":
			result, taskErr = agent.executeCompletionTask(task, accumulatedResults, originalQuestion, streamCallback)
		default:
			result, taskErr = agent.executeCompletionTask(task, accumulatedResults, originalQuestion, streamCallback)
		}

		if taskErr != nil {
			agent.log.Error("Error executing task %s: %v", task.ID, taskErr)
			agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
				Content: NewMessageContent(fmt.Sprintf("\n**Error on task %s**: %s\n", task.ID, taskErr.Error())),
			}, nil)
			accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): ERROR - %s", task.ID, task.Description, taskErr.Error()))
			continue
		}

		accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): %s", task.ID, task.Description, result))
		agent.log.Info("✅ Task %s completed", task.ID)
	}

	// Stream completion
	agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{
		Content: NewMessageContent("\n---\n**All tasks completed.**\n"),
	}, nil)

	// Preserve conversation history: add the original question and a summary
	// of all task results so the chat agent remembers this exchange.
	agent.currentChatAgent.AddMessage(roles.User, originalQuestion)
	agent.currentChatAgent.AddMessage(roles.Assistant, serverbase.BuildTaskContext(accumulatedResults))

	fr := "stop"
	agent.writeStreamChunk(w, flusher, completionID, modelName, &ChatCompletionDelta{}, &fr)
	agent.writeStreamDone(w, flusher)
}

// executePlanNonStreaming executes a plan and returns a single JSON response.
func (agent *GatewayServerAgent) executePlanNonStreaming(
	w http.ResponseWriter,
	req ChatCompletionRequest,
	plan *agents.Plan,
	originalQuestion string,
) {
	completionID := generateCompletionID()

	var fullResponse strings.Builder
	fullResponse.WriteString(serverbase.FormatPlanSummary(plan))

	var accumulatedResults []string

	for _, task := range plan.Tasks {
		agent.log.Info("▶️  Executing task %s: %s (responsible: %s)", task.ID, task.Description, task.Responsible)

		fmt.Fprintf(&fullResponse, "\n---\n**Task %s**: %s\n", task.ID, task.Description)

		collectCallback := func(chunk string, finishReason string) error {
			fullResponse.WriteString(chunk)
			return nil
		}

		var result string
		var taskErr error

		switch task.Responsible {
		case "tool":
			result, taskErr = agent.executeToolTask(task, accumulatedResults, collectCallback)
		case "completion", "developer":
			result, taskErr = agent.executeCompletionTask(task, accumulatedResults, originalQuestion, collectCallback)
		default:
			result, taskErr = agent.executeCompletionTask(task, accumulatedResults, originalQuestion, collectCallback)
		}

		if taskErr != nil {
			agent.log.Error("Error executing task %s: %v", task.ID, taskErr)
			fmt.Fprintf(&fullResponse, "\n**Error on task %s**: %s\n", task.ID, taskErr.Error())
			accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): ERROR - %s", task.ID, task.Description, taskErr.Error()))
			continue
		}

		accumulatedResults = append(accumulatedResults, fmt.Sprintf("Task %s (%s): %s", task.ID, task.Description, result))
		agent.log.Info("✅ Task %s completed", task.ID)
	}

	fullResponse.WriteString("\n---\n**All tasks completed.**\n")

	finishReason := "stop"
	content := fullResponse.String()
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
		Usage: agent.estimateUsage(req.Messages, content),
	}

	// Preserve conversation history: add the original question and a summary
	// of all task results so the chat agent remembers this exchange.
	agent.currentChatAgent.AddMessage(roles.User, originalQuestion)
	agent.currentChatAgent.AddMessage(roles.Assistant, serverbase.BuildTaskContext(accumulatedResults))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		agent.log.Error("Failed to encode plan response: %v", err)
	}
}

// executeToolTask executes a "tool" task via the tools agent.
func (agent *GatewayServerAgent) executeToolTask(
	task agents.Task,
	previousResults []string,
	callback chat.StreamCallback,
) (string, error) {
	return serverbase.ExecuteToolTask(agent.currentChatAgent, agent.toolsAgent, task, previousResults, agent.executeFn, agent.confirmationFn, callback)
}

// executeCompletionTask executes a "completion" or "developer" task via the current chat agent.
func (agent *GatewayServerAgent) executeCompletionTask(
	task agents.Task,
	previousResults []string,
	originalQuestion string,
	callback chat.StreamCallback,
) (string, error) {
	return serverbase.ExecuteCompletionTask(agent.currentChatAgent, task, previousResults, originalQuestion, callback)
}
