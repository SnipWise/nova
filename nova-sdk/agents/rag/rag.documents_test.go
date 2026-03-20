package rag

import (
	"errors"
	"testing"

	"github.com/snipwise/nova/nova-sdk/agents/rag/stores"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// stubStore is a minimal VectorStore used in tests.
type stubStore struct {
	records     []stores.VectorRecord
	resetCalled bool
	resetErr    error
}

func (s *stubStore) GetAll() ([]stores.VectorRecord, error) { return s.records, nil }
func (s *stubStore) Save(r stores.VectorRecord) (stores.VectorRecord, error) {
	s.records = append(s.records, r)
	return r, nil
}
func (s *stubStore) SearchSimilarities(r stores.VectorRecord, limit float64) ([]stores.VectorRecord, error) {
	return nil, nil
}
func (s *stubStore) SearchTopNSimilarities(r stores.VectorRecord, limit float64, max int) ([]stores.VectorRecord, error) {
	return nil, nil
}
func (s *stubStore) ResetMemory() error {
	s.resetCalled = true
	if s.resetErr != nil {
		return s.resetErr
	}
	s.records = nil
	return nil
}

// newTestBaseAgent builds a minimal BaseAgent wired to the given store.
func newTestBaseAgent(store stores.VectorStore) *BaseAgent {
	return &BaseAgent{
		store: store,
		log:   logger.GetLoggerFromEnv(),
	}
}

// ── applyLoadModePolicy ───────────────────────────────────────────────────────

func TestApplyLoadModePolicy_Merge_EmptyStore_Proceeds(t *testing.T) {
	a := newTestBaseAgent(&stubStore{})
	if !applyLoadModePolicy(a, DocumentLoadModeMerge, false) {
		t.Error("Merge with empty store should return true (proceed)")
	}
}

func TestApplyLoadModePolicy_Merge_WithData_Proceeds(t *testing.T) {
	a := newTestBaseAgent(&stubStore{})
	if !applyLoadModePolicy(a, DocumentLoadModeMerge, true) {
		t.Error("Merge with existing data should still return true")
	}
}

func TestApplyLoadModePolicy_Skip_EmptyStore_Proceeds(t *testing.T) {
	a := newTestBaseAgent(&stubStore{})
	if !applyLoadModePolicy(a, DocumentLoadModeSkip, false) {
		t.Error("Skip with empty store should proceed")
	}
}

func TestApplyLoadModePolicy_Skip_WithData_Stops(t *testing.T) {
	a := newTestBaseAgent(&stubStore{})
	if applyLoadModePolicy(a, DocumentLoadModeSkip, true) {
		t.Error("Skip with existing data should return false (stop)")
	}
}

func TestApplyLoadModePolicy_Error_EmptyStore_Proceeds(t *testing.T) {
	a := newTestBaseAgent(&stubStore{})
	if !applyLoadModePolicy(a, DocumentLoadModeError, false) {
		t.Error("Error mode with empty store should proceed")
	}
}

func TestApplyLoadModePolicy_Error_WithData_Stops(t *testing.T) {
	a := newTestBaseAgent(&stubStore{})
	if applyLoadModePolicy(a, DocumentLoadModeError, true) {
		t.Error("Error mode with existing data should return false")
	}
}

func TestApplyLoadModePolicy_SkipDuplicates_EmptyStore_Proceeds(t *testing.T) {
	a := newTestBaseAgent(&stubStore{})
	if !applyLoadModePolicy(a, DocumentLoadModeSkipDuplicates, false) {
		t.Error("SkipDuplicates with empty store should proceed")
	}
}

func TestApplyLoadModePolicy_SkipDuplicates_WithData_Proceeds(t *testing.T) {
	a := newTestBaseAgent(&stubStore{})
	// SkipDuplicates does per-document checks — the pre-check always proceeds
	if !applyLoadModePolicy(a, DocumentLoadModeSkipDuplicates, true) {
		t.Error("SkipDuplicates with existing data should still proceed (per-document check)")
	}
}

func TestApplyLoadModePolicy_Overwrite_EmptyStore_Proceeds(t *testing.T) {
	s := &stubStore{}
	a := newTestBaseAgent(s)
	if !applyLoadModePolicy(a, DocumentLoadModeOverwrite, false) {
		t.Error("Overwrite with empty store should proceed without reset")
	}
	if s.resetCalled {
		t.Error("ResetMemory should not be called when store is empty")
	}
}

func TestApplyLoadModePolicy_Overwrite_WithData_ResetsAndProceeds(t *testing.T) {
	s := &stubStore{records: []stores.VectorRecord{{Prompt: "existing"}}}
	a := newTestBaseAgent(s)
	if !applyLoadModePolicy(a, DocumentLoadModeOverwrite, true) {
		t.Error("Overwrite with data should return true after reset")
	}
	if !s.resetCalled {
		t.Error("ResetMemory should have been called")
	}
}

func TestApplyLoadModePolicy_Overwrite_ResetError_Stops(t *testing.T) {
	s := &stubStore{
		records:  []stores.VectorRecord{{Prompt: "x"}},
		resetErr: errors.New("reset failed"),
	}
	a := newTestBaseAgent(s)
	if applyLoadModePolicy(a, DocumentLoadModeOverwrite, true) {
		t.Error("Overwrite with reset error should return false")
	}
}

// ── handleOverwriteMode ───────────────────────────────────────────────────────

// noResetStore implements VectorStore without the optional ResetMemory method.
type noResetStore struct{}

func (s *noResetStore) GetAll() ([]stores.VectorRecord, error)                  { return nil, nil }
func (s *noResetStore) Save(r stores.VectorRecord) (stores.VectorRecord, error) { return r, nil }
func (s *noResetStore) SearchSimilarities(r stores.VectorRecord, _ float64) ([]stores.VectorRecord, error) {
	return nil, nil
}
func (s *noResetStore) SearchTopNSimilarities(r stores.VectorRecord, _ float64, _ int) ([]stores.VectorRecord, error) {
	return nil, nil
}

func TestHandleOverwriteMode_EmptyStore_Proceeds(t *testing.T) {
	a := newTestBaseAgent(&stubStore{})
	if !handleOverwriteMode(a, false) {
		t.Error("handleOverwriteMode with empty store should return true")
	}
}

func TestHandleOverwriteMode_WithData_ResetsAndProceeds(t *testing.T) {
	s := &stubStore{records: []stores.VectorRecord{{Prompt: "doc"}}}
	a := newTestBaseAgent(s)
	if !handleOverwriteMode(a, true) {
		t.Error("handleOverwriteMode with resettable store should return true")
	}
	if !s.resetCalled {
		t.Error("ResetMemory should have been called")
	}
}

func TestHandleOverwriteMode_ResetError_Stops(t *testing.T) {
	s := &stubStore{
		records:  []stores.VectorRecord{{Prompt: "doc"}},
		resetErr: errors.New("disk full"),
	}
	a := newTestBaseAgent(s)
	if handleOverwriteMode(a, true) {
		t.Error("handleOverwriteMode with reset error should return false")
	}
}

func TestHandleOverwriteMode_NoResetSupport_Stops(t *testing.T) {
	a := newTestBaseAgent(&noResetStore{})
	if handleOverwriteMode(a, true) {
		t.Error("handleOverwriteMode on non-resettable store should return false")
	}
}
