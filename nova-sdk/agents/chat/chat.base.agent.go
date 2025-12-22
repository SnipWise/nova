package chat

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

type BaseAgent struct {
	ctx    context.Context
	config agents.Config

	chatCompletionParams openai.ChatCompletionNewParams
	openaiClient         openai.Client
	log                  logger.Logger
	streamCanceled       bool
}

type AgentOption func(*BaseAgent)

// NewBaseAgent creates a new ChatAgent instance
func NewBaseAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig openai.ChatCompletionNewParams,
	options ...AgentOption,
) (chatAgent *BaseAgent, err error) {

	client, log, err := agents.InitializeConnection(ctx, agentConfig, models.Config{
		Name: modelConfig.Model,
	})
	if err != nil {
		return nil, err
	}

	chatAgent = &BaseAgent{
		ctx:                  ctx,
		config:               agentConfig,
		chatCompletionParams: modelConfig,
		openaiClient:         client,
		log:                  log,
	}

	chatAgent.chatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{}

	chatAgent.chatCompletionParams.Messages = append(chatAgent.chatCompletionParams.Messages, openai.SystemMessage(agentConfig.SystemInstructions))

	for _, option := range options {
		option(chatAgent)
	}

	return chatAgent, nil
}

func (agent *BaseAgent) Kind() (kind agents.Kind) {
	return agents.Chat
}

func (agent *BaseAgent) GetMessages() (messages []openai.ChatCompletionMessageParamUnion) {
	return agent.chatCompletionParams.Messages
}

// AddMessage adds a new message to the agent's message history
func (agent *BaseAgent) AddMessage(message openai.ChatCompletionMessageParamUnion) {
	agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, message)
}

// GetStringMessages converts all messages to a slice of StringMessage with role and content as strings
func (agent *BaseAgent) GetStringMessages() (stringMessages []messages.Message) {

	stringMessages = messages.ConvertFromOpenAIMessages(agent.chatCompletionParams.Messages)

	return stringMessages
}

func (agent *BaseAgent) GetCurrentContextSize() (contextSize int) {
	stringMessages := agent.GetStringMessages()
	//var totalSize int
	for _, msg := range stringMessages {
		contextSize += len(msg.Content)
	}
	return contextSize + len(agent.config.SystemInstructions)
}

// StopStream interrupts the current streaming operation
func (agent *BaseAgent) StopStream() {
	agent.streamCanceled = true
}

// ResetMessages clears the agent's message history except for the initial system message
func (agent *BaseAgent) ResetMessages() {
	// Remove existing messages except the first system message if it's a system message
	if len(agent.chatCompletionParams.Messages) > 0 {
		firstMsg := agent.chatCompletionParams.Messages[0]
		if firstMsg.OfSystem != nil {
			agent.chatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{firstMsg}
		} else {
			agent.chatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{}
		}
	}
}

func (agent *BaseAgent) GenerateCompletion(messages []openai.ChatCompletionMessageParamUnion) (response string, finishReason string, err error) {
	// Preserve existing system messages from agent.Params
	// Combine existing system messages with new messages
	agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, messages...)

	completion, err := agent.openaiClient.Chat.Completions.New(agent.ctx, agent.chatCompletionParams)

	if err != nil {
		return "", "", err
	}

	if len(completion.Choices) > 0 {
		// Append the full response as an assistant message to the agent's messages
		agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, openai.AssistantMessage(completion.Choices[0].Message.Content))

		response = completion.Choices[0].Message.Content
		finishReason = completion.Choices[0].FinishReason

		return response, finishReason, nil
	} else {
		return "", "", errors.New("no choices found")
	}
}

// GenerateCompletionWithReasoning executes a chat completion with the provided messages.
// It sends the messages to the model and returns the first choice's content and reasoning.
//
// Parameters:
//   - messages: The conversation messages to send to the model
//
// Returns:
//   - string: The content of the first choice from the model's response
//   - string: The reasoning content from the model's response
func (agent *BaseAgent) GenerateCompletionWithReasoning(messages []openai.ChatCompletionMessageParamUnion) (response string, reasoning string, finishReason string, err error) {

	// Combine existing system messages with new messages
	agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, messages...)
	completion, err := agent.openaiClient.Chat.Completions.New(agent.ctx, agent.chatCompletionParams)

	if err != nil {
		return "", "", "", err
	}
	finishReason = completion.Choices[0].FinishReason

	if len(completion.Choices) > 0 {
		jsonResponse := completion.Choices[0].Message.RawJSON()

		// extract the content of the reasoning_content field from the jsonResponse
		var reasoningContent struct {
			ReasoningContent string `json:"reasoning_content"`
		}
		err := json.Unmarshal([]byte(jsonResponse), &reasoningContent)
		if err != nil {
			return response, reasoning, finishReason, err
		}

		reasoning = reasoningContent.ReasoningContent
		response = completion.Choices[0].Message.Content
		finishReason = completion.Choices[0].FinishReason

		// Append the full response as an assistant message to the agent's messages
		agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, openai.AssistantMessage(response))

		return response, reasoning, finishReason, nil
	} else {
		return response, reasoning, finishReason, errors.New("no choices found")
	}
}

func (agent *BaseAgent) GenerateStreamCompletion(
	messages []openai.ChatCompletionMessageParamUnion,
	callBack func(partialResponse string, finishReason string) error) (response string, finishReason string, err error) {

	// Reset cancellation flag at the start of streaming
	agent.streamCanceled = false

	// Combine existing system messages with new messages
	agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, messages...)
	stream := agent.openaiClient.Chat.Completions.NewStreaming(agent.ctx, agent.chatCompletionParams)

	var callBackError error
	finalFinishReason := ""

	for stream.Next() {
		// Check if stream was canceled
		if agent.streamCanceled {
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

	// QUESTION: IMPORTANT: what happens if it's something other than stop?
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
	agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, openai.AssistantMessage(response))

	return response, finalFinishReason, nil
}

func (agent *BaseAgent) GenerateStreamCompletionWithReasoning(
	messages []openai.ChatCompletionMessageParamUnion,
	reasoningCallback func(partialReasoning string, finishReason string) error,
	responseCallback func(partialResponse string, finishReason string) error,
) (response string, reasoning string, finishReason string, err error) {

	// Reset cancellation flag at the start of streaming
	agent.streamCanceled = false

	// Combine existing system messages with new messages
	agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, messages...)
	stream := agent.openaiClient.Chat.Completions.NewStreaming(agent.ctx, agent.chatCompletionParams)

	var callBackError error
	var hasReceivedReasoning bool
	var reasoningEnded bool

	for stream.Next() {
		// Check if stream was canceled
		if agent.streamCanceled {
			callBackError = errors.New("stream canceled by user")
			break
		}

		chunk := stream.Current()

		// Capture finishReason if present (even if there's no content)
		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			finishReason = chunk.Choices[0].FinishReason
		}

		// NOTE: Reasoning
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
				//callBackError = reasoningCallback("", finishReason)
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
	agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, openai.AssistantMessage(response))

	return response, reasoning, finishReason, nil
}
