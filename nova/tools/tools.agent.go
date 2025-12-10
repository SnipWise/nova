package tools

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova/agents"
	"github.com/snipwise/nova/nova/messages"
	"github.com/snipwise/nova/nova/models"
	"github.com/snipwise/nova/nova/roles"
	"github.com/snipwise/nova/nova/toolbox/logger"
)

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
	ctx         context.Context
	config      agents.AgentConfig
	modelConfig models.Config
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
	openaiModelConfig := models.ConvertToOpenAIModelConfig(modelConfig)

	internalAgent, err := NewBaseAgent(ctx, agentConfig, openaiModelConfig)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		ctx:         ctx,
		config:      agentConfig,
		modelConfig: modelConfig,
		internalAgent: internalAgent,
		log:           log,
	}

	// Add system instruction as first message
	agent.internalAgent.AddMessage(
		openai.SystemMessage(agentConfig.SystemInstructions),
	)
	return agent, nil
}

// Kind returns the agent type
func (agent *Agent) Kind() agents.Kind {
	return agents.Tools
}

// GetMessages returns all conversation messages
func (agent *Agent) GetMessages() []messages.Message {
	openaiMessages := agent.internalAgent.GetMessages()
	agentMessages := messages.ConvertFromOpenAIMessages(openaiMessages)
	return agentMessages
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

// DetectToolCalls sends messages and detects tool calls, executing them via callback
func (agent *Agent) DetectToolCalls(
	userMessages []messages.Message,
	toolCallback ToolCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCalls(openaiMessages, toolCallback)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history if present
	if lastAssistantMessage != "" {
		agent.internalAgent.AddMessage(
			openai.AssistantMessage(lastAssistantMessage),
		)
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}

// DetectToolCallsStream sends messages and detects tool calls with streaming
func (agent *Agent) DetectToolCallsStream(
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
		agent.internalAgent.AddMessage(
			openai.AssistantMessage(lastAssistantMessage),
		)
	}

	return &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}, nil
}
