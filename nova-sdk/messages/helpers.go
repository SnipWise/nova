package messages

import (
	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// ConvertToOpenAIMessage converts a single message to OpenAI format
func ConvertToOpenAIMessage(message Message) openai.ChatCompletionMessageParamUnion {
	switch message.Role {
	case roles.System:
		return openai.SystemMessage(message.Content)
	case roles.User:
		return openai.UserMessage(message.Content)
	case roles.Assistant:
		return openai.AssistantMessage(message.Content)
	case roles.Developer:
		return openai.DeveloperMessage(message.Content)
	default:
		// Default to user message for unknown roles
		return openai.UserMessage(message.Content)
	}
}

// ConvertToOpenAIMessages converts simplified messages to OpenAI format
func ConvertToOpenAIMessages(messages []Message) []openai.ChatCompletionMessageParamUnion {
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		openaiMessages = append(openaiMessages, ConvertToOpenAIMessage(msg))
	}

	return openaiMessages
}

func ConvertFromOpenAIMessages(openaiMessages []openai.ChatCompletionMessageParamUnion) []Message {

	stringMessages := []Message{}

	for _, msg := range openaiMessages {
		var role roles.Role
		var content string

		// Determine the role
		if msg.OfSystem != nil {
			role = roles.System
			content = msg.OfSystem.Content.OfString.Value
		} else if msg.OfUser != nil {
			role = roles.User
			content = msg.OfUser.Content.OfString.Value
		} else if msg.OfAssistant != nil {
			role = roles.Assistant
			content = msg.OfAssistant.Content.OfString.Value
			// } else if msg.OfTool != nil {
			// 	role = "tool"
			// } else if msg.OfFunction != nil {
			// 	role = "function"
		} else if msg.OfDeveloper != nil {
			role = roles.Developer
			content = msg.OfDeveloper.Content.OfString.Value
		} else {
			role = "unknown"
			content = "Unknown message type"
		}

		stringMessages = append(stringMessages, Message{
			Role:    role,
			Content: content,
		})
	}

	return stringMessages
}
