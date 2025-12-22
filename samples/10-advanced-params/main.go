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
	engineConfig := agents.Config{
		EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
		SystemInstructions: "You are a helpful AI assistant.",
	}

	// Example 1: Using TopK and MinP for more controlled sampling
	display.Info("Example 1: TopK and MinP sampling")
	agent1, err := chat.NewAgent(
		ctx,
		engineConfig,
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.8).
			WithTopK(40).   // Only consider top 40 most likely tokens
			WithMinP(0.05). // Minimum probability threshold
			WithMaxTokens(100),
	)
	if err != nil {
		panic(err)
	}

	result1, _ := agent1.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "What is machine learning?"},
	})
	display.KeyValue("Response", result1.Response)
	display.NewLine()
	display.Separator()

	// Example 2: Using RepeatPenalty to reduce repetition
	display.Info("Example 2: Repeat penalty")
	agent2, err := chat.NewAgent(
		ctx,
		engineConfig,
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.7).
			WithRepeatPenalty(1.2). // Penalize repeated tokens
			WithMaxTokens(150),
	)
	if err != nil {
		panic(err)
	}

	result2, _ := agent2.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Tell me about artificial intelligence"},
	})
	display.KeyValue("Response", result2.Response)
	display.NewLine()
	display.Separator()

	// Example 3: Using Stop sequences
	display.Info("Example 3: Stop sequences")
	agent3, err := chat.NewAgent(
		ctx,
		engineConfig,
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.7).
			WithStop(".", "!", "?"). // Stop at sentence end
			WithMaxTokens(100),
	)
	if err != nil {
		panic(err)
	}

	result3, _ := agent3.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Complete this: The future of AI is"},
	})
	display.KeyValue("Response", result3.Response)
	display.KeyValue("Finish reason", result3.FinishReason)
	display.NewLine()
	display.Separator()

	// Example 4: Full configuration with all parameters
	display.Info("Example 4: Complete configuration")
	agent4, err := chat.NewAgent(
		ctx,
		engineConfig,
		models.Config{
			Name:             "ai/qwen2.5:1.5B-F16",
			Temperature:      models.Float(0.7),
			TopP:             models.Float(0.9),
			TopK:             models.Int(50),
			MinP:             models.Float(0.05),
			MaxTokens:        models.Int(200),
			FrequencyPenalty: models.Float(0.3),
			PresencePenalty:  models.Float(0.3),
			RepeatPenalty:    models.Float(1.1),
			Seed:             models.Int(123),
			Stop:             []string{"\n\n", "END"},
			N:                models.Int(1),
		},
	)
	if err != nil {
		panic(err)
	}

	result4, _ := agent4.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Explain quantum computing briefly"},
	})
	display.KeyValue("Response", result4.Response)
	display.NewLine()
	display.Separator()

	display.Success("All advanced parameter examples completed!")
	display.Info("Note: Some parameters like TopK, MinP, RepeatPenalty may be model-specific")
}
