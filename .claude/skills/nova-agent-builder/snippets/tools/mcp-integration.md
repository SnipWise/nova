---
id: mcp-integration
name: MCP Tools Integration
category: tools
complexity: advanced
sample_source: 40-42
description: Agent that integrates Model Context Protocol (MCP) tools via stdio or HTTP
---

# MCP Tools Integration

## Description

Creates Nova agents that can use tools from MCP (Model Context Protocol) servers. Supports both stdio-based and HTTP-based MCP servers, allowing agents to leverage external tool ecosystems.

## Use Cases

- Integration with existing MCP tool servers
- Docker container MCP tools
- Remote tool execution
- Extending agent capabilities with MCP ecosystem
- Multi-server tool orchestration

## Complete Code

### Stdio MCP Client

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/mcptools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	// === CREATE STDIO MCP CLIENT ===
	// Connects to MCP server via stdio (e.g., Docker command)
	mcpClient, err := mcptools.NewStdioMCPClient(
		ctx,
		"docker",                          // Command to run
		[]string{},                        // Environment variables
		"mcp",                             // Additional args...
		"gateway",
		"run",
	)
	if err != nil {
		panic(err)
	}

	// List available tools from MCP server
	fmt.Println("Available MCP Tools:")
	for _, tool := range mcpClient.GetTools() {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}

	// === CREATE AGENT WITH MCP TOOLS ===
	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "mcp-agent",
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are an AI assistant with access to MCP tools.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true),
		},
		tools.WithMCPTools(mcpClient), // Add MCP tools
	)
	if err != nil {
		panic(err)
	}

	// === USE AGENT WITH MCP TOOLS ===
	result, err := agent.DetectToolCallsLoop(
		[]messages.Message{
			{Role: roles.User, Content: "Use the available MCP tools to help me"},
		},
		func(functionName, arguments string) (string, error) {
			fmt.Printf("Executing MCP tool: %s\n", functionName)
			return mcpClient.ExecTool(functionName, arguments)
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Assistant: %s\n", result.LastAssistantMessage)
}
```

### HTTP MCP Client

```go
package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/mcptools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	// === CREATE HTTP MCP CLIENT ===
	// Connects to MCP server via HTTP endpoint
	mcpClient, err := mcptools.NewStreamableHttpMCPClient(
		ctx,
		"http://localhost:9011", // MCP server URL
	)
	if err != nil {
		panic(err)
	}

	// List available tools
	fmt.Println("Available HTTP MCP Tools:")
	for _, tool := range mcpClient.GetTools() {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}

	// === EXECUTE TOOL DIRECTLY ===
	// You can execute MCP tools directly without an agent
	type ToolInput struct {
		Name string `json:"name"`
	}

	result, err := mcpClient.ExecToolWithAny(
		"hello_world_with_name",
		ToolInput{Name: "Alice"},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Direct tool execution:", result)

	// === CREATE AGENT WITH HTTP MCP TOOLS ===
	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "http-mcp-agent",
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are an AI assistant with HTTP MCP tools.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true),
		},
		tools.WithMCPTools(mcpClient),
	)
	if err != nil {
		panic(err)
	}

	// Use agent
	response, err := agent.DetectToolCallsLoop(
		[]messages.Message{
			{Role: roles.User, Content: "Say hello to Bob"},
		},
		func(functionName, arguments string) (string, error) {
			return mcpClient.ExecTool(functionName, arguments)
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Assistant: %s\n", response.LastAssistantMessage)
}
```

## Configuration

```yaml
# Stdio MCP
MCP_COMMAND: "docker"
MCP_ARGS: "mcp gateway run"

# HTTP MCP
MCP_SERVER_URL: "http://localhost:9011"

# Agent config
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TEMPERATURE: 0.0
PARALLEL_TOOL_CALLS: true
```

## Key API

### mcptools.NewStdioMCPClient

```go
import "github.com/snipwise/nova/nova-sdk/mcptools"

// Create stdio-based MCP client
mcpClient, err := mcptools.NewStdioMCPClient(
    ctx,
    "docker",              // Command to execute
    []string{},            // Environment variables
    "mcp", "gateway", "run", // Command arguments
)
```

### mcptools.NewStreamableHttpMCPClient

```go
// Create HTTP-based MCP client
mcpClient, err := mcptools.NewStreamableHttpMCPClient(
    ctx,
    "http://localhost:9011", // MCP server URL
)
```

### Adding MCP Tools to Agent

```go
import "github.com/snipwise/nova/nova-sdk/agents/tools"

agent, err := tools.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    tools.WithMCPTools(mcpClient), // Add all MCP tools
)
```

### Execute MCP Tool

```go
// Execute tool with JSON string
result, err := mcpClient.ExecTool(toolName, argumentsJSON)

// Execute tool with any struct
result, err := mcpClient.ExecToolWithAny(toolName, argumentsStruct)

// Get available tools
toolsList := mcpClient.GetTools()
```

## Customization

### Multiple MCP Servers

```go
// Create multiple MCP clients
dockerMCP, _ := mcptools.NewStdioMCPClient(ctx, "docker", []string{}, "mcp", "gateway", "run")
httpMCP, _ := mcptools.NewStreamableHttpMCPClient(ctx, "http://localhost:9011")

// Combine tools from multiple sources
localTools := []*tools.Tool{
    tools.NewTool("local_tool").SetDescription("A local tool"),
}

agent, err := tools.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    tools.WithTools(localTools),        // Local tools
    tools.WithMCPTools(dockerMCP),      // Docker MCP tools
    tools.WithMCPTools(httpMCP),        // HTTP MCP tools
)
```

### Environment Variable Configuration

```go
import "github.com/snipwise/nova/nova-sdk/toolbox/env"

func main() {
    ctx := context.Background()

    // Get MCP config from environment
    mcpServerURL := env.GetEnvOrDefault("MCP_SERVER_URL", "http://localhost:9011")

    mcpClient, err := mcptools.NewStreamableHttpMCPClient(ctx, mcpServerURL)
    if err != nil {
        panic(err)
    }

    // Rest of agent setup...
}
```

### Tool Selection and Filtering

```go
// Get all MCP tools
allTools := mcpClient.GetTools()

// Filter tools by name pattern
var selectedTools []*tools.Tool
for _, tool := range allTools {
    if strings.HasPrefix(tool.Name, "data_") {
        selectedTools = append(selectedTools, tool)
    }
}

// Use only selected tools
agent, err := tools.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    tools.WithTools(selectedTools),
)
```

## Docker MCP Server Setup

### docker-compose.yml

```yaml
version: '3.8'

services:
  mcp-server:
    build: ./mcp-server
    ports:
      - "9011:9011"
    environment:
      - MCP_PORT=9011
    restart: unless-stopped

  agent:
    build: .
    depends_on:
      - mcp-server
    environment:
      - MCP_SERVER_URL=http://mcp-server:9011
      - ENGINE_URL=http://host.docker.internal:12434/engines/llama.cpp/v1
```

## Important Notes

### DO:
- Use `mcptools.NewStdioMCPClient()` for local/Docker MCP servers
- Use `mcptools.NewStreamableHttpMCPClient()` for remote HTTP MCP servers
- Enable `ParallelToolCalls: true` to execute multiple MCP tools concurrently
- Set `Temperature: 0.0` for deterministic tool calling
- Check available tools with `mcpClient.GetTools()` before use
- Use `ExecToolWithAny()` for type-safe tool execution
- Combine MCP tools with local tools using multiple `WithTools()` calls

### DON'T:
- Don't forget to check MCP server availability before creating client
- Don't mix stdio and HTTP clients for the same server
- Don't ignore errors from `ExecTool()` - handle them gracefully
- Don't use high temperature for tool-calling agents
- Don't assume all MCP tools support parallel execution

## MCP Server Requirements

To use this integration, you need an MCP server running:

```bash
# For HTTP MCP server
cd samples/mcp-servers
docker compose up --build

# Verify server is running
curl http://localhost:9011/tools

# For stdio MCP (using Docker)
docker run -it mcp-gateway run
```

## Troubleshooting

### Connection Errors
- Verify MCP server is running: `curl http://localhost:9011/health`
- Check network connectivity in Docker: Use `host.docker.internal` instead of `localhost`
- Ensure correct port mapping in docker-compose.yml

### Tool Execution Failures
- Verify tool exists: `mcpClient.GetTools()`
- Check argument format: Must be valid JSON string or struct
- Review tool description for required parameters
- Enable logging: `os.Setenv("NOVA_LOG_LEVEL", "DEBUG")`

### Performance Issues
- Enable parallel tool calls: `ParallelToolCalls: true`
- Use HTTP MCP for better performance than stdio
- Monitor tool execution time
- Consider caching tool results if appropriate
