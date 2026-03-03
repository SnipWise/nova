---
name: qlty-check
description: Analyzes code quality of a project with the qlty CLI, generates a detailed report with explanations, and saves data to compare with future checks.
---

# Qlty Code Quality Check Skill

## Description

Performs a comprehensive code quality analysis of a project or directory using the qlty CLI. Generates a structured report with explanations, saves timestamped results, and allows tracking quality evolution over time.

## When to Use

- `/qlty-check` : to run a code quality analysis
- When measuring quality before/after a refactoring
- To track code quality degradation or improvement over time
- Before a PR or a release

## Instructions

You are a code quality analysis skill using the qlty tool. Follow these phases in order.

---

### Phase 1: Context Collection

Use **AskUserQuestion** to ask:

1. **Target directory**: Which directory to analyze? (default: current directory)
2. **Analysis type**: What do you want to analyze?
   - `all`: Lint + Metrics + Smells (recommended)
   - `lint`: Lint issues only (`qlty check --all`)
   - `metrics`: Complexity metrics only (`qlty metrics --all`)
   - `smells`: Code smells only (`qlty smells --all`)

---

### Phase 2: Environment Check

Verify that qlty is available:

```bash
qlty --version
```

If the command fails:
- Display: `qlty is not installed. Install it with: curl https://qlty.sh | sh`
- **STOP** and wait for the user to install qlty before continuing

Check if qlty is initialized in the target directory (look for `.qlty/qlty.toml`):

```bash
ls [target_directory]/.qlty/qlty.toml 2>/dev/null && echo "CONFIGURED" || echo "NOT_CONFIGURED"
```

If **NOT_CONFIGURED**:
- Ask via **AskUserQuestion**: "qlty is not initialized in this directory. Do you want to initialize it now? (`qlty init` will be executed)"
- If yes → run `cd [target_directory] && qlty init`
- If no → **STOP** and explain that qlty must be initialized

---

### Phase 3: Running the Analysis

Determine the reports save directory:

```
[target_directory]/.qlty-reports/
```

Create this directory if it doesn't exist:

```bash
mkdir -p [target_directory]/.qlty-reports
```

Generate a timestamp for this report:

```bash
date +"%Y%m%d_%H%M%S"
```

Run the requested analyses. Use `--json` when available for structured data. For each command, capture the output for the report.

**Lint check:**
```bash
cd [target_directory] && qlty check --all 2>&1
```

**Metrics:**
```bash
cd [target_directory] && qlty metrics --all --max-depth=2 --sort complexity --limit 20 2>&1
```

**Code smells:**
```bash
cd [target_directory] && qlty smells --all 2>&1
```

---

### Phase 4: Report Generation

Create a timestamped Markdown report file:

**File name**: `[target_directory]/.qlty-reports/report_[TIMESTAMP].md`

**Report structure**:

```markdown
# Quality Report — [PROJECT_NAME]
**Date**: [DATE_TIME]
**Analyzed directory**: [PATH]
**Tool**: qlty [VERSION]

---

## Executive Summary

| Category | Result |
|----------|--------|
| Lint issues | [COUNT] |
| Code smells | [COUNT] |
| Max complexity | [VALUE] |
| Files analyzed | [COUNT] |

### Overall verdict
[GOOD/WARNING/CRITICAL based on issue count]
- ✅ GOOD: 0-5 lint issues, 0-3 smells
- ⚠️ WARNING: 6-20 lint issues, 4-10 smells
- ❌ CRITICAL: >20 lint issues or >10 smells

---

## Detail: Lint & Issues

[Paste formatted output of `qlty check --all`]

### Explanation of main issues
For each issue category found, explain:
- What it is
- Why it is problematic
- How to fix it (suggestion)

---

## Detail: Complexity Metrics

[Paste formatted output of `qlty metrics`]

### Metrics explanation
- **Complexity**: Cyclomatic complexity — number of distinct execution paths. >10 = hard to test, >20 = must refactor
- **Lines**: Line count — functions >50 lines should be split
- **Functions**: Function density per file

### Top 3 most complex files
[List and explain why they are problematic]

---

## Detail: Code Smells

[Paste formatted output of `qlty smells --all`]

### Explanation of main smells
For each smell found, explain:
- The smell type (duplication, long method, etc.)
- The impact on maintainability
- The recommended refactoring strategy

---

## Priority Recommendations

### 🔴 High Priority (fix first)
[Critical or blocking issues]

### 🟡 Medium Priority (next iteration)
[Smells and high complexity]

### 🟢 Low Priority (continuous improvement)
[Small style and convention improvements]

---

## Raw JSON Data

\`\`\`json
{
  "timestamp": "[TIMESTAMP]",
  "project": "[PROJECT_NAME]",
  "directory": "[PATH]",
  "qlty_version": "[VERSION]",
  "lint": {
    "total_issues": [COUNT],
    "by_severity": {
      "error": [COUNT],
      "warning": [COUNT],
      "info": [COUNT]
    }
  },
  "metrics": {
    "max_complexity": [VALUE],
    "avg_complexity": [VALUE],
    "files_analyzed": [COUNT]
  },
  "smells": {
    "total": [COUNT],
    "by_type": {}
  }
}
\`\`\`
```

Also save the raw data in a JSON file for comparison:

**File name**: `[target_directory]/.qlty-reports/data_[TIMESTAMP].json`

---

### Phase 5: Comparison with Previous Report

Look for the previous report:

```bash
ls -t [target_directory]/.qlty-reports/data_*.json 2>/dev/null | sed -n '2p'
```

If a previous report exists, perform the comparison and add a section to the report:

```markdown
---

## Comparison with Previous Report

**Previous report**: [PREVIOUS_REPORT_DATE]

| Metric | Previous | Current | Evolution |
|--------|----------|---------|-----------|
| Lint issues | [N] | [N] | [+/-N] [↑/↓/=] |
| Code smells | [N] | [N] | [+/-N] [↑/↓/=] |
| Max complexity | [N] | [N] | [+/-N] [↑/↓/=] |

### Overall trend
[Improvement / Stable / Degradation] + explanation
```

Evolution legend:
- `↑ +N` red = degradation (more issues)
- `↓ -N` green = improvement (fewer issues)
- `= 0` grey = stable

---

### Phase 6: Displaying Results

Display a compact summary in the terminal:

```
=== QUALITY REPORT: [PROJECT] ===

Date    : [DATE_TIME]
Path    : [DIRECTORY]

📊 SUMMARY
─────────────────────────────
Lint issues    : [N]  [TREND vs previous if available]
Code smells    : [N]  [TREND]
Max complexity : [N]  [TREND]

[VERDICT: ✅ GOOD / ⚠️ WARNING / ❌ CRITICAL]

📁 SAVED REPORTS
─────────────────────────────
Markdown : .qlty-reports/report_[TIMESTAMP].md
JSON     : .qlty-reports/data_[TIMESTAMP].json

💡 TOP 3 RECOMMENDATIONS
─────────────────────────────
1. [Most urgent recommendation]
2. [Second recommendation]
3. [Third recommendation]
```

---

### Strict Rules

**You MUST NEVER:**
- ❌ Modify source code to fix found issues (unless explicitly requested)
- ❌ Run `qlty fmt` or `qlty fix` without user confirmation
- ❌ Delete or overwrite existing reports
- ❌ Continue if `qlty --version` fails

**You MUST ALWAYS:**
- ✅ Timestamp each report for traceability
- ✅ Explain each issue/smell in plain language, not just list them
- ✅ Propose concrete corrective actions
- ✅ Preserve the report history (never delete)
- ✅ Display the trend if a previous report exists
- ✅ Add `.qlty-reports/` to `.gitignore` if requested (do not do it automatically)

---

### Error Handling

**qlty check fails on certain files:**
- Continue the analysis on other files
- List the failing files in the report under "Non-analyzable files"

**No issues found:**
- Good news! Report with ✅ GOOD verdict
- Mention that the score is excellent but vigilance remains important

**Empty or non-code directory:**
- Report that no analyzable files were found
- List file types supported by qlty

---

### Notes

- Reports are kept indefinitely in `.qlty-reports/` — this is intentional for history tracking
- The raw JSON data format allows integrating these metrics into external tools
- qlty supports many languages (Go, JS/TS, Python, Ruby, etc.) via its plugins
- To add new linters: `qlty plugins list` then `qlty plugins enable [plugin]`
