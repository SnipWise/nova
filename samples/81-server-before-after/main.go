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
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	callCount := 0

	// Create and set the tools agent
	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-tools-agent",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},
		tools.WithTools(getToolsIndex()),
	)
	if err != nil {
		panic(err)
	}

	// Create the server agent with lifecycle hooks
	agent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-server-agent",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.4),
		},
		server.WithPort(3500),
		server.WithToolsAgent(toolsAgent),
		server.WithExecuteFn(executeFunction),
		// BeforeCompletion hook: called before each completion (HTTP and CLI)
		server.BeforeCompletion(func(a *server.ServerAgent) {
			callCount++
			display.Info(">> [BeforeCompletion] Agent: " + a.GetName() + " - Call #" + conversion.IntToString(callCount))
		}),
		// AfterCompletion hook: called after each completion (HTTP and CLI)
		server.AfterCompletion(func(a *server.ServerAgent) {
			display.Info("<< [AfterCompletion] Agent: " + a.GetName() + " - Call #" + conversion.IntToString(callCount))
		}),
	)
	if err != nil {
		panic(err)
	}

	// Start the HTTP server
	fmt.Printf("Starting server agent on http://localhost%s\n", agent.GetPort())
	display.Info("Hooks will be triggered on each POST /completion request")
	log.Fatal(agent.StartServer())
}

func getToolsIndex() []*tools.Tool {
	calculateSumTool := tools.NewTool("calculate_sum").
		SetDescription("Calculate the sum of two numbers").
		AddParameter("a", "number", "The first number", true).
		AddParameter("b", "number", "The second number", true)

	sayHelloTool := tools.NewTool("say_hello").
		SetDescription("Say hello to the given name").
		AddParameter("name", "string", "The name to greet", true)

	return []*tools.Tool{
		calculateSumTool,
		sayHelloTool,
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	log.Printf("Executing function: %s with arguments: %s", functionName, arguments)

	switch functionName {
	case "say_hello":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for say_hello"}`, nil
		}
		hello := fmt.Sprintf("Hello, %s!", args.Name)
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

	default:
		return `{"error": "Unknown function"}`, errors.New("unknown function: " + functionName)
	}
}
