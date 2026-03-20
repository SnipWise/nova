package tools

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/base"
)

const (
	errFunctionCallRequest = "Error making function call request: %v"
	finishReasonToolCalls  = "tool_calls"
	finishReasonStop       = "stop"
	msgNoToolCalls         = "😢 No tool calls found in response"
	msgUnexpectedResponse  = "🔴 Unexpected response: %s\n"
)

// WIP:
// GOAL: be able to check state of tool calls across multiple invocations
type LastToolCallsState struct {
	// If a tool call is awaiting user confirmation
	// Possible values: `Confirmed`, `Denied`, `Quit`
	// Denied: do not execute the tool call, but continue the flow
	// Quit: stop the entire agent execution (exit loop)
	Confirmation    ConfirmationResponse
	ExecutionResult ToolExecutionResult
}

// BaseAgent wraps the shared base.Agent for tools-specific functionality
type BaseAgent struct {
	*base.Agent
	// State of the last tool calls processed
	lastState LastToolCallsState
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

// DetectParallelToolCalls detects and executes parallel tool calls.
// Note: not all LLMs with tool support implement parallel tool calls.
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
		agent.Log.Error(errFunctionCallRequest, err)
		return "", results, "", err
	}

	agent.SaveLastResponse(completion)

	finishReason = completion.Choices[0].FinishReason

	switch finishReason {
	case finishReasonToolCalls:
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
			//agent.Log.Warn("😢 No tool calls found in response")
			agent.Log.Info(msgNoToolCalls)

		}

	case finishReasonStop:
		workingMessages, lastAssistantMessage = agent.handleStopReason(workingMessages, completion.Choices[0].Message.Content)

	default:
		agent.Log.Error(fmt.Sprintf(msgUnexpectedResponse, finishReason))
	}

	// Only update agent's conversation history if KeepConversationHistory is true
	if agent.Config.KeepConversationHistory {
		agent.ChatCompletionParams.Messages = workingMessages
	}

	return finishReason, results, lastAssistantMessage, nil
}

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
		agent.Log.Error(errFunctionCallRequest, err)
		return "", results, "", err
	}

	agent.SaveLastResponse(completion)

	finishReason = completion.Choices[0].FinishReason

	switch finishReason {
	case finishReasonToolCalls:
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
			//agent.Log.Warn("😢 No tool calls found in response")
			agent.Log.Info(msgNoToolCalls)
		}

	case finishReasonStop:
		workingMessages, lastAssistantMessage = agent.handleStopReason(workingMessages, completion.Choices[0].Message.Content)

	default:
		agent.Log.Error(fmt.Sprintf(msgUnexpectedResponse, finishReason))
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
			agent.Log.Error(errFunctionCallRequest, err)
			return "", results, "", err
		}

		agent.SaveLastResponse(completion)

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case finishReasonToolCalls:
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallBack, nil)
			} else {
				//agent.Log.Warn("😢 No tool calls found in response")
				agent.Log.Info(msgNoToolCalls)
			}

		case finishReasonStop:
			stopped = true
			workingMessages, lastAssistantMessage = agent.handleStopReason(workingMessages, completion.Choices[0].Message.Content)

		default:
			agent.Log.Error(fmt.Sprintf(msgUnexpectedResponse, finishReason))
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
			agent.Log.Error(errFunctionCallRequest, err)
			return "", results, "", err
		}

		agent.SaveLastResponse(completion)

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case finishReasonToolCalls:
			detectedToolCalls := completion.Choices[0].Message.ToolCalls
			if len(detectedToolCalls) == 0 {
				agent.Log.Info(msgNoToolCalls)
				break
			}
			workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallBack, confirmationCallBack)
			if stopped && finishReason == "user_quit" {
				agent.saveHistoryIfNeeded(workingMessages)
				return finishReason, results, lastAssistantMessage, nil
			}

		case finishReasonStop:
			stopped = true
			workingMessages, lastAssistantMessage = agent.handleStopReason(workingMessages, completion.Choices[0].Message.Content)

		default:
			agent.Log.Error(fmt.Sprintf(msgUnexpectedResponse, finishReason))
			stopped = true
		}
	}

	agent.saveHistoryIfNeeded(workingMessages)
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

		response, err := agent.collectStreamResponse(paramsForCall, streamCallback)
		if err != nil {
			return "", results, "", err
		}

		// Make a non-streaming call to get tool calls
		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)
		if err != nil {
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case finishReasonToolCalls:
			detectedToolCalls := completion.Choices[0].Message.ToolCalls
			if len(detectedToolCalls) == 0 {
				agent.Log.Info(msgNoToolCalls)
				break
			}
			workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallback, nil)

		case finishReasonStop:
			stopped = true
			workingMessages, _ = agent.handleStopReason(workingMessages, response)
			lastAssistantMessage = response

		default:
			agent.Log.Error(fmt.Sprintf(msgUnexpectedResponse, finishReason))
			stopped = true
		}
	}

	agent.saveHistoryIfNeeded(workingMessages)
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

		response, err := agent.collectStreamResponse(paramsForCall, streamCallback)
		if err != nil {
			return "", results, "", err
		}

		// Make a non-streaming call to get tool calls
		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, paramsForCall)
		if err != nil {
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case finishReasonToolCalls:
			detectedToolCalls := completion.Choices[0].Message.ToolCalls
			if len(detectedToolCalls) == 0 {
				agent.Log.Warn(msgNoToolCalls)
				break
			}
			workingMessages, stopped, finishReason = agent.processToolCalls(workingMessages, detectedToolCalls, &results, toolCallback, confirmationCallBack)
			if stopped && finishReason == "user_quit" {
				agent.saveHistoryIfNeeded(workingMessages)
				return finishReason, results, lastAssistantMessage, nil
			}

		case finishReasonStop:
			stopped = true
			workingMessages, _ = agent.handleStopReason(workingMessages, response)
			lastAssistantMessage = response

		default:
			agent.Log.Error(fmt.Sprintf(msgUnexpectedResponse, finishReason))
			stopped = true
		}
	}

	agent.saveHistoryIfNeeded(workingMessages)
	return finishReason, results, lastAssistantMessage, nil
}
