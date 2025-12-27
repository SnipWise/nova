---
id: server-with-compressor
name: Server Agent with Compressor
category: server
complexity: intermediate
sample_source: 54
description: HTTP server agent with automatic context compression for long conversations
---

# Server Agent with Compressor

## Description

Creates an HTTP server agent with context compression capabilities. When conversations become too long, the compressor agent automatically summarizes the history to maintain context while staying within token limits.

## Use Cases

- Long-running conversation APIs
- Chat applications with extended sessions
- Token-limited models
- Cost-optimized AI services
- Memory-efficient servers

## Complete Code

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/models"
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
			Name:               "Bob",                                           // Agent name
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",  // LLM Engine URL
			SystemInstructions: "You are Bob, a helpful AI assistant.",         // System instructions
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

	// === COMPRESSOR AGENT CONFIGURATION ===
	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "compressor-agent",                             // Compressor name
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: compressor.Instructions.Minimalist,             // Use minimalist compression
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",    // Model for compression
			Temperature: models.Float64(0.0),                    // Deterministic compression
		},
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),  // Compression strategy
	)
	if err != nil {
		panic(err)
	}

	// === ATTACH COMPRESSOR AGENT ===
	serverAgent.SetCompressorAgent(compressorAgent)

	// === CONFIGURE COMPRESSION THRESHOLD ===
	serverAgent.SetContextSizeLimit(3000)  // Compress when context exceeds 3000 tokens

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
COMPRESSOR_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TEMPERATURE_CHAT: 0.4
TEMPERATURE_COMPRESSOR: 0.0
PORT: ":8080"
CONTEXT_SIZE_LIMIT: 3000
```

## How Compression Works

### Automatic Compression Workflow

1. **Normal Operation** (context < limit):
   - Conversation proceeds normally
   - All messages retained in memory

2. **Warning Threshold** (context > 80% of limit):
   - Server logs a warning
   - No compression yet

3. **Compression Threshold** (context > 90% of limit):
   - Compressor agent is triggered
   - Conversation history is summarized
   - Older messages are replaced with summary
   - Recent messages are preserved

4. **After Compression**:
   - Context size is reduced
   - Conversation continues with compressed history

## Compression Strategies

### Minimalist Compression (Default)

```go
compressor.NewAgent(
	ctx,
	agents.Config{
		SystemInstructions: compressor.Instructions.Minimalist,
	},
	modelConfig,
	compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
)
```

**Effect**: Very concise summaries, maximum token reduction

### Detailed Compression

```go
compressor.NewAgent(
	ctx,
	agents.Config{
		SystemInstructions: compressor.Instructions.Detailed,
	},
	modelConfig,
	compressor.WithCompressionPrompt(compressor.Prompts.Detailed),
)
```

**Effect**: More context preserved, less aggressive compression

### Custom Compression

```go
customInstructions := `You are a compression agent. Summarize the conversation
while preserving key information about: names, dates, decisions, and action items.`

customPrompt := `Summarize this conversation history, keeping all important details:`

compressor.NewAgent(
	ctx,
	agents.Config{
		SystemInstructions: customInstructions,
	},
	modelConfig,
	compressor.WithCompressionPrompt(customPrompt),
)
```

## API Usage

### Normal Conversation (No Compression Yet)

```bash
# First message
curl -N -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "Hello!"}}'

# Check token count
curl http://localhost:8080/memory/messages/tokens
# Response: {"tokens": 150}
```

### Long Conversation (Triggers Compression)

```bash
# Many messages later...
curl -N -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "What did we discuss earlier?"}}'

# Check token count after compression
curl http://localhost:8080/memory/messages/tokens
# Response: {"tokens": 1500}  (reduced from 3100)
```

## Customization

### Adjust Compression Threshold

```go
// More aggressive compression (compress sooner)
serverAgent.SetContextSizeLimit(2000)

// Less aggressive compression (compress later)
serverAgent.SetContextSizeLimit(6000)

// For small models (2K context)
serverAgent.SetContextSizeLimit(1500)

// For large models (8K context)
serverAgent.SetContextSizeLimit(7000)
```

### Manual Compression

While the server agent compresses automatically, you can also compress manually:

```go
// In a custom endpoint handler
func handleManualCompress(w http.ResponseWriter, r *http.Request) {
	// Get current messages
	messages := serverAgent.GetMessages()

	// Compress
	compressed, err := compressorAgent.Compress(messages)

	// Reset and add compressed version
	serverAgent.ResetMessages()
	serverAgent.AddMessage(roles.System, compressed)
}
```

### Compression-Aware Context Size

```go
// Monitor context size
currentSize := serverAgent.GetContextSize()
limit := 3000

if currentSize > limit {
	fmt.Printf("⚠️  Context size (%d) exceeds limit (%d)\n", currentSize, limit)
}
```

### Different Models for Compression

```go
// Use a smaller, faster model for compression
models.Config{
	Name:        "ai/qwen2.5:1.5B-F16",  // Faster model
	Temperature: models.Float64(0.0),
}

// Use same model for consistency
models.Config{
	Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
	Temperature: models.Float64(0.0),
}
```

## Monitoring

### Check Current Token Count

```bash
curl http://localhost:8080/memory/messages/tokens
```

### View Conversation History

```bash
curl http://localhost:8080/memory/messages/list
```

After compression, you'll see compressed summaries in the history.

### Reset Conversation

```bash
curl -X POST http://localhost:8080/memory/reset
```

## Compression vs. Context Window

| Model Context | Recommended Limit | Warning (80%) | Compress (90%) |
|---------------|-------------------|---------------|----------------|
| 2K tokens     | 1500             | 1200          | 1350           |
| 4K tokens     | 3000             | 2400          | 2700           |
| 8K tokens     | 7000             | 5600          | 6300           |
| 16K tokens    | 14000            | 11200         | 12600          |

## Important Notes

- Compression is **automatic** when context exceeds 90% of limit
- Warning appears at 80% of limit
- Compression is **deterministic** (temperature 0.0)
- Recent messages are **preserved**, older ones are summarized
- Compression reduces token count but may lose some context details
- Use minimalist compression for maximum token reduction
- Use detailed compression to preserve more information
- System instructions are never compressed

## Best Practices

1. **Set appropriate limits**: Based on your model's context window
2. **Use deterministic compression**: Temperature 0.0 for consistency
3. **Monitor token counts**: Use `/memory/messages/tokens` endpoint
4. **Choose compression strategy**: Minimalist vs. Detailed based on use case
5. **Test compression quality**: Verify important context is preserved

## Error Handling

```go
compressorAgent, err := compressor.NewAgent(ctx, config, models)
if err != nil {
	// Fallback: Continue without compression
	log.Printf("Warning: Compressor not available: %v", err)
	serverAgent.SetContextSizeLimit(99999) // Disable auto-compression
}
```

## Performance Considerations

- **Compression overhead**: Adds latency when triggered
- **Frequency**: Only happens when threshold is exceeded
- **Model choice**: Faster models reduce compression latency
- **Strategy choice**: Minimalist is faster than Detailed

## Related Patterns

- For basic server: See `basic-server.md`
- For RAG support: See `server-with-rag.md`
- For tools support: See `server-with-tools.md`
- For full-featured: See `server-full-featured.md`
