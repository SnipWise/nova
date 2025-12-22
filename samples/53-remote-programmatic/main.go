package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/remote"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// Create a remote agent
	agent, err := remote.NewAgent(
		ctx,
		agents.Config{
			Name: "Programmatic Remote Client",
		},
		"http://localhost:8080",
	)
	if err != nil {
		panic(err)
	}

	display.Colorf(display.ColorCyan, "🌐 Connected to remote agent\n\n")

	display.Colorf(display.ColorYellow, "=== Programmatic Operation Management Demo ===\n\n")

	// Example 1: Auto-validate all operations
	display.Colorf(display.ColorYellow, "Example 1: Auto-validate all tool calls\n")
	autoValidateExample(agent)

	time.Sleep(2 * time.Second)

	// Reset conversation
	agent.ResetMessages()
	display.Colorf(display.ColorGreen, "\n✅ Conversation reset\n\n")

	// Clear the callback from example 1
	agent.SetToolCallCallback(nil)

	// Example 2: Cancel all operations
	display.Colorf(display.ColorYellow, "Example 2: Cancel all tool calls\n")
	cancelAllExample(agent)
}

func autoValidateExample(agent *remote.Agent) {
	// Track detected operations
	pendingOps := make(chan string, 10)
	validationDone := make(chan bool)

	// Set up tool call callback to capture operation IDs
	agent.SetToolCallCallback(func(operationID string, message string) error {
		// Send operation ID to validation goroutine
		pendingOps <- operationID
		return nil
	})

	// Goroutine to auto-validate operations as they appear
	go func() {
		for opID := range pendingOps {
			display.Colorf(display.ColorGreen, "🤖 Auto-validating operation: %s\n", opID)
			time.Sleep(200 * time.Millisecond) // Small delay for demonstration
			if err := agent.ValidateOperation(opID); err != nil {
				display.Colorf(display.ColorRed, "❌ Validation error: %v\n", err)
			} else {
				display.Colorf(display.ColorGreen, "✅ Operation validated successfully\n")
			}
		}
		validationDone <- true
	}()

	// Send question
	question := []messages.Message{
		{
			Content: "Say hello to Alice and Bob, then calculate 10 + 20",
			Role:    roles.User,
		},
	}

	display.Colorf(display.ColorMagenta, "Question: %s\n\n", question[0].Content)
	display.Colorf(display.ColorCyan, "🤖 Auto-validation enabled: Operations will be approved automatically\n\n")

	// Start streaming
	_, err := agent.GenerateStreamCompletion(question, func(chunk string, finishReason string) error {
		if chunk != "" {
			fmt.Print(chunk)
		}
		if finishReason == "stop" {
			fmt.Println()
		}
		return nil
	})

	if err != nil {
		display.Colorf(display.ColorRed, "Stream error: %v\n", err)
	}

	// Close the channel and wait for validation to complete
	close(pendingOps)
	<-validationDone

	display.Colorf(display.ColorGreen, "\n✅ Auto-validation complete\n\n")
}

func cancelAllExample(agent *remote.Agent) {
	// Send question
	question := []messages.Message{
		{
			Content: "Calculate 5 + 10, say hello to Sam, and calculate 20 + 30",
			Role:    roles.User,
		},
	}

	display.Colorf(display.ColorMagenta, "Question: %s\n\n", question[0].Content)
	display.Colorf(display.ColorCyan, "🛑 Strategy: Cancel all operations immediately\n\n")

	// Set up callback to auto-cancel operations as they appear
	agent.SetToolCallCallback(func(operationID string, message string) error {
		display.Colorf(display.ColorYellow, "🛑 Auto-cancelling operation: %s\n", operationID)
		return agent.CancelOperation(operationID)
	})

	// Start streaming - operations will be cancelled as they appear
	_, err := agent.GenerateStreamCompletion(question, func(chunk string, finishReason string) error {
		if chunk != "" {
			fmt.Print(chunk)
		}
		if finishReason == "stop" {
			fmt.Println()
		}
		return nil
	})

	if err != nil {
		display.Colorf(display.ColorRed, "Stream error: %v\n", err)
	}

	display.Colorf(display.ColorGreen, "\n✅ All operations were cancelled\n")
}
