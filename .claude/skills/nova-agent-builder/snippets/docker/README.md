# Docker & Deployment Snippets

This directory contains templates and guides for **dockerizing and deploying Nova agents** using Docker and Docker Agentic Compose.

## 📦 Available Snippets

### 1. [dockerfile-template.md](dockerfile-template.md)
**Multi-stage Dockerfile template for Nova agents**

- **Use when**: Containerizing any Nova agent
- **Features**:
  - Multi-stage build (builder + runtime)
  - Alpine-based minimal image (~15-25 MB)
  - Security best practices
  - Customization examples for all agent types
- **Output**: Optimized, production-ready Dockerfile

### 2. [docker-compose-simple.md](docker-compose-simple.md)
**Docker Compose configurations for simple agents**

- **Use when**: Deploying chat, RAG, tools, structured, or server agents
- **Templates**:
  1. Single chat agent
  2. RAG agent with persistence
  3. Server agent (HTTP API)
  4. Multiple agents in parallel
- **Features**: Docker Agentic Compose, YAML anchors, volume management

### 3. [docker-compose-complex.md](docker-compose-complex.md)
**Docker Compose configurations for complex multi-agent systems**

- **Use when**: Deploying crew, pipeline, or orchestrator agents
- **Templates**:
  1. Crew agent with orchestrator
  2. Pipeline agent
  3. Multi-service crew (distributed)
  4. RAG + Crew + Compressor (full-featured)
- **Features**: Multi-service orchestration, inter-service communication, advanced networking

### 4. [dockerization-guide.md](dockerization-guide.md)
**Complete step-by-step dockerization guide**

- **Use when**: First-time dockerization or need comprehensive tutorial
- **Content**:
  - Prerequisites and installation
  - Step-by-step tutorial (chat agent example)
  - Code adaptation patterns (hard-coded → env vars)
  - Deployment strategies (Compose, Swarm, K8s, Cloud)
  - Production checklist
  - Troubleshooting

## 🚀 Quick Start

### Basic Workflow

1. **Choose your agent type**:
   - Simple agent (chat/RAG/tools/server) → Use `docker-compose-simple.md`
   - Complex agent (crew/pipeline) → Use `docker-compose-complex.md`

2. **Generate Dockerfile**:
   - Use template from `dockerfile-template.md`
   - Customize for your agent type

3. **Generate compose.yml**:
   - Copy appropriate template
   - Adapt environment variables
   - Configure models

4. **Adapt code**:
   - Replace hard-coded values with `os.Getenv()` or `env.GetEnvOrDefault()`
   - Make paths configurable

5. **Build and run**:
   ```bash
   docker compose build
   docker compose up -d
   ```

## 🐳 What is Docker Agentic Compose?

**Docker Agentic Compose** (Docker Compose v2.38.0+) is an extension that treats **AI models as first-class resources**:

### Key Features

1. **Declarative models**: Define AI models in `compose.yml`
2. **Auto-injection**: Environment variables automatically injected
3. **Lifecycle management**: Docker Model Runner handles model lifecycle
4. **Portability**: Same config across all environments

### Example

```yaml
# Define models globally
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

**Result**: Docker automatically injects:
- `ENGINE_URL=http://host.docker.internal:12434/engines/llama.cpp/v1`
- `CHAT_MODEL_ID=ai/qwen2.5:1.5B-F16`

## 📋 Prerequisites

### Required Software

- **Docker Desktop 4.36+** (includes Docker Compose 2.38+)
  - Download: https://www.docker.com/products/docker-desktop
- **Docker Model Runner** (included in Docker Desktop)
  - Verify: `docker model --version`
- **Go 1.25.4+**
  - Download: https://go.dev/dl/

### Required Models

Models are auto-pulled on first run. To pre-download:

```bash
# Essential models
docker model pull ai/qwen2.5:1.5B-F16              # Chat (1.5B)
docker model pull ai/mxbai-embed-large             # Embeddings
docker model pull hf.co/menlo/jan-nano-gguf:q4_k_m # Tools (160M)
docker model pull ai/qwen2.5:0.5B-F16              # Compressor (0.5B)

# Verify
docker model list
```

## 🎯 Common Use Cases

### Use Case 1: Dockerize a Chat Agent

**User request**: "Dockerize my chat agent"

**Steps**:
1. Read [dockerfile-template.md](dockerfile-template.md)
2. Read [docker-compose-simple.md](docker-compose-simple.md) - Template 1
3. Adapt main.go with environment variables
4. Generate Dockerfile, compose.yml, .dockerignore
5. Provide build/run instructions

### Use Case 2: Dockerize a RAG Agent with Persistence

**User request**: "Create docker compose for my RAG agent with persistent store"

**Steps**:
1. Read [dockerfile-template.md](dockerfile-template.md)
2. Read [docker-compose-simple.md](docker-compose-simple.md) - Template 2
3. Configure volumes for documents and embeddings store
4. Adapt code for configurable paths
5. Generate complete setup

### Use Case 3: Dockerize a Multi-Agent Crew

**User request**: "Deploy my content creation crew to production"

**Steps**:
1. Read [dockerfile-template.md](dockerfile-template.md)
2. Read [docker-compose-complex.md](docker-compose-complex.md) - Template 1
3. Configure multiple models (orchestrator, research, writer, editor)
4. Set up inter-agent communication
5. Generate production-ready deployment

### Use Case 4: Complete Dockerization Guide

**User request**: "How do I dockerize my Nova agent?"

**Steps**:
1. Direct user to [dockerization-guide.md](dockerization-guide.md)
2. Walk through prerequisites
3. Follow step-by-step tutorial
4. Learn code adaptation patterns
5. Choose deployment strategy

## 🔧 Code Adaptation Patterns

### Pattern 1: Hard-Coded → Environment Variables

**Before**:
```go
engineURL := "http://localhost:12434/engines/llama.cpp/v1"
modelName := "ai/qwen2.5:1.5B-F16"
```

**After**:
```go
import "github.com/snipwise/nova/nova-sdk/toolbox/env"

engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
modelName := env.GetEnvOrDefault("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
```

### Pattern 2: Hard-Coded Paths → Configurable

**Before**:
```go
storePath := "./store/embeddings.json"
```

**After**:
```go
import (
    "path/filepath"
    "github.com/snipwise/nova/nova-sdk/toolbox/env"
)

storePath := env.GetEnvOrDefault("STORE_PATH", "./store")
storeFile := env.GetEnvOrDefault("STORE_FILE", "embeddings.json")
storeFilePath := filepath.Join(storePath, storeFile)
```

### Pattern 3: Flush Output for Docker Visibility

**Critical for Docker**: Add `os.Stdout.Sync()` to ensure all output is visible in Docker logs.

**Problem**: Docker buffers stdout/stderr, causing output truncation.

**Solution**:
```go
import "os"

// After important output (conversation history, JSON export)
fmt.Println(jsonData)
os.Stdout.Sync()  // Force flush to make visible in Docker logs

// Before program exit
os.Stdout.Sync()
os.Stderr.Sync()
```

**When to use**:
- After loops that print output (conversation history)
- After JSON export or large data dumps
- Before program termination
- After critical status messages

**Example**: See sample 67 for complete implementation.

## 📚 Deployment Strategies

### Local Development
```bash
docker compose up -d
```

### Multi-Host (Docker Swarm)
```bash
docker swarm init
docker stack deploy -c compose.yml agent-stack
```

### Kubernetes
```bash
kompose convert -f compose.yml
kubectl apply -f .
```

### Cloud Platforms
- **AWS**: ECS, Fargate, EKS
- **Azure**: Container Instances, AKS
- **GCP**: Cloud Run, GKE

## 🐛 Troubleshooting

### Issue: "ENGINE_URL not set"
- **Cause**: Docker Model Runner not running or model not declared
- **Solution**: Check `docker model list` and verify model declaration in `compose.yml`

### Issue: "Permission denied on volume mount"
- **Cause**: File permission mismatch
- **Solution**: `chmod -R 755 ./data ./store` or use named volumes

### Issue: "Port already in use"
- **Cause**: Another service using same port
- **Solution**: Change port mapping in `compose.yml` or stop conflicting service

### Issue: "Missing output in Docker logs (truncated output)"
- **Cause**: Docker buffers stdout/stderr, causing output to be truncated or invisible
- **Solution**: Add `os.Stdout.Sync()` after important output sections
- **Example**: See Pattern 3 in Code Adaptation section and sample 67

## 📖 Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Docker Agentic Compose](https://docs.docker.com/ai/compose/models-and-compose/)
- [Docker Model Runner](https://docs.docker.com/ai/model-runner/)
- [Nova SDK](https://github.com/snipwise/nova)

## 🎓 Learning Path

1. **Start here**: [dockerization-guide.md](dockerization-guide.md)
2. **Simple agents**: [docker-compose-simple.md](docker-compose-simple.md)
3. **Complex systems**: [docker-compose-complex.md](docker-compose-complex.md)
4. **Optimization**: [dockerfile-template.md](dockerfile-template.md)
5. **Production**: Apply deployment strategies and checklist

## ✅ Best Practices

### DO:
- ✅ Use multi-stage builds for small images
- ✅ Use environment variables for all configuration
- ✅ Mount volumes for persistent data
- ✅ Use specific model versions (not `:latest`)
- ✅ Add health checks for server agents
- ✅ Set resource limits to prevent overconsumption
- ✅ Use YAML anchors for reusable configs
- ✅ **Add `os.Stdout.Sync()` after important output (Docker buffering fix)**
- ✅ Flush stdout/stderr before program exit

### DON'T:
- ❌ Hard-code values in application code
- ❌ Ignore .dockerignore file
- ❌ Run containers as root in production
- ❌ Expose unnecessary ports
- ❌ Skip health checks for critical services
- ❌ Use `:latest` tag in production
- ❌ Forget to document build/run commands

## 🎯 Summary

This directory provides **everything needed to dockerize Nova agents**:
- **Templates**: Ready-to-use Dockerfile and compose.yml
- **Guides**: Step-by-step tutorials and best practices
- **Patterns**: Code adaptation examples
- **Strategies**: Deployment options for all scenarios

Choose the appropriate snippet based on your agent type and deployment needs!
