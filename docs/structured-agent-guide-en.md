# Structured Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Defining Output Types](#3-defining-output-types)
4. [Agent Configuration](#4-agent-configuration)
5. [Model Configuration](#5-model-configuration)
6. [Generating Structured Data](#6-generating-structured-data)
7. [Conversation History and Messages](#7-conversation-history-and-messages)
8. [Lifecycle Hooks (StructuredAgentOption)](#8-lifecycle-hooks-structuredagentoption)
9. [Context and State Management](#9-context-and-state-management)
10. [JSON Export and Debugging](#10-json-export-and-debugging)
11. [API Reference](#11-api-reference)

---

## 1. Introduction

### What is a Structured Agent?

The `structured.Agent[Output]` is a generic agent provided by the Nova SDK (`github.com/snipwise/nova`) that generates structured JSON output conforming to a Go struct type. It uses JSON Schema generation from your Go struct to instruct the LLM to return data in a precise, typed format.

Unlike a chat agent that returns free-form text, the structured agent always returns a parsed Go struct. This makes it ideal for data extraction, classification, entity recognition, and any task where you need structured, typed output from an LLM.

### When to use a Structured Agent

| Scenario | Recommended agent |
|---|---|
| Extract structured data from text (entities, facts, etc.) | `structured.Agent[YourType]` |
| Classify text into predefined categories with metadata | `structured.Agent[Classification]` |
| Parse natural language into typed Go structs | `structured.Agent[YourType]` |
| Topic/intent detection for routing | `orchestrator.Agent` (wraps structured agent) |
| Free-form conversational AI | `chat.Agent` |
| Function calling / tool use | `tools.Agent` |

### Key capabilities

- **Generic typed output**: Define any Go struct as the output type; the agent ensures LLM responses conform to it.
- **JSON Schema enforcement**: Automatically generates a JSON Schema from your Go struct and uses strict mode for guaranteed valid output.
- **Conversation history**: Optionally maintain conversation history across multiple calls.
- **Lifecycle hooks**: Execute custom logic before and after each data generation.
- **JSON export**: Export conversation history for debugging or persistence.

---

## 2. Quick Start

### Minimal example

```go
package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/structured"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

type Country struct {
    Name       string   `json:"name"`
    Capital    string   `json:"capital"`
    Population int      `json:"population"`
    Languages  []string `json:"languages"`
}

func main() {
    ctx := context.Background()

    agent, err := structured.NewAgent[Country](
        ctx,
        agents.Config{
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are an assistant that answers questions about countries.",
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.0),
        },
    )
    if err != nil {
        panic(err)
    }

    response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
        {Role: roles.User, Content: "Tell me about Canada."},
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("Name:", response.Name)
    fmt.Println("Capital:", response.Capital)
    fmt.Println("Population:", response.Population)
    fmt.Println("Languages:", strings.Join(response.Languages, ", "))
    fmt.Println("Finish reason:", finishReason)
}
```

---

## 3. Defining Output Types

The structured agent uses Go generics to enforce the output type. Any Go struct with JSON tags can be used.

### Simple struct

```go
type Country struct {
    Name       string   `json:"name"`
    Capital    string   `json:"capital"`
    Population int      `json:"population"`
    Languages  []string `json:"languages"`
}

agent, _ := structured.NewAgent[Country](ctx, agentConfig, modelConfig)
```

### Nested structs

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type Person struct {
    Name    string  `json:"name"`
    Age     int     `json:"age"`
    Address Address `json:"address"`
}

agent, _ := structured.NewAgent[Person](ctx, agentConfig, modelConfig)
```

### Slice output

You can use a slice as the output type for generating lists:

```go
type Item struct {
    Name  string `json:"name"`
    Price float64 `json:"price"`
}

agent, _ := structured.NewAgent[[]Item](ctx, agentConfig, modelConfig)
```

### JSON Schema generation

The agent automatically generates a JSON Schema from your Go struct using reflection. The schema is passed to the LLM with `strict: true` to ensure the response always matches the expected format. Supported Go types map to JSON Schema types:

| Go type | JSON Schema type |
|---|---|
| `string` | `string` |
| `int`, `int64` | `integer` |
| `float64` | `number` |
| `bool` | `boolean` |
| `[]T` | `array` of T |
| `struct` | `object` with properties |

---

## 4. Agent Configuration

```go
agents.Config{
    Name:                    "structured-agent",   // Agent name (optional)
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1", // LLM engine URL (required)
    APIKey:                  "your-api-key",        // API key (optional)
    SystemInstructions:      "You are an assistant that extracts country information.", // System prompt (recommended)
    KeepConversationHistory: false,                 // Usually false for extraction tasks
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `Name` | `string` | No | Agent identifier for logging. |
| `EngineURL` | `string` | Yes | URL of the OpenAI-compatible LLM engine. |
| `APIKey` | `string` | No | API key for authenticated engines. |
| `SystemInstructions` | `string` | Recommended | System prompt defining the extraction/generation task. |
| `KeepConversationHistory` | `bool` | No | Usually `false` for stateless extraction. Default: `false`. |

---

## 5. Model Configuration

```go
models.Config{
    Name:        "ai/qwen2.5:1.5B-F16",    // Model ID (required)
    Temperature: models.Float64(0.0),        // 0.0 for deterministic extraction
    MaxTokens:   models.Int(2000),            // Max response length
}
```

### Recommended settings

- **Temperature**: `0.0` for deterministic, factual extraction. Higher values for creative generation.
- **Model**: Models with good JSON/instruction-following capabilities work best (Qwen, Llama, etc.).

---

## 6. Generating Structured Data

### GenerateStructuredData

The primary method for generating typed output:

```go
response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Tell me about France."},
})
if err != nil {
    // handle error
}

// response is *Country (typed)
fmt.Println(response.Name)       // "France"
fmt.Println(response.Capital)    // "Paris"
fmt.Println(response.Population) // 67390000
fmt.Println(response.Languages)  // ["French"]
fmt.Println(finishReason)        // "stop"
```

**Return values:**
- `response *Output`: A pointer to the parsed output struct.
- `finishReason string`: Why generation stopped (`"stop"`, `"length"`, etc.).
- `err error`: Error if generation or parsing failed.

### Sending multiple messages

You can provide conversation context:

```go
response, _, err := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "I'm interested in European countries."},
    {Role: roles.Assistant, Content: `{"name":"","capital":"","population":0,"languages":[]}`},
    {Role: roles.User, Content: "Tell me about Germany."},
})
```

---

## 7. Conversation History and Messages

### Managing messages

```go
// Get all messages in history
msgs := agent.GetMessages()

// Add a single message
agent.AddMessage(roles.User, "A manual message")

// Add multiple messages at once
agent.AddMessages([]messages.Message{
    {Role: roles.User, Content: "First message"},
    {Role: roles.Assistant, Content: "First response"},
})

// Clear all messages except the system instruction
agent.ResetMessages()
```

### Export conversation to JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
if err != nil {
    // handle error
}
fmt.Println(jsonStr)
```

---

## 8. Lifecycle Hooks (StructuredAgentOption)

Lifecycle hooks allow you to execute custom logic before and after each structured data generation. They are configured as functional options when creating the agent.

### StructuredAgentOption

The option type is generic, matching the agent's output type:

```go
type StructuredAgentOption[Output any] func(*Agent[Output])
```

Options are passed as variadic arguments to `NewAgent`:

```go
agent, err := structured.NewAgent[Country](ctx, agentConfig, modelConfig,
    structured.BeforeCompletion[Country](fn),
    structured.AfterCompletion[Country](fn),
)
```

**Note:** Go can often infer the type parameter, so you may omit it:

```go
agent, err := structured.NewAgent[Country](ctx, agentConfig, modelConfig,
    structured.BeforeCompletion(fn),
    structured.AfterCompletion(fn),
)
```

### BeforeCompletion

Called before each structured data generation. The hook receives a reference to the typed agent.

```go
structured.BeforeCompletion[Country](func(a *structured.Agent[Country]) {
    fmt.Println("About to generate structured data...")
    fmt.Printf("Messages count: %d\n", len(a.GetMessages()))
})
```

**Use cases:**
- Logging and monitoring
- Metrics collection
- Pre-generation state inspection

### AfterCompletion

Called after each structured data generation, once the result has been parsed. The hook receives a reference to the typed agent.

```go
structured.AfterCompletion[Country](func(a *structured.Agent[Country]) {
    fmt.Println("Structured data generation completed.")
    fmt.Printf("Messages count: %d\n", len(a.GetMessages()))
})
```

**Use cases:**
- Logging generation results
- Post-generation metrics
- Triggering downstream actions
- Auditing/tracking

### Complete example with hooks

```go
agent, err := structured.NewAgent[Country](
    ctx,
    agents.Config{
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "You are an assistant that answers questions about countries.",
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
    },
    structured.BeforeCompletion[Country](func(a *structured.Agent[Country]) {
        fmt.Printf("[BEFORE] Agent: %s, Messages: %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
    structured.AfterCompletion[Country](func(a *structured.Agent[Country]) {
        fmt.Printf("[AFTER] Agent: %s, Messages: %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
)
```

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. The `...StructuredAgentOption[Output]` parameter is variadic, so existing code without hooks continues to work without any changes.

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
agent.Kind()       // Returns agents.Structured
agent.GetName()    // Returns the agent name from config
agent.GetModelID() // Returns the model name from model config
```

---

## 10. JSON Export and Debugging

### Export conversation to JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
if err != nil {
    // handle error
}
fmt.Println(jsonStr)
```

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
func NewAgent[Output any](
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    opts ...StructuredAgentOption[Output],
) (*Agent[Output], error)
```

Creates a new structured data agent. The type parameter `Output` defines the expected output struct. The `opts` parameter accepts zero or more `StructuredAgentOption[Output]` functional options.

---

### Types

```go
// StructuredAgentOption is a functional option for configuring an Agent during creation
type StructuredAgentOption[Output any] func(*Agent[Output])

// StructuredResult represents the result of structured data generation
type StructuredResult[Output any] struct {
    Data         *Output
    FinishReason string
}
```

---

### Option Functions

| Function | Description |
|---|---|
| `BeforeCompletion[Output any](fn func(*Agent[Output]))` | Sets a hook called before each structured data generation. |
| `AfterCompletion[Output any](fn func(*Agent[Output]))` | Sets a hook called after each structured data generation. |

---

### Methods

| Method | Description |
|---|---|
| `GenerateStructuredData(msgs []messages.Message) (*Output, string, error)` | Generate structured data from messages. Returns the typed output, finish reason, and error. |
| `GetMessages() []messages.Message` | Get all conversation messages. |
| `AddMessage(role roles.Role, content string)` | Add a single message to history. |
| `AddMessages(msgs []messages.Message)` | Add multiple messages to history. |
| `ResetMessages()` | Clear all messages except system instruction. |
| `ExportMessagesToJSON() (string, error)` | Export conversation history as JSON. |
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
| `Kind() agents.Kind` | Returns `agents.Structured`. |
| `GetName() string` | Returns the agent name. |
| `GetModelID() string` | Returns the model name. |
