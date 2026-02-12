package gatewayserver

import (
	"encoding/json"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// --- Conversion helpers ---
// These functions convert between gateway types and OpenAI SDK types.

// convertToOpenAIMessages converts gateway ChatCompletionMessage to OpenAI SDK messages.
func (agent *GatewayServerAgent) convertToOpenAIMessages(msgs []ChatCompletionMessage) []openai.ChatCompletionMessageParamUnion {
	var result []openai.ChatCompletionMessageParamUnion

	for _, msg := range msgs {
		content := ""
		if msg.Content != nil {
			content = msg.Content.String()
		}

		switch msg.Role {
		case "system":
			result = append(result, openai.SystemMessage(content))
		case "user":
			result = append(result, openai.UserMessage(content))
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls
				toolCallParams := make([]openai.ChatCompletionMessageToolCallUnionParam, len(msg.ToolCalls))
				for i, tc := range msg.ToolCalls {
					toolCallParams[i] = openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID: tc.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						},
					}
				}
				result = append(result, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						Content:   openai.ChatCompletionAssistantMessageParamContentUnion{OfString: openai.Opt(content)},
						ToolCalls: toolCallParams,
					},
				})
			} else {
				result = append(result, openai.AssistantMessage(content))
			}
		case "tool":
			result = append(result, openai.ToolMessage(content, msg.ToolCallID))
		case "developer":
			result = append(result, openai.DeveloperMessage(content))
		}
	}

	return result
}

// convertToOpenAITools converts gateway ToolDefinition to OpenAI SDK tool params.
func (agent *GatewayServerAgent) convertToOpenAITools(toolDefs []ToolDefinition) []openai.ChatCompletionToolUnionParam {
	if len(toolDefs) == 0 {
		return nil
	}

	var result []openai.ChatCompletionToolUnionParam
	for _, td := range toolDefs {
		// Convert parameters to shared.FunctionParameters (map[string]any)
		var params shared.FunctionParameters
		if td.Function.Parameters != nil {
			// Marshal then unmarshal to convert any type to map[string]any
			paramsJSON, _ := json.Marshal(td.Function.Parameters)
			_ = json.Unmarshal(paramsJSON, &params)
		}

		result = append(result, openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        td.Function.Name,
			Description: openai.String(td.Function.Description),
			Parameters:  params,
		}))
	}

	return result
}

// convertFromOpenAIToolCalls converts OpenAI SDK tool calls to gateway ToolCall format.
func (agent *GatewayServerAgent) convertFromOpenAIToolCalls(openaiCalls []openai.ChatCompletionMessageToolCallUnion) []ToolCall {
	result := make([]ToolCall, len(openaiCalls))
	for i, tc := range openaiCalls {
		idx := i
		result[i] = ToolCall{
			Index: &idx,
			ID:    tc.ID,
			Type:  "function",
			Function: FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return result
}

// convertFromOpenAIStreamToolCalls converts streaming tool call deltas.
func (agent *GatewayServerAgent) convertFromOpenAIStreamToolCalls(deltas []openai.ChatCompletionChunkChoiceDeltaToolCall) []ToolCall {
	result := make([]ToolCall, len(deltas))
	for i, d := range deltas {
		idx := int(d.Index)
		result[i] = ToolCall{
			Index: &idx,
			ID:    d.ID,
			Type:  "function",
			Function: FunctionCall{
				Name:      d.Function.Name,
				Arguments: d.Function.Arguments,
			},
		}
	}
	return result
}

// --- Helper functions ---

// buildToolCallHistory creates a message history for tool detection in server-side execution mode.
func (agent *GatewayServerAgent) buildToolCallHistory(question string) []messages.Message {
	history := agent.currentChatAgent.GetMessages()
	return append(history, messages.Message{
		Role:    roles.User,
		Content: question,
	})
}
