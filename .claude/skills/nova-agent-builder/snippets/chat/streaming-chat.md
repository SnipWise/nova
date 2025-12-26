---
id: streaming-chat
name: Streaming Chat Agent
category: chat
complexity: beginner
sample_source: 05
description: Chat agent that generates responses in real-time using streaming
---

# Streaming Chat Agent

## Description

Creates a chat agent that streams responses token by token, providing a more interactive and responsive user experience.

## Use Cases

- Interactive chatbots
- Real-time assistants
- Interfaces requiring immediate feedback
- Long responses where waiting would be frustrating

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
			Name:               "streaming-assistant",                              // Agent name
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",      // LLM Engine URL
			SystemInstructions: "You are a helpful and friendly assistant.",        // System instructions
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",   // Model to use
			Temperature: models.Float64(0.7),     // Creativity (0.0 = deterministic, 1.0 = creative)
			MaxTokens:   models.Int(2000),        // Maximum response tokens
		},
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	fmt.Println("🤖 Streaming Chat Agent")
	fmt.Println("Type 'quit' to exit")
	fmt.Println(strings.Repeat("-", 40))

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
		if strings.ToLower(input) == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		fmt.Print("🤖 Assistant: ")

		// Streaming call with callback
		result, err := agent.GenerateStreamCompletion(
			[]messages.Message{
				{Role: roles.User, Content: input},
			},
			func(chunk string, finishReason string) error {
				// Called for each received chunk
				fmt.Print(chunk)
				return nil
			},
		)
		
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		fmt.Println() // New line after response

		// Optional: Display metadata
		if result.FinishReason != "" {
			fmt.Printf("   [finish_reason: %s]\n", result.FinishReason)
		}
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

### Streaming with Context History

```go
var conversationHistory []messages.Message

// Add user message
conversationHistory = append(conversationHistory, messages.Message{
    Role:    roles.User,
    Content: input,
})

// Stream with full history
result, err := agent.GenerateStreamCompletion(
    conversationHistory,
    func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        return nil
    },
)

// Add assistant response to history
conversationHistory = append(conversationHistory, messages.Message{
    Role:    roles.Assistant,
    Content: result.Response,
})
```

### Streaming with Progress Indicator

```go
tokenCount := 0
result, err := agent.GenerateStreamCompletion(
    msgs,
    func(chunk string, finishReason string) error {
        tokenCount++
        fmt.Print(chunk)
        // Show token count every 10 tokens
        if tokenCount%10 == 0 {
            fmt.Printf(" [%d]", tokenCount)
        }
        return nil
    },
)
```

## Important Notes

- Streaming requires a compatible model
- The callback is called for each token/chunk
- `finishReason` is empty until the end of generation
- Use `result.Response` to get the complete response after streaming
