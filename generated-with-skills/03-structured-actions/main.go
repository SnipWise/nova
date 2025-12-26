package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// === OUTPUT STRUCTURE ===

// Action represents a single action to be performed
type Action struct {
	ID          string   `json:"id" description:"Unique identifier for the action (1, 2, 3...)"`
	Description string   `json:"description" description:"Clear description of what needs to be done"`
	Priority    string   `json:"priority" description:"Priority level: high, medium, or low"`
	Category    string   `json:"category" description:"Action category: work, personal, shopping, health, etc."`
	DueDate     string   `json:"due_date,omitempty" description:"Optional due date if mentioned (YYYY-MM-DD)"`
	Tags        []string `json:"tags,omitempty" description:"Optional tags or keywords"`
	Estimated   string   `json:"estimated_time,omitempty" description:"Estimated time to complete (e.g., '30min', '2h')"`
}

// ActionList represents the complete list of extracted actions
type ActionList struct {
	Actions     []Action `json:"actions" description:"List of all identified actions"`
	TotalCount  int      `json:"total_count" description:"Total number of actions identified"`
	Summary     string   `json:"summary" description:"Brief summary of all actions"`
	HasDeadline bool     `json:"has_deadline" description:"True if any action has a deadline"`
}

func main() {
	ctx := context.Background()

	// === CREATE STRUCTURED AGENT ===
	agent, err := structured.NewAgent[ActionList](
		ctx,
		agents.Config{
			Name:      "action-extractor",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `You are an expert at identifying and organizing action items from text.

Extract all actionable tasks, to-dos, and action items from the provided text.
For each action:
- Assign a unique ID
- Write a clear, actionable description
- Determine priority (high, medium, low)
- Categorize the action (work, personal, shopping, health, etc.)
- Extract any mentioned deadlines or dates
- Add relevant tags
- Estimate time if possible

Be thorough and identify even implicit action items.
Return a complete JSON structure matching the ActionList schema.`,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0), // 0 for deterministic output
		},
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	// === TEST SCENARIOS ===
	testScenarios := []struct {
		Name  string
		Input string
	}{
		{
			Name: "Email with Multiple Tasks",
			Input: `
Hi team,

After today's meeting, here are the action items:

1. John needs to update the documentation by Friday
2. Sarah will schedule a follow-up meeting with the client for next week
3. I'll review the code changes and provide feedback by EOD tomorrow
4. Don't forget to submit your timesheets before the end of the month
5. We should order new office supplies - we're running low on printer paper

Also, can someone pick up coffee for the break room?

Thanks!
`,
		},
		{
			Name: "Personal Daily Tasks",
			Input: `
Tomorrow's plan:
- Wake up at 7am and go for a 30-minute jog
- Buy groceries (milk, bread, eggs, vegetables)
- Call mom to wish her happy birthday
- Finish reading chapter 5 of the book
- Pay the electricity bill (due on the 15th)
- Schedule dentist appointment
- Clean the garage (been postponing this for weeks!)
`,
		},
		{
			Name: "Project Management Note",
			Input: `
URGENT: The production deployment is scheduled for this Friday.

High priority items:
- Run full regression tests on staging
- Update database migration scripts
- Notify all stakeholders about the maintenance window
- Prepare rollback plan

Medium priority:
- Update user documentation
- Create release notes

Low priority:
- Clean up old log files
- Archive deprecated features
`,
		},
	}

	fmt.Println("📋 Action Item Extraction Demo")
	fmt.Println(strings.Repeat("=", 80))

	for i, scenario := range testScenarios {
		fmt.Printf("\n[Scenario %d] %s\n", i+1, scenario.Name)
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("Input:\n%s\n", scenario.Input)
		fmt.Println(strings.Repeat("-", 80))

		// Extract actions
		actionList, finishReason, err := agent.GenerateStructuredData([]messages.Message{
			{
				Role: roles.User,
				Content: fmt.Sprintf(
					"Extract all action items from this text:\n\n%s",
					scenario.Input,
				),
			},
		})

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		// Display results
		fmt.Printf("\n✅ Extracted %d actions (finish_reason: %s)\n\n", actionList.TotalCount, finishReason)
		fmt.Printf("📊 Summary: %s\n\n", actionList.Summary)

		if len(actionList.Actions) > 0 {
			fmt.Println("📝 Action Items:")
			for _, action := range actionList.Actions {
				fmt.Printf("\n   [%s] %s\n", action.ID, action.Description)
				fmt.Printf("       Priority: %s | Category: %s", getPriorityEmoji(action.Priority), action.Category)

				if action.DueDate != "" {
					fmt.Printf(" | Due: %s", action.DueDate)
				}
				if action.Estimated != "" {
					fmt.Printf(" | Time: %s", action.Estimated)
				}

				if len(action.Tags) > 0 {
					fmt.Printf("\n       Tags: %v", action.Tags)
				}
				fmt.Println()
			}
		}

		// Display statistics
		fmt.Println("\n📈 Statistics:")
		priorities := countByPriority(actionList.Actions)
		categories := countByCategory(actionList.Actions)

		fmt.Printf("   Priorities: High: %d | Medium: %d | Low: %d\n",
			priorities["high"], priorities["medium"], priorities["low"])

		fmt.Print("   Categories: ")
		for cat, count := range categories {
			fmt.Printf("%s: %d | ", cat, count)
		}
		fmt.Println()

		if actionList.HasDeadline {
			fmt.Println("   ⏰ Contains time-sensitive actions!")
		}

		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 80))

	// === EXPORT EXAMPLE ===
	fmt.Println("\n💾 Export Example (JSON):")
	if len(testScenarios) > 0 {
		// Re-extract first scenario for export demo
		actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
			{Role: roles.User, Content: "Extract actions: " + testScenarios[0].Input},
		})

		jsonData, _ := json.MarshalIndent(actionList, "", "  ")
		fmt.Println(string(jsonData))
	}
}

// Helper functions

func getPriorityEmoji(priority string) string {
	switch strings.ToLower(priority) {
	case "high":
		return "🔴 High"
	case "medium":
		return "🟡 Medium"
	case "low":
		return "🟢 Low"
	default:
		return "⚪ " + priority
	}
}

func countByPriority(actions []Action) map[string]int {
	counts := map[string]int{"high": 0, "medium": 0, "low": 0}
	for _, action := range actions {
		counts[strings.ToLower(action.Priority)]++
	}
	return counts
}

func countByCategory(actions []Action) map[string]int {
	counts := make(map[string]int)
	for _, action := range actions {
		counts[strings.ToLower(action.Category)]++
	}
	return counts
}
