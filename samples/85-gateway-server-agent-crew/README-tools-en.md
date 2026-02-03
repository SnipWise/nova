# 🛠️ Tools Usage Guide with Gateway Server

This document explains how to use tools (tools/functions) with the gateway server in **Passthrough** mode.

## 📋 Table of Contents

- [What is Passthrough Mode?](#what-is-passthrough-mode)
- [How Does It Work?](#how-does-it-work)
- [Usage with qwen-code](#usage-with-qwen-code)
- [Tools Format](#tools-format)
- [Practical Examples](#practical-examples)
- [Auto-Execute vs Passthrough Mode](#auto-execute-vs-passthrough-mode)

## What is Passthrough Mode?

**Passthrough mode** (transparency) is the default mode of the gateway server. In this mode:

- 🔄 The gateway **transmits** tool calls from the LLM to the client
- 💻 The **client** (qwen-code, aider, continue.dev, etc.) **executes** the tools
- 📤 The client sends the results back to the gateway
- 🔁 The gateway transmits the results to the LLM to continue the conversation

### Advantages of Passthrough Mode

✅ **Flexibility**: The client controls which tools are available
✅ **Security**: Tools execute in the client's environment, not on the server
✅ **Simplicity**: No need to configure tools on the server side
✅ **Compatible**: Works with all standard OpenAI clients

## How Does It Work?

### Communication Flow

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│             │         │             │         │             │
│  qwen-code  │◄───────►│   Gateway   │◄───────►│     LLM     │
│  (Client)   │         │   Server    │         │   Backend   │
│             │         │             │         │             │
└─────────────┘         └─────────────┘         └─────────────┘
      │                                                 │
      │ 1. Sends request with tools                     │
      │────────────────────────────────────────────────►│
      │                                                 │
      │ 2. LLM decides to call a tool                   │
      │◄────────────────────────────────────────────────│
      │                                                 │
      │ 3. Client executes the tool                     │
      │                                                 │
      │ 4. Sends back the result                        │
      │────────────────────────────────────────────────►│
      │                                                 │
      │ 5. LLM generates final response                 │
      │◄────────────────────────────────────────────────│
```

### Detailed Steps

1. **The client sends a request** with the list of available tools
2. **The gateway routes** to the appropriate agent (coder, thinker, generic)
3. **The LLM decides** whether to use a tool
4. **The gateway returns** the tool call to the client (with `finish_reason: "tool_calls"`)
5. **The client executes** the tool locally
6. **The client sends back** the result with `role: "tool"`
7. **The LLM generates** the final response based on the result

## Usage with qwen-code

### Configuration

```bash
export OPENAI_BASE_URL=http://localhost:8080/v1
export OPENAI_API_KEY=none
export OPENAI_MODEL=crew
```

### Launch

```bash
# Terminal 1: Start the gateway
cd samples/85-gateway-server-agent-crew
go run main.go

# Terminal 2: Use qwen-code
qwen-code
```

### Tool Configuration in qwen-code

Qwen-code must be configured with available tools. Here's an example configuration:

```json
{
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "read_file",
        "description": "Read the contents of a file",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "Path to the file to read"
            }
          },
          "required": ["path"]
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "write_file",
        "description": "Write content to a file",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "Path to the file to write"
            },
            "content": {
              "type": "string",
              "description": "Content to write to the file"
            }
          },
          "required": ["path", "content"]
        }
      }
    }
  ]
}
```

## Tools Format

### Tool Definition (sent by the client)

```json
{
  "type": "function",
  "function": {
    "name": "function_name",
    "description": "Description of what the function does",
    "parameters": {
      "type": "object",
      "properties": {
        "param1": {
          "type": "string",
          "description": "Parameter description"
        },
        "param2": {
          "type": "number",
          "description": "Parameter description"
        }
      },
      "required": ["param1"]
    }
  }
}
```

### Tool Call (returned by the LLM)

```json
{
  "id": "call_abc123",
  "type": "function",
  "function": {
    "name": "function_name",
    "arguments": "{\"param1\":\"value1\",\"param2\":42}"
  }
}
```

### Tool Result (sent by the client)

```json
{
  "role": "tool",
  "content": "{\"result\": \"success\", \"data\": \"...\"}",
  "tool_call_id": "call_abc123"
}
```

## Practical Examples

### Example 1: Simple Request with Tools

**Initial client request:**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What time is it?"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_current_time",
          "description": "Get the current time",
          "parameters": {"type": "object", "properties": {}}
        }
      }
    ]
  }'
```

**Gateway response (tool call):**

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "model": "crew",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": null,
        "tool_calls": [
          {
            "id": "call_xyz",
            "type": "function",
            "function": {
              "name": "get_current_time",
              "arguments": "{}"
            }
          }
        ]
      },
      "finish_reason": "tool_calls"
    }
  ]
}
```

**Client request with result:**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What time is it?"},
      {
        "role": "assistant",
        "content": null,
        "tool_calls": [
          {
            "id": "call_xyz",
            "type": "function",
            "function": {
              "name": "get_current_time",
              "arguments": "{}"
            }
          }
        ]
      },
      {
        "role": "tool",
        "content": "{\"time\": \"14:30:00\"}",
        "tool_call_id": "call_xyz"
      }
    ],
    "tools": [...]
  }'
```

**Final response:**

```json
{
  "id": "chatcmpl-124",
  "object": "chat.completion",
  "model": "crew",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "It is currently 14:30:00 (2:30 PM)."
      },
      "finish_reason": "stop"
    }
  ]
}
```

### Example 2: With qwen-code (automatic)

Qwen-code automatically handles this flow:

```
👤 User: "Read the file package.json"

🤖 LLM: [calls read_file with path="package.json"]
         ↓
💻 qwen-code: [executes file read]
         ↓
🤖 LLM: "Here is the content of package.json: ..."
```

## Auto-Execute vs Passthrough Mode

| Aspect | Passthrough (default) | Auto-Execute |
|--------|----------------------|--------------|
| **Execution** | Client | Server |
| **Configuration** | Tools defined by client | Tools defined server-side |
| **Security** | Tools in client environment | Tools in server environment |
| **Flexibility** | Client controls tools | Server controls tools |
| **Use case** | Applications with local access (IDE, CLI) | Web services, APIs |

### When to use Passthrough?

✅ Desktop applications (qwen-code, IDE extensions)
✅ CLI tools that have access to the local file system
✅ When the client needs to control available tools
✅ For security reasons (tool isolation)

### When to use Auto-Execute?

✅ Web services without an intelligent client
✅ Public APIs with predefined tools
✅ When all clients should have the same tools
✅ Simple web chatbots

## Advanced Configuration

### Enable Auto-Execute Mode

If you want to switch to Auto-Execute mode, modify [main.go](main.go):

```go
gateway, err := gatewayserver.NewAgent(
    ctx,
    gatewayserver.WithAgentCrew(agentCrew, "generic"),
    gatewayserver.WithPort(8080),

    // Enable Auto-Execute mode
    gatewayserver.WithToolMode(gatewayserver.ToolModeAutoExecute),
    gatewayserver.WithToolsAgent(toolsAgent),
    gatewayserver.WithExecuteFn(executeFunction),
)
```

And uncomment the tool definitions in `getToolsDefinitions()`.

## Debugging

### Check Requests

Enable detailed logs:

```go
if err := os.Setenv("NOVA_LOG_LEVEL", "DEBUG"); err != nil {
    panic(err)
}
```

### Diagnostic Messages

The gateway displays:
- 📥 `Request received (current agent: X)`: Request received
- 📤 `Response sent (agent used: X)`: Response sent
- 🔵 `Matching agent for topic: X`: Selected agent

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `400 Invalid request body: json: cannot unmarshal array` | Incorrect content format | ✅ Fixed in this version |
| `finish_reason: "tool_calls"` but no tool_calls | LLM misconfigured | Verify that the model supports tools |
| No response after tool call | Client didn't send back result | Check client implementation |

## Multi-modal Support

The gateway now supports **three formats** for the `content` field:

### 1. Simple String (legacy)
```json
{"role": "user", "content": "Hello"}
```

### 2. Array of Strings (qwen-code)
```json
{"role": "user", "content": ["Hello", "world"]}
```

### 3. Array of Objects (multi-modal OpenAI)
```json
{
  "role": "user",
  "content": [
    {"type": "text", "text": "Hello"},
    {"type": "image_url", "image_url": {"url": "..."}}
  ]
}
```

All formats are automatically converted to plain text by the gateway.

## Resources

- [OpenAI Tools Documentation](https://platform.openai.com/docs/guides/function-calling)
- [Qwen Code GitHub](https://github.com/QwenLM/qwen-code)
- [Nova SDK Documentation](../../README.md)

---

**Version:** 1.0.0
**Last updated:** 2026-02-02
