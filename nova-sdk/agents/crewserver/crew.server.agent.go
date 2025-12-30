package crewserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
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

// NewAgent creates a new crew server agent
func NewAgent(
	ctx context.Context,
	agentCrew map[string]*chat.Agent,
	selectedAgentId string,
	port string,
	matchAgentIdToTopicFn func(string, string) string,
	executeFn func(string, string) (string, error),
) (*CrewServerAgent, error) {

	firstSelectedAgent, exists := agentCrew[selectedAgentId]
	if !exists {
		return nil, fmt.Errorf("selected agent ID %s does not exist in the provided crew", selectedAgentId)
	}

	baseAgent := serverbase.NewBaseServerAgent(ctx, port, firstSelectedAgent, executeFn)

	agent := &CrewServerAgent{
		BaseServerAgent:  baseAgent,
		chatAgents:       agentCrew,
		currentChatAgent: firstSelectedAgent,
		selectedAgentId:  selectedAgentId,
	}

	// Set matchAgentIdToTopicFn
	if matchAgentIdToTopicFn != nil {
		agent.matchAgentIdToTopicFn = matchAgentIdToTopicFn
	} else {
		// Default function: return the first agent ID in the map ignoring the topic if no function is provided
		agent.matchAgentIdToTopicFn = func(currentAgent, topic string) string {
			var agentId string
			for key := range agent.chatAgents {
				agentId = key
				break
			}
			return agentId
		}
	}

	// Set executeFunction to default if not provided
	if executeFn == nil {
		agent.ExecuteFn = agent.executeFunction
	}

	agent.Log.Info("👥 CrewServerAgent initialized with chat agents, starting with agent ID: %s", selectedAgentId)

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


// StartServer starts the HTTP server with all routes
func (agent *CrewServerAgent) StartServer() error {
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
