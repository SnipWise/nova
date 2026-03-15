package serverbase

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// SetupSSEHeaders configures SSE streaming headers and returns a flusher.
func (agent *BaseServerAgent) SetupSSEHeaders(w http.ResponseWriter) (http.Flusher, error) {
	w.Header().Set(headerContentType, contentTypeSSE)
	w.Header().Set(headerCacheControl, cacheControlNoCache)
	w.Header().Set(headerConnection, connectionKeepAlive)
	w.Header().Set(headerAccessControl, accessControlWildcard)
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}
	return flusher, nil
}

// HandleCompletionStop handles the stop stream endpoint.
func (agent *BaseServerAgent) HandleCompletionStop(w http.ResponseWriter, r *http.Request) {
	select {
	case agent.StopStreamChan <- true:
		w.Header().Set(headerContentType, contentTypeJSON)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Stream stopped",
		}); err != nil {
			agent.Log.Error("Failed to encode completion stop response: %v", err)
		}
	default:
		w.Header().Set(headerContentType, contentTypeJSON)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "No stream to stop",
		}); err != nil {
			agent.Log.Error("Failed to encode completion stop response: %v", err)
		}
	}
}

// SetupNotificationChannel creates and registers a notification channel.
func (agent *BaseServerAgent) SetupNotificationChannel() chan ToolCallNotification {
	notificationChan := make(chan ToolCallNotification, 10)
	agent.NotificationChanMutex.Lock()
	agent.CurrentNotificationChan = notificationChan
	agent.NotificationChanMutex.Unlock()
	return notificationChan
}

// StartNotificationStreaming starts a goroutine to stream tool call notifications to the client.
func (agent *BaseServerAgent) StartNotificationStreaming(
	w http.ResponseWriter,
	r *http.Request,
	flusher http.Flusher,
	notificationChan chan ToolCallNotification,
) {
	go func() {
		for {
			select {
			case notification, ok := <-notificationChan:
				if !ok {
					return
				}
				agent.SendToolCallNotification(w, flusher, notification)
			case <-r.Context().Done():
				return
			}
		}
	}()
}

// SendToolCallNotification sends a single tool call notification via SSE.
func (agent *BaseServerAgent) SendToolCallNotification(
	w http.ResponseWriter,
	flusher http.Flusher,
	notification ToolCallNotification,
) {
	notifData := map[string]interface{}{
		"kind":         "tool_call",
		"status":       "pending",
		"operation_id": notification.OperationID,
		"message":      notification.Message,
	}
	jsonData, _ := json.Marshal(notifData)
	if _, err := fmt.Fprintf(w, sseDataFmt, string(jsonData)); err != nil {
		agent.Log.Error("Failed to write notification: %v", err)
		return
	}
	flusher.Flush()
}

// CloseNotificationChannel safely closes the notification channel.
func (agent *BaseServerAgent) CloseNotificationChannel(notificationChan chan ToolCallNotification) {
	close(notificationChan)
	agent.NotificationChanMutex.Lock()
	if agent.CurrentNotificationChan == notificationChan {
		agent.CurrentNotificationChan = nil
	}
	agent.NotificationChanMutex.Unlock()
}

// HandleToolCallsWithNotifications executes tool calls if toolsAgent is configured.
// historyMessages should include chat history with the user question appended.
func (agent *BaseServerAgent) HandleToolCallsWithNotifications(
	chatAgent ChatAgent,
	historyMessages []messages.Message,
	w http.ResponseWriter,
	flusher http.Flusher,
	notificationChan chan ToolCallNotification,
) error {
	if agent.ToolsAgent == nil {
		return nil
	}

	agent.ToolsAgent.ResetMessages()

	var toolCallsResult *tools.ToolCallResult
	var err error

	modelConfig := agent.ToolsAgent.GetModelConfig()
	isParallel := modelConfig.ParallelToolCalls != nil && *modelConfig.ParallelToolCalls

	if isParallel {
		if agent.ConfirmationPromptFn != nil {
			agent.Log.Info("🔄 Using DetectParallelToolCallsWithConfirmation")
			toolCallsResult, err = agent.ToolsAgent.DetectParallelToolCallsWithConfirmation(
				historyMessages,
				agent.ExecuteFn,
				agent.ConfirmationPromptFn,
			)
		} else {
			agent.Log.Info("🔄 Using DetectParallelToolCalls")
			toolCallsResult, err = agent.ToolsAgent.DetectParallelToolCalls(
				historyMessages,
				agent.ExecuteFn,
			)
		}
	} else {
		if agent.ConfirmationPromptFn != nil {
			agent.Log.Info("🔄 Using DetectToolCallsLoopWithConfirmation (custom confirmation)")
			toolCallsResult, err = agent.ToolsAgent.DetectToolCallsLoopWithConfirmation(
				historyMessages,
				agent.ExecuteFn,
				agent.ConfirmationPromptFn,
			)
		} else {
			agent.Log.Info("🔄 Using DetectToolCallsLoopWithConfirmation (web confirmation)")
			toolCallsResult, err = agent.ToolsAgent.DetectToolCallsLoopWithConfirmation(
				historyMessages,
				agent.ExecuteFn,
				agent.WebConfirmationPrompt,
			)
		}
	}

	if err != nil {
		return err
	}

	finishReason := agent.ToolsAgent.GetLastStateToolCalls().ExecutionResult.ExecFinishReason
	agent.LogToolExecutionStatus(finishReason)

	if agent.ToolsExecutedSuccessfully(toolCallsResult, finishReason) {
		agent.AddToolResultsToContextAndStream(chatAgent, toolCallsResult, w, flusher)
	}

	return nil
}

// LogToolExecutionStatus logs the finish reason of tool execution.
func (agent *BaseServerAgent) LogToolExecutionStatus(finishReason string) {
	if finishReason == "" {
		agent.Log.Info("1️⃣ finishReasonOfExecution: %s", "empty")
	} else {
		agent.Log.Info("1️⃣ finishReasonOfExecution: %s", finishReason)
	}
}

// ToolsExecutedSuccessfully checks if tools were executed successfully.
func (agent *BaseServerAgent) ToolsExecutedSuccessfully(result *tools.ToolCallResult, finishReason string) bool {
	return len(result.Results) > 0 && finishReason == "function_executed"
}

// AddToolResultsToContextAndStream adds tool results to the chat context and streams them.
func (agent *BaseServerAgent) AddToolResultsToContextAndStream(
	chatAgent ChatAgent,
	result *tools.ToolCallResult,
	w http.ResponseWriter,
	flusher http.Flusher,
) {
	agent.Log.Info("✅ Tool calls executed successfully.")
	agent.Log.Info("📝 Tool calls results: %s", result.Results)
	agent.Log.Info("😁 Last assistant message: %s", result.LastAssistantMessage)

	chatAgent.AddMessage(roles.System, result.LastAssistantMessage)

	data := map[string]string{"message": "<hr>" + result.LastAssistantMessage + "<hr>"}
	jsonData, _ := json.Marshal(data)
	if _, err := fmt.Fprintf(w, sseDataFmt, string(jsonData)); err != nil {
		agent.Log.Error("Failed to write message: %v", err)
	}
	flusher.Flush()
}

// ShouldGenerateCompletion determines if a completion should be generated after tool execution.
func (agent *BaseServerAgent) ShouldGenerateCompletion() bool {
	if agent.ToolsAgent == nil {
		return true
	}

	state := agent.ToolsAgent.GetLastStateToolCalls()
	confirmation := state.Confirmation
	finishReason := state.ExecutionResult.ExecFinishReason

	agent.Log.Info("2️⃣ lastExecConfirmation: %v", confirmation)
	agent.Log.Info("3️⃣ lastExecFinishReason: %v", finishReason)

	return confirmation == 0 &&
		(finishReason == "user_quit" ||
			finishReason == "user_denied" ||
			finishReason == "")
}

// CleanupToolState resets tool agent state after completion.
func (agent *BaseServerAgent) CleanupToolState() {
	if agent.ToolsAgent != nil {
		agent.ToolsAgent.ResetLastStateToolCalls()
		agent.ToolsAgent.ResetMessages()
	}
}

// AddRAGContext performs similarity search and adds relevant context to the chat agent.
func (agent *BaseServerAgent) AddRAGContext(chatAgent ChatAgent, question string) {
	if agent.RagAgent == nil {
		return
	}

	similarities, err := agent.RagAgent.SearchTopN(question, agent.SimilarityLimit, agent.MaxSimilarities)
	if err != nil {
		agent.Log.Error("Error during similarity search: %v", err)
		return
	}

	if len(similarities) == 0 {
		agent.Log.Info("No relevant contexts found for the query")
		return
	}

	relevantContext := ""
	for _, sim := range similarities {
		agent.Log.Debug("Adding relevant context with similarity: %s", sim.Prompt)
		relevantContext += sim.Prompt + "\n---\n"
	}

	agent.Log.Info("Added %d similar contexts from RAG agent", len(similarities))
	chatAgent.AddMessage(
		roles.System,
		"Relevant information to help you answer the question:\n"+relevantContext,
	)
}

// StreamCompletionResponse streams the completion response via SSE.
func (agent *BaseServerAgent) StreamCompletionResponse(
	chatAgent ChatAgent,
	question string,
	w http.ResponseWriter,
	flusher http.Flusher,
) {
	agent.Log.Info("🚀 Generating streaming completion for question: %s", question)

	stopped := false
	_, errCompletion := chatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: question},
		},
		func(chunk string, finishReason string) error {
			select {
			case <-agent.StopStreamChan:
				stopped = true
				return errors.New("stream stopped by user")
			default:
			}

			if chunk != "" {
				agent.WriteSSEChunk(w, flusher, chunk)
			}

			if finishReason == "stop" && !stopped {
				agent.WriteSSEFinish(w, flusher)
			}

			return nil
		},
	)

	if errCompletion != nil && !stopped {
		agent.WriteSSEError(w, flusher, errCompletion)
	}
}

// WebConfirmationPrompt sends a confirmation prompt via web interface and waits for user response.
func (agent *BaseServerAgent) WebConfirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
	operationID := fmt.Sprintf("op_%p", &arguments)

	agent.Log.Info("🟡 Tool call detected: %s with args: %s (ID: %s)", functionName, arguments, operationID)

	responseChan := make(chan tools.ConfirmationResponse)

	agent.OperationsMutex.Lock()
	agent.PendingOperations[operationID] = &PendingOperation{
		ID:           operationID,
		FunctionName: functionName,
		Arguments:    arguments,
		Response:     responseChan,
	}
	agent.OperationsMutex.Unlock()

	message := fmt.Sprintf("Tool call detected: %s", functionName)
	agent.NotificationChanMutex.Lock()
	if agent.CurrentNotificationChan != nil {
		agent.CurrentNotificationChan <- ToolCallNotification{
			OperationID:  operationID,
			FunctionName: functionName,
			Arguments:    arguments,
			Message:      message,
		}
	}
	agent.NotificationChanMutex.Unlock()

	agent.Log.Info("⏳ Waiting for validation of operation %s", operationID)

	response := <-responseChan

	agent.Log.Info("✅ Operation %s resolved with response: %v", operationID, response)

	return response
}

// WriteSSEChunk writes a chunk of content via SSE.
func (agent *BaseServerAgent) WriteSSEChunk(w http.ResponseWriter, flusher http.Flusher, chunk string) {
	data := map[string]string{"message": chunk}
	jsonData, _ := json.Marshal(data)
	if _, err := fmt.Fprintf(w, sseDataFmt, string(jsonData)); err != nil {
		agent.Log.Error("Failed to write chunk: %v", err)
	}
	flusher.Flush()
}

// WriteSSEFinish writes a finish message via SSE.
func (agent *BaseServerAgent) WriteSSEFinish(w http.ResponseWriter, flusher http.Flusher) {
	data := map[string]string{"message": "", "finish_reason": "stop"}
	jsonData, _ := json.Marshal(data)
	if _, err := fmt.Fprintf(w, sseDataFmt, string(jsonData)); err != nil {
		agent.Log.Error("Failed to write finish reason: %v", err)
	}
	flusher.Flush()
}

// WriteSSEError writes an error message via SSE.
func (agent *BaseServerAgent) WriteSSEError(w http.ResponseWriter, flusher http.Flusher, err error) {
	data := map[string]string{"error": err.Error()}
	jsonData, _ := json.Marshal(data)
	if _, writeErr := fmt.Fprintf(w, sseDataFmt, string(jsonData)); writeErr != nil {
		agent.Log.Error("Failed to write error: %v", writeErr)
	}
	flusher.Flush()
}
