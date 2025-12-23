# Chat Agent with Telemetry

This example demonstrates how to use the telemetry features of the Nova SDK to track and monitor LLM requests and responses.

## Features Demonstrated

- **Request Metadata**: Capture information about each request sent to the LLM
  - Model name
  - Context length
  - Temperature, max tokens, and other parameters
  - Request timestamp

- **Response Metadata**: Track detailed response information
  - Response ID
  - Token usage (prompt, completion, total)
  - Response time in milliseconds
  - Finish reason
  - Response timestamp

- **Session Tracking**: Monitor cumulative metrics across multiple requests
  - Total tokens used in the session
  - Full conversation history as JSON

## Key Methods

### Request Telemetry
```go
// Get last request as JSON
requestJSON, _ := agent.GetLastRequestJSON()

// Get request metadata
metadata := agent.GetLastRequestMetadata()
fmt.Printf("Context Length: %d\n", metadata.ContextLength)
fmt.Printf("Temperature: %.2f\n", metadata.Temperature)
```

### Response Telemetry
```go
// Get last response as JSON
responseJSON, _ := agent.GetLastResponseJSON()

// Get response metadata
metadata := agent.GetLastResponseMetadata()
fmt.Printf("Total Tokens: %d\n", metadata.TotalTokens)
fmt.Printf("Response Time: %d ms\n", metadata.ResponseTime)
```

### Session Telemetry
```go
// Get cumulative token usage
totalTokens := agent.GetTotalTokensUsed()

// Get full conversation history
historyJSON, _ := agent.GetConversationHistoryJSON()
```

## Use Cases

This telemetry functionality is useful for:

1. **Cost Tracking**: Monitor token usage to estimate API costs
2. **Performance Monitoring**: Track response times and identify bottlenecks
3. **Debugging**: Inspect full request/response payloads for troubleshooting
4. **Analytics**: Gather metrics for usage patterns and optimization
5. **Logging**: Create detailed audit trails of LLM interactions
6. **Rate Limiting**: Track request frequency and token consumption

## Running the Example

```bash
# Make sure you have a compatible LLM server running
# For example, with Ollama:
# ollama serve

# Run the example
cd samples/57-chat-agent-telemetry
go run main.go
```

## Expected Output

The example will:
1. Send two questions to the LLM
2. Display the responses
3. Show detailed telemetry for each request/response
4. Print the full request and response JSON
5. Display cumulative token usage
6. Export the conversation history

## Next Steps

See example [58-chat-agent-telemetry-callback](../58-chat-agent-telemetry-callback) for real-time telemetry monitoring using callbacks.
