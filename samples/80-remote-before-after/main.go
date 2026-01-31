package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/remote"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	callCount := 0

	agent, err := remote.NewAgent(
		ctx,
		agents.Config{
			Name: "Remote Bob Client",
		},
		"http://localhost:3500",
		// BeforeCompletion hook: called before each completion
		remote.BeforeCompletion(func(a *remote.Agent) {
			callCount++
			display.Info(">> [BeforeCompletion] Agent: " + a.GetName() + " - Call #" + conversion.IntToString(callCount))
		}),
		// AfterCompletion hook: called after each completion
		remote.AfterCompletion(func(a *remote.Agent) {
			display.Info("<< [AfterCompletion] Agent: " + a.GetName() + " - Call #" + conversion.IntToString(callCount))
		}),
	)
	if err != nil {
		panic(err)
	}

	// Check server health
	if !agent.IsHealthy() {
		display.Error("Server is not available at http://localhost:8080")
		display.Info("Please start a server agent before running this sample.")
		return
	}

	display.Success("Server is healthy")

	// === Test 1: Streaming completion with hooks ===
	display.NewLine()
	display.Separator()
	display.Title("Streaming completion with BeforeCompletion / AfterCompletion hooks")
	display.Separator()

	result1, err := agent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: "Who is James T Kirk?"},
		},
		func(chunk string, finishReason string) error {
			fmt.Print(chunk)
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.KeyValue("Finish reason", result1.FinishReason)

	// === Test 2: Non-streaming completion with hooks ===
	display.NewLine()
	display.Separator()
	display.Title("Non-streaming completion (also triggers hooks via internal streaming)")
	display.Separator()

	result2, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Who is Spock?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result2.Response)
	display.KeyValue("Finish reason", result2.FinishReason)

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("Total completion calls: " + conversion.IntToString(callCount))
	display.Info("Both streaming and non-streaming completions triggered the hooks.")
	display.Info("Hooks are in GenerateStreamCompletion, which all other methods delegate to.")
	display.Separator()
}
