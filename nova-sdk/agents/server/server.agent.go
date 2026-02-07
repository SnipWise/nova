package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// ServerAgent wraps BaseServerAgent with chat-specific functionality
type ServerAgent struct {
	*serverbase.BaseServerAgent
	chatAgent *chat.Agent

	// HTTP server multiplexer for custom routes
	Mux *http.ServeMux

	// Temporary fields to store config before BaseServerAgent is created
	portConfig               string
	executeFnConfig          func(string, string) (string, error)
	confirmationPromptFnConfig func(string, string) tools.ConfirmationResponse
	toolsAgentConfig         *tools.Agent
	ragAgentConfig           *rag.Agent
	compressorAgentConfig    *compressor.Agent
	similarityLimitConfig    float64
	maxSimilaritiesConfig    int
	contextSizeLimitConfig   int

	// TLS/HTTPS configuration
	tlsCertData []byte
	tlsKeyData  []byte
	tlsCertPath string
	tlsKeyPath  string

	// Lifecycle hooks
	beforeCompletion func(*ServerAgent)
	afterCompletion  func(*ServerAgent)
}

// Re-export types from serverbase for backward compatibility
type (
	ToolCallNotification = serverbase.ToolCallNotification
	PendingOperation     = serverbase.PendingOperation
	CompletionRequest    = serverbase.CompletionRequest
	OperationRequest     = serverbase.OperationRequest
	MemoryResponse       = serverbase.MemoryResponse
	TokensResponse       = serverbase.ContextSizeResponse
)

// ServerAgentOption is a function that configures a ServerAgent
type ServerAgentOption func(*ServerAgent) error

// WithPort sets the HTTP server port
func WithPort(port int) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.portConfig = fmt.Sprintf(":%d", port)
		return nil
	}
}

// WithExecuteFn sets the custom function executor
func WithExecuteFn(fn func(string, string) (string, error)) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.executeFnConfig = fn
		return nil
	}
}

// WithConfirmationPromptFn sets the confirmation prompt function
func WithConfirmationPromptFn(fn func(string, string) tools.ConfirmationResponse) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.confirmationPromptFnConfig = fn
		return nil
	}
}

// WithTLSCert sets the TLS certificate and key data for HTTPS support.
// When provided, the server will use HTTPS instead of HTTP.
// certData and keyData should be PEM-encoded certificate and private key.
func WithTLSCert(certData, keyData []byte) ServerAgentOption {
	return func(agent *ServerAgent) error {
		if len(certData) == 0 || len(keyData) == 0 {
			return fmt.Errorf("TLS certificate and key data cannot be empty")
		}
		agent.tlsCertData = certData
		agent.tlsKeyData = keyData
		return nil
	}
}

// WithTLSCertFromFile sets the TLS certificate and key file paths for HTTPS support.
// When provided, the server will use HTTPS instead of HTTP.
// certPath and keyPath should point to PEM-encoded certificate and private key files.
func WithTLSCertFromFile(certPath, keyPath string) ServerAgentOption {
	return func(agent *ServerAgent) error {
		if certPath == "" || keyPath == "" {
			return fmt.Errorf("TLS certificate and key file paths cannot be empty")
		}
		agent.tlsCertPath = certPath
		agent.tlsKeyPath = keyPath
		return nil
	}
}

// WithToolsAgent sets the tools agent
func WithToolsAgent(toolsAgent *tools.Agent) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.toolsAgentConfig = toolsAgent
		return nil
	}
}

// WithCompressorAgent sets the compressor agent
func WithCompressorAgent(compressorAgent *compressor.Agent) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.compressorAgentConfig = compressorAgent
		return nil
	}
}

// WithCompressorAgentAndContextSize sets the compressor agent and context size limit
func WithCompressorAgentAndContextSize(compressorAgent *compressor.Agent, contextSizeLimit int) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.compressorAgentConfig = compressorAgent
		agent.contextSizeLimitConfig = contextSizeLimit
		return nil
	}
}

// WithRagAgent sets the RAG agent
func WithRagAgent(ragAgent *rag.Agent) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.ragAgentConfig = ragAgent
		return nil
	}
}

// WithRagAgentAndSimilarityConfig sets the RAG agent, similarity limit and max similarities
func WithRagAgentAndSimilarityConfig(ragAgent *rag.Agent, similarityLimit float64, maxSimilarities int) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.ragAgentConfig = ragAgent
		agent.similarityLimitConfig = similarityLimit
		agent.maxSimilaritiesConfig = maxSimilarities
		return nil
	}
}

// BeforeCompletion sets a hook that is called before each completion (HTTP and CLI)
func BeforeCompletion(fn func(*ServerAgent)) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.beforeCompletion = fn
		return nil
	}
}

// AfterCompletion sets a hook that is called after each completion (HTTP and CLI)
func AfterCompletion(fn func(*ServerAgent)) ServerAgentOption {
	return func(agent *ServerAgent) error {
		agent.afterCompletion = fn
		return nil
	}
}

// NewAgent creates a new server agent with options
//
// Available options:
//   - WithPort(port) - Sets the HTTP server port as int (default: 8080)
//   - WithExecuteFn(fn) - Sets the custom function executor for tool execution
//   - WithConfirmationPromptFn(fn) - Sets the confirmation prompt function for human-in-the-loop
//   - WithTLSCert(certData, keyData) - Enables HTTPS with PEM-encoded certificate and key data
//   - WithTLSCertFromFile(certPath, keyPath) - Enables HTTPS with certificate and key files
//   - WithToolsAgent(toolsAgent) - Attaches a tools agent for function calling capabilities
//   - WithCompressorAgent(compressorAgent) - Attaches a compressor agent for context compression
//   - WithCompressorAgentAndContextSize(compressorAgent, contextSizeLimit) - Attaches a compressor agent and sets the context size limit
//   - WithRagAgent(ragAgent) - Attaches a RAG agent for document retrieval
//   - WithRagAgentAndSimilarityConfig(ragAgent, similarityLimit, maxSimilarities) - Attaches a RAG agent and configures similarity settings
//   - BeforeCompletion(fn) - Sets a hook called before each completion (HTTP and CLI)
//   - AfterCompletion(fn) - Sets a hook called after each completion (HTTP and CLI)
//
// Example:
//   agent, err := NewAgent(ctx, agentConfig, modelConfig,
//       WithPort(8080),
//       WithToolsAgent(toolsAgent),
//       WithRagAgent(ragAgent),
//   )
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
	options ...ServerAgentOption,
) (*ServerAgent, error) {
	chatAgent, err := chat.NewAgent(ctx, agentConfig, modelConfig)
	if err != nil {
		return nil, err
	}

	// Create agent with defaults
	agent := &ServerAgent{
		chatAgent:  chatAgent,
		portConfig: ":8080", // Default port
	}

	// Apply all options
	for _, option := range options {
		if err := option(agent); err != nil {
			return nil, err
		}
	}

	// Create base server agent with the configured port and executeFn
	baseAgent := serverbase.NewBaseServerAgent(ctx, agent.portConfig, chatAgent, agent.executeFnConfig)
	agent.BaseServerAgent = baseAgent

	// Apply configuration from temporary fields to BaseServerAgent
	if agent.toolsAgentConfig != nil {
		agent.ToolsAgent = agent.toolsAgentConfig
	}
	if agent.ragAgentConfig != nil {
		agent.RagAgent = agent.ragAgentConfig
		if agent.similarityLimitConfig != 0 {
			agent.SimilarityLimit = agent.similarityLimitConfig
		}
		if agent.maxSimilaritiesConfig != 0 {
			agent.MaxSimilarities = agent.maxSimilaritiesConfig
		}
	}
	if agent.compressorAgentConfig != nil {
		agent.CompressorAgent = agent.compressorAgentConfig
		if agent.contextSizeLimitConfig != 0 {
			agent.ContextSizeLimit = agent.contextSizeLimitConfig
		}
	}

	// Set executeFunction to default if not provided
	if agent.ExecuteFn == nil {
		agent.ExecuteFn = agent.executeFunction
	}

	// Set confirmationPromptFn to provided or default CLI confirmation
	if agent.confirmationPromptFnConfig != nil {
		agent.ConfirmationPromptFn = agent.confirmationPromptFnConfig
	} else {
		agent.ConfirmationPromptFn = agent.cliConfirmationPrompt
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

// SetConfirmationPromptFn sets the confirmation prompt function for CLI mode
func (agent *ServerAgent) SetConfirmationPromptFn(fn func(string, string) tools.ConfirmationResponse) {
	agent.ConfirmationPromptFn = fn
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

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins (can be restricted to specific origins if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Allowed HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Allowed headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")

		// Allow credentials (cookies, auth)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Preflight request (OPTIONS) - return 200 OK without calling next handler
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}

// StartServer starts the HTTP server with all routes
func (agent *ServerAgent) StartServer() error {
	mux := http.NewServeMux()

	// Expose mux for custom routes
	agent.Mux = mux

	// Routes using base handlers
	mux.HandleFunc("POST /completion", agent.handleCompletion)
	mux.HandleFunc("POST /completion/stop", agent.handleCompletionStop)
	mux.HandleFunc("POST /memory/reset", agent.HandleMemoryReset)
	mux.HandleFunc("GET /memory/messages/list", agent.HandleMessagesList)
	mux.HandleFunc("GET /memory/messages/context-size", agent.HandleContextSize)
	mux.HandleFunc("POST /operation/validate", agent.HandleOperationValidate)
	mux.HandleFunc("POST /operation/cancel", agent.HandleOperationCancel)
	mux.HandleFunc("POST /operation/reset", agent.HandleOperationReset)
	mux.HandleFunc("GET /models", agent.HandleModelsInformation)
	mux.HandleFunc("GET /health", agent.HandleHealth)

	// Apply CORS middleware
	handler := corsMiddleware(mux)

	// Check if TLS is configured
	if len(agent.tlsCertData) > 0 && len(agent.tlsKeyData) > 0 {
		// HTTPS mode with certificate data in memory
		cert, err := tls.X509KeyPair(agent.tlsCertData, agent.tlsKeyData)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		server := &http.Server{
			Addr:      agent.Port,
			Handler:   handler,
			TLSConfig: tlsConfig,
		}
		agent.Log.Info("🔒 Server started on https://localhost%s", agent.Port)
		return server.ListenAndServeTLS("", "")
	} else if agent.tlsCertPath != "" && agent.tlsKeyPath != "" {
		// HTTPS mode with certificate files
		agent.Log.Info("🔒 Server started on https://localhost%s", agent.Port)
		return http.ListenAndServeTLS(agent.Port, agent.tlsCertPath, agent.tlsKeyPath, handler)
	}

	// Default HTTP mode (backward compatible)
	agent.Log.Info("🚀 Server started on http://localhost%s", agent.Port)
	return http.ListenAndServe(agent.Port, handler)
}

// Helper functions

// jsonEscape escapes a string for safe JSON embedding
func jsonEscape(s string) string {
	return serverbase.JSONEscape(s)
}
