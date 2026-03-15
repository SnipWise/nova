package remote

import (
	"errors"
	"strings"
	"testing"

	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// newTestAgent returns a minimal Agent suitable for unit tests.
func newTestAgent() *Agent {
	return &Agent{log: logger.GetLoggerFromEnv()}
}

// ── handleSSEToolCall ─────────────────────────────────────────────────────────

func TestHandleSSEToolCall_NoCallback_NoError(t *testing.T) {
	a := newTestAgent()
	event := map[string]interface{}{
		"kind":         "tool_call",
		"operation_id": "op-123",
		"message":      "Running tool X",
	}
	if err := handleSSEToolCall(a, event); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestHandleSSEToolCall_WithCallback_Called(t *testing.T) {
	a := newTestAgent()
	var gotID, gotMsg string
	a.toolCallCallback = func(id, msg string) error {
		gotID = id
		gotMsg = msg
		return nil
	}
	event := map[string]interface{}{
		"operation_id": "op-456",
		"message":      "Executing search",
	}
	if err := handleSSEToolCall(a, event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != "op-456" {
		t.Errorf("operation_id: want %q, got %q", "op-456", gotID)
	}
	if gotMsg != "Executing search" {
		t.Errorf("message: want %q, got %q", "Executing search", gotMsg)
	}
}

func TestHandleSSEToolCall_CallbackReturnsError_Propagated(t *testing.T) {
	a := newTestAgent()
	a.toolCallCallback = func(id, msg string) error {
		return errors.New("callback failed")
	}
	err := handleSSEToolCall(a, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error from callback to be propagated")
	}
	if err.Error() != "callback failed" {
		t.Errorf("want %q, got %q", "callback failed", err.Error())
	}
}

func TestHandleSSEToolCall_MissingFields_NoError(t *testing.T) {
	// Empty event — operation_id and message should default to ""
	a := newTestAgent()
	if err := handleSSEToolCall(a, map[string]interface{}{}); err != nil {
		t.Errorf("empty event: expected no error, got %v", err)
	}
}

// ── handleSSEChunk ────────────────────────────────────────────────────────────

func TestHandleSSEChunk_AppendsToResponse(t *testing.T) {
	var resp strings.Builder
	finishReason := ""
	event := map[string]interface{}{"message": "hello "}
	if err := handleSSEChunk(event, nil, &resp, &finishReason); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.String() != "hello " {
		t.Errorf("want %q, got %q", "hello ", resp.String())
	}
}

func TestHandleSSEChunk_CallsCallback(t *testing.T) {
	var got string
	cb := StreamCallback(func(chunk, _ string) error {
		got = chunk
		return nil
	})
	var resp strings.Builder
	fin := ""
	event := map[string]interface{}{"message": "world"}
	if err := handleSSEChunk(event, cb, &resp, &fin); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "world" {
		t.Errorf("callback: want %q, got %q", "world", got)
	}
}

func TestHandleSSEChunk_SetsFinishReason(t *testing.T) {
	var resp strings.Builder
	fin := ""
	event := map[string]interface{}{
		"message":       "last chunk",
		"finish_reason": "stop",
	}
	if err := handleSSEChunk(event, nil, &resp, &fin); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fin != "stop" {
		t.Errorf("finish_reason: want %q, got %q", "stop", fin)
	}
}

func TestHandleSSEChunk_CallbackError_Propagated(t *testing.T) {
	cb := StreamCallback(func(chunk, _ string) error { return errors.New("stream error") })
	var resp strings.Builder
	fin := ""
	event := map[string]interface{}{"message": "oops"}
	err := handleSSEChunk(event, cb, &resp, &fin)
	if err == nil {
		t.Fatal("expected error to be propagated")
	}
}

func TestHandleSSEChunk_NoMessageKey_NoOp(t *testing.T) {
	called := false
	cb := StreamCallback(func(chunk, _ string) error { called = true; return nil })
	var resp strings.Builder
	fin := ""
	event := map[string]interface{}{"finish_reason": "stop"} // no "message"
	if err := handleSSEChunk(event, cb, &resp, &fin); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("callback should not be called when there is no message key")
	}
	if resp.Len() != 0 {
		t.Errorf("response should remain empty, got %q", resp.String())
	}
}

func TestHandleSSEChunk_MultipleChunks_Accumulate(t *testing.T) {
	var resp strings.Builder
	fin := ""
	for _, chunk := range []string{"foo", "bar", "baz"} {
		event := map[string]interface{}{"message": chunk}
		if err := handleSSEChunk(event, nil, &resp, &fin); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if resp.String() != "foobarbaz" {
		t.Errorf("accumulated: want %q, got %q", "foobarbaz", resp.String())
	}
}
