# Chat Agent

Simplified conversational agent that abstracts OpenAI SDK details.

## Creation

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/models"
)

agent, err := chat.NewAgent(ctx, agentConfig, modelConfig)
```

## Core Features

### 1. Response Generation

#### Simple Completion
```go
result, err := agent.GenerateCompletion(userMessages)
// result.Response (string)
// result.FinishReason (string)
```

#### Streaming Completion
```go
callback := func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
}

result, err := agent.GenerateStreamCompletion(userMessages, callback)
```

#### Completion with Reasoning
```go
result, err := agent.GenerateCompletionWithReasoning(userMessages)
// result.Response (string)
// result.Reasoning (string)
// result.FinishReason (string)
```

#### Streaming with Reasoning
```go
reasoningCallback := func(chunk string, finishReason string) error {
    fmt.Print("[Reasoning] ", chunk)
    return nil
}

responseCallback := func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
}

result, err := agent.GenerateStreamCompletionWithReasoning(
    userMessages,
    reasoningCallback,
    responseCallback,
)
```

### 2. Message Management

```go
// Add a message
agent.AddMessage(roles.User, "Hello")

// Add multiple messages
agent.AddMessages([]messages.Message{...})

// Get all messages
msgs := agent.GetMessages()

// Remove last N messages (except system message)
agent.RemoveLastNMessages(3)

// Reset (keeps only system message)
agent.ResetMessages()
```

### 3. System Instructions

```go
// Set/update system instructions
agent.SetSystemInstructions("You are a helpful assistant...")
```

### 4. Pre/Post Directives for User Messages

**Use case**: Consistently frame all user messages with additional context or instructions.

```go
// Add context BEFORE the last user message
agent.SetUserMessagePreDirectives("Context: You are a technical support agent.")

// Add instructions AFTER the last user message
agent.SetUserMessagePostDirectives("Always respond in French.")

// Get directives
pre := agent.GetUserMessagePreDirectives()
post := agent.GetUserMessagePostDirectives()
```

**How it works**: Directives are automatically added to the last user message during each call to `GenerateCompletion`, `GenerateStreamCompletion`, etc.

**Example**:
```go
agent.SetUserMessagePreDirectives("You are an expert in Go programming.")
agent.SetUserMessagePostDirectives("Keep your answer under 100 words.")

// User message: "How do I use goroutines?"
// Actual message sent to model:
// "You are an expert in Go programming.\n\nHow do I use goroutines?\n\nKeep your answer under 100 words."
```

### 5. Stream Control

```go
// Interrupt ongoing streaming
agent.StopStream()
```

### 6. Context and Metadata

```go
// Approximate current context size
size := agent.GetContextSize()

// Agent type
kind := agent.Kind() // agents.Chat

// Agent name
name := agent.GetName()

// Model ID
modelID := agent.GetModelID()
```

### 7. Export and Inspection

```go
// Export conversation as JSON
jsonStr, err := agent.ExportMessagesToJSON()

// Last request (raw JSON)
rawReq := agent.GetLastRequestRawJSON()

// Last response (raw JSON)
rawResp := agent.GetLastResponseRawJSON()

// Last request (formatted JSON)
reqJSON, err := agent.GetLastRequestJSON()

// Last response (formatted JSON)
respJSON, err := agent.GetLastResponseJSON()
```

### 8. Configuration

```go
// Get agent configuration
config := agent.GetConfig()

// Update agent configuration
agent.SetConfig(newConfig)

// Get model configuration
modelConfig := agent.GetModelConfig()

// Update model configuration
agent.SetModelConfig(newModelConfig)
```

### 9. Go Context

```go
// Get context.Context
ctx := agent.GetContext()

// Set new context.Context
agent.SetContext(newCtx)
```

## Types

### CompletionResult
```go
type CompletionResult struct {
    Response     string
    FinishReason string
}
```

### ReasoningResult
```go
type ReasoningResult struct {
    Response     string
    Reasoning    string
    FinishReason string
}
```

### StreamCallback
```go
type StreamCallback func(chunk string, finishReason string) error
```

## Important Notes

- Conversation history is automatically managed by `BaseAgent` according to `KeepConversationHistory`
- System message is preserved when calling `ResetMessages()`
- Pre/post directives are automatically applied to the last user message
- Streaming can be interrupted with `StopStream()`
