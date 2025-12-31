package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

var (
	chatAgent  *chat.Agent
	toolsAgent *tools.Agent
	ctx        context.Context

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
	var err error
	ctx = context.Background()

	// Initialiser les agents
	chatAgent, err = chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-chat-agent",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.NewConfig("hf.co/menlo/jan-nano-gguf:q4_k_m").WithTemperature(0.4),
	)
	if err != nil {
		panic(err)
	}

	toolsAgent, err = tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-tools-agent",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.NewConfig("hf.co/menlo/jan-nano-gguf:q4_k_m").
			WithTemperature(0.0).
			WithParallelToolCalls(true),

		tools.WithTools(GetToolsIndex()),
	)
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
	mux.HandleFunc("GET /memory/messages/context-size", handleTokensCount)
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
	chatAgent.ResetMessages()
	toolsAgent.ResetMessages()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Memory reset successfully",
	})
}

func handleMessagesList(w http.ResponseWriter, r *http.Request) {
	messages := chatAgent.GetMessages()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MemoryResponse{Messages: messages})
}

func handleTokensCount(w http.ResponseWriter, r *http.Request) {
	count := len(chatAgent.GetMessages())
	tokens := chatAgent.GetContextSize()
	limit := 9999
	//limit := chatAgent.GetContextLimit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokensResponse{
		Count:  count,
		Tokens: tokens,
		Limit:  limit,
	})
}

func handleModelsInformation(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":           "ok",
		"chat_model":       chatAgent.GetName(),
		"embeddings_model": "to be defined",
		"tools_model":      toolsAgent.GetName(),
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

	// Détection des tool calls avec confirmation web
	toolCallsResult, err := toolsAgent.DetectParallelToolCallsWithConfirmation(
		[]messages.Message{
			{Role: roles.User, Content: question},
		},
		executeFunction,
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

	// Ajouter les résultats des tool calls à l'agent de chat
	if len(toolCallsResult.Results) > 0 {
		chatAgent.AddMessage(roles.System, toolCallsResult.LastAssistantMessage)
		toolsAgent.ResetMessages()

		// Ligne de séparation après la fin des validations
		data := map[string]string{"message": "<hr>"}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		flusher.Flush()
	}

	// Générer la réponse en streaming
	stopped := false
	_, err = chatAgent.GenerateStreamCompletion(
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

func GetToolsIndex() []*tools.Tool {
	calculateSumTool := tools.NewTool("calculate_sum").
		SetDescription("Calculate the sum of two numbers").
		AddParameter("a", "number", "The first number", true).
		AddParameter("b", "number", "The second number", true)

	sayHelloTool := tools.NewTool("say_hello").
		SetDescription("Say hello to the given name").
		AddParameter("name", "string", "The name to greet", true)

	sayExit := tools.NewTool("say_exit").
		SetDescription("Say exit")

	return []*tools.Tool{
		calculateSumTool,
		sayHelloTool,
		sayExit,
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	log.Printf("🟢 Executing function: %s with arguments: %s", functionName, arguments)

	switch functionName {
	case "say_hello":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for say_hello"}`, nil
		}
		hello := fmt.Sprintf("👋 Hello, %s!🙂", args.Name)
		return fmt.Sprintf(`{"message": "%s"}`, hello), nil

	case "calculate_sum":
		var args struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for calculate_sum"}`, nil
		}
		sum := args.A + args.B
		return fmt.Sprintf(`{"result": %g}`, sum), nil

	case "say_exit":
		return fmt.Sprintf(`{"message": "%s"}`, "❌ EXIT"), errors.New("exit_loop")

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}

// jsonEscape échappe une chaîne pour l'inclusion dans JSON
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
