---
id: compressor-agent
name: Context Compressor Agent
category: compressor
complexity: advanced
sample_source: 28
description: Agent that compresses conversation context to save tokens using compressor.NewAgent
---

# Context Compressor Agent

## Description

Creates a Nova agent that automatically compresses conversation context when it becomes too long. Uses `compressor.NewAgent` which provides the `Compress()` method to summarize old messages while preserving essential information.

## Use Cases

- Long conversations (customer support, coaching)
- Assistants with extended memory
- API cost reduction
- Maintaining context across sessions
- Agents with strict token limits

## Complete Code

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// === CONFIGURATION ===
const (
	MaxContextSize    = 4000 // Trigger compression above this
	TargetContextSize = 2000 // Target size after compression
	KeepRecentCount   = 4    // Keep last N messages uncompressed
)

func main() {
	ctx := context.Background()

	// === MAIN CHAT AGENT ===
	mainAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful AI assistant with excellent memory.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.7),
		},
	)
	if err != nil {
		panic(err)
	}

	// === COMPRESSOR AGENT ===
	// Best practice: Use predefined Instructions and Prompts from compressor package
	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:      "compressor",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			// Use built-in instructions for optimal compression
			// Options: compressor.Instructions.Expert, .Effective, .Basic
			SystemInstructions: compressor.Instructions.Effective,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0), // Deterministic for summaries
		},
		// Use built-in compression prompts for consistent results
		// Options: compressor.Prompts.UltraShort, .Minimalist, .Balanced, .Detailed
		compressor.WithCompressionPrompt(compressor.Prompts.UltraShort),
	)
	if err != nil {
		panic(err)
	}

	// === CONTEXT MANAGER ===
	cm := &ContextManager{
		mainAgent:         mainAgent,
		compressorAgent:   compressorAgent,
		maxSize:           MaxContextSize,
		targetSize:        TargetContextSize,
		history:           []messages.Message{},
		compressedSummary: "",
	}

	// === SIMULATE LONG CONVERSATION ===
	exchanges := []string{
		"Hello! My name is Marie and I'm looking for advice on learning Go.",
		"I have 5 years of Python experience and want to diversify.",
		"What are the key concepts to master first?",
		"How does concurrency management work in Go?",
		"Can you give me a goroutine example?",
		"And channels, how do they work?",
		"What's the difference between buffered and unbuffered channels?",
		"How to handle errors properly in Go?",
		"Tell me about interfaces in Go.",
		"How to structure a professional Go project?",
		"What testing tools do you recommend?",
		"How to do TDD in Go?",
		"Let's talk about design patterns in Go.",
	}

	fmt.Println("=== Conversation with Context Compression ===\n")

	for i, userMessage := range exchanges {
		fmt.Printf("--- Exchange %d ---\n", i+1)
		fmt.Printf("👤 User: %s\n", userMessage)

		response, err := cm.Chat(userMessage)
		if err != nil {
			panic(err)
		}

		// Display truncated response
		displayResponse := response
		if len(displayResponse) > 200 {
			displayResponse = displayResponse[:200] + "..."
		}
		fmt.Printf("🤖 Assistant: %s\n", displayResponse)
		fmt.Printf("📊 Context: %d characters\n\n", cm.GetContextSize())
	}

	// Display final summary
	if cm.compressedSummary != "" {
		fmt.Println("=== Compressed Summary ===")
		fmt.Println(cm.compressedSummary)
	}
}

// === CONTEXT MANAGER ===
type ContextManager struct {
	mainAgent         *chat.Agent
	compressorAgent   *compressor.Agent
	maxSize           int
	targetSize        int
	history           []messages.Message
	compressedSummary string
}

func (cm *ContextManager) GetContextSize() int {
	size := len(cm.compressedSummary)
	for _, msg := range cm.history {
		size += len(msg.Content)
	}
	return size
}

func (cm *ContextManager) Chat(userMessage string) (string, error) {
	// Add user message
	cm.history = append(cm.history, messages.Message{
		Role:    roles.User,
		Content: userMessage,
	})

	// Check if compression needed
	if cm.GetContextSize() > cm.maxSize {
		fmt.Println("⚡ Compressing context...")
		if err := cm.compress(); err != nil {
			return "", err
		}
		fmt.Printf("⚡ Context compressed: %d characters\n", cm.GetContextSize())
	}

	// Build complete context
	contextMessages := cm.buildContextMessages()

	// Generate response
	result, err := cm.mainAgent.GenerateCompletion(contextMessages)
	if err != nil {
		return "", err
	}

	// Add response to history
	cm.history = append(cm.history, messages.Message{
		Role:    roles.Assistant,
		Content: result.Response,
	})

	return result.Response, nil
}

func (cm *ContextManager) buildContextMessages() []messages.Message {
	var msgs []messages.Message

	// Add compressed summary if present
	if cm.compressedSummary != "" {
		msgs = append(msgs, messages.Message{
			Role:    roles.System,
			Content: fmt.Sprintf("Previous context (summary):\n%s", cm.compressedSummary),
		})
	}

	// Add recent history
	msgs = append(msgs, cm.history...)

	return msgs
}

func (cm *ContextManager) compress() error {
	// Prepare text to compress
	var toCompress strings.Builder

	if cm.compressedSummary != "" {
		toCompress.WriteString("Previous summary:\n")
		toCompress.WriteString(cm.compressedSummary)
		toCompress.WriteString("\n\nNew conversation:\n")
	}

	// Keep recent messages (don't compress)
	keepRecent := KeepRecentCount
	if len(cm.history) <= keepRecent {
		return nil // Not enough messages to compress
	}

	// Messages to compress
	toCompressMessages := cm.history[:len(cm.history)-keepRecent]
	recentMessages := cm.history[len(cm.history)-keepRecent:]

	for _, msg := range toCompressMessages {
		role := "User"
		if msg.Role == roles.Assistant {
			role = "Assistant"
		}
		toCompress.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	// Compress using the compressor agent's Compress method
	summary, err := cm.compressorAgent.Compress(toCompress.String())
	if err != nil {
		return err
	}

	// Update state
	cm.compressedSummary = summary
	cm.history = recentMessages

	return nil
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL_NAME: "ai/qwen2.5:1.5B-F16"

# Compression settings
MAX_CONTEXT_SIZE: 4000      # Trigger threshold
TARGET_CONTEXT_SIZE: 2000   # Target after compression
KEEP_RECENT_MESSAGES: 4     # Recent messages to preserve
```

## Key API

### compressor.NewAgent

```go
import "github.com/snipwise/nova/nova-sdk/agents/compressor"

// RECOMMENDED: Use built-in instructions and prompts
agent, err := compressor.NewAgent(
    ctx,
    agents.Config{
        Name:               "compressor",
        EngineURL:          engineURL,
        // Use predefined instructions for best results
        SystemInstructions: compressor.Instructions.Effective,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
    },
    // Use built-in compression prompts
    compressor.WithCompressionPrompt(compressor.Prompts.UltraShort),
)

// Available Instructions:
// - compressor.Instructions.Expert     - Most detailed compression
// - compressor.Instructions.Effective  - Balanced (RECOMMENDED)
// - compressor.Instructions.Basic      - Simple compression

// Available Prompts:
// - compressor.Prompts.UltraShort   - Maximum compression (RECOMMENDED)
// - compressor.Prompts.Minimalist   - Very concise
// - compressor.Prompts.Balanced     - Balance detail/brevity
// - compressor.Prompts.Detailed     - Preserve more information
```

### Compress Method

```go
// Compress text and return summary
summary, err := compressorAgent.Compress(textToCompress)
```

## Customization

### With Compression Statistics

```go
type CompressionStats struct {
    OriginalSize     int
    CompressedSize   int
    CompressionRatio float64
    Timestamp        time.Time
}

func (cm *ContextManager) CompressWithStats() (*CompressionStats, error) {
    originalSize := cm.GetContextSize()
    
    err := cm.compress()
    if err != nil {
        return nil, err
    }
    
    compressedSize := cm.GetContextSize()
    
    return &CompressionStats{
        OriginalSize:     originalSize,
        CompressedSize:   compressedSize,
        CompressionRatio: float64(compressedSize) / float64(originalSize),
        Timestamp:        time.Now(),
    }, nil
}
```

### With Session Persistence

```go
import "encoding/json"

func (cm *ContextManager) SaveState(filepath string) error {
    state := map[string]interface{}{
        "summary":   cm.compressedSummary,
        "history":   cm.history,
        "timestamp": time.Now(),
    }
    data, _ := json.Marshal(state)
    return os.WriteFile(filepath, data, 0644)
}

func (cm *ContextManager) LoadState(filepath string) error {
    data, err := os.ReadFile(filepath)
    if err != nil {
        return err
    }
    var state struct {
        Summary string             `json:"summary"`
        History []messages.Message `json:"history"`
    }
    if err := json.Unmarshal(data, &state); err != nil {
        return err
    }
    cm.compressedSummary = state.Summary
    cm.history = state.History
    return nil
}
```

## Important Notes

- Use `compressor.NewAgent` (not `chat.NewAgent`) for the compressor
- **BEST PRACTICE**: Use `compressor.Instructions.Effective` for SystemInstructions
- **BEST PRACTICE**: Use `compressor.Prompts.UltraShort` for maximum compression
- The `Compress()` method takes a string and returns a compressed summary
- The `CompressContextStream()` method compresses messages with streaming support
- Always keep recent messages uncompressed for conversation flow
- Temperature 0.0 for compressor ensures consistent summaries
- Save summaries for session persistence
- Quality of summary depends on the model used

## Built-in Presets

### Instructions Presets (SystemInstructions)

```go
compressor.Instructions.Expert     // Most sophisticated compression
compressor.Instructions.Effective  // Balanced approach (RECOMMENDED)
compressor.Instructions.Basic      // Simple, straightforward compression
```

### Compression Prompts (WithCompressionPrompt)

```go
compressor.Prompts.UltraShort   // Maximum token reduction (RECOMMENDED)
compressor.Prompts.Minimalist   // Very concise summaries
compressor.Prompts.Balanced     // Balance between detail and brevity
compressor.Prompts.Detailed     // Preserve more contextual information
```
