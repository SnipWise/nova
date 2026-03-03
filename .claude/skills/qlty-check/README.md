# Skill: qlty-check

Analyzes code quality of a project using the [qlty](https://docs.qlty.sh/cli/quickstart) CLI, generates a detailed report with explanations, and saves data to track evolution over time.

## Prerequisites

- **qlty CLI** installed: `curl https://qlty.sh | sh`
- **qlty initialized** in the target directory: `qlty init` (the skill offers to do it if missing)

## Usage

```
/qlty-check
```

The skill asks two questions at startup:
1. The directory to analyze (default: current directory)
2. The desired analysis type

## Analysis Types

| Option | Commands executed | Description |
|--------|-------------------|-------------|
| `all` *(recommended)* | `qlty check --all` + `qlty metrics --all` + `qlty smells --all` | Full analysis |
| `lint` | `qlty check --all` | Lint issues only |
| `metrics` | `qlty metrics --all --max-depth=2 --sort complexity --limit 20` | Complexity and metrics |
| `smells` | `qlty smells --all` | Duplication and code smells |

## Generated Outputs

Reports are saved in `.qlty-reports/` at the root of the analyzed directory:

```
[project]/
└── .qlty-reports/
    ├── report_20240315_143022.md   ← readable report + recommendations
    ├── data_20240315_143022.json   ← raw data for comparison
    ├── report_20240320_091500.md
    └── data_20240320_091500.json
```

### Markdown Report (`report_[TIMESTAMP].md`)

Contains:
- **Executive summary** with overall verdict (✅ GOOD / ⚠️ WARNING / ❌ CRITICAL)
- **Lint issue details** with explanation of each category and fix suggestions
- **Complexity metrics** with cyclomatic complexity explanation and top problematic files
- **Code smells** with smell type, impact, and refactoring strategy
- **Prioritized recommendations** (🔴 high / 🟡 medium / 🟢 low)
- **Comparison with previous report** (if available)

### JSON Data (`data_[TIMESTAMP].json`)

Structured data for automated tracking:

```json
{
  "timestamp": "20240315_143022",
  "project": "my-project",
  "directory": "/path/to/project",
  "qlty_version": "1.x.x",
  "lint": {
    "total_issues": 12,
    "by_severity": { "error": 2, "warning": 8, "info": 2 }
  },
  "metrics": {
    "max_complexity": 18,
    "avg_complexity": 4.2,
    "files_analyzed": 34
  },
  "smells": {
    "total": 5,
    "by_type": {}
  }
}
```

## Overall Verdict Thresholds

| Verdict | Lint issues | Code smells |
|---------|-------------|-------------|
| ✅ GOOD | 0–5 | 0–3 |
| ⚠️ WARNING | 6–20 | 4–10 |
| ❌ CRITICAL | > 20 | > 10 |

## Time-based Comparison

When a previous report exists, the skill automatically displays evolution:

```
| Metric         | Previous | Current | Evolution |
|----------------|----------|---------|-----------|
| Lint issues    | 15       | 8       | ↓ -7  ✅  |
| Code smells    | 3        | 5       | ↑ +2  ⚠️  |
| Max complexity | 22       | 18      | ↓ -4  ✅  |
```

## What the Skill Does NOT Do

- It does not auto-fix code (`qlty fmt` / `qlty fix` are not run without confirmation)
- It does not delete existing reports (history is preserved)
- It does not modify `.gitignore` automatically (offered on request)

## Tip: Exclude Reports from Version Control

```bash
echo ".qlty-reports/" >> .gitignore
```

Or ask the skill to do it during the analysis.
