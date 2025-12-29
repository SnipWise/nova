# CLAUDE.md - Nova Demos Project

## ⚠️ IMPORTANT RULES

**This project EXCLUSIVELY uses:**
- **Language**: Go (golang)
- **Framework**: Nova SDK (github.com/snipwise/nova)
- **NOT**: Node.js, Python, JavaScript, TypeScript

**When asked to generate an agent, ALWAYS:**
1. Use the `nova-agent-builder` skill in `.claude/skills/`
2. Generate **Go** code with Nova SDK
3. Follow patterns from available snippets

## About

This project contains demos and templates for the Nova SDK framework (AI agents in Go).

## Project Structure

```
nova-demos/
├── CLAUDE.md                    # This file
└── .claude/
    └── skills/
        └── nova-agent-builder/
            ├── SKILL.md                 # Main skill
            ├── snippets/                # Code templates
            │   ├── chat/                # Chat agents
            │   ├── rag/                 # RAG agents
            │   ├── tools/               # Tools agents
            │   ├── structured/          # Structured output agents
            │   ├── compressor/          # Compressor agents
            │   ├── orchestrator/        # Orchestrator agents (topic detection)
            │   ├── server/              # Server agents (HTTP API)
            │   ├── remote/              # Remote agents (clients)
            │   ├── complex/             # Crew & Pipeline
            │   └── docker/              # Docker & Deployment
            └── references/              # API documentation
```

## Claude Code Commands

### Agent Generation

To generate an agent, simply ask:

```
generate a chat agent with streaming
create a RAG agent for my FAQ system
generate a tools agent to calculate and send emails
```

**Important**: Claude should automatically detect the `nova-agent-builder` skill and generate Go code.

### Complex Agents

For complex agents (crew, pipeline), Claude will ask questions:

```
create a crew agent to write articles
# Claude will ask:
# - Number of agents
# - Role of each agent
# - Interaction type
# - Required tools

create a pipeline to process documents
# Claude will ask:
# - Number of steps
# - Transformations
# - Error handling
```

### If Claude Generates Wrong Code

If Claude generates Node.js or Python instead of Go, specify:

```
generate a chat agent with streaming IN GO with Nova SDK
use the nova-agent-builder skill to create a RAG agent
```

## Available Snippets

| Category | Snippet | Description | Sample |
|----------|---------|-------------|--------|
| chat | streaming-chat | Chat with streaming responses | 05 |
| chat | contextual-chat | Chat with context memory | 06 |
| rag | basic-rag | RAG with in-memory vectorstore | 13 |
| rag | jsonstore-rag | RAG with JSON persistent store | 69 |
| tools | simple-tools | Agent with function calling | 18 |
| tools | parallel-tools | Parallel tool execution | 19 |
| tools | confirmation-tools | Human-in-the-loop | 47 |
| structured | structured-output | Structured output (Go struct) | 23 |
| structured | structured-schema | Output with JSON Schema | 24 |
| structured | structured-validation | Advanced validation + retry | 25 |
| compressor | compressor-agent | Context compression | 28 |
| **server** | **basic-server** | **HTTP/REST API with SSE streaming** | **70** |
| **server** | **server-with-tools** | **API with function calling** | **49** |
| **server** | **server-with-rag** | **API with document retrieval** | **54** |
| **server** | **server-with-compressor** | **API with context compression** | **54** |
| **server** | **server-full-featured** | **Complete API (tools+RAG+compress)** | **54** |
| **remote** | **basic-remote** | **Client connecting to Server Agent** | **71** |
| orchestrator | topic-detection | Topic/intent detection for routing | 55 |
| complex | crew-agent | Multi-agent collaboration (local) | 55 |
| complex | crew-server-agent | HTTP agent server (API) | 56 |
| complex | remote-agent | Client for Crew Server | 51 |
| complex | pipeline-agent | Chained agents | 56 |
| **docker** | **dockerfile-template** | **Multi-stage Dockerfile for agents** | - |
| **docker** | **docker-compose-simple** | **Docker Compose for simple agents** | - |
| **docker** | **docker-compose-complex** | **Docker Compose for crew/pipeline** | - |
| **docker** | **dockerization-guide** | **Complete dockerization guide** | - |

## Docker & Deployment

### Dockerize Agents

To dockerize your agents for production deployment:

```
dockerize my chat agent
create docker compose for my RAG agent
deploy my crew to production
```

**Claude will generate**:
- Multi-stage Dockerfile (optimized, ~15-25 MB)
- docker-compose.yml with Docker Agentic Compose
- .dockerignore file
- Environment-variable-based configuration
- Build and run instructions

### Docker Agentic Compose

**What is it?** Docker Compose extension (v2.38.0+) that manages AI models as first-class resources.

**Key features**:
- Declarative model management in `compose.yml`
- Automatic environment variable injection (`ENGINE_URL`, `MODEL_ID`)
- Model lifecycle management via Docker Model Runner
- Multi-model orchestration

**Example**:
```yaml
services:
  chat-agent:
    models:
      chat-model:
        endpoint_var: ENGINE_URL      # Auto-injected
        model_var: CHAT_MODEL_ID      # Auto-injected

models:
  chat-model:
    model: ai/qwen2.5:1.5B-F16
```

### Available Docker Templates

| Template | Use Case | Command |
|----------|----------|---------|
| **Dockerfile** | Any agent | `docker build -t my-agent .` |
| **Simple Compose** | Chat/RAG/Tools/Server | `docker compose up -d` |
| **Complex Compose** | Crew/Pipeline/Orchestrator | `docker compose up -d` |
| **Complete Guide** | Step-by-step tutorial | Read `docker/dockerization-guide.md` |

### Deployment Strategies

- **Local**: Docker Compose (development, demos)
- **Multi-host**: Docker Swarm (small-medium production)
- **Cloud**: Kubernetes, AWS ECS, Azure ACI, GCP Cloud Run

### Prerequisites

- Docker Desktop 4.36+ (includes Docker Compose 2.38+)
- Docker Model Runner (included in Docker Desktop)
- Pre-downloaded models (optional, auto-pulled on first run)

```bash
# Verify installation
docker --version
docker compose version
docker model list

# Pull models
docker model pull ai/qwen2.5:1.5B-F16
docker model pull ai/mxbai-embed-large
```

## Orchestrator Agent

The **orchestrator agent** is a specialized agent for **topic detection and query routing** in multi-agent systems:

```
generate an orchestrator agent for topic detection
create an orchestrator to route queries to specialized agents
```

**Key Features:**
- Detects topics/intents from user input
- Routes queries to appropriate specialized agents
- Integrates seamlessly with crew agents
- Fast classification with low latency
- Uses `agents.Intent` for standardized output

**Usage in Crew:**
```go
orchestratorAgent, _ := orchestrator.NewAgent(ctx, agentConfig, modelConfig)
crewAgent.SetOrchestratorAgent(orchestratorAgent)
// Crew now auto-routes queries based on detected topics
```

See `orchestrator/topic-detection` snippet for complete examples.

## Default Configuration

```yaml
# config.yaml or .env
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
CHAT_MODEL: "ai/qwen2.5:1.5B-F16"
TOOLS_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
ORCHESTRATOR_MODEL: "hf.co/menlo/lucy-gguf:q4_k_m"
EMBEDDING_MODEL: "ai/mxbai-embed-large"
```

## Development Workflow

1. **Ask**: Describe the desired agent
2. **Answer**: Claude's questions (for complex agents)
3. **Receive**: Generated code based on snippets
4. **Adapt**: Customize to your needs
5. **Test**: Run and iterate

## Important Notes

- Snippets are in `.claude/skills/nova-agent-builder/snippets/`
- API documentation is in `.claude/skills/nova-agent-builder/references/`
- Always check model compatibility with features (tools, streaming, etc.)
