package gatewayserver

import (
	"encoding/json"
	"net/http"
	"time"
)

// handleListModels returns the available models in OpenAI format (GET /v1/models).
func (agent *GatewayServerAgent) handleListModels(w http.ResponseWriter, r *http.Request) {
	var models []ModelEntry

	// Add an entry for each chat agent in the crew
	for id, chatAgent := range agent.chatAgents {
		models = append(models, ModelEntry{
			ID:      chatAgent.GetModelID(),
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "nova-gateway:" + id,
		})
	}

	// Add tools model if configured
	if agent.toolsAgent != nil {
		models = append(models, ModelEntry{
			ID:      agent.toolsAgent.GetModelID(),
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "nova-gateway:tools",
		})
	}

	response := ModelsResponse{
		Object: "list",
		Data:   models,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		agent.log.Error("Failed to encode models response: %v", err)
	}
}

// HandleHealthForTest exposes handleHealth for testing.
func (agent *GatewayServerAgent) HandleHealthForTest(w http.ResponseWriter, r *http.Request) {
	agent.handleHealth(w, r)
}

// HandleListModelsForTest exposes handleListModels for testing.
func (agent *GatewayServerAgent) HandleListModelsForTest(w http.ResponseWriter, r *http.Request) {
	agent.handleListModels(w, r)
}

// handleHealth handles the health check endpoint (GET /health).
func (agent *GatewayServerAgent) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		agent.log.Error("Failed to encode health response: %v", err)
	}
}
