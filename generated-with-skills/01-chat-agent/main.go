package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	// === CONFIGURATION - CUSTOMIZE HERE ===
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "streaming-assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful and friendly assistant.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.7),
			MaxTokens:   models.Int(2000),
		},
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	fmt.Println("🤖 Streaming Chat Agent")
	fmt.Println("Type 'quit' to exit")
	fmt.Println(strings.Repeat("-", 40))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n👤 You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		fmt.Print("🤖 Assistant: ")

		// Streaming call with callback
		result, err := agent.GenerateStreamCompletion(
			[]messages.Message{
				{Role: roles.User, Content: input},
			},
			func(chunk string, finishReason string) error {
				// Called for each received chunk
				fmt.Print(chunk)
				return nil
			},
		)

		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		fmt.Println() // New line after response

		// Optional: Display metadata
		if result.FinishReason != "" {
			fmt.Printf("   [finish_reason: %s]\n", result.FinishReason)
		}
	}
}
