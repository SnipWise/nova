package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func TestSimpleChatAgent(t *testing.T) {

	ctx := context.Background()

	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
			MaxTokens:   models.Int(2000),
		},
	)
	if err != nil {
		panic(err)
	}

	display := func(result *chat.CompletionResult) {
		fmt.Println()
		fmt.Println("Response:\n", result.Response)
		fmt.Println()
		fmt.Println("Finish reason:\n", result.FinishReason)
		fmt.Println(strings.Repeat("-", 40))
	}

	// Simple chat using only Message structs
	result, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "[Brief] who is James T Kirk?"},
	})

	if err != nil {
		panic(err)
	}
	display(result)

	// Context is maintained automatically
	// Continue the conversation
	result, err = agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "[Brief] who is his best friend?"},
	})

	if err != nil {
		panic(err)
	}
	display(result)

}

// go test -v -run TestSimpleChatAgent ./getting-started/tests
