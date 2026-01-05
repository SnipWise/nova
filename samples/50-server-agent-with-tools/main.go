package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	// Enable logging
	if err := os.Setenv("NOVA_LOG_LEVEL", "INFO"); err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Create the tools agent
	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "Bob Tools",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true),
		}, // IMPORTANT: Enable parallel tool calls
		tools.WithTools(GetToolsIndex()),
	)
	if err != nil {
		panic(err)
	}


	// Create the server agent
	serverAgent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:               "Bob",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.4),
		},
		server.WithPort(8080),
		server.WithToolsAgent(toolsAgent),
		server.WithExecuteFn(executeFunction),
	)
	if err != nil {
		panic(err)
	}


	display.Colorf(display.ColorCyan, "🚀 Server starting on http://localhost%s\n", serverAgent.GetPort())
	display.Colorf(display.ColorYellow, "📡 Endpoints:\n")
	display.Colorf(display.ColorYellow, "  POST   /completion\n")
	display.Colorf(display.ColorYellow, "  POST   /completion/stop\n")
	display.Colorf(display.ColorYellow, "  POST   /memory/reset\n")
	display.Colorf(display.ColorYellow, "  GET    /memory/messages/list\n")
	display.Colorf(display.ColorYellow, "  GET    /memory/messages/context-size\n")
	display.Colorf(display.ColorYellow, "  POST   /operation/validate\n")
	display.Colorf(display.ColorYellow, "  POST   /operation/cancel\n")
	display.Colorf(display.ColorYellow, "  POST   /operation/reset\n")
	display.Colorf(display.ColorYellow, "  GET    /models\n")
	display.Colorf(display.ColorYellow, "  GET    /health\n")

	// Start the server
	if err := serverAgent.StartServer(); err != nil {
		panic(err)
	}
}

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

func executeFunction(functionName string, arguments string) (string, error) {
	display.Colorf(display.ColorGreen, "🟢 Executing function: %s with arguments: %s\n", functionName, arguments)

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
		// NOTE: Returning a message and an error to stop further processing
		return fmt.Sprintf(`{"message": "%s"}`, "❌ EXIT"), errors.New("exit_loop")

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
