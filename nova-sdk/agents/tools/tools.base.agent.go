package tools

import (
	"context"
	"fmt"
	"time"

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

	agent.Log.Info("⏳ [DetectParallelToolCalls] Making function call request...")
	agent.ChatCompletionParams.Messages = messages

	// Capture request for telemetry
	agent.CaptureRequest(agent.ChatCompletionParams)
	startTime := time.Now()

	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)
	if err != nil {
		agent.Log.Error("Error making function call request:", err)
		agent.CaptureError(err, "DetectParallelToolCalls")
		return "", results, "", err
	}

	// Capture response for telemetry
	agent.CaptureResponse(completion, startTime)

	finishReason = completion.Choices[0].FinishReason

	switch finishReason {
	case "tool_calls":
		detectedToolCalls := completion.Choices[0].Message.ToolCalls

		if len(detectedToolCalls) > 0 {
			var stopped bool
			messages, stopped, finishReason = agent.processToolCalls(messages, detectedToolCalls, &results, toolCallBack, nil)
			if stopped {
				return finishReason, results, lastAssistantMessage, nil
			}
		} else {
			agent.Log.Warn("😢 No tool calls found in response")
		}

	case "stop":
		messages, lastAssistantMessage = agent.handleStopReason(messages, completion.Choices[0].Message.Content)

	default:
		agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
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

	agent.Log.Info("⏳ [DetectParallelToolCallsWitConfirmation] Making function call request...")
	agent.ChatCompletionParams.Messages = messages

	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)
	if err != nil {
		agent.Log.Error("Error making function call request:", err)
		return "", results, "", err
	}

	finishReason = completion.Choices[0].FinishReason

	switch finishReason {
	case "tool_calls":
		detectedToolCalls := completion.Choices[0].Message.ToolCalls

		if len(detectedToolCalls) > 0 {
			var stopped bool
			messages, stopped, finishReason = agent.processToolCalls(messages, detectedToolCalls, &results, toolCallBack, confirmationCallBack)
			if stopped {
				return finishReason, results, lastAssistantMessage, nil
			}
		} else {
			agent.Log.Warn("😢 No tool calls found in response")
		}

	case "stop":
		messages, lastAssistantMessage = agent.handleStopReason(messages, completion.Choices[0].Message.Content)

	default:
		agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
	}

	return finishReason, results, lastAssistantMessage, nil
}

func (agent *BaseAgent) DetectToolCallsLoop(messages []openai.ChatCompletionMessageParamUnion, toolCallBack func(functionName string, arguments string) (string, error)) (string, []string, string, error) {

	stopped := false
	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	for !stopped {
		agent.Log.Info("⏳ [DetectToolCallsLoop] Making function call request...")

		agent.ChatCompletionParams.Messages = messages

		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)
		if err != nil {
			agent.Log.Error("Error making function call request:", err)
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				messages, stopped, finishReason = agent.processToolCalls(messages, detectedToolCalls, &results, toolCallBack, nil)
			} else {
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			stopped = true
			messages, lastAssistantMessage = agent.handleStopReason(messages, completion.Choices[0].Message.Content)

		default:
			agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
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

	for !stopped {
		agent.Log.Info("⏳ [LOOP][DetectToolCallsLoopWithConfirmation] Making function call request...")

		agent.ChatCompletionParams.Messages = messages

		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)
		if err != nil {
			agent.Log.Error("Error making function call request:", err)
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				messages, stopped, finishReason = agent.processToolCalls(messages, detectedToolCalls, &results, toolCallBack, confirmationCallBack)
				if stopped && finishReason == "user_quit" {
					return finishReason, results, lastAssistantMessage, nil
				}
			} else {
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			stopped = true
			messages, lastAssistantMessage = agent.handleStopReason(messages, completion.Choices[0].Message.Content)

		default:
			agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
	}
	return finishReason, results, lastAssistantMessage, nil
}

func (agent *BaseAgent) DetectToolCallsLoopStream(messages []openai.ChatCompletionMessageParamUnion, toolCallback func(functionName string, arguments string) (string, error), streamCallback func(content string) error) (string, []string, string, error) {
	stopped := false
	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	for !stopped {
		agent.Log.Info("⏳ [LOOP][DetectToolCallsLoopStream] Making function call request...")

		agent.ChatCompletionParams.Messages = messages

		stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, agent.ChatCompletionParams)
		var response string
		var cbkRes error

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				cbkRes = streamCallback(chunk.Choices[0].Delta.Content)
				response += chunk.Choices[0].Delta.Content
			}

			if cbkRes != nil {
				agent.Log.Error("Error in stream callback:", cbkRes)
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

		// Make a non-streaming call to get tool calls
		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)
		if err != nil {
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				messages, stopped, finishReason = agent.processToolCalls(messages, detectedToolCalls, &results, toolCallback, nil)
			} else {
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			stopped = true
			messages, _ = agent.handleStopReason(messages, response)
			lastAssistantMessage = response

		default:
			agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
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

	for !stopped {
		agent.Log.Info("⏳ [LOOP][DetectToolCallsLoopWithConfirmationStream] Making function call request...")

		agent.ChatCompletionParams.Messages = messages

		stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, agent.ChatCompletionParams)
		var response string
		var cbkRes error

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				cbkRes = streamCallback(chunk.Choices[0].Delta.Content)
				response += chunk.Choices[0].Delta.Content
			}

			if cbkRes != nil {
				agent.Log.Error("Error in stream callback:", cbkRes)
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

		// Make a non-streaming call to get tool calls
		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)
		if err != nil {
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				messages, stopped, finishReason = agent.processToolCalls(messages, detectedToolCalls, &results, toolCallback, confirmationCallBack)
				if stopped && finishReason == "user_quit" {
					return finishReason, results, lastAssistantMessage, nil
				}
			} else {
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			stopped = true
			messages, _ = agent.handleStopReason(messages, response)
			lastAssistantMessage = response

		default:
			agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
	}
	return finishReason, results, lastAssistantMessage, nil
}
