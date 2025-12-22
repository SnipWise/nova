package crewserver

import (
	"encoding/json"
	"net/http"
)

// ----------------------------------------
// HTTP Handlers
// ----------------------------------------

func (agent *CrewServerAgent) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		agent.log.Error("Failed to encode health response: %v", err)
	}
}

func (agent *CrewServerAgent) handleMemoryReset(w http.ResponseWriter, r *http.Request) {
	agent.currentChatAgent.ResetMessages()
	if agent.toolsAgent != nil {
		agent.toolsAgent.ResetMessages()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Memory reset successfully",
	}); err != nil {
		agent.log.Error("Failed to encode memory reset response: %v", err)
	}
}

func (agent *CrewServerAgent) handleMessagesList(w http.ResponseWriter, r *http.Request) {
	messages := agent.currentChatAgent.GetMessages()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(MemoryResponse{Messages: messages}); err != nil {
		agent.log.Error("Failed to encode messages list response: %v", err)
	}
}

func (agent *CrewServerAgent) handleTokensCount(w http.ResponseWriter, r *http.Request) {
	count := len(agent.currentChatAgent.GetMessages())
	tokens := agent.currentChatAgent.GetContextSize()
	limit := 9999

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(TokensResponse{
		Count:  count,
		Tokens: tokens,
		Limit:  limit,
	}); err != nil {
		agent.log.Error("Failed to encode tokens count response: %v", err)
	}
}

func (agent *CrewServerAgent) handleModelsInformation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	toolsModel := "none"
	if agent.toolsAgent != nil {
		toolsModel = agent.toolsAgent.GetModelID()
	}
	embeddingsModel := "none"
	if agent.ragAgent != nil {
		embeddingsModel = agent.ragAgent.GetModelID()
	}

	if err := json.NewEncoder(w).Encode(map[string]any{
		"status":           "ok",
		"chat_model":       agent.currentChatAgent.GetModelID(),
		"embeddings_model": embeddingsModel,
		"tools_model":      toolsModel,
	}); err != nil {
		agent.log.Error("Failed to encode models information response: %v", err)
	}
}
