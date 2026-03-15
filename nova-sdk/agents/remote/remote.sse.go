package remote

import "strings"

// handleSSEToolCall processes a "tool_call" kind SSE event.
// It logs the operation details and invokes the agent's toolCallCallback if set.
func handleSSEToolCall(a *Agent, event map[string]interface{}) error {
	operationID := ""
	if id, ok := event["operation_id"].(string); ok {
		operationID = id
	}
	message := ""
	if msg, ok := event["message"].(string); ok {
		message = msg
	}

	a.log.Info("\n🔔 Tool Call Detected: %s", message)
	a.log.Info("📝 Operation ID: %s", operationID)
	a.log.Info("✅ To validate: curl -X POST http://localhost:8080/operation/validate -d '{\"operation_id\":\"%s\"}'", operationID)
	a.log.Info("⛔️ To cancel:   curl -X POST http://localhost:8080/operation/cancel -d '{\"operation_id\":\"%s\"}'\n", operationID)

	if a.toolCallCallback != nil {
		return a.toolCallCallback(operationID, message)
	}
	return nil
}

// handleSSEChunk processes a message-chunk SSE event.
// It appends the chunk to resp, updates finishReason, and forwards the chunk to cb.
func handleSSEChunk(event map[string]interface{}, cb StreamCallback, resp *strings.Builder, finishReason *string) error {
	message, ok := event["message"].(string)
	if !ok {
		return nil
	}
	resp.WriteString(message)

	fr := ""
	if v, ok := event["finish_reason"].(string); ok {
		fr = v
		*finishReason = v
	}

	if cb != nil {
		return cb(message, fr)
	}
	return nil
}
