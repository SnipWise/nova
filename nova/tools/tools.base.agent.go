package tools

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared/constant"
	"github.com/snipwise/nova/nova/agents"
	"github.com/snipwise/nova/nova/messages"
	"github.com/snipwise/nova/nova/toolbox/logger"
)

type BaseAgent struct {
	ctx    context.Context
	config agents.AgentConfig

	chatCompletionParams openai.ChatCompletionNewParams
	openaiClient         openai.Client
	log                  logger.Logger
}

type AgentOption func(*BaseAgent)

// NewBaseAgent creates a simplified Tools agent
func NewBaseAgent(
	ctx context.Context,
	agentConfig agents.AgentConfig,
	modelConfig openai.ChatCompletionNewParams,
	options ...AgentOption,
) (toolsAgent *BaseAgent, err error) {

	client, log, err := agents.InitializeConnection(ctx, agentConfig.EngineURL, modelConfig.Model)
	if err != nil {
		return nil, err
	}

	toolsAgent = &BaseAgent{
		ctx:                  ctx,
		config:               agentConfig,
		chatCompletionParams: modelConfig,
		openaiClient:         client,
		log:                  log,
	}

	toolsAgent.chatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{}

	toolsAgent.chatCompletionParams.Messages = append(toolsAgent.chatCompletionParams.Messages, openai.SystemMessage(agentConfig.SystemInstructions))

	for _, option := range options {
		option(toolsAgent)
	}

	return toolsAgent, nil
}

func (agent *BaseAgent) Kind() (kind agents.Kind) {
	return agents.Tools
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

func (agent *BaseAgent) DetectToolCalls(messages []openai.ChatCompletionMessageParamUnion, toolCallBack func(functionName string, arguments string) (string, error)) (string, []string, string, error) {

	stopped := false
	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	for !stopped {
		// TOOL: Make a function call request
		agent.log.Info("⏳ Making function call request...")

		agent.chatCompletionParams.Messages = messages

		completion, err := agent.openaiClient.Chat.Completions.New(agent.ctx, agent.chatCompletionParams)
		if err != nil {
			agent.log.Error("Error making function call request:", err)
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		// Extract reasoning_content from RawJSON
		// completion.Choices[0].Message.RawJSON()

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {

				toolCallParams := make([]openai.ChatCompletionMessageToolCallUnionParam, len(detectedToolCalls))
				for i, toolCall := range detectedToolCalls {
					toolCallParams[i] = openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID:   toolCall.ID,
							Type: constant.Function("function"),
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      toolCall.Function.Name,
								Arguments: toolCall.Function.Arguments,
							},
						},
					}
				}

				// Create assistant message with tool calls using proper union type
				assistantMessage := openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: toolCallParams,
					},
				}

				// Add the assistant message with tool calls to the conversation history
				messages = append(messages, assistantMessage)

				// TOOL: Process each detected tool call
				agent.log.Info("🚀 Processing tool calls...")

				for _, toolCall := range detectedToolCalls {
					functionName := toolCall.Function.Name
					functionArgs := toolCall.Function.Arguments
					callID := toolCall.ID

					// TOOL: Execute the function with the provided arguments
					agent.log.Info(fmt.Sprintf("▶️ Executing function: %s with args: %s\n", functionName, functionArgs))

					resultContent, errExec := toolCallBack(functionName, functionArgs)

					if errExec != nil {
						agent.log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))
						stopped = true
						finishReason = "exit_loop"
						resultContent = fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error())
					}
					if resultContent == "" {
						resultContent = `{"error": "Function execution returned empty result"}`
					}
					results = append(results, resultContent)

					//fmt.Printf("Function result: %s with CallID: %s\n\n", resultContent, callID)
					//agent.log.Info(fmt.Sprintf("✅ Function %s executed successfully.\n", functionName))
					agent.log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

					// Add the tool call result to the conversation history
					messages = append(
						messages,
						openai.ToolMessage(
							resultContent,
							toolCall.ID,
						),
					)
				}

			} else {
				// TODO: Handle case where no tool calls were detected
				agent.log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			agent.log.Info("🟥 Stopping due to 'stop' finish reason.")
			stopped = true
			lastAssistantMessage = completion.Choices[0].Message.Content

			agent.log.Info(fmt.Sprintf("🤖 %s\n", lastAssistantMessage))

			// Add final assistant message to conversation history
			messages = append(messages, openai.AssistantMessage(lastAssistantMessage))

		default:
			agent.log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}

	}
	return finishReason, results, lastAssistantMessage, nil
}

func (agent *BaseAgent) DetectToolCallsStream(messages []openai.ChatCompletionMessageParamUnion, toolCallback func(functionName string, arguments string) (string, error), streamCallback func(content string) error) (string, []string, string, error) {
	stopped := false
	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	for !stopped {
		// TOOL: Make a function call request
		agent.log.Info("⏳ Making function call request...")

		agent.chatCompletionParams.Messages = messages

		stream := agent.openaiClient.Chat.Completions.NewStreaming(agent.ctx, agent.chatCompletionParams)
		var response string
		var cbkRes error

		for stream.Next() {
			chunk := stream.Current()
			// Stream each chunk as it arrives
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				cbkRes = streamCallback(chunk.Choices[0].Delta.Content)
				response += chunk.Choices[0].Delta.Content
			}

			if cbkRes != nil {
				agent.log.Error("Error in stream callback:", cbkRes)
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

		// Make a non-streaming call to get tool calls (streaming doesn't provide tool calls properly)
		completion, err := agent.openaiClient.Chat.Completions.New(agent.ctx, agent.chatCompletionParams)
		if err != nil {
			return "", results, "", err
		}

		finishReason = completion.Choices[0].FinishReason

		switch finishReason {
		case "tool_calls":
			detectedToolCalls := completion.Choices[0].Message.ToolCalls

			if len(detectedToolCalls) > 0 {
				toolCallParams := make([]openai.ChatCompletionMessageToolCallUnionParam, len(detectedToolCalls))
				for i, toolCall := range detectedToolCalls {
					toolCallParams[i] = openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID:   toolCall.ID,
							Type: constant.Function("function"),
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      toolCall.Function.Name,
								Arguments: toolCall.Function.Arguments,
							},
						},
					}
				}

				// Create assistant message with tool calls
				assistantMessage := openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: toolCallParams,
					},
				}

				messages = append(messages, assistantMessage)

				// Execute each tool call
				for _, toolCall := range detectedToolCalls {
					functionName := toolCall.Function.Name
					functionArgs := toolCall.Function.Arguments
					callID := toolCall.ID

					resultContent, errExec := toolCallback(functionName, functionArgs)

					if errExec != nil {
						agent.log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))
						stopped = true
						finishReason = "exit_loop"
						resultContent = fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error())
					}

					if resultContent == "" {
						resultContent = `{"error": "Function execution returned empty result"}`
					}
					results = append(results, resultContent)
					agent.log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

					// Add the tool call result to the conversation history
					messages = append(
						messages,
						openai.ToolMessage(
							resultContent,
							toolCall.ID,
						),
					)
				}

			} else {
				fmt.Println("😢 No tool calls found in response")
			}

		case "stop":
			agent.log.Info("🟥 Stopping due to 'stop' finish reason.")
			stopped = true
			lastAssistantMessage = response

			agent.log.Info(fmt.Sprintf("🤖 %s\n", lastAssistantMessage))

			// Add final assistant message to conversation history
			messages = append(messages, openai.AssistantMessage(lastAssistantMessage))

		default:
			agent.log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
	}
	return finishReason, results, lastAssistantMessage, nil
}
