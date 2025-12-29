# Complete Dockerization Guide for Nova Agents

This guide provides a **complete step-by-step process** for dockerizing and deploying Nova agent applications using **Docker** and **Docker Agentic Compose**.

## Category
**Docker & Deployment - Complete Guide**

## Table of Contents

1. [Introduction](#introduction)
2. [Prerequisites](#prerequisites)
3. [Dockerization Workflow](#dockerization-workflow)
4. [Step-by-Step Tutorial](#step-by-step-tutorial)
5. [Code Adaptation Guide](#code-adaptation-guide)
6. [Common Patterns](#common-patterns)
7. [Deployment Strategies](#deployment-strategies)
8. [Production Checklist](#production-checklist)
9. [Troubleshooting](#troubleshooting)

## Introduction

### What is Docker Agentic Compose?

**Docker Agentic Compose** is an extension of Docker Compose (v2.38.0+) that allows declaring **AI models as first-class resources** in containerized applications.

**Key Features:**
- Declarative model management in `compose.yml`
- Automatic environment variable injection
- Model lifecycle management via Docker Model Runner (DMR)
- Portability across development, staging, production
- Multi-model orchestration

### Benefits of Dockerizing Nova Agents

1. **Portability**: Same configuration across all environments
2. **Isolation**: Each agent runs in isolated container
3. **Scalability**: Easy horizontal scaling with Compose
4. **Reproducibility**: Version-locked dependencies and models
5. **Simplified Deployment**: One command to deploy entire stack
6. **Model Management**: Automatic model lifecycle handling

## Prerequisites

**CRITICAL Requirements for Nova Agents**:
- ✅ Go 1.25.4 in go.mod
- ✅ Nova SDK latest: `github.com/snipwise/nova latest`
- ✅ Dockerfile: `FROM golang:1.25.5-alpine`
- ✅ No local replace directives in go.mod for Docker builds
- 📚 Complete example: See sample 67 (dockerized chat agent)

### Required Software

#### 1. Docker Desktop 4.36+ (includes Docker Compose 2.38+)

**macOS / Windows:**
```bash
# Download from https://www.docker.com/products/docker-desktop
# Install and verify
docker --version
docker compose version
```

**Linux:**
```bash
# Install Docker Engine
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose
sudo apt-get update
sudo apt-get install docker-compose-plugin

# Verify
docker --version
docker compose version
```

#### 2. Docker Model Runner

**Verify installation:**
```bash
docker model --version
```

If not found:
```bash
# macOS/Linux
ln -s ~/.docker/cli-plugins/docker-model /usr/local/bin/docker-model

# Windows (PowerShell as admin)
New-Item -ItemType SymbolicLink -Path "C:\Program Files\Docker\cli-plugins\docker-model" -Target "$env:USERPROFILE\.docker\cli-plugins\docker-model.exe"
```

#### 3. Go 1.25.4+

```bash
# Verify Go installation
go version

# If not installed, download from https://go.dev/dl/
```

### Resource Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **RAM** | 8 GB | 16 GB |
| **Disk** | 20 GB free | 50 GB free |
| **CPU** | 4 cores | 8 cores |
| **GPU** | None | Apple Silicon / NVIDIA / AMD |

### AI Models

Models are auto-downloaded by Docker Agentic Compose. To pre-download:

```bash
# Essential models
docker model pull ai/qwen2.5:1.5B-F16              # Chat (1.5B)
docker model pull ai/mxbai-embed-large             # Embeddings
docker model pull hf.co/menlo/jan-nano-gguf:q4_k_m # Tools (160M)
docker model pull ai/qwen2.5:0.5B-F16              # Compressor (0.5B)

# Optional larger models
docker model pull huggingface.co/tensorblock/nvidia_nemotron-mini-4b-instruct-gguf:q4_k_m  # Writer (4B)

# Verify
docker model list
```

## Dockerization Workflow

### High-Level Process

```
1. Develop Agent Locally
   ├── Write main.go
   ├── Add dependencies (go.mod)
   └── Test with local LLM

2. Adapt for Docker
   ├── Replace hard-coded values with env vars
   ├── Use env.GetEnvOrDefault()
   └── Make paths configurable

3. Create Dockerfile
   ├── Multi-stage build
   ├── Go build stage
   └── Minimal runtime stage

4. Create compose.yml
   ├── Define services
   ├── Declare models
   ├── Configure environment
   └── Set up volumes

5. Build & Test
   ├── docker compose build
   ├── docker compose up
   └── Test functionality

6. Deploy
   ├── Push to registry (optional)
   └── Deploy to cloud/kubernetes
```

## Step-by-Step Tutorial

### Tutorial: Dockerize a Chat Agent

#### Step 1: Create Project Structure

```bash
# Create project directory
mkdir my-chat-agent
cd my-chat-agent

# Create necessary files
touch main.go go.mod Dockerfile compose.yml .dockerignore
```

#### Step 2: Write Initial Go Code

**main.go** (local development version):
```go
package main

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

func main() {
    ctx := context.Background()

    // Hard-coded configuration (to be replaced)
    engineURL := "http://localhost:12434/engines/llama.cpp/v1"
    modelName := "ai/qwen2.5:1.5B-F16"

    agent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "chat-agent",
            EngineURL:          engineURL,
            SystemInstructions: "You are a helpful assistant.",
        },
        models.Config{
            Name:        modelName,
            Temperature: models.Float64(0.7),
            MaxTokens:   models.Int(2000),
        },
    )
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    fmt.Println("Chat Agent Ready (type 'exit' to quit)")

    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("\nYou: ")
        if !scanner.Scan() {
            break
        }

        userInput := strings.TrimSpace(scanner.Text())
        if userInput == "exit" {
            break
        }

        chunkChan, err := agent.GenerateStreamingCompletion(
            []messages.Message{{Role: roles.User, Content: userInput}},
        )
        if err != nil {
            log.Printf("Error: %v\n", err)
            continue
        }

        fmt.Print("Agent: ")
        for chunk := range chunkChan {
            if chunk.Err != nil {
                log.Printf("\nError: %v", chunk.Err)
                break
            }
            fmt.Print(chunk.Content)
        }
        fmt.Println()
    }
}
```

**go.mod**:
```go
module my-chat-agent

go 1.25.4

require (
    github.com/snipwise/nova latest
)
```

#### Step 3: Test Locally

```bash
# Initialize module
go mod tidy

# Run locally
go run main.go
```

#### Step 4: Adapt for Docker (Environment Variables)

**main.go** (Docker-ready version):
```go
package main

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

// Helper function to get environment variables with defaults
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
    modelName := getEnv("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
    agentName := getEnv("AGENT_NAME", "chat-agent")
    systemInstructions := getEnv("SYSTEM_INSTRUCTIONS", "You are a helpful assistant.")

    agent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               agentName,
            EngineURL:          engineURL,
            SystemInstructions: systemInstructions,
        },
        models.Config{
            Name:        modelName,
            Temperature: models.Float64(0.7),
            MaxTokens:   models.Int(2000),
        },
    )
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    fmt.Printf("✅ Chat Agent Ready: %s (model: %s)\n", agentName, modelName)
    fmt.Println("Type 'exit' to quit")

    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("\nYou: ")
        if !scanner.Scan() {
            break
        }

        userInput := strings.TrimSpace(scanner.Text())
        if userInput == "exit" {
            break
        }

        chunkChan, err := agent.GenerateStreamingCompletion(
            []messages.Message{{Role: roles.User, Content: userInput}},
        )
        if err != nil {
            log.Printf("Error: %v\n", err)
            continue
        }

        fmt.Print("Agent: ")
        for chunk := range chunkChan {
            if chunk.Err != nil {
                log.Printf("\nError: %v", chunk.Err)
                break
            }
            fmt.Print(chunk.Content)
        }
        fmt.Println()
    }
}
```

#### Step 5: Create Dockerfile

**Dockerfile**:
```dockerfile
# === BUILD STAGE ===
FROM golang:1.25.5-alpine AS builder

WORKDIR /build

# Copy source code
COPY . .

# Download dependencies
RUN go mod download

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-agent .

# === RUNTIME STAGE ===
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /build/chat-agent .

# Run agent
CMD ["./chat-agent"]
```

#### Step 6: Create .dockerignore

**.dockerignore**:
```
.git
.gitignore
.vscode
.idea
*.swp
*_test.go
test/
tests/
README.md
*.md
tmp/
temp/
*.log
.DS_Store
Thumbs.db
```

#### Step 7: Create compose.yml

**compose.yml**:
```yaml
services:
  chat-agent:
    build:
      context: .
      dockerfile: Dockerfile

    # For interactive CLI
    stdin_open: true
    tty: true

    environment:
      # Application configuration
      AGENT_NAME: "helpful-assistant"
      SYSTEM_INSTRUCTIONS: "You are a helpful and concise assistant."

      # Auto-injected by Docker Agentic Compose:
      # ENGINE_URL: <auto>
      # CHAT_MODEL_ID: <auto>

    # Declare AI model dependencies
    models:
      chat-model:
        endpoint_var: ENGINE_URL
        model_var: CHAT_MODEL_ID

# Define AI models
models:
  chat-model:
    model: ai/qwen2.5:1.5B-F16
    # Optional: context_size: 32768
```

#### Step 8: Build and Run

```bash
# Build Docker image
docker compose build

# Start service
docker compose up -d

# Interact with agent
docker compose exec chat-agent ./chat-agent

# View logs
docker compose logs -f chat-agent

# Stop
docker compose down
```

## Code Adaptation Guide

### Pattern 1: Hard-Coded → Environment Variables

**Before (Hard-Coded):**
```go
engineURL := "http://localhost:12434/engines/llama.cpp/v1"
modelName := "ai/qwen2.5:1.5B-F16"
temperature := 0.7
```

**After (Environment Variables):**
```go
import "os"
import "strconv"

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
    if value := os.Getenv(key); value != "" {
        if f, err := strconv.ParseFloat(value, 64); err == nil {
            return f
        }
    }
    return defaultValue
}

// Usage
engineURL := getEnv("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
modelName := getEnv("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
temperature := getEnvFloat("TEMPERATURE", 0.7)
```

### Pattern 2: Hard-Coded Paths → Configurable Paths

**Before:**
```go
docsPath := "./docs"
storePath := "./store/embeddings.json"
outputPath := "./output/results.txt"
```

**After:**
```go
import "path/filepath"

docsPath := getEnv("DOCS_PATH", "./docs")
storePath := getEnv("STORE_PATH", "./store")
storeFile := getEnv("STORE_FILE", "embeddings.json")
storeFilePath := filepath.Join(storePath, storeFile)

outputPath := getEnv("OUTPUT_PATH", "./output")
outputFile := getEnv("OUTPUT_FILE", "results.txt")
outputFilePath := filepath.Join(outputPath, outputFile)
```

**compose.yml:**
```yaml
environment:
  DOCS_PATH: ./docs
  STORE_PATH: ./store
  STORE_FILE: embeddings.json
  OUTPUT_PATH: ./output
  OUTPUT_FILE: results.txt

volumes:
  - ./docs:/app/docs:ro
  - ./store:/app/store:rw
  - ./output:/app/output:rw
```

### Pattern 3: Multiple Models

**Before:**
```go
chatModel := "ai/qwen2.5:1.5B-F16"
embeddingModel := "ai/mxbai-embed-large"
compressorModel := "ai/qwen2.5:0.5B-F16"
```

**After:**
```go
chatModel := getEnv("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
embeddingModel := getEnv("EMBEDDING_MODEL_ID", "ai/mxbai-embed-large")
compressorModel := getEnv("COMPRESSOR_MODEL_ID", "ai/qwen2.5:0.5B-F16")
```

**compose.yml:**
```yaml
models:
  chat-model:
    endpoint_var: ENGINE_URL
    model_var: CHAT_MODEL_ID

  embedding-model:
    endpoint_var: ENGINE_URL
    model_var: EMBEDDING_MODEL_ID

  compressor-model:
    endpoint_var: ENGINE_URL
    model_var: COMPRESSOR_MODEL_ID

models:
  chat-model:
    model: ai/qwen2.5:1.5B-F16
  embedding-model:
    model: ai/mxbai-embed-large
  compressor-model:
    model: ai/qwen2.5:0.5B-F16
```

## Common Patterns

### Pattern: Chat Agent

See [docker-compose-simple.md](docker-compose-simple.md) - Template 1

### Pattern: RAG Agent with Persistence

See [docker-compose-simple.md](docker-compose-simple.md) - Template 2

### Pattern: Server Agent (HTTP API)

See [docker-compose-simple.md](docker-compose-simple.md) - Template 3

### Pattern: Multi-Agent Crew

See [docker-compose-complex.md](docker-compose-complex.md) - Template 1

### Pattern: Pipeline Agent

See [docker-compose-complex.md](docker-compose-complex.md) - Template 2

## Deployment Strategies

### Strategy 1: Single Host (Docker Compose)

**Best for**: Development, small production, demos

```bash
# Deploy
docker compose up -d

# Update
docker compose pull
docker compose up -d

# Rollback
docker compose down
docker compose up -d
```

### Strategy 2: Docker Swarm

**Best for**: Multi-host, medium-scale production

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c compose.yml agent-stack

# Scale service
docker service scale agent-stack_chat-agent=3

# Update
docker service update agent-stack_chat-agent

# Remove
docker stack rm agent-stack
```

### Strategy 3: Kubernetes

**Best for**: Large-scale, cloud-native production

```bash
# Convert compose to K8s
kompose convert -f compose.yml

# Apply to cluster
kubectl apply -f .

# Scale
kubectl scale deployment chat-agent --replicas=5

# Update
kubectl set image deployment/chat-agent chat-agent=myimage:v2
```

### Strategy 4: Cloud Platforms

#### AWS ECS (Elastic Container Service)

```bash
# Create ECR repository
aws ecr create-repository --repository-name chat-agent

# Build and push
docker build -t chat-agent .
docker tag chat-agent:latest <account>.dkr.ecr.<region>.amazonaws.com/chat-agent:latest
docker push <account>.dkr.ecr.<region>.amazonaws.com/chat-agent:latest

# Create ECS task definition and service (via AWS Console or CLI)
```

#### Azure Container Instances

```bash
# Create resource group
az group create --name agent-rg --location eastus

# Deploy container
az container create \
  --resource-group agent-rg \
  --name chat-agent \
  --image myregistry.azurecr.io/chat-agent:latest \
  --cpu 1 --memory 2 \
  --environment-variables ENGINE_URL=... CHAT_MODEL_ID=...
```

#### Google Cloud Run

```bash
# Build and push to GCR
gcloud builds submit --tag gcr.io/PROJECT-ID/chat-agent

# Deploy
gcloud run deploy chat-agent \
  --image gcr.io/PROJECT-ID/chat-agent \
  --platform managed \
  --set-env-vars ENGINE_URL=...,CHAT_MODEL_ID=...
```

## Production Checklist

### Security

- [ ] Remove `stdin_open` and `tty` for non-interactive services
- [ ] Run containers as non-root user
- [ ] Use Docker secrets for sensitive data
- [ ] Enable TLS for HTTP endpoints
- [ ] Implement authentication/authorization
- [ ] Set up firewall rules
- [ ] Scan images for vulnerabilities (`docker scan`)
- [ ] Use specific image tags (not `:latest`)

### Performance

- [ ] Set resource limits (CPU, memory)
- [ ] Configure health checks
- [ ] Enable logging with rotation
- [ ] Use multi-stage builds for small images
- [ ] Optimize model selection
- [ ] Implement caching strategies
- [ ] Add request queuing for high load
- [ ] Monitor with Prometheus/Grafana

### Reliability

- [ ] Set restart policies (`unless-stopped`)
- [ ] Add liveness/readiness probes
- [ ] Implement graceful shutdown
- [ ] Use persistent volumes for data
- [ ] Configure backups
- [ ] Set up alerting
- [ ] Test failure scenarios
- [ ] Document recovery procedures

### Observability

- [ ] Centralized logging (ELK, Splunk)
- [ ] Metrics collection (Prometheus)
- [ ] Distributed tracing (Jaeger)
- [ ] Dashboard visualization (Grafana)
- [ ] Error tracking (Sentry)
- [ ] Performance monitoring (APM)

## Troubleshooting

### Build Issues

#### Error: "go.mod not found"

```bash
# Initialize Go module
go mod init my-agent
go mod tidy
```

#### Error: "Cannot download dependencies"

```bash
# Inside Dockerfile, ensure proper network
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download
```

### Runtime Issues

#### Error: "ENGINE_URL not set"

```bash
# Check Docker Model Runner
docker model list

# Verify model declaration in compose.yml
models:
  chat-model:
    model: ai/qwen2.5:1.5B-F16
```

#### Error: "Connection refused to host.docker.internal:12434"

```bash
# macOS/Windows: Use host.docker.internal (auto-injected)
# Linux: Use --add-host=host.docker.internal:host-gateway

# Or in compose.yml
services:
  chat-agent:
    extra_hosts:
      - "host.docker.internal:host-gateway"
```

#### Error: "Permission denied on volume mount"

```bash
# Fix permissions on host
chmod -R 755 ./data ./store

# Or use named volumes
volumes:
  - agent-data:/app/data
```

### Model Issues

#### Error: "Model not found"

```bash
# Manual pull
docker model pull ai/qwen2.5:1.5B-F16

# Verify
docker model list
```

#### Error: "Out of memory"

```bash
# Reduce model size or increase container memory
services:
  chat-agent:
    deploy:
      resources:
        limits:
          memory: 4G
```

## Next Steps

1. **Explore Templates**: See [docker-compose-simple.md](docker-compose-simple.md) and [docker-compose-complex.md](docker-compose-complex.md)
2. **Learn Patterns**: Study the code adaptation patterns
3. **Practice**: Dockerize your existing Nova agents
4. **Deploy**: Choose appropriate deployment strategy
5. **Monitor**: Set up observability tools
6. **Scale**: Add replicas or services as needed

## Related Snippets

- [dockerfile-template.md](dockerfile-template.md) - Dockerfile templates
- [docker-compose-simple.md](docker-compose-simple.md) - Simple deployments
- [docker-compose-complex.md](docker-compose-complex.md) - Complex deployments

## Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Docker Agentic Compose](https://docs.docker.com/ai/compose/models-and-compose/)
- [Docker Model Runner](https://docs.docker.com/ai/model-runner/)
- [Nova SDK](https://github.com/snipwise/nova)
- [Go in Docker](https://docs.docker.com/language/golang/)
