# 🚀 Quick Start - Gateway Server with qwen-code

Quick guide to use the gateway server with qwen-code and tools.

## 📦 Prerequisites

1. **LLM Server**: A running llama.cpp engine
   ```bash
   # The server must be accessible at http://localhost:12434
   ```

2. **qwen-code**: Install qwen-code
   ```bash
   npm install -g @qwen-code/qwen-code
   ```

## 🎯 Getting Started in 3 Steps

### Step 1: Start the gateway server

```bash
cd samples/85-gateway-server-agent-crew
go run main.go
```

You should see:
```
🚀 Gateway crew server starting on http://localhost:8080
📡 OpenAI-compatible endpoint: POST /v1/chat/completions
👥 Crew agents: coder, thinker, generic
🔧 Tools mode: passthrough (client-side)
```

### Step 2: Configure environment variables

```bash
export OPENAI_BASE_URL=http://localhost:8080/v1
export OPENAI_API_KEY=none
export OPENAI_MODEL=crew
```

### Step 3: Launch qwen-code

```bash
qwen-code
```

That's it! 🎉

## ✅ Quick Test

Once qwen-code is running, test with:

```
You: Write a hello world in Go
```

The gateway should automatically route to the **coder** agent and generate the code.

## 🛠️ Using Tools

Qwen-code automatically handles tools. For example:

```
You: Read the file package.json and tell me the version
```

Qwen-code will:
1. Declare the `read_file` tool to the gateway
2. The LLM decides to use the tool
3. Qwen-code executes the file read
4. The LLM generates the response with the content

**Everything is automatic!** 🎯

## 🔍 Operating Modes

### Current Mode: **Passthrough** (default)

- ✅ Qwen-code handles the tools
- ✅ Client-side execution
- ✅ Maximum security
- ✅ Full flexibility

### Alternative Mode: **Auto-Execute**

To enable Auto-Execute mode (server-side tools), see [README-tools.md](README-tools.md#advanced-configuration).

## 📊 Available Agents

The gateway automatically routes to the right agent:

| Agent | Triggers | Use case |
|-------|----------|----------|
| **coder** | coding, programming, development, code, software | Code, debug, refactoring |
| **thinker** | philosophy, thinking, ideas, psychology, math, science | Reflection, analysis, complex problems |
| **generic** | everything else | General questions |

## 🧪 Testing with curl

If you want to test without qwen-code:

```bash
# Simple test
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "stream": false
  }' | jq .
```

For more advanced examples with tools:

```bash
./examples-tools.sh
```

## 🐛 Troubleshooting

### Error: "connection refused"

**Cause:** The LLM server is not started

**Solution:** Verify that llama.cpp is running on `localhost:12434`

### Error: "model not found"

**Cause:** The models are not downloaded

**Solution:** Verify that the models in `main.go` are available:
- `hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m`
- `hf.co/menlo/lucy-gguf:q4_k_m`
- `hf.co/menlo/jan-nano-gguf:q4_k_m`

### Qwen-code can't find the model

**Cause:** Variable `OPENAI_MODEL` not defined

**Solution:**
```bash
export OPENAI_MODEL=crew
```

### Tools are not working

**Cause:** Qwen-code must be configured to use tools

**Solution:** Check qwen-code configuration for available tools

## 📚 Complete Documentation

- [README-tools.md](README-tools.md) - Complete guide on tools
- [examples-tools.sh](examples-tools.sh) - Practical examples with curl
- [test.sh](test.sh) - Gateway test suite

## 🎨 Customization

### Change the port

Modify in `main.go`:

```go
gatewayserver.WithPort(8080), // Change 8080 to your port
```

### Modify agents

Add, remove or modify agents in the `main()` function:

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "generic": genericAgent,
    // Add your agents here
}
```

### Customize routing

Modify the `matchAgentFunction` function to change routing rules:

```go
matchAgentFunction := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "coding":
        return "coder"
    case "philosophy":
        return "thinker"
    // Add your rules here
    default:
        return "generic"
    }
}
```

## 💡 Usage Tips

### 1. Always specify context

❌ Bad: "Fix this"
✅ Good: "Fix the syntax error in the Go function reverseString"

### 2. Use the right keywords for routing

- For code: "write", "debug", "fix", "code", "function"
- For reflection: "explain", "why", "philosophy", "analyze"
- For general questions: everything else

### 3. Take advantage of qwen-code tools

Qwen-code has access to your local file system, use it!

```
You: Read all .go files in the current directory and find potential bugs
```

## 🌟 Advanced Features

### Automatic Compression

The gateway automatically compresses history when it exceeds 7000 characters:

```go
gatewayserver.WithCompressorAgentAndContextSize(compressorAgent, 7000)
```

### Multi-agent Orchestration

The orchestrator automatically analyzes the topic and routes to the right agent:

```go
gatewayserver.WithOrchestratorAgent(orchestratorAgent)
```

### Lifecycle Hooks

You can add hooks before/after each request:

```go
gatewayserver.BeforeCompletion(func(agent *gatewayserver.GatewayServerAgent) {
    fmt.Printf("📥 Request received\n")
})
```

## 🔗 Useful Links

- [Nova SDK Documentation](../../README.md)
- [Qwen Code GitHub](https://github.com/QwenLM/qwen-code)
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)

---

**Need help?** Check [README-tools.md](README-tools.md) for more details.
