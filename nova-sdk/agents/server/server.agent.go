package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

type ServerAgent struct {
	chatAgent  *chat.Agent
	toolsAgent *tools.Agent

	ragAgent *rag.Agent

	similarityLimit float64
	maxSimilarities int

	contextSizeLimit int
	compressorAgent  *compressor.Agent

	port string
	ctx  context.Context
	log  logger.Logger

	// Pending operations management
	pendingOperations       map[string]*PendingOperation
	operationsMutex         sync.RWMutex
	stopStreamChan          chan bool
	currentNotificationChan chan ToolCallNotification
	notificationChanMutex   sync.Mutex

	// Custom function executor
	executeFn func(string, string) (string, error)
}

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

type MemoryResponse struct {
	Messages []messages.Message `json:"messages"`
}

type TokensResponse struct {
	Count  int `json:"count"`
	Tokens int `json:"tokens"`
	Limit  int `json:"limit"`
}

// NewAgent creates a new server agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
	port string,
	executeFn func(string, string) (string, error),
) (*ServerAgent, error) {
	chatAgent, err := chat.NewAgent(ctx, agentConfig, modelConfig)
	if err != nil {
		return nil, err
	}

	agent := &ServerAgent{
		chatAgent:         chatAgent,
		toolsAgent:        nil,
		ragAgent:          nil,
		similarityLimit:   0.6,
		maxSimilarities:   3,
		contextSizeLimit:  8000,
		compressorAgent:   nil,
		port:              port,
		ctx:               ctx,
		log:               logger.GetLoggerFromEnv(),
		pendingOperations: make(map[string]*PendingOperation),
		stopStreamChan:    make(chan bool, 1),
	}

	// Set executeFunction (use provided or default)
	if executeFn != nil {
		agent.executeFn = executeFn
	} else {
		agent.executeFn = agent.executeFunction
	}

	return agent, nil
}

// SetPort sets the HTTP port
func (agent *ServerAgent) SetPort(port string) {
	agent.port = port
}

// GetPort returns the HTTP port
func (agent *ServerAgent) GetPort() string {
	return agent.port
}

// Kind returns the agent type
func (agent *ServerAgent) Kind() agents.Kind {
	return agents.ChatServer
}

// GetName returns the agent name
func (agent *ServerAgent) GetName() string {
	return agent.chatAgent.GetName()
}

// GetModelID returns the model ID
func (agent *ServerAgent) GetModelID() string {
	return agent.chatAgent.GetModelID()
}

// GetMessages returns all conversation messages
func (agent *ServerAgent) GetMessages() []messages.Message {
	return agent.chatAgent.GetMessages()
}

// GetContextSize returns the approximate size of the current context
func (agent *ServerAgent) GetContextSize() int {
	return agent.chatAgent.GetContextSize()
}

// StopStream interrupts the current streaming operation
func (agent *ServerAgent) StopStream() {
	agent.chatAgent.StopStream()
}

// ResetMessages clears all messages except the system instruction
func (agent *ServerAgent) ResetMessages() {
	agent.chatAgent.ResetMessages()
}

// AddMessage adds a message to the conversation history
func (agent *ServerAgent) AddMessage(role roles.Role, content string) {
	agent.chatAgent.AddMessage(role, content)
}

// GenerateCompletion sends messages and returns the completion result
func (agent *ServerAgent) GenerateCompletion(userMessages []messages.Message) (*chat.CompletionResult, error) {
	return agent.chatAgent.GenerateCompletion(userMessages)
}

// GenerateCompletionWithReasoning sends messages and returns the completion result with reasoning
func (agent *ServerAgent) GenerateCompletionWithReasoning(userMessages []messages.Message) (*chat.ReasoningResult, error) {
	return agent.chatAgent.GenerateCompletionWithReasoning(userMessages)
}

// GenerateStreamCompletion sends messages and streams the response via callback
func (agent *ServerAgent) GenerateStreamCompletion(
	userMessages []messages.Message,
	callback chat.StreamCallback,
) (*chat.CompletionResult, error) {
	return agent.chatAgent.GenerateStreamCompletion(userMessages, callback)
}

// GenerateStreamCompletionWithReasoning sends messages and streams both reasoning and response
func (agent *ServerAgent) GenerateStreamCompletionWithReasoning(
	userMessages []messages.Message,
	reasoningCallback chat.StreamCallback,
	responseCallback chat.StreamCallback,
) (*chat.ReasoningResult, error) {
	return agent.chatAgent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCallback, responseCallback)
}

// ExportMessagesToJSON exports the conversation history to JSON
func (agent *ServerAgent) ExportMessagesToJSON() (string, error) {
	return agent.chatAgent.ExportMessagesToJSON()
}

// StartServer starts the HTTP server with all routes
func (agent *ServerAgent) StartServer() error {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("POST /completion", agent.handleCompletion)
	mux.HandleFunc("POST /completion/stop", agent.handleCompletionStop)
	mux.HandleFunc("POST /memory/reset", agent.handleMemoryReset)
	mux.HandleFunc("GET /memory/messages/list", agent.handleMessagesList)
	mux.HandleFunc("GET /memory/messages/tokens", agent.handleTokensCount)
	mux.HandleFunc("POST /operation/validate", agent.handleOperationValidate)
	mux.HandleFunc("POST /operation/cancel", agent.handleOperationCancel)
	mux.HandleFunc("POST /operation/reset", agent.handleOperationReset)
	mux.HandleFunc("GET /models", agent.handleModelsInformation)
	mux.HandleFunc("GET /health", agent.handleHealth)

	agent.log.Info("🚀 Server started on http://localhost%s", agent.port)
	return http.ListenAndServe(agent.port, mux)
}

// Helper functions

// jsonEscape escapes a string for safe JSON embedding
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
