package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func GetClientSideToolsAgent(ctx context.Context, engineURL string) (*tools.Agent, error) {
	modelID := env.GetEnvOrDefault("CLIENT_SIDE_TOOLS_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	return tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "client-side-tools",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a helpful assistant that can use tools when needed.",
			KeepConversationHistory: false, // Tools agent doesn't need history
		},
		models.Config{
			Name:        modelID,
			Temperature: models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},
		tools.BeforeCompletion(func(agent *tools.Agent) {
			display.Styledln("🔀 [CLIENT-SIDE TOOLS] Detecting tool calls...", display.ColorYellow)
		}),
	)
}
