package main

import (
	"context"
	"fmt"
	"os"

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

	// Create a remote agent that connects to the server
	agent, err := remote.NewAgent(
		ctx,
		agents.Config{
			Name: "Remote Bob Client",
		},
		"http://localhost:8080", // Server URL
	)
	if err != nil {
		panic(err)
	}

	display.Colorf(display.ColorCyan, "🌐 Connected to remote agent at %s\n", "http://localhost:8080")
	display.Colorf(display.ColorCyan, "Agent: %s\n", agent.GetName())

	// Check server health
	if agent.IsHealthy() {
		display.Colorf(display.ColorGreen, "✅ Server is healthy\n")
	} else {
		display.Colorf(display.ColorRed, "❌ Server is not available\n")
		return
	}

	// Get models information
	modelsInfo, err := agent.GetModelsInfo()
	if err != nil {
		display.Colorf(display.ColorRed, "Failed to get models info: %v\n", err)
	} else {
		display.Colorf(display.ColorCyan, "Chat Model: %s\n", modelsInfo.ChatModel)
		display.Colorf(display.ColorCyan, "Embeddings Model: %s\n", modelsInfo.EmbeddingsModel)
		display.Colorf(display.ColorCyan, "Tools Model: %s\n", modelsInfo.ToolsModel)
	}
	fmt.Println()

	// Example 1: Simple streaming completion
	display.Colorf(display.ColorYellow, "=== Example 1: Simple Question ===\n")
	simpleQuestion := []messages.Message{
		{
			Content: "What is the capital of France?",
			Role:    roles.User,
		},
	}

	_, err = agent.GenerateStreamCompletion(simpleQuestion, func(chunk string, finishReason string) error {
		if chunk != "" {
			fmt.Print(chunk)
		}
		if finishReason == "stop" {
			fmt.Println()
		}
		return nil
	})
	if err != nil {
		display.Colorf(display.ColorRed, "Error: %v\n", err)
	}

	fmt.Println()
	display.Colorf(display.ColorGreen, "Context size: %d tokens\n\n", agent.GetContextSize())

	// Example 2: Question that triggers tool calls
	display.Colorf(display.ColorYellow, "=== Example 2: Tool Calls (with confirmation) ===\n")
	display.Colorf(display.ColorMagenta, "Note: Tool calls require manual validation.\n")
	display.Colorf(display.ColorMagenta, "Use the operation_id shown below to validate or cancel.\n\n")

	toolQuestion := []messages.Message{
		{
			Content: `
			Make the sum of 40 and 2,
			then say hello to Bob and to Sam,
			make the sum of 5 and 37,
			Say hello to Alice
			`,
			Role: roles.User,
		},
	}

	_, err = agent.GenerateStreamCompletion(toolQuestion, func(chunk string, finishReason string) error {
		if chunk != "" {
			fmt.Print(chunk)
		}
		if finishReason == "stop" {
			fmt.Println()
		}
		return nil
	})
	if err != nil {
		display.Colorf(display.ColorRed, "Error: %v\n", err)
	}

	fmt.Println()
	display.Colorf(display.ColorGreen, "Context size: %d tokens\n\n", agent.GetContextSize())

	// Example 3: Non-streaming completion
	display.Colorf(display.ColorYellow, "=== Example 3: Non-Streaming Completion ===\n")
	nonStreamQuestion := []messages.Message{
		{
			Content: "What is 2 + 2?",
			Role:    roles.User,
		},
	}

	result, err := agent.GenerateCompletion(nonStreamQuestion)
	if err != nil {
		display.Colorf(display.ColorRed, "Error: %v\n", err)
	} else {
		display.Colorf(display.ColorGreen, "Response: %s\n", result.Response)
		display.Colorf(display.ColorGreen, "Finish Reason: %s\n", result.FinishReason)
	}

	fmt.Println()
	display.Colorf(display.ColorGreen, "Context size: %d tokens\n\n", agent.GetContextSize())

	// Display conversation history
	display.Colorf(display.ColorYellow, "=== Conversation History ===\n")
	messages := agent.GetMessages()
	for i, msg := range messages {
		display.Colorf(display.ColorCyan, "[%d] %s: ", i+1, msg.Role)
		if len(msg.Content) > 100 {
			fmt.Printf("%s...\n", msg.Content[:100])
		} else {
			fmt.Printf("%s\n", msg.Content)
		}
	}

	// Export conversation to JSON
	fmt.Println()
	display.Colorf(display.ColorYellow, "=== Export Conversation ===\n")
	jsonData, err := agent.ExportMessagesToJSON()
	if err != nil {
		display.Colorf(display.ColorRed, "Error exporting: %v\n", err)
	} else {
		display.Colorf(display.ColorGreen, "Conversation exported to JSON:\n%s\n", jsonData)
	}

	// Reset conversation
	fmt.Println()
	display.Colorf(display.ColorYellow, "=== Reset Conversation ===\n")
	agent.ResetMessages()
	display.Colorf(display.ColorGreen, "Conversation reset. Context size: %d tokens\n", agent.GetContextSize())
}
