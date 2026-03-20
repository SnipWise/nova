package compressor

import (
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/messages"
)

// buildConversationText converts a list of OpenAI messages into a plain-text
// transcript suitable for the compression prompt.
func buildConversationText(messagesList []openai.ChatCompletionMessageParamUnion) string {
	var textBuilder strings.Builder
	for _, msg := range messages.ConvertFromOpenAIMessages(messagesList) {
		textBuilder.WriteString(fmt.Sprintf("%s: ", msg.Role))
		textBuilder.WriteString(msg.Content)
		textBuilder.WriteString("\n")
	}
	return textBuilder.String()
}
