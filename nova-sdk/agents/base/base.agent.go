package base

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/openai/openai-go/v3"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

const errNoChoices = "no choices found"

// Agent is the shared base agent structure that contains common fields
// used by all agent types (chat, tools, compressor, structured, etc.)
type Agent struct {
	Ctx                  context.Context
	Config               agents.Config
	ChatCompletionParams openai.ChatCompletionNewParams
	OpenaiClient         openai.Client
	Log                  logger.Logger
	StreamCanceled       bool

	// // Telemetry fields for tracking requests and responses
	// lastRequest          *openai.ChatCompletionNewParams
	// lastRequestTime      time.Time
	// lastResponse         *openai.ChatCompletion
	// lastResponseTime     time.Time
	// lastResponseDuration time.Duration
	// telemetryCallback    TelemetryCallback
	// totalTokensUsed      int
	lastRequestJSON  string
	lastResponseJSON string
}

// AgentOption is a functional option for configuring an Agent
type AgentOption func(*Agent)

// NewAgent creates a new base Agent instance with common initialization logic
// This function handles the connection initialization and basic setup that all agents need
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig openai.ChatCompletionNewParams,
	options ...AgentOption,
) (*Agent, error) {

	client, log, err := agents.InitializeConnection(ctx, agentConfig, models.Config{
		Name: modelConfig.Model,
	})
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		Ctx:                  ctx,
		Config:               agentConfig,
		ChatCompletionParams: modelConfig,
		OpenaiClient:         client,
		Log:                  log,
		StreamCanceled:       false,
	}

	// Initialize messages slice
	agent.ChatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{}

	// Add system message if provided
	if agentConfig.SystemInstructions != "" {
		agent.ChatCompletionParams.Messages = append(
			agent.ChatCompletionParams.Messages,
			openai.SystemMessage(agentConfig.SystemInstructions),
		)
	}

	// Apply any additional options
	for _, option := range options {
		option(agent)
	}

	return agent, nil
}

// GetMessages returns the current message history
func (agent *Agent) GetMessages() []openai.ChatCompletionMessageParamUnion {
	return agent.ChatCompletionParams.Messages
}

// AddMessage adds a new message to the agent's message history
func (agent *Agent) AddMessage(message openai.ChatCompletionMessageParamUnion) {
	agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, message)
}

// AddMessages adds multiple messages to the agent's message history
func (agent *Agent) AddMessages(messages []openai.ChatCompletionMessageParamUnion) {
	agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, messages...)
}

// GetStringMessages converts all messages to a slice of Message with role and content as strings
func (agent *Agent) GetStringMessages() []messages.Message {
	return messages.ConvertFromOpenAIMessages(agent.ChatCompletionParams.Messages)
}

// GetCurrentContextSize calculates the total size of the current context
// by summing the length of all message contents plus the system instructions
func (agent *Agent) GetCurrentContextSize() int {
	stringMessages := agent.GetStringMessages()
	contextSize := 0
	for _, msg := range stringMessages {
		contextSize += len(msg.Content)
	}
	return contextSize + len(agent.Config.SystemInstructions)
}

// StopStream interrupts the current streaming operation
func (agent *Agent) StopStream() {
	agent.StreamCanceled = true
}

// ResetMessages clears the agent's message history except for the initial system message
func (agent *Agent) ResetMessages() {
	if len(agent.ChatCompletionParams.Messages) > 0 {
		firstMsg := agent.ChatCompletionParams.Messages[0]
		if firstMsg.OfSystem != nil {
			agent.ChatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{firstMsg}
		} else {
			agent.ChatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{}
		}
	}
}

// RemoveLastNMessages removes the last N messages from the agent's message history
// It will not remove the system message (first message) if it exists
func (agent *Agent) RemoveLastNMessages(n int) {
	if n <= 0 {
		return
	}

	totalMessages := len(agent.ChatCompletionParams.Messages)
	if totalMessages == 0 {
		return
	}

	// Check if first message is a system message
	hasSystemMessage := false
	if totalMessages > 0 && agent.ChatCompletionParams.Messages[0].OfSystem != nil {
		hasSystemMessage = true
	}

	// Calculate how many messages can be removed (excluding system message)
	removableMessages := totalMessages
	if hasSystemMessage {
		removableMessages = totalMessages - 1
	}

	// Don't remove more than available
	if n > removableMessages {
		n = removableMessages
	}

	// Calculate the new length after removal
	newLength := totalMessages - n

	// Keep messages up to newLength
	agent.ChatCompletionParams.Messages = agent.ChatCompletionParams.Messages[:newLength]
}

// SetSystemInstructions updates the system instructions for the agent
// If a system message already exists as the first message, it will be replaced
// Otherwise, a new system message will be prepended to the message list
func (agent *Agent) SetSystemInstructions(instructions string) {
	// Update the config
	agent.Config.SystemInstructions = instructions

	// Check if first message is a system message
	if len(agent.ChatCompletionParams.Messages) > 0 && agent.ChatCompletionParams.Messages[0].OfSystem != nil {
		// Replace existing system message
		agent.ChatCompletionParams.Messages[0] = openai.SystemMessage(instructions)
	} else {
		// Prepend new system message
		agent.ChatCompletionParams.Messages = append(
			[]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(instructions)},
			agent.ChatCompletionParams.Messages...,
		)
	}
}

// GenerateCompletion executes a chat completion with the provided messages
// and returns the response, finish reason, and any error
func (agent *Agent) GenerateCompletion(messages []openai.ChatCompletionMessageParamUnion) (response string, finishReason string, err error) {
	// Prepare messages for the API call
	// If KeepConversationHistory is true, add to history permanently
	// Otherwise, create a temporary message list for this call only
	var messagesToSend []openai.ChatCompletionMessageParamUnion

	if agent.Config.KeepConversationHistory {
		// Add new messages to history permanently
		agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, messages...)
		messagesToSend = agent.ChatCompletionParams.Messages
	} else {
		// Create temporary message list with system + current user messages only
		messagesToSend = append(agent.ChatCompletionParams.Messages, messages...)
	}

	// Update params with messages for this call
	paramsForCall := agent.ChatCompletionParams
	paramsForCall.Messages = messagesToSend

	agent.SaveLastRequest()

	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)

	agent.SaveLastResponse(completion)

	if err != nil {
		return "", "", err
	}

	if len(completion.Choices) > 0 {
		response = completion.Choices[0].Message.Content
		finishReason = completion.Choices[0].FinishReason

		// Only add assistant response to history if KeepConversationHistory is true and response is not empty
		if agent.Config.KeepConversationHistory && response != "" {
			agent.ChatCompletionParams.Messages = append(
				agent.ChatCompletionParams.Messages,
				openai.AssistantMessage(response),
			)
		}

		return response, finishReason, nil
	}

	return "", "", errors.New(errNoChoices)
}

// GenerateCompletionWithReasoning executes a chat completion with the provided messages
// and returns both the response and reasoning content
func (agent *Agent) GenerateCompletionWithReasoning(messages []openai.ChatCompletionMessageParamUnion) (response string, reasoning string, finishReason string, err error) {
	// Prepare messages for the API call
	// If KeepConversationHistory is true, add to history permanently
	// Otherwise, create a temporary message list for this call only
	var messagesToSend []openai.ChatCompletionMessageParamUnion

	if agent.Config.KeepConversationHistory {
		// Add new messages to history permanently
		agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, messages...)
		messagesToSend = agent.ChatCompletionParams.Messages
	} else {
		// Create temporary message list with system + current user messages only
		messagesToSend = append(agent.ChatCompletionParams.Messages, messages...)
	}

	// Update params with messages for this call
	paramsForCall := agent.ChatCompletionParams
	paramsForCall.Messages = messagesToSend

	agent.SaveLastRequest()

	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)

	agent.SaveLastResponse(completion)

	if err != nil {
		return "", "", "", err
	}

	if len(completion.Choices) == 0 {
		return "", "", "", errors.New(errNoChoices)
	}

	finishReason = completion.Choices[0].FinishReason
	jsonResponse := completion.Choices[0].Message.RawJSON()

	// Extract the content of the reasoning_content field from the jsonResponse
	var reasoningContent struct {
		ReasoningContent string `json:"reasoning_content"`
	}

	err = json.Unmarshal([]byte(jsonResponse), &reasoningContent)
	if err != nil {
		return "", "", finishReason, err
	}

	reasoning = reasoningContent.ReasoningContent
	response = completion.Choices[0].Message.Content

	// Only add assistant response to history if KeepConversationHistory is true and response is not empty
	if agent.Config.KeepConversationHistory && response != "" {
		agent.ChatCompletionParams.Messages = append(
			agent.ChatCompletionParams.Messages,
			openai.AssistantMessage(response),
		)
	}

	return response, reasoning, finishReason, nil
}

// GenerateStreamCompletion executes a streaming chat completion with the provided messages
// The callback is called for each chunk of the response
func (agent *Agent) GenerateStreamCompletion(
	messages []openai.ChatCompletionMessageParamUnion,
	callBack func(partialResponse string, finishReason string) error,
) (response string, finishReason string, err error) {

	agent.StreamCanceled = false

	paramsForCall := agent.ChatCompletionParams
	paramsForCall.Messages = agent.prepareMessagesToSend(messages)
	agent.SaveLastRequest()

	stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, paramsForCall)

	var callBackError error

	for stream.Next() {
		if agent.StreamCanceled {
			callBackError = canceledError
			break
		}
		if callBackError = agent.processStreamChunk(stream.Current(), &finishReason, &response, callBack); callBackError != nil {
			break
		}
	}

	if finishReason != "" {
		callBackError = callBack("", finishReason)
		if callBackError != nil {
			return response, finishReason, callBackError
		}
	}

	if callBackError != nil {
		return response, finishReason, callBackError
	}
	if err := agent.finalizeStream(stream); err != nil {
		return response, finishReason, err
	}

	agent.appendAssistantToHistory(response)
	return response, finishReason, nil
}

// GenerateStreamCompletionWithReasoning executes a streaming chat completion with reasoning support
// It calls reasoningCallback for reasoning chunks and responseCallback for response chunks
func (agent *Agent) GenerateStreamCompletionWithReasoning(
	messages []openai.ChatCompletionMessageParamUnion,
	reasoningCallback func(partialReasoning string, finishReason string) error,
	responseCallback func(partialResponse string, finishReason string) error,
) (response string, reasoning string, finishReason string, err error) {

	agent.StreamCanceled = false

	paramsForCall := agent.ChatCompletionParams
	paramsForCall.Messages = agent.prepareMessagesToSend(messages)
	agent.SaveLastRequest()

	stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, paramsForCall)

	var callBackError error
	var hasReceivedReasoning bool
	var reasoningEnded bool

	for stream.Next() {
		if agent.StreamCanceled {
			callBackError = canceledError
			break
		}

		chunk := stream.Current()

		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			agent.SaveLastChunkResponse(&chunk)
			finishReason = chunk.Choices[0].FinishReason
		}

		if callBackError = processReasoningChunk(chunk, finishReason, &reasoning, &hasReceivedReasoning, reasoningCallback); callBackError != nil {
			break
		}
		if callBackError = processResponseChunk(chunk, finishReason, &response, hasReceivedReasoning, &reasoningEnded, reasoningCallback, responseCallback); callBackError != nil {
			break
		}
	}

	if finishReason != "" {
		callBackError = responseCallback("", finishReason)
		if callBackError != nil {
			return response, reasoning, finishReason, callBackError
		}
	}

	if callBackError != nil {
		return response, reasoning, finishReason, callBackError
	}
	if err := agent.finalizeStream(stream); err != nil {
		return response, reasoning, finishReason, err
	}

	agent.appendAssistantToHistory(response)
	return response, reasoning, finishReason, nil
}

// SaveLastRequest stores the last request JSON for telemetry or debugging
func (agent *Agent) SaveLastRequest() error {
	bparam, err := agent.ChatCompletionParams.MarshalJSON()
	if err != nil {
		agent.Log.Error("Error marshaling last request: %v", err)
		return err
	}
	agent.lastRequestJSON = string(bparam)
	agent.Log.Debug("📡 Request Sent:\n%s", agent.lastRequestJSON)
	return nil
}

// SaveLastResponse stores the last response JSON for telemetry or debugging
func (agent *Agent) SaveLastResponse(completion *openai.ChatCompletion) error {
	//Store last request and response JSON for telemetry or debugging
	agent.lastResponseJSON = completion.RawJSON()
	agent.Log.Debug("📝 Response Received:\n%s", agent.lastResponseJSON)
	return nil
}

// SaveLastChunkResponse stores the last chunk response JSON for telemetry or debugging
func (agent *Agent) SaveLastChunkResponse(completion *openai.ChatCompletionChunk) error {
	//Store last request and response JSON for telemetry or debugging
	agent.lastResponseJSON = completion.RawJSON()
	agent.Log.Debug("🍰 Response Received:\n%s", agent.lastResponseJSON)
	return nil
}

func (agent *Agent) GetLastRequestRawJSON() string {
	return agent.lastRequestJSON
}

func (agent *Agent) GetLastResponseRawJSON() string {
	return agent.lastResponseJSON
}

func (agent *Agent) GetLastRequestSON() (string, error) {
	return conversion.PrettyPrint(agent.lastRequestJSON)
}

func (agent *Agent) GetLastResponseJSON() (string, error) {
	return conversion.PrettyPrint(agent.lastResponseJSON)
}

// GetContext returns the agent's context
func (agent *Agent) GetContext() context.Context {
	return agent.Ctx
}

// SetContext updates the agent's context
func (agent *Agent) SetContext(ctx context.Context) {
	agent.Ctx = ctx
}
