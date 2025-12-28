---
id: server-with-rag
name: Server Agent with RAG
category: server
complexity: intermediate
sample_source: 54
description: HTTP server agent with Retrieval-Augmented Generation for context-aware responses
---

# Server Agent with RAG

## Description

Creates an HTTP server agent with RAG (Retrieval-Augmented Generation) capabilities. The agent can search through a vector database of documents to provide context-aware responses based on relevant information.

## Use Cases

- Document Q&A APIs
- Knowledge base servers
- Context-aware chatbots
- Information retrieval systems
- FAQ automation

## Complete Code

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	// Enable logging
	if err := os.Setenv("NOVA_LOG_LEVEL", "INFO"); err != nil {
		panic(err)
	}

	ctx := context.Background()

	// === SERVER AGENT CONFIGURATION ===
	serverAgent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:                    "Bob",                                           // Agent name
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",  // LLM Engine URL
			SystemInstructions:      "You are Bob, a helpful AI assistant.",         // System instructions
			KeepConversationHistory: true,                                           // Keep conversation context
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",    // Model for chat
			Temperature: models.Float64(0.4),
		},
		":8080",  // HTTP port
		// executeFunction is optional - omitted here
	)
	if err != nil {
		panic(err)
	}

	// === RAG AGENT CONFIGURATION ===
	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name: "ai/mxbai-embed-large",  // Embeddings model
		},
	)
	if err != nil {
		panic(err)
	}

	// === INDEX DOCUMENTS ===
	// Load markdown files from data directory
	contents, err := files.GetContentFiles("./data", ".md")
	if err != nil {
		panic(err)
	}

	// Process and index each document
	for idx, content := range contents {
		// Split document into semantic chunks
		piecesOfDoc := chunks.SplitMarkdownBySections(content)

		for chunkIdx, piece := range piecesOfDoc {
			display.Colorf(display.ColorYellow,
				"generating vectors... (docs %d/%d) [chunks: %d/%d]\n",
				idx+1, len(contents), chunkIdx+1, len(piecesOfDoc))

			// Generate and save embedding
			err := ragAgent.SaveEmbedding(piece)
			if err != nil {
				display.Errorf("failed to save embedding for document %d: %v\n", idx, err)
			}
		}
	}

	// === ATTACH RAG AGENT ===
	serverAgent.SetRagAgent(ragAgent)

	// Optional: Configure RAG behavior
	serverAgent.SetSimilarityLimit(0.6)   // Minimum similarity threshold
	serverAgent.SetMaxSimilarities(3)     // Max number of documents to retrieve

	display.Colorf(display.ColorCyan, "🚀 Server starting on http://localhost%s\n", serverAgent.GetPort())

	// Start the server
	if err := serverAgent.StartServer(); err != nil {
		panic(err)
	}
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
CHAT_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
EMBEDDING_MODEL: "ai/mxbai-embed-large"
TEMPERATURE: 0.4
PORT: ":8080"
SIMILARITY_THRESHOLD: 0.6
MAX_DOCUMENTS: 3
DATA_DIR: "./data"
```

## Directory Structure

```
project/
├── main.go
└── data/
    ├── document1.md
    ├── document2.md
    └── faq.md
```

## How RAG Works

1. **Indexing Phase** (startup):
   - Documents are loaded from `./data` directory
   - Each document is split into semantic chunks
   - Embeddings are generated for each chunk
   - Embeddings are stored in an in-memory vector database

2. **Query Phase** (runtime):
   - User sends a question via HTTP
   - Question is converted to an embedding
   - Similar documents are retrieved (similarity search)
   - Retrieved documents are injected into the context
   - LLM generates response with relevant context

## API Usage

### Query with RAG Context

```bash
curl -N -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "message": "What are the features of Nova SDK?"
    }
  }'
```

The server will:
1. Search for relevant documents about "Nova SDK features"
2. Inject top matching documents into context
3. Generate response based on retrieved information
4. Stream response via SSE

## Customization

### Custom Document Chunking

```go
// Custom chunking strategy
func customChunker(content string) []string {
	// Split by paragraphs
	return strings.Split(content, "\n\n")
}

for _, content := range contents {
	chunks := customChunker(content)
	for _, chunk := range chunks {
		ragAgent.SaveEmbedding(chunk)
	}
}
```

### Different File Types

```go
// Load different file types
mdFiles, _ := files.GetContentFiles("./data", ".md")
txtFiles, _ := files.GetContentFiles("./docs", ".txt")

allContent := append(mdFiles, txtFiles...)

for _, content := range allContent {
	// Process and index
}
```

### Persistent Vector Store

```go
// Save embeddings to JSON file
ragAgent.PersistStore("./store/embeddings.json")

// Load on next startup
if ragAgent.StoreFileExists("./store/embeddings.json") {
	ragAgent.LoadStore("./store/embeddings.json")
} else {
	// Index documents
	// ...
	ragAgent.PersistStore("./store/embeddings.json")
}
```

### Adjust RAG Parameters

```go
// More strict similarity (higher threshold)
serverAgent.SetSimilarityLimit(0.8)  // Only very similar docs

// Retrieve more context
serverAgent.SetMaxSimilarities(5)    // Up to 5 documents

// Less strict similarity (lower threshold)
serverAgent.SetSimilarityLimit(0.4)  // More permissive matching
```

### Manual Document Addition

```go
// Add single documents programmatically
ragAgent.SaveEmbedding("Nova SDK is a Go framework for AI agents")
ragAgent.SaveEmbedding("RAG stands for Retrieval-Augmented Generation")
```

## Document Chunking Strategies

### By Sections (Markdown)

```go
chunks := chunks.SplitMarkdownBySections(content)
```

### By Paragraphs

```go
paragraphs := strings.Split(content, "\n\n")
```

### By Size

```go
func chunkBySize(text string, size int) []string {
	var chunks []string
	for i := 0; i < len(text); i += size {
		end := i + size
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}
```

### By Sentences

```go
sentences := strings.Split(content, ". ")
```

## Monitoring RAG Performance

```bash
# Check token count (includes retrieved documents)
curl http://localhost:8080/memory/messages/tokens

# View conversation with injected context
curl http://localhost:8080/memory/messages/list
```

## Important Notes

- RAG searches for similar documents on **every request**
- Documents are injected as **system messages** before generation
- Similarity threshold: `0.6` is a good default (range: 0.0 - 1.0)
- Embeddings model must support the documents' language
- In-memory vector store is lost on server restart (use `PersistStore()` for persistence)
- Chunking strategy affects retrieval quality
- Smaller chunks = more precise but less context
- Larger chunks = more context but less precise

## Error Handling

```go
// Handle indexing errors
for idx, content := range contents {
	chunks := chunks.SplitMarkdownBySections(content)
	for _, chunk := range chunks {
		if err := ragAgent.SaveEmbedding(chunk); err != nil {
			log.Printf("Warning: Failed to index chunk: %v", err)
			// Continue with other chunks
		}
	}
}
```

## Related Patterns

- For basic server: See `basic-server.md`
- For tools support: See `server-with-tools.md`
- For context compression: See `server-with-compressor.md`
- For full-featured: See `server-full-featured.md`
