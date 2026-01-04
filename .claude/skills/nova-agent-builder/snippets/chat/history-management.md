---
id: history-management
name: Conversation History Management
category: chat
complexity: intermediate
sample_source: 60-61
description: Control conversation history with KeepConversationHistory flag for stateful/stateless agents
---

# Conversation History Management

## Description

Creates Nova chat agents with explicit control over conversation history. The `KeepConversationHistory` flag determines whether agents maintain context across requests (stateful) or treat each request independently (stateless).

## Use Cases

- **Stateful (History=true)**: Multi-turn conversations, customer support, personal assistants
- **Stateless (History=false)**: Independent requests, classification tasks, stateless APIs, batch processing
- API endpoints where context isn't needed
- Memory-efficient agents
- Intent detection / routing agents

## Complete Code

### Stateless Agent (No History)

```go
package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// === STATELESS AGENT ===
	// Each request is independent - no context preservation
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: false, // CRITICAL: Disables history
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
			MaxTokens:   models.Int(2000),
		},
	)
	if err != nil {
		panic(err)
	}

	// === REQUEST 1 ===
	display.Title("Request 1: What is your name?")

	result1, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Hello, what is your name?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result1.Response)
	display.KeyValue("Finish reason", result1.FinishReason)

	// Check message history
	messages1 := agent.GetMessages()
	display.KeyValue("Messages count", conversion.IntToString(len(messages1)))
	display.Info("Expected: 1 (only system message)")

	// === REQUEST 2 ===
	display.NewLine()
	display.Title("Request 2: Who is James T Kirk?")

	result2, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Who is James T Kirk?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result2.Response)

	// === REQUEST 3 - NO CONTEXT ===
	display.NewLine()
	display.Title("Request 3: Who is his best friend? (NO CONTEXT)")

	result3, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Who is his best friend?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result3.Response)
	display.Info("Note: Agent doesn't know who 'his' refers to")

	// Verify history is NOT kept
	messages3 := agent.GetMessages()
	display.KeyValue("Final messages count", conversion.IntToString(len(messages3)))
	display.Info("Expected: 1 (still only system message)")

	// Export messages to JSON
	fmt.Println(agent.ExportMessagesToJSON())
}
```

### Stateful Agent (With History)

```go
package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// === STATEFUL AGENT ===
	// Maintains context across requests
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: true, // CRITICAL: Enables history
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.7),
		},
	)
	if err != nil {
		panic(err)
	}

	// === REQUEST 1 ===
	result1, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "My name is Alice and I love Go programming."},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Assistant:", result1.Response)

	// === REQUEST 2 - USES CONTEXT ===
	result2, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "What is my name?"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Assistant:", result2.Response)
	display.Info("Agent remembers: 'Your name is Alice'")

	// === REQUEST 3 - USES FULL CONTEXT ===
	result3, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "What programming language do I like?"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Assistant:", result3.Response)
	display.Info("Agent remembers: 'You like Go programming'")

	// Check accumulated history
	allMessages := agent.GetMessages()
	fmt.Printf("\nTotal messages in history: %d\n", len(allMessages))
	// Expected: 1 system + 3 user + 3 assistant = 7 messages
}
```

## Configuration

```yaml
# Stateless Configuration
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "ai/qwen2.5:1.5B-F16"
TEMPERATURE: 0.0
KEEP_CONVERSATION_HISTORY: false  # No memory

# Stateful Configuration
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "ai/qwen2.5:1.5B-F16"
TEMPERATURE: 0.7
KEEP_CONVERSATION_HISTORY: true   # Full memory
```

## Key API

### KeepConversationHistory Flag

```go
import "github.com/snipwise/nova/nova-sdk/agents"

// Stateless agent (default behavior in some scenarios)
agents.Config{
    KeepConversationHistory: false, // Each request is independent
}

// Stateful agent (recommended for chat)
agents.Config{
    KeepConversationHistory: true, // Maintains full context
}
```

### Message Management Methods

```go
// Get all messages in history
messages := agent.GetMessages()

// Get message count / context size
contextSize := agent.GetContextSize()

// Add message manually
agent.AddMessage(roles.User, "Hello")

// Reset conversation history
agent.ResetMessages()

// Export messages to JSON
jsonString := agent.ExportMessagesToJSON()

// Import messages from JSON
err := agent.ImportMessagesFromJSON(jsonString)
```

## Customization

### Hybrid Approach: Manual History Control

```go
// Create stateless agent
agent, _ := chat.NewAgent(
    ctx,
    agents.Config{
        KeepConversationHistory: false,
        // ...
    },
    modelConfig,
)

// Manually manage context for specific requests
var conversationHistory []messages.Message

// Add user message to manual history
userMsg := messages.Message{Role: roles.User, Content: "What is Go?"}
conversationHistory = append(conversationHistory, userMsg)

// Generate with manual context
result, _ := agent.GenerateCompletion(conversationHistory)

// Add assistant response to manual history
conversationHistory = append(conversationHistory, messages.Message{
    Role:    roles.Assistant,
    Content: result.Response,
})

// Next request uses accumulated manual history
userMsg2 := messages.Message{Role: roles.User, Content: "Tell me more"}
conversationHistory = append(conversationHistory, userMsg2)
result2, _ := agent.GenerateCompletion(conversationHistory)
```

### Sliding Window History

```go
type SlidingWindowAgent struct {
    agent      *chat.Agent
    maxHistory int
    messages   []messages.Message
}

func (s *SlidingWindowAgent) Chat(userMessage string) (string, error) {
    // Add user message
    s.messages = append(s.messages, messages.Message{
        Role:    roles.User,
        Content: userMessage,
    })

    // Keep only last N messages (sliding window)
    if len(s.messages) > s.maxHistory {
        s.messages = s.messages[len(s.messages)-s.maxHistory:]
    }

    // Generate with windowed context
    result, err := s.agent.GenerateCompletion(s.messages)
    if err != nil {
        return "", err
    }

    // Add assistant response
    s.messages = append(s.messages, messages.Message{
        Role:    roles.Assistant,
        Content: result.Response,
    })

    return result.Response, nil
}
```

### Context Persistence

```go
import (
    "encoding/json"
    "os"
)

// Save conversation to file
func SaveConversation(agent *chat.Agent, filepath string) error {
    jsonData := agent.ExportMessagesToJSON()
    return os.WriteFile(filepath, []byte(jsonData), 0644)
}

// Load conversation from file
func LoadConversation(agent *chat.Agent, filepath string) error {
    data, err := os.ReadFile(filepath)
    if err != nil {
        return err
    }
    return agent.ImportMessagesFromJSON(string(data))
}

// Usage
SaveConversation(agent, "./conversations/session-123.json")
LoadConversation(agent, "./conversations/session-123.json")
```

## Use Case Examples

### API Endpoint (Stateless)

```go
// Each HTTP request should be independent
func handleChatRequest(w http.ResponseWriter, r *http.Request) {
    agent, _ := chat.NewAgent(
        r.Context(),
        agents.Config{
            KeepConversationHistory: false, // Stateless
        },
        modelConfig,
    )

    var req struct {
        Message string `json:"message"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    result, _ := agent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: req.Message},
    })

    json.NewEncoder(w).Encode(map[string]string{
        "response": result.Response,
    })
}
```

### Customer Support (Stateful)

```go
// Maintain full conversation context
var supportAgent *chat.Agent

func initSupportAgent() {
    supportAgent, _ = chat.NewAgent(
        context.Background(),
        agents.Config{
            SystemInstructions:      "You are a customer support agent...",
            KeepConversationHistory: true, // Stateful
        },
        modelConfig,
    )
}

func handleSupportMessage(userMessage string) string {
    result, _ := supportAgent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: userMessage},
    })
    return result.Response
}
```

## Important Notes

### DO:
- Use `KeepConversationHistory: false` for stateless APIs and independent requests
- Use `KeepConversationHistory: true` for multi-turn conversations
- Set `Temperature: 0.0` for stateless agents (deterministic)
- Set `Temperature: 0.0-0.8` for stateful agents (more natural)
- Monitor context size with `agent.GetContextSize()` to avoid token limits
- Use `agent.ResetMessages()` to clear history when needed
- Export/import messages for session persistence across restarts
- Manually manage history for fine-grained control

### DON'T:
- Don't assume history is kept - always check `KeepConversationHistory` setting
- Don't let stateful agents accumulate unlimited history - use compression or sliding windows
- Don't use stateful agents for batch processing (memory inefficient)
- Don't forget to reset history between different user sessions
- Don't mix manual and automatic history management

## Behavior Summary

| KeepConversationHistory | Messages Kept | Use Cases | Memory Usage |
|-------------------------|---------------|-----------|--------------|
| `false` | System only | APIs, classification, routing | Low |
| `true` | All messages | Chat, support, assistants | High |

## Troubleshooting

### Context Too Large
```go
// Check context size
if agent.GetContextSize() > 4000 {
    // Option 1: Reset and start fresh
    agent.ResetMessages()

    // Option 2: Use compressor agent (see compressor snippet)

    // Option 3: Use sliding window
}
```

### Lost Context
```go
// Verify history is enabled
config := agents.Config{
    KeepConversationHistory: true, // Must be true
}

// Check messages are being kept
fmt.Println("Message count:", len(agent.GetMessages()))
```

### Memory Leaks
```go
// For long-running stateful agents, periodically clear old sessions
if sessionExpired {
    agent.ResetMessages()
}

// Or use a fresh agent per session
agent, _ = chat.NewAgent(ctx, config, modelConfig)
```
