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

// Predefined model configurations for common use cases
var (
	// Deterministic configuration for consistent outputs
	DeterministicConfig = models.NewConfig("ai/qwen2.5:1.5B-F16").
				WithTemperature(0.0).
				WithSeed(42)

	// Creative configuration for storytelling
	CreativeConfig = models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.9).
			WithTopP(0.95).
			WithPresencePenalty(0.6)

	// Balanced configuration for general use
	BalancedConfig = models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.7).
			WithMaxTokens(2000)

	// Concise configuration for brief responses
	ConciseConfig = models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.3).
			WithMaxTokens(500).
			WithFrequencyPenalty(0.5)
)

func main() {
	ctx := context.Background()
	engineConfig := agents.Config{
		EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
		SystemInstructions: "You are a helpful AI assistant.",
	}

	// Test 1: Deterministic agent
	display.Info("Test 1: Deterministic agent (temperature=0.0, seed=42)")
	deterministicAgent, err := chat.NewAgent(ctx, engineConfig, DeterministicConfig)
	if err != nil {
		panic(err)
	}

	result1, _ := deterministicAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Count from 1 to 5"},
	})
	display.KeyValue("Response", result1.Response)

	// Run the same query again - should get the same result
	result2, _ := deterministicAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Now count backwards from 5 to 1"},
	})
	display.KeyValue("Response", result2.Response)

	display.NewLine()
	display.Separator()

	// Test 2: Creative agent
	display.Info("Test 2: Creative agent (temperature=0.9)")
	creativeAgent, err := chat.NewAgent(ctx, engineConfig, CreativeConfig)
	if err != nil {
		panic(err)
	}

	result3, _ := creativeAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Tell me a creative idea for a story"},
	})
	display.KeyValue("Response", result3.Response)

	display.NewLine()
	display.Separator()

	// Test 3: Balanced agent
	display.Info("Test 3: Balanced agent (temperature=0.7)")
	balancedAgent, err := chat.NewAgent(ctx, engineConfig, BalancedConfig)
	if err != nil {
		panic(err)
	}

	result4, _ := balancedAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "What is artificial intelligence?"},
	})
	display.KeyValue("Response", result4.Response)

	display.NewLine()
	display.Separator()

	// Test 4: Concise agent
	display.Info("Test 4: Concise agent (temperature=0.3, max_tokens=500)")
	conciseAgent, err := chat.NewAgent(ctx, engineConfig, ConciseConfig)
	if err != nil {
		panic(err)
	}

	result5, _ := conciseAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Explain quantum computing in simple terms"},
	})
	display.KeyValue("Response", result5.Response)

	display.NewLine()
	display.Separator()
	display.Success("All configuration tests completed!")
}
