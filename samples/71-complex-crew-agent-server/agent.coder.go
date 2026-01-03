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

func CreateCoderAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	temperature := env.GetEnvFloatOrDefault("CODER_TEMPERATURE", 0.8)
	topP := env.GetEnvFloatOrDefault("CODER_TOP_P", 0.9)

	coderAgentModelID := env.GetEnvOrDefault("CODER_MODEL_ID", "hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m")
	directivesPathFile := env.GetEnvOrDefault("CODER_DIRECTIVES_PATH_FILE", "./directives/coder.agent.system.instructions.md")
	coderAgentSystemInstructionsContent, err := files.ReadTextFile(directivesPathFile)
	if err != nil {
		return nil, err
	}

	// Create coder agent
	coderAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "coder",
			EngineURL:          engineURL,
			SystemInstructions: coderAgentSystemInstructionsContent,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        coderAgentModelID,
			Temperature: models.Float64(temperature),
			TopP:        models.Float64(topP),
		},
	)
	if err != nil {
		display.Errorf("❌ Error creating coder agent: %v", err)
		return nil, err
	}
	display.Infof("✅ Coder agent created for code generation")

	return coderAgent, nil
}
