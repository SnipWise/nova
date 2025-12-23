package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func TestParallelToolCallsWithConfirmationAgent(t *testing.T) {
	ctx := context.Background()

	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
			ParallelToolCalls: models.Bool(true), // IMPORTANT: Enable parallel tool calls
		},

		tools.WithTools(getConfirmationToolsIndex()),
	)

	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

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

	result, err := agent.DetectParallelToolCallsWithConfirmation(
		messages,
		executeConfirmationFunction,
		confirmationPrompt)

	if err != nil {
		t.Fatalf("DetectParallelToolCallsWithConfirmation failed: %v", err)
	}

	// Display results
	fmt.Println("Finish Reason:", result.FinishReason)
	for _, value := range result.Results {
		fmt.Println("Result for tool:", value)
	}
	fmt.Println("Assistant Message:", result.LastAssistantMessage)

	// Verify we got some results
	if len(result.Results) == 0 {
		t.Error("Expected at least one tool result")
	}
}

func confirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
	fmt.Printf("🟢 Detected function: %s with arguments: %s\n", functionName, arguments)

	// For automated testing, we auto-approve all tool calls
	// In a real application, you would use prompt.HumanConfirmation here
	fmt.Printf("Auto-approving execution of %s with %s\n", functionName, arguments)
	return tools.Confirmed
}

func getConfirmationToolsIndex() []*tools.Tool {
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

func executeConfirmationFunction(functionName string, arguments string) (string, error) {
	fmt.Printf("🟢 Executing function: %s with arguments: %s\n", functionName, arguments)

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

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}

// go test -v -run TestParallelToolCallsWithConfirmationAgent ./getting-started/tests
