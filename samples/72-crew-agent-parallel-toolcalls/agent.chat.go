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

func CreateChatAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {
	chatModelID := env.GetEnvOrDefault("CHAT_MODEL_ID", "ai/qwen2.5:1.5B-F16")
	directivesPathFile := env.GetEnvOrDefault("CHAT_DIRECTIVES_PATH_FILE", "./directives/chat.agent.system.instructions.md")
	systemInstructions, err := files.ReadTextFile(directivesPathFile)
	if err != nil {
		return nil, err
	}

	chatAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "chat-agent",
			EngineURL:          engineURL,
			SystemInstructions: systemInstructions,
		},
		models.Config{
			Name:        chatModelID,
			Temperature: models.Float64(0.7),
		},
	)
	if err != nil {
		display.Errorf("❌ Error creating chat agent: %v", err)
		return nil, err
	}
	display.Infof("✅ Chat agent created")
	return chatAgent, nil
}
