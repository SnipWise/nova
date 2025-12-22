package chat

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// CompletionResult represents the result of a chat completion
type CompletionResult struct {
	Response     string
	FinishReason string
}

// ReasoningResult represents the result of a chat completion with reasoning
type ReasoningResult struct {
	Response     string
	Reasoning    string
	FinishReason string
}

// StreamCallback is a function called for each chunk of streaming response
type StreamCallback func(chunk string, finishReason string) error

// Agent represents a simplified chat agent that hides OpenAI SDK details
type Agent struct {
	ctx           context.Context
	config        agents.Config
	modelConfig   models.Config
	internalAgent *BaseAgent
	log           logger.Logger
}

// NewAgent creates a new simplified chat agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
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
		ctx:           ctx,
		config:        agentConfig,
		modelConfig:   modelConfig,
		internalAgent: internalAgent,
		log:           log,
	}

	agent.internalAgent.AddMessage(
		openai.SystemMessage(agentConfig.SystemInstructions),
	)
	return agent, nil
}

// Kind returns the agent type
func (agent *Agent) Kind() agents.Kind {
	return agents.Chat
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

// GetContextSize returns the approximate size of the current context
func (agent *Agent) GetContextSize() int {
	return agent.internalAgent.GetCurrentContextSize()
}

// StopStream interrupts the current streaming operation
func (agent *Agent) StopStream() {
	agent.internalAgent.StopStream()
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

// GenerateCompletion sends messages and returns the completion result
func (agent *Agent) GenerateCompletion(userMessages []messages.Message) (*CompletionResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	response, finishReason, err := agent.internalAgent.GenerateCompletion(openaiMessages)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history
	agent.internalAgent.AddMessage(
		openai.AssistantMessage(response),
	)

	return &CompletionResult{
		Response:     response,
		FinishReason: finishReason,
	}, nil
}

// GenerateCompletionWithReasoning sends messages and returns the completion result with reasoning
func (agent *Agent) GenerateCompletionWithReasoning(userMessages []messages.Message) (*ReasoningResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	response, reasoning, finishReason, err := agent.internalAgent.GenerateCompletionWithReasoning(openaiMessages)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history
	agent.internalAgent.AddMessage(
		openai.AssistantMessage(response),
	)

	return &ReasoningResult{
		Response:     response,
		Reasoning:    reasoning,
		FinishReason: finishReason,
	}, nil
}

// GenerateStreamCompletion sends messages and streams the response via callback
func (agent *Agent) GenerateStreamCompletion(
	userMessages []messages.Message,
	callback StreamCallback,
) (*CompletionResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent with streaming
	response, finishReason, err := agent.internalAgent.GenerateStreamCompletion(openaiMessages, callback)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history
	agent.internalAgent.AddMessage(
		openai.AssistantMessage(response),
	)

	return &CompletionResult{
		Response:     response,
		FinishReason: finishReason,
	}, nil
}

// GenerateStreamCompletionWithReasoning sends messages and streams both reasoning and response
func (agent *Agent) GenerateStreamCompletionWithReasoning(
	userMessages []messages.Message,
	reasoningCallback StreamCallback,
	responseCallback StreamCallback,
) (*ReasoningResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Add user messages to history
	//agent.messages = append(agent.messages, userMessages...)

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent with streaming
	response, reasoning, finishReason, err := agent.internalAgent.GenerateStreamCompletionWithReasoning(
		openaiMessages,
		reasoningCallback,
		responseCallback,
	)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history
	agent.internalAgent.AddMessage(
		openai.AssistantMessage(response),
	)

	return &ReasoningResult{
		Response:     response,
		Reasoning:    reasoning,
		FinishReason: finishReason,
	}, nil
}

// ExportMessagesToJSON exports the conversation history to JSON
func (agent *Agent) ExportMessagesToJSON() (string, error) {
	messagesList := agent.GetMessages()
	jsonData, err := json.MarshalIndent(messagesList, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
