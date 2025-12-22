package crewserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)


type CrewServerAgent struct {

	// Multiple chat agents for different purposes
	// The crewserver can route between them based on the task
	chatAgents map[string]*chat.Agent

	currentChatAgent *chat.Agent
	toolsAgent       *tools.Agent

	ragAgent *rag.Agent

	similarityLimit float64
	maxSimilarities int

	contextSizeLimit int
	compressorAgent  *compressor.Agent

	// Routing / Orchestration agent
	orchestratorAgent *structured.Agent[agents.Intent]

	// Retrieval function: from a topic determine which agent to use
	matchAgentIdToTopicFn func(string) string

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
	agentCrew map[string]*chat.Agent,
	selectedAgentId string,
	port string,
	matchAgentIdToTopicFn func(string) string,
	executeFn func(string, string) (string, error),
) (*CrewServerAgent, error) {

	firstSelectedAgent, exists := agentCrew[selectedAgentId]
	if !exists {
		return nil, fmt.Errorf("selected agent ID %s does not exist in the provided crew", selectedAgentId)
	}

	agent := &CrewServerAgent{
		chatAgents:        agentCrew,
		currentChatAgent:  firstSelectedAgent,
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

	agent.log.Info("👥 CrewServerAgent initialized with chat agents, starting with agent ID: %s", selectedAgentId)

	// Set matchAgentIdToTopicFn
	if matchAgentIdToTopicFn != nil {
		agent.matchAgentIdToTopicFn = matchAgentIdToTopicFn
	} else {
		// Default function: return the first agent ID in the map ignoring the topic if no function is provided
		agent.matchAgentIdToTopicFn = func(topic string) string {
			var agentId string
			for key := range agent.chatAgents {
				agentId = key
				break
			}
			return agentId
		}
	}

	// Set executeFunction (use provided or default)
	if executeFn != nil {
		agent.executeFn = executeFn
	} else {
		// executeFunction is a placeholder that should be overridden by the user
		agent.executeFn = agent.executeFunction
	}

	return agent, nil
}

// SetPort sets the HTTP port
func (agent *CrewServerAgent) SetPort(port string) {
	agent.port = port
}

// GetPort returns the HTTP port
func (agent *CrewServerAgent) GetPort() string {
	return agent.port
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

// StartServer starts the HTTP server with all routes
func (agent *CrewServerAgent) StartServer() error {
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
