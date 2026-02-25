# 129 — Crew Agent with Single Agent & Tasks

This sample demonstrates a **simplified Crew Agent** that uses a single chat agent (no orchestrator, no routing) combined with a tasks planner, tools agent, compressor, and a skills system. Everything is configurable via `config.yaml`.

## Architecture

```
User input
    │
    ├──► /skill <name> <input>  ──► Prompt enrichment ──┐
    │                                                    │
    ▼                                                    ▼
┌─────────────────┐    plan   ┌─────────────────┐    ┌──────────────────┐
│ Tasks Planner   │──────────►│  Tools Agent    │    │   Coder Agent    │
│ (task breakdown)│           │ (file ops, bash)│    │ (code generation)│
└─────────────────┘           └─────────────────┘    └──────────────────┘
                                     │                      │
                                     ▼                      ▼
                              ┌────────────────┐        Streaming
                              │   Compressor   │        response
                              │ (context mgmt) │
                              └────────────────┘
```

### Agents

| Agent | Role | Config key |
|-------|------|------------|
| **Tasks Planner** | Breaks user requests into ordered tasks with dependencies | `tasks_planner` |
| **Coder** | Single chat agent for code generation and conversation | `coder` |
| **Tools** | Executes file system operations (read, write, find, grep, bash...) | `tools` |
| **Compressor** | Summarises conversation history when context exceeds the limit | `compressor` |

### Key difference with sample 128

Sample 128 uses `crew.WithAgentCrew()` with an orchestrator that routes topics to different chat agents (coder, generic). This sample uses `crew.WithSingleAgent()` — **no orchestrator, no routing, one agent handles everything**.

```go
// Sample 128 (multi-agent with orchestrator)
crewAgent, _ := crew.NewAgent(ctx,
    crew.WithAgentCrew(chatAgents, defaultAgent),
    crew.WithOrchestratorAgent(orchestratorAgent),
)

// Sample 129 (single agent, no orchestrator)
crewAgent, _ := crew.NewAgent(ctx,
    crew.WithSingleAgent(chatAgent),
)
```

## Files

| File | Description |
|------|-------------|
| `main.go` | Entry point: REPL loop, banner, commands, crew assembly |
| `config.go` | YAML config structs and loading |
| `config.yaml` | All configuration: agents, tools, skills |
| `agent.chat.go` | Creates the single coder chat agent |
| `agent.tasks.go` | Creates the tasks planner agent with spinner |
| `agent.tools.go` | Creates the tools agent and builds the tool index |
| `agent.compressor.go` | Creates the compressor agent with spinner |
| `tools.execution.go` | Tool execution (shell commands) and confirmation prompt |

## Configuration

All configuration lives in `config.yaml`. No Go code changes needed to:

- Change models or temperatures
- Edit agent instructions
- Add/remove tools
- Add/remove skills

### Agents

```yaml
agents:
  tasks_planner:
    model: "huggingface.co/menlo/jan-nano-128k-gguf:Q4_K_M"
    temperature: 0.0
    instructions: |
      You are an expert task planner...

  coder:
    model: "huggingface.co/unsloth/rnj-1-instruct-gguf:Q4_K_M"
    temperature: 0.8
    instructions: |
      You are an expert programming assistant...
```

### Tools

Tools are shell command templates with `{{param}}` placeholders:

```yaml
tools:
  - name: read_file
    description: "Read the content of a text file"
    command: "cat '{{path}}'"
    parameters:
      - name: path
        type: string
        description: "The file path to read"
        required: true
```

Built-in tools: `read_file`, `write_file`, `list_directory`, `find_files`, `grep_files`, `create_directory`, `bash`.

### Skills

Skills are reusable prompt templates invocable via `/skill <name> <input>`:

```yaml
skills:
  - name: explain
    description: "Explain a source file in plain language"
    prompt: |
      Read the file {{input}} and explain what it does in plain language.
      Be concise, highlight the main purpose, key functions, and any important details.
```

Built-in skills: `explain`, `summarize`, `review`, `test`, `doc`.

## REPL Commands

| Command | Description |
|---------|-------------|
| `/new` | Reset all agents memory and start fresh |
| `/pack` | Force context packing (compress history) |
| `/skill <name> <input>` | Run a skill |
| `/bye` | Exit the program |
| `/help` | Show available commands and skills |

## Spinners

Each agent displays a spinner while working:

```
  ⣾ Analyzing your plan...    (tasks planner)
  ⣾ Generating response...    (coder)
  ⣾ Executing the tool...     (tools)
  ⣾ Compressing context...    (compressor)
```

## Usage examples

### Direct conversation

```
🧑 You: I need a hello world program in golang, create a ./demos directory, and save this program to hello.go into ./demos
```

The tasks planner will break this into 3 tasks:
1. `create_directory` (tool) — create `./demos`
2. Generate code (developer) — produce the hello world program
3. `write_file` (tool) — save the code to `./demos/hello.go`

### Using skills

```
🧑 You: /skill explain ./main.go
🧑 You: /skill review ./config.go
🧑 You: /skill test ./utils.go
```

### Context management

```
🧑 You: /pack
🗜️  Context packed! New size: 4521 characters

🧑 You: /new
🧹 All agents memory has been reset. Starting fresh!
```

## Running

```bash
go run .
```

Make sure the LLM engine is running at the URL specified in `config.yaml` (`engine_url`).
