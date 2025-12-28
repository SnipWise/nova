---
id: simple-tools
name: Simple Tools Agent
category: tools
complexity: intermediate
sample_source: 18
description: Agent with function calling capability to execute tools
---

# Simple Tools Agent

## Description

Creates an agent capable of using tools (function calling) to perform actions like calculations, web searches, API calls, etc.

## Use Cases

- Agents needing external actions
- Calculator assistants
- Systems with API integrations
- Automation bots

## Complete Code

```go
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	// === DEFINE TOOLS ===
	availableTools := []*tools.Tool{
		tools.NewTool("get_current_time").
			SetDescription("Get the current date and time").
			AddParameter("timezone", "string", "Timezone (e.g., 'UTC', 'America/New_York')", false),

		tools.NewTool("calculate").
			SetDescription("Perform a mathematical calculation").
			AddParameter("expression", "string", "Mathematical expression to evaluate", true),

		tools.NewTool("get_weather").
			SetDescription("Get current weather for a city").
			AddParameter("city", "string", "City name", true).
			AddParameter("units", "string", "Units: 'celsius' or 'fahrenheit'", false),
	}

	// === CREATE AGENT ===
	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "tools-assistant",
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are a helpful assistant with access to tools. Use them when needed to answer user questions.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m", // Model supporting tools
			Temperature: models.Float64(0.0),                 // 0 for determinism
		},
		tools.WithTools(availableTools),
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	fmt.Println("🛠️ Tools Agent")
	fmt.Println("Available tools: get_current_time, calculate, get_weather")
	fmt.Println("Type 'quit' to exit")
	fmt.Println(strings.Repeat("-", 40))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n👤 You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "quit" {
			break
		}

		// Detect and execute tools in a loop
		result, err := agent.DetectToolCallsLoop(
			[]messages.Message{{Role: roles.User, Content: input}},
			executeTool, // Tool execution function
		)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Display tool results
		if len(result.Results) > 0 {
			fmt.Println("🔧 Tools executed:")
			for _, r := range result.Results {
				fmt.Printf("   - %s\n", r)
			}
		}

		fmt.Printf("🤖 Assistant: %s\n", result.LastAssistantMessage)
	}
}

// === TOOL EXECUTION ===
func executeTool(name string, argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}

	switch name {
	case "get_current_time":
		tz := "UTC"
		if v, ok := args["timezone"].(string); ok && v != "" {
			tz = v
		}
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return fmt.Sprintf(`{"error": "invalid timezone: %s"}`, tz), nil
		}
		now := time.Now().In(loc)
		return fmt.Sprintf(`{"time": "%s", "timezone": "%s"}`, 
			now.Format("2006-01-02 15:04:05"), tz), nil

	case "calculate":
		expr, _ := args["expression"].(string)
		// Simple implementation - use a real library in production
		result := evalSimpleExpr(expr)
		return fmt.Sprintf(`{"expression": "%s", "result": %s}`, expr, result), nil

	case "get_weather":
		city, _ := args["city"].(string)
		units := "celsius"
		if v, ok := args["units"].(string); ok && v != "" {
			units = v
		}
		// Simulated response - use real API in production
		return fmt.Sprintf(`{"city": "%s", "temperature": 22, "units": "%s", "condition": "sunny"}`,
			city, units), nil

	default:
		return fmt.Sprintf(`{"error": "unknown tool: %s"}`, name), nil
	}
}

func evalSimpleExpr(expr string) string {
	// Simplified implementation
	// Use github.com/Knetic/govaluate for real expressions
	return "42"
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"  # Must support function calling
TEMPERATURE: 0.0  # Deterministic for tools
```

## Customization

### Tools with Enum

```go
tools.NewTool("set_priority").
    SetDescription("Set task priority").
    AddEnumParameter("level", "string", "Priority level", 
        []string{"low", "medium", "high", "critical"}, true)
```

### Tool Calling Without Loop

```go
// Single detection (without automatic execution)
result, err := agent.DetectToolCalls(messages)
if err != nil {
    // Handle error
}

for _, call := range result.ToolCalls {
    fmt.Printf("Tool: %s\n", call.Name)
    fmt.Printf("Args: %s\n", call.Arguments)
    
    // Manual execution
    output, _ := executeTool(call.Name, call.Arguments)
    fmt.Printf("Result: %s\n", output)
}
```

### With API Integration

```go
func executeTool(name string, argsJSON string) (string, error) {
    switch name {
    case "search_web":
        var args struct {
            Query string `json:"query"`
        }
        json.Unmarshal([]byte(argsJSON), &args)
        
        // Real API call
        resp, err := http.Get(fmt.Sprintf(
            "https://api.search.com/search?q=%s", 
            url.QueryEscape(args.Query),
        ))
        if err != nil {
            return `{"error": "search failed"}`, nil
        }
        defer resp.Body.Close()
        
        body, _ := io.ReadAll(resp.Body)
        return string(body), nil
    }
    return `{"error": "unknown tool"}`, nil
}
```

## Important Notes

- Model must support function calling
- Temperature 0.0 recommended for tool consistency
- Return JSON format for results
- Handle errors in tool execution gracefully
- `DetectToolCallsLoop` handles multi-step tool chains
