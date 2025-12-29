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
  - "docker", "dockerize", "containerize" + agent
  - "docker compose", "compose", "deployment" + agent
  - "production", "deploy", "cloud" + agent
  
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

**CRITICAL for Docker/Production**:
- When dockerizing agents, ALWAYS use Go 1.25.4 in both go.mod AND Dockerfile
- ALWAYS use `github.com/snipwise/nova latest` (not local replace directives)
- Example: `FROM golang:1.25.5-alpine` in Dockerfile
- See sample 67 for reference implementation

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
| "docker", "dockerfile", "containerize" | Docker Template | `snippets/docker/dockerfile-template.md` |
| "docker compose", "compose", "simple agent" + docker | Docker Compose Simple | `snippets/docker/docker-compose-simple.md` |
| "docker compose", "crew", "pipeline" + docker | Docker Compose Complex | `snippets/docker/docker-compose-complex.md` |
| "dockerization", "deploy", "production", "guide" | Dockerization Guide | `snippets/docker/dockerization-guide.md` |

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
- **Best practice**: Use `compressor.Instructions.Effective` + `compressor.Prompts.UltraShort`

**Server Agents:** (`snippets/server/`)
- `basic-server.md` - HTTP/REST API server with streaming (sample 70)
- `server-with-tools.md` - Server with function calling and validation (sample 49)
- `server-with-rag.md` - Server with document retrieval (sample 54)
- `server-with-compressor.md` - Server with context compression (sample 54)
- `server-full-featured.md` - Complete server with all features (sample 54)

**Remote Agents:** (`snippets/remote/`)
- `basic-remote.md` - Remote client connecting to Server Agent (sample 71)

**Docker & Deployment:** (`snippets/docker/`)
- `dockerfile-template.md` - Multi-stage Dockerfile for Nova agents
- `docker-compose-simple.md` - Docker Compose for simple agents (chat, RAG, tools, server)
- `docker-compose-complex.md` - Docker Compose for complex agents (crew, pipeline, orchestrator)
- `dockerization-guide.md` - Complete step-by-step dockerization guide
- **CRITICAL**: When dockerizing, ALWAYS use Go 1.25.4 and Nova latest version in go.mod
- **Reference**: See sample 67 for complete dockerized chat agent example

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

## Docker & Deployment

### When to Dockerize

Use Docker deployment when:
- **Production deployment**: Deploying to cloud or on-premises servers
- **Multi-environment**: Need same config across dev, staging, production
- **Scalability**: Need to run multiple agent instances
- **Isolation**: Want containerized, isolated execution
- **Team collaboration**: Standardized deployment across team
- **CI/CD**: Automated build and deployment pipelines

### Dockerization Workflow

When user requests dockerization ("dockerize my agent", "create docker compose", "deploy to production"):

1. **Identify agent type**: Simple (chat/RAG/tools/server) or Complex (crew/pipeline)
2. **Choose appropriate snippet**:
   - Simple agents → `docker-compose-simple.md`
   - Complex agents → `docker-compose-complex.md`
   - Need guide → `dockerization-guide.md`
3. **Generate Dockerfile**: Use `dockerfile-template.md` pattern
4. **Generate compose.yml**: Based on agent complexity
5. **Adapt code**: Replace hard-coded values with environment variables
6. **Provide instructions**: Build, run, test commands

### Docker Snippets Usage

#### Dockerfile Template (`dockerfile-template.md`)

**Use when**:
- User asks: "create dockerfile", "containerize agent"
- Generating any dockerized agent

**Features**:
- Multi-stage build (builder + runtime)
- Alpine-based minimal image (~15-25 MB)
- Security best practices
- Customization examples for all agent types

#### Docker Compose Simple (`docker-compose-simple.md`)

**Use when**:
- User asks: "docker compose for chat agent", "deploy RAG agent"
- Simple agents: chat, RAG, tools, structured, server
- Single or multiple instances of simple agents

**Templates provided**:
1. Single chat agent
2. RAG agent with persistence
3. Server agent (HTTP API)
4. Multiple agents in parallel

#### Docker Compose Complex (`docker-compose-complex.md`)

**Use when**:
- User asks: "dockerize crew", "deploy multi-agent system"
- Complex agents: crew, pipeline, orchestrator
- Multi-service deployments

**Templates provided**:
1. Crew agent with orchestrator
2. Pipeline agent
3. Multi-service crew (distributed)
4. RAG + Crew + Compressor (full-featured)

#### Dockerization Guide (`dockerization-guide.md`)

**Use when**:
- User asks: "how to dockerize", "complete docker guide"
- User needs step-by-step tutorial
- First-time dockerization

**Content**:
- Prerequisites and installation
- Step-by-step tutorial (chat agent example)
- Code adaptation patterns
- Deployment strategies
- Production checklist
- Troubleshooting

### Docker Agentic Compose Key Concepts

**Docker Agentic Compose** (v2.38.0+) manages AI models as resources:

```yaml
# Declare models globally
models:
  chat-model:
    model: ai/qwen2.5:1.5B-F16

# Reference in services
services:
  my-agent:
    models:
      chat-model:
        endpoint_var: ENGINE_URL      # Auto-injected
        model_var: CHAT_MODEL_ID      # Auto-injected
```

**Environment variables automatically injected**:
- `ENGINE_URL`: LLM endpoint (e.g., `http://host.docker.internal:12434/engines/llama.cpp/v1`)
- `CHAT_MODEL_ID`: Model name (e.g., `ai/qwen2.5:1.5B-F16`)

### Code Adaptation for Docker

**Pattern**: Hard-coded → Environment Variables

**Before**:
```go
engineURL := "http://localhost:12434/engines/llama.cpp/v1"
modelName := "ai/qwen2.5:1.5B-F16"
```

**After**:
```go
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

engineURL := getEnv("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
modelName := getEnv("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
```

**Alternative (using Nova SDK helper)**:
```go
import "github.com/snipwise/nova/nova-sdk/toolbox/env"

engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
modelName := env.GetEnvOrDefault("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
```

### Docker Generation Rules

When generating Docker-related code:

1. **Always include**:
   - Multi-stage Dockerfile
   - docker-compose.yml with Docker Agentic Compose syntax
   - .dockerignore file
   - Environment variable adaptation in Go code
   - Usage instructions (build, run, test)

2. **Always use**:
   - `env.GetEnvOrDefault()` for configuration
   - YAML anchors for reusable compose configs
   - Specific model tags (avoid `:latest` in production)
   - Volume mounts for persistent data

3. **Always document**:
   - Prerequisites (Docker Desktop, models)
   - Build commands
   - Run commands
   - Common troubleshooting

4. **Deployment strategies**:
   - Local: Docker Compose
   - Multi-host: Docker Swarm
   - Cloud: Kubernetes, AWS ECS, Azure ACI, GCP Cloud Run

### Example: Dockerize Chat Agent Request

**User**: "Dockerize my chat agent"

**Response workflow**:
1. Read `dockerfile-template.md`
2. Read `docker-compose-simple.md` (Template 1: Single Chat Agent)
3. Generate adapted main.go with environment variables
4. Generate Dockerfile
5. Generate compose.yml
6. Generate .dockerignore
7. Provide usage instructions

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

### Compressor Pattern

**RECOMMENDED**: Use built-in instructions and prompts for optimal compression.

```go
import "github.com/snipwise/nova/nova-sdk/agents/compressor"

agent, err := compressor.NewAgent(
    ctx,
    agents.Config{
        Name:      "compressor",
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        // BEST PRACTICE: Use Effective for balanced compression
        SystemInstructions: compressor.Instructions.Effective,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),  // Always use 0.0 for deterministic compression
    },
    // BEST PRACTICE: Use UltraShort for maximum token reduction
    compressor.WithCompressionPrompt(compressor.Prompts.UltraShort),
)

// Available presets:
// Instructions: Expert, Effective (RECOMMENDED), Basic
// Prompts: UltraShort (RECOMMENDED), Minimalist, Balanced, Detailed
```

## Generation Rules

1. **Always include** necessary imports
2. **Always handle** errors with `if err != nil`
3. **Always use** `context.Background()` or appropriate context
4. **Comment** customizable parts
5. **Provide** functional default values
6. **Document** the agent with comprehensive comments
7. **Include** usage examples in code comments
8. **Add** best practices as inline documentation

## Code Documentation Standards

**MANDATORY**: Every generated agent MUST include:

### 1. File Header with Overview
```go
/*
Package main implements a [TYPE] agent using Nova SDK.

Agent Purpose:
  - [Brief description of what this agent does]
  - [Key capabilities]

Architecture:
  - Type: [Chat|RAG|Tools|Server|etc.]
  - Model: [Model name]
  - Features: [Streaming|Persistence|Tools|etc.]

Dependencies:
  - Nova SDK: github.com/snipwise/nova
  - Go version: 1.25.4+
*/
package main
```

### 2. Usage Examples Before main()
```go
/*
USAGE EXAMPLES:

Basic Usage:
  go run main.go

Example Interaction:
  User: "Hello, how are you?"
  Agent: "I'm doing well, thank you! How can I assist you today?"

Configuration:
  - Modify ENGINE_URL to point to your LLM server
  - Update MODEL_NAME for different models
  - Adjust TEMPERATURE (0.0-1.0) for creativity
  - Change MAX_TOKENS for response length

Best Practices:
  1. [Practice 1]
  2. [Practice 2]
  3. [Practice 3]

Common Pitfalls:
  ❌ Don't manually manage conversation history
  ✅ Let the agent handle history automatically

  ❌ Don't ignore error handling
  ✅ Always check err != nil

Production Considerations:
  - Add graceful shutdown handling
  - Implement request timeouts
  - Add logging for debugging
  - Monitor token usage
*/
```

### 3. Configuration Comments
```go
const (
    // LLM Engine URL - supports multiple backends:
    // - Local LLaMA.cpp: "http://localhost:12434/engines/llama.cpp/v1"
    // - Ollama: "http://localhost:11434/v1"
    // - LM Studio: "http://localhost:1234/v1"
    ENGINE_URL = "http://localhost:12434/engines/llama.cpp/v1"

    // Model selection impacts performance and capabilities
    // - Small/Fast: "ai/qwen2.5:1.5B-F16"
    // - Balanced: "hf.co/menlo/jan-nano-gguf:q4_k_m"
    // - Tools: Models supporting function calling
    MODEL_NAME = "ai/qwen2.5:1.5B-F16"
)
```

### 4. Agent-Specific Best Practices

#### For Chat Agents:
```go
/*
CHAT AGENT BEST PRACTICES:

✅ DO:
  - Use clear, specific system instructions
  - Set temperature=0.0 for deterministic responses
  - Test with various conversation lengths
  - Use streaming for better UX

❌ DON'T:
  - Manually append to conversation history
  - Ignore context window limits
  - Use very high temperatures for factual tasks

Memory Management:
  agent.GetMessages()      // View conversation history
  agent.ResetMessages()    // Clear history
  agent.GetContextSize()   // Check token usage

Example System Instructions:
  "You are a helpful assistant that provides concise, accurate answers."
  "You are a coding expert specializing in Go programming."
*/
```

#### For RAG Agents:
```go
/*
RAG AGENT BEST PRACTICES:

✅ DO:
  - Use appropriate similarity thresholds (0.5-0.8)
  - Chunk documents into 200-500 token pieces
  - Use descriptive metadata for filtering
  - Persist store for production use

❌ DON'T:
  - Index entire documents as single chunks
  - Use very low similarity thresholds (<0.3)
  - Forget to handle empty search results

Document Indexing:
  // Good chunk size
  agent.SaveEmbedding("Focused 200-500 token chunk")

  // Too large (avoid)
  agent.SaveEmbedding("Entire 10,000 word document...")

Similarity Search:
  results, _ := agent.SearchSimilar(query, 0.6)
  if len(results) == 0 {
      // Handle no results case
  }

Persistence:
  // Save store after indexing
  agent.PersistStore("./store/data.json")

  // Load on startup
  if agent.StoreFileExists("./store/data.json") {
      agent.LoadStore("./store/data.json")
  }
*/
```

#### For Compressor Agents:
```go
/*
COMPRESSOR AGENT BEST PRACTICES:

✅ DO:
  - Use compressor.Instructions.Effective for balanced compression
  - Use compressor.Prompts.UltraShort for maximum token reduction
  - Set temperature=0.0 for deterministic, consistent compression
  - Keep recent messages uncompressed (preserve conversation flow)
  - Test compression quality with your use case
  - Monitor context size before/after compression

❌ DON'T:
  - Use custom instructions when built-in presets work well
  - Set temperature > 0.0 (causes unpredictable compression)
  - Compress too aggressively (may lose important context)
  - Forget to preserve recent conversation turns
  - Skip validation of compression results

Agent Creation Pattern:
  compressorAgent, err := compressor.NewAgent(
      ctx,
      agents.Config{
          Name:      "compressor",
          EngineURL: engineURL,
          // RECOMMENDED: Use built-in instructions
          SystemInstructions: compressor.Instructions.Effective,
      },
      models.Config{
          Name:        modelName,
          Temperature: models.Float64(0.0),  // CRITICAL: Always 0.0
      },
      // RECOMMENDED: Use built-in prompts
      compressor.WithCompressionPrompt(compressor.Prompts.UltraShort),
  )

Available Presets:
  // Instructions (SystemInstructions)
  compressor.Instructions.Expert      // Most sophisticated
  compressor.Instructions.Effective   // Balanced (RECOMMENDED)
  compressor.Instructions.Basic       // Simple

  // Prompts (WithCompressionPrompt)
  compressor.Prompts.UltraShort   // Maximum reduction (RECOMMENDED)
  compressor.Prompts.Minimalist   // Very concise
  compressor.Prompts.Balanced     // Balance detail/brevity
  compressor.Prompts.Detailed     // Preserve more info

Compression Methods:
  // Compress text directly
  summary, err := agent.Compress(textToCompress)

  // Compress conversation messages with streaming
  result, err := agent.CompressContextStream(
      messages,
      func(partial string, reason string) error {
          fmt.Print(partial)
          return nil
      },
  )

Context Management:
  // Check if compression is needed
  if chatAgent.GetContextSize() > threshold {
      compressed, _ := compressorAgent.CompressContextStream(
          chatAgent.GetMessages(),
          streamCallback,
      )
      chatAgent.ResetMessages()
      chatAgent.AddMessage(roles.System, compressed.CompressedText)
  }
*/
```

#### For Tools Agents:
```go
/*
TOOLS AGENT BEST PRACTICES:

✅ DO:
  - Provide clear tool descriptions
  - Use descriptive parameter names
  - Validate inputs in tool functions
  - Return structured error messages
  - Use required=true for mandatory params

❌ DON'T:
  - Create tools with ambiguous names
  - Skip input validation
  - Use complex nested parameters
  - Ignore tool execution errors

Tool Definition Pattern:
  tool := tools.NewTool("tool_name").
      SetDescription("Clear, action-oriented description").
      AddParameter("param1", "string", "What this parameter does", true).
      AddParameter("param2", "integer", "Optional parameter", false).
      SetFunction(func(args map[string]interface{}) (string, error) {
          // Validate inputs
          param1, ok := args["param1"].(string)
          if !ok || param1 == "" {
              return "", errors.New("param1 is required")
          }

          // Perform action
          result := performAction(param1)

          // Return structured response
          return fmt.Sprintf("Result: %v", result), nil
      })

Example Tool Names (Action-Oriented):
  ✅ "calculate_sum", "send_email", "get_weather"
  ❌ "calculator", "email", "weather"
*/
```

#### For Server Agents:
```go
/*
SERVER AGENT BEST PRACTICES:

✅ DO:
  - Implement proper CORS for web clients
  - Use SSE for streaming responses
  - Add request validation
  - Implement rate limiting for production
  - Log requests for debugging

❌ DON'T:
  - Expose without authentication
  - Skip error response handling
  - Use polling instead of SSE
  - Ignore timeout configurations

Endpoints:
  POST /chat          - Single completion (JSON response)
  POST /chat/stream   - Streaming completion (SSE)
  GET  /health        - Health check
  POST /reset         - Clear conversation history

Request Format:
  {
    "message": "User message here",
    "temperature": 0.7,  // Optional
    "max_tokens": 1000   // Optional
  }

Response Format:
  {
    "response": "Agent response",
    "usage": {
      "prompt_tokens": 123,
      "completion_tokens": 456
    }
  }

Testing:
  curl -X POST http://localhost:8080/chat \
    -H "Content-Type: application/json" \
    -d '{"message": "Hello"}'
*/
```

### 5. Inline Code Comments

**For Configuration:**
```go
agent, err := chat.NewAgent(
    ctx,
    agents.Config{
        Name: "helpful-assistant",  // Agent identifier for logging
        EngineURL: ENGINE_URL,      // LLM backend server
        SystemInstructions: "...",  // Defines agent behavior/persona
    },
    models.Config{
        Name: MODEL_NAME,           // Must be available on LLM server
        Temperature: models.Float64(0.0),  // 0.0=deterministic, 1.0=creative
        MaxTokens: models.Int(2000),       // Response length limit
    },
)
```

**For Critical Operations:**
```go
// Generate completion - conversation history managed automatically
// No need to manually track messages between calls
response, err := agent.GenerateCompletion(
    []messages.Message{{Role: roles.User, Content: userInput}},
)
if err != nil {
    // Production: log error and return user-friendly message
    log.Printf("Error generating completion: %v", err)
    return
}
```

### 6. Template Structure

Every generated agent should follow this structure:

```go
/*
[FILE HEADER - Agent overview]
*/
package main

/*
[USAGE EXAMPLES - How to use this agent]
*/

import (...)

const (
    // [CONFIGURATION with explanatory comments]
)

/*
[AGENT-SPECIFIC BEST PRACTICES]
*/

func main() {
    // Step 1: Initialize context
    // [Comment explaining why]

    // Step 2: Configure and create agent
    // [Comment explaining configuration choices]

    // Step 3: Main interaction loop
    // [Comment explaining flow]

    // Step 4: Handle responses
    // [Comment explaining best practices]
}
```

## Example: Fully Documented Agent

See below for a complete example incorporating all documentation standards:

```go
/*
Package main implements a conversational chat agent using Nova SDK.

Agent Purpose:
  - Provides interactive chat capabilities with streaming responses
  - Maintains conversation context automatically
  - Supports customizable personas via system instructions

Architecture:
  - Type: Chat Agent with Streaming
  - Model: Qwen 2.5 1.5B (lightweight, fast)
  - Features: Streaming, automatic history management

Dependencies:
  - Nova SDK: github.com/snipwise/nova
  - Go version: 1.25.4+
*/
package main

/*
USAGE EXAMPLES:

Basic Usage:
  go run main.go

Example Interaction:
  You: "What is the capital of France?"
  Agent: "The capital of France is Paris."
  You: "What is its population?"
  Agent: "Paris has approximately 2.2 million inhabitants in the city proper..."

Configuration:
  - Modify ENGINE_URL to point to your LLM server (Ollama, LM Studio, etc.)
  - Update MODEL_NAME for different models (ensure model supports chat)
  - Adjust TEMPERATURE: 0.0 for factual, 0.7 for creative
  - Change MAX_TOKENS to control response length

Best Practices:
  1. Use streaming for better user experience
  2. Set temperature=0.0 for factual/deterministic responses
  3. Monitor context size with agent.GetContextSize()
  4. Clear history periodically with agent.ResetMessages()

Common Pitfalls:
  ❌ Don't manually manage conversation history
  ✅ Let the agent handle history automatically

  ❌ Don't use very high MaxTokens (causes delays)
  ✅ Use 1000-2000 for most use cases

  ❌ Don't ignore streaming errors
  ✅ Always check chunkChan for errors

Production Considerations:
  - Add request timeouts to prevent hanging
  - Implement graceful shutdown (handle SIGINT/SIGTERM)
  - Add structured logging (not fmt.Printf)
  - Monitor token usage to prevent excessive costs
  - Consider adding rate limiting for multi-user scenarios
*/

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

const (
    // LLM Engine URL - supports multiple backends:
    // - Local LLaMA.cpp: "http://localhost:12434/engines/llama.cpp/v1"
    // - Ollama: "http://localhost:11434/v1"
    // - LM Studio: "http://localhost:1234/v1"
    ENGINE_URL = "http://localhost:12434/engines/llama.cpp/v1"

    // Model selection impacts performance and capabilities
    // Small/Fast: ai/qwen2.5:1.5B-F16 (good for most tasks)
    // Medium: hf.co/menlo/jan-nano-gguf:q4_k_m (better quality)
    MODEL_NAME = "ai/qwen2.5:1.5B-F16"

    // Temperature: 0.0=deterministic, 1.0=very creative
    TEMPERATURE = 0.0

    // Maximum tokens in response (1 token ≈ 0.75 words)
    MAX_TOKENS = 2000
)

func main() {
    // Step 1: Initialize context
    // Context allows for cancellation and timeout management
    ctx := context.Background()

    // Step 2: Configure and create agent
    // Agent automatically manages conversation history
    agent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name: "helpful-assistant",  // Identifier for logging/debugging
            EngineURL: ENGINE_URL,      // LLM backend server endpoint
            SystemInstructions: "You are a helpful and concise assistant.",
        },
        models.Config{
            Name: MODEL_NAME,                      // Must exist on LLM server
            Temperature: models.Float64(TEMPERATURE),  // Controls randomness
            MaxTokens: models.Int(MAX_TOKENS),        // Limits response length
        },
    )
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    fmt.Println("💬 Chat Agent Ready (type 'exit' to quit)")
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

    // Step 3: Main interaction loop
    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("\nYou: ")
        if !scanner.Scan() {
            break
        }

        userInput := strings.TrimSpace(scanner.Text())
        if userInput == "" {
            continue
        }
        if userInput == "exit" {
            fmt.Println("👋 Goodbye!")
            break
        }

        // Step 4: Generate streaming response
        // Agent automatically appends user message to history
        fmt.Print("Agent: ")

        // GenerateStreamingCompletion returns a channel for real-time output
        chunkChan, err := agent.GenerateStreamingCompletion(
            []messages.Message{{Role: roles.User, Content: userInput}},
        )
        if err != nil {
            log.Printf("Error: %v\n", err)
            continue
        }

        // Stream and display response chunks as they arrive
        fullResponse := ""
        for chunk := range chunkChan {
            if chunk.Err != nil {
                log.Printf("\nStreaming error: %v", chunk.Err)
                break
            }
            fmt.Print(chunk.Content)
            fullResponse += chunk.Content
        }
        fmt.Println() // New line after complete response

        // Note: Agent automatically added assistant response to history
        // No manual history management needed!
    }

    // Optional: Display conversation statistics
    contextSize := agent.GetContextSize()
    fmt.Printf("\n📊 Context used: %d tokens\n", contextSize)
}
```

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
