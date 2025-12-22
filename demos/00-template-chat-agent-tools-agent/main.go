package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

//  make the sum of 40 and 2 and say hello to bob

func main() {
	ctx := context.Background()

	// Create a simple chat chatAgent
	chatAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.4),
		},
	)
	if err != nil {
		panic(err)
	}

	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},

		models.NewConfig("hf.co/menlo/jan-nano-gguf:q4_k_m").
			WithTemperature(0.0).
			WithParallelToolCalls(true),

		tools.WithTools(GetToolsIndex()),
	)

	if err != nil {
		panic(err)
	}

	for {

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("tell me something: ")
		question, _ := reader.ReadString('\n')

		toolCallsResult, err := toolsAgent.DetectParallelToolCallsWithConfirmation(
			[]messages.Message{
				{Role: roles.User, Content: question},
			},
			executeFunction,
			confirmationPrompt,
		)
		if err != nil {
			panic(err)
		}


		if len(toolCallsResult.Results) > 0 {
			chatAgent.AddMessage(roles.System, toolCallsResult.LastAssistantMessage)
		} 

		_, err = chatAgent.GenerateStreamCompletion(
			[]messages.Message{
				{Role: roles.User, Content: question},
			},
			func(chunk string, finishReason string) error {
				if chunk != "" {
					fmt.Print(chunk)
				}
				if finishReason == "stop" {
					fmt.Println()
				}
				return nil
			},
		)
		if err != nil {
			panic(err)
		}
	}

}

func confirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {

	fmt.Printf("🟢 Detected function: %s with arguments: %s\n", functionName, arguments)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("y/n/q: ")
	choice, _ := reader.ReadString('\n')

	switch choice {
	case "q":
		return tools.Quit
	case "n":
		return tools.Denied
	case "y":
		return tools.Confirmed
	default:
		return tools.Denied
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

	case "say_exit":

		// NOTE: Returning a message and an ExitToolCallsLoopError to stop further processing
		return fmt.Sprintf(`{"message": "%s"}`, "❌ EXIT"), errors.New("exit_loop")

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
