# Compressor Agent

## Description

The **Compressor Agent** is a specialized agent for compressing conversation context. It takes a list of messages and generates a concise summary that preserves essential information while reducing context size.

## Features

- **Context compression** : Summarizes long conversations while preserving key facts
- **Streaming** : Generate summary with streaming or in one shot
- **Customizable prompts** : Multiple predefined compression prompts and ability to create custom prompts
- **Configurable instructions** : Predefined system instructions for different compression styles

## Creating a Compressor Agent

### Basic syntax

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

// Agent configuration
agentConfig := agents.Config{
    Name: "Compressor",
    Instructions: compressor.Instructions.Minimalist,
}

// Model configuration
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    ModelID: "qwen2.5:1.5b",
}

// Create agent with default prompt (Minimalist)
agent, err := compressor.NewAgent(ctx, agentConfig, modelConfig)

// Create agent with a custom prompt
agent, err := compressor.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    compressor.WithCompressionPrompt(compressor.Prompts.Structured),
)
```

### Available options

| Option | Description |
|--------|-------------|
| `WithCompressionPrompt(prompt)` | Sets the compression prompt to use |

## Predefined system instructions

The package provides three predefined system instructions:

### `Instructions.Minimalist` (recommended default)
```
You are a context compression assistant. Your task is to summarize
conversations concisely, preserving key facts, decisions, and context
needed for continuation.
```

### `Instructions.Expert`
Detailed instructions with:
- Preservation of critical information
- Redundancy elimination
- Chronology maintenance
- Structured output format
- Specific compression guidelines

### `Instructions.Effective`
Structured format with sections:
- Conversation Summary
- Key Points
- To Remember

## Predefined compression prompts

The package provides four compression prompts:

| Prompt | Description | Use case |
|--------|-------------|----------|
| `Prompts.Minimalist` ⭐ | Concise summary preserving key facts, decisions and context | **Recommended** - General usage |
| `Prompts.Structured` | Structured format with topics, decisions, context (< 200 words) | Organized summaries |
| `Prompts.UltraShort` | Extract facts, decisions and essential context only | Maximum compression |
| `Prompts.ContinuityFocus` | Preserve all information needed to continue naturally | Conversation continuity |

**Default prompt** : `Prompts.Minimalist`

## Main methods

### Compression without streaming

```go
// Compress a list of messages
result, err := agent.CompressContext(messagesList)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Compressed text:", result.CompressedText)
fmt.Println("Finish reason:", result.FinishReason)
```

**Return** : `*CompressionResult`
- `CompressedText` : The compressed text
- `FinishReason` : The finish reason (usually "stop")

### Compression with streaming

```go
// Compress with streaming
result, err := agent.CompressContextStream(messagesList, func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
})
if err != nil {
    log.Fatal(err)
}

fmt.Println("\nFinal compressed text:", result.CompressedText)
```

### Change compression prompt

```go
// Change prompt after creation
agent.SetCompressionPrompt(compressor.Prompts.UltraShort)

// Or use a custom prompt
customPrompt := "Summarize this conversation in 3 sentences maximum."
agent.SetCompressionPrompt(customPrompt)
```

### Getters and Setters

```go
// Configuration
config := agent.GetConfig()
agent.SetConfig(newConfig)

modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)

// Information
name := agent.GetName()
modelID := agent.GetModelID()
kind := agent.GetKind() // Returns agents.Compressor

// Context
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requests/Responses (debugging)
lastRequestJSON, _ := agent.GetLastRequestJSON()
lastResponseJSON, _ := agent.GetLastResponseJSON()
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
```

## Complete example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Configuration
    agentConfig := agents.Config{
        Name:         "Compressor",
        Instructions: compressor.Instructions.Minimalist,
    }
    modelConfig := models.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        ModelID:   "qwen2.5:1.5b",
    }

    // Create agent with structured prompt
    agent, err := compressor.NewAgent(
        ctx,
        agentConfig,
        modelConfig,
        compressor.WithCompressionPrompt(compressor.Prompts.Structured),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Messages to compress
    messagesList := []messages.Message{
        {Role: roles.User, Content: "Hello, I want to create a REST API."},
        {Role: roles.Assistant, Content: "Sure! Which language do you prefer?"},
        {Role: roles.User, Content: "I'd like to use Go."},
        {Role: roles.Assistant, Content: "Excellent choice. Here's how to create a REST API in Go..."},
        // ... many more messages
    }

    // Compression with streaming
    fmt.Println("🗜️  Compressing context...")
    result, err := agent.CompressContextStream(messagesList, func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        if finishReason != "" {
            fmt.Printf("\n[Finish: %s]\n", finishReason)
        }
        return nil
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n✅ Compression complete. Original: %d messages → Compressed: %d chars\n",
        len(messagesList), len(result.CompressedText))
}
```

## Usage with other agents

The Compressor Agent is typically used with Server, Crew or Chat agents to automatically manage context compression:

```go
// With Server Agent
serverAgent, _ := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)

// With Crew Agent
crewAgent, _ := crew.NewAgent(
    ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)

// Compression happens automatically when limit is reached
```

## Compression format

The Compressor Agent:
1. Converts messages to text format:
   ```
   user: User message
   assistant: Assistant response
   system: System message
   ```
2. Sends the text with the compression prompt
3. Returns the summary generated by the model

## Notes

- **Kind** : Returns `agents.Compressor`
- **Streaming** : Uses OpenAI SDK internally for streaming
- **Default prompt** : `Prompts.Minimalist`
- **Default instructions** : No default instructions - must be set in `agentConfig.Instructions`
- **Empty error** : Returns an error if `messagesList` is empty
- **Automatic conversion** : Messages are automatically converted to OpenAI format internally

## Recommendations

- **Recommended prompt** : `Prompts.Minimalist` for most cases
- **Recommended instructions** : `Instructions.Minimalist` for general use, `Instructions.Expert` for advanced compression
- **Streaming** : Use `CompressContextStream` to see progress in real-time
- **Context size** : Configure an appropriate limit (e.g., 8000 characters) when using with Server/Crew agents
