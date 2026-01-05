# RAG Agent

## Description

The **RAG Agent** (Retrieval-Augmented Generation) is a specialized agent for vector embedding generation and similarity search. It allows storing textual content as vectors and searching for the most similar content to a given query.

## Features

- **Embedding generation** : Converts text to numerical vectors
- **Vector storage** : Saves embeddings in memory
- **Similarity search** : Finds the most similar content via cosine similarity
- **Persistence** : Saves and loads vector store from JSON file
- **Top-N Search** : Retrieves the N best similar results

## Use cases

The RAG Agent is used to:
- **Enrich context** of chat agents with relevant information
- **Create a knowledge base** queryable by semantic similarity
- **Semantic search** in documents, FAQs, documentation
- **Recommend** similar content

## Creating a RAG Agent

### Basic syntax

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

// Agent configuration
agentConfig := agents.Config{
    Name: "RAG",
}

// Embedding model configuration
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    Name:      "mxbai-embed-large", // Embedding model
}

// Create the agent
agent, err := rag.NewAgent(ctx, agentConfig, modelConfig)
if err != nil {
    log.Fatal(err)
}
```

## VectorRecord structure

Search results return `VectorRecord` objects:

```go
type VectorRecord struct {
    ID         string         // Unique record identifier
    Prompt     string         // Original textual content
    Embedding  []float64      // Embedding vector
    Metadata   map[string]any // Optional metadata
    Similarity float64        // Cosine similarity score (0.0 - 1.0)
}
```

## Main methods

### Embedding generation

```go
// Generate an embedding for text
embedding, err := agent.GenerateEmbedding("How to make a pizza?")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Embedding vector: %d dimensions\n", len(embedding))

// Get the embedding dimension of the model
dimension := agent.GetEmbeddingDimension()
fmt.Printf("Model dimension: %d\n", dimension) // e.g., 1024
```

### Saving embeddings

```go
// Save an embedding in the in-memory vector store
err := agent.SaveEmbedding("Neapolitan pizza is prepared with tipo 00 flour.")
if err != nil {
    log.Fatal(err)
}

// Alternative (same function)
err = agent.SaveEmbeddingIntoMemoryVectorStore("The dough must rise for 24 hours.")
```

### Similarity search

```go
// Search all similar content with a similarity threshold
results, err := agent.SearchSimilar("How to prepare pizza dough?", 0.6)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Similarity: %.2f - Content: %s\n", result.Similarity, result.Prompt)
}
```

**Parameters**:
- `content` : The search text
- `limit` : Minimum cosine similarity threshold (0.0 - 1.0)
  - 1.0 = exact match
  - 0.8-1.0 = very similar
  - 0.6-0.8 = similar
  - 0.0-0.6 = not very similar

### Top-N search

```go
// Search for the top 3 results with a threshold of 0.6
results, err := agent.SearchTopN("How to make dough rise?", 0.6, 3)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d results:\n", len(results))
for i, result := range results {
    fmt.Printf("%d. [%.2f] %s\n", i+1, result.Similarity, result.Prompt)
}
```

**Parameters**:
- `content` : The search text
- `limit` : Minimum similarity threshold (0.0 - 1.0)
- `n` : Maximum number of results to return

### Vector store persistence

```go
// Save the vector store to a JSON file
err := agent.PersistStore("./data/knowledge.json")
if err != nil {
    log.Fatal(err)
}

// Check if the file exists
exists := agent.StoreFileExists("./data/knowledge.json")
fmt.Printf("Store file exists: %v\n", exists)

// Load the vector store from a JSON file
err = agent.LoadStore("./data/knowledge.json")
if err != nil {
    log.Fatal(err)
}
```

### Getters and Setters

```go
// Configuration
config := agent.GetConfig()
agent.SetConfig(newConfig)

modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig) // Note: Requires recreating the agent

// Information
name := agent.GetName()
modelID := agent.GetModelID()
kind := agent.Kind() // Returns agents.Rag

// Context
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requests/Responses (debugging)
lastRequestJSON, _ := agent.GetLastRequestJSON()
lastResponseJSON, _ := agent.GetLastResponseJSON()
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
```

## Usage with other agents

The RAG Agent is typically used with Server or Crew agents to automatically enrich context:

```go
// Create the RAG agent
ragAgent, _ := rag.NewAgent(ctx, agentConfig, modelConfig)

// Populate the knowledge base
ragAgent.SaveEmbedding("Neapolitan pizza cooks at 450°C for 90 seconds.")
ragAgent.SaveEmbedding("Tipo 00 flour is ideal for pizza.")
ragAgent.SaveEmbedding("Buffalo mozzarella is traditionally used.")

// Use with Server Agent
serverAgent, _ := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithRagAgentAndSimilarityConfig(ragAgent, 0.6, 3),
)

// Use with Crew Agent
crewAgent, _ := crew.NewAgent(
    ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithRagAgentAndSimilarityConfig(ragAgent, 0.6, 3),
)

// During a request, the context is automatically enriched
// with the 3 most similar contents (threshold 0.6)
```

## Complete example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Configuration
    agentConfig := agents.Config{
        Name: "PizzaKnowledge",
    }
    modelConfig := models.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        Name:      "mxbai-embed-large",
    }

    // Create the RAG agent
    agent, err := rag.NewAgent(ctx, agentConfig, modelConfig)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Embedding dimension: %d\n", agent.GetEmbeddingDimension())

    // Populate the knowledge base
    knowledge := []string{
        "Neapolitan pizza cooks at 450°C for 90 seconds in a wood-fired oven.",
        "Tipo 00 flour is the best for Neapolitan pizza dough.",
        "Buffalo mozzarella campana DOP is traditionally used.",
        "The dough must rise for at least 8 hours, ideally 24-48 hours.",
        "The tomato sauce is made with San Marzano DOP tomatoes.",
        "Extra virgin olive oil is added after cooking.",
    }

    for _, content := range knowledge {
        if err := agent.SaveEmbedding(content); err != nil {
            log.Printf("Error saving: %v", err)
        }
    }

    // Save to a file
    if err := agent.PersistStore("./pizza-knowledge.json"); err != nil {
        log.Fatal(err)
    }
    fmt.Println("✅ Knowledge base saved")

    // Similarity search
    query := "What temperature to cook pizza?"
    results, err := agent.SearchTopN(query, 0.5, 2)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n🔍 Query: %s\n", query)
    fmt.Printf("Found %d results:\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. [Similarity: %.2f]\n   %s\n\n",
            i+1, result.Similarity, result.Prompt)
    }
}
```

**Expected output**:
```
Embedding dimension: 1024
✅ Knowledge base saved

🔍 Query: What temperature to cook pizza?
Found 2 results:
1. [Similarity: 0.87]
   Neapolitan pizza cooks at 450°C for 90 seconds in a wood-fired oven.

2. [Similarity: 0.62]
   The tomato sauce is made with San Marzano DOP tomatoes.
```

## Cosine similarity

The RAG Agent uses **cosine similarity** to compare vectors:

- **1.0** : Identical vectors (perfect match)
- **0.8-1.0** : Very similar
- **0.6-0.8** : Moderately similar
- **0.4-0.6** : Slightly similar
- **0.0-0.4** : Very little similarity
- **0.0** : No similarity

**Threshold recommendations**:
- `0.7-0.8` : For precise matches
- `0.6` : Good balance (recommended)
- `0.5` : For more results, less precise

## Notes

- **Kind** : Returns `agents.Rag`
- **Vector Store** : In-memory storage with JSON persistence
- **Dimension** : Depends on the model (e.g., `mxbai-embed-large` = 1024 dimensions)
- **Empty error** : Returns an error if `content` is empty
- **Top-N** : Returns at most N results, sorted by descending similarity
- **Persistence** : JSON format, can be shared between instances

## Recommendations

### Recommended embedding models

- **mxbai-embed-large** : 1024 dimensions, excellent quality/speed balance
- **nomic-embed-text** : 768 dimensions, fast and efficient
- **all-minilm** : 384 dimensions, very fast, less precise

### Best practices

1. **Chunking** : Split long documents into 200-500 word chunks
2. **Similarity threshold** : Start with 0.6, adjust according to your needs
3. **Top-N** : Limit to 3-5 results to avoid noise
4. **Persistence** : Save the vector store regularly
5. **Initial loading** : Check if the file exists before populating

```go
// Load or create
if agent.StoreFileExists("./knowledge.json") {
    agent.LoadStore("./knowledge.json")
} else {
    // Populate the base
    for _, content := range knowledge {
        agent.SaveEmbedding(content)
    }
    agent.PersistStore("./knowledge.json")
}
```
