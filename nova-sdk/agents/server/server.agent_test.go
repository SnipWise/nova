package server

import (
	"context"
	"testing"

	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
)

// newTestServerAgent builds a minimal *ServerAgent wired to a real BaseServerAgent.
// It bypasses NewAgent intentionally to avoid needing a live LLM endpoint.
func newTestServerAgent(t *testing.T) *ServerAgent {
	t.Helper()
	base := serverbase.NewBaseServerAgent(context.Background(), ":8080", nil, nil)
	return &ServerAgent{BaseServerAgent: base}
}

// ── applyConfigFields ─────────────────────────────────────────────────────────

func TestApplyConfigFields_NilConfigs_LeavesAgentsNil(t *testing.T) {
	agent := newTestServerAgent(t)
	agent.applyConfigFields()
	if agent.ToolsAgent != nil {
		t.Error("expected ToolsAgent to remain nil when no config was set")
	}
	if agent.RagAgent != nil {
		t.Error("expected RagAgent to remain nil when no config was set")
	}
	if agent.CompressorAgent != nil {
		t.Error("expected CompressorAgent to remain nil when no config was set")
	}
}

func TestApplyConfigFields_ExecuteFn_DefaultsToExecuteFunction(t *testing.T) {
	agent := newTestServerAgent(t)
	// ExecuteFn is nil on fresh agent — applyConfigFields should set the default
	agent.applyConfigFields()
	if agent.ExecuteFn == nil {
		t.Error("expected ExecuteFn to be set to agent.executeFunction")
	}
}

func TestApplyConfigFields_ExecuteFn_PreservesExisting(t *testing.T) {
	agent := newTestServerAgent(t)
	custom := func(name, args string) (string, error) { return "custom", nil }
	agent.ExecuteFn = custom
	agent.applyConfigFields()
	// The custom function pointer should survive; check by calling it
	result, _ := agent.ExecuteFn("", "")
	if result != "custom" {
		t.Errorf("expected custom ExecuteFn to be preserved, got result=%q", result)
	}
}

func TestApplyConfigFields_ConfirmationFn_DefaultsToCliPrompt(t *testing.T) {
	agent := newTestServerAgent(t)
	agent.applyConfigFields()
	if agent.ConfirmationPromptFn == nil {
		t.Error("expected ConfirmationPromptFn to be set to default cliConfirmationPrompt")
	}
}
