---
id: server-with-tools
name: Server Agent with Tools
category: server
complexity: intermediate
sample_source: 49
description: HTTP server agent with function calling capabilities and human-in-the-loop validation
---

# Server Agent with Tools

## Description

Creates an HTTP server agent with function calling (tools) capabilities. Tools are detected automatically from user requests and require human validation before execution via a web-based confirmation workflow.

## Use Cases

- API servers with external function calls
- Web applications requiring tool validation
- Microservices with controlled side effects
- Interactive assistants with actions
- Human-in-the-loop AI systems

## Complete Code

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// === SERVER AGENT CONFIGURATION ===
	agent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-server-agent",                              // Agent name
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",   // LLM Engine URL
			SystemInstructions: "You are Bob, a helpful AI assistant.",          // System instructions
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",    // Model for chat
			Temperature: models.Float64(0.4),
		},
		":3500",          // HTTP port
		executeFunction,  // Custom execute function for tools
	)
	if err != nil {
		panic(err)
	}

	// === TOOLS AGENT CONFIGURATION ===
	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-tools-agent",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",    // Model for tools
			Temperature:       models.Float64(0.0),                    // Deterministic for tools
			ParallelToolCalls: models.Bool(true),                      // Enable parallel execution
		},
		tools.WithTools(GetToolsIndex()),  // Register tools
	)
	if err != nil {
		panic(err)
	}

	// Attach tools agent to server agent
	agent.SetToolsAgent(toolsAgent)

	// Start the HTTP server
	fmt.Printf("🚀 Starting server agent on http://localhost%s\n", agent.GetPort())
	log.Fatal(agent.StartServer())
}

// === TOOL DEFINITIONS ===
func GetToolsIndex() []*tools.Tool {
	calculateSumTool := tools.NewTool("calculate_sum").
		SetDescription("Calculate the sum of two numbers").
		AddParameter("a", "number", "The first number", true).
		AddParameter("b", "number", "The second number", true)

	sayHelloTool := tools.NewTool("say_hello").
		SetDescription("Say hello to the given name").
		AddParameter("name", "string", "The name to greet", true)

	sayExit := tools.NewTool("say_exit").
		SetDescription("Say exit")

	return []*tools.Tool{
		calculateSumTool,
		sayHelloTool,
		sayExit,
	}
}

// === TOOL IMPLEMENTATIONS ===
func executeFunction(functionName string, arguments string) (string, error) {
	log.Printf("🟢 Executing function: %s with arguments: %s", functionName, arguments)

	switch functionName {
	case "say_hello":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for say_hello"}`, nil
		}
		hello := fmt.Sprintf("👋 Hello, %s!🙂", args.Name)
		return fmt.Sprintf(`{"message": "%s"}`, hello), nil

	case "calculate_sum":
		var args struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for calculate_sum"}`, nil
		}
		sum := args.A + args.B
		return fmt.Sprintf(`{"result": %g}`, sum), nil

	case "say_exit":
		return fmt.Sprintf(`{"message": "%s"}`, "❌ EXIT"), errors.New("exit_loop")

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
CHAT_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TOOLS_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TEMPERATURE_CHAT: 0.4
TEMPERATURE_TOOLS: 0.0
PARALLEL_TOOL_CALLS: true
PORT: ":3500"
```

## Tool Validation Workflow

### 1. Client sends request with tool intent
```bash
curl -N -X POST http://localhost:3500/completion \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "message": "Can you calculate 42 + 58 and say hello to Alice?"
    }
  }'
```

### 2. Server detects tools and sends notification (SSE)
```
data: {"type": "tool_notification", "operation_id": "abc123", "tools": [...]}
```

### 3. Client validates the operation
```bash
curl -X POST http://localhost:3500/operation/validate \
  -H "Content-Type: application/json" \
  -d '{"operation_id": "abc123"}'
```

### 4. Server executes tools and streams response
```
data: {"chunk": "I calculated..."}
data: {"finish_reason": "stop"}
```

## API Endpoints (Additional for Tools)

### POST /operation/validate
Approve pending tool operations.

```bash
curl -X POST http://localhost:3500/operation/validate \
  -H "Content-Type: application/json" \
  -d '{"operation_id": "operation-id-from-notification"}'
```

### POST /operation/cancel
Cancel pending tool operations.

```bash
curl -X POST http://localhost:3500/operation/cancel \
  -H "Content-Type: application/json" \
  -d '{"operation_id": "operation-id-from-notification"}'
```

### POST /operation/reset
Clear all pending operations.

```bash
curl -X POST http://localhost:3500/operation/reset
```

## Customization

### Adding More Tools

```go
func GetToolsIndex() []*tools.Tool {
	// Add your custom tools
	sendEmailTool := tools.NewTool("send_email").
		SetDescription("Send an email to a recipient").
		AddParameter("to", "string", "Email address", true).
		AddParameter("subject", "string", "Email subject", true).
		AddParameter("body", "string", "Email body", true)

	// Return all tools
	return []*tools.Tool{
		calculateSumTool,
		sayHelloTool,
		sendEmailTool,
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	switch functionName {
	case "send_email":
		var args struct {
			To      string `json:"to"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments"}`, nil
		}
		// Implement email sending logic
		return `{"status": "sent"}`, nil

	// Other cases...
	}
}
```

### Using Default ExecuteFunction

If you prefer using the tools agent's internal execution:

```go
// Omit executeFunction parameter
agent, err := server.NewAgent(
	ctx,
	agentConfig,
	modelConfig,
	":3500",
	// No executeFunction - uses default
)
```

### Disable Tool Validation (Auto-execute)

```go
// NOTE: Server agent always requires validation by default
// This is a safety feature for web-based systems
// To auto-execute, use a regular tools agent instead
```

## Client Example (JavaScript with SSE)

```javascript
const evtSource = new EventSource('http://localhost:3500/completion', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    data: { message: 'Calculate 10 + 20 and say hello to Bob' }
  })
});

evtSource.onmessage = (event) => {
  const data = JSON.parse(event.data);

  // Handle tool notification
  if (data.type === 'tool_notification') {
    console.log('Tools detected:', data.tools);

    // Auto-approve (or show UI to user)
    fetch('http://localhost:3500/operation/validate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ operation_id: data.operation_id })
    });
  }

  // Handle response chunks
  if (data.chunk) {
    console.log(data.chunk);
  }

  // Handle completion
  if (data.finish_reason) {
    evtSource.close();
  }
};
```

## Important Notes

- Tools require **human-in-the-loop validation** before execution
- Use `ParallelToolCalls: true` to execute multiple tools simultaneously
- Temperature should be `0.0` for tools agent (deterministic)
- The `executeFunction` parameter is now optional
- Tool notifications are sent via SSE before execution
- Always validate tool calls before they execute (security best practice)

## Error Handling

```go
func executeFunction(functionName string, arguments string) (string, error) {
	switch functionName {
	case "risky_operation":
		// Return error to prevent execution
		return "", fmt.Errorf("operation not allowed")

	case "safe_operation":
		// Return JSON error (execution considered successful)
		return `{"error": "Invalid input", "code": 400}`, nil
	}
}
```

## Related Patterns

- For basic server: See `basic-server.md`
- For RAG support: See `server-with-rag.md`
- For full-featured: See `server-full-featured.md`
