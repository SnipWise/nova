package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func GetToolsAgent(ctx context.Context, engineURL string) (*tools.Agent, error) {
	modelID := env.GetEnvOrDefault("TOOLS_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")

	getToolsIndex := func() []*tools.Tool {

		calculateSumTool := tools.NewTool("calculate_sum").
			SetDescription("Calculate the sum of two numbers").
			AddParameter("a", "number", "The first number", true).
			AddParameter("b", "number", "The second number", true)

		sayHelloTool := tools.NewTool("say_hello").
			SetDescription("Say hello to the given name").
			AddParameter("name", "string", "The name to greet", true)

		klingonGreetingTool := tools.NewTool("klingon_greeting").
			SetDescription("Greet someone in Klingon").
			AddParameter("name", "string", "The name to greet in Klingon", true)

		return []*tools.Tool{
			calculateSumTool,
			sayHelloTool,
			klingonGreetingTool,
		}
	}

	executeFunction := func(functionName string, arguments string) (string, error) {

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

		case "klingon_greeting":
			var args struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal([]byte(arguments), &args); err != nil {
				return `{"error": "Invalid arguments for klingon_greeting"}`, nil
			}
			greeting := fmt.Sprintf("nuqneH, %s! Qapla'! 🖖", args.Name)
			return fmt.Sprintf(`{"message": "%s"}`, greeting), nil

		default:
			return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
		}
	}

	return tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "tools",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a helpful assistant that can use tools when needed.",
			KeepConversationHistory: false, // Tools agent doesn't need history
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.0),
		},
		tools.BeforeCompletion(func(agent *tools.Agent) {
			display.Styledln("[TOOLS] Detecting tool calls...", display.ColorYellow)
		}),
		tools.WithTools(getToolsIndex()),
		tools.WithExecuteFn(executeFunction),
	)

}
