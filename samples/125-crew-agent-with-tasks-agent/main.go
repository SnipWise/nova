package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/crew"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/agents/tasks"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	ctx := context.Background()

	// --- 1. Create the Tasks Agent ---
	tasksAgent, err := tasks.NewAgent(
		ctx,
		agents.Config{
			Name:      "project-planner",
			EngineURL: engineURL,
			SystemInstructions: `You are an expert task planner for an AI coding assistant.
				Your job is to analyze a user request and break it down into an ordered list of tasks.

				## Task Classification
				Each task must be classified by its "responsible" field:
				- "tool": the task requires calling an external tool (file system operation, API call, etc.)
				- "completion": the task requires a generalist LLM to generate text, analysis, documentation, or explanations
				- "developer": the task requires a code-specialized LLM to generate, review, or refactor source code

				## Available Tools
				The following tools are available for "tool" tasks:
				- read_file: read the content of a file (arguments: "path")
				- save_file: save/write content to a file (arguments: "path")
				- create_directory: create a directory/folder (arguments: "path")

				When responsible is "tool", you MUST also set:
				- "tool_name": the exact tool name from the list above
				- "arguments": a map of argument names to values (e.g. {"path": "./demo"})

				## Task Complexity
				Each task must have a "complexity" field:
				- "simple": trivial operation, no reasoning needed
				- "moderate": requires some logic or moderate generation
				- "complex": requires deep reasoning, code generation, or analysis

				## Task Dependencies
				Each task must have a "depends_on" field: a list of task IDs that must be completed before this task can start.

				## Rules for Task Ordering
				CRITICAL: Reorder tasks based on logical dependencies, NOT based on the order in the user request.

				## Output Format
				For each task:
				- "id": sequential number as string ("1", "2", "3", ...)
				- "description": clear, actionable description
				- "responsible": "tool", "completion", or "developer"
				- "tool_name": (only when responsible is "tool")
				- "arguments": (only when responsible is "tool")
				- "depends_on": list of task IDs (empty list [] if none)
				- "complexity": "simple", "moderate", or "complex"
				`,
		},
		models.Config{
			Name:        "huggingface.co/menlo/jan-nano-128k-gguf:Q4_K_M",
			Temperature: models.Float64(0.0),
		},
		tasks.BeforeCompletion(func(a *tasks.Agent) {
			fmt.Println("🔄 Analyzing request and identifying plan...")
		}),
		tasks.AfterCompletion(func(a *tasks.Agent) {
			fmt.Println("✅ Plan identification completed!")
		}),
	)
	if err != nil {
		panic(err)
	}

	// --- 2. Create the Tools Agent ---
	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "file-tools-agent",
			EngineURL:          engineURL,
			SystemInstructions: "You are a file system assistant. Execute the requested tool operations.",
		},
		models.Config{
			Name:        "huggingface.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
		},
		tools.WithTools(getToolsIndex()),
	)
	if err != nil {
		panic(err)
	}

	// --- 3. Create the Agent Crew ---
	coderAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "coder",
			EngineURL:               engineURL,
			SystemInstructions:      "You are an expert programming assistant. You write clean, efficient, and well-documented code.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "huggingface.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.8),
		},
	)
	if err != nil {
		panic(err)
	}

	genericAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "generic",
			EngineURL:               engineURL,
			SystemInstructions:      "You respond appropriately to different types of questions. Always start with the most important information.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "huggingface.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.8),
		},
	)
	if err != nil {
		panic(err)
	}

	agentCrew := map[string]*chat.Agent{
		"coder":   coderAgent,
		"generic": genericAgent,
	}

	// --- 4. Create the Orchestrator Agent ---
	orchestratorAgent, err := orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:      "orchestrator-agent",
			EngineURL: engineURL,
			SystemInstructions: `You are good at identifying the topic of a conversation.
				Given a user's input, identify the main topic of discussion in only one word.
				The possible topics are: Technology, Coding, Programming, Science, Mathematics,
				Food, Education, Finance, History, Literature, Art, Music.
				Respond in JSON format with the field 'topic_discussion'.`,
		},
		models.Config{
			Name:        "huggingface.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	matchAgentFunction := func(currentAgentId, topic string) string {
		fmt.Println("🔵 Matching agent for topic:", topic)
		var agentId string
		switch strings.ToLower(topic) {
		case "coding", "programming", "development", "code", "software", "technology":
			agentId = "coder"
		default:
			agentId = "generic"
		}
		fmt.Println("🟢 Matched agent ID:", agentId)
		return agentId
	}

	// --- 5. Create the Crew Agent with Tasks Agent ---
	crewAgent, err := crew.NewAgent(
		ctx,
		crew.WithAgentCrew(agentCrew, "generic"),
		crew.WithMatchAgentIdToTopicFn(matchAgentFunction),
		crew.WithOrchestratorAgent(orchestratorAgent),
		crew.WithTasksAgent(tasksAgent),
		crew.WithToolsAgent(toolsAgent),
		crew.WithExecuteFn(executeFunction),
		crew.WithConfirmationPromptFn(confirmationPromptFunction),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("🤖 Crew Agent with Tasks Agent (CLI Mode)")
	fmt.Println("The tasks agent will analyze your request, create a plan,")
	fmt.Println("and the orchestrator will route each task to the right crew member.")
	fmt.Println("Type '/bye' to quit")
	fmt.Println("---")

	for {
		markdownParser := display.NewMarkdownChunkParser()

		input := prompt.NewWithColor("🧑 You: ")
		question, err := input.RunWithEdit()
		if err != nil {
			display.Errorf("Error reading input: %v", err)
			continue
		}

		if strings.HasPrefix(question, "/bye") {
			display.Infof("👋 Goodbye!")
			break
		}

		if question == "" {
			continue
		}

		display.NewLine()

		result, err := crewAgent.StreamCompletion(question, func(chunk string, finishReason string) error {
			if chunk != "" {
				display.MarkdownChunk(markdownParser, chunk)
			}
			if finishReason == "stop" {
				markdownParser.Flush()
				markdownParser.Reset()
				display.NewLine()
			}
			return nil
		})

		if err != nil {
			display.Errorf("❌ Error: %v", err)
			continue
		}

		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", result.FinishReason)
		display.KeyValue("Context size", fmt.Sprintf("%d characters", crewAgent.GetContextSize()))
		display.Separator()
	}
}

func getToolsIndex() []*tools.Tool {
	readFileTool := tools.NewTool("read_file").
		SetDescription("Read the content of a file").
		AddParameter("path", "string", "The file path to read", true)

	saveFileTool := tools.NewTool("save_file").
		SetDescription("Save/write content to a file").
		AddParameter("path", "string", "The file path to save to", true)

	createDirectoryTool := tools.NewTool("create_directory").
		SetDescription("Create a directory/folder").
		AddParameter("path", "string", "The directory path to create", true)

	return []*tools.Tool{
		readFileTool,
		saveFileTool,
		createDirectoryTool,
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	fmt.Printf("\n🔧 Executing: %s\n", functionName)

	switch functionName {
	case "read_file":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments"}`, nil
		}
		return fmt.Sprintf(`{"content": "Content of file %s (simulated)"}`, args.Path), nil

	case "save_file":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments"}`, nil
		}
		return fmt.Sprintf(`{"status": "File saved to %s (simulated)"}`, args.Path), nil

	case "create_directory":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments"}`, nil
		}
		return fmt.Sprintf(`{"status": "Directory %s created (simulated)"}`, args.Path), nil

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}

func confirmationPromptFunction(functionName string, arguments string) tools.ConfirmationResponse {
	display.Colorf(display.ColorGreen, "🟢 Detected function: %s with arguments: %s\n", functionName, arguments)
	choice := prompt.HumanConfirmation(fmt.Sprintf("Execute %s with %v?", functionName, arguments))
	return choice
}
