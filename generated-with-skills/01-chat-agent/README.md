# 01 - Streaming Chat Agent

> Generated with Nova Agent Builder skill

## Description

Interactive chat agent that streams responses token by token for a more responsive user experience.

## Features

- Real-time streaming responses
- Simple command-line interface
- Conversation loop with context
- Configurable model and temperature

## Prerequisites

- Model `ai/qwen2.5:1.5B-F16` available

## Installation

```bash
cd generated-with-skills/01-chat-agent
go mod init chat-agent
go mod tidy
```

## Usage

```bash
go run main.go
```

### Interactive Session

```
🤖 Streaming Chat Agent
Type 'quit' to exit
----------------------------------------

👤 You: Hello!
🤖 Assistant: Hello! How can I help you today?
   [finish_reason: stop]

👤 You: What is Go?
🤖 Assistant: Go is a statically typed, compiled programming language designed at Google...
   [finish_reason: stop]

👤 You: quit
Goodbye!
```

## Configuration

Edit the configuration in `main.go`:

```go
agents.Config{
    Name:               "streaming-assistant",
    EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
    SystemInstructions: "You are a helpful and friendly assistant.",
}

models.Config{
    Name:        "ai/qwen2.5:1.5B-F16",
    Temperature: models.Float64(0.7),
    MaxTokens:   models.Int(2000),
}
```

### Available Parameters

- **Name**: Agent identifier
- **EngineURL**: Nova engine endpoint
- **SystemInstructions**: Agent's personality/role
- **Model**: LLM model to use
- **Temperature**: Creativity (0.0-1.0)
- **MaxTokens**: Maximum response length

## Customization

### Add Conversation History

Modify the code to maintain context across messages:

```go
var conversationHistory []messages.Message

// Add system message at start
conversationHistory = append(conversationHistory, messages.Message{
    Role:    roles.System,
    Content: "You are a helpful assistant.",
})

// In the loop, add user message
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

### Different Models

Change the model to suit your needs:

```go
// Fast, lightweight
Name: "ai/qwen2.5:0.5B-F16"

// Balanced
Name: "ai/qwen2.5:1.5B-F16"

// More capable
Name: "hf.co/menlo/lucy-gguf:q4_k_m"

// Specialized for coding
Name: "hf.co/quantfactory/deepseek-coder-7b-instruct-v1.5-gguf:q4_k_m"
```

## Related Snippets

- **contextual-chat**: Chat with persistent conversation memory
- **chat/streaming-chat**: This snippet
- See [CLAUDE.md](../../CLAUDE.md) for all available snippets

## Reference

- Sample: `samples/05-chat-agent`
- Snippet: `.claude/skills/nova-agent-builder/snippets/chat/streaming-chat.md`
- Category: `chat`
- Complexity: `beginner`
