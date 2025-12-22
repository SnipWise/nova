package server

import (
	"encoding/json"
	"net/http"
)

// ----------------------------------------
// HTTP Handlers
// ----------------------------------------

func (agent *ServerAgent) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		agent.log.Error("Failed to encode health response: %v", err)
	}
}

func (agent *ServerAgent) handleMemoryReset(w http.ResponseWriter, r *http.Request) {
	agent.chatAgent.ResetMessages()
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

func (agent *ServerAgent) handleMessagesList(w http.ResponseWriter, r *http.Request) {
	messages := agent.chatAgent.GetMessages()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(MemoryResponse{Messages: messages}); err != nil {
		agent.log.Error("Failed to encode messages list response: %v", err)
	}
}

func (agent *ServerAgent) handleTokensCount(w http.ResponseWriter, r *http.Request) {
	count := len(agent.chatAgent.GetMessages())
	tokens := agent.chatAgent.GetContextSize()
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

func (agent *ServerAgent) handleModelsInformation(w http.ResponseWriter, r *http.Request) {
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
		"chat_model":       agent.chatAgent.GetModelID(),
		"embeddings_model": embeddingsModel,
		"tools_model":      toolsModel,
	}); err != nil {
		agent.log.Error("Failed to encode models information response: %v", err)
	}
}
