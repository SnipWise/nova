package compressor

import (
	"context"
	"errors"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// CompressionResult represents the result of a context compression
type CompressionResult struct {
	CompressedText string
	FinishReason   string
}

// StreamCallback is a function called for each chunk of streaming response
type StreamCallback func(chunk string, finishReason string) error

// Agent represents a simplified compressor agent that hides OpenAI SDK details
type Agent struct {
	ctx           context.Context
	config        agents.Config
	modelConfig   models.Config
	internalAgent *BaseAgent
	log           logger.Logger
}

// NewAgent creates a new simplified compressor agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
	options ...AgentOption,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	// Create internal OpenAI-based agent with converted parameters
	openaiModelConfig := models.ConvertToOpenAIModelConfig(modelConfig)

	internalAgent, err := NewBaseAgent(ctx, agentConfig, openaiModelConfig, options...)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		ctx:           ctx,
		config:        agentConfig,
		modelConfig:   modelConfig,
		internalAgent: internalAgent,
		log:           log,
	}

	return agent, nil
}

func (agent *Agent) GetKind() agents.Kind {
	return agents.Compressor
}

func (agent *Agent) GetName() string {
	return agent.config.Name
}

func (agent *Agent) GetModelID() string {
	return agent.modelConfig.Name
}

// SetCompressionPrompt sets a custom compression prompt for the agent
func (agent *Agent) SetCompressionPrompt(prompt string) {
	agent.internalAgent.SetCompressionPrompt(prompt)
}

// CompressMessages compresses a list of messages and returns the compressed result
func (agent *Agent) CompressContext(messagesList []messages.Message) (*CompressionResult, error) {
	if len(messagesList) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(messagesList)

	// Call internal agent
	response, finishReason, err := agent.internalAgent.CompressContext(openaiMessages)
	if err != nil {
		return nil, err
	}

	return &CompressionResult{
		CompressedText: response,
		FinishReason:   finishReason,
	}, nil
}

// CompressMessagesStream compresses a list of messages and streams the result via callback
func (agent *Agent) CompressContextStream(
	messagesList []messages.Message,
	callback StreamCallback,
) (*CompressionResult, error) {
	if len(messagesList) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(messagesList)

	// Call internal agent with streaming
	response, finishReason, err := agent.internalAgent.CompressContextStream(openaiMessages, callback)
	if err != nil {
		return nil, err
	}

	return &CompressionResult{
		CompressedText: response,
		FinishReason:   finishReason,
	}, nil
}
