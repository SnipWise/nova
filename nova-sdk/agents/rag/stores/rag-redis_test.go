package stores

import (
	"testing"
)

// ── parseScoreField ────────────────────────────────────────────────────────────

func TestParseScoreField_ValidString(t *testing.T) {
	// distance 0.2 → similarity 0.8
	sim, ok := parseScoreField("0.2")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if sim != 0.8 {
		t.Errorf("want 0.8, got %f", sim)
	}
}

func TestParseScoreField_ZeroDistance(t *testing.T) {
	// distance 0.0 → similarity 1.0 (perfect match)
	sim, ok := parseScoreField("0.0")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if sim != 1.0 {
		t.Errorf("want 1.0, got %f", sim)
	}
}

func TestParseScoreField_FullDistance(t *testing.T) {
	// distance 1.0 → similarity 0.0
	sim, ok := parseScoreField("1.0")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if sim != 0.0 {
		t.Errorf("want 0.0, got %f", sim)
	}
}

func TestParseScoreField_NotAString(t *testing.T) {
	_, ok := parseScoreField(42)
	if ok {
		t.Error("non-string raw: expected ok=false")
	}
}

func TestParseScoreField_InvalidFloat(t *testing.T) {
	_, ok := parseScoreField("not-a-number")
	if ok {
		t.Error("invalid float string: expected ok=false")
	}
}

func TestParseScoreField_NilRaw(t *testing.T) {
	_, ok := parseScoreField(nil)
	if ok {
		t.Error("nil raw: expected ok=false")
	}
}

// ── parseDocumentFields ────────────────────────────────────────────────────────

func TestParseDocumentFields_ID(t *testing.T) {
	record := parseDocumentFields("abc-123", []interface{}{})
	if record.Id != "abc-123" {
		t.Errorf("Id: want 'abc-123', got %q", record.Id)
	}
}

func TestParseDocumentFields_Prompt(t *testing.T) {
	fields := []interface{}{"prompt", "hello world"}
	record := parseDocumentFields("id1", fields)
	if record.Prompt != "hello world" {
		t.Errorf("Prompt: want 'hello world', got %q", record.Prompt)
	}
}

func TestParseDocumentFields_Score(t *testing.T) {
	fields := []interface{}{"score", "0.3"}
	record := parseDocumentFields("id1", fields)
	want := 1.0 - 0.3
	if record.CosineSimilarity != want {
		t.Errorf("CosineSimilarity: want %f, got %f", want, record.CosineSimilarity)
	}
}

func TestParseDocumentFields_PromptAndScore(t *testing.T) {
	fields := []interface{}{"prompt", "test prompt", "score", "0.1"}
	record := parseDocumentFields("id2", fields)
	if record.Prompt != "test prompt" {
		t.Errorf("Prompt: want 'test prompt', got %q", record.Prompt)
	}
	if record.CosineSimilarity != 0.9 {
		t.Errorf("CosineSimilarity: want 0.9, got %f", record.CosineSimilarity)
	}
}

func TestParseDocumentFields_UnknownField_Ignored(t *testing.T) {
	fields := []interface{}{"unknown_field", "some value"}
	record := parseDocumentFields("id3", fields)
	// Should not panic or alter known fields
	if record.Prompt != "" {
		t.Errorf("unknown field: Prompt should be empty, got %q", record.Prompt)
	}
}

func TestParseDocumentFields_OddFields_SafelyHandled(t *testing.T) {
	// Odd number of fields: last entry has no value — should break cleanly
	fields := []interface{}{"prompt", "hi", "score"}
	record := parseDocumentFields("id4", fields)
	if record.Prompt != "hi" {
		t.Errorf("prompt: want 'hi', got %q", record.Prompt)
	}
	// score has no value pair, CosineSimilarity stays 0
	if record.CosineSimilarity != 0 {
		t.Errorf("CosineSimilarity: want 0 for incomplete pair, got %f", record.CosineSimilarity)
	}
}

func TestParseDocumentFields_EmptyFields(t *testing.T) {
	record := parseDocumentFields("empty", []interface{}{})
	if record.Id != "empty" {
		t.Errorf("Id: want 'empty', got %q", record.Id)
	}
	if record.Prompt != "" || record.CosineSimilarity != 0 {
		t.Errorf("empty fields: expected zero-value record, got %+v", record)
	}
}
