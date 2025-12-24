# Tools Agent With Conversation History

This example demonstrates how to use a tools agent with `KeepConversationHistory` set to `true`.

## Behavior

When `KeepConversationHistory` is set to `true`:
- System messages are kept in the conversation history
- User messages are sent to the LLM and stored in history
- Tool calls and results are stored in history
- Assistant responses are generated and stored in history
- Each request has context from all previous requests and tool executions

## Use Cases

This is useful when:
- You want stateful tool-based workflows (each request builds on previous ones)
- You need the LLM to remember previous tool calls and results
- You want to maintain conversation context in complex multi-step tasks
- You're building agents that need to track their actions over time
- You need follow-up tool calls to work with context from previous ones

## Running the Example

```bash
go run main.go
```

## Expected Output

The example makes two requests:
1. "Make the sum of 40 and 2" - calls calculate_sum tool
2. "Say hello to Alice" - calls say_hello tool

After each request, it shows the message count. With `KeepConversationHistory=true`, all messages (system, user, tool, and assistant) are kept in history, allowing the agent to maintain context across tool calls.

## Tools Available

- **calculate_sum**: Calculates the sum of two numbers
- **say_hello**: Says hello to a given name
