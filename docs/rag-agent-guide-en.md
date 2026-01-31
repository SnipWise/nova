# RAG Agent Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Quick Start](#2-quick-start)
3. [Agent Configuration](#3-agent-configuration)
4. [Model Configuration](#4-model-configuration)
5. [Generating Embeddings](#5-generating-embeddings)
6. [Saving Embeddings](#6-saving-embeddings)
7. [Searching for Similar Content](#7-searching-for-similar-content)
8. [Store Persistence](#8-store-persistence)
9. [Chunking Utilities](#9-chunking-utilities)
10. [Options: AgentOption and RagAgentOption](#10-options-agentoption-and-ragagentoption)
11. [Lifecycle Hooks (RagAgentOption)](#11-lifecycle-hooks-ragagentoption)
12. [Context and State Management](#12-context-and-state-management)
13. [JSON Export and Debugging](#13-json-export-and-debugging)
14. [API Reference](#14-api-reference)

---

## 1. Introduction

### What is a RAG Agent?

The `rag.Agent` is a specialized agent provided by the Nova SDK (`github.com/snipwise/nova`) that handles Retrieval-Augmented Generation (RAG) workflows. It generates vector embeddings from text content and provides similarity search over an in-memory vector store.

Unlike chat or structured agents that use the Chat Completions API, the RAG agent uses the **Embeddings API** to convert text into numerical vectors, then uses cosine similarity to find semantically similar content.

### When to use a RAG Agent

| Scenario | Recommended agent |
|---|---|
| Generate vector embeddings from text | `rag.Agent` |
| Semantic similarity search | `rag.Agent` |
| Build a knowledge base for contextual retrieval | `rag.Agent` |
| Free-form conversational AI | `chat.Agent` |
| Structured data extraction | `structured.Agent[T]` |
| Function calling / tool use | `tools.Agent` |
| Intent detection and routing | `orchestrator.Agent` |
| Context compression | `compressor.Agent` |

### Key capabilities

- **Embedding generation**: Convert text content into vector embeddings using any OpenAI-compatible embedding model.
- **In-memory vector store**: Save and manage embeddings with automatic ID generation.
- **Similarity search**: Find semantically similar content using cosine similarity with configurable thresholds.
- **Top-N search**: Retrieve the top N most similar results above a threshold.
- **Store persistence**: Save and load the vector store to/from JSON files.
- **Chunking utilities**: Built-in text chunking helpers for splitting documents before embedding.
- **Lifecycle hooks**: Execute custom logic before and after each embedding generation.

---

## 2. Quick Start

### Minimal example

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := rag.NewAgent(
        ctx,
        agents.Config{
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        },
        models.Config{
            Name: "ai/mxbai-embed-large",
        },
    )
    if err != nil {
        panic(err)
    }

    // Generate an embedding
    embedding, err := agent.GenerateEmbedding("James T Kirk is the captain of the USS Enterprise.")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Embedding dimension: %d\n", len(embedding))

    // Save documents to the vector store
    agent.SaveEmbedding("Spock is the science officer aboard the Enterprise.")
    agent.SaveEmbedding("Leonard McCoy is the chief medical officer.")

    // Search for similar content
    results, err := agent.SearchSimilar("Who is the doctor?", 0.5)
    if err != nil {
        panic(err)
    }

    for _, r := range results {
        fmt.Printf("Match: %s (similarity: %.4f)\n", r.Prompt, r.Similarity)
    }
}
```

---

## 3. Agent Configuration

```go
agents.Config{
    Name:      "RAG",                                              // Agent name (optional)
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",      // LLM engine URL (required)
    APIKey:    "your-api-key",                                     // API key (optional)
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `Name` | `string` | No | Agent identifier for logging. |
| `EngineURL` | `string` | Yes | URL of the OpenAI-compatible LLM engine. |
| `APIKey` | `string` | No | API key for authenticated engines. |

**Note:** Unlike chat or structured agents, the RAG agent does not use `SystemInstructions` since it works with the Embeddings API, not Chat Completions.

---

## 4. Model Configuration

```go
models.Config{
    Name: "ai/mxbai-embed-large",    // Embedding model ID (required)
}
```

### Recommended models

- **mxbai-embed-large**: Good general-purpose embedding model with 1024 dimensions.
- Choose a model that matches your semantic search needs and available resources.

---

## 5. Generating Embeddings

### GenerateEmbedding

Generate a vector embedding for a given text:

```go
embedding, err := agent.GenerateEmbedding("Some text content")
if err != nil {
    // handle error
}

fmt.Printf("Dimension: %d\n", len(embedding)) // e.g., 1024
fmt.Printf("First value: %f\n", embedding[0])
```

**Return values:**
- `[]float64`: The embedding vector.
- `error`: Error if embedding generation failed.

### GetEmbeddingDimension

Get the dimension of the embedding vectors produced by the model:

```go
dimension := agent.GetEmbeddingDimension()
fmt.Printf("Embedding dimension: %d\n", dimension) // e.g., 1024
```

**Note:** This method makes a test API call to determine the dimension.

---

## 6. Saving Embeddings

### SaveEmbedding / SaveEmbeddingIntoMemoryVectorStore

Generate an embedding and save it to the in-memory vector store:

```go
err := agent.SaveEmbedding("Spock is a half-Vulcan science officer.")
if err != nil {
    // handle error
}
```

Each saved embedding is automatically assigned a unique ID. The store maps content to its vector representation for later similarity search.

### Saving multiple documents

```go
documents := []string{
    "James T Kirk is the captain of the Enterprise.",
    "Spock is the science officer.",
    "Leonard McCoy is the chief medical officer.",
}

for _, doc := range documents {
    err := agent.SaveEmbedding(doc)
    if err != nil {
        fmt.Printf("Failed to save: %v\n", err)
    }
}
```

---

## 7. Searching for Similar Content

### SearchSimilar

Search for all documents above a similarity threshold:

```go
results, err := agent.SearchSimilar("Who is the doctor?", 0.5)
if err != nil {
    // handle error
}

for _, r := range results {
    fmt.Printf("Content: %s\n", r.Prompt)
    fmt.Printf("Similarity: %.4f\n", r.Similarity)
}
```

**Parameters:**
- `content string`: The query text to search for.
- `limit float64`: Minimum cosine similarity threshold (1.0 = exact match, 0.0 = no similarity).

### SearchTopN

Search for the top N most similar documents above a threshold:

```go
results, err := agent.SearchTopN("Who is the captain?", 0.5, 3)
if err != nil {
    // handle error
}
```

**Parameters:**
- `content string`: The query text.
- `limit float64`: Minimum cosine similarity threshold.
- `n int`: Maximum number of results to return.

### VectorRecord

Search results are returned as `[]VectorRecord`:

```go
type VectorRecord struct {
    ID         string
    Prompt     string
    Embedding  []float64
    Metadata   map[string]any
    Similarity float64
}
```

---

## 8. Store Persistence

### Saving the store to disk

```go
err := agent.PersistStore("./store/knowledge.json")
if err != nil {
    // handle error
}
```

### Loading the store from disk

```go
err := agent.LoadStore("./store/knowledge.json")
if err != nil {
    // handle error
}
```

### Checking if a store file exists

```go
if agent.StoreFileExists("./store/knowledge.json") {
    agent.LoadStore("./store/knowledge.json")
} else {
    // Build the store from scratch
}
```

### Typical persistence workflow

```go
storeFile := "./store/data.json"

if agent.StoreFileExists(storeFile) {
    agent.LoadStore(storeFile)
} else {
    // Save documents
    for _, doc := range documents {
        agent.SaveEmbedding(doc)
    }
    // Persist for next run
    agent.PersistStore(storeFile)
}
```

---

## 9. Chunking Utilities

The `chunks` subpackage provides utilities for splitting documents before embedding.

### ChunkText

Split text into fixed-size chunks with overlap:

```go
import "github.com/snipwise/nova/nova-sdk/agents/rag/chunks"

pieces := chunks.ChunkText(longText, 512, 64) // chunkSize=512, overlap=64
for _, piece := range pieces {
    agent.SaveEmbedding(piece)
}
```

### SplitMarkdownBySections

Split Markdown content by sections (headers):

```go
sections := chunks.SplitMarkdownBySections(markdownContent)
for _, section := range sections {
    agent.SaveEmbedding(section)
}
```

---

## 10. Options: AgentOption and RagAgentOption

The RAG agent supports two distinct option types, both passed as variadic `...any` arguments to `NewAgent`:

### AgentOption (base-level)

`AgentOption` operates on the internal `*BaseAgent` and configures low-level behavior:

```go
// Currently available for extensibility
```

### RagAgentOption (agent-level)

`RagAgentOption` operates on the high-level `*Agent` and configures lifecycle hooks:

```go
rag.BeforeCompletion(func(a *rag.Agent) { ... })
rag.AfterCompletion(func(a *rag.Agent) { ... })
```

### Mixing both option types

Both option types can be passed together to `NewAgent`:

```go
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    // RagAgentOption (agent-level)
    rag.BeforeCompletion(func(a *rag.Agent) {
        fmt.Println("Before embedding generation...")
    }),
    rag.AfterCompletion(func(a *rag.Agent) {
        fmt.Println("After embedding generation...")
    }),
)
```

---

## 11. Lifecycle Hooks (RagAgentOption)

Lifecycle hooks allow you to execute custom logic before and after each embedding generation via the `GenerateEmbedding` method. They are configured as functional options when creating the agent.

### RagAgentOption

```go
type RagAgentOption func(*Agent)
```

### BeforeCompletion

Called before each embedding generation in `GenerateEmbedding`. The hook receives a reference to the agent.

```go
rag.BeforeCompletion(func(a *rag.Agent) {
    fmt.Printf("About to generate embedding... Agent: %s (%s)\n",
        a.GetName(), a.GetModelID())
})
```

**Use cases:**
- Logging and monitoring
- Metrics collection (e.g., count embedding generations)
- Rate limiting or throttling

### AfterCompletion

Called after each embedding generation in `GenerateEmbedding`. The hook receives a reference to the agent.

```go
rag.AfterCompletion(func(a *rag.Agent) {
    fmt.Printf("Embedding generated. Agent: %s (%s)\n",
        a.GetName(), a.GetModelID())
})
```

**Use cases:**
- Logging results
- Post-generation metrics
- Triggering downstream actions
- Auditing/tracking

### Complete example with hooks

```go
embeddingCount := 0

agent, err := rag.NewAgent(
    ctx,
    agents.Config{
        Name:      "RAG",
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    },
    models.Config{
        Name: "ai/mxbai-embed-large",
    },
    rag.BeforeCompletion(func(a *rag.Agent) {
        embeddingCount++
        fmt.Printf("[BEFORE] Agent: %s, Embedding #%d\n", a.GetName(), embeddingCount)
    }),
    rag.AfterCompletion(func(a *rag.Agent) {
        fmt.Printf("[AFTER] Agent: %s, Embedding #%d\n", a.GetName(), embeddingCount)
    }),
)
```

### Important note on hook scope

The hooks are triggered only by direct calls to `GenerateEmbedding`. Other methods like `SaveEmbedding`, `SearchSimilar`, and `SearchTopN` use the internal `BaseAgent.GenerateEmbeddingVector` directly and do **not** trigger the hooks.

### Hooks are optional

If no hooks are provided, the agent behaves exactly as before. The `...any` parameter is variadic, so existing code without hooks continues to work without any changes.

---

## 12. Context and State Management

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
agent.Kind()       // Returns agents.Rag
agent.GetName()    // Returns the agent name from config
agent.GetModelID() // Returns the model name from model config
```

---

## 13. JSON Export and Debugging

### Raw request/response JSON

```go
// Raw (unformatted) JSON of the last embedding request/response
rawReq := agent.GetLastRequestRawJSON()
rawResp := agent.GetLastResponseRawJSON()

// Pretty-printed JSON
prettyReq, err := agent.GetLastRequestJSON()
prettyResp, err := agent.GetLastResponseJSON()
```

---

## 14. API Reference

### Constructor

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    options ...any,
) (*Agent, error)
```

Creates a new RAG agent. The `options` parameter accepts both `AgentOption` (base-level) and `RagAgentOption` (agent-level hooks). The constructor separates them internally using type assertion.

---

### Types

```go
// VectorRecord represents a vector record with prompt and embedding
type VectorRecord struct {
    ID         string
    Prompt     string
    Embedding  []float64
    Metadata   map[string]any
    Similarity float64
}

// RagAgentOption configures the high-level Agent (e.g., lifecycle hooks)
type RagAgentOption func(*Agent)

// AgentOption configures the internal BaseAgent
type AgentOption func(*BaseAgent)
```

---

### Option Functions

| Function | Type | Description |
|---|---|---|
| `BeforeCompletion(fn func(*Agent))` | `RagAgentOption` | Sets a hook called before each embedding generation in `GenerateEmbedding`. |
| `AfterCompletion(fn func(*Agent))` | `RagAgentOption` | Sets a hook called after each embedding generation in `GenerateEmbedding`. |

---

### Methods

| Method | Description |
|---|---|
| `GenerateEmbedding(content string) ([]float64, error)` | Generate a vector embedding for the given text. Triggers lifecycle hooks. |
| `GetEmbeddingDimension() int` | Get the dimension of embedding vectors produced by the model. |
| `SaveEmbedding(content string) error` | Generate and save an embedding to the in-memory vector store. |
| `SaveEmbeddingIntoMemoryVectorStore(content string) error` | Alias for `SaveEmbedding`. |
| `SearchSimilar(content string, limit float64) ([]VectorRecord, error)` | Search for similar records above a similarity threshold. |
| `SearchTopN(content string, limit float64, n int) ([]VectorRecord, error)` | Search for top N similar records above a threshold. |
| `LoadStore(path string) error` | Load the vector store from a JSON file. |
| `PersistStore(path string) error` | Save the vector store to a JSON file. |
| `StoreFileExists(path string) bool` | Check if a store file exists at the given path. |
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
| `Kind() agents.Kind` | Returns `agents.Rag`. |
| `GetName() string` | Returns the agent name. |
| `GetModelID() string` | Returns the model name. |
