---
id: contextual-chat
name: Contextual Chat Agent
category: chat
complexity: beginner
sample_source: 06
description: Chat agent that maintains conversation history across multiple turns
---

# Contextual Chat Agent

## Description

Creates a chat agent that maintains conversation context, allowing coherent multi-turn conversations where the agent remembers previous exchanges.

## Use Cases

- Multi-turn conversations
- Assistants requiring memory
- Customer support chatbots
- Tutoring systems

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
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	// === CONFIGURATION - CUSTOMIZE HERE ===
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "contextual-assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful assistant. Remember our conversation context to provide relevant responses.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.7),
			MaxTokens:   models.Int(2000),
		},
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	fmt.Println("🤖 Contextual Chat Agent")
	fmt.Println("Type 'quit' to exit, 'clear' to reset history")
	fmt.Println(strings.Repeat("-", 40))

	// Conversation history
	var history []messages.Message
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n👤 You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "quit":
			fmt.Println("Goodbye!")
			return
		case "clear":
			history = nil
			fmt.Println("📝 History cleared")
			continue
		}

		// Add user message to history
		history = append(history, messages.Message{
			Role:    roles.User,
			Content: input,
		})

		// Generate response with full context
		result, err := agent.GenerateCompletion(history)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			// Remove failed message from history
			history = history[:len(history)-1]
			continue
		}

		// Add assistant response to history
		history = append(history, messages.Message{
			Role:    roles.Assistant,
			Content: result.Response,
		})

		fmt.Printf("🤖 Assistant: %s\n", result.Response)
		fmt.Printf("   [history: %d messages]\n", len(history))
	}
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "ai/qwen2.5:1.5B-F16"
TEMPERATURE: 0.7
MAX_TOKENS: 2000
```

## Customization

### With Maximum History Limit

```go
const MaxHistorySize = 20 // Keep last 20 messages

func trimHistory(history []messages.Message, maxSize int) []messages.Message {
    if len(history) <= maxSize {
        return history
    }
    // Keep recent messages
    return history[len(history)-maxSize:]
}

// Before generating
history = trimHistory(history, MaxHistorySize)
```

### With History Persistence

```go
import "encoding/json"

func saveHistory(history []messages.Message, filename string) error {
    data, err := json.Marshal(history)
    if err != nil {
        return err
    }
    return os.WriteFile(filename, data, 0644)
}

func loadHistory(filename string) ([]messages.Message, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    var history []messages.Message
    err = json.Unmarshal(data, &history)
    return history, err
}
```

### With Conversation Summary

```go
func summarizeIfNeeded(agent *chat.Agent, history []messages.Message, threshold int) []messages.Message {
    if len(history) < threshold {
        return history
    }
    
    // Create summary of old messages
    oldMessages := history[:len(history)-4] // Keep last 4 intact
    summaryPrompt := "Summarize this conversation in 2-3 sentences:\n"
    for _, msg := range oldMessages {
        summaryPrompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
    }
    
    result, _ := agent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: summaryPrompt},
    })
    
    // Return summary + recent messages
    return append(
        []messages.Message{{Role: roles.System, Content: "Previous context: " + result.Response}},
        history[len(history)-4:]...,
    )
}
```

## Important Notes

- History grows with each exchange - consider limiting
- Each request sends full history (token cost)
- Use `clear` to reset if context becomes confusing
- Consider compression for long conversations (see compressor-agent)
