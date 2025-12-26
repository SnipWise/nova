# Nova SDK Examples Catalog

Reference of available snippets and sample mappings.

## Available Snippets

### 💬 Chat Agents (`snippets/chat/`)

| ID | Name | Complexity | Sample Source | Mode |
|----|------|------------|---------------|------|
| streaming-chat | Streaming Chat | Beginner | 05 | Direct |
| contextual-chat | Contextual Chat | Beginner | 06 | Direct |

### 🔍 RAG Agents (`snippets/rag/`)

| ID | Name | Complexity | Sample Source | Mode |
|----|------|------------|---------------|------|
| basic-rag | Basic RAG (in-memory) | Intermediate | 13 | Direct |
| jsonstore-rag | RAG with JSON Store | Intermediate | 69 | Direct |

### 🔧 Tools Agents (`snippets/tools/`)

| ID | Name | Complexity | Sample Source | Mode |
|----|------|------------|---------------|------|
| simple-tools | Simple Tools | Intermediate | 18 | Direct |
| parallel-tools | Parallel Tools | Intermediate | 19 | Direct |
| confirmation-tools | Tools + Confirmation | Advanced | 47 | Direct |

### 📋 Structured Output Agents (`snippets/structured/`)

| ID | Name | Complexity | Sample Source | Mode |
|----|------|------------|---------------|------|
| structured-output | Structured (Go Struct) | Intermediate | 23 | Direct |
| structured-schema | Structured (JSON Schema) | Intermediate | 24 | Direct |
| structured-validation | Structured + Validation | Advanced | 25 | Direct |

### 🗜️ Compressor Agents (`snippets/compressor/`)

| ID | Name | Complexity | Sample Source | Mode |
|----|------|------------|---------------|------|
| compressor-agent | Context Compressor | Advanced | 28 | Direct |

### 🏗️ Complex Agents - Interactive (`snippets/complex/`)

| ID | Name | Complexity | Sample Source | Mode |
|----|------|------------|---------------|------|
| crew-agent | Crew Multi-Agents (local) | Advanced | 55 | Interactive |
| crew-server-agent | Crew Server (HTTP API) | Advanced | 56 | Interactive |
| remote-agent | Remote Agent (Client) | Intermediate | 51 | Interactive |
| pipeline-agent | Pipeline (Chained) | Advanced | 56 | Interactive |

---

## Authorized Sample Sources

Snippets are generated from the following Nova project samples:

### Direct Generation
- `05` - Chat streaming
- `06` - Contextual chat
- `13` - Basic RAG (in-memory)
- `18` - Simple tools
- `19` - Parallel tools
- `23` - Structured output (Go struct)
- `24` - Structured output (JSON Schema)
- `25` - Structured output with validation
- `28` - Compressor agent
- `47` - Tools with confirmation
- `69` - RAG with JSON persistent store

### Interactive Mode (Questions/Answers)
- `51` - Remote agent (client)
- `55` - Crew agent (local)
- `56` - Crew server agent / Pipeline agent

---

## Request → Snippet Mapping

| Request Keywords | Selected Snippet |
|------------------|------------------|
| "chat", "conversation", "chatbot" | streaming-chat or contextual-chat |
| "streaming", "real-time" | streaming-chat |
| "context", "memory", "history" | contextual-chat |
| "RAG", "search", "embeddings" | basic-rag |
| "persistent", "store", "save embeddings" | jsonstore-rag |
| "tools", "function", "API" | simple-tools |
| "parallel", "concurrent" | parallel-tools |
| "confirmation", "human validation" | confirmation-tools |
| "structured", "JSON", "schema" | structured-output |
| "JSON Schema", "constraints" | structured-schema |
| "validation", "retry" | structured-validation |
| "compression", "long context" | compressor-agent |
| "crew", "team", "multi-agents" | crew-agent (interactive) |
| "server", "API", "expose" | crew-server-agent (interactive) |
| "remote", "client", "distant", "connect" | remote-agent (interactive) |
| "pipeline", "chain", "steps" | pipeline-agent (interactive) |

---

## Usage Examples

```bash
# Simple Chat
"generate a chat agent with streaming"
→ Uses: snippets/chat/streaming-chat.md

# RAG
"create a RAG agent for FAQ"
→ Uses: snippets/rag/basic-rag.md

# RAG with persistence
"create a RAG agent that saves embeddings to JSON"
→ Uses: snippets/rag/jsonstore-rag.md

# Tools
"generate an agent with tools to calculate and get time"
→ Uses: snippets/tools/simple-tools.md

# Structured
"create an agent that extracts data in JSON"
→ Uses: snippets/structured/structured-output.md

# Complex (Interactive)
"create a multi-agent crew"
→ Uses: snippets/complex/crew-agent.md
→ Claude asks configuration questions
```
