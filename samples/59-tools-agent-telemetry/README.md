# Tools Agent with Telemetry

This example demonstrates telemetry tracking for a tools agent that uses function calling.

## Features

- Define multiple tools (weather, calculator)
- Execute parallel tool calls
- Track telemetry for tool-enhanced interactions
- Monitor token usage including tool definitions
- Inspect request/response with tool calls in JSON

## Key Points

When using tools, the telemetry captures:
- **Tool definitions** in the request (increases context length)
- **Tool call results** in the messages
- **Multiple LLM interactions** (initial request + tool result processing)
- **Token usage** for both tool calling and response generation

## Running

```bash
cd samples/59-tools-agent-telemetry
go run main.go
```

## Expected Behavior

1. Agent receives user question requiring multiple tools
2. LLM makes parallel tool calls
3. Tools are executed and results returned
4. LLM generates final response based on tool results
5. Telemetry shows complete interaction metrics
