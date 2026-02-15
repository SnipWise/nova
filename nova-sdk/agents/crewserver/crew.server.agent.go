package crewserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
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
)

// CrewServerAgent wraps BaseServerAgent with crew-specific functionality
type CrewServerAgent struct {
	*serverbase.BaseServerAgent

	// Multiple chat agents for different purposes
	// The crewserver can route between them based on the task
	chatAgents      map[string]*chat.Agent
	selectedAgentId string

	currentChatAgent *chat.Agent

	// Routing / Orchestration agent
	orchestratorAgent agents.OrchestratorAgent

	// Retrieval function: from a topic determine which agent to use
	matchAgentIdToTopicFn func(string, string) string

	// HTTP server multiplexer for custom routes
	Mux *http.ServeMux

	// Temporary fields to store config before BaseServerAgent is created
	portConfig             string
	executeFnConfig        func(string, string) (string, error)
	toolsAgentConfig       *tools.Agent
	ragAgentConfig         *rag.Agent
	compressorAgentConfig  *compressor.Agent
	similarityLimitConfig  float64
	maxSimilaritiesConfig  int
	contextSizeLimitConfig int

	// TLS/HTTPS configuration
	tlsCertData []byte
	tlsKeyData  []byte
	tlsCertPath string
	tlsKeyPath  string

	// Confirmation prompt function config (for tool call confirmation)
	confirmationPromptFnConfig func(string, string) tools.ConfirmationResponse

	// Lifecycle hooks
	beforeCompletion func(*CrewServerAgent)
	afterCompletion  func(*CrewServerAgent)
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

// CrewServerAgentOption is a function that configures a CrewServerAgent
type CrewServerAgentOption func(*CrewServerAgent) error

// WithAgentCrew sets the crew of agents and the selected agent ID
func WithAgentCrew(agentCrew map[string]*chat.Agent, selectedAgentId string) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		if agentCrew == nil || len(agentCrew) == 0 {
			return fmt.Errorf("agent crew cannot be nil or empty")
		}
		if selectedAgentId == "" {
			return fmt.Errorf("selected agent ID cannot be empty")
		}
		firstSelectedAgent, exists := agentCrew[selectedAgentId]
		if !exists {
			return fmt.Errorf("selected agent ID %s does not exist in the provided crew", selectedAgentId)
		}
		agent.chatAgents = agentCrew
		agent.selectedAgentId = selectedAgentId
		agent.currentChatAgent = firstSelectedAgent
		return nil
	}
}

// WithSingleAgent creates a crew with a single agent with the key "single"
func WithSingleAgent(chatAgent *chat.Agent) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		if chatAgent == nil {
			return fmt.Errorf("chat agent cannot be nil")
		}
		agent.chatAgents = map[string]*chat.Agent{"single": chatAgent}
		agent.selectedAgentId = "single"
		agent.currentChatAgent = chatAgent
		return nil
	}
}

// WithPort sets the HTTP server port (default: 3500)
func WithPort(port int) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.portConfig = fmt.Sprintf(":%d", port)
		return nil
	}
}

// WithMatchAgentIdToTopicFn sets the function to match agent ID to topic
func WithMatchAgentIdToTopicFn(fn func(string, string) string) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.matchAgentIdToTopicFn = fn
		return nil
	}
}

// WithExecuteFn sets the custom function executor
func WithExecuteFn(fn func(string, string) (string, error)) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.executeFnConfig = fn
		return nil
	}
}

// WithToolsAgent sets the tools agent
func WithToolsAgent(toolsAgent *tools.Agent) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.toolsAgentConfig = toolsAgent
		return nil
	}
}

// WithCompressorAgent sets the compressor agent
func WithCompressorAgent(compressorAgent *compressor.Agent) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.compressorAgentConfig = compressorAgent
		return nil
	}
}

// WithCompressorAgentAndContextSize sets the compressor agent and context size limit
func WithCompressorAgentAndContextSize(compressorAgent *compressor.Agent, contextSizeLimit int) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.compressorAgentConfig = compressorAgent
		agent.contextSizeLimitConfig = contextSizeLimit
		return nil
	}
}

// WithRagAgent sets the RAG agent
func WithRagAgent(ragAgent *rag.Agent) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.ragAgentConfig = ragAgent
		return nil
	}
}

// WithRagAgentAndSimilarityConfig sets the RAG agent, similarity limit and max similarities
func WithRagAgentAndSimilarityConfig(ragAgent *rag.Agent, similarityLimit float64, maxSimilarities int) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.ragAgentConfig = ragAgent
		agent.similarityLimitConfig = similarityLimit
		agent.maxSimilaritiesConfig = maxSimilarities
		return nil
	}
}

// WithConfirmationPromptFn sets the confirmation prompt function for tool call confirmation.
// When provided, this function is used instead of the default web-based confirmation prompt.
// When combined with ParallelToolCalls enabled on the tools agent, DetectParallelToolCallsWithConfirmation is used.
func WithConfirmationPromptFn(fn func(string, string) tools.ConfirmationResponse) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.confirmationPromptFnConfig = fn
		return nil
	}
}

// WithTLSCert sets the TLS certificate and key data for HTTPS support.
// When provided, the server will use HTTPS instead of HTTP.
// certData and keyData should be PEM-encoded certificate and private key.
func WithTLSCert(certData, keyData []byte) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
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
func WithTLSCertFromFile(certPath, keyPath string) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		if certPath == "" || keyPath == "" {
			return fmt.Errorf("TLS certificate and key file paths cannot be empty")
		}
		agent.tlsCertPath = certPath
		agent.tlsKeyPath = keyPath
		return nil
	}
}

// WithOrchestratorAgent sets the orchestrator agent for routing/topic detection
// Automatically configures matchAgentIdToTopicFn to use the orchestrator's GetAgentForTopic method
// unless explicitly overridden with WithMatchAgentIdToTopicFn.
func WithOrchestratorAgent(orchestratorAgent agents.OrchestratorAgent) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.orchestratorAgent = orchestratorAgent

		// Auto-configure routing function using the orchestrator's GetAgentForTopic method
		// This can still be overridden by calling WithMatchAgentIdToTopicFn after this
		if agent.matchAgentIdToTopicFn == nil {
			agent.matchAgentIdToTopicFn = func(currentAgentId, topic string) string {
				return orchestratorAgent.GetAgentForTopic(topic)
			}
		}

		return nil
	}
}

// BeforeCompletion sets a hook that is called before each handleCompletion call
func BeforeCompletion(fn func(*CrewServerAgent)) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.beforeCompletion = fn
		return nil
	}
}

// AfterCompletion sets a hook that is called after each handleCompletion call
func AfterCompletion(fn func(*CrewServerAgent)) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.afterCompletion = fn
		return nil
	}
}

// NewAgent creates a new crew server agent with options
//
// Available options:
//   - WithAgentCrew(agentCrew, selectedAgentId) - Sets the crew of agents and the selected agent ID
//   - WithSingleAgent(chatAgent) - Creates a crew with a single agent (key: "single", selectedAgentId: "single")
//   - WithPort(port) - Sets the HTTP server port as int (default: 3500)
//   - WithMatchAgentIdToTopicFn(fn) - Sets the function to match agent ID to topic for routing
//   - WithExecuteFn(fn) - Sets the custom function executor for tool execution
//   - WithToolsAgent(toolsAgent) - Attaches a tools agent for function calling capabilities
//   - WithCompressorAgent(compressorAgent) - Attaches a compressor agent for context compression
//   - WithCompressorAgentAndContextSize(compressorAgent, contextSizeLimit) - Attaches a compressor agent and sets the context size limit
//   - WithRagAgent(ragAgent) - Attaches a RAG agent for document retrieval
//   - WithRagAgentAndSimilarityConfig(ragAgent, similarityLimit, maxSimilarities) - Attaches a RAG agent and configures similarity settings
//   - WithConfirmationPromptFn(fn) - Sets a custom confirmation prompt function for tool call confirmation
//   - WithTLSCert(certData, keyData) - Enables HTTPS with PEM-encoded certificate and key data
//   - WithTLSCertFromFile(certPath, keyPath) - Enables HTTPS with certificate and key files
//   - WithOrchestratorAgent(orchestratorAgent) - Attaches an orchestrator agent for routing/topic detection
//   - BeforeCompletion(fn) - Sets a hook called before each handleCompletion call
//   - AfterCompletion(fn) - Sets a hook called after each handleCompletion call
//
// At least one of WithAgentCrew or WithSingleAgent must be provided.
func NewAgent(ctx context.Context, options ...CrewServerAgentOption) (*CrewServerAgent, error) {
	// Create agent with defaults
	agent := &CrewServerAgent{
		chatAgents:       nil,
		currentChatAgent: nil,
		selectedAgentId:  "",
		portConfig:       ":3500", // Default port
	}

	// Apply all options
	for _, option := range options {
		if err := option(agent); err != nil {
			return nil, err
		}
	}

	// Validate that agent crew was set
	if agent.chatAgents == nil || len(agent.chatAgents) == 0 {
		return nil, fmt.Errorf("agent crew must be set using WithAgentCrew or WithSingleAgent option")
	}

	// Create base server agent with the configured port and executeFn
	baseAgent := serverbase.NewBaseServerAgent(ctx, agent.portConfig, agent.currentChatAgent, agent.executeFnConfig)
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

	// Set confirmationPromptFn if provided
	if agent.confirmationPromptFnConfig != nil {
		agent.ConfirmationPromptFn = agent.confirmationPromptFnConfig
	}

	// Set default matchAgentIdToTopicFn if not provided
	if agent.matchAgentIdToTopicFn == nil {
		agent.matchAgentIdToTopicFn = func(currentAgent, topic string) string {
			var agentId string
			for key := range agent.chatAgents {
				agentId = key
				break
			}
			return agentId
		}
	}

	// Set default executeFn if not provided
	if agent.ExecuteFn == nil {
		agent.ExecuteFn = agent.executeFunction
	}

	agent.Log.Info("👥 CrewServerAgent initialized with chat agents, starting with agent ID: %s", agent.selectedAgentId)

	return agent, nil
}

// SetPort sets the HTTP port
func (agent *CrewServerAgent) SetPort(port string) {
	agent.Port = port
}

// GetPort returns the HTTP port
func (agent *CrewServerAgent) GetPort() string {
	return agent.Port
}

// GetChatAgents returns the map of chat agents
func (agent *CrewServerAgent) GetChatAgents() map[string]*chat.Agent {
	return agent.chatAgents
}

// SetChatAgents sets the map of chat agents
func (agent *CrewServerAgent) SetChatAgents(chatAgents map[string]*chat.Agent) {
	agent.chatAgents = chatAgents
}

// AddChatAgentToCrew adds a new chat agent to the crew
func (agent *CrewServerAgent) AddChatAgentToCrew(id string, chatAgent *chat.Agent) error {
	if chatAgent == nil {
		return fmt.Errorf("cannot add nil chat agent")
	}
	if id == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}
	if _, exists := agent.chatAgents[id]; exists {
		return fmt.Errorf("agent with ID %s already exists in the crew", id)
	}
	agent.chatAgents[id] = chatAgent
	return nil
}

// RemoveChatAgentFromCrew removes a chat agent from the crew
func (agent *CrewServerAgent) RemoveChatAgentFromCrew(id string) error {
	if id == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}
	if _, exists := agent.chatAgents[id]; !exists {
		return fmt.Errorf("agent with ID %s does not exist in the crew", id)
	}
	// Prevent removing the current active agent
	if agent.currentChatAgent == agent.chatAgents[id] {
		return fmt.Errorf("cannot remove the currently active agent (ID: %s)", id)
	}
	delete(agent.chatAgents, id)
	return nil
}

// Kind returns the agent type
func (agent *CrewServerAgent) Kind() agents.Kind {
	return agents.ChatServer
}

// GetName returns the agent name
func (agent *CrewServerAgent) GetName() string {
	return agent.currentChatAgent.GetName()
}

// GetModelID returns the model ID
func (agent *CrewServerAgent) GetModelID() string {
	return agent.currentChatAgent.GetModelID()
}

// GetMessages returns all conversation messages
func (agent *CrewServerAgent) GetMessages() []messages.Message {
	return agent.currentChatAgent.GetMessages()
}

// GetContextSize returns the approximate size of the current context
func (agent *CrewServerAgent) GetContextSize() int {
	return agent.currentChatAgent.GetContextSize()
}

// StopStream interrupts the current streaming operation
func (agent *CrewServerAgent) StopStream() {
	agent.currentChatAgent.StopStream()
}

// ResetMessages clears all messages except the system instruction
func (agent *CrewServerAgent) ResetMessages() {
	agent.currentChatAgent.ResetMessages()
}

// AddMessage adds a message to the conversation history
func (agent *CrewServerAgent) AddMessage(role roles.Role, content string) {
	agent.currentChatAgent.AddMessage(role, content)
}

// GenerateCompletion sends messages and returns the completion result
func (agent *CrewServerAgent) GenerateCompletion(userMessages []messages.Message) (*chat.CompletionResult, error) {
	return agent.currentChatAgent.GenerateCompletion(userMessages)
}

// GenerateCompletionWithReasoning sends messages and returns the completion result with reasoning
func (agent *CrewServerAgent) GenerateCompletionWithReasoning(userMessages []messages.Message) (*chat.ReasoningResult, error) {
	return agent.currentChatAgent.GenerateCompletionWithReasoning(userMessages)
}

// GenerateStreamCompletion sends messages and streams the response via callback
func (agent *CrewServerAgent) GenerateStreamCompletion(
	userMessages []messages.Message,
	callback chat.StreamCallback,
) (*chat.CompletionResult, error) {
	return agent.currentChatAgent.GenerateStreamCompletion(userMessages, callback)
}

// GenerateStreamCompletionWithReasoning sends messages and streams both reasoning and response
func (agent *CrewServerAgent) GenerateStreamCompletionWithReasoning(
	userMessages []messages.Message,
	reasoningCallback chat.StreamCallback,
	responseCallback chat.StreamCallback,
) (*chat.ReasoningResult, error) {
	return agent.currentChatAgent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCallback, responseCallback)
}

// ExportMessagesToJSON exports the conversation history to JSON
func (agent *CrewServerAgent) ExportMessagesToJSON() (string, error) {
	return agent.currentChatAgent.ExportMessagesToJSON()
}

func (agent *CrewServerAgent) SetSelectedAgentId(agentId string) error {
	chatAgent, exists := agent.chatAgents[agentId]
	if !exists {
		return fmt.Errorf("no chat agent found with ID: %s", agentId)
	}
	agent.selectedAgentId = agentId
	agent.currentChatAgent = chatAgent
	agent.Log.Info("🔀 Switched to agent ID: %s", agentId)
	return nil
}

// GetSelectedAgentId returns the currently selected agent ID
func (agent *CrewServerAgent) GetSelectedAgentId() string {
	return agent.selectedAgentId
}

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
func (agent *CrewServerAgent) StartServer() error {
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
	mux.HandleFunc("GET /current-agent", agent.handleCurrentAgent)

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

// handleCurrentAgent returns information about the currently selected agent
func (agent *CrewServerAgent) handleCurrentAgent(w http.ResponseWriter, r *http.Request) {
	agentId := agent.GetSelectedAgentId()
	modelId := agent.GetModelID()
	agentName := agent.GetName()

	if agentName == "" {
		agentName = "Unnamed Agent"
	}

	//agent.Log.Info("ℹ️ Current agent requested: ID=%s, Name=%s, Model=%s", agentId, agentName, modelId)

	response := map[string]string{
		"agent_id": agentId,
		"model_id": modelId,
		"agent_name": agentName,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Helper functions

// jsonEscape escapes a string for safe JSON embedding
func jsonEscape(s string) string {
	return serverbase.JSONEscape(s)
}
