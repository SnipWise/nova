package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tasks"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	// Create a Tasks Agent with lifecycle hooks for monitoring
	agent, err := tasks.NewAgent(
		ctx,
		agents.Config{
			Name:      "project-planner",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `You are an expert task planner for an AI coding assistant.
				Your job is to analyze a user request and break it down into an ordered list of tasks.

				## Task Classification

				Each task must be classified by its "responsible" field:
				- "tool": the task requires calling an external tool (file system operation, API call, etc.)
				- "completion": the task requires the LLM to generate text, code, or analysis

				## Available Tools

				The following tools are available. Use these names to detect tool-related tasks:
				- read_file: read the content of a file
				- save_file: save/write content to a file
				- create_directory: create a directory/folder

				## Rules for Task Ordering

				CRITICAL: You must reorder tasks based on logical dependencies, NOT based on the order they appear in the user request.
				- If a task requires saving a file into a directory, the directory creation task MUST come before the file saving task.
				- If a task requires reading a file to generate content, the read task MUST come before the generation task.
				- If a task generates content that will be saved, the generation task MUST come before the save task.
				- Always identify prerequisites and ensure they are scheduled first.

				## Output Format

				For each task:
				- "id": a sequential number as a string ("1", "2", "3", ...) representing the execution order
				- "description": a clear, specific, actionable description of what needs to be done
				- "responsible": either "tool" or "completion"
				`,
		},
		/*
			Lucy is faster
			But Jan Nano is more accurate for complex reasoning tasks: for ordering the tasks correctly, understanding the dependencies
		*/
		models.Config{
			//Name: "huggingface.co/menlo/lucy-128k-gguf:Q4_K_M",
			Name: "huggingface.co/menlo/jan-nano-128k-gguf:Q4_K_M",
			Temperature: models.Float64(0.0),
			//MaxTokens:   models.Int(4000),
		},
		// Add lifecycle hooks for monitoring
		tasks.BeforeCompletion(func(a *tasks.Agent) {
			fmt.Println("🔄 Analyzing project description...")
		}),
		tasks.AfterCompletion(func(a *tasks.Agent) {
			fmt.Println("✅ Plan identification completed!")
		}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🎯 N.O.V.A. Tasks Agent - Project Planning Demo")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	fmt.Println("📋 Example 1: Software specification")
	fmt.Println(strings.Repeat("-", 80))

	plan, err := agent.IdentifyPlanFromText(`
		Read the file in ./specification.md.  
		Use these document to generate golang code.
		Check carefully the syntax of the code.
		Make the code simple and add remarks to explain the code.
		Once the code ready,
		Then save the source code into a ./demo/main.go file in the same folder,
		Then create a new markdown document that explains how the code works,
		and save it into ./demo/explanation.md file.
		create the demo folder if it does not exist.
	`)
	if err != nil {
		panic(err)
	}


	fmt.Println(strings.Repeat("-", 80))

	planJSON, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println("Plan exported as JSON:")
	fmt.Println(string(planJSON))

	// Show debug information
	fmt.Println("\n🔍 Debug Information")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Agent Name: %s\n", agent.GetName())
	fmt.Printf("Agent Kind: %s\n", agent.Kind())
	fmt.Printf("Model ID: %s\n", agent.GetModelID())
	fmt.Printf("Total messages in history: %d\n", len(agent.GetMessages()))

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✨ Demo completed successfully!")
	fmt.Println(strings.Repeat("=", 80))
}
/* TODO:
- run an action for every tasks
- the result of the former taks is used by the next task
  - if completion: added to the messages of the completion agent (chat agent)

- in a future version of Nova: add fields to the task struct: 
  - tool name
  - model ID
*/