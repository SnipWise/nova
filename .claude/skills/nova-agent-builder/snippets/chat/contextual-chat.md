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

## Conversation History Management

**IMPORTANT**: The agent **automatically maintains** conversation history internally. You do NOT need to manually track it.

### Retrieve History

```go
// Get all conversation messages
history := agent.GetMessages()
for i, msg := range history {
    fmt.Printf("[%d] %s: %s\n", i, msg.Role, msg.Content)
}
```

### Reset History

```go
// Clear all messages except system instructions
agent.ResetMessages()
```

### Check Context Size

```go
// Get approximate context size in characters
contextSize := agent.GetContextSize()
fmt.Printf("Current context: %d characters\n", contextSize)
```

### Export to JSON

```go
// Export conversation history to JSON
jsonData, err := agent.ExportMessagesToJSON()
if err != nil {
    log.Fatal(err)
}
fmt.Println(jsonData)

// Save to file
os.WriteFile("conversation.json", []byte(jsonData), 0644)
```

## Important Notes

- **Automatic History Management**: The agent maintains conversation history automatically - DO NOT create manual `conversationHistory` arrays
- **Simple Usage**: Always pass only the current user message: `[]messages.Message{{Role: roles.User, Content: input}}`
- **History Methods**: Use `GetMessages()`, `ResetMessages()`, `GetContextSize()`, `ExportMessagesToJSON()`
- History grows with each exchange - monitor with `GetContextSize()`
- For long conversations, consider using a Compressor Agent (see `compressor/compressor-agent.md`)

## What NOT to Do

```go
// ❌ WRONG - Don't manually maintain history
var history []messages.Message
history = append(history, messages.Message{Role: roles.User, Content: input})
result, _ := agent.GenerateCompletion(history)
history = append(history, messages.Message{Role: roles.Assistant, Content: result.Response})
```

## What to Do Instead

```go
// ✅ CORRECT - Just pass the current message
result, _ := agent.GenerateCompletion(
    []messages.Message{{Role: roles.User, Content: input}},
)
// History is automatically updated by the agent
```

## Customization
- Each request sends full history (token cost)
- Use `clear` to reset if context becomes confusing
- Consider compression for long conversations (see compressor-agent)
