package remote

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/messages"
)

const (
	errServerStatus   = "server returned status %d: %s"
	errMarshalRequest = "failed to marshal request: %w"
	errCreateRequest  = "failed to create request: %w"
	errSendRequest    = "failed to send request: %w"
	remoteContentType = "Content-Type"
	remoteMIMEJSON    = "application/json"
)

// buildMessageContent combines multiple messages into a single newline-separated string.
func buildMessageContent(userMessages []messages.Message) string {
	contents := make([]string, len(userMessages))
	for i, msg := range userMessages {
		contents[i] = msg.Content
	}
	return strings.Join(contents, "\n")
}

// processSingleSSELine parses one line from the SSE stream.
// Returns nil to continue the scanning loop, or an error to stop immediately.
func (a *Agent) processSingleSSELine(
	line string,
	callback StreamCallback,
	fullResponse *strings.Builder,
	lastFinishReason *string,
) error {
	if !strings.HasPrefix(line, "data: ") {
		return nil
	}
	data := strings.TrimPrefix(line, "data: ")
	if data == "" {
		return nil
	}

	var event map[string]interface{}
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		a.log.Error("Failed to parse SSE event: %v", err)
		return nil
	}
	if kind, ok := event["kind"].(string); ok && kind == "tool_call" {
		return handleSSEToolCall(a, event)
	}
	if err := handleSSEChunk(event, callback, fullResponse, lastFinishReason); err != nil {
		return err
	}
	if errMsg, ok := event["error"].(string); ok {
		return fmt.Errorf("server error: %s", errMsg)
	}
	return nil
}
