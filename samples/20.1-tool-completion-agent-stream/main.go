package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()
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

		tools.WithTools(GetToolsIndex()),
		// NEW: Set the tool execution function via option
		tools.WithExecuteFn(executeFunction),
	)
	if err != nil {
		panic(err)
	}
	// Say "Exit" to stop the process
	messages := []messages.Message{
		{
			Content: `
			Make the sum of 40 and 2,
			then say hello to Bob and to Sam,
			make the sum of 5 and 37
			Say hello to Alice
			`,
			Role: roles.User,
		},
	}

	// Stream callback for real-time content display
	streamCallback := func(content string) error {
		fmt.Print(content)
		return nil
	}

	display.Colorf(display.ColorCyan, "🚀 Starting streaming tool completion...\n")
	display.Separator()

	// NEW: Only streamCallback is required as parameter - executeFunction is set via option
	result, err := agent.DetectToolCallsLoopStream(messages, streamCallback)
	if err != nil {
		panic(err)
	}
	display.NewLine()
	display.Separator()

	display.KeyValue("Finish Reason", result.FinishReason)
	for _, value := range result.Results {
		display.KeyValue("Result for tool", value)
	}
	display.KeyValue("Assistant Message", result.LastAssistantMessage)

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

	display.Colorf(display.ColorGreen, "🟢 Detected function: %s with arguments: %s\n", functionName, arguments)

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

		// NOTE: Returning a message and an ExitToolCallsLoopError to stop further processing
		return fmt.Sprintf(`{"message": "%s"}`, "❌ EXIT"), errors.New("exit_loop")

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
