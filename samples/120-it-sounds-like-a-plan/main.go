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
			SystemInstructions: `
			You are an expert project planner. Break down project descriptions into clear, actionable tasks with:
				- Unique hierarchical IDs (1, 1.1, 1.2, 2, etc.)
				- Specific, actionable descriptions
				- Appropriate responsibility assignments (team/role/person)
				- Logical task grouping and subtask breakdown
				- Additional context when helpful
			`,
		},
		models.Config{
			//Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			//Name: "hf.co/menlo/lucy-gguf:q4_k_m",
			Name: "ai/qwen2.5:3B-F16",
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
		Create a dungeon and dragons style text-based adventure game with the following features:
		- Character creation with customizable attributes and classes
		- Procedurally generated dungeons with multiple levels
				
		Assign responsibilities to appropriate teams (frontend, backend, devops).
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
