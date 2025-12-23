package crewserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

func (agent *CrewServerAgent) handleCompletion(w http.ResponseWriter, r *http.Request) {

	// ------------------------------------------------------------
	// NOTE: Context packing
	// ------------------------------------------------------------
	newSize, err := agent.CompressChatAgentContextIfOverLimit()
	if err != nil {
		agent.Log.Error("Error during context compression: %v", err)
	} else if newSize > 0 {
		agent.Log.Info("🗜️  Chat agent context compressed to %d bytes", newSize)
	}

	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	question := req.Data.Message

	// Setup SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Create a notification channel for tool call updates
	notificationChan := make(chan ToolCallNotification, 10)

	// Define the current notification channel
	agent.NotificationChanMutex.Lock()
	agent.CurrentNotificationChan = notificationChan
	agent.NotificationChanMutex.Unlock()

	// Goroutine to listen for notifications and stream them to the client
	notificationDone := make(chan bool)
	go func() {
		defer close(notificationDone)
		for {
			select {
			case notification, ok := <-notificationChan:
				if !ok {
					return
				}
				// Envoyer la notification au client
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
			case <-r.Context().Done():
				return
			}
		}
	}()

	// ------------------------------------------------------------
	// NOTE: Tool calls detection and execution if toolsAgent is set
	// ------------------------------------------------------------
	// Tool calls detection and execution if toolsAgent is set
	if agent.ToolsAgent != nil {
		toolCallsResult, err := agent.ToolsAgent.DetectParallelToolCallsWithConfirmation(
			[]messages.Message{
				{Role: roles.User, Content: question},
			},
			agent.ExecuteFn,
			agent.webConfirmationPrompt,
		)

		// Closing notification channel and cleanup
		close(notificationChan)
		agent.NotificationChanMutex.Lock()
		if agent.CurrentNotificationChan == notificationChan {
			agent.CurrentNotificationChan = nil
		}
		agent.NotificationChanMutex.Unlock()

		if err != nil {
			if _, writeErr := fmt.Fprintf(w, "data: %s\n\n", jsonEscape(fmt.Sprintf("Error: %v", err))); writeErr != nil {
				agent.Log.Error("Failed to write error response: %v", writeErr)
			}
			flusher.Flush()
			return
		}

		// Add tool results to chat agent context
		if len(toolCallsResult.Results) > 0 {
			agent.currentChatAgent.AddMessage(roles.System, toolCallsResult.LastAssistantMessage)
			agent.ToolsAgent.ResetMessages()

			data := map[string]string{"message": "<hr>"}
			jsonData, _ := json.Marshal(data)
			if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
				agent.Log.Error("Failed to write separator: %v", err)
			}
			flusher.Flush()
		}
	} else {
		// Close notification channel and cleanup
		close(notificationChan)
		agent.NotificationChanMutex.Lock()
		if agent.CurrentNotificationChan == notificationChan {
			agent.CurrentNotificationChan = nil
		}
		agent.NotificationChanMutex.Unlock()
	}

	// ------------------------------------------------------------
	// NOTE: Similarity search and add to context if RAG agent is set
	// ------------------------------------------------------------
	if agent.RagAgent != nil {
		relevantContext := ""
		similarities, err := agent.RagAgent.SearchTopN(question, agent.SimilarityLimit, agent.MaxSimilarities)
		if err == nil && len(similarities) > 0 {
			for _, sim := range similarities {
				agent.Log.Debug("Adding relevant context with similarity: %s", sim.Prompt)
				relevantContext += sim.Prompt + "\n---\n"
			}
			agent.Log.Info("Added %d similar contexts from RAG agent", len(similarities))
			agent.currentChatAgent.AddMessage(
				roles.System,
				"Relevant information to help you answer the question:\n"+relevantContext,
			)
		} else {
			if err != nil {
				agent.Log.Error("Error during similarity search: %v", err)
			} else {
				agent.Log.Info("No relevant contexts found for the query")
			}
		}

	}

	// ------------------------------------------------------------
	// NOTE: Detect if we need to select another agent based on topic
	// ------------------------------------------------------------
	if agent.orchestratorAgent != nil {
		detectedAgentId, err := agent.DetectTopicThenGetAgentId(question)
		if err != nil {
			agent.Log.Error("Error during topic detection: %v", err)
		} else if detectedAgentId != "" && agent.chatAgents[detectedAgentId] != agent.currentChatAgent {
			agent.Log.Info("💡 Switching to detected agent ID: %s", detectedAgentId)
			agent.currentChatAgent = agent.chatAgents[detectedAgentId]
		}
	}
	// ------------------------------------------------------------
	// NOTE: Generate streaming completion
	// ------------------------------------------------------------

	agent.Log.Info("🚀 Generating streaming completion for question: %s", question)
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

			if chunk != "" {
				data := map[string]string{"message": chunk}
				jsonData, _ := json.Marshal(data)
				if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
					agent.Log.Error("Failed to write chunk: %v", err)
				}
				flusher.Flush()
			}

			if finishReason == "stop" && !stopped {
				data := map[string]string{"message": "", "finish_reason": "stop"}
				jsonData, _ := json.Marshal(data)
				if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
					agent.Log.Error("Failed to write finish reason: %v", err)
				}
				flusher.Flush()
			}

			return nil
		},
	)

	if errCompletion != nil && !stopped {
		data := map[string]string{"error": err.Error()}
		jsonData, _ := json.Marshal(data)
		if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
			agent.Log.Error("Failed to write error: %v", err)
		}
		flusher.Flush()
	}
}
