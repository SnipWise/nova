package tools

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/base"
)

// BaseAgent wraps the shared base.Agent for tools-specific functionality
type BaseAgent struct {
	*base.Agent
}

type AgentOption func(*BaseAgent)

// NewBaseAgent creates a simplified Tools agent using the shared base agent
func NewBaseAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig openai.ChatCompletionNewParams,
	options ...AgentOption,
) (toolsAgent *BaseAgent, err error) {

	// Create the shared base agent
	baseAgent, err := base.NewAgent(ctx, agentConfig, modelConfig)
	if err != nil {
		return nil, err
	}

	toolsAgent = &BaseAgent{
		Agent: baseAgent,
	}

	// Apply tools-specific options
	for _, option := range options {
		option(toolsAgent)
	}

	return toolsAgent, nil
}

func (agent *BaseAgent) Kind() (kind agents.Kind) {
	return agents.Tools
}

// NOTE: IMPORTANT: Not all LLMs with tool support support parallel tool calls.
func (agent *BaseAgent) DetectParallelToolCalls(messages []openai.ChatCompletionMessageParamUnion, toolCallBack func(functionName string, arguments string) (string, error)) (string, []string, string, error) {

	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	// Prepare messages: combine system message with user messages
	workingMessages := append(agent.ChatCompletionParams.Messages, messages...)

	agent.Log.Info("⏳ [DetectParallelToolCalls] Making function call request...")

	// Create params for this call
	paramsForCall := agent.ChatCompletionParams
	paramsForCall.Messages = workingMessages

	agent.SaveLastRequest()

	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)
	if err != nil {
		agent.Log.Error("Error making function call request: %v", err)
		return "", results, "", err
	}

	agent.SaveLastResponse(completion)

	finishReason = completion.Choices[0].FinishReason

	switch finishReason {
	case "tool_calls":
		detectedToolCalls := completion.Choices[0].Message.ToolCalls

		if len(detectedToolCalls) > 0 {
			var stopped bool
			workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallBack, nil)
			if stopped {
				// Only update if keeping history
				if agent.Config.KeepConversationHistory {
					agent.ChatCompletionParams.Messages = workingMessages
				}
				return finishReason, results, lastAssistantMessage, nil
			}
		} else {
			agent.Log.Warn("😢 No tool calls found in response")
		}

	case "stop":
		workingMessages, lastAssistantMessage = agent.handleStopReason(workingMessages, completion.Choices[0].Message.Content)

	default:
		agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
	}

	// Only update agent's conversation history if KeepConversationHistory is true
	if agent.Config.KeepConversationHistory {
		agent.ChatCompletionParams.Messages = workingMessages
	}

	return finishReason, results, lastAssistantMessage, nil
}

// TODO: -> Tools.Agent
func (agent *BaseAgent) DetectParallelToolCallsWitConfirmation(
	messages []openai.ChatCompletionMessageParamUnion,
	toolCallBack func(functionName string, arguments string) (string, error),
	confirmationCallBack func(functionName string, arguments string) ConfirmationResponse) (string, []string, string, error) {

	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	// Prepare messages: combine system message with user messages
	workingMessages := append(agent.ChatCompletionParams.Messages, messages...)

	agent.Log.Info("⏳ [DetectParallelToolCallsWitConfirmation] Making function call request...")

	// Create params for this call
	paramsForCall := agent.ChatCompletionParams
	paramsForCall.Messages = workingMessages

	agent.SaveLastRequest()

	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)
	if err != nil {
		agent.Log.Error("Error making function call request: %v", err)
		return "", results, "", err
	}

	agent.SaveLastResponse(completion)

	finishReason = completion.Choices[0].FinishReason

	switch finishReason {
	case "tool_calls":
		detectedToolCalls := completion.Choices[0].Message.ToolCalls

		if len(detectedToolCalls) > 0 {
			var stopped bool
			workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallBack, confirmationCallBack)
			if stopped {
				// Only update if keeping history
				if agent.Config.KeepConversationHistory {
					agent.ChatCompletionParams.Messages = workingMessages
				}
				return finishReason, results, lastAssistantMessage, nil
			}
		} else {
			agent.Log.Warn("😢 No tool calls found in response")
		}

	case "stop":
		workingMessages, lastAssistantMessage = agent.handleStopReason(workingMessages, completion.Choices[0].Message.Content)

	default:
		agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
	}

	// Only update agent's conversation history if KeepConversationHistory is true
	if agent.Config.KeepConversationHistory {
		agent.ChatCompletionParams.Messages = workingMessages
	}

	return finishReason, results, lastAssistantMessage, nil
}

func (agent *BaseAgent) DetectToolCallsLoop(messages []openai.ChatCompletionMessageParamUnion, toolCallBack func(functionName string, arguments string) (string, error)) (string, []string, string, error) {

	stopped := false
	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	// Prepare messages: combine system message with user messages
	// Build on top of existing messages (which include system message)
	workingMessages := append(agent.ChatCompletionParams.Messages, messages...)

	for !stopped {
		agent.Log.Info("⏳ [DetectToolCallsLoop] Making function call request...")

		// Create params for this call with current working messages
		paramsForCall := agent.ChatCompletionParams
		paramsForCall.Messages = workingMessages

		agent.SaveLastRequest()

		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)
		if err != nil {
			agent.Log.Error("Error making function call request: %v", err)
			return "", results, "", err
		}

		agent.SaveLastResponse(completion)

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallBack, nil)
			} else {
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			stopped = true
			workingMessages, lastAssistantMessage = agent.handleStopReason(workingMessages, completion.Choices[0].Message.Content)

		default:
			agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
	}

	// Only update agent's conversation history if KeepConversationHistory is true
	if agent.Config.KeepConversationHistory {
		agent.ChatCompletionParams.Messages = workingMessages
	}

	return finishReason, results, lastAssistantMessage, nil
}

func (agent *BaseAgent) DetectToolCallsLoopWithConfirmation(
	messages []openai.ChatCompletionMessageParamUnion,
	toolCallBack func(functionName string, arguments string) (string, error),
	confirmationCallBack func(functionName string, arguments string) ConfirmationResponse) (string, []string, string, error) {

	stopped := false
	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	// Prepare messages: combine system message with user messages
	workingMessages := append(agent.ChatCompletionParams.Messages, messages...)

	for !stopped {
		agent.Log.Info("⏳ [LOOP][DetectToolCallsLoopWithConfirmation] Making function call request...")

		// Create params for this call with current working messages
		paramsForCall := agent.ChatCompletionParams
		paramsForCall.Messages = workingMessages

		agent.SaveLastRequest()

		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)
		if err != nil {
			agent.Log.Error("Error making function call request: %v", err)
			return "", results, "", err
		}

		agent.SaveLastResponse(completion)

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallBack, confirmationCallBack)
				if stopped && finishReason == "user_quit" {
					// Only update if keeping history
					if agent.Config.KeepConversationHistory {
						agent.ChatCompletionParams.Messages = workingMessages
					}
					return finishReason, results, lastAssistantMessage, nil
				}
			} else {
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			stopped = true
			workingMessages, lastAssistantMessage = agent.handleStopReason(workingMessages, completion.Choices[0].Message.Content)

		default:
			agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
	}

	// Only update agent's conversation history if KeepConversationHistory is true
	if agent.Config.KeepConversationHistory {
		agent.ChatCompletionParams.Messages = workingMessages
	}

	return finishReason, results, lastAssistantMessage, nil
}

func (agent *BaseAgent) DetectToolCallsLoopStream(messages []openai.ChatCompletionMessageParamUnion, toolCallback func(functionName string, arguments string) (string, error), streamCallback func(content string) error) (string, []string, string, error) {
	stopped := false
	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	// Prepare messages: combine system message with user messages
	workingMessages := append(agent.ChatCompletionParams.Messages, messages...)

	for !stopped {
		agent.Log.Info("⏳ [LOOP][DetectToolCallsLoopStream] Making function call request...")

		// Create params for this call with current working messages
		paramsForCall := agent.ChatCompletionParams
		paramsForCall.Messages = workingMessages

		agent.SaveLastRequest()

		stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, paramsForCall)
		var response string
		var cbkRes error

		for stream.Next() {
			chunk := stream.Current()

			// Capture finishReason if present (even if there's no content)
			if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
				agent.SaveLastChunkResponse(&chunk)
				//finalFinishReason = chunk.Choices[0].FinishReason
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				cbkRes = streamCallback(chunk.Choices[0].Delta.Content)
				response += chunk.Choices[0].Delta.Content
			}

			if cbkRes != nil {
				agent.Log.Error("Error in stream callback: %v", cbkRes)
				break
			}
		}

		if cbkRes != nil {
			return "", results, "", cbkRes
		}
		if err := stream.Err(); err != nil {
			return "", results, "", err
		}
		if err := stream.Close(); err != nil {
			return "", results, "", err
		}

		// QUESTION: 🤔 can I use `finalFinishReason`, see below
		// Make a non-streaming call to get tool calls
		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)
		if err != nil {
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallback, nil)
			} else {
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			stopped = true
			workingMessages, _ = agent.handleStopReason(workingMessages, response)
			lastAssistantMessage = response

		default:
			agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
	}

	// Only update agent's conversation history if KeepConversationHistory is true
	if agent.Config.KeepConversationHistory {
		agent.ChatCompletionParams.Messages = workingMessages
	}

	return finishReason, results, lastAssistantMessage, nil
}

func (agent *BaseAgent) DetectToolCallsLoopWithConfirmationStream(
	messages []openai.ChatCompletionMessageParamUnion,
	toolCallback func(functionName string, arguments string) (string, error),
	confirmationCallBack func(functionName string, arguments string) ConfirmationResponse,
	streamCallback func(content string) error) (string, []string, string, error) {

	stopped := false
	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	// Prepare messages: combine system message with user messages
	workingMessages := append(agent.ChatCompletionParams.Messages, messages...)

	for !stopped {
		agent.Log.Info("⏳ [LOOP][DetectToolCallsLoopWithConfirmationStream] Making function call request...")

		// Create params for this call with current working messages
		paramsForCall := agent.ChatCompletionParams
		paramsForCall.Messages = workingMessages

		agent.SaveLastRequest()

		stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, paramsForCall)
		var response string
		var cbkRes error

		for stream.Next() {
			chunk := stream.Current()

			// Capture finishReason if present (even if there's no content)
			if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
				agent.SaveLastChunkResponse(&chunk)
				//finalFinishReason = chunk.Choices[0].FinishReason
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				cbkRes = streamCallback(chunk.Choices[0].Delta.Content)
				response += chunk.Choices[0].Delta.Content
			}

			if cbkRes != nil {
				agent.Log.Error("Error in stream callback: %v", cbkRes)
				break
			}
		}

		if cbkRes != nil {
			return "", results, "", cbkRes
		}
		if err := stream.Err(); err != nil {
			return "", results, "", err
		}
		if err := stream.Close(); err != nil {
			return "", results, "", err
		}

		// QUESTION: 🤔 can I use `finalFinishReason`, see below
		// Make a non-streaming call to get tool calls
		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)
		if err != nil {
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallback, confirmationCallBack)
				if stopped && finishReason == "user_quit" {
					// Only update if keeping history
					if agent.Config.KeepConversationHistory {
						agent.ChatCompletionParams.Messages = workingMessages
					}
					return finishReason, results, lastAssistantMessage, nil
				}
			} else {
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			stopped = true
			workingMessages, _ = agent.handleStopReason(workingMessages, response)
			lastAssistantMessage = response

		default:
			agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
	}

	// Only update agent's conversation history if KeepConversationHistory is true
	if agent.Config.KeepConversationHistory {
		agent.ChatCompletionParams.Messages = workingMessages
	}

	return finishReason, results, lastAssistantMessage, nil
}
