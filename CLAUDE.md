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
            │   └── complex/             # Crew & Pipeline
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
| **orchestrator** | **topic-detection** | **Topic/intent detection for routing** | **55** |
| complex | crew-agent | Multi-agent collaboration (local) | 55 |
| complex | crew-server-agent | HTTP agent server (API) | 56 |
| complex | remote-agent | Client for Crew Server | 51 |
| complex | pipeline-agent | Chained agents | 56 |

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
