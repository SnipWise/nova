# Docker Compose - Complex Agent Deployment (Crew & Pipeline)

This snippet provides **Docker Compose** configurations for deploying complex Nova agent systems (crews, pipelines, orchestrators) using **Docker Agentic Compose**.

## Category
**Docker & Deployment - Complex Multi-Agent Systems**

## Use Case
- Deploy multi-agent crews with collaboration
- Deploy agent pipelines with sequential processing
- Deploy orchestrator-based routing systems
- Scale complex agent architectures
- Manage multiple specialized models per agent
- Implement production-ready multi-agent systems

## What is a Complex Agent System?

### Crew Agent
Multiple specialized agents working collaboratively on tasks:
- **Research Agent** + **Writer Agent** + **Editor Agent**
- Each agent has specific role and tools
- Agents can communicate and share context
- Coordinated by orchestrator or hierarchy

### Pipeline Agent
Sequential agent chain with transformations:
- **Input Processor** → **Analyzer** → **Generator** → **Formatter**
- Each step transforms/enriches data
- Output of one agent feeds next agent
- Error handling at each stage

### Orchestrator Agent
Topic detection and query routing:
- Analyzes user queries to detect intent/topic
- Routes queries to appropriate specialized agents
- Fast classification with minimal latency
- Integrates with crew agents

## Prerequisites

**CRITICAL Requirements**:
- Go 1.25.4 in go.mod
- Nova SDK latest version: `github.com/snipwise/nova latest`
- Dockerfile with `FROM golang:1.25.5-alpine`
- Docker Desktop 4.36+ with Docker Compose 2.38+
- See sample 67 for go.mod/Dockerfile setup

## Template 1: Crew Agent with Orchestrator

Deploy a collaborative multi-agent crew for content creation.

### Project Structure
```
content-crew/
├── Dockerfile
├── compose.yml
├── main.go
├── go.mod
├── docs/                 # System instructions per agent
│   ├── research.instructions.md
│   ├── writer.instructions.md
│   └── editor.instructions.md
└── data/                 # Shared data between agents
    └── output/
```

### compose.yml

```yaml
# === COMMON CONFIGURATION ===
# Shared environment variables
x-common-environment: &common-env
  NOVA_LOG_LEVEL: INFO
  DOCS_PATH: ./docs
  DATA_PATH: ./data

# Shared volumes
x-common-volumes: &common-volumes
  - ./docs:/app/docs:ro      # Read-only instructions
  - ./data:/app/data:rw      # Read-write shared data

# === SERVICES ===
services:
  # Main orchestrator + crew service
  content-crew:
    build:
      context: .
      dockerfile: Dockerfile

    stdin_open: true
    tty: true

    environment:
      <<: *common-env

      # Crew configuration
      CREW_NAME: "content-creation-crew"
      CREW_MODE: "collaborative"  # collaborative | hierarchical

      # Orchestrator configuration
      ENABLE_ORCHESTRATOR: "true"
      ORCHESTRATOR_TOPICS: "research,writing,editing"

      # Agent roles
      AGENT_1_ROLE: "researcher"
      AGENT_1_INSTRUCTIONS_FILE: "research.instructions.md"

      AGENT_2_ROLE: "writer"
      AGENT_2_INSTRUCTIONS_FILE: "writer.instructions.md"

      AGENT_3_ROLE: "editor"
      AGENT_3_INSTRUCTIONS_FILE: "editor.instructions.md"

    volumes: *common-volumes

    # Multiple models for different agents
    models:
      # Orchestrator model (fast, small)
      orchestrator-model:
        endpoint_var: ENGINE_URL
        model_var: ORCHESTRATOR_MODEL_ID

      # Research agent (medium, factual)
      research-model:
        endpoint_var: ENGINE_URL
        model_var: RESEARCH_MODEL_ID

      # Writer agent (large, creative)
      writer-model:
        endpoint_var: ENGINE_URL
        model_var: WRITER_MODEL_ID

      # Editor agent (medium, precise)
      editor-model:
        endpoint_var: ENGINE_URL
        model_var: EDITOR_MODEL_ID

      # RAG for knowledge retrieval
      embedding-model:
        endpoint_var: ENGINE_URL
        model_var: EMBEDDING_MODEL_ID

      # Compressor for long contexts
      compressor-model:
        endpoint_var: ENGINE_URL
        model_var: COMPRESSOR_MODEL_ID

# === GLOBAL MODELS ===
models:
  # Fast orchestrator (topic detection)
  orchestrator-model:
    model: hf.co/menlo/lucy-gguf:q4_k_m
    # context_size: 16384

  # Research agent
  research-model:
    model: ai/qwen2.5:1.5B-F16
    # context_size: 32768

  # Writer agent (more creative)
  writer-model:
    model: huggingface.co/tensorblock/nvidia_nemotron-mini-4b-instruct-gguf:q4_k_m
    # context_size: 32768

  # Editor agent
  editor-model:
    model: ai/qwen2.5:1.5B-F16
    # context_size: 32768

  # Embedding model
  embedding-model:
    model: ai/mxbai-embed-large

  # Compressor (small, fast)
  compressor-model:
    model: ai/qwen2.5:0.5B-F16
    # context_size: 16384
```

### main.go (Crew with Orchestrator)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/crew"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
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
    docsPath := getEnv("DOCS_PATH", "./docs")

    // === CREATE ORCHESTRATOR AGENT ===
    orchestratorModelID := getEnv("ORCHESTRATOR_MODEL_ID", "hf.co/menlo/lucy-gguf:q4_k_m")

    orchestratorAgent, err := orchestrator.NewAgent(
        ctx,
        agents.Config{
            Name:      "orchestrator",
            EngineURL: engineURL,
        },
        models.Config{
            Name:        orchestratorModelID,
            Temperature: models.Float64(0.0),
        },
    )
    if err != nil {
        log.Fatalf("Failed to create orchestrator: %v", err)
    }

    // === CREATE SPECIALIZED AGENTS ===

    // Research Agent
    researchInstructions, _ := files.ReadTextFile(filepath.Join(docsPath, "research.instructions.md"))
    researchAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "researcher",
            EngineURL:          engineURL,
            SystemInstructions: researchInstructions,
        },
        models.Config{
            Name:        getEnv("RESEARCH_MODEL_ID", "ai/qwen2.5:1.5B-F16"),
            Temperature: models.Float64(0.3), // Factual
        },
    )
    if err != nil {
        log.Fatalf("Failed to create research agent: %v", err)
    }

    // Writer Agent
    writerInstructions, _ := files.ReadTextFile(filepath.Join(docsPath, "writer.instructions.md"))
    writerAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "writer",
            EngineURL:          engineURL,
            SystemInstructions: writerInstructions,
        },
        models.Config{
            Name:        getEnv("WRITER_MODEL_ID", "ai/qwen2.5:1.5B-F16"),
            Temperature: models.Float64(0.8), // Creative
        },
    )
    if err != nil {
        log.Fatalf("Failed to create writer agent: %v", err)
    }

    // Editor Agent
    editorInstructions, _ := files.ReadTextFile(filepath.Join(docsPath, "editor.instructions.md"))
    editorAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "editor",
            EngineURL:          engineURL,
            SystemInstructions: editorInstructions,
        },
        models.Config{
            Name:        getEnv("EDITOR_MODEL_ID", "ai/qwen2.5:1.5B-F16"),
            Temperature: models.Float64(0.2), // Precise
        },
    )
    if err != nil {
        log.Fatalf("Failed to create editor agent: %v", err)
    }

    // === CREATE CREW ===
    crewAgent, err := crew.NewAgent(
        ctx,
        agents.Config{
            Name: getEnv("CREW_NAME", "content-crew"),
        },
    )
    if err != nil {
        log.Fatalf("Failed to create crew: %v", err)
    }

    // Add agents to crew
    crewAgent.AddAgent(researchAgent)
    crewAgent.AddAgent(writerAgent)
    crewAgent.AddAgent(editorAgent)

    // Set orchestrator for routing
    crewAgent.SetOrchestratorAgent(orchestratorAgent)

    fmt.Println("✅ Content Creation Crew Ready")
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    fmt.Printf("Orchestrator: %s\n", orchestratorModelID)
    fmt.Printf("Crew Members: researcher, writer, editor\n")

    // ... rest of crew interaction logic
}
```

## Template 2: Pipeline Agent

Deploy a sequential pipeline for document processing.

### compose.yml

```yaml
x-common-environment: &common-env
  NOVA_LOG_LEVEL: INFO
  PIPELINE_NAME: "document-pipeline"

x-common-volumes: &common-volumes
  - ./input:/app/input:ro
  - ./output:/app/output:rw
  - ./temp:/app/temp:rw

services:
  document-pipeline:
    build:
      context: .
      dockerfile: Dockerfile

    environment:
      <<: *common-env

      # Pipeline configuration
      PIPELINE_STAGES: "extractor,analyzer,summarizer,formatter"

      # Stage 1: Extractor
      EXTRACTOR_ROLE: "Extract structured data from documents"
      EXTRACTOR_MODEL: "EXTRACTOR_MODEL_ID"

      # Stage 2: Analyzer
      ANALYZER_ROLE: "Analyze extracted data for insights"
      ANALYZER_MODEL: "ANALYZER_MODEL_ID"

      # Stage 3: Summarizer
      SUMMARIZER_ROLE: "Generate concise summary"
      SUMMARIZER_MODEL: "SUMMARIZER_MODEL_ID"

      # Stage 4: Formatter
      FORMATTER_ROLE: "Format output as markdown report"
      FORMATTER_MODEL: "FORMATTER_MODEL_ID"

    volumes: *common-volumes

    models:
      extractor-model:
        endpoint_var: ENGINE_URL
        model_var: EXTRACTOR_MODEL_ID

      analyzer-model:
        endpoint_var: ENGINE_URL
        model_var: ANALYZER_MODEL_ID

      summarizer-model:
        endpoint_var: ENGINE_URL
        model_var: SUMMARIZER_MODEL_ID

      formatter-model:
        endpoint_var: ENGINE_URL
        model_var: FORMATTER_MODEL_ID

      compressor-model:
        endpoint_var: ENGINE_URL
        model_var: COMPRESSOR_MODEL_ID

models:
  extractor-model:
    model: hf.co/menlo/jan-nano-gguf:q4_k_m
    # context_size: 16384

  analyzer-model:
    model: ai/qwen2.5:1.5B-F16
    # context_size: 32768

  summarizer-model:
    model: ai/qwen2.5:1.5B-F16
    # context_size: 32768

  formatter-model:
    model: hf.co/menlo/jan-nano-gguf:q4_k_m
    # context_size: 16384

  compressor-model:
    model: ai/qwen2.5:0.5B-F16
    # context_size: 16384
```

## Template 3: Multi-Service Crew (Distributed)

Deploy crew agents as separate services for better scalability.

### compose.yml

```yaml
# === COMMON CONFIGURATION ===
x-common-environment: &common-env
  NOVA_LOG_LEVEL: INFO

x-common-volumes: &common-volumes
  - ./shared:/app/shared:rw

# === SERVICES ===
services:
  # Orchestrator Service
  orchestrator:
    build:
      context: ./orchestrator
      dockerfile: Dockerfile
    ports:
      - "8080:8080"  # HTTP API for routing
    environment:
      <<: *common-env
      SERVICE_NAME: "orchestrator"
      PORT: "8080"
    models:
      orchestrator-model:
        endpoint_var: ENGINE_URL
        model_var: ORCHESTRATOR_MODEL_ID
    networks:
      - crew-network

  # Research Agent Service
  researcher:
    build:
      context: ./agents/researcher
      dockerfile: Dockerfile
    ports:
      - "8081:8080"
    environment:
      <<: *common-env
      SERVICE_NAME: "researcher"
      PORT: "8080"
      ORCHESTRATOR_URL: http://orchestrator:8080
    volumes: *common-volumes
    models:
      research-model:
        endpoint_var: ENGINE_URL
        model_var: RESEARCH_MODEL_ID
      embedding-model:
        endpoint_var: ENGINE_URL
        model_var: EMBEDDING_MODEL_ID
    depends_on:
      - orchestrator
    networks:
      - crew-network

  # Writer Agent Service
  writer:
    build:
      context: ./agents/writer
      dockerfile: Dockerfile
    ports:
      - "8082:8080"
    environment:
      <<: *common-env
      SERVICE_NAME: "writer"
      PORT: "8080"
      ORCHESTRATOR_URL: http://orchestrator:8080
    volumes: *common-volumes
    models:
      writer-model:
        endpoint_var: ENGINE_URL
        model_var: WRITER_MODEL_ID
      compressor-model:
        endpoint_var: ENGINE_URL
        model_var: COMPRESSOR_MODEL_ID
    depends_on:
      - orchestrator
    networks:
      - crew-network

  # Editor Agent Service
  editor:
    build:
      context: ./agents/editor
      dockerfile: Dockerfile
    ports:
      - "8083:8080"
    environment:
      <<: *common-env
      SERVICE_NAME: "editor"
      PORT: "8080"
      ORCHESTRATOR_URL: http://orchestrator:8080
    volumes: *common-volumes
    models:
      editor-model:
        endpoint_var: ENGINE_URL
        model_var: EDITOR_MODEL_ID
    depends_on:
      - orchestrator
    networks:
      - crew-network

# === GLOBAL MODELS ===
models:
  orchestrator-model:
    model: hf.co/menlo/lucy-gguf:q4_k_m

  research-model:
    model: ai/qwen2.5:1.5B-F16

  writer-model:
    model: huggingface.co/tensorblock/nvidia_nemotron-mini-4b-instruct-gguf:q4_k_m

  editor-model:
    model: ai/qwen2.5:1.5B-F16

  embedding-model:
    model: ai/mxbai-embed-large

  compressor-model:
    model: ai/qwen2.5:0.5B-F16

# === NETWORKS ===
networks:
  crew-network:
    driver: bridge
```

### Project Structure (Multi-Service)

```
multi-service-crew/
├── compose.yml
├── orchestrator/
│   ├── Dockerfile
│   ├── main.go
│   └── go.mod
└── agents/
    ├── researcher/
    │   ├── Dockerfile
    │   ├── main.go
    │   └── go.mod
    ├── writer/
    │   ├── Dockerfile
    │   ├── main.go
    │   └── go.mod
    └── editor/
        ├── Dockerfile
        ├── main.go
        └── go.mod
```

### Usage (Multi-Service)

```bash
# Build and start entire crew
docker compose up --build -d

# Check all services
docker compose ps

# View orchestrator logs
docker compose logs -f orchestrator

# Test orchestrator API
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{"query": "Research the history of AI"}'

# Scale a specific service
docker compose up -d --scale researcher=3

# Stop all
docker compose down
```

## Template 4: RAG + Crew + Compressor (Full-Featured)

Complete multi-agent system with all advanced features.

### compose.yml

```yaml
x-common-environment: &common-env
  NOVA_LOG_LEVEL: INFO
  DOCS_PATH: ./docs
  STORE_PATH: ./store
  SHEETS_PATH: ./sheets

x-common-volumes: &common-volumes
  - ./docs:/app/docs:rw
  - ./store:/app/store:rw
  - ./sheets:/app/sheets:rw

x-common-models: &common-models
  embedding-model:
    endpoint_var: ENGINE_URL
    model_var: EMBEDDING_MODEL_ID

  compressor-model:
    endpoint_var: ENGINE_URL
    model_var: COMPRESSOR_MODEL_ID

  metadata-model:
    endpoint_var: ENGINE_URL
    model_var: METADATA_MODEL_ID

services:
  # Agent with RAG + Compressor
  advanced-agent:
    build:
      context: .
      dockerfile: Dockerfile

    stdin_open: true
    tty: true

    environment:
      <<: *common-env

      AGENT_NAME: "advanced-knowledge-agent"
      AGENT_ROLE: "Knowledge expert with long-term memory"

      # RAG configuration
      RAG_ENABLED: "true"
      RAG_SIMILARITY_THRESHOLD: "0.6"
      RAG_MAX_RESULTS: "5"
      RAG_STORE_FILE: "knowledge-base.json"

      # Compressor configuration
      COMPRESSOR_ENABLED: "true"
      COMPRESSOR_THRESHOLD: "8000"  # Compress when context exceeds 8000 chars
      COMPRESSOR_PRESERVE_RECENT: "3"  # Keep last 3 messages uncompressed

      # Metadata extraction
      METADATA_ENABLED: "true"

    volumes: *common-volumes

    models:
      <<: *common-models

      # Main agent model
      agent-model:
        endpoint_var: ENGINE_URL
        model_var: AGENT_MODEL_ID

models:
  agent-model:
    model: ai/qwen2.5:1.5B-F16

  embedding-model:
    model: ai/mxbai-embed-large

  compressor-model:
    model: ai/qwen2.5:0.5B-F16

  metadata-model:
    model: hf.co/menlo/jan-nano-gguf:q4_k_m
```

## Model Selection Guidelines

### By Agent Role

| Agent Role | Recommended Model | Characteristics |
|------------|------------------|-----------------|
| **Orchestrator** | `hf.co/menlo/lucy-gguf:q4_k_m` | Fast, small, topic detection |
| **Research** | `ai/qwen2.5:1.5B-F16` | Factual, medium speed |
| **Creative Writer** | `nvidia_nemotron-mini-4b-instruct` | Large, creative |
| **Code Generator** | `ai/qwen2.5:1.5B-F16` | Precise, deterministic |
| **Data Analyst** | `ai/qwen2.5:1.5B-F16` | Analytical, structured |
| **Compressor** | `ai/qwen2.5:0.5B-F16` | Fast, minimal |
| **Metadata Extractor** | `hf.co/menlo/jan-nano-gguf` | Small, structured |
| **Embeddings (RAG)** | `ai/mxbai-embed-large` | High-quality vectors |

### By Use Case

| Use Case | Recommended Setup |
|----------|------------------|
| **Content Creation** | Writer (4B) + Editor (1.5B) + Research (1.5B) |
| **Customer Support** | Orchestrator (160M) + FAQ (1.5B) + Escalation (1.5B) |
| **Code Review** | Analyzer (1.5B) + Suggester (1.5B) + Formatter (160M) |
| **Document Processing** | Extractor (160M) + Analyzer (1.5B) + Summarizer (1.5B) |

## Advanced Features

### Health Checks

```yaml
services:
  api-agent:
    # ... other config

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      start_period: 5s
      retries: 3
```

### Resource Limits

```yaml
services:
  heavy-agent:
    # ... other config

    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G
```

### Restart Policies

```yaml
services:
  critical-agent:
    # ... other config

    restart: unless-stopped  # always | on-failure | unless-stopped
```

### Logging Configuration

```yaml
services:
  production-agent:
    # ... other config

    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Common Commands

```bash
# === BUILD & START ===
# Build and start all services
docker compose up --build -d

# Start with logs visible
docker compose up --build --no-log-prefix

# Build without cache
docker compose build --no-cache

# === MANAGEMENT ===
# View all services status
docker compose ps

# View logs (all services)
docker compose logs -f

# View logs (specific service)
docker compose logs -f researcher

# Execute in container
docker compose exec content-crew ./agent-binary

# === SCALING ===
# Scale specific service
docker compose up -d --scale researcher=3

# === CLEANUP ===
# Stop all services
docker compose down

# Stop and remove volumes
docker compose down -v

# Remove all (including images)
docker compose down --rmi all

# === MONITORING ===
# View resource usage
docker stats

# View network info
docker network ls
docker network inspect <network-name>
```

## Troubleshooting

### Issue: "Services can't communicate"

**Cause**: Network configuration problem.

**Solution**:
```yaml
# Ensure all services are on same network
services:
  service-1:
    networks:
      - crew-network
  service-2:
    networks:
      - crew-network

networks:
  crew-network:
    driver: bridge
```

### Issue: "Model not found for specific service"

**Cause**: Model not declared or wrong model reference.

**Solution**:
```yaml
# Ensure model is declared globally
models:
  my-model:
    model: ai/qwen2.5:1.5B-F16

# And referenced in service
services:
  my-service:
    models:
      my-model:
        endpoint_var: ENGINE_URL
        model_var: MY_MODEL_ID
```

### Issue: "High memory usage"

**Cause**: Too many models loaded simultaneously.

**Solution**:
```yaml
# Add resource limits
services:
  agent:
    deploy:
      resources:
        limits:
          memory: 2G

# Or share models across services
models:
  shared-model:
    model: ai/qwen2.5:1.5B-F16
```

## Best Practices

### ✅ DO:

1. **Use orchestrator** for intelligent routing
2. **Share models** when agents have similar needs
3. **Use volumes** for inter-agent data sharing
4. **Add health checks** for production deployments
5. **Set resource limits** to prevent overconsumption
6. **Use networks** for service isolation
7. **Implement retry logic** in agent communication
8. **Monitor performance** with docker stats

### ❌ DON'T:

1. **Don't load all models** if not necessary
2. **Don't skip error handling** in agent chains
3. **Don't ignore dependencies** between services
4. **Don't expose all ports** publicly
5. **Don't use same model** for all agents (specialize!)
6. **Don't forget logging** configuration
7. **Don't skip testing** individual services first

## Production Deployment

### Security Checklist

- [ ] Remove `stdin_open` and `tty` for non-interactive services
- [ ] Add authentication to HTTP endpoints
- [ ] Use secrets management (Docker secrets, env files)
- [ ] Run containers as non-root user
- [ ] Enable TLS for inter-service communication
- [ ] Set up firewall rules
- [ ] Implement rate limiting
- [ ] Add API key validation

### Performance Checklist

- [ ] Set appropriate resource limits
- [ ] Use health checks
- [ ] Configure restart policies
- [ ] Enable logging with rotation
- [ ] Monitor with Prometheus/Grafana
- [ ] Use caching where applicable
- [ ] Optimize model selection
- [ ] Implement request queuing

## Next Steps

1. **Design architecture**: Plan agent roles and interactions
2. **Create Dockerfiles**: One per service or shared
3. **Configure compose.yml**: Define services, models, networks
4. **Implement agents**: Adapt to use environment variables
5. **Test locally**: `docker compose up`
6. **Scale**: Add replicas or services
7. **Deploy**: Cloud platforms (AWS ECS, Azure, GCP, K8s)

## Related Snippets

- [dockerfile-template.md](dockerfile-template.md) - Dockerfile creation
- [docker-compose-simple.md](docker-compose-simple.md) - Simple agents
- [dockerization-guide.md](dockerization-guide.md) - Complete guide

## Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Docker Agentic Compose](https://docs.docker.com/ai/compose/models-and-compose/)
- [Docker Networking](https://docs.docker.com/network/)
- [Nova SDK Crew Agents](https://github.com/snipwise/nova)
