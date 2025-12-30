---
id: crew-agent
name: Crew Agent (Collaborative Multi-Agents)
category: complex
complexity: advanced
sample_source: 55
description: Collaborative multi-agent system with specialized roles
interactive: true
---

# Crew Agent (Collaborative Multi-Agents)

## Description

Creates a system of multiple AI agents collaborating to accomplish complex tasks. Each agent has a specialized role and they communicate with each other to achieve a common goal.

## Use Cases

- Content creation pipelines (research → write → review)
- Complex analysis with multiple perspectives
- Automated workflows with validation
- Systems requiring specialized expertise
- Tasks too complex for a single agent

## ⚠️ Interactive Mode

This snippet requires specific information. Answer the following questions:

### Configuration Questions

1. **How many agents in your crew?**
   - Recommended minimum: 2
   - Practical maximum: 5-6

2. **What role for each agent?**
   - Examples: Researcher, Writer, Reviewer, Analyst, Manager
   - Each role should have a clear responsibility

3. **What interactions between agents?**
   - Sequential (A → B → C)
   - Hierarchical (Manager supervises others)
   - Collaborative (all can interact)

4. **What tools should each agent have?**
   - Research, creation, validation tools, etc.

5. **What is the overall workflow?**
   - Describe the end-to-end workflow

---

## Base Template

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// === AGENT DEFINITION ===
type CrewAgent struct {
	Name        string
	Role        string
	Agent       *chat.Agent
	CanDelegate []string // Names of agents this one can delegate to
}

// === CREW ===
type Crew struct {
	agents  map[string]*CrewAgent
	manager string // Name of manager agent
}

func NewCrew(ctx context.Context, engineURL string) (*Crew, error) {
	crew := &Crew{
		agents: make(map[string]*CrewAgent),
	}

	// === AGENT 1: RESEARCHER ===
	researcher, err := chat.NewAgent(ctx,
		agents.Config{
			Name:                    "researcher",
			EngineURL:               engineURL,
			KeepConversationHistory: true,
			SystemInstructions: `You are a researcher. Your role is to:
- Gather relevant information on the given topic
- Identify key facts and data
- Summarize findings clearly
- Flag areas needing more investigation

Output format: structured summary with key points.`,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.3),
		},
	)
	if err != nil {
		return nil, err
	}
	crew.agents["researcher"] = &CrewAgent{
		Name:  "researcher",
		Role:  "Research and information gathering",
		Agent: researcher,
	}

	// === AGENT 2: WRITER ===
	writer, err := chat.NewAgent(ctx,
		agents.Config{
			Name:                    "writer",
			EngineURL:               engineURL,
			KeepConversationHistory: true,
			SystemInstructions: `You are a professional writer. Your role is to:
- Transform research into well-written content
- Maintain clear and engaging style
- Structure content logically
- Adapt tone to audience

Output format: polished, publishable content.`,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.7),
		},
	)
	if err != nil {
		return nil, err
	}
	crew.agents["writer"] = &CrewAgent{
		Name:  "writer",
		Role:  "Content writing",
		Agent: writer,
	}

	// === AGENT 3: REVIEWER ===
	reviewer, err := chat.NewAgent(ctx,
		agents.Config{
			Name:                    "reviewer",
			EngineURL:               engineURL,
			KeepConversationHistory: true,
			SystemInstructions: `You are a critical reviewer. Your role is to:
- Check content accuracy
- Identify errors and inconsistencies
- Suggest improvements
- Validate quality before publication

Output format: review with specific feedback and recommendations.`,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.2),
		},
	)
	if err != nil {
		return nil, err
	}
	crew.agents["reviewer"] = &CrewAgent{
		Name:  "reviewer",
		Role:  "Quality review",
		Agent: reviewer,
	}

	return crew, nil
}

// === SEQUENTIAL EXECUTION ===
func (c *Crew) ExecuteSequential(task string) (string, error) {
	fmt.Println("🚀 Starting crew execution...")
	fmt.Println(strings.Repeat("=", 50))

	// Step 1: Research
	fmt.Println("\n📚 Step 1: Research")
	researchResult, err := c.agents["researcher"].Agent.GenerateCompletion(
		[]messages.Message{{Role: roles.User, Content: task}},
	)
	if err != nil {
		return "", fmt.Errorf("research failed: %v", err)
	}
	fmt.Printf("Research output: %s\n", truncate(researchResult.Response, 200))

	// Step 2: Writing
	fmt.Println("\n✍️ Step 2: Writing")
	writePrompt := fmt.Sprintf("Based on this research, write the content:\n\n%s", researchResult.Response)
	writeResult, err := c.agents["writer"].Agent.GenerateCompletion(
		[]messages.Message{{Role: roles.User, Content: writePrompt}},
	)
	if err != nil {
		return "", fmt.Errorf("writing failed: %v", err)
	}
	fmt.Printf("Writing output: %s\n", truncate(writeResult.Response, 200))

	// Step 3: Review
	fmt.Println("\n🔍 Step 3: Review")
	reviewPrompt := fmt.Sprintf("Review this content for quality and accuracy:\n\n%s", writeResult.Response)
	reviewResult, err := c.agents["reviewer"].Agent.GenerateCompletion(
		[]messages.Message{{Role: roles.User, Content: reviewPrompt}},
	)
	if err != nil {
		return "", fmt.Errorf("review failed: %v", err)
	}
	fmt.Printf("Review: %s\n", truncate(reviewResult.Response, 200))

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✅ Crew execution completed")

	return writeResult.Response, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// === MAIN ===
func main() {
	ctx := context.Background()

	crew, err := NewCrew(ctx, "http://localhost:12434/engines/llama.cpp/v1")
	if err != nil {
		fmt.Printf("Error creating crew: %v\n", err)
		return
	}

	task := "Write a short article about the benefits of local AI models for privacy"

	result, err := crew.ExecuteSequential(task)
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}

	fmt.Println("\n📄 Final Result:")
	fmt.Println(result)
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "ai/qwen2.5:1.5B-F16"

# Temperature per role
RESEARCHER_TEMP: 0.3  # More factual
WRITER_TEMP: 0.7      # More creative
REVIEWER_TEMP: 0.2    # More rigorous
```

## Customization

### With Manager Agent

```go
// Add hierarchical coordination
manager, _ := chat.NewAgent(ctx,
    agents.Config{
        Name:      "manager",
        EngineURL: engineURL,
        SystemInstructions: `You are a project manager. Your role is to:
- Decompose tasks into subtasks
- Assign work to appropriate agents
- Review progress and redirect if needed
- Make final decisions on quality`,
    },
    models.Config{Temperature: models.Float64(0.3)},
)

func (c *Crew) ExecuteWithManager(task string) (string, error) {
    // 1. Manager decomposes task
    // 2. Manager assigns to agents
    // 3. Manager reviews and iterates
    // 4. Manager approves final output
}
```

### With Tools per Agent

```go
// Researcher with web search tool
researcherTools := []*tools.Tool{
    tools.NewTool("web_search").SetDescription("Search the web"),
}

// Writer with file creation tool
writerTools := []*tools.Tool{
    tools.NewTool("save_draft").SetDescription("Save draft to file"),
}
```

### With Orchestrator Agent (Auto-Routing)

For automatic topic detection and agent routing, use the `orchestrator` agent with `crew` agent:

```go
import (
    "github.com/snipwise/nova/nova-sdk/agents/crew"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
)

// Create orchestrator for topic detection
orchestratorAgent, _ := orchestrator.NewAgent(
    ctx,
    agents.Config{
        Name:      "orchestrator",
        EngineURL: engineURL,
        SystemInstructions: `
Identify the main topic from user input.
Topics: Technology, Science, Business, Health, Entertainment
Respond in JSON: {"topic_discussion": "TopicName"}`,
    },
    models.Config{
        Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
        Temperature: models.Float64(0.0),
    },
)

// Define routing logic
matchAgentFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "technology", "programming":
        return "researcher"
    case "business", "finance":
        return "writer"
    default:
        return "reviewer"
    }
}

// Create crew with auto-routing
crewAgent, _ := crew.NewAgent(
    ctx,
    agentMap,         // map[string]*chat.Agent
    "reviewer",       // default agent
    matchAgentFn,     // routing function
    executeFn,        // tool executor
    confirmFn,        // confirmation
)

// Attach orchestrator
crewAgent.SetOrchestratorAgent(orchestratorAgent)

// Now queries are automatically routed to the right agent!
response, _ := crewAgent.StreamCompletion("How to optimize database queries?", streamCallback)
// Orchestrator detects "Technology" → routes to researcher
```

See the `orchestrator/topic-detection` snippet for more details on topic detection.

## Important Notes

- Each agent should have a clearly defined role
- Temperature affects output style (low=factual, high=creative)
- Sequential flow is simplest; hierarchical is more flexible
- Consider timeouts for complex workflows
- Log agent outputs for debugging
- **Use orchestrator agent for automatic routing in complex crews**
