package tools

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared/constant"
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
			agent.Log.Info("🚀 Processing tool calls...")

			for _, toolCall := range detectedToolCalls {
				functionName := toolCall.Function.Name
				functionArgs := toolCall.Function.Arguments
				callID := toolCall.ID

				// TOOL: Execute the function with the provided arguments
				agent.Log.Info(fmt.Sprintf("▶️ Executing function: %s with args: %s\n", functionName, functionArgs))

				resultContent, errExec := toolCallBack(functionName, functionArgs)

				if errExec != nil {
					agent.Log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))
					//stopped = true
					finishReason = "exit_loop"
					resultContent = fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error())
				}
				if resultContent == "" {
					resultContent = `{"error": "Function execution returned empty result"}`
				}
				results = append(results, resultContent)

				//fmt.Printf("Function result: %s with CallID: %s\n\n", resultContent, callID)
				//agent.Log.Info(fmt.Sprintf("✅ Function %s executed successfully.\n", functionName))
				agent.Log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

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
			agent.Log.Warn("😢 No tool calls found in response")
		}

	case "stop":
		agent.Log.Info("✋ Stopping due to 'stop' finish reason.")
		//stopped = true
		lastAssistantMessage = completion.Choices[0].Message.Content

		agent.Log.Info(fmt.Sprintf("🤖 %s\n", lastAssistantMessage))

		// Add final assistant message to conversation history
		messages = append(messages, openai.AssistantMessage(lastAssistantMessage))

	default:
		agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
		//stopped = true
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
			agent.Log.Info("🚀 Processing tool calls...")

			for _, toolCall := range detectedToolCalls {
				functionName := toolCall.Function.Name
				functionArgs := toolCall.Function.Arguments
				callID := toolCall.ID

				// --> NOTE: HERE

				// Ask for confirmation before executing the tool
				agent.Log.Info(fmt.Sprintf("⁉️ Requesting confirmation for function: %s with args: %s\n", functionName, functionArgs))
				confirmation := confirmationCallBack(functionName, functionArgs)

				switch confirmation {
				case Confirmed:
					// TOOL: Execute the function with the provided arguments
					agent.Log.Info(fmt.Sprintf("▶️ Executing function: %s with args: %s\n", functionName, functionArgs))

					resultContent, errExec := toolCallBack(functionName, functionArgs)

					if errExec != nil {
						agent.Log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))
						//stopped = true
						finishReason = "exit_loop"
						resultContent = fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error())
					}
					if resultContent == "" {
						resultContent = `{"error": "Function execution returned empty result"}`
					}
					results = append(results, resultContent)

					agent.Log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

					// Add the tool call result to the conversation history
					messages = append(
						messages,
						openai.ToolMessage(
							resultContent,
							toolCall.ID,
						),
					)

				case Denied:
					// Skip execution but add a message indicating the tool was denied
					agent.Log.Warn(fmt.Sprintf("⛔ Tool execution denied for function: %s\n", functionName))
					resultContent := `{"status": "denied", "message": "Tool execution was denied by user"}`
					results = append(results, resultContent)

					// Add the tool call result to the conversation history
					messages = append(
						messages,
						openai.ToolMessage(
							resultContent,
							toolCall.ID,
						),
					)

				case Quit:
					// Exit the function immediately
					agent.Log.Warn(fmt.Sprintf("🛑 Quit requested for function: %s\n", functionName))
					//stopped = true
					finishReason = "user_quit"
					return finishReason, results, lastAssistantMessage, nil
				}
			}

		} else {
			// TODO: Handle case where no tool calls were detected
			agent.Log.Warn("😢 No tool calls found in response")
		}

	case "stop":
		agent.Log.Info("✋ Stopping due to 'stop' finish reason.")
		//stopped = true
		lastAssistantMessage = completion.Choices[0].Message.Content

		agent.Log.Info(fmt.Sprintf("🤖 %s\n", lastAssistantMessage))

		// Add final assistant message to conversation history
		messages = append(messages, openai.AssistantMessage(lastAssistantMessage))

	default:
		agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
		//stopped = true
	}

	return finishReason, results, lastAssistantMessage, nil
}

func (agent *BaseAgent) DetectToolCallsLoop(messages []openai.ChatCompletionMessageParamUnion, toolCallBack func(functionName string, arguments string) (string, error)) (string, []string, string, error) {

	stopped := false
	results := []string{}
	lastAssistantMessage := ""
	finishReason := ""

	for !stopped {
		// TOOL: Make a function call request
		agent.Log.Info("⏳ [DetectToolCallsLoop] Making function call request...")

		agent.ChatCompletionParams.Messages = messages

		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)
		if err != nil {
			agent.Log.Error("Error making function call request:", err)
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
				agent.Log.Info("🚀 Processing tool calls...")

				for _, toolCall := range detectedToolCalls {
					functionName := toolCall.Function.Name
					functionArgs := toolCall.Function.Arguments
					callID := toolCall.ID

					// TOOL: Execute the function with the provided arguments
					agent.Log.Info(fmt.Sprintf("▶️ Executing function: %s with args: %s\n", functionName, functionArgs))

					resultContent, errExec := toolCallBack(functionName, functionArgs)

					if errExec != nil {
						agent.Log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))
						stopped = true
						finishReason = "exit_loop"
						resultContent = fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error())
					}
					if resultContent == "" {
						resultContent = `{"error": "Function execution returned empty result"}`
					}
					results = append(results, resultContent)

					//fmt.Printf("Function result: %s with CallID: %s\n\n", resultContent, callID)
					//agent.Log.Info(fmt.Sprintf("✅ Function %s executed successfully.\n", functionName))
					agent.Log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

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
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			agent.Log.Info("✋ Stopping due to 'stop' finish reason.")
			stopped = true
			lastAssistantMessage = completion.Choices[0].Message.Content

			agent.Log.Info(fmt.Sprintf("🤖 %s\n", lastAssistantMessage))

			// Add final assistant message to conversation history
			messages = append(messages, openai.AssistantMessage(lastAssistantMessage))

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
		// TOOL: Make a function call request
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
				agent.Log.Info("🚀 Processing tool calls...")

				for _, toolCall := range detectedToolCalls {
					functionName := toolCall.Function.Name
					functionArgs := toolCall.Function.Arguments
					callID := toolCall.ID

					// Ask for confirmation before executing the tool
					agent.Log.Info(fmt.Sprintf("⁉️ Requesting confirmation for function: %s with args: %s\n", functionName, functionArgs))
					confirmation := confirmationCallBack(functionName, functionArgs)

					switch confirmation {
					case Confirmed:
						// TOOL: Execute the function with the provided arguments
						agent.Log.Info(fmt.Sprintf("▶️ Executing function: %s with args: %s\n", functionName, functionArgs))

						resultContent, errExec := toolCallBack(functionName, functionArgs)

						if errExec != nil {
							agent.Log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))
							stopped = true
							finishReason = "exit_loop"
							resultContent = fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error())
						}
						if resultContent == "" {
							resultContent = `{"error": "Function execution returned empty result"}`
						}
						results = append(results, resultContent)

						agent.Log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

						// Add the tool call result to the conversation history
						messages = append(
							messages,
							openai.ToolMessage(
								resultContent,
								toolCall.ID,
							),
						)

					case Denied:
						// Skip execution but add a message indicating the tool was denied
						agent.Log.Warn(fmt.Sprintf("⛔ Tool execution denied for function: %s\n", functionName))
						resultContent := `{"status": "denied", "message": "Tool execution was denied by user"}`
						results = append(results, resultContent)

						// Add the tool call result to the conversation history
						messages = append(
							messages,
							openai.ToolMessage(
								resultContent,
								toolCall.ID,
							),
						)

					case Quit:
						// Exit the function immediately
						agent.Log.Warn(fmt.Sprintf("🛑 Quit requested for function: %s\n", functionName))
						stopped = true
						finishReason = "user_quit"
						return finishReason, results, lastAssistantMessage, nil
					}
				}

			} else {
				// TODO: Handle case where no tool calls were detected
				agent.Log.Warn("😢 No tool calls found in response")
			}

		case "stop":
			agent.Log.Info("✋ Stopping due to 'stop' finish reason.")
			stopped = true
			lastAssistantMessage = completion.Choices[0].Message.Content

			agent.Log.Info(fmt.Sprintf("🤖 %s\n", lastAssistantMessage))

			// Add final assistant message to conversation history
			messages = append(messages, openai.AssistantMessage(lastAssistantMessage))

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
		// TOOL: Make a function call request
		agent.Log.Info("⏳ [LOOP][DetectToolCallsLoopStream] Making function call request...")

		agent.ChatCompletionParams.Messages = messages

		stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, agent.ChatCompletionParams)
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

		// Make a non-streaming call to get tool calls (streaming doesn't provide tool calls properly)
		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)
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
						agent.Log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))
						stopped = true
						finishReason = "exit_loop"
						resultContent = fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error())
					}

					if resultContent == "" {
						resultContent = `{"error": "Function execution returned empty result"}`
					}
					results = append(results, resultContent)
					agent.Log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

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
			agent.Log.Info("✋ Stopping due to 'stop' finish reason.")
			stopped = true
			lastAssistantMessage = response

			agent.Log.Info(fmt.Sprintf("🤖 %s\n", lastAssistantMessage))

			// Add final assistant message to conversation history
			messages = append(messages, openai.AssistantMessage(lastAssistantMessage))

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
		// TOOL: Make a function call request
		agent.Log.Info("⏳ [LOOP][DetectToolCallsLoopWithConfirmationStream] Making function call request...")

		agent.ChatCompletionParams.Messages = messages

		stream := agent.OpenaiClient.Chat.Completions.NewStreaming(agent.Ctx, agent.ChatCompletionParams)
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

		// Make a non-streaming call to get tool calls (streaming doesn't provide tool calls properly)
		completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)
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

					// Ask for confirmation before executing the tool
					agent.Log.Info(fmt.Sprintf("⁉️ Requesting confirmation for function: %s with args: %s\n", functionName, functionArgs))
					confirmation := confirmationCallBack(functionName, functionArgs)

					switch confirmation {
					case Confirmed:
						// TOOL: Execute the function with the provided arguments
						agent.Log.Info(fmt.Sprintf("▶️ Executing function: %s with args: %s\n", functionName, functionArgs))

						resultContent, errExec := toolCallback(functionName, functionArgs)

						if errExec != nil {
							agent.Log.Error(fmt.Sprintf("🔴 Error executing function %s: %s\n", functionName, errExec.Error()))
							stopped = true
							finishReason = "exit_loop"
							resultContent = fmt.Sprintf(`{"error": "Function execution failed: %s"}`, errExec.Error())
						}

						if resultContent == "" {
							resultContent = `{"error": "Function execution returned empty result"}`
						}
						results = append(results, resultContent)
						agent.Log.Info(fmt.Sprintf("✅ Function result: %s with CallID: %s\n\n", resultContent, callID))

						// Add the tool call result to the conversation history
						messages = append(
							messages,
							openai.ToolMessage(
								resultContent,
								toolCall.ID,
							),
						)

					case Denied:
						// Skip execution but add a message indicating the tool was denied
						agent.Log.Warn(fmt.Sprintf("⛔ Tool execution denied for function: %s\n", functionName))
						resultContent := `{"status": "denied", "message": "Tool execution was denied by user"}`
						results = append(results, resultContent)

						// Add the tool call result to the conversation history
						messages = append(
							messages,
							openai.ToolMessage(
								resultContent,
								toolCall.ID,
							),
						)

					case Quit:
						// Exit the function immediately
						agent.Log.Warn(fmt.Sprintf("🛑 Quit requested for function: %s\n", functionName))
						stopped = true
						finishReason = "user_quit"
						return finishReason, results, lastAssistantMessage, nil
					}
				}

			} else {
				fmt.Println("😢 No tool calls found in response")
			}

		case "stop":
			agent.Log.Info("✋ Stopping due to 'stop' finish reason.")
			stopped = true
			lastAssistantMessage = response

			agent.Log.Info(fmt.Sprintf("🤖 %s\n", lastAssistantMessage))

			// Add final assistant message to conversation history
			messages = append(messages, openai.AssistantMessage(lastAssistantMessage))

		default:
			agent.Log.Error(fmt.Sprintf("🔴 Unexpected response: %s\n", finishReason))
			stopped = true
		}
	}
	return finishReason, results, lastAssistantMessage, nil
}
