# Docker Compose - Simple Agent Deployment

This snippet provides **Docker Compose** configurations for deploying simple Nova agents (chat, RAG, tools, structured, server) using **Docker Agentic Compose**.

## Category
**Docker & Deployment - Simple Agents**

## Use Case
- Deploy single or multiple instances of simple agents
- Leverage Docker Agentic Compose for automatic AI model management
- Configure agents via environment variables
- Persist data with Docker volumes
- Scale agents horizontally

## What is Docker Agentic Compose?

**Docker Agentic Compose** (Docker Compose 2.38.0+) allows declaring **AI models as first-class resources**:

1. **Declares models** in `compose.yml` file
2. **Automatically injects** environment variables (URL, model name) into services
3. **Manages lifecycle** via Docker Model Runner (DMR)
4. **Ensures portability** across environments

## Prerequisites

**CRITICAL Requirements**:
- Go 1.25.4 in go.mod
- Nova SDK latest version: `github.com/snipwise/nova latest`
- Dockerfile with `FROM golang:1.25.5-alpine`
- See sample 67 for complete working example

### Required Software

1. **Docker Desktop 4.36+** (includes Docker Compose 2.38+)
   - Download: https://www.docker.com/products/docker-desktop

2. **Docker Model Runner** (included in Docker Desktop)
   ```bash
   # Verify installation
   docker model --version
   ```

### Required Models

Models are automatically pulled by Docker Agentic Compose on first run. To pre-download:

```bash
# Chat models
docker model pull ai/qwen2.5:1.5B-F16
docker model pull hf.co/menlo/jan-nano-gguf:q4_k_m

# Embedding models (for RAG)
docker model pull ai/mxbai-embed-large
docker model pull ai/embeddinggemma:latest

# Specialized models
docker model pull ai/qwen2.5:0.5B-F16  # Compressor

# List available models
docker model list
```

## Template 1: Single Chat Agent

### Project Structure
```
my-chat-agent/
├── Dockerfile
├── compose.yml
├── main.go
└── go.mod
```

### compose.yml

```yaml
# === SERVICES ===
services:
  chat-agent:
    build:
      context: .
      dockerfile: Dockerfile

    # For interactive CLI agents
    stdin_open: true
    tty: true

    environment:
      # Application configuration
      NOVA_LOG_LEVEL: INFO
      AGENT_NAME: "helpful-assistant"
      SYSTEM_INSTRUCTIONS: "You are a helpful and concise assistant."

      # Model configuration injected by Docker Agentic Compose
      # ENGINE_URL: (auto-injected by Docker Model Runner)
      # CHAT_MODEL_ID: (auto-injected from models section)

    # AI models used by this service
    models:
      chat-model:
        endpoint_var: ENGINE_URL      # Environment variable for LLM endpoint
        model_var: CHAT_MODEL_ID      # Environment variable for model name

# === GLOBAL MODELS ===
# Define AI models available to services
models:
  chat-model:
    model: ai/qwen2.5:1.5B-F16
    # Optional: context_size: 32768
```

### main.go (adapted for environment variables)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/models"
)

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func main() {
    ctx := context.Background()

    // Configuration from environment variables (injected by Docker Compose)
    engineURL := getEnv("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
    modelID := getEnv("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
    agentName := getEnv("AGENT_NAME", "assistant")
    systemInstructions := getEnv("SYSTEM_INSTRUCTIONS", "You are a helpful assistant.")

    agent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               agentName,
            EngineURL:          engineURL,
            SystemInstructions: systemInstructions,
        },
        models.Config{
            Name:        modelID,
            Temperature: models.Float64(0.7),
            MaxTokens:   models.Int(2000),
        },
    )
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    fmt.Printf("Chat agent started with model: %s\n", modelID)
    // ... rest of your chat logic
}
```

### Usage

```bash
# Build and start
docker compose up --build -d

# Interact with agent
docker compose exec chat-agent ./agent-binary

# View logs
docker compose logs -f chat-agent

# Stop
docker compose down
```

## Template 2: RAG Agent with Persistence

### Project Structure
```
my-rag-agent/
├── Dockerfile
├── compose.yml
├── main.go
├── go.mod
├── documents/       # Documents to index
└── store/          # Persistent embeddings store
```

### compose.yml

```yaml
services:
  rag-agent:
    build:
      context: .
      dockerfile: Dockerfile

    stdin_open: true
    tty: true

    environment:
      NOVA_LOG_LEVEL: INFO
      AGENT_NAME: "knowledge-assistant"
      SYSTEM_INSTRUCTIONS: "You are a knowledgeable assistant with access to document retrieval."

      # Paths for documents and store
      DOCS_PATH: ./documents
      STORE_PATH: ./store
      STORE_FILE: knowledge-base.json

    # Mount volumes for persistence
    volumes:
      - ./documents:/app/documents:ro   # Read-only documents
      - ./store:/app/store:rw           # Read-write embeddings store

    models:
      embedding-model:
        endpoint_var: ENGINE_URL
        model_var: EMBEDDING_MODEL_ID

      chat-model:
        endpoint_var: ENGINE_URL
        model_var: CHAT_MODEL_ID

models:
  embedding-model:
    model: ai/mxbai-embed-large

  chat-model:
    model: ai/qwen2.5:1.5B-F16
```

### main.go (RAG with persistence)

```go
package main

import (
    "context"
    "log"
    "os"
    "path/filepath"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/models"
    "github.com/snipwise/nova/nova-sdk/toolbox/files"
)

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func main() {
    ctx := context.Background()

    engineURL := getEnv("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
    embeddingModel := getEnv("EMBEDDING_MODEL_ID", "ai/mxbai-embed-large")
    chatModel := getEnv("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")

    docsPath := getEnv("DOCS_PATH", "./documents")
    storePath := getEnv("STORE_PATH", "./store")
    storeFile := getEnv("STORE_FILE", "knowledge-base.json")
    storeFilePath := filepath.Join(storePath, storeFile)

    // Create RAG agent
    ragAgent, err := rag.NewAgent(
        ctx,
        agents.Config{
            Name:      "rag-agent",
            EngineURL: engineURL,
        },
        models.Config{
            Name: embeddingModel,
        },
    )
    if err != nil {
        log.Fatalf("Failed to create RAG agent: %v", err)
    }

    // Load or create embeddings store
    if files.FileExists(storeFilePath) {
        log.Printf("Loading existing store: %s", storeFilePath)
        ragAgent.LoadStore(storeFilePath)
    } else {
        log.Printf("Creating new store from documents: %s", docsPath)
        // Index documents from docsPath
        // ... your indexing logic here

        // Persist store
        ragAgent.PersistStore(storeFilePath)
    }

    log.Println("RAG agent ready")
    // ... rest of your RAG logic
}
```

## Template 3: Server Agent (HTTP API)

### compose.yml

```yaml
services:
  api-server:
    build:
      context: .
      dockerfile: Dockerfile

    # Expose HTTP port
    ports:
      - "8080:8080"

    environment:
      NOVA_LOG_LEVEL: INFO
      SERVER_PORT: "8080"
      AGENT_NAME: "api-assistant"
      SYSTEM_INSTRUCTIONS: "You are a helpful API assistant."

    models:
      chat-model:
        endpoint_var: ENGINE_URL
        model_var: CHAT_MODEL_ID

models:
  chat-model:
    model: ai/qwen2.5:1.5B-F16
```

### Usage

```bash
# Start server
docker compose up --build -d

# Test API
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, how are you?"}'

# Test streaming endpoint
curl -X POST http://localhost:8080/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "Tell me a story"}'

# Health check
curl http://localhost:8080/health

# View logs
docker compose logs -f api-server
```

## Template 4: Multiple Agents (Parallel Deployment)

Deploy multiple agent instances with different configurations.

### compose.yml

```yaml
# === COMMON CONFIGURATION ===
x-common-environment: &common-env
  NOVA_LOG_LEVEL: INFO

x-common-models: &common-models
  chat-model:
    endpoint_var: ENGINE_URL
    model_var: CHAT_MODEL_ID

# === SERVICES ===
services:
  # Agent 1: Creative Writer
  writer-agent:
    build:
      context: .
      dockerfile: Dockerfile
    stdin_open: true
    tty: true
    environment:
      <<: *common-env
      AGENT_NAME: "creative-writer"
      SYSTEM_INSTRUCTIONS: "You are a creative writer specializing in storytelling."
      TEMPERATURE: "0.9"
    models:
      <<: *common-models

  # Agent 2: Code Expert
  coder-agent:
    build:
      context: .
      dockerfile: Dockerfile
    stdin_open: true
    tty: true
    environment:
      <<: *common-env
      AGENT_NAME: "code-expert"
      SYSTEM_INSTRUCTIONS: "You are an expert programmer specializing in Go."
      TEMPERATURE: "0.0"
    models:
      <<: *common-models

  # Agent 3: Data Analyst
  analyst-agent:
    build:
      context: .
      dockerfile: Dockerfile
    stdin_open: true
    tty: true
    environment:
      <<: *common-env
      AGENT_NAME: "data-analyst"
      SYSTEM_INSTRUCTIONS: "You are a data analyst providing insights."
      TEMPERATURE: "0.3"
    models:
      <<: *common-models

models:
  chat-model:
    model: ai/qwen2.5:1.5B-F16
```

### Usage

```bash
# Start all agents
docker compose up --build -d

# Interact with specific agent
docker compose exec writer-agent ./agent-binary
docker compose exec coder-agent ./agent-binary
docker compose exec analyst-agent ./agent-binary

# View all logs
docker compose logs -f

# View specific agent logs
docker compose logs -f writer-agent

# Stop all
docker compose down
```

## YAML Anchors Explained

YAML anchors allow reusing configurations:

```yaml
# Define anchor with &name
x-common-variables: &common-vars
  NOVA_LOG_LEVEL: INFO
  TIMEOUT: "30s"

x-common-models: &common-models
  chat-model:
    endpoint_var: ENGINE_URL
    model_var: CHAT_MODEL_ID

services:
  agent-1:
    environment:
      <<: *common-vars      # Merge object (for maps)
      AGENT_NAME: "agent-1"
    models:
      <<: *common-models    # Merge object

  agent-2:
    environment:
      <<: *common-vars
      AGENT_NAME: "agent-2"
    models:
      <<: *common-models
```

**Syntax**:
- `&anchor-name` - Define anchor
- `<<: *anchor-name` - Merge anchor (for objects/maps)
- `*anchor-name` - Reference anchor (for lists/arrays)

## Environment Variable Injection

Docker Agentic Compose automatically injects these variables:

```bash
# Automatically injected by Docker Model Runner
ENGINE_URL=http://host.docker.internal:12434/engines/llama.cpp/v1
CHAT_MODEL_ID=ai/qwen2.5:1.5B-F16
EMBEDDING_MODEL_ID=ai/mxbai-embed-large

# You define these in compose.yml
NOVA_LOG_LEVEL=INFO
AGENT_NAME=my-agent
SYSTEM_INSTRUCTIONS=...
```

Access in Go code:
```go
engineURL := os.Getenv("ENGINE_URL")
modelID := os.Getenv("CHAT_MODEL_ID")
```

## Volume Management

### Persistent Data Volumes

```yaml
services:
  my-agent:
    volumes:
      # Bind mount (host directory)
      - ./data:/app/data:rw           # Read-write
      - ./config:/app/config:ro       # Read-only

      # Named volume (managed by Docker)
      - agent-data:/app/store

volumes:
  # Define named volumes
  agent-data:
    driver: local
```

### Common Volume Patterns

```yaml
# Pattern 1: Configuration files (read-only)
volumes:
  - ./config.yaml:/app/config.yaml:ro
  - ./prompts:/app/prompts:ro

# Pattern 2: Persistent data (read-write)
volumes:
  - ./store:/app/store:rw
  - ./logs:/app/logs:rw

# Pattern 3: Document repository
volumes:
  - ./documents:/app/documents:ro   # Read-only source docs
  - ./store:/app/store:rw           # Read-write embeddings
```

## Networking

### Service-to-Service Communication

```yaml
services:
  api-server:
    ports:
      - "8080:8080"
    networks:
      - agent-network

  worker-agent:
    depends_on:
      - api-server
    environment:
      API_URL: http://api-server:8080
    networks:
      - agent-network

networks:
  agent-network:
    driver: bridge
```

## Common Commands

```bash
# Build and start all services
docker compose up --build -d

# Start specific service
docker compose up -d chat-agent

# View logs (all services)
docker compose logs -f

# View logs (specific service)
docker compose logs -f chat-agent

# Execute command in running container
docker compose exec chat-agent ./agent-binary
docker compose exec -it chat-agent /bin/sh

# Stop services
docker compose down

# Stop and remove volumes
docker compose down -v

# Rebuild without cache
docker compose build --no-cache

# View running services
docker compose ps

# View resource usage
docker stats
```

## Troubleshooting

### Issue: "ENGINE_URL not set"

**Cause**: Docker Model Runner not accessible or model not declared.

**Solution**:
```bash
# Verify Docker Model Runner is running
docker model list

# Check model declaration in compose.yml
models:
  chat-model:
    model: ai/qwen2.5:1.5B-F16  # Must be valid model
```

### Issue: "Permission denied on volume mount"

**Cause**: File permission mismatch between host and container.

**Solution**:
```bash
# Set proper permissions on host
chmod -R 755 ./data ./store

# Or use named volumes instead of bind mounts
volumes:
  - agent-data:/app/data
```

### Issue: "Port already in use"

**Cause**: Another service using the same port.

**Solution**:
```bash
# Change port mapping in compose.yml
ports:
  - "8081:8080"  # Map to different host port

# Or stop conflicting service
lsof -i :8080  # Find process
kill -9 <PID>
```

## Best Practices

### ✅ DO:

1. **Use YAML anchors** for common configurations
2. **Use environment variables** for all configurations
3. **Mount volumes** for persistent data
4. **Use meaningful service names**
5. **Add health checks** for server agents
6. **Use specific model versions** in production
7. **Set resource limits** to prevent overconsumption

### ❌ DON'T:

1. **Don't hard-code values** in application code
2. **Don't use :latest tag** in production
3. **Don't ignore volume permissions**
4. **Don't expose ports** unnecessarily
5. **Don't skip health checks** for critical services

## Resource Limits (Production)

```yaml
services:
  chat-agent:
    build:
      context: .
      dockerfile: Dockerfile

    # Resource limits
    deploy:
      resources:
        limits:
          cpus: '1.0'      # Max 1 CPU core
          memory: 2G       # Max 2GB RAM
        reservations:
          cpus: '0.5'      # Reserve 0.5 CPU
          memory: 512M     # Reserve 512MB RAM

    # Restart policy
    restart: unless-stopped
```

## Next Steps

1. **Create Dockerfile**: See [dockerfile-template.md](dockerfile-template.md)
2. **Adapt main.go**: Use environment variables for configuration
3. **Test locally**: `docker compose up --build`
4. **Scale up**: Add more services for complex deployments
5. **Deploy to cloud**: See [dockerization-guide.md](dockerization-guide.md)

## Related Snippets

- [dockerfile-template.md](dockerfile-template.md) - Dockerfile creation
- [docker-compose-complex.md](docker-compose-complex.md) - Multi-agent systems
- [dockerization-guide.md](dockerization-guide.md) - Complete guide

## Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Docker Agentic Compose](https://docs.docker.com/ai/compose/models-and-compose/)
- [Docker Model Runner](https://docs.docker.com/ai/model-runner/)
