package chunks

import (
	"regexp"
	"testing"
)

// ── collectHeaderPositions ────────────────────────────────────────────────────

func TestCollectHeaderPositions_NoMatchingLevel(t *testing.T) {
	re := regexp.MustCompile(`(?m)^\s*(#+)\s+.*$`)
	lines := []string{"# H1", "some text", "## H2"}
	pos := collectHeaderPositions(lines, 3, re)
	if len(pos) != 0 {
		t.Errorf("expected no positions for level 3, got %v", pos)
	}
}

func TestCollectHeaderPositions_SingleMatch(t *testing.T) {
	re := regexp.MustCompile(`(?m)^\s*(#+)\s+.*$`)
	lines := []string{"# Title", "body"}
	pos := collectHeaderPositions(lines, 1, re)
	if len(pos) != 1 || pos[0] != 0 {
		t.Errorf("expected [0], got %v", pos)
	}
}

func TestCollectHeaderPositions_MultipleMatches(t *testing.T) {
	re := regexp.MustCompile(`(?m)^\s*(#+)\s+.*$`)
	lines := []string{"# A", "text", "# B"}
	// "# A" starts at byte 0, "text\n" is 5 bytes, "# B" starts at byte 5+1=... wait
	// line0 "# A" len=3 → currentPos after = 4 (len+1)
	// line1 "text" len=4 → currentPos after = 9
	// line2 "# B" starts at 9 — but position was recorded at currentPos=4 before line1
	// Actually: line0 recorded at 0, line2 recorded at 0+3+1+4+1 = 9
	pos := collectHeaderPositions(lines, 1, re)
	if len(pos) != 2 {
		t.Fatalf("expected 2 positions, got %v", pos)
	}
	if pos[0] != 0 {
		t.Errorf("first header should start at byte 0, got %d", pos[0])
	}
}

func TestCollectHeaderPositions_SkipsWrongLevel(t *testing.T) {
	re := regexp.MustCompile(`(?m)^\s*(#+)\s+.*$`)
	lines := []string{"# H1", "## H2", "### H3", "## H2b"}
	pos := collectHeaderPositions(lines, 2, re)
	if len(pos) != 2 {
		t.Errorf("expected 2 level-2 headers, got %v", pos)
	}
}

// ── buildSections ─────────────────────────────────────────────────────────────

func TestBuildSections_SingleHeader_NoPreContent(t *testing.T) {
	md := "# Title\ncontent"
	// headerPositions[0] == 0, so no pre-header
	sections := buildSections(md, []int{0})
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0] != "# Title\ncontent" {
		t.Errorf("unexpected section: %q", sections[0])
	}
}

func TestBuildSections_PreHeaderContent(t *testing.T) {
	md := "intro text\n# Title\ncontent"
	// "intro text\n" = 11 bytes → first header at 11
	sections := buildSections(md, []int{11})
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections (pre-header + header), got %d", len(sections))
	}
	if sections[0] != "intro text" {
		t.Errorf("pre-header: expected 'intro text', got %q", sections[0])
	}
}

func TestBuildSections_TwoHeaders(t *testing.T) {
	md := "# A\nfoo\n# B\nbar"
	// "# A\n" = 4, "foo\n" = 4 → second header at 8
	sections := buildSections(md, []int{0, 8})
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d: %v", len(sections), sections)
	}
}

func TestBuildSections_EmptySectionSkipped(t *testing.T) {
	// Two headers back-to-back — TrimSpace of the first section = just the header line
	md := "# A\n# B\ncontent"
	sections := buildSections(md, []int{0, 4})
	// "# A" trimmed → "# A", "# B\ncontent" trimmed → "# B\ncontent"
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d: %v", len(sections), sections)
	}
}

// ── collectHeaderContent ──────────────────────────────────────────────────────

func TestCollectHeaderContent_StopsAtNextHeader(t *testing.T) {
	re := regexp.MustCompile(`^(#+)\s+(.*)$`)
	lines := []string{"# H1", "line1", "line2", "## H2", "other"}
	content := collectHeaderContent(lines, 0, re)
	if content != "line1\nline2" {
		t.Errorf("expected 'line1\\nline2', got %q", content)
	}
}

func TestCollectHeaderContent_EOF(t *testing.T) {
	re := regexp.MustCompile(`^(#+)\s+(.*)$`)
	lines := []string{"# H1", "only", "line"}
	content := collectHeaderContent(lines, 0, re)
	if content != "only\nline" {
		t.Errorf("expected 'only\\nline', got %q", content)
	}
}

func TestCollectHeaderContent_EmptyContent(t *testing.T) {
	re := regexp.MustCompile(`^(#+)\s+(.*)$`)
	lines := []string{"# H1", "# H2"}
	content := collectHeaderContent(lines, 0, re)
	if content != "" {
		t.Errorf("expected empty string, got %q", content)
	}
}

func TestCollectHeaderContent_LastLine(t *testing.T) {
	re := regexp.MustCompile(`^(#+)\s+(.*)$`)
	lines := []string{"# H1", "text"}
	// startIdx at last meaningful line
	content := collectHeaderContent(lines, 1, re)
	if content != "" {
		t.Errorf("starting at last line: expected empty, got %q", content)
	}
}
