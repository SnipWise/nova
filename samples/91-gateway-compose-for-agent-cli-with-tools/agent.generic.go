package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func GetGenericAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {
	modelID := env.GetEnvOrDefault("GENERIC_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	return chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "generic",
			EngineURL:               engineURL,
			SystemInstructions:      "You respond appropriately to different types of questions. Always start with the most important information.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.8),
		},
		chat.BeforeCompletion(func(agent *chat.Agent) {
			display.Styledln("💬 [GENERIC AGENT] Processing request...", display.ColorGreen)
		}),
	)
}
