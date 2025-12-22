package crewserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
)

// ----------------------------------------
// HTTP Handlers
// ----------------------------------------

func (agent *CrewServerAgent) handleOperationValidate(w http.ResponseWriter, r *http.Request) {
	var req OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Configurer le streaming SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	agent.operationsMutex.Lock()
	op, exists := agent.pendingOperations[req.OperationID]
	if exists {
		delete(agent.pendingOperations, req.OperationID)
	}
	agent.operationsMutex.Unlock()

	if !exists {
		data := map[string]string{"message": fmt.Sprintf("❌ Operation %s not found", req.OperationID)}
		jsonData, _ := json.Marshal(data)
		if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
			agent.log.Error("Failed to write operation not found response: %v", err)
		}
		flusher.Flush()
		return
	}

	// Envoyer un message de confirmation à l'UI
	data := map[string]string{"message": fmt.Sprintf("✅ Operation %s validated\n", req.OperationID)}
	jsonData, _ := json.Marshal(data)
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		agent.log.Error("Failed to write validation response: %v", err)
		return
	}
	flusher.Flush()

	// Envoyer la confirmation au canal
	op.Response <- tools.Confirmed
}

func (agent *CrewServerAgent) handleOperationCancel(w http.ResponseWriter, r *http.Request) {
	var req OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Configurer le streaming SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	agent.operationsMutex.Lock()
	op, exists := agent.pendingOperations[req.OperationID]
	if exists {
		delete(agent.pendingOperations, req.OperationID)
	}
	agent.operationsMutex.Unlock()

	if !exists {
		data := map[string]string{"message": fmt.Sprintf("❌ Operation %s not found", req.OperationID)}
		jsonData, _ := json.Marshal(data)
		if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
			agent.log.Error("Failed to write operation not found response: %v", err)
		}
		flusher.Flush()
		return
	}

	// Envoyer un message d'annulation à l'UI
	data := map[string]string{"message": fmt.Sprintf("⛔️ Operation %s cancelled", req.OperationID)}
	jsonData, _ := json.Marshal(data)
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		agent.log.Error("Failed to write cancellation response: %v", err)
		return
	}
	flusher.Flush()

	// Envoyer le refus au canal
	op.Response <- tools.Denied
}

func (agent *CrewServerAgent) handleOperationReset(w http.ResponseWriter, r *http.Request) {
	// Configurer le streaming SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	agent.operationsMutex.Lock()
	count := len(agent.pendingOperations)
	// Refuser toutes les opérations en attente
	for id, op := range agent.pendingOperations {
		op.Response <- tools.Quit
		delete(agent.pendingOperations, id)
	}
	agent.operationsMutex.Unlock()

	// Envoyer un message de confirmation à l'UI
	data := map[string]string{"message": fmt.Sprintf("🔄 All pending operations cancelled (%d operations)", count)}
	jsonData, _ := json.Marshal(data)
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		agent.log.Error("Failed to write reset response: %v", err)
	}
	flusher.Flush()
}
