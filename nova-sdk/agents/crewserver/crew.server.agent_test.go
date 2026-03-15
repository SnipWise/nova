package crewserver

import (
	"context"
	"testing"

	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
)

// newTestCrewAgent creates a minimal *CrewServerAgent suitable for unit tests.
// It does NOT call NewAgent — it bypasses option validation intentionally.
func newTestCrewAgent(t *testing.T) *CrewServerAgent {
	t.Helper()
	base := serverbase.NewBaseServerAgent(context.Background(), ":3500", nil, nil)
	return &CrewServerAgent{
		BaseServerAgent: base,
		chatAgents:      map[string]*chat.Agent{"agent1": nil},
	}
}

// ── NewAgent validation ───────────────────────────────────────────────────────

func TestNewAgent_MissingCrew_ReturnsError(t *testing.T) {
	_, err := NewAgent(context.Background())
	if err == nil {
		t.Error("expected error when no agent crew is provided, got nil")
	}
}

// ── applyDefaultFunctions ─────────────────────────────────────────────────────

func TestApplyDefaultFunctions_SetsMatchFn_WhenNil(t *testing.T) {
	agent := newTestCrewAgent(t)
	// matchAgentIdToTopicFn is nil on a freshly constructed agent
	agent.applyDefaultFunctions()

	if agent.matchAgentIdToTopicFn == nil {
		t.Fatal("matchAgentIdToTopicFn should be non-nil after applyDefaultFunctions")
	}
	// The default closure returns the first (and only) key in chatAgents
	result := agent.matchAgentIdToTopicFn("", "")
	if result != "agent1" {
		t.Errorf("default matchFn: expected 'agent1', got %q", result)
	}
}

func TestApplyDefaultFunctions_KeepsExistingMatchFn(t *testing.T) {
	agent := newTestCrewAgent(t)
	customFn := func(_, _ string) string { return "custom" }
	agent.matchAgentIdToTopicFn = customFn

	agent.applyDefaultFunctions()

	result := agent.matchAgentIdToTopicFn("", "")
	if result != "custom" {
		t.Errorf("existing matchFn must not be overridden, got %q", result)
	}
}

func TestApplyDefaultFunctions_SetsExecuteFn_WhenNil(t *testing.T) {
	agent := newTestCrewAgent(t)
	// ExecuteFn is nil because NewBaseServerAgent was called with nil executeFn

	agent.applyDefaultFunctions()

	if agent.ExecuteFn == nil {
		t.Error("ExecuteFn should be set to agent.executeFunction after applyDefaultFunctions")
	}
}
