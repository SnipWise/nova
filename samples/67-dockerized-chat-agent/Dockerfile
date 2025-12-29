# === STAGE 1: BUILD ===
# Use official Go image with Alpine for minimal size
FROM golang:1.25.5-alpine AS builder

# Set working directory for build
WORKDIR /build

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download Go module dependencies
# This layer is cached unless go.mod/go.sum change
RUN go mod download

# Copy application source code
COPY . .

# Build the application binary
# CGO_ENABLED=0: Static binary (no C dependencies)
# GOOS=linux: Target Linux OS
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-agent .

# === STAGE 2: RUNTIME ===
# Use minimal Alpine image for runtime
FROM alpine:latest

# Set working directory for runtime
WORKDIR /app

# Install CA certificates for HTTPS requests
# Required for connecting to LLM servers
RUN apk --no-cache add ca-certificates

# Copy compiled binary from builder stage
COPY --from=builder /build/chat-agent .

# Optional: Set environment variables with defaults
# These will be overridden by docker-compose.yml or Docker Agentic Compose
ENV NOVA_LOG_LEVEL=INFO

# Run the agent binary
# Note: CMD can be overridden in docker-compose.yml or docker run
CMD ["./chat-agent"]
