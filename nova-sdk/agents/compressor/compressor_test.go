package compressor

import (
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
)

// buildConversationText is tested with nil / empty input since constructing
// populated openai.ChatCompletionMessageParamUnion values requires no mock infrastructure.

func TestBuildConversationText_NilInput_ReturnsEmpty(t *testing.T) {
	result := buildConversationText(nil)
	if result != "" {
		t.Errorf("expected empty string for nil input, got %q", result)
	}
}

func TestBuildConversationText_EmptySlice_ReturnsEmpty(t *testing.T) {
	result := buildConversationText([]openai.ChatCompletionMessageParamUnion{})
	if result != "" {
		t.Errorf("expected empty string for empty slice, got %q", result)
	}
}

func TestBuildConversationText_NilInput_NoTrailingNewline(t *testing.T) {
	result := buildConversationText(nil)
	if strings.HasSuffix(result, "\n") {
		t.Error("empty input must not produce a trailing newline")
	}
}
