package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
)

func getCoderAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	coderAgentSystemInstructionsContent := `
        You are an expert programming assistant. You write clean, efficient, and well-documented code. Always:
        - Provide complete, working code
        - Include error handling
        - Add helpful comments
        - Follow best practices for the language
        - Explain your approach briefly
	`

	coderAgentModel := env.GetEnvOrDefault("CODER_AGENT_MODEL", "hf.co/quantfactory/deepseek-coder-7b-instruct-v1.5-gguf:q4_k_m")
	coderAgentModelTemperatureStr := env.GetEnvOrDefault("CODER_AGENT_MODEL_TEMPERATURE", "0.8")
	coderAgentModelTemperature := conversion.StringToFloat(coderAgentModelTemperatureStr)
	coderAgentSystemInstructions := env.GetEnvOrDefault("CODER_AGENT_SYSTEM_INSTRUCTIONS", coderAgentSystemInstructionsContent)

	coderAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "coder",
			EngineURL:          engineURL,
			SystemInstructions: coderAgentSystemInstructions,
		},

		models.NewConfig(coderAgentModel).
			WithTemperature(coderAgentModelTemperature),
	)
	if err != nil {
		return nil, err
	}

	return coderAgent, nil
}

func getThinkerAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	thinkerAgentSystemInstructionsContent := `
        You are a thoughtful conversational assistant. 
        - Listen carefully to the user
        - Think before responding
        - Ask clarifying questions when needed
        - Discuss topics with curiosity and respect
        - Admit when you don't know something
        Keep responses natural and conversational.	
	`

	thinkerModel := env.GetEnvOrDefault("THINKER_MODEL", "hf.co/menlo/lucy-gguf:q4_k_m")
	thinkerModelTemperatureStr := env.GetEnvOrDefault("THINKER_MODEL_TEMPERATURE", "0.8")
	thinkerModelTemperature := conversion.StringToFloat(thinkerModelTemperatureStr)
	thinkerAgentSystemInstructions := env.GetEnvOrDefault("THINKER_AGENT_SYSTEM_INSTRUCTIONS", thinkerAgentSystemInstructionsContent)

	thinkerAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "thinker",
			EngineURL:          engineURL,
			SystemInstructions: thinkerAgentSystemInstructions,
		},
		models.NewConfig(thinkerModel).
			WithTemperature(thinkerModelTemperature),
	)
	if err != nil {
		return nil, err
	}

	return thinkerAgent, nil
}

func getGenericAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	genericAgentSystemInstructionsContent := `
        You respond appropriately to different types of questions.
        For factual questions: Give direct answers with key facts
        For how-to questions: Provide step-by-step guidance
        For opinion questions: Present balanced perspectives
        For complex topics: Break into digestible parts

        Always start with the most important information.	
	`

	genericModel := env.GetEnvOrDefault("GENERIC_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	genericModelTemperatureStr := env.GetEnvOrDefault("GENERIC_MODEL_TEMPERATURE", "0.8")
	genericModelTemperature := conversion.StringToFloat(genericModelTemperatureStr)
	genericAgentSystemInstructions := env.GetEnvOrDefault("GENERIC_MODEL_SYSTEM_INSTRUCTIONS", genericAgentSystemInstructionsContent)

	genericAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "generic",
			EngineURL:          engineURL,
			SystemInstructions: genericAgentSystemInstructions,
		},
		models.NewConfig(genericModel).
			WithTemperature(genericModelTemperature),
	)
	if err != nil {
		return nil, err
	}

	return genericAgent, nil
}

func (ca *CompositeAgent) initializeAgenticSquad(ctx context.Context, engineURL string) error {

	coderAgent, err := getCoderAgent(ctx, engineURL)
	if err != nil {
		return err
	}
	ca.chatAgents[coderAgent.GetName()] = coderAgent

	thinkerAgent, err := getThinkerAgent(ctx, engineURL)
	if err != nil {
		return err
	}
	ca.chatAgents[thinkerAgent.GetName()] = thinkerAgent

	genericAgent, err := getGenericAgent(ctx, engineURL)
	if err != nil {
		return err
	}
	ca.chatAgents[genericAgent.GetName()] = genericAgent

	/*
		teamAgents := map[string]*chat.Agent{
			"coder":      coderAgent,
			"thinker":    thinkerAgent,
			"cooking":    cookingAgent,
			"translator": tranlatorAgent,
			"generic":    genericAgent,
		}
	*/
	return nil

}
