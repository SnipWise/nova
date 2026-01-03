package crewserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// ----------------------------------------
// HTTP Handlers
// ----------------------------------------

func (agent *CrewServerAgent) handleCompletionStop(w http.ResponseWriter, r *http.Request) {
	select {
	case agent.StopStreamChan <- true:
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Stream stopped",
		}); err != nil {
			agent.Log.Error("Failed to encode completion stop response: %v", err)
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "No stream to stop",
		}); err != nil {
			agent.Log.Error("Failed to encode completion stop response: %v", err)
		}
	}
}

// handleCompletion processes completion requests with SSE streaming:
// 1. Compresses context if needed
// 2. Parses request and sets up SSE streaming
// 3. Manages tool call notifications
// 4. Executes tool calls if detected
// 5. Adds RAG context if available
// 6. Routes to appropriate agent if orchestrator is configured
// 7. Generates streaming completion
func (agent *CrewServerAgent) handleCompletion(w http.ResponseWriter, r *http.Request) {
	// Step 1: Compress context if needed
	agent.compressContextIfNeeded()

	// Step 2: Parse request
	question, err := agent.parseCompletionRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Step 3: Setup SSE streaming
	flusher, err := agent.setupSSEHeaders(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 4: Setup and start notification streaming
	notificationChan := agent.setupNotificationChannel()
	agent.startNotificationStreaming(w, r, flusher, notificationChan)

	// Step 5: Handle tool calls if configured
	if err := agent.handleToolCallsWithNotifications(question, w, flusher, notificationChan); err != nil {
		agent.closeNotificationChannel(notificationChan)
		agent.writeSSEError(w, flusher, err)
		return
	}

	// Step 6: Close notification channel after tool execution
	agent.closeNotificationChannel(notificationChan)

	// Step 7: Generate completion if needed
	if agent.shouldGenerateCompletion() {
		agent.generateStreamingCompletion(question, w, flusher)
	}

	// Step 8: Cleanup tool state
	agent.cleanupToolState()
}

// compressContextIfNeeded compresses the chat context if compressor is configured
func (agent *CrewServerAgent) compressContextIfNeeded() {
	if agent.CompressorAgent == nil {
		return
	}

	newSize, err := agent.CompressChatAgentContextIfOverLimit()
	if err != nil {
		agent.Log.Error("Error during context compression: %v", err)
		return
	}

	if newSize > 0 {
		agent.Log.Info("🗜️  Chat agent context compressed to %d bytes", newSize)
	}
}

// parseCompletionRequest parses the incoming completion request
func (agent *CrewServerAgent) parseCompletionRequest(r *http.Request) (string, error) {
	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", err
	}
	return req.Data.Message, nil
}

// setupSSEHeaders configures SSE streaming headers
func (agent *CrewServerAgent) setupSSEHeaders(w http.ResponseWriter) (http.Flusher, error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}
	return flusher, nil
}

// setupNotificationChannel creates and registers a notification channel
func (agent *CrewServerAgent) setupNotificationChannel() chan ToolCallNotification {
	notificationChan := make(chan ToolCallNotification, 10)
	agent.NotificationChanMutex.Lock()
	agent.CurrentNotificationChan = notificationChan
	agent.NotificationChanMutex.Unlock()
	return notificationChan
}

// startNotificationStreaming starts a goroutine to stream notifications to the client
func (agent *CrewServerAgent) startNotificationStreaming(
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
				agent.sendToolCallNotification(w, flusher, notification)
			case <-r.Context().Done():
				return
			}
		}
	}()
}

// sendToolCallNotification sends a single tool call notification via SSE
func (agent *CrewServerAgent) sendToolCallNotification(
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
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		agent.Log.Error("Failed to write notification: %v", err)
		return
	}
	flusher.Flush()
}

// closeNotificationChannel safely closes the notification channel
func (agent *CrewServerAgent) closeNotificationChannel(notificationChan chan ToolCallNotification) {
	close(notificationChan)
	agent.NotificationChanMutex.Lock()
	if agent.CurrentNotificationChan == notificationChan {
		agent.CurrentNotificationChan = nil
	}
	agent.NotificationChanMutex.Unlock()
}

// handleToolCallsWithNotifications executes tool calls if toolsAgent is configured
func (agent *CrewServerAgent) handleToolCallsWithNotifications(
	question string,
	w http.ResponseWriter,
	flusher http.Flusher,
	notificationChan chan ToolCallNotification,
) error {
	if agent.ToolsAgent == nil {
		return nil
	}

	agent.ToolsAgent.ResetMessages()

	// Prepare message history
	historyMessages := agent.buildToolCallHistory(question)

	// Detect and execute tool calls
	toolCallsResult, err := agent.ToolsAgent.DetectToolCallsLoopWithConfirmation(
		historyMessages,
		agent.ExecuteFn,
		agent.webConfirmationPrompt,
	)
	if err != nil {
		return err
	}

	// Process tool execution results
	finishReason := agent.ToolsAgent.GetLastStateToolCalls().ExecutionResult.ExecFinishReason
	agent.logToolExecutionStatus(finishReason)

	// Add tool results to chat context if execution succeeded
	if agent.toolsExecutedSuccessfully(toolCallsResult, finishReason) {
		agent.addToolResultsToContextAndStream(toolCallsResult, w, flusher)
	}

	return nil
}

// buildToolCallHistory creates message history for tool detection
func (agent *CrewServerAgent) buildToolCallHistory(question string) []messages.Message {
	history := agent.currentChatAgent.GetMessages()
	return append(history, messages.Message{
		Role:    roles.User,
		Content: question,
	})
}

// logToolExecutionStatus logs the finish reason of tool execution
func (agent *CrewServerAgent) logToolExecutionStatus(finishReason string) {
	if finishReason == "" {
		agent.Log.Info("1️⃣ finishReasonOfExecution: %s", "empty")
	} else {
		agent.Log.Info("1️⃣ finishReasonOfExecution: %s", finishReason)
	}
}

// toolsExecutedSuccessfully checks if tools were executed successfully
func (agent *CrewServerAgent) toolsExecutedSuccessfully(result *tools.ToolCallResult, finishReason string) bool {
	return len(result.Results) > 0 && finishReason == "function_executed"
}

// addToolResultsToContextAndStream adds tool results to context and streams them
func (agent *CrewServerAgent) addToolResultsToContextAndStream(
	result *tools.ToolCallResult,
	w http.ResponseWriter,
	flusher http.Flusher,
) {
	agent.Log.Info("✅ Tool calls executed successfully.")
	agent.Log.Info("📝 Tool calls results: %s", result.Results)
	agent.Log.Info("😁 Last assistant message: %s", result.LastAssistantMessage)

	agent.currentChatAgent.AddMessage(roles.System, result.LastAssistantMessage)

	data := map[string]string{"message": "<hr>" + result.LastAssistantMessage + "<hr>"}
	jsonData, _ := json.Marshal(data)
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		agent.Log.Error("Failed to write message: %v", err)
	}
	flusher.Flush()
}

// shouldGenerateCompletion determines if we should generate a completion
func (agent *CrewServerAgent) shouldGenerateCompletion() bool {
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

// generateStreamingCompletion generates the final streaming completion
func (agent *CrewServerAgent) generateStreamingCompletion(
	question string,
	w http.ResponseWriter,
	flusher http.Flusher,
) {
	agent.Log.Info("👋 No tool execution was performed.")

	// Add RAG context if available
	agent.addRAGContext(question)

	// Switch agent based on topic if orchestrator is configured
	agent.routeToAppropriateAgent(question)

	// Generate streaming response
	agent.streamCompletionResponse(question, w, flusher)
}

// addRAGContext performs similarity search and adds relevant context
func (agent *CrewServerAgent) addRAGContext(question string) {
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
	agent.currentChatAgent.AddMessage(
		roles.System,
		"Relevant information to help you answer the question:\n"+relevantContext,
	)
}

// routeToAppropriateAgent detects topic and switches to appropriate agent
func (agent *CrewServerAgent) routeToAppropriateAgent(question string) {
	if agent.orchestratorAgent == nil {
		return
	}

	detectedAgentId, err := agent.DetectTopicThenGetAgentId(question)
	if err != nil {
		agent.Log.Error("Error during topic detection: %v", err)
		return
	}

	if detectedAgentId != "" && agent.chatAgents[detectedAgentId] != agent.currentChatAgent {
		agent.Log.Info("💡 Switching to detected agent ID: %s", detectedAgentId)
		agent.currentChatAgent = agent.chatAgents[detectedAgentId]
		agent.selectedAgentId = detectedAgentId
	}
}

// streamCompletionResponse streams the completion response via SSE
func (agent *CrewServerAgent) streamCompletionResponse(
	question string,
	w http.ResponseWriter,
	flusher http.Flusher,
) {
	agent.Log.Info("🚀 Generating streaming completion for question: %s", question)

	// NOTE: hooks/hacks to add some commands
	// BEGIN: hacks
	if strings.HasPrefix(question, "[select-agent") {
		parts := strings.Split(question, " ")
		if len(parts) >= 2 {
			agentId := strings.Split(parts[1], "]")[0]

			if _, exists := agent.chatAgents[agentId]; exists {
				agent.currentChatAgent = agent.chatAgents[agentId]
				agent.selectedAgentId = agentId
				agent.Log.Info("💡 Manually switched to agent ID: %s", agentId)
				agent.writeSSEChunk(w, flusher, fmt.Sprintf("<b>Switched to agent: %s.</b><br>", agentId))
				agent.writeSSEFinish(w, flusher)

				agent.Log.Info("✅ Current agent ID is now: %s", agent.GetModelID())
				//return
			} else {
				agent.Log.Info("❌ Agent ID %s does not exist", agentId)
				agent.writeSSEChunk(w, flusher, fmt.Sprintf("<b>Agent ID %s does not exist.</b><br>", agentId))
				agent.writeSSEFinish(w, flusher)
				//return
			}
		}
	}
	if strings.HasPrefix(question, "[agent-list]") {
		agentList := "Available agents:<br>"
		for id := range agent.chatAgents {
			agentList += fmt.Sprintf("- %s<br>", id)
		}
		agent.writeSSEChunk(w, flusher, agentList)
		agent.writeSSEFinish(w, flusher)
		return
	}

	// END: hacks

	stopped := false
	_, errCompletion := agent.currentChatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: question},
		},
		func(chunk string, finishReason string) error {
			// Check if stop signal received
			select {
			case <-agent.StopStreamChan:
				stopped = true
				return errors.New("stream stopped by user")
			default:
			}

			// Stream chunk if not empty
			if chunk != "" {
				agent.writeSSEChunk(w, flusher, chunk)
			}

			// Send finish reason if stream completed
			if finishReason == "stop" && !stopped {
				agent.writeSSEFinish(w, flusher)
			}

			return nil
		},
	)

	// Handle completion error
	if errCompletion != nil && !stopped {
		agent.writeSSEError(w, flusher, errCompletion)
	}
}

// writeSSEChunk writes a chunk of content via SSE
func (agent *CrewServerAgent) writeSSEChunk(w http.ResponseWriter, flusher http.Flusher, chunk string) {
	data := map[string]string{"message": chunk}
	jsonData, _ := json.Marshal(data)
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		agent.Log.Error("Failed to write chunk: %v", err)
	}
	flusher.Flush()
}

// writeSSEFinish writes a finish message via SSE
func (agent *CrewServerAgent) writeSSEFinish(w http.ResponseWriter, flusher http.Flusher) {
	data := map[string]string{"message": "", "finish_reason": "stop"}
	jsonData, _ := json.Marshal(data)
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		agent.Log.Error("Failed to write finish reason: %v", err)
	}
	flusher.Flush()
}

// writeSSEError writes an error message via SSE
func (agent *CrewServerAgent) writeSSEError(w http.ResponseWriter, flusher http.Flusher, err error) {
	data := map[string]string{"error": err.Error()}
	jsonData, _ := json.Marshal(data)
	if _, writeErr := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); writeErr != nil {
		agent.Log.Error("Failed to write error: %v", writeErr)
	}
	flusher.Flush()
}

// cleanupToolState resets tool agent state after completion
func (agent *CrewServerAgent) cleanupToolState() {
	if agent.ToolsAgent != nil {
		agent.ToolsAgent.ResetLastStateToolCalls()
		agent.ToolsAgent.ResetMessages()
	}
}
