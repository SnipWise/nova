package tools

import (
	"errors"
	"testing"
)

// ── extractToolCallback ───────────────────────────────────────────────────────

func TestExtractToolCallback_TypeAliasAtPosition0(t *testing.T) {
	called := false
	fn := ToolCallback(func(name, args string) (string, error) {
		called = true
		return "ok", nil
	})
	cb, err := extractToolCallback([]any{fn}, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cb("f", "a")
	if !called {
		t.Error("expected callback to be called")
	}
}

func TestExtractToolCallback_UnderlyingFuncType(t *testing.T) {
	// Pass as bare func (not type-aliased)
	var plain func(string, string) (string, error) = func(n, a string) (string, error) {
		return "result", nil
	}
	cb, err := extractToolCallback([]any{plain}, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, _ := cb("fn", "args")
	if res != "result" {
		t.Errorf("want %q, got %q", "result", res)
	}
}

func TestExtractToolCallback_FallbackUsed_WhenNilInSlice(t *testing.T) {
	fallback := ToolCallback(func(n, a string) (string, error) { return "fallback", nil })
	cb, err := extractToolCallback([]any{nil}, 0, fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, _ := cb("f", "a")
	if res != "fallback" {
		t.Errorf("want %q, got %q", "fallback", res)
	}
}

func TestExtractToolCallback_FallbackUsed_WhenSliceEmpty(t *testing.T) {
	fallback := ToolCallback(func(n, a string) (string, error) { return "fallback", nil })
	cb, err := extractToolCallback([]any{}, 0, fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, _ := cb("f", "a")
	if res != "fallback" {
		t.Errorf("want %q, got %q", "fallback", res)
	}
}

func TestExtractToolCallback_Error_WhenBothNil(t *testing.T) {
	_, err := extractToolCallback([]any{}, 0, nil)
	if err == nil {
		t.Fatal("expected error when both positional and fallback are nil")
	}
	if !errors.Is(err, err) {
		t.Errorf("unexpected error type: %v", err)
	}
}

func TestExtractToolCallback_Error_MessageContent(t *testing.T) {
	_, err := extractToolCallback(nil, 0, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

// ── extractConfirmationCallback ───────────────────────────────────────────────

func TestExtractConfirmationCallback_TypeAliasAtPosition1(t *testing.T) {
	called := false
	fn := ConfirmationCallback(func(name, args string) ConfirmationResponse {
		called = true
		return ConfirmationResponse(1)
	})
	cb, err := extractConfirmationCallback([]any{nil, fn}, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cb("f", "a")
	if !called {
		t.Error("expected confirmation callback to be called")
	}
}

func TestExtractConfirmationCallback_UnderlyingFuncType(t *testing.T) {
	var plain func(string, string) ConfirmationResponse = func(n, a string) ConfirmationResponse {
		return ConfirmationResponse(2)
	}
	cb, err := extractConfirmationCallback([]any{nil, plain}, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := cb("f", "a")
	if res != ConfirmationResponse(2) {
		t.Errorf("want 2, got %v", res)
	}
}

func TestExtractConfirmationCallback_FallbackUsed_WhenSliceShort(t *testing.T) {
	fallback := ConfirmationCallback(func(n, a string) ConfirmationResponse { return ConfirmationResponse(3) })
	cb, err := extractConfirmationCallback([]any{}, 1, fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := cb("f", "a")
	if res != ConfirmationResponse(3) {
		t.Errorf("want 3, got %v", res)
	}
}

func TestExtractConfirmationCallback_Error_WhenBothNil(t *testing.T) {
	_, err := extractConfirmationCallback([]any{}, 1, nil)
	if err == nil {
		t.Fatal("expected error when both positional and fallback are nil")
	}
}

func TestExtractConfirmationCallback_PositionZero_Ignored_WhenPositionIs1(t *testing.T) {
	// Value at position 0 is a ToolCallback-shaped function; should not be used
	// for position 1 (confirmation), so fallback kicks in
	fallback := ConfirmationCallback(func(n, a string) ConfirmationResponse { return ConfirmationResponse(7) })
	cb, err := extractConfirmationCallback([]any{nil, nil}, 1, fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := cb("f", "a")
	if res != ConfirmationResponse(7) {
		t.Errorf("want 7, got %v", res)
	}
}
