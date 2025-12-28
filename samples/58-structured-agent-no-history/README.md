# Structured Agent Without Conversation History

This example demonstrates how to use a structured data agent with `KeepConversationHistory` set to `false`.

## Behavior

When `KeepConversationHistory` is set to `false`:
- Only the system message is kept in the conversation history
- User messages are sent to the LLM but not stored in history
- Assistant responses are generated but not stored in history
- Each request is independent and doesn't have context from previous requests

## Use Cases

This is useful when:
- You want stateless requests (each request is independent)
- You want to minimize memory usage
- You don't need conversation context between requests
- You want to prevent context from growing over time

## Running the Example

```bash
go run main.go
```

## Expected Output

The example makes two requests:
1. "Tell me about Canada"
2. "Tell me about France"

After each request, it shows the message count. With `KeepConversationHistory=false`, only the system message(s) remain in history.
