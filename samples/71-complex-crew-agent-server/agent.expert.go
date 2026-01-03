package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func CreateExpertAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	temperature := env.GetEnvFloatOrDefault("EXPERT_TEMPERATURE", 0.8)
	topP := env.GetEnvFloatOrDefault("EXPERT_TOP_P", 0.9)

	expertAgentModelID := env.GetEnvOrDefault("EXPERT_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	directivesPathFile := env.GetEnvOrDefault("EXPERT_DIRECTIVES_PATH_FILE", "./directives/expert.agent.system.instructions.md")
	expertAgentSystemInstructionsContent, err := files.ReadTextFile(directivesPathFile)
	if err != nil {
		return nil, err
	}

	// Create expert agent
	expertAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "expert",
			EngineURL:          engineURL,
			SystemInstructions: expertAgentSystemInstructionsContent,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        expertAgentModelID,
			Temperature: models.Float64(temperature),
			TopP:        models.Float64(topP),
		},
	)
	if err != nil {
		display.Errorf("❌ Error creating expert agent: %v", err)
		return nil, err
	}
	display.Infof("✅ Expert agent created for expert assistance")

	return expertAgent, nil
}
