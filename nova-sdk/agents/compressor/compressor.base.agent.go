package compressor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/base"
	"github.com/snipwise/nova/nova-sdk/messages"
)

// BaseAgent wraps the shared base.Agent and adds compression-specific functionality
type BaseAgent struct {
	*base.Agent
	compressionPrompt string
}

type AgentOption func(*BaseAgent)

// WithCompressionPrompt sets a custom compression prompt for the agent
func WithCompressionPrompt(prompt string) AgentOption {
	return func(agent *BaseAgent) {
		agent.compressionPrompt = prompt
	}
}

// NewBaseAgent creates a new CompressorAgent instance using the shared base agent
func NewBaseAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig openai.ChatCompletionNewParams,
	options ...AgentOption,
) (compressorAgent *BaseAgent, err error) {

	// Create the shared base agent
	baseAgent, err := base.NewAgent(ctx, agentConfig, modelConfig)
	if err != nil {
		return nil, err
	}

	compressorAgent = &BaseAgent{
		Agent:             baseAgent,
		compressionPrompt: Prompts.Minimalist,
	}

	// Apply compressor-specific options
	for _, option := range options {
		option(compressorAgent)
	}

	return compressorAgent, nil
}

// resetMessages uses the base ResetMessages method
func (agent *BaseAgent) resetMessages() {
	agent.ResetMessages()
}

func (agent *BaseAgent) SetCompressionPrompt(prompt string) {
	agent.compressionPrompt = prompt
}

func (agent *BaseAgent) CompressContext(messagesList []openai.ChatCompletionMessageParamUnion) (response string, finishReason string, err error) {

	// Start with reset messages
	agent.resetMessages()

	// Add compression prompt as user message
	agent.ChatCompletionParams.Messages = append(
		agent.ChatCompletionParams.Messages,
		openai.UserMessage(agent.compressionPrompt),
	)

	// Convert messages to text format
	var textBuilder strings.Builder
	stringMessages := messages.ConvertFromOpenAIMessages(messagesList)

	for _, msg := range stringMessages {
		textBuilder.WriteString(fmt.Sprintf("%s: ", msg.Role))
		textBuilder.WriteString(msg.Content)
		textBuilder.WriteString("\n")
	}

	text := textBuilder.String()

	agent.ChatCompletionParams.Messages = append(
		agent.ChatCompletionParams.Messages,
		openai.UserMessage("CONVERSATION:\n"+text),
	)

	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)

	if err != nil {
		return "", "", err
	}

	if len(completion.Choices) > 0 {
		response = completion.Choices[0].Message.Content
		finishReason = completion.Choices[0].FinishReason
		return response, finishReason, nil
	} else {
		return "", "", errors.New("no choices found")
	}
}

func (agent *BaseAgent) CompressContextStream(
	messagesList []openai.ChatCompletionMessageParamUnion,
	callBack func(partialResponse string, finishReason string) error) (response string, finishReason string, err error) {

	// Start with reset messages
	agent.resetMessages()

	// Add compression prompt as user message
	agent.ChatCompletionParams.Messages = append(
		agent.ChatCompletionParams.Messages,
		openai.UserMessage(agent.compressionPrompt),
	)

	// Convert messages to text format
	var textBuilder strings.Builder
	stringMessages := messages.ConvertFromOpenAIMessages(messagesList)

	for _, msg := range stringMessages {
		textBuilder.WriteString(fmt.Sprintf("%s: ", msg.Role))
		textBuilder.WriteString(msg.Content)
		textBuilder.WriteString("\n")
	}

	text := textBuilder.String()

	agent.ChatCompletionParams.Messages = append(
		agent.ChatCompletionParams.Messages,
		openai.UserMessage("CONVERSATION:\n"+text),
	)

	stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, agent.ChatCompletionParams)

	var callBackError error
	finalFinishReason := ""

	for stream.Next() {
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

	return response, finalFinishReason, nil
}

type SystemInstructions struct {
	Minimalist string
	Expert     string
}

var Instructions = SystemInstructions{
	Minimalist: `You are a context compression assistant. Your task is to summarize conversations concisely, preserving key facts, decisions, and context needed for continuation.`,
	Expert: `
	You are a context compression specialist. Your task is to analyze the conversation history and compress it while preserving all essential information.

	## Instructions:
	1. **Preserve Critical Information**: Keep all important facts, decisions, code snippets, file paths, function names, and technical details
	2. **Remove Redundancy**: Eliminate repetitive discussions, failed attempts, and conversational fluff
	3. **Maintain Chronology**: Keep the logical flow and order of important events
	4. **Summarize Discussions**: Convert long discussions into concise summaries with key takeaways
	5. **Keep Context**: Ensure the compressed version provides enough context for continuing the conversation

	## Output Format:
	Return a compressed version of the conversation that:
	- Uses clear, concise language
	- Groups related topics together
	- Highlights key decisions and outcomes
	- Preserves technical accuracy
	- Maintains references to files, functions, and code

	## Compression Guidelines:
	- Remove: Greetings, acknowledgments, verbose explanations, failed attempts
	- Keep: Facts, code, decisions, file paths, function signatures, error messages, requirements
	- Summarize: Long discussions into bullet points with essential information
	`,
}

type CompressionPrompts struct {
	Minimalist      string
	Structured      string
	UltraShort      string
	ContinuityFocus string
}

var Prompts = CompressionPrompts{
	//recommended
	Minimalist: `Summarize the conversation history concisely, preserving key facts, decisions, and context needed for continuation.`,
	Structured: `Compress this conversation into a brief summary including:
		- Main topics discussed
		- Key decisions/conclusions
		- Important context for next exchanges
		Keep it under 200 words.
	`,
	UltraShort:      `Summarize this conversation: extract key facts, decisions, and essential context only.`,
	ContinuityFocus: `Create a compact summary of this conversation that preserves all information needed to continue the discussion naturally.`,
}
