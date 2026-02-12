package gatewayserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// GatewayServerAgent exposes an OpenAI-compatible API (POST /v1/chat/completions)
// backed by a N.O.V.A. crew of agents. External clients see a single "model".
type GatewayServerAgent struct {
	ctx context.Context
	log logger.Logger

	// Crew of chat agents
	chatAgents      map[string]*chat.Agent
	selectedAgentId string
	currentChatAgent *chat.Agent

	// Orchestration
	orchestratorAgent     agents.OrchestratorAgent
	matchAgentIdToTopicFn func(string, string) string

	// Tools - Server-side execution
	toolsAgent     *tools.Agent
	executeFn      func(string, string) (string, error)
	confirmationFn func(string, string) tools.ConfirmationResponse

	// Tools - Client-side execution
	clientSideToolsAgent *tools.Agent

	// Agent execution order
	agentExecutionOrder []AgentExecutionType

	// RAG
	ragAgent        *rag.Agent
	similarityLimit float64
	maxSimilarities int

	// Compression
	compressorAgent  *compressor.Agent
	contextSizeLimit int

	// Server
	port string
	Mux  *http.ServeMux

	// TLS/HTTPS configuration
	tlsCertData []byte
	tlsKeyData  []byte
	tlsCertPath string
	tlsKeyPath  string

	// Stream control
	stopStreamChan chan bool
	streamMutex    sync.Mutex

	// Lifecycle hooks
	beforeCompletion func(*GatewayServerAgent)
	afterCompletion  func(*GatewayServerAgent)
}

// GatewayServerAgentOption is a function that configures a GatewayServerAgent.
type GatewayServerAgentOption func(*GatewayServerAgent) error

// WithAgentCrew sets the crew of agents and the initially selected agent ID.
func WithAgentCrew(agentCrew map[string]*chat.Agent, selectedAgentId string) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		if len(agentCrew) == 0 {
			return fmt.Errorf("agent crew cannot be nil or empty")
		}
		if selectedAgentId == "" {
			return fmt.Errorf("selected agent ID cannot be empty")
		}
		firstAgent, exists := agentCrew[selectedAgentId]
		if !exists {
			return fmt.Errorf("selected agent ID %s does not exist in the provided crew", selectedAgentId)
		}
		agent.chatAgents = agentCrew
		agent.selectedAgentId = selectedAgentId
		agent.currentChatAgent = firstAgent
		return nil
	}
}

// WithSingleAgent creates a crew with a single agent (key: "single").
func WithSingleAgent(chatAgent *chat.Agent) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		if chatAgent == nil {
			return fmt.Errorf("chat agent cannot be nil")
		}
		agent.chatAgents = map[string]*chat.Agent{"single": chatAgent}
		agent.selectedAgentId = "single"
		agent.currentChatAgent = chatAgent
		return nil
	}
}

// WithPort sets the HTTP server port (default: 8080).
func WithPort(port int) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.port = fmt.Sprintf(":%d", port)
		return nil
	}
}

// WithToolsAgent attaches a tools agent for server-side function calling capabilities.
// Tools are executed on the server and results are used to continue the completion.
func WithToolsAgent(toolsAgent *tools.Agent) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.toolsAgent = toolsAgent
		return nil
	}
}

// WithClientSideToolsAgent attaches a tools agent for client-side tool execution.
// The gateway detects tool calls and returns them to the client in OpenAI format.
// The client executes the tools and sends results back as messages with role "tool".
// This is used by clients like qwen-code, aider, continue.dev, etc.
func WithClientSideToolsAgent(toolsAgent *tools.Agent) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.clientSideToolsAgent = toolsAgent
		return nil
	}
}

// WithAgentExecutionOrder sets the order in which different agent types process requests.
// The default order is: ClientSideTools -> ServerSideTools -> Orchestrator.
// Each handler in the order can either:
//  - Handle the request and return (stopping the chain)
//  - Skip and let the next handler process the request
//
// Example custom order:
//   WithAgentExecutionOrder([]AgentExecutionType{
//       AgentExecutionOrchestrator,      // Route first
//       AgentExecutionClientSideTools,   // Then check for client tools
//   })
func WithAgentExecutionOrder(order []AgentExecutionType) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		if len(order) == 0 {
			return fmt.Errorf("agent execution order cannot be empty")
		}
		agent.agentExecutionOrder = order
		return nil
	}
}

// WithExecuteFn sets the function executor for server-side tool execution.
func WithExecuteFn(fn func(string, string) (string, error)) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.executeFn = fn
		return nil
	}
}

// WithConfirmationPromptFn sets the confirmation prompt for tool call confirmation.
func WithConfirmationPromptFn(fn func(string, string) tools.ConfirmationResponse) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.confirmationFn = fn
		return nil
	}
}

// WithTLSCert sets the TLS certificate and key data for HTTPS support.
// When provided, the server will use HTTPS instead of HTTP.
// certData and keyData should be PEM-encoded certificate and private key.
func WithTLSCert(certData, keyData []byte) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
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
func WithTLSCertFromFile(certPath, keyPath string) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		if certPath == "" || keyPath == "" {
			return fmt.Errorf("TLS certificate and key file paths cannot be empty")
		}
		agent.tlsCertPath = certPath
		agent.tlsKeyPath = keyPath
		return nil
	}
}

// WithRagAgent attaches a RAG agent for document retrieval.
func WithRagAgent(ragAgent *rag.Agent) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.ragAgent = ragAgent
		return nil
	}
}

// WithRagAgentAndSimilarityConfig attaches a RAG agent and configures similarity settings.
func WithRagAgentAndSimilarityConfig(ragAgent *rag.Agent, similarityLimit float64, maxSimilarities int) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.ragAgent = ragAgent
		agent.similarityLimit = similarityLimit
		agent.maxSimilarities = maxSimilarities
		return nil
	}
}

// WithCompressorAgent attaches a compressor agent for context compression.
func WithCompressorAgent(compressorAgent *compressor.Agent) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.compressorAgent = compressorAgent
		return nil
	}
}

// WithCompressorAgentAndContextSize attaches a compressor agent and sets the context size limit.
func WithCompressorAgentAndContextSize(compressorAgent *compressor.Agent, contextSizeLimit int) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.compressorAgent = compressorAgent
		agent.contextSizeLimit = contextSizeLimit
		return nil
	}
}

// WithOrchestratorAgent attaches an orchestrator agent for topic detection and routing.
func WithOrchestratorAgent(orchestratorAgent agents.OrchestratorAgent) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.orchestratorAgent = orchestratorAgent
		return nil
	}
}

// WithMatchAgentIdToTopicFn sets the function to match agent ID to detected topic.
func WithMatchAgentIdToTopicFn(fn func(string, string) string) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.matchAgentIdToTopicFn = fn
		return nil
	}
}

// BeforeCompletion sets a hook called before each completion request.
func BeforeCompletion(fn func(*GatewayServerAgent)) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.beforeCompletion = fn
		return nil
	}
}

// AfterCompletion sets a hook called after each completion request.
func AfterCompletion(fn func(*GatewayServerAgent)) GatewayServerAgentOption {
	return func(agent *GatewayServerAgent) error {
		agent.afterCompletion = fn
		return nil
	}
}

// NewAgent creates a new GatewayServerAgent with the provided options.
//
// At least one of WithAgentCrew or WithSingleAgent must be provided.
//
// Available options:
//   - WithAgentCrew(agentCrew, selectedAgentId) - Sets the crew of agents
//   - WithSingleAgent(chatAgent) - Creates a single-agent crew
//   - WithPort(port) - Sets the HTTP server port (default: 8080)
//   - WithToolsAgent(toolsAgent) - Attaches a tools agent for server-side execution
//   - WithClientSideToolsAgent(toolsAgent) - Attaches a tools agent for client-side execution
//   - WithAgentExecutionOrder(order) - Sets the agent execution order
//   - WithExecuteFn(fn) - Sets the tool executor (for server-side tools)
//   - WithConfirmationPromptFn(fn) - Sets tool confirmation prompt
//   - WithTLSCert(certData, keyData) - Enables HTTPS with PEM-encoded certificate and key data
//   - WithTLSCertFromFile(certPath, keyPath) - Enables HTTPS with certificate and key files
//   - WithRagAgent(ragAgent) - Attaches a RAG agent
//   - WithRagAgentAndSimilarityConfig(ragAgent, limit, max) - RAG with similarity config
//   - WithCompressorAgent(compressorAgent) - Attaches a compressor agent
//   - WithCompressorAgentAndContextSize(compressorAgent, limit) - Compressor with size limit
//   - WithOrchestratorAgent(orchestratorAgent) - Attaches an orchestrator
//   - WithMatchAgentIdToTopicFn(fn) - Sets topic-to-agent routing
//   - BeforeCompletion(fn) - Hook before completion
//   - AfterCompletion(fn) - Hook after completion
func NewAgent(ctx context.Context, options ...GatewayServerAgentOption) (*GatewayServerAgent, error) {
	agent := &GatewayServerAgent{
		ctx:                 ctx,
		log:                 logger.GetLoggerFromEnv(),
		port:                ":8080",
		agentExecutionOrder: DefaultAgentExecutionOrder,
		similarityLimit:     0.6,
		maxSimilarities:     3,
		contextSizeLimit:    8000,
		stopStreamChan:      make(chan bool, 1),
	}

	for _, option := range options {
		if err := option(agent); err != nil {
			return nil, err
		}
	}

	if len(agent.chatAgents) == 0 {
		return nil, fmt.Errorf("agent crew must be set using WithAgentCrew or WithSingleAgent option")
	}

	// Default matchAgentIdToTopicFn: return first available agent
	if agent.matchAgentIdToTopicFn == nil {
		agent.matchAgentIdToTopicFn = func(currentAgent, topic string) string {
			return currentAgent
		}
	}

	// Note: No default executeFn - if not configured, toolsAgent will use its own configured callbacks

	agent.log.Info("🌐 GatewayServerAgent initialized (agent: %s)", agent.selectedAgentId)

	return agent, nil
}

// StartServer starts the HTTP server with OpenAI-compatible routes.
func (agent *GatewayServerAgent) StartServer() error {
	mux := http.NewServeMux()
	agent.Mux = mux

	// OpenAI-compatible routes
	mux.HandleFunc("POST /v1/chat/completions", agent.handleChatCompletions)
	mux.HandleFunc("GET /v1/models", agent.handleListModels)

	// Utility routes
	mux.HandleFunc("GET /health", agent.handleHealth)

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
			Addr:      agent.port,
			Handler:   handler,
			TLSConfig: tlsConfig,
		}
		agent.log.Info("🔒 Gateway server started on https://localhost%s", agent.port)
		agent.log.Info("📡 OpenAI-compatible endpoint: POST /v1/chat/completions")
		return server.ListenAndServeTLS("", "")
	} else if agent.tlsCertPath != "" && agent.tlsKeyPath != "" {
		// HTTPS mode with certificate files
		agent.log.Info("🔒 Gateway server started on https://localhost%s", agent.port)
		agent.log.Info("📡 OpenAI-compatible endpoint: POST /v1/chat/completions")
		return http.ListenAndServeTLS(agent.port, agent.tlsCertPath, agent.tlsKeyPath, handler)
	}

	// Default HTTP mode (backward compatible)
	agent.log.Info("🚀 Gateway server started on http://localhost%s", agent.port)
	agent.log.Info("📡 OpenAI-compatible endpoint: POST /v1/chat/completions")
	return http.ListenAndServe(agent.port, handler)
}

// corsMiddleware adds CORS headers to all responses.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- Accessor methods ---

// GetChatAgents returns the map of chat agents.
func (agent *GatewayServerAgent) GetChatAgents() map[string]*chat.Agent {
	return agent.chatAgents
}

// SetChatAgents sets the map of chat agents.
func (agent *GatewayServerAgent) SetChatAgents(chatAgents map[string]*chat.Agent) {
	agent.chatAgents = chatAgents
}

// AddChatAgentToCrew adds a new chat agent to the crew.
func (agent *GatewayServerAgent) AddChatAgentToCrew(id string, chatAgent *chat.Agent) error {
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

// RemoveChatAgentFromCrew removes a chat agent from the crew.
func (agent *GatewayServerAgent) RemoveChatAgentFromCrew(id string) error {
	if id == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}
	if _, exists := agent.chatAgents[id]; !exists {
		return fmt.Errorf("agent with ID %s does not exist in the crew", id)
	}
	if agent.currentChatAgent == agent.chatAgents[id] {
		return fmt.Errorf("cannot remove the currently active agent (ID: %s)", id)
	}
	delete(agent.chatAgents, id)
	return nil
}

// SetSelectedAgentId switches the active agent.
func (agent *GatewayServerAgent) SetSelectedAgentId(agentId string) error {
	chatAgent, exists := agent.chatAgents[agentId]
	if !exists {
		return fmt.Errorf("no chat agent found with ID: %s", agentId)
	}
	agent.selectedAgentId = agentId
	agent.currentChatAgent = chatAgent
	agent.log.Info("🔀 Switched to agent ID: %s", agentId)
	return nil
}

// GetSelectedAgentId returns the currently selected agent ID.
func (agent *GatewayServerAgent) GetSelectedAgentId() string {
	return agent.selectedAgentId
}

// Kind returns the agent type.
func (agent *GatewayServerAgent) Kind() agents.Kind {
	return agents.ChatServer
}

// GetName returns the current agent name.
func (agent *GatewayServerAgent) GetName() string {
	return agent.currentChatAgent.GetName()
}

// GetModelID returns the current model ID.
func (agent *GatewayServerAgent) GetModelID() string {
	return agent.currentChatAgent.GetModelID()
}

// GetMessages returns all conversation messages.
func (agent *GatewayServerAgent) GetMessages() []messages.Message {
	return agent.currentChatAgent.GetMessages()
}

// GetContextSize returns the approximate context size.
func (agent *GatewayServerAgent) GetContextSize() int {
	return agent.currentChatAgent.GetContextSize()
}

// StopStream interrupts the current streaming operation.
func (agent *GatewayServerAgent) StopStream() {
	agent.currentChatAgent.StopStream()
}

// ResetMessages clears all messages except the system instruction.
func (agent *GatewayServerAgent) ResetMessages() {
	agent.currentChatAgent.ResetMessages()
}

// AddMessage adds a message to the conversation history.
func (agent *GatewayServerAgent) AddMessage(role roles.Role, content string) {
	agent.currentChatAgent.AddMessage(role, content)
}

// GetPort returns the HTTP port.
func (agent *GatewayServerAgent) GetPort() string {
	return agent.port
}

// GetAgentExecutionOrder returns the current agent execution order.
func (agent *GatewayServerAgent) GetAgentExecutionOrder() []AgentExecutionType {
	return agent.agentExecutionOrder
}

// SetAgentExecutionOrder changes the agent execution order.
func (agent *GatewayServerAgent) SetAgentExecutionOrder(order []AgentExecutionType) {
	agent.agentExecutionOrder = order
}
