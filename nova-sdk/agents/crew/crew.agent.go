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

// CrewAgentOption is a function that configures a CrewAgent
type CrewAgentOption func(*CrewAgent) error

// WithAgentCrew sets the crew of agents and the selected agent ID
func WithAgentCrew(agentCrew map[string]*chat.Agent, selectedAgentId string) CrewAgentOption {
	return func(agent *CrewAgent) error {
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
func WithSingleAgent(chatAgent *chat.Agent) CrewAgentOption {
	return func(agent *CrewAgent) error {
		if chatAgent == nil {
			return fmt.Errorf("chat agent cannot be nil")
		}
		agent.chatAgents = map[string]*chat.Agent{"single": chatAgent}
		agent.selectedAgentId = "single"
		agent.currentChatAgent = chatAgent
		return nil
	}
}

// WithMatchAgentIdToTopicFn sets the function to match agent ID to topic
func WithMatchAgentIdToTopicFn(fn func(string, string) string) CrewAgentOption {
	return func(agent *CrewAgent) error {
		agent.matchAgentIdToTopicFn = fn
		return nil
	}
}

// WithExecuteFn sets the custom function executor
func WithExecuteFn(fn func(string, string) (string, error)) CrewAgentOption {
	return func(agent *CrewAgent) error {
		agent.executeFn = fn
		return nil
	}
}

// WithConfirmationPromptFn sets the confirmation prompt function
func WithConfirmationPromptFn(fn func(string, string) tools.ConfirmationResponse) CrewAgentOption {
	return func(agent *CrewAgent) error {
		agent.confirmationPromptFn = fn
		return nil
	}
}

// WithToolsAgent sets the tools agent
func WithToolsAgent(toolsAgent *tools.Agent) CrewAgentOption {
	return func(agent *CrewAgent) error {
		agent.toolsAgent = toolsAgent
		return nil
	}
}

// WithCompressorAgent sets the compressor agent
func WithCompressorAgent(compressorAgent *compressor.Agent) CrewAgentOption {
	return func(agent *CrewAgent) error {
		agent.compressorAgent = compressorAgent
		return nil
	}
}

// WithCompressorAgentAndContextSize sets the compressor agent and context size limit
func WithCompressorAgentAndContextSize(compressorAgent *compressor.Agent, contextSizeLimit int) CrewAgentOption {
	return func(agent *CrewAgent) error {
		agent.compressorAgent = compressorAgent
		agent.contextSizeLimit = contextSizeLimit
		return nil
	}
}

// WithRagAgent sets the RAG agent
func WithRagAgent(ragAgent *rag.Agent) CrewAgentOption {
	return func(agent *CrewAgent) error {
		agent.ragAgent = ragAgent
		return nil
	}
}

// WithRagAgentAndSimilarityConfig sets the RAG agent, similarity limit and max similarities
func WithRagAgentAndSimilarityConfig(ragAgent *rag.Agent, similarityLimit float64, maxSimilarities int) CrewAgentOption {
	return func(agent *CrewAgent) error {
		agent.ragAgent = ragAgent
		agent.similarityLimit = similarityLimit
		agent.maxSimilarities = maxSimilarities
		return nil
	}
}

// WithOrchestratorAgent sets the orchestrator agent for routing/topic detection
func WithOrchestratorAgent(orchestratorAgent agents.OrchestratorAgent) CrewAgentOption {
	return func(agent *CrewAgent) error {
		agent.orchestratorAgent = orchestratorAgent
		return nil
	}
}

// NewAgent creates a new crew agent with options
//
// Available options:
//   - WithAgentCrew(agentCrew, selectedAgentId) - Sets the crew of agents and the selected agent ID
//   - WithSingleAgent(chatAgent) - Creates a crew with a single agent (key: "single", selectedAgentId: "single")
//   - WithMatchAgentIdToTopicFn(fn) - Sets the function to match agent ID to topic for routing
//   - WithExecuteFn(fn) - Sets the custom function executor for tool execution
//   - WithConfirmationPromptFn(fn) - Sets the confirmation prompt function for human-in-the-loop
//   - WithToolsAgent(toolsAgent) - Attaches a tools agent for function calling capabilities
//   - WithCompressorAgent(compressorAgent) - Attaches a compressor agent for context compression
//   - WithCompressorAgentAndContextSize(compressorAgent, contextSizeLimit) - Attaches a compressor agent and sets the context size limit
//   - WithRagAgent(ragAgent) - Attaches a RAG agent for document retrieval
//   - WithRagAgentAndSimilarityConfig(ragAgent, similarityLimit, maxSimilarities) - Attaches a RAG agent and configures similarity settings
//   - WithOrchestratorAgent(orchestratorAgent) - Attaches an orchestrator agent for routing/topic detection
//
// At least one of WithAgentCrew or WithSingleAgent must be provided.
func NewAgent(ctx context.Context, options ...CrewAgentOption) (*CrewAgent, error) {
	agent := &CrewAgent{
		chatAgents:       nil,
		currentChatAgent: nil,
		selectedAgentId:  "",
		toolsAgent:       nil,
		ragAgent:         nil,
		similarityLimit:  0.6,
		maxSimilarities:  3,
		contextSizeLimit: 8000,
		compressorAgent:  nil,
		ctx:              ctx,
		log:              logger.GetLoggerFromEnv(),
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

	agent.log.Info("👥 CrewAgent initialized with chat agents, starting with agent ID: %s", agent.selectedAgentId)

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
	if agent.executeFn == nil {
		agent.executeFn = agent.executeFunction
	}

	// Set default confirmationPromptFn if not provided
	if agent.confirmationPromptFn == nil {
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
