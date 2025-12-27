package server

import (
	"context"
	"net/http"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// ServerAgent wraps BaseServerAgent with chat-specific functionality
type ServerAgent struct {
	*serverbase.BaseServerAgent
	chatAgent *chat.Agent
}

// Re-export types from serverbase for backward compatibility
type (
	ToolCallNotification = serverbase.ToolCallNotification
	PendingOperation     = serverbase.PendingOperation
	CompletionRequest    = serverbase.CompletionRequest
	OperationRequest     = serverbase.OperationRequest
	MemoryResponse       = serverbase.MemoryResponse
	TokensResponse       = serverbase.TokensResponse
)

// NewAgent creates a new server agent
// executeFn is optional - if not provided, uses the default executeFunction method
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
	port string,
	executeFn ...func(string, string) (string, error),
) (*ServerAgent, error) {
	chatAgent, err := chat.NewAgent(ctx, agentConfig, modelConfig)
	if err != nil {
		return nil, err
	}

	// Use provided executeFn or nil (will be set to default later)
	var execFn func(string, string) (string, error)
	if len(executeFn) > 0 && executeFn[0] != nil {
		execFn = executeFn[0]
	}

	baseAgent := serverbase.NewBaseServerAgent(ctx, port, chatAgent, execFn)

	agent := &ServerAgent{
		BaseServerAgent: baseAgent,
		chatAgent:       chatAgent,
	}

	// Set executeFunction to default if not provided
	if execFn == nil {
		agent.ExecuteFn = agent.executeFunction
	}

	return agent, nil
}

// SetPort sets the HTTP port
func (agent *ServerAgent) SetPort(port string) {
	agent.Port = port
}

// GetPort returns the HTTP port
func (agent *ServerAgent) GetPort() string {
	return agent.Port
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

// Note: SetToolsAgent, GetToolsAgent, SetRagAgent, GetRagAgent, etc.
// are defined in methods.*.related.go files

// StartServer starts the HTTP server with all routes
func (agent *ServerAgent) StartServer() error {
	mux := http.NewServeMux()

	// Routes using base handlers
	mux.HandleFunc("POST /completion", agent.handleCompletion)
	mux.HandleFunc("POST /completion/stop", agent.handleCompletionStop)
	mux.HandleFunc("POST /memory/reset", agent.HandleMemoryReset)
	mux.HandleFunc("GET /memory/messages/list", agent.HandleMessagesList)
	mux.HandleFunc("GET /memory/messages/tokens", agent.HandleTokensCount)
	mux.HandleFunc("POST /operation/validate", agent.HandleOperationValidate)
	mux.HandleFunc("POST /operation/cancel", agent.HandleOperationCancel)
	mux.HandleFunc("POST /operation/reset", agent.HandleOperationReset)
	mux.HandleFunc("GET /models", agent.HandleModelsInformation)
	mux.HandleFunc("GET /health", agent.HandleHealth)

	agent.Log.Info("🚀 Server started on http://localhost%s", agent.Port)
	return http.ListenAndServe(agent.Port, mux)
}

// Helper functions

// jsonEscape escapes a string for safe JSON embedding
func jsonEscape(s string) string {
	return serverbase.JSONEscape(s)
}
