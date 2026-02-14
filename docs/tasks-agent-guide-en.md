# Tasks Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Understanding the Plan Output Type](#3-understanding-the-plan-output-type)
4. [Agent Configuration](#4-agent-configuration)
5. [Model Configuration](#5-model-configuration)
6. [Identifying Plans](#6-identifying-plans)
7. [Conversation History and Messages](#7-conversation-history-and-messages)
8. [Lifecycle Hooks (TasksAgentOption)](#8-lifecycle-hooks-tasksagentoption)
9. [Context and State Management](#9-context-and-state-management)
10. [JSON Export and Debugging](#10-json-export-and-debugging)
11. [API Reference](#11-api-reference)

---

## 1. Introduction

### What is a Tasks Agent?

The `tasks.Agent` is a specialized structured agent provided by the Nova SDK (`github.com/snipwise/nova`) that identifies and extracts structured task plans from natural language input. It uses the `agents.Plan` type as its output, converting user descriptions into organized task lists with clear responsibilities.

Unlike a general chat agent, the tasks agent focuses on understanding project requirements and generating actionable, structured plans. It's built on top of the structured agent framework, ensuring type-safe, predictable output.

### When to use a Tasks Agent

| Scenario | Recommended agent |
|---|---|
| Break down project descriptions into task lists | `tasks.Agent` |
| Extract structured plans from meeting notes or requirements | `tasks.Agent` |
| Convert natural language goals into actionable tasks | `tasks.Agent` |
| Generate organized project task lists | `tasks.Agent` |
| General text extraction or classification | `structured.Agent[YourType]` |
| Free-form conversational AI | `chat.Agent` |
| Function calling / tool use | `tools.Agent` |

### Key capabilities

- **Plan identification**: Automatically converts natural language descriptions into structured plans with tasks and responsibilities.
- **Sequential organization**: Generates ordered task lists for clear project execution.
- **Type-safe output**: Always returns a properly structured `agents.Plan` object.
- **Conversation history**: Optionally maintain conversation history for iterative plan refinement.
- **Lifecycle hooks**: Execute custom logic before and after plan identification.
- **JSON export**: Export conversation history and plans for debugging or persistence.

---

## 2. Quick Start

### Minimal example

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/tasks"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := tasks.NewAgent(
        ctx,
        agents.Config{
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a project planning assistant that breaks down project descriptions into structured task plans.",
        },
        models.Config{
            Name:        "ai/qwen2.5:7B-Q8_0",
            Temperature: models.Float64(0.3),
        },
    )
    if err != nil {
        panic(err)
    }

    plan, err := agent.IdentifyPlanFromText(
        "Create a web application for task management with user authentication and a dashboard.",
    )
    if err != nil {
        panic(err)
    }

    fmt.Println("Plan identified successfully!")
    fmt.Printf("Number of main tasks: %d\n", len(plan.Tasks))
}
```

---

## 3. Understanding the Plan Output Type

The tasks agent returns an `agents.Plan` struct, which contains a list of tasks:

### Plan structure

```go
type Plan struct {
    Tasks []Task `json:"tasks"`
}

type Task struct {
    ID          string `json:"id"`
    Description string `json:"description"`
    Responsible string `json:"responsible"`
}
```

### Example plan output

```json
{
  "tasks": [
    {
      "id": "1",
      "description": "Setup project infrastructure and initialize Git repository",
      "responsible": "DevOps Team"
    },
    {
      "id": "2",
      "description": "Configure CI/CD pipeline with automated testing",
      "responsible": "DevOps Engineer"
    },
    {
      "id": "3",
      "description": "Implement user authentication and authorization system",
      "responsible": "Backend Team"
    },
    {
      "id": "4",
      "description": "Design and develop the frontend dashboard",
      "responsible": "Frontend Team"
    }
  ]
}
```

### Task properties

- **ID**: Unique identifier for the task (sequential: "1", "2", "3", etc.)
- **Description**: Clear, actionable description of what needs to be done
- **Responsible**: Who is responsible for completing the task (team, role, or person)

---

## 4. Agent Configuration

```go
agents.Config{
    Name:                    "tasks-agent",        // Agent name (optional)
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1", // LLM engine URL (required)
    APIKey:                  "your-api-key",        // API key (optional)
    SystemInstructions:      "You are a project planning assistant...", // System prompt (recommended)
    KeepConversationHistory: false,                 // Usually false for stateless extraction
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `Name` | `string` | No | Agent identifier for logging. |
| `EngineURL` | `string` | Yes | URL of the OpenAI-compatible LLM engine. |
| `APIKey` | `string` | No | API key for authenticated engines. |
| `SystemInstructions` | `string` | Recommended | System prompt defining the planning task. A good default is provided if omitted. |
| `KeepConversationHistory` | `bool` | No | Usually `false` for stateless plan extraction. Default: `false`. |

### Recommended system instructions

```go
SystemInstructions: `You are an expert project planner. Break down project descriptions into clear, actionable tasks with:
- Unique sequential IDs (1, 2, 3, etc.)
- Specific, actionable descriptions
- Appropriate responsibility assignments (team/role/person)
- Logical task organization and ordering`
```

---

## 5. Model Configuration

```go
models.Config{
    Name:        "ai/qwen2.5:7B-Q8_0",    // Model ID (required)
    Temperature: models.Float64(0.3),      // 0.3 for structured but slightly creative planning
    MaxTokens:   models.Int(4000),          // Allow for larger plans
}
```

### Recommended settings

- **Temperature**: `0.3` - structured enough for consistent formatting, but allows for creative task breakdown
- **Model**: Use models with strong instruction-following and JSON capabilities (Qwen 2.5 7B+ recommended)
- **MaxTokens**: 3000-4000+ for complex projects with many tasks

---

## 6. Identifying Plans

### IdentifyPlanFromText

The simplest method for extracting a plan from text:

```go
plan, err := agent.IdentifyPlanFromText(
    "Build a REST API with authentication, user management, and real-time notifications",
)
if err != nil {
    // handle error
}

// Access the tasks
for _, task := range plan.Tasks {
    fmt.Printf("[%s] %s\n", task.ID, task.Description)
    fmt.Printf("     Responsible: %s\n\n", task.Responsible)
}
```

### IdentifyPlan

For more control, use the full method with messages:

```go
userMessages := []messages.Message{
    {
        Role:    roles.User,
        Content: "Create a mobile app for expense tracking",
    },
}

plan, finishReason, err := agent.IdentifyPlan(userMessages)
if err != nil {
    // handle error
}

fmt.Println("Finish reason:", finishReason) // "stop"
```

**Return values:**
- `plan *agents.Plan`: The extracted task plan
- `finishReason string`: Why generation stopped (`"stop"`, `"length"`, etc.)
- `err error`: Error if generation or parsing failed

### Multi-turn plan refinement

```go
// Initial planning
plan, _, err := agent.IdentifyPlan([]messages.Message{
    {Role: roles.User, Content: "Build an e-commerce platform"},
})

// Add feedback and refine
agent.AddMessage(roles.User, "Add tasks for payment integration and inventory management")

// Generate refined plan
refinedPlan, _, err := agent.IdentifyPlan([]messages.Message{
    {Role: roles.User, Content: "Refine the plan with the new requirements"},
})
```

---

## 7. Conversation History and Messages

### Managing messages

```go
// Get all messages in history
msgs := agent.GetMessages()

// Add a single message
agent.AddMessage(roles.User, "Add security audit tasks")

// Add multiple messages at once
agent.AddMessages([]messages.Message{
    {Role: roles.User, Content: "Consider scalability"},
    {Role: roles.Assistant, Content: "..."},
})

// Clear all messages except the system instruction
agent.ResetMessages()
```

---

## 8. Lifecycle Hooks (TasksAgentOption)

Lifecycle hooks allow you to execute custom logic before and after plan identification.

### TasksAgentOption

```go
type TasksAgentOption func(*Agent)
```

Options are passed as variadic arguments to `NewAgent`:

```go
agent, err := tasks.NewAgent(ctx, agentConfig, modelConfig,
    tasks.BeforeCompletion(fn),
    tasks.AfterCompletion(fn),
)
```

### BeforeCompletion

Called before each plan identification:

```go
tasks.BeforeCompletion(func(a *tasks.Agent) {
    fmt.Println("About to identify plan...")
    fmt.Printf("Messages count: %d\n", len(a.GetMessages()))
})
```

**Use cases:**
- Logging and monitoring
- Metrics collection
- Pre-identification state inspection

### AfterCompletion

Called after each plan identification:

```go
tasks.AfterCompletion(func(a *tasks.Agent) {
    fmt.Println("Plan identification completed.")
    fmt.Printf("Messages count: %d\n", len(a.GetMessages()))
})
```

**Use cases:**
- Logging identification results
- Post-identification metrics
- Triggering downstream actions (e.g., storing plans in database)
- Auditing/tracking

### Complete example with hooks

```go
agent, err := tasks.NewAgent(
    ctx,
    agents.Config{
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "You are a project planning assistant.",
    },
    models.Config{
        Name:        "ai/qwen2.5:7B-Q8_0",
        Temperature: models.Float64(0.3),
    },
    tasks.BeforeCompletion(func(a *tasks.Agent) {
        fmt.Printf("[BEFORE] Agent: %s, Messages: %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
    tasks.AfterCompletion(func(a *tasks.Agent) {
        fmt.Printf("[AFTER] Agent: %s, Messages: %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
)
```

---

## 9. Context and State Management

### Getting and setting context

```go
ctx := agent.GetContext()
agent.SetContext(newCtx)
```

### Getting and setting configuration

```go
// Agent configuration
config := agent.GetConfig()
agent.SetConfig(newConfig)

// Model configuration
modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)
```

### Agent metadata

```go
agent.Kind()       // Returns agents.Tasks
agent.GetName()    // Returns the agent name from config
agent.GetModelID() // Returns the model name from model config
```

---

## 10. JSON Export and Debugging

### Raw request/response JSON

```go
// Raw (unformatted) JSON
rawReq := agent.GetLastRequestRawJSON()
rawResp := agent.GetLastResponseRawJSON()

// Pretty-printed JSON
prettyReq, err := agent.GetLastRequestJSON()
prettyResp, err := agent.GetLastResponseJSON()
```

---

## 11. API Reference

### Constructor

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    opts ...TasksAgentOption,
) (*Agent, error)
```

Creates a new tasks agent. The `opts` parameter accepts zero or more `TasksAgentOption` functional options.

---

### Types

```go
// TasksAgentOption is a functional option for configuring an Agent during creation
type TasksAgentOption func(*Agent)
```

---

### Option Functions

| Function | Description |
|---|---|
| `BeforeCompletion(fn func(*Agent))` | Sets a hook called before each plan identification. |
| `AfterCompletion(fn func(*Agent))` | Sets a hook called after each plan identification. |

---

### Methods

| Method | Description |
|---|---|
| `IdentifyPlanFromText(text string) (*agents.Plan, error)` | Identify a plan from a simple text description. |
| `IdentifyPlan(userMessages []messages.Message) (*agents.Plan, string, error)` | Identify a plan from messages. Returns the plan, finish reason, and error. |
| `GetMessages() []messages.Message` | Get all conversation messages. |
| `AddMessage(role roles.Role, content string)` | Add a single message to history. |
| `AddMessages(msgs []messages.Message)` | Add multiple messages to history. |
| `ResetMessages()` | Clear all messages except system instruction. |
| `GetConfig() agents.Config` | Get the agent configuration. |
| `SetConfig(config agents.Config)` | Update the agent configuration. |
| `GetModelConfig() models.Config` | Get the model configuration. |
| `SetModelConfig(config models.Config)` | Update the model configuration. |
| `GetContext() context.Context` | Get the agent's context. |
| `SetContext(ctx context.Context)` | Update the agent's context. |
| `GetLastRequestRawJSON() string` | Get the raw JSON of the last request. |
| `GetLastResponseRawJSON() string` | Get the raw JSON of the last response. |
| `GetLastRequestJSON() (string, error)` | Get the pretty-printed JSON of the last request. |
| `GetLastResponseJSON() (string, error)` | Get the pretty-printed JSON of the last response. |
| `Kind() agents.Kind` | Returns `agents.Tasks`. |
| `GetName() string` | Returns the agent name. |
| `GetModelID() string` | Returns the model name. |
