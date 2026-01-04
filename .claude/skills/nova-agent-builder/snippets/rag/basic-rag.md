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

### With Document Chunking (Built-in Utilities)

```go
import "github.com/snipwise/nova/nova-sdk/agents/rag/chunks"

// METHOD 1: Split by markdown sections (RECOMMENDED for .md files)
// Preserves document structure and semantic boundaries
markdownContent := `
# Introduction
This is the intro...

## Features
- Feature 1
- Feature 2
`

// Split at section level (h2 by default)
sections := chunks.SplitMarkdownBySections(markdownContent)
for _, section := range sections {
    agent.SaveEmbedding(section)
}

// Split at specific heading level
sections := chunks.SplitMarkdownBySection(2, markdownContent) // h2 level
```

```go
// METHOD 2: Character-based chunking with overlap
// Good for plain text or when structure doesn't matter
longText := "Your long document content..."

textChunks := chunks.ChunkText(
    longText,
    512,  // chunk size in characters
    64,   // overlap size in characters
)

for _, chunk := range textChunks {
    agent.SaveEmbedding(chunk)
}
```

```go
// METHOD 3: Load and chunk files automatically
import "github.com/snipwise/nova/nova-sdk/toolbox/files"

// Get all markdown files from directory
contents, err := files.GetContentFiles("./data", ".md")
if err != nil {
    panic(err)
}

// Index with automatic chunking
for idx, content := range contents {
    // Split by sections for better semantic chunks
    piecesOfDoc := chunks.SplitMarkdownBySections(content)

    for chunkIdx, piece := range piecesOfDoc {
        fmt.Printf("Indexing doc %d/%d, chunk %d/%d\n",
            idx+1, len(contents), chunkIdx+1, len(piecesOfDoc))

        err := agent.SaveEmbedding(piece)
        if err != nil {
            fmt.Printf("Error indexing: %v\n", err)
        }
    }
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

### DO:
- Use `chunks.SplitMarkdownBySections()` for markdown documents (preserves structure)
- Use `chunks.ChunkText()` for plain text with consistent chunk sizes
- Use `files.GetContentFiles()` to load all files from a directory
- Set similarity threshold to 0.5-0.7 for balanced retrieval
- Choose embedding model carefully: `ai/mxbai-embed-large` (recommended)
- Chunk large documents before indexing (512-1024 chars per chunk)
- Use overlap (64-128 chars) to avoid losing context at boundaries
- Index chunks with metadata for better traceability

### DON'T:
- Don't index entire large documents without chunking
- Don't use overly small chunks (< 100 chars) - loses context
- Don't use overly large chunks (> 2000 chars) - loses precision
- Don't ignore chunk overlap - it preserves semantic continuity
- Don't mix different chunking strategies in same index
- Don't forget to test different chunk sizes for your use case

### Chunking Best Practices:
- **Markdown files**: Use `chunks.SplitMarkdownBySections()` - preserves document structure
- **Plain text**: Use `chunks.ChunkText(text, 512, 64)` - 512 chars, 64 overlap
- **Code files**: Use `chunks.SplitMarkdownBySections()` if well-commented
- **Long articles**: Combine both - split by sections, then chunk large sections

### Recommended Chunk Sizes:
- **Short FAQs**: 200-300 chars (one Q&A per chunk)
- **Documentation**: 512-1024 chars (one concept per chunk)
- **Articles/Blogs**: 1024-1536 chars (one paragraph or section)
- **Technical docs**: Use section-based splitting (adaptive size)
