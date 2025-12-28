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

	// Create agent with KeepConversationHistory set to true
	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: true, // Enable conversation history
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},
		tools.WithTools(GetToolsIndex()),
	)

	if err != nil {
		panic(err)
	}

	// First request - calculate sum
	display.NewLine()
	display.Separator()
	display.Title("First Request: Calculate sum of 40 and 2")
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

	// Check messages after first request
	messages1 := agent.GetMessages()
	display.NewLine()
	display.KeyValue("Messages count after first request", conversion.IntToString(len(messages1)))
	display.Info("Expected: Multiple messages (system + user + tool + assistant)")

	// Second request - say hello
	display.NewLine()
	display.Separator()
	display.Title("Second Request: Say hello to Alice")
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

	// Check messages after second request
	messages2 := agent.GetMessages()
	display.NewLine()
	display.KeyValue("Messages count after second request", conversion.IntToString(len(messages2)))
	display.Info("Expected: Even more messages (history accumulates)")

	// Display all messages to verify
	display.NewLine()
	display.Separator()
	display.Title("All Messages in History")
	display.Separator()
	for i, msg := range messages2 {
		contentPreview := msg.Content
		if len(contentPreview) > 60 {
			contentPreview = contentPreview[:60] + "..."
		}
		display.KeyValue("Message "+conversion.IntToString(i+1), string(msg.Role)+": "+contentPreview)
	}

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("With KeepConversationHistory=true, all messages should be kept.")
	display.Info("The history includes system, user, assistant, and tool messages.")
	display.Info("Each request has full context from previous interactions.")
	display.Separator()

	fmt.Println(agent.ExportMessagesToJSON())
	fmt.Println(agent.GetMessages())


}

func GetToolsIndex() []*tools.Tool {
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

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
