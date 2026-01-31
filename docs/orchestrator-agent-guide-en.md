# Orchestrator Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration](#3-agent-configuration)
4. [Model Configuration](#4-model-configuration)
5. [Intent Detection Methods](#5-intent-detection-methods)
6. [Conversation History and Messages](#6-conversation-history-and-messages)
7. [Lifecycle Hooks (OrchestratorAgentOption)](#7-lifecycle-hooks-orchestratoragentoption)
8. [Usage with Crew Agent](#8-usage-with-crew-agent)
9. [Context and State Management](#9-context-and-state-management)
10. [JSON Debugging](#10-json-debugging)
11. [API Reference](#11-api-reference)

---

## 1. Introduction

### What is an Orchestrator Agent?

The `orchestrator.Agent` is a specialized agent provided by the Nova SDK (`github.com/snipwise/nova`) for topic and intent detection from user messages. It wraps a `structured.Agent[agents.Intent]` internally to generate structured JSON output containing the identified discussion topic.

Unlike a chat agent that generates free-form text responses, the orchestrator agent always returns a structured `Intent` object with a `TopicDiscussion` field. This makes it ideal for routing, classification, and decision-making in multi-agent systems.

### When to use an Orchestrator Agent

| Scenario | Recommended agent |
|---|---|
| Topic/intent classification from user input | `orchestrator.Agent` |
| Routing queries to specialized agents in a crew | `orchestrator.Agent` with `CrewServerAgent` |
| Simple question classification | `orchestrator.Agent` with `IdentifyTopicFromText` |
| Free-form conversational AI | `chat.Agent` |
| Function calling / tool use | `tools.Agent` |

### Key capabilities

- **Topic detection**: Identifies the main topic of a conversation from user messages.
- **Structured output**: Always returns an `Intent` object with a `TopicDiscussion` field in JSON format.
- **Intent routing**: Designed to work with Crew Agents for automatic query routing to specialized agents.
- **Lifecycle hooks**: Execute custom logic before and after each intent identification.
- **Lightweight**: Uses fast, small models for quick classification without requiring large language models.

---

## 2. Quick Start

### Minimal example

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := orchestrator.NewAgent(
        ctx,
        agents.Config{
            Name:      "topic-detector",
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: `You are good at identifying the topic of a conversation.
Given a user's input, identify the main topic of discussion in only one word.
The possible topics are: Technology, Health, Sports, Entertainment, Science.
Respond in JSON format with the field 'topic_discussion'.`,
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.0),
        },
    )
    if err != nil {
        panic(err)
    }

    topic, err := agent.IdentifyTopicFromText("How does quantum computing work?")
    if err != nil {
        panic(err)
    }

    fmt.Println("Detected topic:", topic)
    // Output: Detected topic: Technology
}
```

---

## 3. Agent Configuration

The `agents.Config` struct controls the agent's identity and behavior:

```go
agents.Config{
    Name:                    "orchestrator-agent",  // Agent name (optional)
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1", // LLM engine URL (required)
    APIKey:                  "your-api-key",        // API key (optional)
    SystemInstructions:      "...",                 // Classification instructions (critical)
    KeepConversationHistory: false,                 // Usually false for classification
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `Name` | `string` | No | Agent identifier for logging and multi-agent setups. |
| `EngineURL` | `string` | Yes | URL of the OpenAI-compatible LLM engine. |
| `APIKey` | `string` | No | API key for authenticated engines. |
| `SystemInstructions` | `string` | Critical | Instructions that define the classification categories and format. This is the most important field -- the quality of topic detection depends directly on these instructions. |
| `KeepConversationHistory` | `bool` | No | Usually `false` for stateless classification. Default: `false`. |

### Writing effective system instructions

The system instructions must clearly define:
1. The possible topic categories
2. The expected output format (JSON with `topic_discussion` field)
3. Any classification rules or priorities

```go
SystemInstructions: `You are a topic classification assistant.
Analyze the user's message and identify the main topic of discussion.

Categories:
- coding: programming, software development, debugging
- cooking: recipes, food, ingredients
- philosophy: ideas, concepts, thinking
- science: physics, chemistry, biology
- generic: everything else

Return only the topic category in JSON format with the field 'topic_discussion'.`
```

---

## 4. Model Configuration

The `models.Config` struct controls the model's generation parameters:

```go
models.Config{
    Name:        "ai/qwen2.5:1.5B-F16",    // Model ID (required)
    Temperature: models.Float64(0.0),        // Use 0.0 for deterministic classification
}
```

### Recommended settings for classification

- **Temperature**: `0.0` -- Deterministic output is critical for consistent classification.
- **Model size**: Use small, fast models (1-3B parameters). Topic detection does not require large models.
- **Recommended models**: `qwen2.5:1.5b`, `lucy`, `jan-nano` -- fast and sufficient for classification.

---

## 5. Intent Detection Methods

### IdentifyIntent

The primary method for detecting topics from messages. Returns the full `Intent` object along with the finish reason.

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

intent, finishReason, err := agent.IdentifyIntent([]messages.Message{
    {Role: roles.User, Content: "How to make a Neapolitan pizza?"},
})
if err != nil {
    // handle error
}

fmt.Println("Topic:", intent.TopicDiscussion) // "cooking" or "food"
fmt.Println("Finish reason:", finishReason)   // "stop"
```

**Return type:** `*agents.Intent`

```go
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}
```

### IdentifyTopicFromText

A convenience method that takes a plain text string and returns only the detected topic. It internally creates a user message and calls `IdentifyIntent`.

```go
topic, err := agent.IdentifyTopicFromText("Explain the Factory pattern in Go")
if err != nil {
    // handle error
}

fmt.Println("Detected topic:", topic) // "coding"
```

### OrchestratorAgent interface

The orchestrator agent implements the `agents.OrchestratorAgent` interface:

```go
type OrchestratorAgent interface {
    IdentifyIntent(userMessages []messages.Message) (intent *Intent, finishReason string, err error)
    IdentifyTopicFromText(text string) (string, error)
}
```

This interface is used by `CrewServerAgent` for automatic query routing.

---

## 6. Conversation History and Messages

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

### When to use conversation history

For most classification use cases, `KeepConversationHistory` should be `false`. Each classification call is typically independent. However, in some scenarios you may want history enabled -- for example, if the orchestrator needs to consider previous messages to classify ambiguous queries.

---

## 7. Lifecycle Hooks (OrchestratorAgentOption)

Lifecycle hooks allow you to execute custom logic before and after each intent identification. They are configured as functional options when creating the agent.

### OrchestratorAgentOption

```go
type OrchestratorAgentOption func(*Agent)
```

Options are passed as variadic arguments to `NewAgent`:

```go
agent, err := orchestrator.NewAgent(ctx, agentConfig, modelConfig,
    orchestrator.BeforeCompletion(fn),
    orchestrator.AfterCompletion(fn),
)
```

### BeforeCompletion

Called before each intent identification (`IdentifyIntent` and, by extension, `IdentifyTopicFromText`). The hook receives a reference to the agent.

```go
orchestrator.BeforeCompletion(func(a *orchestrator.Agent) {
    fmt.Println("About to identify intent...")
    fmt.Printf("Agent: %s (%s)\n", a.GetName(), a.GetModelID())
    fmt.Printf("Messages count: %d\n", len(a.GetMessages()))
})
```

**Use cases:**
- Logging and monitoring
- Metrics collection (count classification requests)
- Pre-classification state inspection

### AfterCompletion

Called after each intent identification, once the result has been received. The hook receives a reference to the agent.

```go
orchestrator.AfterCompletion(func(a *orchestrator.Agent) {
    fmt.Println("Intent identification completed.")
    fmt.Printf("Messages count: %d\n", len(a.GetMessages()))
})
```

**Use cases:**
- Logging classification results
- Post-classification metrics
- Triggering downstream actions based on classification
- Auditing/tracking topic distribution

### Complete example with hooks

```go
agent, err := orchestrator.NewAgent(
    ctx,
    agents.Config{
        Name:      "orchestrator-agent",
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: `You are good at identifying the topic of a conversation.
Given a user's input, identify the main topic of discussion in only one word.
Respond in JSON format with the field 'topic_discussion'.`,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
    },
    orchestrator.BeforeCompletion(func(a *orchestrator.Agent) {
        fmt.Printf("[BEFORE] Agent: %s, Messages: %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
    orchestrator.AfterCompletion(func(a *orchestrator.Agent) {
        fmt.Printf("[AFTER] Agent: %s, Messages: %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
)
```

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. Hooks are only called when they have been set. The `...OrchestratorAgentOption` parameter is variadic, so existing code without hooks continues to work without any changes.

### Hooks apply to both detection methods

Since `IdentifyTopicFromText` calls `IdentifyIntent` internally, hooks are triggered for both methods:

| Method | BeforeCompletion | AfterCompletion |
|---|---|---|
| `IdentifyIntent` | Yes | Yes |
| `IdentifyTopicFromText` | Yes (via IdentifyIntent) | Yes (via IdentifyIntent) |

---

## 8. Usage with Crew Agent

The orchestrator agent is designed to work with `CrewServerAgent` for automatic query routing in multi-agent systems.

### Basic crew routing

```go
import (
    "strings"
    "github.com/snipwise/nova/nova-sdk/agents/crewserver"
)

// Create specialized chat agents
agentCrew := map[string]*chat.Agent{
    "expert":  expertAgent,
    "coder":   coderAgent,
    "thinker": thinkerAgent,
}

// Create orchestrator agent
orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:      "orchestrator-agent",
        EngineURL: engineURL,
        SystemInstructions: "Classify queries into: code_generation, complex_thinking, code_question.",
    },
    models.Config{
        Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
        Temperature: models.Float64(0.0),
    },
)

// Define routing function
matchFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "code_generation", "write code":
        return "coder"
    case "complex_thinking", "reasoning":
        return "thinker"
    default:
        return "expert"
    }
}

// Assemble the crew server
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "expert"),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.WithMatchAgentIdToTopicFn(matchFn),
    crewserver.WithPort(3500),
)

crewServerAgent.StartServer()
```

### How routing works

1. A user sends a message to the crew server.
2. The orchestrator agent classifies the message topic.
3. The match function maps the topic to an agent ID.
4. The crew server routes the request to the matched agent.
5. The matched agent generates the response.

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
agent.Kind()       // Returns agents.Orchestrator
agent.GetName()    // Returns the agent name from config
agent.GetModelID() // Returns the model name from model config
```

---

## 10. JSON Debugging

For debugging, you can access the raw JSON sent to and received from the LLM engine:

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
    opts ...OrchestratorAgentOption,
) (*Agent, error)
```

Creates a new orchestrator agent. The `opts` parameter accepts zero or more `OrchestratorAgentOption` functional options.

---

### Types

```go
// OrchestratorAgentOption is a functional option for configuring an Agent during creation
type OrchestratorAgentOption func(*Agent)

// Intent represents the structured output from the orchestrator
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}

// OrchestratorAgent is the interface implemented by the orchestrator agent
type OrchestratorAgent interface {
    IdentifyIntent(userMessages []messages.Message) (intent *Intent, finishReason string, err error)
    IdentifyTopicFromText(text string) (string, error)
}
```

---

### Option Functions

| Function | Description |
|---|---|
| `BeforeCompletion(fn func(*Agent))` | Sets a hook called before each intent identification. |
| `AfterCompletion(fn func(*Agent))` | Sets a hook called after each intent identification. |

---

### Methods

| Method | Description |
|---|---|
| `IdentifyIntent(msgs []messages.Message) (*agents.Intent, string, error)` | Identify intent from messages. Returns the intent, finish reason, and error. |
| `IdentifyTopicFromText(text string) (string, error)` | Convenience method to detect topic from a text string. |
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
| `Kind() agents.Kind` | Returns `agents.Orchestrator`. |
| `GetName() string` | Returns the agent name. |
| `GetModelID() string` | Returns the model name. |
