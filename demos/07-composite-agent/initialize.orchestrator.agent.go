package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
)

type Intent struct {
	TopicDiscussion string `json:"topic_discussion"`
}

func (ca *CompositeAgent) initializeOrchestratorAgent(ctx context.Context, engineURL string) error {

	orchestratorAgentSystemInstructionsContent := `
        You are good at identifying the topic of a conversation.
        Given a user's input, identify the main topic of discussion in only one word.
        The possible topics are: Technology, Health, Sports, Entertainment, Politics, Science, Mathematics,
        Travel, Food, Education, Finance, Environment, Fashion, History, Literature, Art,
        Music, Psychology, Relationships, Philosophy, Religion, Automotive, Gaming, Translation.
        Respond in JSON format with the field 'topic_discussion'.
	`

	orchestratorModel := env.GetEnvOrDefault("ORCHESTRATOR_MODEL", "hf.co/menlo/lucy-gguf:q4_k_m")
	orchestratorAgentSystemInstructions := env.GetEnvOrDefault("ORCHESTRATOR_AGENT_SYSTEM_INSTRUCTIONS", orchestratorAgentSystemInstructionsContent)

	agent, err := structured.NewAgent[Intent](
		ctx,
		agents.Config{
			Name:               "orchestrator-agent",
			EngineURL:          engineURL,
			SystemInstructions: orchestratorAgentSystemInstructions,
		},
		models.Config{
			Name:        orchestratorModel,
			Temperature: models.Float64(0.0),
		},
	)
	if err != nil {
		return err
	}

	ca.orchestratorAgent = agent

	return nil
}
