package messages

import (
	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova/roles"
)

// ConvertToOpenAIMessages converts simplified messages to OpenAI format
func ConvertToOpenAIMessages(messages []Message) []openai.ChatCompletionMessageParamUnion {
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case roles.System:
			openaiMessages = append(openaiMessages, openai.SystemMessage(msg.Content))
		case roles.User:
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		case roles.Assistant:
			openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
		case roles.Developer:
			openaiMessages = append(openaiMessages, openai.DeveloperMessage(msg.Content))
		}
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
