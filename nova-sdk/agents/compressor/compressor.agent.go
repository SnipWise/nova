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

// CompressorAgentOption is a functional option for configuring an Agent during creation
type CompressorAgentOption func(*Agent)

// BeforeCompletion sets a hook that is called before each compression (standard and streaming)
func BeforeCompletion(fn func(*Agent)) CompressorAgentOption {
	return func(a *Agent) {
		a.beforeCompletion = fn
	}
}

// AfterCompletion sets a hook that is called after each compression (standard and streaming)
func AfterCompletion(fn func(*Agent)) CompressorAgentOption {
	return func(a *Agent) {
		a.afterCompletion = fn
	}
}

// Agent represents a simplified compressor agent that hides OpenAI SDK details
type Agent struct {
	config        agents.Config
	modelConfig   models.Config
	internalAgent *BaseAgent
	log           logger.Logger

	// Lifecycle hooks
	beforeCompletion func(*Agent)
	afterCompletion  func(*Agent)
}

// NewAgent creates a new simplified compressor agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
	options ...any,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	// Separate AgentOption (for BaseAgent) from CompressorAgentOption (for Agent)
	var baseOptions []AgentOption
	var agentOptions []CompressorAgentOption
	for _, opt := range options {
		switch o := opt.(type) {
		case AgentOption:
			baseOptions = append(baseOptions, o)
		case CompressorAgentOption:
			agentOptions = append(agentOptions, o)
		}
	}

	// Create internal OpenAI-based agent with converted parameters
	openaiModelConfig := models.ConvertToOpenAIModelConfig(modelConfig)

	internalAgent, err := NewBaseAgent(ctx, agentConfig, openaiModelConfig, baseOptions...)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		config:        agentConfig,
		modelConfig:   modelConfig,
		internalAgent: internalAgent,
		log:           log,
	}

	// Apply CompressorAgentOption configurations
	for _, opt := range agentOptions {
		opt(agent)
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

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(messagesList)

	// Call internal agent
	response, finishReason, err := agent.internalAgent.CompressContext(openaiMessages)
	if err != nil {
		return nil, err
	}

	result := &CompressionResult{
		CompressedText: response,
		FinishReason:   finishReason,
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return result, nil
}

// CompressMessagesStream compresses a list of messages and streams the result via callback
func (agent *Agent) CompressContextStream(
	messagesList []messages.Message,
	callback StreamCallback,
) (*CompressionResult, error) {
	if len(messagesList) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(messagesList)

	// Call internal agent with streaming
	response, finishReason, err := agent.internalAgent.CompressContextStream(openaiMessages, callback)
	if err != nil {
		return nil, err
	}

	result := &CompressionResult{
		CompressedText: response,
		FinishReason:   finishReason,
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return result, nil
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
	agent.internalAgent.ChatCompletionParams = models.ConvertToOpenAIModelConfig(config)
}

func (agent *Agent) GetLastRequestRawJSON() string {
	return agent.internalAgent.GetLastRequestRawJSON()
}
func (agent *Agent) GetLastResponseRawJSON() string {
	return agent.internalAgent.GetLastResponseRawJSON()
}

func (agent *Agent) GetLastRequestJSON() (string, error) {
	return agent.internalAgent.GetLastRequestSON()
}

func (agent *Agent) GetLastResponseJSON() (string, error) {
	return agent.internalAgent.GetLastResponseJSON()
}

// GetContext returns the agent's context
func (agent *Agent) GetContext() context.Context {
	return agent.internalAgent.GetContext()
}

// SetContext updates the agent's context
func (agent *Agent) SetContext(ctx context.Context) {
	agent.internalAgent.SetContext(ctx)
}
