package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// Example 1: Simple model config with just the name
	display.Info("Example 1: Simple configuration")
	agent1, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful assistant.",
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16"),
	)
	if err != nil {
		panic(err)
	}
	result1, _ := agent1.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Say hello!"},
	})
	display.KeyValue("Response", result1.Response)

	display.NewLine()
	display.Separator()

	// Example 2: Full configuration with all parameters
	display.Info("Example 2: Full configuration")
	agent2, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a creative storyteller.",
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.9).
			WithTopP(0.95).
			WithMaxTokens(1000).
			WithFrequencyPenalty(0.5).
			WithPresencePenalty(0.3).
			WithSeed(42),
	)
	if err != nil {
		panic(err)
	}
	result2, _ := agent2.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Tell me a very short story."},
	})
	display.KeyValue("Response", result2.Response)

	display.NewLine()
	display.Separator()

	// Example 3: Deterministic output with seed
	display.Info("Example 3: Deterministic output (with seed)")
	agent3, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful assistant.",
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.0).
			WithSeed(123),
	)
	if err != nil {
		panic(err)
	}
	result3, _ := agent3.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Count from 1 to 5"},
	})
	display.KeyValue("Response", result3.Response)

	display.NewLine()
	display.Separator()

	// Example 4: Using pointer helpers directly
	display.Info("Example 4: Using ModelConfig struct directly")
	agent4, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful assistant.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float(0.5),
			MaxTokens:   models.Int(500),
		},
	)
	if err != nil {
		panic(err)
	}
	result4, _ := agent4.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "What is Go?"},
	})
	display.KeyValue("Response", result4.Response)

	display.NewLine()
	display.Separator()
	display.Success("All examples completed successfully!")
}
