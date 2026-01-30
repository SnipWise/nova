package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func CreateToolsAgent(ctx context.Context, engineURL string) (*tools.Agent, error) {
	toolsModelID := env.GetEnvOrDefault("TOOLS_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	directivesPathFile := env.GetEnvOrDefault("TOOLS_DIRECTIVES_PATH_FILE", "./directives/tools.agent.system.instructions.md")
	systemInstructions, err := files.ReadTextFile(directivesPathFile)
	if err != nil {
		return nil, err
	}

	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:                    "tools-agent",
			EngineURL:               engineURL,
			SystemInstructions:      systemInstructions,
			KeepConversationHistory: false,
		},
		models.Config{
			Name:              toolsModelID,
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true), // Enable parallel tool calls
		},
		tools.WithTools(GetToolsIndex()),
	)
	if err != nil {
		display.Errorf("❌ Error creating tools agent: %v", err)
		return nil, err
	}
	display.Infof("✅ Tools agent created with ParallelToolCalls enabled")
	return toolsAgent, nil
}

func GetToolsIndex() []*tools.Tool {
	addTool := tools.NewTool("add_numbers").
		SetDescription("Add two numbers together").
		AddParameter("a", "number", "First number", true).
		AddParameter("b", "number", "Second number", true)

	multiplyTool := tools.NewTool("multiply_numbers").
		SetDescription("Multiply two numbers together").
		AddParameter("a", "number", "First number", true).
		AddParameter("b", "number", "Second number", true)

	sayHelloTool := tools.NewTool("say_hello").
		SetDescription("Say hello to the given name").
		AddParameter("name", "string", "The name to greet", true)

	return []*tools.Tool{addTool, multiplyTool, sayHelloTool}
}
