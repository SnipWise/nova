package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/mcptools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

/*
To execute this sample, make sure you have an MCP server running locally.

```
cd ../mcp-servers
docker compose up --build
```
*/

func main() {
	ctx := context.Background()

	mcpClient, err := mcptools.NewStreamableHttpMCPClient(ctx, "http://localhost:9011")

	if err != nil {
		panic(err)
	}

	// Print available tools
	for _, tool := range mcpClient.GetTools() {
		println("Tool:", tool.Name, "-", tool.Description)
	}

	fmt.Println(strings.Repeat("=", 50))

	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name: "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:        models.Float64(0.0),
			ParallelToolCalls:  models.Bool(false),
		},

		tools.WithMCPTools(mcpClient.GetTools()),
	)

	if err != nil {
		panic(err)
	}

	messages := []messages.Message{
		{
			Content: `
			Say hello to Alice
			Say hello to Bob Morane
			`,
			Role: roles.User,
		},
	}

	executeFunction := func(functionName, arguments string) (string, error) {
		display.Colorf(display.ColorGreen, "🟢 Executing function: %s with arguments: %s\n", functionName, arguments)
		switch functionName {
		case "hello_world_with_name":
			type input struct {
				Name string `json:"name"`
			}
			params, err := conversion.FromJSON[input](arguments)
			if err != nil {
				return "", err
			}
			result, err := mcpClient.ExecToolWithAny("hello_world_with_name", params)

			if err != nil {
				return "", err
			}
			return result, nil
		default:
			return "", fmt.Errorf("unknown function: %s", functionName)
		}
	}

	result, err := agent.DetectToolCallsLoop(messages, executeFunction)
	if err != nil {
		panic(err)
	}
	display.KeyValue("Finish Reason", result.FinishReason)

	for _, value := range result.Results {
		display.KeyValue("Result for tool", value)
	}

	display.KeyValue("Assistant Message", result.LastAssistantMessage)

}
