# Structured Agent With Conversation History

This example demonstrates how to use a structured data agent with `KeepConversationHistory` set to `true`.

## Behavior

When `KeepConversationHistory` is set to `true`:
- System messages are kept in the conversation history
- User messages are sent to the LLM and stored in history
- Assistant responses are generated and stored in history
- Each request has context from all previous requests

## Use Cases

This is useful when:
- You want stateful conversations (each request builds on previous ones)
- You need the LLM to remember previous interactions
- You want to maintain conversation context
- You're building chatbots or interactive assistants

## Running the Example

```bash
go run main.go
```

## Expected Output

The example makes two requests:
1. "Tell me about Canada"
2. "Tell me about France"

After each request, it shows the message count. With `KeepConversationHistory=true`, all messages (system, user, and assistant) are kept in history.
