package chat

import (
	"context"
	"encoding/json"
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
	config        agents.AgentConfig
	modelConfig   models.Config
	messages      []Message
	internalAgent *BaseAgent
	log           logger.Logger
}

// NewAgent creates a new simplified chat agent
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
	// Note: TopK, MinP, RepeatPenalty, and Stop are model-specific parameters
	// that may need to be passed differently depending on the model/engine

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
	return agents.Chat
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
	agent.internalAgent.ResetMessages()
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

// Chat sends messages and returns the completion result
func (agent *Agent) Chat(userMessages []Message) (*CompletionResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Add user messages to history
	agent.messages = append(agent.messages, userMessages...)

	// Convert to OpenAI format
	openaiMessages := agent.convertToOpenAIMessages(userMessages)

	// Call internal agent
	response, finishReason, err := agent.internalAgent.Run(openaiMessages)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history
	agent.messages = append(agent.messages, Message{
		Role:    "assistant",
		Content: response,
	})

	return &CompletionResult{
		Response:     response,
		FinishReason: finishReason,
	}, nil
}

// ChatWithReasoning sends messages and returns the completion result with reasoning
func (agent *Agent) ChatWithReasoning(userMessages []Message) (*ReasoningResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Add user messages to history
	agent.messages = append(agent.messages, userMessages...)

	// Convert to OpenAI format
	openaiMessages := agent.convertToOpenAIMessages(userMessages)

	// Call internal agent
	response, reasoning, finishReason, err := agent.internalAgent.RunWithReasoning(openaiMessages)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history
	agent.messages = append(agent.messages, Message{
		Role:    "assistant",
		Content: response,
	})

	return &ReasoningResult{
		Response:     response,
		Reasoning:    reasoning,
		FinishReason: finishReason,
	}, nil
}

// ChatStream sends messages and streams the response via callback
func (agent *Agent) ChatStream(
	userMessages []Message,
	callback StreamCallback,
) (*CompletionResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Add user messages to history
	agent.messages = append(agent.messages, userMessages...)

	// Convert to OpenAI format
	openaiMessages := agent.convertToOpenAIMessages(userMessages)

	// Call internal agent with streaming
	response, finishReason, err := agent.internalAgent.RunStream(openaiMessages, callback)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history
	agent.messages = append(agent.messages, Message{
		Role:    "assistant",
		Content: response,
	})

	return &CompletionResult{
		Response:     response,
		FinishReason: finishReason,
	}, nil
}

// ChatStreamWithReasoning sends messages and streams both reasoning and response
func (agent *Agent) ChatStreamWithReasoning(
	userMessages []Message,
	reasoningCallback StreamCallback,
	responseCallback StreamCallback,
) (*ReasoningResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Add user messages to history
	agent.messages = append(agent.messages, userMessages...)

	// Convert to OpenAI format
	openaiMessages := agent.convertToOpenAIMessages(userMessages)

	// Call internal agent with streaming
	response, reasoning, finishReason, err := agent.internalAgent.RunStreamWithReasoning(
		openaiMessages,
		reasoningCallback,
		responseCallback,
	)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history
	agent.messages = append(agent.messages, Message{
		Role:    "assistant",
		Content: response,
	})

	return &ReasoningResult{
		Response:     response,
		Reasoning:    reasoning,
		FinishReason: finishReason,
	}, nil
}

// ExportMessagesToJSON exports the conversation history to JSON
func (agent *Agent) ExportMessagesToJSON() (string, error) {
	jsonData, err := json.MarshalIndent(agent.messages, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ImportMessagesFromJSON imports conversation history from JSON
func (agent *Agent) ImportMessagesFromJSON(jsonData string) error {
	var messages []Message
	err := json.Unmarshal([]byte(jsonData), &messages)
	if err != nil {
		return err
	}
	agent.messages = messages
	return nil
}
