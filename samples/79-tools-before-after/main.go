package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	callCount := 0

	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},
		// ToolAgentOption: configure tools (existing option type)
		tools.WithTools(getToolsIndex()),
		// ToolsAgentOption: BeforeCompletion hook
		tools.BeforeCompletion(func(a *tools.Agent) {
			callCount++
			display.Info(">> [BeforeCompletion] Agent: " + a.GetName() + " (" + a.GetModelID() + ") - Call #" + conversion.IntToString(callCount))
			display.Info(">> [BeforeCompletion] Messages count: " + conversion.IntToString(len(a.GetMessages())))
		}),
		// ToolsAgentOption: AfterCompletion hook
		tools.AfterCompletion(func(a *tools.Agent) {
			display.Info("<< [AfterCompletion] Agent: " + a.GetName() + " (" + a.GetModelID() + ") - Call #" + conversion.IntToString(callCount))
			display.Info("<< [AfterCompletion] Messages count: " + conversion.IntToString(len(a.GetMessages())))
		}),
	)
	if err != nil {
		panic(err)
	}

	// === Test 1: DetectToolCallsLoop with hooks ===
	display.NewLine()
	display.Separator()
	display.Title("DetectToolCallsLoop with BeforeCompletion / AfterCompletion hooks")
	display.Separator()

	result1, err := agent.DetectToolCallsLoop(
		[]messages.Message{
			{Role: roles.User, Content: "Make the sum of 40 and 2"},
		},
		executeFunction,
	)
	if err != nil {
		panic(err)
	}

	display.KeyValue("Finish Reason", result1.FinishReason)
	for _, value := range result1.Results {
		display.KeyValue("Tool Result", value)
	}
	display.KeyValue("Assistant Message", result1.LastAssistantMessage)

	// === Test 2: Another tool call with hooks ===
	display.NewLine()
	display.Separator()
	display.Title("Another DetectToolCallsLoop call")
	display.Separator()

	result2, err := agent.DetectToolCallsLoop(
		[]messages.Message{
			{Role: roles.User, Content: "Say hello to Alice"},
		},
		executeFunction,
	)
	if err != nil {
		panic(err)
	}

	display.KeyValue("Finish Reason", result2.FinishReason)
	for _, value := range result2.Results {
		display.KeyValue("Tool Result", value)
	}
	display.KeyValue("Assistant Message", result2.LastAssistantMessage)

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("Total tool call detections: " + conversion.IntToString(callCount))
	display.Info("Both DetectToolCallsLoop calls triggered the BeforeCompletion and AfterCompletion hooks.")
	display.Separator()
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
	display.Colorf(display.ColorGreen, "Executing function: %s with arguments: %s\n", functionName, arguments)

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
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
