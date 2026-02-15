package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func GetCoderAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {
	modelID := env.GetEnvOrDefault("CODER_MODEL_ID", "hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m")
	return chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "coder",
			EngineURL:               engineURL,
			SystemInstructions:      "You are an expert programming assistant. You write clean, efficient, and well-documented code.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.8),
		},
		chat.BeforeCompletion(func(agent *chat.Agent) {
			display.Styledln("🔧 [CODER AGENT] Processing request...", display.ColorCyan)
		}),
	)
}
