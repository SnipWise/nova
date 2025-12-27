package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/remote"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"

	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)


func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// Create a remote agent that connects to the server
	agent, err := remote.NewAgent(
		ctx,
		agents.Config{
			Name: "Interactive Remote Bob Client",
		},
		"http://localhost:3500",
	)
	if err != nil {
		panic(err)
	}

	display.Colorf(display.ColorCyan, "🌐 Connected to remote agent at %s\n", "http://localhost:8080")
	display.Colorf(display.ColorCyan, "Agent: %s\n", agent.GetName())
	display.Colorf(display.ColorCyan, "Model: %s\n\n", agent.GetModelID())

	for {

		input := prompt.NewWithColor("🤖 Ask me something?").
			SetMessageColor(prompt.ColorBrightCyan).
			SetInputColor(prompt.ColorBrightWhite)

		question, err := input.Run()
		if err != nil {
			log.Fatal(err)
		}

		if strings.HasPrefix(question, "/bye") {
			fmt.Println("Goodbye!")
			break
		}

		result, err := agent.GenerateStreamCompletion(
			[]messages.Message{
				{
					Role:    roles.User,
					Content: question,
				},
			},
			func(chunk string, finishReason string) error {

				// Use markdown chunk parser for colorized streaming output
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
		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", result.FinishReason)
		display.KeyValue("Context size", fmt.Sprintf("%d characters", agent.GetContextSize()))
		display.Separator()

	}

}
