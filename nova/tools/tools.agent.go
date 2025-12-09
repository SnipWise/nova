package tools

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova/agents"
	"github.com/snipwise/nova/nova/models"
	"github.com/snipwise/nova/nova/roles"
	"github.com/snipwise/nova/nova/toolbox/logger"
)

// Message represents a conversation message with a role and content
type Message struct {
	Role    roles.Role
	Content string
}

// ToolCallResult represents the result of tool call detection
type ToolCallResult struct {
	FinishReason         string
	Results              []string
	LastAssistantMessage string
}

// ToolCallback is a function called when a tool needs to be executed
type ToolCallback func(functionName string, arguments string) (string, error)

// StreamCallback is a function called for each chunk of streaming response
type StreamCallback func(chunk string) error

// Agent represents a simplified tools agent that hides OpenAI SDK details
type Agent struct {
	ctx           context.Context
	config        agents.AgentConfig
	modelConfig   models.Config
	messages      []Message
	internalAgent *BaseAgent
	log           logger.Logger
}

// NewAgent creates a new simplified tools agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.AgentConfig,
	modelConfig models.Config,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	// Create internal OpenAI-based agent with converted parameters
	openaiModelConfig := openai.ChatCompletionNewParams{
		Model: modelConfig.Name,
	}

	// Set optional parameters if provided
	if modelConfig.Temperature != nil {
		openaiModelConfig.Temperature = openai.Float(*modelConfig.Temperature)
	}
	if modelConfig.TopP != nil {
		openaiModelConfig.TopP = openai.Float(*modelConfig.TopP)
	}
	if modelConfig.MaxTokens != nil {
		openaiModelConfig.MaxTokens = openai.Int(*modelConfig.MaxTokens)
	}
	if modelConfig.FrequencyPenalty != nil {
		openaiModelConfig.FrequencyPenalty = openai.Float(*modelConfig.FrequencyPenalty)
	}
	if modelConfig.PresencePenalty != nil {
		openaiModelConfig.PresencePenalty = openai.Float(*modelConfig.PresencePenalty)
	}
	if modelConfig.Seed != nil {
		openaiModelConfig.Seed = openai.Int(*modelConfig.Seed)
	}
	if modelConfig.N != nil {
		openaiModelConfig.N = openai.Int(*modelConfig.N)
	}
	if modelConfig.ToolChoice != nil {
		openaiModelConfig.ToolChoice = *modelConfig.ToolChoice
	}
	if modelConfig.ParallelToolCalls != nil {
		openaiModelConfig.ParallelToolCalls = openai.Bool(*modelConfig.ParallelToolCalls)
	}
	if modelConfig.Tools != nil {
		openaiModelConfig.Tools = modelConfig.Tools
	}

	internalAgent, err := NewBaseAgent(ctx, agentConfig, openaiModelConfig)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		ctx:           ctx,
		config:        agentConfig,
		modelConfig:   modelConfig,
		messages:      []Message{},
		internalAgent: internalAgent,
		log:           log,
	}

	// Add system instruction as first message
	agent.messages = append(agent.messages, Message{
		Role:    "system",
		Content: agentConfig.SystemInstructions,
	})

	return agent, nil
}

// Kind returns the agent type
func (agent *Agent) Kind() agents.Kind {
	return agents.Tools
}

// GetMessages returns all conversation messages
func (agent *Agent) GetMessages() []Message {
	return agent.messages
}

// GetContextSize returns the approximate size of the current context
func (agent *Agent) GetContextSize() int {
	totalSize := 0
	for _, msg := range agent.messages {
		totalSize += len(msg.Content)
	}
	return totalSize
}

// ResetMessages clears all messages except the system instruction
func (agent *Agent) ResetMessages() {
	if len(agent.messages) > 0 && agent.messages[0].Role == "system" {
		agent.messages = []Message{agent.messages[0]}
	} else {
		agent.messages = []Message{}
	}
}

// AddMessage adds a message to the conversation history
func (agent *Agent) AddMessage(role roles.Role, content string) {
	agent.messages = append(agent.messages, Message{
		Role:    role,
		Content: content,
	})
}

// convertToOpenAIMessages converts simplified messages to OpenAI format
func (agent *Agent) convertToOpenAIMessages(messages []Message) []openai.ChatCompletionMessageParamUnion {
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			openaiMessages = append(openaiMessages, openai.SystemMessage(msg.Content))
		case "user":
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		case "assistant":
			openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
		case "developer":
			openaiMessages = append(openaiMessages, openai.DeveloperMessage(msg.Content))
		}
	}

	return openaiMessages
}

// DetectToolCalls sends messages and detects tool calls, executing them via callback
func (agent *Agent) DetectToolCalls(
	userMessages []Message,
	toolCallback ToolCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Add user messages to history
	agent.messages = append(agent.messages, userMessages...)

	// Convert to OpenAI format
	openaiMessages := agent.convertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCalls(openaiMessages, toolCallback)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history if present
	if lastAssistantMessage != "" {
		agent.messages = append(agent.messages, Message{
			Role:    "assistant",
			Content: lastAssistantMessage,
		})
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}

// DetectToolCallsStream sends messages and detects tool calls with streaming
func (agent *Agent) DetectToolCallsStream(
	userMessages []Message,
	toolCallback ToolCallback,
	streamCallback StreamCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Add user messages to history
	agent.messages = append(agent.messages, userMessages...)

	// Convert to OpenAI format
	openaiMessages := agent.convertToOpenAIMessages(userMessages)

	// Call internal agent with streaming
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCallsStream(
		openaiMessages,
		toolCallback,
		streamCallback,
	)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history if present
	if lastAssistantMessage != "" {
		agent.messages = append(agent.messages, Message{
			Role:    "assistant",
			Content: lastAssistantMessage,
		})
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}
