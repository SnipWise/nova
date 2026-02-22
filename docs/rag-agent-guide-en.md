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
9. [JSON Store with WithJsonStore](#9-json-store-with-withjsonstore)
10. [Document Initialization with WithDocuments](#10-document-initialization-with-withdocuments)
11. [Redis Vector Store](#11-redis-vector-store)
12. [Chunking Utilities](#12-chunking-utilities)
13. [Options: AgentOption and RagAgentOption](#13-options-agentoption-and-ragagentoption)
14. [Lifecycle Hooks (RagAgentOption)](#14-lifecycle-hooks-ragagentoption)
15. [Context and State Management](#15-context-and-state-management)
16. [JSON Export and Debugging](#16-json-export-and-debugging)
17. [API Reference](#17-api-reference)

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
- **Redis vector store**: Use Redis as a persistent backend with HNSW indexing for ultra-fast and scalable search.
- **Similarity search**: Find semantically similar content using cosine similarity with configurable thresholds.
- **Top-N search**: Retrieve the top N most similar results above a threshold.
- **Store persistence**: Save and load the vector store to/from JSON files (Memory) or Redis.
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

## 9. JSON Store with WithJsonStore

### Introduction

The `WithJsonStore` option provides a convenient way to automatically load and persist your vector store from a JSON file during agent creation. This eliminates the need for manual `LoadStore`/`PersistStore` calls in many common scenarios.

### Basic Usage

```go
agent, err := rag.NewAgent(
    ctx,
    agents.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    },
    models.Config{
        Name: "ai/mxbai-embed-large:latest",
    },
    rag.WithJsonStore("./store/embeddings.json"),
)
```

### How it works

1. **On Creation**: The agent attempts to load existing embeddings from the specified JSON file
2. **File Exists**: Data is loaded into memory automatically
3. **File Missing**: An empty in-memory store is created
4. **Automatic Persistence**: When combined with `WithDocuments`, the store is automatically persisted if new documents are added
5. **Manual Persistence**: You can still call `agent.PersistStore(filePath)` manually to save changes at any time

### ✨ Automatic Persistence (New!)

When you use `WithJsonStore` together with `WithDocuments`, the agent **automatically persists** the store if new documents are added during initialization:

```go
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithJsonStore("./store/embeddings.json"),
    rag.WithDocuments(documents),  // Automatically persists if documents are added!
)
// No need to call agent.PersistStore() manually!
```

**How it works:**
- If the store is empty or new documents are added → **automatic persistence**
- If using `DocumentLoadModeSkip` and the store already has data → no persistence (no new documents added)
- If using `DocumentLoadModeSkipDuplicates` → persists only if non-duplicate documents are added
- The parent directory is automatically created if it doesn't exist

### Complete Example

```go
package main

import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()
    storeFile := "./store/knowledge.json"

    // Create agent with JSON store - automatically loads if file exists
    agent, err := rag.NewAgent(
        ctx,
        agents.Config{
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        },
        models.Config{
            Name: "ai/mxbai-embed-large:latest",
        },
        rag.WithJsonStore(storeFile),
    )
    if err != nil {
        panic(err)
    }

    // Add new documents
    agent.SaveEmbedding("New document to add")

    // Save changes to disk
    agent.PersistStore(storeFile)
}
```

### Benefits

- ✅ Automatic loading on agent creation
- ✅ Clean, declarative configuration
- ✅ No need to check if file exists
- ✅ Seamless fallback to empty store
- ✅ Full control over when to persist

### When to use WithJsonStore

| Scenario | Recommended Approach |
|----------|---------------------|
| Simple JSON persistence | `WithJsonStore` ✅ |
| Manual load/persist control | `LoadStore`/`PersistStore` |
| Production, large datasets | `WithRedisStore` (see next section) |
| Temporary/testing | Default in-memory store |

---

## 10. Document Initialization with WithDocuments

### Introduction

The `WithDocuments` option allows you to initialize your RAG agent with a predefined list of documents. This is perfect for:
- Pre-loading a knowledge base
- Seeding the agent with initial data
- Simplifying agent setup with known content

### Basic Usage

```go
documents := []string{
    "Squirrels run in the forest",
    "Birds fly in the sky",
    "Frogs swim in the pond",
}

agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithDocuments(documents),
)
```

### Document Load Modes

When using `WithDocuments`, you can specify how to handle existing data in the store:

```go
type DocumentLoadMode string

const (
    DocumentLoadModeOverwrite  // Clear existing data and load new documents
    DocumentLoadModeMerge      // Add documents to existing data (default)
    DocumentLoadModeSkip       // Skip loading if store already has data
    DocumentLoadModeError      // Log error if store is not empty
)
```

### Usage with Different Modes

#### Merge Mode (Default)

Adds new documents to existing ones:

```go
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithJsonStore(storeFile),
    rag.WithDocuments(documents, rag.DocumentLoadModeMerge), // or just rag.WithDocuments(documents)
)
```

#### Overwrite Mode

Replaces all existing data:

```go
rag.WithDocuments(documents, rag.DocumentLoadModeOverwrite)
```

#### Skip Mode

Preserves existing data, skips loading if store has content:

```go
rag.WithDocuments(documents, rag.DocumentLoadModeSkip)
```

#### Error Mode

Prevents accidental overwrites by logging an error:

```go
rag.WithDocuments(documents, rag.DocumentLoadModeError)
```

### Complete Example with JSON Store

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
    storeFile := "./store/animals.json"

    // Initial knowledge base
    documents := []string{
        "Squirrels run in the forest",
        "Birds fly in the sky",
        "Frogs swim in the pond",
        "Fishes swim in the sea",
        "Lions roar in the savannah",
    }

    // Create agent with both JSON store and initial documents
    agent, err := rag.NewAgent(
        ctx,
        agents.Config{
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        },
        models.Config{
            Name: "ai/mxbai-embed-large:latest",
        },
        rag.WithJsonStore(storeFile),  // Load existing store (if it exists)
        rag.WithDocuments(documents, rag.DocumentLoadModeSkipDuplicates), // Add documents, skip duplicates
    )
    if err != nil {
        panic(err)
    }

    // ✨ No need to call agent.PersistStore() - it's automatic!
    // The store was automatically persisted when documents were added

    // Search
    results, _ := agent.SearchSimilar("What animals live in water?", 0.6)
    for _, r := range results {
        fmt.Printf("Match: %s (%.3f)\n", r.Prompt, r.Similarity)
    }
}
```

### ⚠️ Important: Option Order

Always apply `WithDocuments` **AFTER** `WithJsonStore` or `WithRedisStore`:

```go
// ✅ Correct order
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithJsonStore(storeFile),      // 1. Configure store first
    rag.WithDocuments(documents),      // 2. Then load documents
)

// ❌ Wrong order - documents will be loaded into default store, then overwritten by JsonStore
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithDocuments(documents),      // ❌ Loaded into default store
    rag.WithJsonStore(storeFile),      // ❌ Replaces the store, losing documents
)
```

### Persistence Behavior

#### With JSON Store

✨ **Automatic persistence** when combined with `WithDocuments`:

```go
agent, err := rag.NewAgent(
    ctx, agentConfig, modelConfig,
    rag.WithJsonStore(storeFile),
    rag.WithDocuments(documents),
)
// ✅ Automatically persisted if new documents were added!
```

**When automatic persistence occurs:**
- ✅ New documents added (first run or using Merge/Overwrite modes)
- ✅ Non-duplicate documents added (using SkipDuplicates mode)
- ❌ No new documents added (using Skip mode with existing data)

**Manual persistence still available:**
```go
// Add documents after agent creation
agent.SaveEmbedding("New document")
// Manually persist changes
agent.PersistStore(storeFile)
```

#### With Redis Store

Documents are automatically persisted to Redis:

```go
agent, err := rag.NewAgent(
    ctx, agentConfig, modelConfig,
    rag.WithRedisStore(redisConfig, dimension),
    rag.WithDocuments(documents),
)
// Documents are automatically saved to Redis
```

### Use Cases

| Use Case | Recommended Mode |
|----------|-----------------|
| First-time setup | `Overwrite` or `Merge` |
| Daily updates | `Merge` |
| Keep existing data unchanged | `Skip` |
| Prevent accidental changes | `Error` |
| Testing with clean slate | `Overwrite` |

### Sample Code

See `samples/110-rag-agent-with-json-store/` for a complete working example demonstrating both `WithJsonStore` and `WithDocuments` with different modes.

---

## 11. Redis Vector Store

### Introduction: Redis vs In-Memory

By default, the RAG Agent uses an **in-memory vector store** that stores embeddings in RAM. This is perfect for prototyping and small datasets, but the data is lost when the application restarts.

The **Redis Vector Store** offers a persistent and scalable alternative:
- 💾 **Persistence**: Data survives restarts
- 🔄 **Sharing**: Multiple applications can access the same data
- 📈 **Scalability**: Support for millions of vectors
- ⚡ **Performance**: HNSW indexing for ultra-fast search

### When to use Redis vs In-Memory

| Criterion | In-Memory | Redis |
|-----------|-----------|-------|
| **Persistence** | ❌ Lost on restart | ✅ Survives restarts |
| **Multi-process sharing** | ❌ Single process | ✅ Multiple applications |
| **Scalability** | Limited by RAM | Millions of vectors |
| **Speed** | Very fast | Very fast (HNSW) |
| **Setup** | None needed | Requires Redis |
| **Use case** | Prototyping, small datasets | Production, large datasets |

### Redis Configuration

To use Redis as a backend, you need to configure the connection via `RedisConfig`:

```go
type RedisConfig struct {
    Address   string // Redis server address (e.g., "localhost:6379")
    Password  string // Redis password (empty string for no password)
    DB        int    // Redis database number (default: 0)
    IndexName string // Redis search index name (default: "nova_rag_index")
}
```

### Using WithRedisStore

To create a RAG agent with Redis as the backend, use the `WithRedisStore` option:

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/agents/rag/stores"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

agent, err := rag.NewAgent(
    ctx,
    agents.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    },
    models.Config{
        Name: "ai/mxbai-embed-large", // 1024 dimensions
    },
    // Redis option
    rag.WithRedisStore(stores.RedisConfig{
        Address:   "localhost:6379",
        Password:  "",                    // Empty if no password
        DB:        0,                     // Default database
        IndexName: "my_knowledge_base",   // Custom index name
    }, 1024), // ⚠️ Dimension MUST match embedding model
)
if err != nil {
    panic(err)
}

// Usage is identical to in-memory store
agent.SaveEmbedding("James T Kirk is the captain of the USS Enterprise.")
agent.SaveEmbedding("Spock is the science officer.")

// Search
results, _ := agent.SearchSimilar("Who is the captain?", 0.5)
```

### ⚠️ Important: Embedding Dimension

The `dimension` parameter in `WithRedisStore` **MUST** match the dimension of vectors produced by your embedding model:

| Model | Dimension |
|-------|-----------|
| `ai/mxbai-embed-large` | 1024 |
| `text-embedding-3-small` | 1536 |
| `text-embedding-3-large` | 3072 |
| `text-embedding-ada-002` | 1536 |

You can verify the dimension with:
```go
dimension := agent.GetEmbeddingDimension()
fmt.Printf("Dimension: %d\n", dimension)
```

### Complete Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/agents/rag/stores"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Create agent with Redis
    agent, err := rag.NewAgent(
        ctx,
        agents.Config{
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        },
        models.Config{
            Name: "ai/mxbai-embed-large",
        },
        rag.WithRedisStore(stores.RedisConfig{
            Address:   "localhost:6379",
            Password:  "",
            DB:        0,
            IndexName: "star_trek_knowledge",
        }, 1024),
    )
    if err != nil {
        panic(err)
    }

    // Save documents
    documents := []string{
        "James T Kirk is the captain of the Enterprise.",
        "Spock is a half-Vulcan science officer.",
        "Leonard McCoy is the chief medical officer.",
        "Montgomery Scott is the chief engineer.",
    }

    for _, doc := range documents {
        err := agent.SaveEmbedding(doc)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    }

    // Search
    results, err := agent.SearchSimilar("Who is the doctor?", 0.5)
    if err != nil {
        panic(err)
    }

    for _, r := range results {
        fmt.Printf("Result: %s (similarity: %.4f)\n", r.Prompt, r.Similarity)
    }
}
```

### Prerequisites: Starting Redis

Redis must be running with vector search support (Redis Stack or RediSearch module):

```bash
# With Docker
docker run -d \
  --name redis-vector-store \
  -p 6379:6379 \
  redis/redis-stack-server:latest

# Verify Redis is running
docker exec -it redis-vector-store redis-cli ping
# Should return: PONG
```

### Inspecting Data in Redis

You can inspect stored data using Redis CLI:

```bash
# Access Redis CLI
docker exec -it redis-vector-store redis-cli

# List all indexes
FT._LIST

# View index details
FT.INFO my_knowledge_base

# List all document keys
KEYS doc:*

# View a specific document
HGETALL doc:<uuid>

# Count documents
DBSIZE
```

### Persistence and Restarts

The main advantage of Redis is **automatic persistence**:

```bash
# First run - save data
go run main.go

# Stop the program (Ctrl+C)

# Rerun - data is still there!
go run main.go
# Previously saved embeddings are accessible
```

To start fresh:
```bash
# Delete index and all data
docker exec -it redis-vector-store redis-cli
FT.DROPINDEX my_knowledge_base DD  # DD = delete documents
```

### Troubleshooting

#### Redis Connection Error

```
❌ Failed to create RAG agent: failed to connect to Redis: dial tcp [::1]:6379: connect: connection refused
```

**Solution**: Start Redis with the Docker command above.

#### Dimension Mismatch

```
Error: vector dimension mismatch
```

**Solution**: Verify the `dimension` parameter in `WithRedisStore` matches your model:
```go
dimension := agent.GetEmbeddingDimension()
fmt.Printf("Model dimension: %d\n", dimension)
```

#### Index Already Exists

Redis reuses existing indexes. To create a fresh index:
```bash
docker exec -it redis-vector-store redis-cli
FT.DROPINDEX my_knowledge_base DD
```

### Performance and Scalability

The Redis Vector Store uses the **HNSW algorithm** (Hierarchical Navigable Small World) for ultra-fast similarity search:

- ⚡ Constant time search O(log n)
- 📊 Support for millions of vectors
- 🎯 High precision with cosine similarity
- 🔄 Real-time updates

**Recommendations:**
- Use Redis for datasets > 10,000 documents
- Index in batches for better performance
- Configure Redis persistence (RDB or AOF) according to your needs

---

## 10. Chunking Utilities

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

### ChunkXML

Split XML content into chunks based on a specified target tag:

```go
xml := `<menu>
  <item id="1">
    <name>Margherita Pizza</name>
    <price currency="USD">12.99</price>
  </item>
  <item id="2">
    <name>Caesar Salad</name>
    <price currency="USD">8.50</price>
  </item>
</menu>`

chunks := chunks.ChunkXML(xml, "item")
for _, chunk := range chunks {
    agent.SaveEmbedding(chunk)
}
// Each chunk contains: <item id="1">...</item>, <item id="2">...</item>, etc.
```

**Features:**
- Extracts all elements matching the target tag name
- Preserves all XML attributes automatically
- Supports self-closing tags (`<item ... />`)
- Supports tags with content (`<item>...</item>`)
- Handles nested elements correctly

### ChunkYAML

Split YAML content into chunks based on a specified target key. The target key can be a simple key (e.g., `"snippet"`) or a list item key (e.g., `"- id"`).

**Example with list items (`- id`):**

```go
yaml := `snippets:
  - id: 1
    name: hello_world
    language: swiftlang
    code: |
      print("Hello, World!")
  - id: 2
    name: variables
    language: swiftlang
    code: |
      var name = "Alice"
      let pi = 3.14`

yamlChunks := chunks.ChunkYAML(yaml, "- id")
for _, chunk := range yamlChunks {
    agent.SaveEmbedding(chunk)
}
// Each chunk contains one complete list item with all its nested content
```

**Example with simple key (`snippet`):**

```go
yaml := `snippets:
  snippet:
    name: "Chunk YAML"
    language: "python"
    code: |
      print("hello")
  snippet:
    name: "Example"
    language: "python"
    code: |
      print("world")`

yamlChunks := chunks.ChunkYAML(yaml, "snippet")
// Each chunk contains one complete snippet block
```

**Features:**
- Extracts all blocks matching the target key at the same indentation level
- Supports simple keys (`snippet:`) and list item keys (`- id:`)
- Preserves the complete nested content of each block
- Automatically detects indentation level from the first match

---

## 11. Options: AgentOption and RagAgentOption

The RAG agent supports two distinct option types, both passed as variadic `...any` arguments to `NewAgent`:

### AgentOption (base-level)

`AgentOption` operates on the internal `*BaseAgent` and configures low-level behavior such as storage backend:

```go
// Store configuration options
rag.WithInMemoryStore()
rag.WithJsonStore(storeFilePath)
rag.WithRedisStore(stores.RedisConfig{...}, dimension)
rag.WithDocuments(documents, mode)
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
    // Use Redis as backend (optional)
    rag.WithRedisStore(stores.RedisConfig{
        Address:   "localhost:6379",
        Password:  "",
        DB:        0,
        IndexName: "my_index",
    }, 1024),
)
```

---

## 12. Lifecycle Hooks (RagAgentOption)

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

## 13. Context and State Management

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

## 14. JSON Export and Debugging

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

## 15. API Reference

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

// RedisConfig configures the Redis connection for the vector store
type RedisConfig struct {
    Address   string // Redis server address (e.g., "localhost:6379")
    Password  string // Redis password (empty string for no password)
    DB        int    // Redis database number (default: 0)
    IndexName string // Redis search index name (default: "nova_rag_index")
}

// DocumentLoadMode defines how documents should be loaded when store has existing data
type DocumentLoadMode string

const (
    DocumentLoadModeOverwrite  // Clear existing data and load new documents
    DocumentLoadModeMerge      // Add documents to existing data (default)
    DocumentLoadModeSkip       // Skip loading if store already has data
    DocumentLoadModeError      // Log error if store is not empty
)
```

---

### Option Functions

| Function | Type | Description |
|---|---|---|
| `BeforeCompletion(fn func(*Agent))` | `RagAgentOption` | Sets a hook called before each embedding generation in `GenerateEmbedding`. |
| `AfterCompletion(fn func(*Agent))` | `RagAgentOption` | Sets a hook called after each embedding generation in `GenerateEmbedding`. |
| `WithInMemoryStore()` | `AgentOption` | Configures the agent to use in-memory vector storage (default behavior). |
| `WithJsonStore(storePathFile string)` | `AgentOption` | Configures the agent to use JSON file-based storage. Automatically loads existing data from the file if it exists. |
| `WithRedisStore(config RedisConfig, dimension int)` | `AgentOption` | Configures Redis as the vector store backend. The `dimension` parameter must match the embedding model's dimension. |
| `WithDocuments(documents []string, mode ...DocumentLoadMode)` | `AgentOption` | Initializes the agent with predefined documents. Optional mode parameter controls behavior when store has existing data (default: `DocumentLoadModeMerge`). Must be applied AFTER store configuration options. |

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
| `PersistStore(path string) error` | Save the vector store to a JSON file. Note: Automatically called when using `WithJsonStore` + `WithDocuments` if new documents are added. |
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
