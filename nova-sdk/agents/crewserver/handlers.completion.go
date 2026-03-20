package crewserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

const handlerSSEData = "data: %s\n\n"

// ----------------------------------------
// HTTP Handlers
// ----------------------------------------

func (agent *CrewServerAgent) handleCompletion(w http.ResponseWriter, r *http.Request) {
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

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

	agent.compressContextIfNeeded(w, flusher)

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
	if err := agent.HandleToolCallsWithNotifications(agent.currentChatAgent, historyMessages, w, flusher, notificationChan); err != nil {
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
// and sends SSE notifications to the client about the compression process
func (agent *CrewServerAgent) compressContextIfNeeded(w http.ResponseWriter, flusher http.Flusher) {
	if agent.CompressorAgent == nil {
		return
	}

	contextSizeBefore := agent.currentChatAgent.GetContextSize()
	if contextSizeBefore <= agent.ContextSizeLimit {
		return
	}

	startData := map[string]interface{}{
		"role":    "information",
		"content": "🗜️ Context size limit reached. Compressing conversation history...",
	}
	jsonData, _ := json.Marshal(startData)
	if _, err := fmt.Fprintf(w, handlerSSEData, string(jsonData)); err != nil {
		agent.Log.Error("Failed to write compression start notification: %v", err)
	}
	flusher.Flush()

	newSize, err := agent.CompressChatAgentContextIfOverLimit()

	if err != nil {
		agent.Log.Error("Error during context compression: %v", err)
		errData := map[string]interface{}{
			"role":    "information",
			"content": fmt.Sprintf("❌ Compression failed: %s", err.Error()),
		}
		jsonData, _ = json.Marshal(errData)
		if _, writeErr := fmt.Fprintf(w, handlerSSEData, string(jsonData)); writeErr != nil {
			agent.Log.Error("Failed to write compression error notification: %v", writeErr)
		}
		flusher.Flush()
		return
	}

	if newSize > 0 {
		agent.Log.Info("🗜️  Chat agent context compressed from %d to %d bytes", contextSizeBefore, newSize)
		successData := map[string]interface{}{
			"role":    "information",
			"content": fmt.Sprintf("✅ Compression completed. Context reduced from %d to %d bytes.", contextSizeBefore, newSize),
		}
		jsonData, _ = json.Marshal(successData)
		if _, writeErr := fmt.Fprintf(w, handlerSSEData, string(jsonData)); writeErr != nil {
			agent.Log.Error("Failed to write compression success notification: %v", writeErr)
		}
		flusher.Flush()
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

// buildToolCallHistory creates message history for tool detection
func (agent *CrewServerAgent) buildToolCallHistory(question string) []messages.Message {
	history := agent.currentChatAgent.GetMessages()
	return append(history, messages.Message{
		Role:    roles.User,
		Content: question,
	})
}

// generateStreamingCompletion generates the final streaming completion
func (agent *CrewServerAgent) generateStreamingCompletion(
	question string,
	w http.ResponseWriter,
	flusher http.Flusher,
) {
	agent.Log.Info("👋 No tool execution was performed.")

	agent.AddRAGContext(agent.currentChatAgent, question)
	agent.routeToAppropriateAgent(question)
	agent.streamCompletionResponse(question, w, flusher)
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

// handleSelectAgentCommand handles the [select-agent <id>] command hack.
// Switches the current agent if the given ID exists, then falls through to streaming.
func (agent *CrewServerAgent) handleSelectAgentCommand(question string, w http.ResponseWriter, flusher http.Flusher) {
	if !strings.HasPrefix(question, "[select-agent") {
		return
	}
	parts := strings.Split(question, " ")
	if len(parts) < 2 {
		return
	}
	agentId := strings.Split(parts[1], "]")[0]
	if _, exists := agent.chatAgents[agentId]; exists {
		agent.currentChatAgent = agent.chatAgents[agentId]
		agent.selectedAgentId = agentId
		agent.Log.Info("💡 Manually switched to agent ID: %s", agentId)
		agent.WriteSSEChunk(w, flusher, fmt.Sprintf("<b>Switched to agent: %s.</b><br>", agentId))
		agent.WriteSSEFinish(w, flusher)
		agent.Log.Info("✅ Current agent ID is now: %s", agent.GetModelID())
	} else {
		agent.Log.Info("❌ Agent ID %s does not exist", agentId)
		agent.WriteSSEChunk(w, flusher, fmt.Sprintf("<b>Agent ID %s does not exist.</b><br>", agentId))
		agent.WriteSSEFinish(w, flusher)
	}
}

// streamCompletionResponse streams the completion response via SSE
func (agent *CrewServerAgent) streamCompletionResponse(
	question string,
	w http.ResponseWriter,
	flusher http.Flusher,
) {
	agent.Log.Info("🚀 Generating streaming completion for question: %s", question)

	// BEGIN: command hacks ([select-agent], [agent-list])
	agent.handleSelectAgentCommand(question, w, flusher)
	if strings.HasPrefix(question, "[agent-list]") {
		agentList := "Available agents:<br>"
		for id := range agent.chatAgents {
			agentList += fmt.Sprintf("- %s<br>", id)
		}
		agent.WriteSSEChunk(w, flusher, agentList)
		agent.WriteSSEFinish(w, flusher)
		return
	}
	// END: hacks

	stopped := false
	_, errCompletion := agent.currentChatAgent.GenerateStreamCompletion(
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
