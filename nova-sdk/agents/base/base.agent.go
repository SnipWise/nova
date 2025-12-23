package base

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/openai/openai-go/v3"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// Agent is the shared base agent structure that contains common fields
// used by all agent types (chat, tools, compressor, structured, etc.)
type Agent struct {
	Ctx                  context.Context
	Config               agents.Config
	ChatCompletionParams openai.ChatCompletionNewParams
	OpenaiClient         openai.Client
	Log                  logger.Logger
	StreamCanceled       bool
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
	// Combine existing messages with new messages
	agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, messages...)

	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)

	if err != nil {
		return "", "", err
	}

	if len(completion.Choices) > 0 {
		// Append the full response as an assistant message to the agent's messages
		agent.ChatCompletionParams.Messages = append(
			agent.ChatCompletionParams.Messages,
			openai.AssistantMessage(completion.Choices[0].Message.Content),
		)

		response = completion.Choices[0].Message.Content
		finishReason = completion.Choices[0].FinishReason

		return response, finishReason, nil
	}

	return "", "", errors.New("no choices found")
}

// GenerateCompletionWithReasoning executes a chat completion with the provided messages
// and returns both the response and reasoning content
func (agent *Agent) GenerateCompletionWithReasoning(messages []openai.ChatCompletionMessageParamUnion) (response string, reasoning string, finishReason string, err error) {
	// Combine existing messages with new messages
	agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, messages...)

	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)

	if err != nil {
		return "", "", "", err
	}

	if len(completion.Choices) == 0 {
		return "", "", "", errors.New("no choices found")
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

	// Append the full response as an assistant message to the agent's messages
	agent.ChatCompletionParams.Messages = append(
		agent.ChatCompletionParams.Messages,
		openai.AssistantMessage(response),
	)

	return response, reasoning, finishReason, nil
}

// GenerateStreamCompletion executes a streaming chat completion with the provided messages
// The callback is called for each chunk of the response
func (agent *Agent) GenerateStreamCompletion(
	messages []openai.ChatCompletionMessageParamUnion,
	callBack func(partialResponse string, finishReason string) error,
) (response string, finishReason string, err error) {

	// Reset cancellation flag at the start of streaming
	agent.StreamCanceled = false

	// Combine existing messages with new messages
	agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, messages...)
	stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, agent.ChatCompletionParams)

	var callBackError error
	finalFinishReason := ""

	for stream.Next() {
		// Check if stream was canceled
		if agent.StreamCanceled {
			callBackError = errors.New("stream canceled by user")
			break
		}

		chunk := stream.Current()

		// Capture finishReason if present (even if there's no content)
		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			finalFinishReason = chunk.Choices[0].FinishReason
		}

		// Stream each chunk as it arrives
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			callBackError = callBack(chunk.Choices[0].Delta.Content, finalFinishReason)
			response += chunk.Choices[0].Delta.Content
		}

		if callBackError != nil {
			break
		}
	}

	// Call callback one last time with the final finishReason and empty content
	if finalFinishReason != "" {
		callBackError = callBack("", finalFinishReason)
		if callBackError != nil {
			return response, finalFinishReason, callBackError
		}
	}

	if callBackError != nil {
		return response, finalFinishReason, callBackError
	}
	if err := stream.Err(); err != nil {
		return response, finalFinishReason, err
	}
	if err := stream.Close(); err != nil {
		return response, finalFinishReason, err
	}

	// Append the full response as an assistant message to the agent's messages
	agent.ChatCompletionParams.Messages = append(
		agent.ChatCompletionParams.Messages,
		openai.AssistantMessage(response),
	)

	return response, finalFinishReason, nil
}

// GenerateStreamCompletionWithReasoning executes a streaming chat completion with reasoning support
// It calls reasoningCallback for reasoning chunks and responseCallback for response chunks
func (agent *Agent) GenerateStreamCompletionWithReasoning(
	messages []openai.ChatCompletionMessageParamUnion,
	reasoningCallback func(partialReasoning string, finishReason string) error,
	responseCallback func(partialResponse string, finishReason string) error,
) (response string, reasoning string, finishReason string, err error) {

	// Reset cancellation flag at the start of streaming
	agent.StreamCanceled = false

	// Combine existing messages with new messages
	agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, messages...)
	stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, agent.ChatCompletionParams)

	var callBackError error
	var hasReceivedReasoning bool
	var reasoningEnded bool

	for stream.Next() {
		// Check if stream was canceled
		if agent.StreamCanceled {
			callBackError = errors.New("stream canceled by user")
			break
		}

		chunk := stream.Current()

		// Capture finishReason if present (even if there's no content)
		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			finishReason = chunk.Choices[0].FinishReason
		}

		// Extract and stream reasoning content if available
		if len(chunk.Choices) > 0 {
			jsonResponse := chunk.Choices[0].Delta.RawJSON()
			var reasoningContent struct {
				ReasoningContent string `json:"reasoning_content"`
			}
			err := json.Unmarshal([]byte(jsonResponse), &reasoningContent)

			if err == nil && reasoningContent.ReasoningContent != "" {
				hasReceivedReasoning = true
				reasoningChunk := reasoningContent.ReasoningContent
				if reasoningChunk != "" {
					callBackError = reasoningCallback(reasoningChunk, finishReason)
					reasoning += reasoningChunk
					if callBackError != nil {
						break
					}
				}
			}
		}

		// Stream content chunk as it arrives
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			// If we had reasoning and this is the first content chunk, signal end of reasoning
			if hasReceivedReasoning && !reasoningEnded {
				reasoningEnded = true
				callBackError = reasoningCallback("", "end_of_reasoning")
				if callBackError != nil {
					break
				}
			}

			callBackError = responseCallback(chunk.Choices[0].Delta.Content, finishReason)
			response += chunk.Choices[0].Delta.Content
			if callBackError != nil {
				break
			}
		}
	}

	// Call callbacks one last time with the final finishReason and empty content
	if finishReason != "" {
		// Call response callback with empty content to signal response completion
		callBackError = responseCallback("", finishReason)
		if callBackError != nil {
			return response, reasoning, finishReason, callBackError
		}
	}

	if callBackError != nil {
		return response, reasoning, finishReason, callBackError
	}
	if err := stream.Err(); err != nil {
		return response, reasoning, finishReason, err
	}
	if err := stream.Close(); err != nil {
		return response, reasoning, finishReason, err
	}

	// Append the full response as an assistant message to the agent's messages
	agent.ChatCompletionParams.Messages = append(
		agent.ChatCompletionParams.Messages,
		openai.AssistantMessage(response),
	)

	return response, reasoning, finishReason, nil
}
