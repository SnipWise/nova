---
name: nova-agent-builder
description: |
  ALWAYS use this skill when the user asks to generate AI agent code in Go.
  This skill generates Go code with the Nova SDK framework (github.com/snipwise/nova).
  
  MANDATORY TRIGGERS - Use this skill if the request contains:
  - "agent" + "Go" or "golang"
  - "Nova SDK" or "Nova" + "agent"
  - "chat agent", "chatbot", "conversational" + Go
  - "RAG agent", "retrieval", "embeddings", "vector" + Go
  - "tools agent", "function calling" + Go
  - "server agent", "HTTP API", "REST API", "API server" + Go
  - "server with tools", "API with functions"
  - "server with RAG", "API with knowledge base"
  - "remote agent", "client agent", "connect to server"
  - "crew agent", "multi-agents", "agent team"
  - "pipeline agent", "chained agents"
  - "structured output", "JSON schema" + agent
  - "compressor", "context compression"
  - "remote agent", "crew server"
  - "streaming" + agent + Go
  
  DO NOT use if: Node.js, Python, JavaScript, TypeScript explicitly requested.
  
  This project uses Nova SDK - a Go framework for building local AI agents.
---

# Nova Agent Builder Skill

Specialized skill for generating AI agent code with the Nova SDK framework in Go.

## Version Requirements

**CRITICAL**: Always use these versions when generating code:

- **Go**: `1.25.4` (minimum)
- **Nova SDK**: Latest release from `github.com/snipwise/nova`

### go.mod Template

When generating new projects, always include:

```go
module your-project-name

go 1.25.4

require (
    github.com/snipwise/nova latest
)
```

**Installation command**:
```bash
go get github.com/snipwise/nova@latest
go mod tidy
```

## Nova SDK Architecture

Nova SDK uses a modular architecture:

```
nova-sdk/
├── agents/
│   ├── chat/      # Conversational agents
│   ├── rag/       # RAG agents
│   └── tools/     # Function calling agents
├── messages/      # Message handling
├── models/        # Model configuration
└── prompt/        # Prompting utilities
```

## Generation Workflow

### 1. Identify Agent Type

| User Request | Type | Snippet to Use |
|--------------|------|----------------|
| "chat agent", "chatbot", "conversation" | Chat | `snippets/chat/` |
| "streaming", "real-time" | Chat Streaming | `snippets/chat/streaming-chat.md` |
| "RAG", "search", "embeddings", "vector" | RAG | `snippets/rag/` |
| "persistent RAG", "json store", "save embeddings" | RAG Persistant | `snippets/rag/jsonstore-rag.md` |
| "tools", "function calling", "utilities" | Tools | `snippets/tools/` |
| "structured", "JSON", "schema", "extraction" | Structured | `snippets/structured/` |
| "compression", "long context", "long memory" | Compressor | `snippets/compressor/` |
| "server agent", "HTTP API", "REST", "SSE" | Server Agent | `snippets/server/basic-server.md` |
| "server with tools", "API with functions" | Server + Tools | `snippets/server/server-with-tools.md` |
| "server with RAG", "API with knowledge base" | Server + RAG | `snippets/server/server-with-rag.md` |
| "server with compression", "API with long context" | Server + Compressor | `snippets/server/server-with-compressor.md` |
| "full server", "complete API", "production server" | Full Server | `snippets/server/server-full-featured.md` |
| "remote agent", "client agent", "connect to server" | Remote Agent | `snippets/remote/basic-remote.md` |
| "multi-agents", "crew", "team" | Crew | `snippets/complex/crew-agent.md` |
| "crew server", "multi-agent API", "expose crew" | Crew Server | `snippets/complex/crew-server-agent.md` |
| "remote", "client", "distant", "connect server" | Remote Agent | `snippets/complex/remote-agent.md` |
| "pipeline", "chained", "sequence" | Pipeline | `snippets/complex/pipeline-agent.md` |

### 2. Load Appropriate Snippet

Read the corresponding snippet in `snippets/` and adapt to needs.

### 3. Customize Code

Adapt according to user parameters:
- Agent name
- System instructions
- LLM engine URL
- Model to use
- Specific tools (for tools agent)

## Available Snippets

### Simple Agents

Files in `snippets/` - direct generation without questions:

**Chat Agents:** (`snippets/chat/`)
- `streaming-chat.md` - Chat with streaming (sample 05)
- `contextual-chat.md` - Chat with maintained context (sample 06)

**RAG Agents:** (`snippets/rag/`)
- `basic-rag.md` - Basic RAG agent with in-memory vectorstore (sample 13)
- `jsonstore-rag.md` - RAG with JSON persistence for embeddings (sample 69)
- **Default embedding model**: `ai/mxbai-embed-large`

**Tools Agents:** (`snippets/tools/`)
- `simple-tools.md` - Agent with simple tools (sample 18)
- `parallel-tools.md` - Parallel tools (sample 19)
- `confirmation-tools.md` - Tools with human-in-the-loop confirmation (sample 47)

**Structured Output Agents:** (`snippets/structured/`)
- `structured-output.md` - Structured output with Go struct (sample 23)
- `structured-schema.md` - Output with explicit JSON Schema (sample 24)
- `structured-validation.md` - Advanced validation and retry (sample 25)

**Compressor Agents:** (`snippets/compressor/`)
- `compressor-agent.md` - Context compression (sample 28)

**Server Agents:** (`snippets/server/`)
- `basic-server.md` - HTTP/REST API server with streaming (sample 70)
- `server-with-tools.md` - Server with function calling and validation (sample 49)
- `server-with-rag.md` - Server with document retrieval (sample 54)
- `server-with-compressor.md` - Server with context compression (sample 54)
- `server-full-featured.md` - Complete server with all features (sample 54)

**Remote Agents:** (`snippets/remote/`)
- `basic-remote.md` - Remote client connecting to Server Agent (sample 71)

### Complex Agents (Interactive Mode)

For these agents, ask questions before generating:

#### Crew Agent (`snippets/complex/crew-agent.md`) - Sample 55

Local collaborative multi-agents.

Questions to ask:
1. "How many agents in your crew?"
2. "What role for each agent?"
3. "What interactions between agents?"
4. "Hierarchy between agents?"
5. "Tools per agent?"

#### Crew Server Agent (`snippets/complex/crew-server-agent.md`) - Sample 56

HTTP server exposing an agent crew via REST API.

Questions to ask:
1. "Which agents to expose (chat, rag, tools)?"
2. "Which API (REST, WebSocket, gRPC)?"
3. "Authentication required?"
4. "Which endpoints?"
5. "Rate limiting / scaling?"

#### Remote Agent (`snippets/complex/remote-agent.md`) - Sample 51

Client to connect to a remote Crew Server.

Questions to ask:
1. "Crew Server URL?"
2. "Which endpoints to use?"
3. "Connection mode (HTTP, WebSocket)?"
4. "Error handling (retry, fallback)?"
5. "Response caching?"

#### Pipeline Agent (`snippets/complex/pipeline-agent.md`) - Sample 56

Chained agents with transformations.

Questions to ask:
1. "How many steps?"
2. "Transformation at each step?"
3. "Branching conditions?"
4. "Error handling?"
5. "Parallelization?"

## Nova Code Patterns

### Base Pattern - Chat Agent

```go
agent, err := chat.NewAgent(
    ctx,
    agents.Config{
        Name:               "agent-name",
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "System instructions...",
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
        MaxTokens:   models.Int(2000),
    },
)
```

### Tools Pattern

```go
tool := tools.NewTool("tool_name").
    SetDescription("Description").
    AddParameter("param", "type", "description", true)
```

### RAG Pattern

**Default embedding model**: `ai/mxbai-embed-large`

```go
agent, err := rag.NewAgent(ctx, agentConfig, models.Config{
    Name: "ai/mxbai-embed-large",  // Default embedding model
})
agent.SaveEmbedding("text chunk")
similarities, _ := agent.SearchSimilar(query, 0.6)
```

### RAG with Persistence Pattern

```go
// Check and load existing store
if agent.StoreFileExists("./store/data.json") {
    agent.LoadStore("./store/data.json")
} else {
    // Index documents and save
    agent.SaveEmbedding("text chunk")
    agent.PersistStore("./store/data.json")
}
```

## Generation Rules

1. **Always include** necessary imports
2. **Always handle** errors with `if err != nil`
3. **Always use** `context.Background()` or appropriate context
4. **Comment** customizable parts
5. **Provide** functional default values

## CRITICAL: Conversation History

**DO NOT manually manage conversation history** - agents handle this automatically:

### ✅ CORRECT Pattern
```go
// Just pass the current user message
agent.GenerateCompletion(
    []messages.Message{{Role: roles.User, Content: userInput}},
)
// Agent automatically maintains full conversation history
```

### ❌ WRONG Pattern (Never Generate This)
```go
// DON'T create manual history arrays
var conversationHistory []messages.Message
conversationHistory = append(conversationHistory, userMessage)
agent.GenerateCompletion(conversationHistory)
conversationHistory = append(conversationHistory, assistantMessage)
```

### History Management Methods
```go
agent.GetMessages()           // Retrieve history
agent.ResetMessages()         // Clear history
agent.GetContextSize()        // Check size
agent.ExportMessagesToJSON()  // Export to JSON
```

**This applies to**: Chat agents, Tools agents, RAG agents, Structured agents, Server agents, Remote agents.

## Recommended Configuration

```yaml
# .env or config
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
# Alternatives:
# - Ollama: "http://localhost:11434/v1"
# - LM Studio: "http://localhost:1234/v1"

# Recommended models by usage:
CHAT_MODEL: "ai/qwen2.5:1.5B-F16"
EMBEDDING_MODEL: "ai/mxbai-embed-large"  # Default for RAG agents
TOOLS_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
```

## References

For more details, see:
- `references/nova-api.md` - Complete Nova SDK API
- `references/examples-catalog.md` - Examples catalog
