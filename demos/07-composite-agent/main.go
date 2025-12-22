package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/mcptools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
)

var (
	compositeAgent *CompositeAgent
	mcpClient      *mcptools.MCPClient

	// Gestion des opérations en attente de validation
	pendingOperations = make(map[string]*PendingOperation)
	operationsMutex   sync.RWMutex

	// Canal pour arrêter le streaming
	stopStreamChan chan bool

	// Canal de notification actuel (par session)
	currentNotificationChan chan ToolCallNotification
	notificationChanMutex   sync.Mutex
)

type ToolCallNotification struct {
	OperationID  string
	FunctionName string
	Arguments    string
	Message      string
}

type PendingOperation struct {
	ID           string
	FunctionName string
	Arguments    string
	Response     chan tools.ConfirmationResponse
}

type CompletionRequest struct {
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

type OperationRequest struct {
	OperationID string `json:"operation_id"`
}

type OperationResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	OperationID string `json:"operation_id,omitempty"`
}

type MemoryResponse struct {
	Messages []messages.Message `json:"messages"`
}

type TokensResponse struct {
	Count  int `json:"count"`
	Tokens int `json:"tokens"`
	Limit  int `json:"limit"`
}

func main() {

	ctx := context.Background()

	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
	mcpServerURL := env.GetEnvOrDefault("MCP_SERVER_URL", "http://localhost:9011")
	var err error
	mcpClient, err = mcptools.NewStreamableHttpMCPClient(ctx, mcpServerURL)
	if err != nil {
		panic(err)
	}

	// Print available tools
	for _, tool := range mcpClient.GetTools() {
		println("- Tool:", tool.Name, ":", tool.Description)
	}

	// Initialize all chat agents
	compositeAgent, err = NewCompositeAgent(ctx, engineURL, mcpClient.GetTools())

	if err != nil {
		panic(err)
	}

	err = compositeAgent.SetCurrentAgent("generic")
	if err != nil {
		panic(err)
	}

	// Initialiser le canal d'arrêt
	stopStreamChan = make(chan bool, 1)

	// Configuration du serveur HTTP
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("POST /completion", handleCompletion)
	mux.HandleFunc("POST /completion/stop", handleCompletionStop)
	mux.HandleFunc("POST /memory/reset", handleMemoryReset)
	mux.HandleFunc("GET /memory/messages/list", handleMessagesList)
	mux.HandleFunc("GET /memory/messages/tokens", handleTokensCount)
	mux.HandleFunc("POST /operation/validate", handleOperationValidate)
	mux.HandleFunc("POST /operation/cancel", handleOperationCancel)
	mux.HandleFunc("POST /operation/reset", handleOperationReset)
	mux.HandleFunc("GET /models", handleModelsInformation)
	mux.HandleFunc("GET /health", handleHealth)

	// Démarrer le serveur
	port := ":3500"
	fmt.Printf("🚀 Server started on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleMemoryReset(w http.ResponseWriter, r *http.Request) {

	compositeAgent.ResetCurrentChatAgentMemory()
	compositeAgent.ResetToolsAgentMemory()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Memory reset successfully",
	})
}

func handleMessagesList(w http.ResponseWriter, r *http.Request) {

	messages := compositeAgent.GetCurrentAgentMessages()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MemoryResponse{Messages: messages})
}

func handleTokensCount(w http.ResponseWriter, r *http.Request) {

	count := len(compositeAgent.GetCurrentAgentMessages())
	tokens := compositeAgent.GetCurrentAgentContextSize()
	limit := compositeAgent.GetContextSizeLimit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokensResponse{
		Count:  count,
		Tokens: tokens,
		Limit:  limit,
	})
}

func handleModelsInformation(w http.ResponseWriter, r *http.Request) {
	var chatModelID string
	chatAgent, err := compositeAgent.GetCurrentAgent()
	if err != nil {
		chatModelID = "unknown"
	} else {
		chatModelID = chatAgent.GetModelID()
	}
	//chatModelID = chatAgent.GetModelID()

	embeddingsModelID := compositeAgent.ragAgent.GetModelID()
	toolsModelID := compositeAgent.toolsAgent.GetModelID()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":           "ok",
		"chat_model":       chatModelID + "[" + compositeAgent.currentAgent.GetName() + "]",
		"embeddings_model": embeddingsModelID + "[" + compositeAgent.ragAgent.GetName() + "]",
		"tools_model":      toolsModelID + "[" + compositeAgent.toolsAgent.GetName() + "]",
	})
}

func handleCompletionStop(w http.ResponseWriter, r *http.Request) {
	select {
	case stopStreamChan <- true:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Stream stopped",
		})
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "No stream to stop",
		})
	}
}

func handleCompletion(w http.ResponseWriter, r *http.Request) {
	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	question := req.Data.Message

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

	// Créer un canal de notification pour cette requête
	notificationChan := make(chan ToolCallNotification, 10)

	// Définir ce canal comme le canal actuel
	notificationChanMutex.Lock()
	currentNotificationChan = notificationChan
	notificationChanMutex.Unlock()

	// Goroutine pour écouter les notifications de tool calls
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
				fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}()

	selectedAgent,err := compositeAgent.SetCurrentAgentByTopic(question)
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", jsonEscape(fmt.Sprintf("Error: %v", err)))
		flusher.Flush()
		return
	}
	// TODO: use logs
	fmt.Println("🤖 "+ selectedAgent.GetName())

	// Tool calls detection
	// We can use:
	// compositeAgent.DetectParallelToolCallsWithConfirmation(
	// query string,
	// toolCallback tools.ToolCallback,
	// confirmationCallback tools.ConfirmationCallback)
	toolCallsResult, err := compositeAgent.toolsAgent.DetectParallelToolCallsWithConfirmation(
		[]messages.Message{
			{Role: roles.User, Content: question},
		},
		// [MCP] Tool execution function
		func(functionName, arguments string) (string, error) {
			// TODO: add logs
			result, err := mcpClient.ExecToolWithString(functionName, arguments)
			if err != nil {
				return "", err
			}
			return result, err
		},
		webConfirmationPrompt,
	)

	// Fermer le canal de notification et nettoyer
	close(notificationChan)
	notificationChanMutex.Lock()
	if currentNotificationChan == notificationChan {
		currentNotificationChan = nil
	}
	notificationChanMutex.Unlock()
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", jsonEscape(fmt.Sprintf("Error: %v", err)))
		flusher.Flush()
		return
	}

	currentChatAgent, err := compositeAgent.GetCurrentAgent()
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", jsonEscape(fmt.Sprintf("Error: %v", err)))
		flusher.Flush()
		return
	}

	// Ajouter les résultats des tool calls à l'agent de chat
	if len(toolCallsResult.Results) > 0 {

		// TODO: create compositeAgent.AddMessageToCurrentAgent IMPORTANT:
		currentChatAgent.AddMessage(roles.System, toolCallsResult.LastAssistantMessage)
		compositeAgent.ResetToolsAgentMemory()

		// Ligne de séparation après la fin des validations
		data := map[string]string{"message": "<hr>"}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		flusher.Flush()
	}

	compositeAgent.SearchSimilaritiesAndAddToCurrentAgentContext(
		question,
		compositeAgent.GetSimilarityLimit(),
		compositeAgent.GetMaxSimilarities(),
	)

	// Générer la réponse en streaming
	stopped := false
	_, err = currentChatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: question},
		},
		func(chunk string, finishReason string) error {
			// Vérifier si on doit arrêter
			select {
			case <-stopStreamChan:
				stopped = true
				return errors.New("stream stopped by user")
			default:
			}

			if chunk != "" {
				data := map[string]string{"message": chunk}
				jsonData, _ := json.Marshal(data)
				fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
				flusher.Flush()
			}

			if finishReason == "stop" && !stopped {
				data := map[string]string{"message": "", "finish_reason": "stop"}
				jsonData, _ := json.Marshal(data)
				fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
				flusher.Flush()
			}

			return nil
		},
	)

	if err != nil && !stopped {
		data := map[string]string{"error": err.Error()}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		flusher.Flush()
	}
}

func handleOperationValidate(w http.ResponseWriter, r *http.Request) {
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

	operationsMutex.Lock()
	op, exists := pendingOperations[req.OperationID]
	if exists {
		delete(pendingOperations, req.OperationID)
	}
	operationsMutex.Unlock()

	if !exists {
		data := map[string]string{"message": fmt.Sprintf("❌ Operation %s not found", req.OperationID)}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		flusher.Flush()
		return
	}

	// Envoyer un message de confirmation à l'UI
	data := map[string]string{"message": fmt.Sprintf("✅ Operation %s validated\n", req.OperationID)}
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	flusher.Flush()

	// Envoyer la confirmation au canal
	op.Response <- tools.Confirmed
}

func handleOperationCancel(w http.ResponseWriter, r *http.Request) {
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

	operationsMutex.Lock()
	op, exists := pendingOperations[req.OperationID]
	if exists {
		delete(pendingOperations, req.OperationID)
	}
	operationsMutex.Unlock()

	if !exists {
		data := map[string]string{"message": fmt.Sprintf("❌ Operation %s not found", req.OperationID)}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		flusher.Flush()
		return
	}

	// Envoyer un message d'annulation à l'UI
	data := map[string]string{"message": fmt.Sprintf("⛔️ Operation %s cancelled", req.OperationID)}
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	flusher.Flush()

	// Envoyer le refus au canal
	op.Response <- tools.Denied
}

func handleOperationReset(w http.ResponseWriter, r *http.Request) {
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

	operationsMutex.Lock()
	count := len(pendingOperations)
	// Refuser toutes les opérations en attente
	for id, op := range pendingOperations {
		op.Response <- tools.Quit
		delete(pendingOperations, id)
	}
	operationsMutex.Unlock()

	// Envoyer un message de confirmation à l'UI
	data := map[string]string{"message": fmt.Sprintf("🔄 All pending operations cancelled (%d operations)", count)}
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	flusher.Flush()
}

// webConfirmationPrompt est une fonction qui attend la validation via HTTP
func webConfirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
	operationID := fmt.Sprintf("op_%p", &arguments)

	log.Printf("🟡 Tool call detected: %s with args: %s (ID: %s)", functionName, arguments, operationID)

	// Créer un canal pour recevoir la réponse
	responseChan := make(chan tools.ConfirmationResponse)

	// Enregistrer l'opération en attente
	operationsMutex.Lock()
	pendingOperations[operationID] = &PendingOperation{
		ID:           operationID,
		FunctionName: functionName,
		Arguments:    arguments,
		Response:     responseChan,
	}
	operationsMutex.Unlock()

	// Envoyer une notification au client via le canal actuel
	message := fmt.Sprintf("Tool call detected: %s", functionName)
	notificationChanMutex.Lock()
	if currentNotificationChan != nil {
		currentNotificationChan <- ToolCallNotification{
			OperationID:  operationID,
			FunctionName: functionName,
			Arguments:    arguments,
			Message:      message,
		}
	}
	notificationChanMutex.Unlock()

	log.Printf("⏳ Waiting for validation of operation %s", operationID)

	// Attendre la réponse via HTTP
	response := <-responseChan

	log.Printf("✅ Operation %s resolved with response: %v", operationID, response)

	return response
}

// jsonEscape échappe une chaîne pour l'inclusion dans JSON
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
