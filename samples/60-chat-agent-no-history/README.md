# Chat Agent Without Conversation History

This example demonstrates how to use a chat agent with `KeepConversationHistory` set to `false`.

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
- You're processing independent queries that don't relate to each other

## Running the Example

```bash
go run main.go
```

## Expected Output

The example makes three requests:
1. "Hello, what is your name?"
2. "Who is James T Kirk?"
3. "Who is his best friend?" (should NOT know the context)

After each request, it shows the message count. With `KeepConversationHistory=false`, only the system message remains in history, and the third request won't know who "his" refers to.
