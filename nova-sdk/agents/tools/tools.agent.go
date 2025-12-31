package tools

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/mcptools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// ToolCallResult represents the result of tool call detection
type ToolCallResult struct {
	FinishReason         string
	Results              []string
	LastAssistantMessage string
}

// ToolCallback is a function called when a tool needs to be executed
type ToolCallback func(functionName string, arguments string) (string, error)
type ConfirmationCallback func(functionName string, arguments string) ConfirmationResponse

// StreamCallback is a function called for each chunk of streaming response
type StreamCallback func(chunk string) error

// Agent represents a simplified tools agent that hides OpenAI SDK details
type Agent struct {
	ctx            context.Context
	config         agents.Config
	modelConfig    models.Config
	internalAgent  *BaseAgent
	log            logger.Logger
	toolsFunctions map[string]func(args ...any) (any, error)
}

// ToolAgentOption is a functional option for configuring an Agent during creation
type ToolAgentOption func(*openai.ChatCompletionNewParams)

// WithTools sets custom tools for the agent
func WithOpenAITools(tools []openai.ChatCompletionToolUnionParam) ToolAgentOption {
	return func(params *openai.ChatCompletionNewParams) {
		params.Tools = tools
	}
}

func WithTools(tools []*Tool) ToolAgentOption {
	return func(params *openai.ChatCompletionNewParams) {
		params.Tools = ToOpenAITools(tools)
	}
}

func WithMCPTools(tools []mcp.Tool) ToolAgentOption {
	return func(params *openai.ChatCompletionNewParams) {
		params.Tools = mcptools.ConvertMCPToolsToOpenAITools(tools)
	}
}

// TODO: WithMCPToolsWithFilter

// NewAgent creates a new simplified tools agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
	opts ...ToolAgentOption,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	// Create internal OpenAI-based agent with converted parameters
	openaiModelConfig := models.ConvertToOpenAIModelConfig(modelConfig)

	// Add tools to model config
	//openaiModelConfig.Tools = ToOpenAITools(tools)
	// Replaced by functional options

	// Apply optional configurations
	for _, opt := range opts {
		opt(&openaiModelConfig)
	}

	internalAgent, err := NewBaseAgent(ctx, agentConfig, openaiModelConfig)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		ctx:            ctx,
		config:         agentConfig,
		modelConfig:    modelConfig,
		internalAgent:  internalAgent,
		log:            log,
		toolsFunctions: make(map[string]func(args ...any) (any, error)),
	}

	// System message is already added by the BaseAgent constructor
	// No need to add it again here

	return agent, nil
}

// Kind returns the agent type
func (agent *Agent) Kind() agents.Kind {
	return agents.Tools
}

func (agent *Agent) GetName() string {
	return agent.config.Name
}

func (agent *Agent) GetModelID() string {
	return agent.modelConfig.Name
}

// GetMessages returns all conversation messages
func (agent *Agent) GetMessages() []messages.Message {
	openaiMessages := agent.internalAgent.GetMessages()
	agentMessages := messages.ConvertFromOpenAIMessages(openaiMessages)
	return agentMessages
}

func (agent *Agent) ExportMessagesToJSON() (string, error) {
	messagesList := agent.GetMessages()
	jsonData, err := json.MarshalIndent(messagesList, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// GetContextSize returns the approximate size of the current context
func (agent *Agent) GetContextSize() int {
	return agent.internalAgent.GetCurrentContextSize()
}

// ResetMessages clears all messages except the system instruction
func (agent *Agent) ResetMessages() {
	agent.internalAgent.ResetMessages()
}

// AddMessage adds a message to the conversation history
func (agent *Agent) AddMessage(role roles.Role, content string) {
	agent.internalAgent.AddMessage(
		messages.ConvertToOpenAIMessage(messages.Message{
			Role:    role,
			Content: content,
		}),
	)
}

// AddMessages adds multiple messages to the conversation history
func (agent *Agent) AddMessages(msgs []messages.Message) {
	openaiMessages := messages.ConvertToOpenAIMessages(msgs)
	agent.internalAgent.AddMessages(openaiMessages)
}

// NOTE: IMPORTANT: Not all LLMs with tool support support parallel tool calls.
func (agent *Agent) DetectParallelToolCalls(
	userMessages []messages.Message,
	toolCallback ToolCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectParallelToolCalls(openaiMessages, toolCallback)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}

// NOTE: IMPORTANT: Not all LLMs with tool support support parallel tool calls.
func (agent *Agent) DetectParallelToolCallsWithConfirmation(
	userMessages []messages.Message,
	toolCallback ToolCallback,
	confirmationCallback ConfirmationCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectParallelToolCallsWitConfirmation(
		openaiMessages,
		toolCallback,
		confirmationCallback,
	)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}

// DetectToolCallsLoop sends messages and detects tool calls, executing them via callback
func (agent *Agent) DetectToolCallsLoop(
	userMessages []messages.Message,
	toolCallback ToolCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCallsLoop(openaiMessages, toolCallback)

	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}

func (agent *Agent) DetectToolCallsLoopWithConfirmation(
	userMessages []messages.Message,
	toolCallback ToolCallback,
	confirmationCallback ConfirmationCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCallsLoopWithConfirmation(
		openaiMessages,
		toolCallback,
		confirmationCallback,
	)

	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}

// DetectToolCallsLoopStream sends messages and detects tool calls with streaming
func (agent *Agent) DetectToolCallsLoopStream(
	userMessages []messages.Message,
	toolCallback ToolCallback,
	streamCallback StreamCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent with streaming
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCallsLoopStream(
		openaiMessages,
		toolCallback,
		streamCallback,
	)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}

// DetectToolCallsStream sends messages and detects tool calls with streaming
func (agent *Agent) DetectToolCallsLoopWithConfirmationStream(
	userMessages []messages.Message,
	toolCallback ToolCallback,
	confirmationPrompt ConfirmationCallback,
	streamCallback StreamCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent with streaming
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCallsLoopWithConfirmationStream(
		openaiMessages,
		toolCallback,
		confirmationPrompt,
		streamCallback,
	)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}

// === Config Getters and Setters ===

// GetConfig returns the agent configuration
func (agent *Agent) GetConfig() agents.Config {
	return agent.config
}

// SetConfig updates the agent configuration
func (agent *Agent) SetConfig(config agents.Config) {
	agent.config = config
	agent.internalAgent.Config = config
}

// GetModelConfig returns the model configuration
func (agent *Agent) GetModelConfig() models.Config {
	return agent.modelConfig
}

// SetModelConfig updates the model configuration
// Note: This updates the stored config but doesn't regenerate the internal OpenAI params
// For most parameters to take effect, create a new agent with the new config
func (agent *Agent) SetModelConfig(config models.Config) {
	agent.modelConfig = config
	// Update the internal OpenAI params with the new config
	openaiModelConfig := models.ConvertToOpenAIModelConfig(config)
	// Preserve the existing tools
	openaiModelConfig.Tools = agent.internalAgent.ChatCompletionParams.Tools
	agent.internalAgent.ChatCompletionParams = openaiModelConfig
}

func (agent *Agent) GetLastRequestRawJSON() string {
	return agent.internalAgent.GetLastRequestRawJSON()
}

func (agent *Agent) GetLastResponseRawJSON() string {
	return agent.internalAgent.GetLastResponseRawJSON()
}

func (agent *Agent) GetLastRequestSON() (string, error) {
	return conversion.PrettyPrint(agent.internalAgent.GetLastRequestRawJSON())
}

func (agent *Agent) GetLastResponseJSON() (string, error) {
	return conversion.PrettyPrint(agent.internalAgent.GetLastResponseRawJSON())
}