# Tools Agent Without Conversation History

This example demonstrates how to use a tools agent with `KeepConversationHistory` set to `false`.

## Behavior

When `KeepConversationHistory` is set to `false`:
- Only the system message is kept in the conversation history
- User messages are sent to the LLM but not stored in history
- Tool calls and results are processed but not stored in history
- Assistant responses are generated but not stored in history
- Each request is independent and doesn't have context from previous requests

## Use Cases

This is useful when:
- You want stateless tool executions (each request is independent)
- You want to minimize memory usage in tool-based workflows
- You don't need conversation context between tool calls
- You're processing independent tasks that don't relate to each other
- You want to prevent context from growing with tool call history

## Running the Example

```bash
go run main.go
```

## Expected Output

The example makes two requests:
1. "Make the sum of 40 and 2" - calls calculate_sum tool
2. "Say hello to Alice" - calls say_hello tool

After each request, it shows the message count. With `KeepConversationHistory=false`, only the system message remains in history.

## Tools Available

- **calculate_sum**: Calculates the sum of two numbers
- **say_hello**: Says hello to a given name
