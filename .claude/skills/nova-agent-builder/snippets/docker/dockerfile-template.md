# Docker - Dockerfile Template for Nova Agents

This snippet provides a **multi-stage Dockerfile** template for containerizing any Nova agent application.

## Category
**Docker & Deployment**

## Use Case
- Containerize Nova agents for production deployment
- Create lightweight, portable Docker images
- Support multi-service deployments with Docker Compose
- Enable cloud deployment (AWS, Azure, GCP, Kubernetes)

## Multi-Stage Dockerfile Benefits

1. **Lightweight images**: Alpine runtime (~5MB) instead of golang (~300MB)
2. **Security**: No development tools in final image
3. **Performance**: Faster startup, less bandwidth
4. **Efficient cache**: Optimized Docker layers for rebuilds

## Dockerfile Template

**CRITICAL Requirements**:
- Use Go 1.25.5-alpine image (supports Go 1.25.4 requirement)
- Ensure go.mod specifies `go 1.25.4`
- Use `github.com/snipwise/nova latest` (not local replace)

```dockerfile
# === STAGE 1: BUILD ===
# Use official Go image with Alpine for minimal size
FROM golang:1.25.5-alpine AS builder

# Set working directory for build
WORKDIR /build

# Copy application source code
# Note: Copy from context directory (where docker build runs)
COPY . .

# Download Go module dependencies
# This layer is cached unless go.mod/go.sum change
RUN go mod download

# Build the application binary
# CGO_ENABLED=0: Static binary (no C dependencies)
# GOOS=linux: Target Linux OS
RUN CGO_ENABLED=0 GOOS=linux go build -o agent-binary .

# === STAGE 2: RUNTIME ===
# Use minimal Alpine image for runtime
FROM alpine:latest

# Set working directory for runtime
WORKDIR /app

# Install CA certificates for HTTPS requests
# Required for connecting to LLM servers
RUN apk --no-cache add ca-certificates

# Copy compiled binary from builder stage
COPY --from=builder /build/agent-binary .

# Optional: Copy additional resources (docs, config files, etc.)
# COPY --from=builder /build/docs ./docs
# COPY --from=builder /build/config ./config

# Optional: Create directories for persistent data
# RUN mkdir -p /app/data /app/logs

# Optional: Expose port if agent runs as server
# EXPOSE 8080

# Optional: Set environment variables
# ENV NOVA_LOG_LEVEL=INFO

# Run the agent binary
# Note: CMD can be overridden in docker-compose.yml or docker run
CMD ["./agent-binary"]
```

## Build Instructions

### Build Image Locally

```bash
# Navigate to agent directory
cd /path/to/your-agent

# Build Docker image
docker build -t your-agent-name:latest .

# Build with custom Dockerfile name
docker build -f Dockerfile.custom -t your-agent-name:latest .

# Build with no cache (force rebuild)
docker build --no-cache -t your-agent-name:latest .
```

### Run Container

```bash
# Run interactively
docker run -it your-agent-name:latest

# Run in detached mode (background)
docker run -d your-agent-name:latest

# Run with environment variables
docker run -e ENGINE_URL=http://host.docker.internal:12434/engines/llama.cpp/v1 \
           -e MODEL_NAME=ai/qwen2.5:1.5B-F16 \
           your-agent-name:latest

# Run with volume mount for persistence
docker run -v $(pwd)/data:/app/data your-agent-name:latest

# Run with port mapping (for server agents)
docker run -p 8080:8080 your-agent-name:latest
```

## Customization Guide

### 1. For Chat/RAG/Tools Agents (CLI)

```dockerfile
# ... (build stage same as above)

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /build/agent-binary .

# For interactive CLI agents
CMD ["./agent-binary"]
```

### 2. For Server Agents (HTTP API)

```dockerfile
# ... (build stage same as above)

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /build/agent-binary .

# Expose HTTP port
EXPOSE 8080

# Run server
CMD ["./agent-binary"]
```

### 3. For Agents with External Resources

```dockerfile
# ... (build stage same as above)

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /build/agent-binary .

# Copy resource directories
COPY --from=builder /build/docs ./docs
COPY --from=builder /build/prompts ./prompts

# Create data directories
RUN mkdir -p /app/data /app/logs /app/store

CMD ["./agent-binary"]
```

### 4. For RAG Agents with Persistent Store

```dockerfile
# ... (build stage same as above)

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /build/agent-binary .

# Create store directory for embeddings
RUN mkdir -p /app/store /app/documents

# Volume will be mounted here for persistence
VOLUME ["/app/store", "/app/documents"]

CMD ["./agent-binary"]
```

## Docker Best Practices

### ✅ DO:

1. **Use multi-stage builds** to keep images small
2. **Install only required dependencies** in runtime stage
3. **Use .dockerignore** to exclude unnecessary files
4. **Set proper file permissions** for security
5. **Use specific image tags** instead of :latest in production
6. **Add health checks** for server agents
7. **Use environment variables** for configuration
8. **Create volumes** for persistent data

### ❌ DON'T:

1. **Don't include development tools** in runtime image
2. **Don't hard-code secrets** in Dockerfile
3. **Don't run as root** (add USER directive for security)
4. **Don't copy unnecessary files** (.git, tests, etc.)
5. **Don't use :latest tag** in production
6. **Don't ignore .dockerignore** file

## .dockerignore Template

Create a `.dockerignore` file in your project root:

```
# Git
.git
.gitignore

# IDE
.vscode
.idea
*.swp

# Build artifacts
*.exe
*.dll
*.so
*.dylib

# Test files
*_test.go
test/
tests/

# Documentation
README.md
docs/
*.md

# Temporary files
tmp/
temp/
*.tmp

# Logs
*.log
logs/

# OS files
.DS_Store
Thumbs.db
```

## Advanced Dockerfile Features

### With Health Check (for server agents)

```dockerfile
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates curl

COPY --from=builder /build/agent-binary .

EXPOSE 8080

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./agent-binary"]
```

### With Non-Root User (security)

```dockerfile
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

COPY --from=builder /build/agent-binary .

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

CMD ["./agent-binary"]
```

### With Build Arguments

```dockerfile
FROM golang:1.25.5-alpine AS builder

# Accept build arguments
ARG GO_VERSION=1.25.5
ARG BUILD_DATE
ARG VERSION

WORKDIR /build
COPY . .

RUN go mod download

# Inject build info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE}" \
    -o agent-binary .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /build/agent-binary .

# Add labels
LABEL version="${VERSION}"
LABEL build-date="${BUILD_DATE}"

CMD ["./agent-binary"]
```

Build with arguments:
```bash
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  -t your-agent:1.0.0 .
```

## Image Size Optimization

### Typical Sizes:

- **Without multi-stage**: ~350 MB (includes Go toolchain)
- **With multi-stage (Alpine)**: ~15-25 MB (binary + Alpine)
- **With scratch**: ~10-15 MB (binary only, no OS)

### Using scratch (minimal image):

```dockerfile
FROM golang:1.25.5-alpine AS builder
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o agent-binary .

# Minimal scratch image (no OS)
FROM scratch

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /build/agent-binary /agent-binary

CMD ["/agent-binary"]
```

**Note**: `scratch` image has no shell or utilities, making debugging difficult. Use Alpine for most cases.

## Testing the Docker Image

```bash
# Build image
docker build -t test-agent:latest .

# Check image size
docker images test-agent:latest

# Inspect image layers
docker history test-agent:latest

# Run container interactively
docker run -it --rm test-agent:latest

# Run with environment override
docker run -it --rm \
  -e ENGINE_URL=http://host.docker.internal:12434/engines/llama.cpp/v1 \
  -e MODEL_NAME=ai/qwen2.5:1.5B-F16 \
  test-agent:latest

# Check logs
docker logs <container-id>

# Execute command in running container
docker exec -it <container-id> /bin/sh
```

## Integration with go.mod

Ensure your `go.mod` is properly configured:

```go
module your-agent-name

go 1.25.4

require (
    github.com/snipwise/nova latest
)
```

Before building Docker image, update dependencies:
```bash
go mod tidy
go mod download
```

## Next Steps

After creating your Dockerfile:

1. **Build the image**: `docker build -t your-agent .`
2. **Test locally**: `docker run -it your-agent`
3. **Create docker-compose.yml**: See `docker-compose-simple.md` or `docker-compose-complex.md`
4. **Push to registry**: For deployment (Docker Hub, ECR, GCR, etc.)

## Related Snippets

- [docker-compose-simple.md](docker-compose-simple.md) - Single agent deployment
- [docker-compose-complex.md](docker-compose-complex.md) - Multi-agent deployment
- [dockerization-guide.md](dockerization-guide.md) - Complete guide

## Resources

- [Docker Multi-Stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/build-images/)
- [Alpine Linux](https://alpinelinux.org/)
