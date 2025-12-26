---
id: basic-rag
name: Basic RAG Agent
category: rag
complexity: intermediate
sample_source: 13
description: RAG agent with in-memory vector store for semantic search
---

# Basic RAG Agent

## Description

Creates a RAG (Retrieval-Augmented Generation) agent that indexes documents, performs semantic search, and generates responses based on retrieved context.

## Use Cases

- FAQ systems
- Documentation search
- Knowledge bases
- Document Q&A
- Contextual assistants

## Complete Code

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	// === CONFIGURATION - CUSTOMIZE HERE ===
	agent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:               "rag-assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are an assistant that answers questions based on provided context. If the context doesn't contain the answer, say so.",
		},
		models.Config{
			Name: "ai/mxbai-embed-large", // Embedding model
		},
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	// === INDEX YOUR DOCUMENTS ===
	documents := []string{
		"Nova SDK is a Go framework for building AI agents.",
		"Nova SDK supports chat agents, RAG agents, and tools agents.",
		"To create an agent, use chat.NewAgent() with a configuration.",
		"The embeddings model converts text to vectors for semantic search.",
		"RAG stands for Retrieval-Augmented Generation.",
		"Nova SDK uses local LLMs via llama.cpp or Ollama.",
	}

	fmt.Println("📚 Indexing documents...")
	for i, doc := range documents {
		if err := agent.SaveEmbedding(doc); err != nil {
			fmt.Printf("Error indexing document %d: %v\n", i, err)
		}
	}
	fmt.Printf("✅ %d documents indexed\n", len(documents))

	fmt.Println("\n🤖 RAG Agent - Ask me about Nova SDK")
	fmt.Println("Type 'quit' to exit")
	fmt.Println(strings.Repeat("-", 40))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n❓ Question: ")
		if !scanner.Scan() {
			break
		}

		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}
		if strings.ToLower(query) == "quit" {
			break
		}

		// Semantic search
		threshold := 0.5 // Minimum similarity (0.0 to 1.0)
		similarities, err := agent.SearchSimilar(query, threshold)
		if err != nil {
			fmt.Printf("Search error: %v\n", err)
			continue
		}

		if len(similarities) == 0 {
			fmt.Println("🔍 No relevant documents found")
			continue
		}

		// Display results
		fmt.Println("\n📄 Retrieved context:")
		for i, sim := range similarities {
			fmt.Printf("   %d. [%.2f] %s\n", i+1, sim.Similarity, sim.Prompt)
		}

		// Generate response based on context
		// (requires separate chat agent for generation)
		fmt.Println("\n💡 Use this context to answer the question")
	}
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
EMBEDDING_MODEL: "ai/mxbai-embed-large"
SIMILARITY_THRESHOLD: 0.5
```

## Customization

### RAG with Response Generation

```go
import (
    "github.com/snipwise/nova/nova-sdk/agents/chat"
)

// Create a chat agent for generation
chatAgent, _ := chat.NewAgent(ctx,
    agents.Config{
        Name:               "rag-generator",
        EngineURL:          engineURL,
        SystemInstructions: "Answer based ONLY on the provided context.",
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.3), // Low for factual responses
    },
)

// Search and generate
similarities, _ := ragAgent.SearchSimilar(query, 0.5)

context := ""
for _, sim := range similarities {
    context += sim.Prompt + "\n"
}

prompt := fmt.Sprintf("Context:\n%s\n\nQuestion: %s", context, query)
result, _ := chatAgent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: prompt},
})
```

### With Document Chunking

```go
func chunkDocument(text string, chunkSize int, overlap int) []string {
    words := strings.Fields(text)
    var chunks []string
    
    for i := 0; i < len(words); i += chunkSize - overlap {
        end := i + chunkSize
        if end > len(words) {
            end = len(words)
        }
        chunk := strings.Join(words[i:end], " ")
        chunks = append(chunks, chunk)
        
        if end == len(words) {
            break
        }
    }
    
    return chunks
}

// Use
chunks := chunkDocument(longDocument, 100, 20)
for _, chunk := range chunks {
    agent.SaveEmbedding(chunk)
}
```

### With Metadata

```go
type DocumentWithMeta struct {
    Content  string
    Source   string
    Page     int
}

// Store with identifier
func indexWithMeta(agent *rag.Agent, doc DocumentWithMeta) {
    // Include metadata in indexed text
    indexed := fmt.Sprintf("[Source: %s, Page: %d] %s", 
        doc.Source, doc.Page, doc.Content)
    agent.SaveEmbedding(indexed)
}
```

## Important Notes

- The embedding model must be compatible (text-to-vector)
- Similarity threshold: 0.5 is a good starting point
- Chunk large documents for better retrieval
- RAG quality depends on chunk quality and size
- Consider persistence for production (database)
