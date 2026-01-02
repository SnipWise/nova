package crew

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

type CrewAgent struct {

	// Multiple chat agents for different purposes
	// The crew agent can route between them based on the task
	chatAgents      map[string]*chat.Agent
	selectedAgentId string

	currentChatAgent *chat.Agent
	toolsAgent       *tools.Agent

	ragAgent *rag.Agent

	similarityLimit float64
	maxSimilarities int

	contextSizeLimit int
	compressorAgent  *compressor.Agent

	// Routing / Orchestration agent
	orchestratorAgent agents.OrchestratorAgent

	// Retrieval function: from a topic determine which agent to use
	matchAgentIdToTopicFn func(string, string) string

	ctx context.Context
	log logger.Logger

	// Custom function executor
	executeFn func(string, string) (string, error)

	confirmationPromptFn func(string, string) tools.ConfirmationResponse

	streamCallbackFn func(chunk string, finishReason string) error
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
	matchAgentIdToTopicFn func(string, string) string,
	executeFn func(string, string) (string, error),
	confirmationPromptFn func(string, string) tools.ConfirmationResponse,
) (*CrewAgent, error) {

	firstSelectedAgent, exists := agentCrew[selectedAgentId]
	if !exists {
		return nil, fmt.Errorf("selected agent ID %s does not exist in the provided crew", selectedAgentId)
	}

	agent := &CrewAgent{
		chatAgents:       agentCrew,
		currentChatAgent: firstSelectedAgent,
		selectedAgentId:  selectedAgentId,
		toolsAgent:       nil,
		ragAgent:         nil,
		similarityLimit:  0.6,
		maxSimilarities:  3,
		contextSizeLimit: 8000,
		compressorAgent:  nil,
		ctx:              ctx,
		log:              logger.GetLoggerFromEnv(),
	}

	agent.log.Info("👥 CrewAgent initialized with chat agents, starting with agent ID: %s", selectedAgentId)

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

	// Set executeFunction (use provided or default)
	if executeFn != nil {
		agent.executeFn = executeFn
	} else {
		// executeFunction is a placeholder that should be overridden by the user
		agent.executeFn = agent.executeFunction
	}

	// Set confirmationPromptFunction (use provided or default)
	if confirmationPromptFn != nil {
		agent.confirmationPromptFn = confirmationPromptFn
	} else {
		agent.confirmationPromptFn = agent.confirmationPrompt
	}

	return agent, nil
}

func NewSimpleAgent(
	ctx context.Context,
	agentCrew map[string]*chat.Agent,
	selectedAgentId string,
) (*CrewAgent, error) {

	firstSelectedAgent, exists := agentCrew[selectedAgentId]
	if !exists {
		return nil, fmt.Errorf("selected agent ID %s does not exist in the provided crew", selectedAgentId)
	}

	agent := &CrewAgent{
		chatAgents:       agentCrew,
		currentChatAgent: firstSelectedAgent,
		selectedAgentId:  selectedAgentId,
		toolsAgent:       nil,
		ragAgent:         nil,
		similarityLimit:  0.6,
		maxSimilarities:  3,
		contextSizeLimit: 8000,
		compressorAgent:  nil,
		ctx:              ctx,
		log:              logger.GetLoggerFromEnv(),
	}

	agent.log.Info("👥 CrewAgent initialized with chat agents, starting with agent ID: %s", selectedAgentId)

	// Set matchAgentIdToTopicFn
	// if matchAgentIdToTopicFn != nil {
	// 	agent.matchAgentIdToTopicFn = matchAgentIdToTopicFn
	// } else {
	// 	// Default function: return the first agent ID in the map ignoring the topic if no function is provided
	// 	agent.matchAgentIdToTopicFn = func(currentAgent, topic string) string {
	// 		var agentId string
	// 		for key := range agent.chatAgents {
	// 			agentId = key
	// 			break
	// 		}
	// 		return agentId
	// 	}
	// }

	// Set executeFunction (use provided or default)
	// if executeFn != nil {
	// 	agent.executeFn = executeFn
	// } else {
	// 	// executeFunction is a placeholder that should be overridden by the user
	// 	agent.executeFn = agent.executeFunction
	// }

	// Set confirmationPromptFunction (use provided or default)
	// if confirmationPromptFn != nil {
	// 	agent.confirmationPromptFn = confirmationPromptFn
	// } else {
	// 	agent.confirmationPromptFn = agent.confirmationPrompt
	// }

	return agent, nil
}



// GetChatAgents returns the map of chat agents
func (agent *CrewAgent) GetChatAgents() map[string]*chat.Agent {
	return agent.chatAgents
}

// SetChatAgents sets the map of chat agents
func (agent *CrewAgent) SetChatAgents(chatAgents map[string]*chat.Agent) {
	agent.chatAgents = chatAgents
}

// AddChatAgentToCrew adds a new chat agent to the crew
func (agent *CrewAgent) AddChatAgentToCrew(id string, chatAgent *chat.Agent) error {
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
func (agent *CrewAgent) RemoveChatAgentFromCrew(id string) error {
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
func (agent *CrewAgent) Kind() agents.Kind {
	return agents.Composite
}

// GetName returns the agent name
func (agent *CrewAgent) GetName() string {
	return agent.currentChatAgent.GetName()
}

// GetModelID returns the model ID
func (agent *CrewAgent) GetModelID() string {
	return agent.currentChatAgent.GetModelID()
}

// GetMessages returns all conversation messages
func (agent *CrewAgent) GetMessages() []messages.Message {
	return agent.currentChatAgent.GetMessages()
}

// GetContextSize returns the approximate size of the current context
func (agent *CrewAgent) GetContextSize() int {
	return agent.currentChatAgent.GetContextSize()
}

// StopStream interrupts the current streaming operation
func (agent *CrewAgent) StopStream() {
	agent.currentChatAgent.StopStream()
}

// ResetMessages clears all messages except the system instruction
func (agent *CrewAgent) ResetMessages() {
	agent.currentChatAgent.ResetMessages()
}

// AddMessage adds a message to the conversation history
func (agent *CrewAgent) AddMessage(role roles.Role, content string) {
	agent.currentChatAgent.AddMessage(role, content)
}

// GenerateCompletion sends messages and returns the completion result
func (agent *CrewAgent) GenerateCompletion(userMessages []messages.Message) (*chat.CompletionResult, error) {
	return agent.currentChatAgent.GenerateCompletion(userMessages)
}

// GenerateCompletionWithReasoning sends messages and returns the completion result with reasoning
func (agent *CrewAgent) GenerateCompletionWithReasoning(userMessages []messages.Message) (*chat.ReasoningResult, error) {
	return agent.currentChatAgent.GenerateCompletionWithReasoning(userMessages)
}

// GenerateStreamCompletion sends messages and streams the response via callback
func (agent *CrewAgent) GenerateStreamCompletion(
	userMessages []messages.Message,
	callback chat.StreamCallback,
) (*chat.CompletionResult, error) {
	return agent.currentChatAgent.GenerateStreamCompletion(userMessages, callback)
}

// GenerateStreamCompletionWithReasoning sends messages and streams both reasoning and response
func (agent *CrewAgent) GenerateStreamCompletionWithReasoning(
	userMessages []messages.Message,
	reasoningCallback chat.StreamCallback,
	responseCallback chat.StreamCallback,
) (*chat.ReasoningResult, error) {
	return agent.currentChatAgent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCallback, responseCallback)
}

// ExportMessagesToJSON exports the conversation history to JSON
func (agent *CrewAgent) ExportMessagesToJSON() (string, error) {
	return agent.currentChatAgent.ExportMessagesToJSON()
}

// GetSelectedAgentId returns the currently selected agent ID
func (agent *CrewAgent) GetSelectedAgentId() string {
	return agent.selectedAgentId
}

// SetSelectedAgentId sets the currently selected agent ID
func (agent *CrewAgent) SetSelectedAgentId(agentId string) error {
	chatAgent, exists := agent.chatAgents[agentId]
	if !exists {
		return fmt.Errorf("no chat agent found with ID: %s", agentId)
	}
	agent.selectedAgentId = agentId
	agent.currentChatAgent = chatAgent
	agent.log.Info("🔀 Switched to agent ID: %s", agentId)
	return nil
}

// GetContext returns the crew agent's context
func (agent *CrewAgent) GetContext() context.Context {
	return agent.ctx
}

// SetContext updates the crew agent's context
func (agent *CrewAgent) SetContext(ctx context.Context) {
	agent.ctx = ctx
}
