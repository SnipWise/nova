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
			SystemInstructions: `You are an expert project planner. Break down project descriptions into clear, actionable tasks with:
				- Unique hierarchical IDs (1, 1.1, 1.2, 2, etc.)
				- Specific, actionable descriptions
				- Appropriate responsibility assignments (team/role/person)
				- Logical task grouping and subtask breakdown
				- Additional context when helpful`,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.3),
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

	// Example 1: Simple web application
	fmt.Println("📋 Example 1: Web Application for Task Management")
	fmt.Println(strings.Repeat("-", 80))

	plan1, err := agent.IdentifyPlanFromText(`
		Create a modern web application for task management with the following features:
		- User authentication and authorization
		- Dashboard with task overview
		- Create, edit, and delete tasks
		- Task categories and tags
		- Real-time notifications
		- Mobile responsive design
		- REST API for mobile apps
	`)
	if err != nil {
		panic(err)
	}

	displayPlan(plan1, "Web Application Project")

	// Example 2: E-commerce platform
	fmt.Println("\n📋 Example 2: E-commerce Platform")
	fmt.Println(strings.Repeat("-", 80))

	plan2, err := agent.IdentifyPlanFromText(`
		Build a full-featured e-commerce platform with:
		- Product catalog with search and filters
		- Shopping cart and checkout process
		- Payment integration (Stripe)
		- Order management system
		- Inventory tracking
		- Customer account management
		- Admin dashboard for managing products and orders
		- Email notifications for orders
	`)
	if err != nil {
		panic(err)
	}

	displayPlan(plan2, "E-commerce Platform")

	// Example 3: Mobile app with complex requirements
	fmt.Println("\n📋 Example 3: Social Media Mobile App")
	fmt.Println(strings.Repeat("-", 80))

	plan3, err := agent.IdentifyPlanFromText(`
		Develop a social media mobile application for both iOS and Android with:
		- User profiles with customizable avatars
		- Post creation with images and videos
		- Like and comment system
		- Follow/unfollow functionality
		- Real-time chat messaging
		- Push notifications
		- Content moderation system
		- Analytics dashboard
		- Backend API with GraphQL
	`)
	if err != nil {
		panic(err)
	}

	displayPlan(plan3, "Social Media Mobile App")

	// Example 4: Export plan to JSON
	fmt.Println("\n💾 Example 4: Exporting Plan to JSON")
	fmt.Println(strings.Repeat("-", 80))

	planJSON, err := json.MarshalIndent(plan1, "", "  ")
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

// displayPlan prints a formatted plan with all tasks and subtasks
func displayPlan(plan *agents.Plan, projectName string) {
	fmt.Printf("\n📊 Project Plan: %s\n", projectName)
	fmt.Printf("Total main tasks: %d\n\n", len(plan.Tasks))

	for _, task := range plan.Tasks {
		displayTask(task, 0)
	}
}

// displayTask recursively displays a task and its subtasks with proper indentation
func displayTask(task agents.Task, level int) {
	indent := strings.Repeat("  ", level)
	icon := "📌"
	if level > 0 {
		icon = "  └─"
	}

	// Display task information
	fmt.Printf("%s%s [%s] %s\n", indent, icon, task.ID, task.Description)
	fmt.Printf("%s   👤 Responsible: %s\n", indent, task.Responsible)


	if level == 0 {
		fmt.Println() // Add spacing between main tasks
	}
}
