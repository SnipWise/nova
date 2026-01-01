# Server Agent in CLI Mode with StreamCompletion

This example demonstrates how to use a **Server Agent** in **CLI mode** using the `StreamCompletion` method, which provides the same functionality as the Crew Agent's streaming method.

## Key Features

- **Dual Mode**: The same agent can be used in both CLI and HTTP server modes
- **StreamCompletion**: Uses the crew-style `StreamCompletion` method for CLI interactions
- **Tool Execution**: Supports tool calls with automatic confirmation in CLI mode
- **Streaming**: Real-time streaming responses
- **Consistent API**: Same interface as Crew Agent for easy code migration

## What's New

The Server Agent now has a `StreamCompletion` method that mirrors the Crew Agent's functionality:

```go
agent.StreamCompletion(question, func(chunk string, finishReason string) error {
    if chunk != "" {
        fmt.Print(chunk)
    }
    return nil
})
```

This method handles:
1. Context compression (if configured)
2. Tool call detection and execution
3. RAG context integration (if configured)
4. Streaming response generation

## CLI vs Server Mode

### CLI Mode (this example)
- Uses `StreamCompletion` method
- Interactive console-based interface
- Auto-confirms tool execution by default
- Can set custom confirmation prompts with `SetConfirmationPromptFunction`

### Server Mode (sample 49-server-agent-stream)
- Uses `StartServer` method
- HTTP API with SSE streaming
- Web-based confirmation via `/operation/validate` endpoint
- Separate endpoints for all operations

## Usage

### Prerequisites
Make sure you have a model server running:
```bash
# Using Docker Model Runner or Ollama
docker model pull ai/qwen2.5:1.5B-F16
# or
docker model pull hf.co/menlo/jan-nano-gguf:q4_k_m
```

### Run the Example
```bash
cd samples/72-server-agent-cli-stream
go mod tidy
go run main.go
```

### Try These Commands
```
🧑 You: Hello! Say hello to Alice
🤖 Bob: [executes say_hello tool and responds]

🧑 You: What is 42 plus 58?
🤖 Bob: [executes calculate_sum tool and responds]

🧑 You: What time is it?
🤖 Bob: [executes get_current_time tool and responds]
```

## Customization

### Custom Confirmation Prompt
By default, tool calls are auto-confirmed in CLI mode. To add manual confirmation:

```go
agent.SetConfirmationPromptFunction(func(functionName string, arguments string) tools.ConfirmationResponse {
    fmt.Printf("\n⚠️  Execute %s? (y/n): ", functionName)
    response, _ := input.ReadInput()
    if response == "y" {
        return tools.Confirmed
    }
    return tools.Denied
})
```

### Add RAG Support
```go
ragAgent, _ := rag.NewAgent(ctx, ragConfig, modelConfig, embeddingConfig)
agent.SetRagAgent(ragAgent)
agent.SetSimilarityLimit(0.7)
agent.SetMaxSimilarities(3)
```

### Add Context Compression
```go
compressorAgent, _ := compressor.NewAgent(ctx, compressorConfig, modelConfig)
agent.SetCompressorAgent(compressorAgent)
agent.SetContextSizeLimit(8000)
```

## Benefits of Using Server Agent in CLI Mode

1. **Code Reusability**: Same agent code works in both CLI and HTTP modes
2. **Testing**: Test your server agent logic in CLI before deploying as HTTP service
3. **Consistency**: Same behavior across crew agents and server agents
4. **Flexibility**: Easy to switch between interactive and server modes

## Related Examples

- **49-server-agent-stream**: Server agent in HTTP mode
- **50-server-agent-with-tools**: Server agent with tools (HTTP mode)
- **54-server-agent-tools-rag-compress**: Full-featured server agent (HTTP mode)
- **55-crew-agent**: Crew agent with StreamCompletion
- **69-simple-crew-agent**: Simple crew agent example

## Architecture

```
ServerAgent
  ├── CLI Mode (this example)
  │   └── StreamCompletion()
  │       ├── Compress context if needed
  │       ├── Detect and execute tool calls
  │       ├── Add RAG context
  │       └── Stream response
  │
  └── HTTP Mode (sample 49)
      └── StartServer()
          └── handleCompletion()
              ├── Compress context if needed
              ├── Detect and execute tool calls
              ├── Add RAG context
              └── Stream response via SSE
```

Both modes share the same underlying logic but differ in:
- Input/output handling
- Tool confirmation mechanism
- Response streaming method
