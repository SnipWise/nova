# Nova SDK API Reference

Quick reference for the main Nova SDK APIs.

## Imports

```go
import (
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/agents/structured"
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)
```

## Agent Configuration

### agents.Config

```go
agents.Config{
    Name:               string,  // Agent name (required)
    EngineURL:          string,  // LLM engine URL (required)
    SystemInstructions: string,  // System prompt
}
```

### models.Config

```go
models.Config{
    Name:              string,              // Model name (required)
    Temperature:       *float64,            // 0.0-1.0
    MaxTokens:         *int,                // Max response tokens
    ParallelToolCalls: *bool,               // Enable parallel tools
}

// Helper functions for pointers
models.Float64(0.7)  // Returns *float64
models.Int(2000)     // Returns *int
models.Bool(true)    // Returns *bool
```

## Chat Agent

### Creation

```go
agent, err := chat.NewAgent(
    ctx,
    agents.Config{...},
    models.Config{...},
)
```

### Methods

```go
// Simple completion
result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Hello"},
})
// result.Response - Response text
// result.FinishReason - Finish reason

// Streaming completion
result, err := agent.GenerateStreamCompletion(
    []messages.Message{{Role: roles.User, Content: "Hello"}},
    func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        return nil
    },
)
```

## RAG Agent

### Creation

```go
agent, err := rag.NewAgent(
    ctx,
    agents.Config{...},
    models.Config{
        Name: "ai/mxbai-embed-large", // Embedding model
    },
)
```

### Methods

```go
// Index a document
err := agent.SaveEmbedding("text to index")

// Semantic search
similarities, err := agent.SearchSimilar(query, threshold)
// threshold: 0.0 to 1.0 (minimum similarity)

// Result structure
type Similarity struct {
    Prompt     string  // Original text
    Similarity float64 // Similarity score
}
```

## Tools Agent

### Tool Creation

```go
tool := tools.NewTool("tool_name").
    SetDescription("Tool description").
    AddParameter("param1", "string", "Parameter description", true).  // required
    AddParameter("param2", "number", "Optional param", false).        // optional
    AddEnumParameter("status", "string", "Status", []string{"a", "b"}, true)
```

### Agent Creation

```go
agent, err := tools.NewAgent(
    ctx,
    agents.Config{...},
    models.Config{
        Name:              "model-with-tools-support",
        Temperature:       models.Float64(0.0),
        ParallelToolCalls: models.Bool(true),
    },
    tools.WithTools([]*tools.Tool{tool1, tool2}),
)
```

### Methods

```go
// Detection with execution loop
result, err := agent.DetectToolCallsLoop(
    []messages.Message{{Role: roles.User, Content: "..."}},
    func(toolName, argsJSON string) (string, error) {
        // Execute tool and return JSON result
        return `{"result": "value"}`, nil
    },
)
// result.Results - Tool execution results
// result.LastAssistantMessage - Final response
// result.FinishReason - Finish reason

// Simple detection (without execution)
result, err := agent.DetectToolCalls(messages)
// result.ToolCalls - Detected tools
```

## Structured Agent

### With Go Struct

```go
type MyOutput struct {
    Field1 string   `json:"field1" description:"Field description"`
    Field2 int      `json:"field2"`
    Tags   []string `json:"tags,omitempty"`
}

agent, err := structured.NewAgent(
    ctx,
    agents.Config{...},
    models.Config{Temperature: models.Float64(0.0)},
    structured.WithOutputSchema(MyOutput{}),
)
```

### With JSON Schema

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "name": map[string]interface{}{
            "type": "string",
            "minLength": 1,
        },
        "status": map[string]interface{}{
            "type": "string",
            "enum": []string{"active", "inactive"},
        },
    },
    "required": []string{"name", "status"},
}

agent, err := structured.NewAgent(
    ctx,
    agents.Config{...},
    models.Config{...},
    structured.WithJSONSchema(schema),
)
```

## Compressor Agent

### Creation

```go
agent, err := compressor.NewAgent(
    ctx,
    agents.Config{
        Name:               "compressor",
        EngineURL:          engineURL,
        SystemInstructions: "You are an expert at summarizing conversations...",
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0), // Deterministic
    },
)
```

### Methods

```go
// Compress text and return summary
summary, err := agent.Compress(textToCompress)
// summary: string - compressed summary of the input text
```

## Messages

### Roles

```go
roles.System    // System message
roles.User      // User message
roles.Assistant // Assistant message
```

### Message Structure

```go
messages.Message{
    Role:    roles.User,
    Content: "Message content",
}
```

## Recommended Models

| Use Case | Model | Notes |
|----------|-------|-------|
| Chat | `ai/qwen2.5:1.5B-F16` | Good quality/speed ratio |
| Embeddings | `ai/mxbai-embed-large` | For RAG |
| Tools | `hf.co/menlo/jan-nano-gguf:q4_k_m` | Function calling support |
| Structured | `ai/qwen2.5:1.5B-F16` | With temperature 0.0 |

## Engine URLs

```yaml
# llama.cpp (default)
http://localhost:12434/engines/llama.cpp/v1

# Ollama
http://localhost:11434/v1

# LM Studio
http://localhost:1234/v1

# OpenAI compatible
https://api.openai.com/v1
```
