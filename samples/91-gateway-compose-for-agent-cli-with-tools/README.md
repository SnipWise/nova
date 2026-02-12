# Gateway Crew with Configuration-Based Agent Routing & Passthrough-First Tool Handling

This sample demonstrates an advanced Gateway Server Agent with:
- **Configuration-based agent routing** (agent-routing.json)
- **Passthrough-first tool handling** - ensures all tool requests are processed by a tool-capable agent
- **Multi-agent crew** with specialized agents (coder, generic, passthrough)
- **Intelligent orchestration** for automatic topic detection and routing
- **Context compression** for long conversations
- **BeforeCompletion hooks** for request tracing

## 🎯 Key Feature: Passthrough-First Architecture

When a client (like `pi`, `qwen-code`, or `aider`) sends a request with tools:

1. **Phase 1: Detection** - The passthrough agent analyzes the request
2. **Phase 2: Decision**:
   - If tool_calls needed → Returns tool_calls to client
   - If no tool_calls → Redirects to appropriate agent (orchestrator/default)

This ensures that **all tool requests are handled by a model that supports tool calling**, preventing errors with models that don't support tools.

```
Client + tools[]
    ↓
🔀 PASSTHROUGH AGENT (always first)
    ├─ Detects tool_calls?
    │  ├─ YES → Returns tool_calls to client
    │  └─ NO  → Redirects to selected agent
    ↓
💬 Selected Agent (coder/generic)
    └─ Responds without tools
```

## Architecture

```
┌────────────────────────────────────────────┐
│         Gateway Server (Port 8080)         │
│                                            │
│  ┌─────────────────────────────────────┐   │
│  │       Agent Crew                    │   │
│  │  ┌────────────┐  ┌──────────────┐   │   │
│  │  │   Coder    │  │   Generic    │   │   │
│  │  │  (coding)  │  │  (default)   │   │   │
│  │  └────────────┘  └──────────────┘   │   │
│  │  ┌────────────────────────────────┐ │   │
│  │  │  Passthrough (tool handling)   │ │   │
│  │  │  - Always checks for tools     │ │   │
│  │  │  - Redirects if no tools needed│ │   │
│  │  └────────────────────────────────┘ │   │
│  └─────────────────────────────────────┘   │
│                                            │
│  ┌─────────────────────────────────────┐   │
│  │     Orchestrator Agent              │   │
│  │  (Detects topic → selects agent)    │   │
│  └─────────────────────────────────────┘   │
│                                            │
│  ┌─────────────────────────────────────┐   │
│  │     Compressor Agent                │   │
│  │  (Compresses context > 16K tokens)  │   │
│  └─────────────────────────────────────┘   │
└────────────────────────────────────────────┘
```

## Configuration Files

### agent-routing.json

Defines which agents handle which topics:

```json
{
  "routing": [
    {
      "topics": ["coding", "programming", "development", "code", "software", "debugging", "technology"],
      "agent": "coder"
    }
  ],
  "default_agent": "generic"
}
```

### orchestrator.instructions.md

Instructions for the orchestrator agent that detects topics:

```markdown
You identify the topic of a conversation in one word.
Possible topics: Computing, Programming, Technology, Health, Science, Mathematics, Philosophy, Food, Education.
Respond in JSON with the field 'topic_discussion'.
```

## Quick Start

### 1. Build

```bash
go build -o gateway-server
```

### 2. Configure Environment Variables (optional)

```bash
export ENGINE_URL="http://localhost:12434/engines/llama.cpp/v1"
export CODER_MODEL_ID="hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m"
export GENERIC_MODEL_ID="hf.co/menlo/jan-nano-gguf:q4_k_m"
export PASSTHROUGHT_MODEL_ID="hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m"  # Tool-capable model
export ORCHESTRATOR_MODEL_ID="hf.co/menlo/lucy-gguf:q4_k_m"
export COMPRESSOR_MODEL_ID="ai/qwen2.5:0.5B-F16"
```

### 3. Start the Gateway

```bash
./gateway-server
```

Output:
```
🚀 Gateway crew server starting on http://localhost:8080
📡 OpenAI-compatible endpoint: POST /v1/chat/completions
👥 Crew agents: coder, generic, passthrough
🔧 Tools mode: passthrough (client-side)
```

## Usage Examples

### With curl (non-streaming)

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Write a Go function to calculate fibonacci"}
    ]
  }'
```

### With curl (streaming)

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Explain quantum computing"}
    ],
    "stream": true
  }'
```

### With curl (with tools - passthrough mode)

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What is the weather in Paris?"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather",
          "description": "Get the current weather in a location",
          "parameters": {
            "type": "object",
            "properties": {
              "location": {"type": "string", "description": "The city name"}
            },
            "required": ["location"]
          }
        }
      }
    ],
    "stream": true
  }'
```

### With qwen-code

```bash
export OPENAI_BASE_URL=http://localhost:8080/v1
export OPENAI_API_KEY=none
export OPENAI_MODEL=crew
qwen-code
```

### With pi (from pi-mono project)

Configure `settings.json`:
```json
{
  "defaultProvider": "dockermodelrunner",
  "defaultModel": "crew"
}
```

Configure `models.json`:
```json
{
  "providers": {
    "dockermodelrunner": {
      "baseUrl": "http://localhost:8080/v1",
      "api": "openai-completions",
      "apiKey": "hello",
      "models": [
        { "id": "crew" }
      ]
    }
  }
}
```

Then run:
```bash
pi
```

### With aider

```bash
aider --openai-api-base http://localhost:8080/v1 --model crew
```

## Testing

Run the test script to validate the passthrough-first behavior:

```bash
./test-passthrough.sh
```

This script tests:
1. ✅ Request with tools that triggers tool_calls (passthrough handles it)
2. ✅ Request with tools but simple question (redirects to appropriate agent)
3. ✅ Request without tools (normal orchestrator behavior)

## Request Flow Examples

### Example 1: Coding Question (No Tools)

```
Client: "Write a Python function to reverse a string"
    ↓
Gateway receives request (no tools)
    ↓
Orchestrator detects topic: "coding"
    ↓
Routes to: Coder Agent
    ↓
🔧 [CODER AGENT] Processing request...
    ↓
Returns: Python code
```

### Example 2: Tool-Required Question

```
Client: "What is the weather in Paris?" + tools[]
    ↓
Gateway receives request (with tools)
    ↓
🔀 [PASSTHROUGH AGENT] Checking for tool calls...
    ├─ Phase 1: Fast detection (non-streaming)
    ├─ Detects: tool_calls needed
    ├─ Phase 2: Streaming call
    └─ Returns: tool_calls to client
    ↓
Client executes get_weather() locally
    ↓
Client sends results back
    ↓
Final response with weather data
```

### Example 3: Simple Question with Tools Present

```
Client: "What is 2+2?" + tools[]
    ↓
Gateway receives request (with tools)
    ↓
🔀 [PASSTHROUGH AGENT] Checking for tool calls...
    ├─ Phase 1: Fast detection
    ├─ Detects: no tool_calls needed
    └─ Redirects to appropriate agent
    ↓
Orchestrator detects topic: "generic"
    ↓
💬 [GENERIC AGENT] Processing request...
    ↓
Returns: "4"
```

## Log Output Examples

### Successful Tool Call Detection (Streaming)

```
📥 Request received (current agent: generic)
🔀 [PASSTHROUGH AGENT] Checking for tool calls...
🔀 Routing to passthrough agent first to check for tool calls
🔍 Phase 1: Fast detection call (non-streaming)
✅ Passthrough agent detected tool_calls
🔄 Phase 2: Making streaming call (client requested stream)
📤 Response sent (agent used: passthrough)
```

### No Tool Calls - Redirection to Generic Agent

```
📥 Request received (current agent: generic)
🔀 [PASSTHROUGH AGENT] Checking for tool calls...
🔀 Routing to passthrough agent first to check for tool calls
🔍 Phase 1: Fast detection call (non-streaming)
⏭️  Passthrough agent found no tool_calls (finish_reason: stop), redirecting to selected agent
💬 [GENERIC AGENT] Processing request...
📤 Response sent (agent used: generic)
```

### Normal Request Without Tools

```
📥 Request received (current agent: generic)
🔶 Orchestrator processing request...
🔵 Matching agent for topic: programming
💬 [GENERIC AGENT] Processing request...
📤 Response sent (agent used: generic)
```

## Key Implementation Details

### Passthrough-First Logic

The gateway implements a two-phase approach for tool requests:

**Phase 1: Detection (Always Non-Streaming)**
- Fast call to passthrough agent to check if tools are needed
- Analyzes `finish_reason` and `tool_calls` in response
- Cost: 1 API call

**Phase 2: Response (Conditional)**
- If tool_calls detected + streaming requested: Makes a second streaming call
- If tool_calls detected + non-streaming: Uses detection response
- If no tool_calls: Redirects to appropriate agent
- Cost: 0-1 additional API call (only if tool_calls + streaming)

### Validation at Startup

The gateway validates that an agent with ID "passthrough" exists in the crew:

```go
if agent.toolMode == ToolModePassthrough {
    if _, exists := agent.chatAgents["passthrough"]; !exists {
        return nil, fmt.Errorf("passthrough mode requires an agent with ID 'passthrough' in the crew")
    }
}
```

## Important Notes

1. **Passthrough Agent Must Support Tools**: The model used for the passthrough agent MUST support tool calling (e.g., Qwen2.5-Coder, GPT-4, Claude, etc.)

2. **Cost Consideration**: When tool_calls are detected in streaming mode, the gateway makes 2 API calls:
   - 1 non-streaming detection call
   - 1 streaming response call

3. **Agent ID is Important**: The agent must have exactly the ID "passthrough" for the validation to work.

4. **Streaming Support**: The passthrough-first logic fully supports both streaming and non-streaming clients (pi, qwen-code, aider, etc.)

## Troubleshooting

### Gateway fails to start

```
Error: passthrough mode requires an agent with ID 'passthrough' in the crew
```

**Solution**: Ensure you have an agent with ID "passthrough" in your crew:

```go
agentCrew := map[string]*chat.Agent{
    "passthrough": passthroughAgent,  // ← Must have this ID
    "coder": coderAgent,
    "generic": genericAgent,
}
```

### Tools not working

**Solution**: Make sure the passthrough agent uses a tool-capable model:

```bash
export PASSTHROUGHT_MODEL_ID="hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m"
```

### Client hangs in streaming mode

This was a bug in the initial implementation. Make sure you're using the updated version that supports streaming in `handlePassthroughFirst`.

## Related Samples

- `samples/84-gateway-server-agent` - Single agent gateway (basic)
- `samples/85-gateway-server-agent-crew` - Multi-agent gateway (crew)
- `samples/86-gateway-compose` - Gateway with orchestrator
- `samples/90-https-server-example` - HTTPS gateway

## Documentation

- [Gateway Server Agent Guide (English)](../../docs/gateway-server-agent-guide-en.md)
- [Gateway Server Agent Guide (Français)](../../docs/gateway-server-agent-guide-fr.md)

## License

Part of the N.O.V.A. SDK - See repository root for license information.