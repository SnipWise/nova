package tools

import (
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/base"
)

// newTestBaseAgent creates a minimal BaseAgent for unit testing helpers.
func newTestBaseAgent(keepHistory bool) *BaseAgent {
	return &BaseAgent{
		Agent: &base.Agent{
			Config: agents.Config{
				KeepConversationHistory: keepHistory,
			},
		},
	}
}

// ── saveHistoryIfNeeded ────────────────────────────────────────────────────────

func TestSaveHistoryIfNeeded_StoresMessages_WhenEnabled(t *testing.T) {
	agent := newTestBaseAgent(true)
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("hello"),
	}
	agent.saveHistoryIfNeeded(msgs)
	if len(agent.ChatCompletionParams.Messages) != 1 {
		t.Errorf("expected 1 message stored, got %d", len(agent.ChatCompletionParams.Messages))
	}
}

func TestSaveHistoryIfNeeded_SkipsMessages_WhenDisabled(t *testing.T) {
	agent := newTestBaseAgent(false)
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("hello"),
	}
	agent.saveHistoryIfNeeded(msgs)
	if len(agent.ChatCompletionParams.Messages) != 0 {
		t.Errorf("expected 0 messages (history disabled), got %d", len(agent.ChatCompletionParams.Messages))
	}
}

func TestSaveHistoryIfNeeded_StoresMultipleMessages(t *testing.T) {
	agent := newTestBaseAgent(true)
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("first"),
		openai.UserMessage("second"),
		openai.UserMessage("third"),
	}
	agent.saveHistoryIfNeeded(msgs)
	if len(agent.ChatCompletionParams.Messages) != 3 {
		t.Errorf("expected 3 messages stored, got %d", len(agent.ChatCompletionParams.Messages))
	}
}

func TestSaveHistoryIfNeeded_DoesNotModify_WhenDisabledAndEmpty(t *testing.T) {
	agent := newTestBaseAgent(false)
	agent.saveHistoryIfNeeded(nil)
	if agent.ChatCompletionParams.Messages != nil {
		t.Error("expected Messages to remain nil when history is disabled")
	}
}
