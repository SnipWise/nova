package server

import (
	"encoding/json"
	"net/http"

	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// ----------------------------------------
// HTTP Handlers
// ----------------------------------------

func (agent *ServerAgent) handleCompletion(w http.ResponseWriter, r *http.Request) {
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	agent.compressContextIfNeeded()

	question, err := agent.parseCompletionRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flusher, err := agent.SetupSSEHeaders(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if planExecuted, planErr := agent.executePlanHTTP(question, w, flusher); planErr != nil {
		agent.WriteSSEError(w, flusher, planErr)
		if agent.afterCompletion != nil {
			agent.afterCompletion(agent)
		}
		return
	} else if planExecuted {
		if agent.afterCompletion != nil {
			agent.afterCompletion(agent)
		}
		return
	}

	notificationChan := agent.SetupNotificationChannel()
	agent.StartNotificationStreaming(w, r, flusher, notificationChan)

	historyMessages := agent.buildToolCallHistory(question)
	if err := agent.HandleToolCallsWithNotifications(agent.chatAgent, historyMessages, w, flusher, notificationChan); err != nil {
		agent.CloseNotificationChannel(notificationChan)
		agent.WriteSSEError(w, flusher, err)
		return
	}

	agent.CloseNotificationChannel(notificationChan)

	if agent.ShouldGenerateCompletion() {
		agent.generateStreamingCompletion(question, w, flusher)
	}

	agent.CleanupToolState()

	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}
}

// compressContextIfNeeded compresses the chat context if compressor is configured
func (agent *ServerAgent) compressContextIfNeeded() {
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
func (agent *ServerAgent) parseCompletionRequest(r *http.Request) (string, error) {
	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", err
	}
	return req.Data.Message, nil
}

// buildToolCallHistory creates message history for tool detection
func (agent *ServerAgent) buildToolCallHistory(question string) []messages.Message {
	history := agent.chatAgent.GetMessages()
	return append(history, messages.Message{
		Role:    roles.User,
		Content: question,
	})
}

// generateStreamingCompletion generates the final streaming completion
func (agent *ServerAgent) generateStreamingCompletion(
	question string,
	w http.ResponseWriter,
	flusher http.Flusher,
) {
	agent.Log.Info("👋 No tool execution was performed.")
	agent.AddRAGContext(agent.chatAgent, question)
	agent.StreamCompletionResponse(agent.chatAgent, question, w, flusher)
}
