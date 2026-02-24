# Sample 122 - Tasks Agent for AI Coding Assistant

This sample demonstrates how to use the N.O.V.A. Tasks Agent to decompose a natural language request into an **ordered, classified execution plan** for an AI coding assistant.

## The Problem

When a user writes a request like:

> Read the file in ./specification.md.
> Use these document to generate golang code.
> Check carefully the syntax of the code.
> Make the code simple and add remarks to explain the code.
> Once the code ready,
> Then save the source code into a ./demo/main.go file in the same folder,
> Then create a new markdown document that explains how the code works,
> and save it into ./demo/explanation.md file.
> create the demo folder if it does not exist.

An orchestrator needs to know, for each step:

1. **What** to do (the description)
2. **How** to do it — is it a tool call, a text generation, or a code generation?
3. **In what order** — respecting logical dependencies, not the order the user wrote them
4. **What depends on what** — to enable parallel execution when possible

## The Task Struct

The `Task` struct (`nova-sdk/agents/agents.plan.go`) provides all the fields needed:

```go
type Task struct {
    ID          string            `json:"id"`
    Description string            `json:"description"`
    Responsible string            `json:"responsible"`
    ToolName    string            `json:"tool_name,omitempty"`
    Arguments   map[string]string `json:"arguments,omitempty"`
    DependsOn   []string          `json:"depends_on,omitempty"`
    Complexity  string            `json:"complexity,omitempty"`
}
```

| Field | Description |
|---|---|
| `ID` | Sequential execution order (`"1"`, `"2"`, `"3"`, ...) |
| `Description` | Clear, actionable description of the task |
| `Responsible` | Who executes the task: `"tool"`, `"completion"`, or `"developer"` |
| `ToolName` | When `responsible == "tool"`: the exact tool name to call |
| `Arguments` | When `responsible == "tool"`: the arguments map (e.g. `{"path": "./demo"}`) |
| `DependsOn` | List of task IDs that must complete before this task can start |
| `Complexity` | `"simple"`, `"moderate"`, or `"complex"` |

## The Three Responsible Types

### `"tool"` — External Tool Call

The task is a concrete operation executed by a tool (file I/O, API call, etc.). The system instructions define the available tools:

| Tool | Description | Arguments |
|---|---|---|
| `read_file` | Read the content of a file | `path` |
| `save_file` | Save/write content to a file | `path` |
| `create_directory` | Create a directory/folder | `path` |

When `responsible` is `"tool"`, the `tool_name` and `arguments` fields are required.

### `"completion"` — Generalist LLM

The task requires a **generalist LLM** to produce text: documentation, explanations, analysis, summaries. In a future orchestrator, these tasks would be routed to a general-purpose model.

### `"developer"` — Code-Specialized LLM

The task requires a **code-specialized LLM** to generate, review, or refactor source code. In a future orchestrator, these tasks would be routed to a code-focused model (e.g. DeepSeek Coder, Codestral) which is typically better at code generation and syntax checking than a generalist model.

## Complexity Levels

Each task is rated by complexity, which can be used by the orchestrator to select the appropriate model size:

| Level | Description | Examples |
|---|---|---|
| `"simple"` | Trivial operation, no reasoning needed | Create a directory, read a file |
| `"moderate"` | Some logic or moderate generation | Write documentation, simple transformations |
| `"complex"` | Deep reasoning, code generation, or analysis | Generate code from a specification, architecture decisions |

## Dependency-Based Ordering

The key challenge is **reordering tasks based on logical dependencies**, not based on the order the user wrote them.

### The Problem with Naive Ordering

In the user request above, "create the demo folder if it does not exist" appears **last**. But saving files to `./demo/main.go` requires the `./demo/` directory to exist. A naive ordering would fail at execution time.

### How the System Prompt Solves This

The system instructions include explicit dependency rules:

- **Directory before file**: If a task saves a file into a directory, `create_directory` must come first
- **Read before generate**: If a task generates content from a file, `read_file` must come first
- **Generate before save**: If a task generates content to be saved, generation must come before `save_file`

A **few-shot example** is included in the system prompt to help small local models (3B parameters) apply these rules correctly. Abstract rules alone are often insufficient — a concrete example showing "mentioned last, scheduled first" makes the pattern reproducible by analogy.

### Expected Output

For the sample user request, the correctly ordered plan should be:

```json
{
  "tasks": [
    {
      "id": "1",
      "description": "Read the content of ./specification.md",
      "responsible": "tool",
      "tool_name": "read_file",
      "arguments": {"path": "./specification.md"},
      "depends_on": [],
      "complexity": "simple"
    },
    {
      "id": "2",
      "description": "Generate golang code based on the specification, with simple syntax and explanatory comments",
      "responsible": "developer",
      "depends_on": ["1"],
      "complexity": "complex"
    },
    {
      "id": "3",
      "description": "Check the syntax of the generated golang code",
      "responsible": "developer",
      "depends_on": ["2"],
      "complexity": "moderate"
    },
    {
      "id": "4",
      "description": "Create the ./demo directory if it does not exist",
      "responsible": "tool",
      "tool_name": "create_directory",
      "arguments": {"path": "./demo"},
      "depends_on": [],
      "complexity": "simple"
    },
    {
      "id": "5",
      "description": "Save the source code to ./demo/main.go",
      "responsible": "tool",
      "tool_name": "save_file",
      "arguments": {"path": "./demo/main.go"},
      "depends_on": ["3", "4"],
      "complexity": "simple"
    },
    {
      "id": "6",
      "description": "Generate a markdown document explaining how the code works",
      "responsible": "completion",
      "depends_on": ["3"],
      "complexity": "moderate"
    },
    {
      "id": "7",
      "description": "Save the explanation to ./demo/explanation.md",
      "responsible": "tool",
      "tool_name": "save_file",
      "arguments": {"path": "./demo/explanation.md"},
      "depends_on": ["4", "6"],
      "complexity": "simple"
    }
  ]
}
```

Notice:
- Task 4 (`create_directory`) has **no dependencies** — it can run in parallel with tasks 1-3
- Task 5 (`save_file`) depends on **both** task 3 (code is ready) **and** task 4 (directory exists)
- Task 6 (`completion`) generates documentation — routed to a generalist LLM
- Task 2 (`developer`) generates code — routed to a code-specialized LLM

## Dependency Graph

```
1 (read_file)──────► 2 (developer)──► 3 (developer)──► 5 (save_file)
                                            │                 ▲
                                            │                 │
                                            ▼                 │
                                       6 (completion)──► 7 (save_file)
                                                              ▲
4 (create_directory)──────────────────────────────────────────┘
```

Tasks 1 and 4 have no mutual dependency — they could execute in parallel.

## Model Choice

This sample uses `jan-nano-128k-gguf:Q4_K_M` with `temperature: 0.0`. The comment in the code notes that:

- **Lucy** (`lucy-128k-gguf`) is faster
- **Jan Nano** (`jan-nano-128k-gguf`) is more accurate for complex reasoning tasks like ordering and dependency detection

## Running the Sample

```bash
go run samples/122-tasks-agent/main.go
```

Requires a local LLM engine running at `http://localhost:12434/engines/llama.cpp/v1`.

## Next Steps

The TODO in the code outlines the future direction:

1. **Orchestrator execution** — iterate over tasks and execute each one based on its `responsible` type
2. **Context propagation** — the result of each task feeds into the next (e.g. file content read by `read_file` becomes context for the `developer` code generation task)
3. **Model routing** — use `responsible` + `complexity` to select the right model for each task
