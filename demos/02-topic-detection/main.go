package main

import (
	"context"
	"strings"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"

	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
	"github.com/snipwise/nova/nova-sdk/ui/spinner"
)

type Intent struct {
	TopicDiscussion string `json:"topic_discussion"`
}

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		display.Warningf("No .env file found or error loading it: %v", err)
	}

	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
	orchestratorModel := env.GetEnvOrDefault("ORCHESTRATOR_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	orchestratorAgentSystemInstructions := env.GetEnvOrDefault("ORCHESTRATOR_AGENT_SYSTEM_INSTRUCTIONS", "")

	thinkingSpinner := spinner.NewWithColor("").SetSuffix("thinking...").SetFrames(spinner.FramesDots)
	thinkingSpinner.SetSuffixColor(spinner.ColorBrightYellow).SetFrameColor(spinner.ColorBrightYellow)

	agent, err := structured.NewAgent[Intent](
		ctx,
		agents.Config{
			Name:               "Orchestrator",
			EngineURL:          engineURL,
			SystemInstructions: orchestratorAgentSystemInstructions,
		},
		models.NewConfig(orchestratorModel).
			WithTemperature(0.0),
	)
	if err != nil {
		display.Errorf("failed to create agent: %v", err)
		return
	}

	for {

		input := prompt.NewWithColor("🤖 What would you like to talk about?")
		question, err := input.RunWithEdit()

		if err != nil {
			display.Errorf("failed to get input: %v", err)
			return
		}
		if strings.HasPrefix(question, "/bye") {
			display.Infof("👋 Goodbye!")
			break
		}

		thinkingSpinner.Start()

		response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
			{
				Role:    roles.User,
				Content: question,
			},
		})
		if err != nil {
			thinkingSpinner.Error("Failed!")
			display.Errorf("failed to get response: %v", err)
			return
		}
		thinkingSpinner.Success("Done!")

		display.NewLine()
		display.Title("Intent Detection")

		display.KeyValue("Topic", response.TopicDiscussion)
		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", finishReason)
		display.Separator()

	}

}
