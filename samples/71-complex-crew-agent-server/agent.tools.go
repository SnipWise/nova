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

	toolsAgentModelID := env.GetEnvOrDefault("TOOLS_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	directivesPathFile := env.GetEnvOrDefault("TOOLS_DIRECTIVES_PATH_FILE", "./directives/tools.agent.system.instructions.md")
	toolsAgentSystemInstructionsContent, err := files.ReadTextFile(directivesPathFile)
	if err != nil {
		return nil, err
	}

	// Create coder agent
	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "tools-agent",
			EngineURL:          engineURL,
			SystemInstructions: toolsAgentSystemInstructionsContent,
			KeepConversationHistory: false,
		},
		models.Config{
			Name:              toolsAgentModelID,
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},
		tools.WithTools(GetToolsIndex()),
	)
	if err != nil {
		display.Errorf("❌ Error creating tools agent: %v", err)
		return nil, err
	}
	display.Infof("✅ Tools agent created for tool calls detection")

	return toolsAgent, nil
}

func GetToolsIndex() []*tools.Tool {
	calculateSumTool := tools.NewTool("calculate_sum").
		SetDescription("Calculate the sum of two numbers").
		AddParameter("a", "number", "The first number", true).
		AddParameter("b", "number", "The second number", true)

	// getHistoryMessagesOfAgentByIdTool := tools.NewTool("get_history_messages_of_agent_by_id").
	// 	SetDescription("Get the history messages of an agent by its ID").
	// 	AddParameter("agent_id", "string", "The ID of the agent", true)

	saveSnippettoFileTool := tools.NewTool("save_snippet").
		SetDescription("Save snippet content to a file").
		AddParameter("file_path", "string", "The path of the file to write to", true).
		AddParameter("content", "string", "The content to write to the file", true)

	// sayHelloTool := tools.NewTool("say_hello").
	// 	SetDescription("Say hello to the given name").
	// 	AddParameter("name", "string", "The name to greet", true)

	return []*tools.Tool{
		calculateSumTool,
		//getHistoryMessagesOfAgentByIdTool,
		saveSnippettoFileTool,
		//sayHelloTool,
	}
}

