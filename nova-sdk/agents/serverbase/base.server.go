package serverbase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

const (
	headerContentType     = "Content-Type"
	contentTypeJSON       = "application/json"
	contentTypeSSE        = "text/event-stream"
	headerCacheControl    = "Cache-Control"
	headerConnection      = "Connection"
	headerAccessControl   = "Access-Control-Allow-Origin"
	cacheControlNoCache   = "no-cache"
	connectionKeepAlive   = "keep-alive"
	accessControlWildcard = "*"

	errStreamingNotSupported = "Streaming not supported"
	sseDataFmt               = "data: %s\n\n"
)

// setSSEHeaders configures SSE streaming headers and returns a flusher.
// Returns nil, false if streaming is not supported by the ResponseWriter.
func setSSEHeaders(w http.ResponseWriter) (http.Flusher, bool) {
	w.Header().Set(headerContentType, contentTypeSSE)
	w.Header().Set(headerCacheControl, cacheControlNoCache)
	w.Header().Set(headerConnection, connectionKeepAlive)
	w.Header().Set(headerAccessControl, accessControlWildcard)
	flusher, ok := w.(http.Flusher)
	return flusher, ok
}

// BaseServerAgent contains common server agent functionality
type BaseServerAgent struct {
	ChatAgent        ChatAgent
	ToolsAgent       *tools.Agent
	RagAgent         *rag.Agent
	SimilarityLimit  float64
	MaxSimilarities  int
	ContextSizeLimit int
	CompressorAgent  *compressor.Agent
	Port             string
	Ctx              context.Context
	Log              logger.Logger

	// Pending operations management
	PendingOperations       map[string]*PendingOperation
	OperationsMutex         sync.RWMutex
	StopStreamChan          chan bool
	CurrentNotificationChan chan ToolCallNotification
	NotificationChanMutex   sync.Mutex

	// Custom function executor
	ExecuteFn func(string, string) (string, error)

	// Custom confirmation prompt function (for CLI mode)
	ConfirmationPromptFn func(string, string) tools.ConfirmationResponse
}

// NewBaseServerAgent creates a new base server agent
func NewBaseServerAgent(ctx context.Context, port string, chatAgent ChatAgent, executeFn func(string, string) (string, error)) *BaseServerAgent {
	agent := &BaseServerAgent{
		ChatAgent:         chatAgent,
		ToolsAgent:        nil,
		RagAgent:          nil,
		SimilarityLimit:   0.6,
		MaxSimilarities:   3,
		ContextSizeLimit:  8000,
		CompressorAgent:   nil,
		Port:              port,
		Ctx:               ctx,
		Log:               logger.GetLoggerFromEnv(),
		PendingOperations: make(map[string]*PendingOperation),
		StopStreamChan:    make(chan bool, 1),
	}

	// Set executeFunction (use provided or default)
	if executeFn != nil {
		agent.ExecuteFn = executeFn
	}

	return agent
}

// HandleHealth handles the health check endpoint
func (agent *BaseServerAgent) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		agent.Log.Error("Failed to encode health response: %v", err)
	}
}

// HandleMemoryReset handles the memory reset endpoint
func (agent *BaseServerAgent) HandleMemoryReset(w http.ResponseWriter, r *http.Request) {
	agent.ChatAgent.ResetMessages()
	if agent.ToolsAgent != nil {
		agent.ToolsAgent.ResetMessages()
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Memory reset successfully",
	}); err != nil {
		agent.Log.Error("Failed to encode memory reset response: %v", err)
	}
}

// HandleMessagesList handles the messages list endpoint
func (agent *BaseServerAgent) HandleMessagesList(w http.ResponseWriter, r *http.Request) {
	messages := agent.ChatAgent.GetMessages()

	w.Header().Set(headerContentType, contentTypeJSON)
	if err := json.NewEncoder(w).Encode(MemoryResponse{Messages: messages}); err != nil {
		agent.Log.Error("Failed to encode messages list response: %v", err)
	}
}

// HandleContextSize handles the tokens count endpoint
func (agent *BaseServerAgent) HandleContextSize(w http.ResponseWriter, r *http.Request) {
	count := len(agent.ChatAgent.GetMessages())
	charactersCount := agent.ChatAgent.GetContextSize()
	limit := agent.ContextSizeLimit

	w.Header().Set(headerContentType, contentTypeJSON)
	if err := json.NewEncoder(w).Encode(ContextSizeResponse{
		MessagesCount:   count,
		CharactersCount: charactersCount,
		Limit:           limit,
	}); err != nil {
		agent.Log.Error("Failed to encode tokens count response: %v", err)
	}
}

// HandleModelsInformation handles the models information endpoint
func (agent *BaseServerAgent) HandleModelsInformation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)

	toolsModel := "none"
	if agent.ToolsAgent != nil {
		toolsModel = agent.ToolsAgent.GetModelID()
	}
	embeddingsModel := "none"
	if agent.RagAgent != nil {
		embeddingsModel = agent.RagAgent.GetModelID()
	}

	if err := json.NewEncoder(w).Encode(map[string]any{
		"status":           "ok",
		"chat_model":       agent.ChatAgent.GetModelID(),
		"embeddings_model": embeddingsModel,
		"tools_model":      toolsModel,
	}); err != nil {
		agent.Log.Error("Failed to encode models information response: %v", err)
	}
}

// handleOperationSSE processes a single pending operation via SSE,
// sending notFoundMsg or successMsg then routing the response to the operation channel.
func (agent *BaseServerAgent) handleOperationSSE(
	w http.ResponseWriter,
	r *http.Request,
	notFoundMsg string,
	successMsg string,
	response tools.ConfirmationResponse,
) {
	var req OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flusher, ok := setSSEHeaders(w)
	if !ok {
		http.Error(w, errStreamingNotSupported, http.StatusInternalServerError)
		return
	}

	agent.OperationsMutex.Lock()
	op, exists := agent.PendingOperations[req.OperationID]
	if exists {
		delete(agent.PendingOperations, req.OperationID)
	}
	agent.OperationsMutex.Unlock()

	if !exists {
		jsonData, _ := json.Marshal(map[string]string{"message": fmt.Sprintf(notFoundMsg, req.OperationID)})
		if _, err := fmt.Fprintf(w, sseDataFmt, string(jsonData)); err != nil {
			agent.Log.Error("Failed to write operation not found response: %v", err)
		}
		flusher.Flush()
		return
	}

	jsonData, _ := json.Marshal(map[string]string{"message": fmt.Sprintf(successMsg, req.OperationID)})
	if _, err := fmt.Fprintf(w, sseDataFmt, string(jsonData)); err != nil {
		agent.Log.Error("Failed to write operation response: %v", err)
		return
	}
	flusher.Flush()
	op.Response <- response
}

// HandleOperationValidate handles the operation validation endpoint
func (agent *BaseServerAgent) HandleOperationValidate(w http.ResponseWriter, r *http.Request) {
	agent.handleOperationSSE(w, r,
		"❌ Operation %s not found",
		"✅ Operation %s validated<br>",
		tools.Confirmed,
	)
}

// HandleOperationCancel handles the operation cancellation endpoint
func (agent *BaseServerAgent) HandleOperationCancel(w http.ResponseWriter, r *http.Request) {
	agent.handleOperationSSE(w, r,
		"❌ Operation %s not found<br>",
		"⛔️ Operation %s cancelled<br>",
		tools.Denied,
	)
}

// HandleOperationReset handles the operation reset endpoint
func (agent *BaseServerAgent) HandleOperationReset(w http.ResponseWriter, r *http.Request) {
	flusher, ok := setSSEHeaders(w)
	if !ok {
		http.Error(w, errStreamingNotSupported, http.StatusInternalServerError)
		return
	}

	agent.OperationsMutex.Lock()
	count := len(agent.PendingOperations)
	for id, op := range agent.PendingOperations {
		op.Response <- tools.Quit
		delete(agent.PendingOperations, id)
	}
	agent.OperationsMutex.Unlock()

	jsonData, _ := json.Marshal(map[string]string{"message": fmt.Sprintf("🔄 All pending operations cancelled (%d operations)", count)})
	if _, err := fmt.Fprintf(w, sseDataFmt, string(jsonData)); err != nil {
		agent.Log.Error("Failed to write reset response: %v", err)
	}
	flusher.Flush()
}

// JSONEscape escapes a string for safe JSON embedding
func JSONEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// GetContext returns the base server agent's context
func (agent *BaseServerAgent) GetContext() context.Context {
	return agent.Ctx
}

// SetContext updates the base server agent's context
func (agent *BaseServerAgent) SetContext(ctx context.Context) {
	agent.Ctx = ctx
}
