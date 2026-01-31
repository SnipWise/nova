package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
			MaxTokens:   models.Int(2000),
		},
		// BeforeCompletion hook: called before each completion
		chat.BeforeCompletion(func(a *chat.Agent) {
			display.Info(">> [BeforeCompletion] Context size: " + conversion.IntToString(a.GetContextSize()))
			display.Info(">> [BeforeCompletion] Messages count: " + conversion.IntToString(len(a.GetMessages())))
		}),
		// AfterCompletion hook: called after each completion
		chat.AfterCompletion(func(a *chat.Agent) {
			display.Info("<< [AfterCompletion] Context size: " + conversion.IntToString(a.GetContextSize()))
			display.Info("<< [AfterCompletion] Messages count: " + conversion.IntToString(len(a.GetMessages())))
		}),
	)
	if err != nil {
		panic(err)
	}

	// === Standard Completion ===
	display.NewLine()
	display.Separator()
	display.Title("Standard Completion with BeforeCompletion / AfterCompletion hooks")
	display.Separator()

	result, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Hello, what is your name?"},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Response", result.Response)
	display.KeyValue("Finish reason", result.FinishReason)

	// === Stream Completion ===
	display.NewLine()
	display.Separator()
	display.Title("Stream Completion with BeforeCompletion / AfterCompletion hooks")
	display.Separator()

	fmt.Print("Response: ")
	streamResult, err := agent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: "Who is James T Kirk?"},
		},
		func(chunk string, finishReason string) error {
			fmt.Print(chunk)
			return nil
		},
	)
	fmt.Println()
	if err != nil {
		panic(err)
	}

	display.KeyValue("Finish reason", streamResult.FinishReason)

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("Both standard and streaming completions triggered the BeforeCompletion and AfterCompletion hooks.")
	display.Separator()
}
