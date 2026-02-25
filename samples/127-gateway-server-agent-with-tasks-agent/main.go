package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
	"github.com/snipwise/nova/nova-sdk/agents/tasks"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
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
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
		},
		tools.WithTools(getToolsIndex()),
	)
	if err != nil {
		panic(err)
	}

	// --- 3. Create the Chat Agent ---
	chatAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "assistant",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a helpful AI assistant that can plan and execute complex tasks.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.4),
		},
	)
	if err != nil {
		panic(err)
	}

	// --- 4. Create the Gateway Server Agent with Tasks Agent ---
	// The gateway exposes an OpenAI-compatible API.
	// When a tasks agent is configured, requests are analyzed and broken
	// into a plan before the normal execution chain runs.
	gateway, err := gatewayserver.NewAgent(
		ctx,
		gatewayserver.WithSingleAgent(chatAgent),
		gatewayserver.WithPort(8080),
		gatewayserver.WithTasksAgent(tasksAgent),
		gatewayserver.WithToolsAgent(toolsAgent),
		gatewayserver.WithExecuteFn(executeFunction),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("🚀 Gateway Server Agent with Tasks Agent")
	fmt.Println("📡 OpenAI-compatible endpoint: POST /v1/chat/completions")
	fmt.Println()
	fmt.Println("The tasks agent will analyze incoming requests, create a plan,")
	fmt.Println("and execute each task step by step.")
	fmt.Println()
	fmt.Println("Usage with curl (non-streaming):")
	fmt.Println(`  curl http://localhost:8080/v1/chat/completions \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"model":"assistant","messages":[{"role":"user","content":"Create a demo directory and read README.md"}]}'`)
	fmt.Println()
	fmt.Println("Usage with curl (streaming):")
	fmt.Println(`  curl http://localhost:8080/v1/chat/completions \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"model":"assistant","messages":[{"role":"user","content":"Create a demo directory and read README.md"}],"stream":true}'`)

	if err := gateway.StartServer(); err != nil {
		panic(err)
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
