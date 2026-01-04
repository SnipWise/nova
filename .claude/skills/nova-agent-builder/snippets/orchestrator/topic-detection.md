---
id: topic-detection
name: Orchestrator Agent (Topic Detection)
category: orchestrator
complexity: intermediate
sample_source: 55
description: Agent specialized in identifying topics/intents from user input for routing
---

# Orchestrator Agent (Topic Detection)

## Description

Creates an orchestrator agent that analyzes user input to identify the main topic or intent. This agent is specialized for routing queries to the appropriate specialized agent in multi-agent systems.

## Use Cases

- Route user queries to specialized agents in a crew
- Detect conversation topics for context switching
- Identify user intent for workflow automation
- Classify requests for proper handling
- Dynamic agent selection in multi-agent systems

## Key Features

- Uses `agents.Intent` structure for topic identification
- Wraps structured agent with topic-specific methods
- Designed for integration with crew/composite agents
- Provides both message-based and text-based APIs
- Optimized for fast topic classification

---

## Base Template

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()
	engineURL := "http://localhost:12434/engines/llama.cpp/v1"

	// === ORCHESTRATOR SYSTEM INSTRUCTIONS ===
	// Define the topics your orchestrator should recognize
	systemInstructions := `
You are good at identifying the topic of a conversation.
Given a user's input, identify the main topic of discussion in only one word.

The possible topics are:
- Technology, Programming, Software Development
- Science, Mathematics, Physics
- Health, Medicine, Fitness
- Business, Finance, Economics
- Arts, Music, Literature
- Sports, Gaming, Entertainment
- Travel, Food, Cooking
- Education, History, Philosophy

Respond in JSON format with the field 'topic_discussion'.
Example: {"topic_discussion": "Technology"}
	`

	// === CREATE ORCHESTRATOR AGENT ===
	orchestratorAgent, err := orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:                    "orchestrator-agent",
			EngineURL:               engineURL,
			SystemInstructions:      systemInstructions,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
			Temperature: models.Float64(0.0), // Low temperature for consistent classification
		},
	)
	if err != nil {
		panic(err)
	}

	// === TEST TOPIC DETECTION ===
	testQueries := []string{
		"How do I implement a binary search tree in Go?",
		"What's the best recipe for chocolate chip cookies?",
		"Tell me about the theory of relativity",
		"What are the latest trends in machine learning?",
	}

	fmt.Println("🔍 Testing Topic Detection")
	fmt.Println(strings.Repeat("=", 60))

	for _, query := range testQueries {
		fmt.Printf("\n📝 Query: %s\n", query)

		// Method 1: Using IdentifyTopicFromText (simple)
		topic, err := orchestratorAgent.IdentifyTopicFromText(query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		fmt.Printf("✅ Detected Topic: %s\n", topic)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
}
```

## Advanced Usage

### With Message History

```go
import (
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// Method 2: Using IdentifyIntent with full message control
msgs := []messages.Message{
	{
		Role:    roles.User,
		Content: "I need help with debugging my Python code",
	},
}

intent, finishReason, err := orchestratorAgent.IdentifyIntent(msgs)
if err != nil {
	fmt.Printf("Error: %v\n", err)
	return
}

fmt.Printf("Topic: %s\n", intent.TopicDiscussion)
fmt.Printf("Finish Reason: %s\n", finishReason)
```

### Integration with Crew Agent

```go
import (
	"github.com/snipwise/nova/nova-sdk/agents/crew"
)

// Define routing function based on detected topic
matchAgentFunction := func(currentAgentId, topic string) string {
	var agentId string
	switch strings.ToLower(topic) {
	case "coding", "programming", "development", "technology", "software":
		agentId = "coder"
	case "science", "mathematics", "physics":
		agentId = "scientist"
	case "cooking", "recipe", "food", "culinary":
		agentId = "cook"
	default:
		agentId = "generic"
	}
	return agentId
}

// Create crew agent with orchestrator
crewAgent, err := crew.NewAgent(
	ctx,
	chatAgents,              // Map of specialized chat agents
	"generic",               // Default agent
	matchAgentFunction,      // Routing function
	executeFunction,         // Tool execution function
	confirmationPromptFn,    // Confirmation function
)

// Attach orchestrator to crew
crewAgent.SetOrchestratorAgent(orchestratorAgent)

// The crew will now automatically route queries to appropriate agents
// based on the orchestrator's topic detection
```

### Custom Topic Categories

```go
// Define your own topic categories
systemInstructions := `
You are an expert at categorizing customer support requests.

Identify the category from these options:
- Technical Support (bugs, errors, crashes)
- Billing (invoices, payments, subscriptions)
- Account (login, password, settings)
- Feature Request (new features, improvements)
- General Question (how-to, documentation)

Respond with ONLY the category name in the 'topic_discussion' field.
`

orchestratorAgent, _ := orchestrator.NewAgent(
	ctx,
	agents.Config{
		Name:                    "support-orchestrator",
		EngineURL:               engineURL,
		SystemInstructions:      systemInstructions,
		KeepConversationHistory: true,
	},
	models.Config{
		Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
		Temperature: models.Float64(0.0),
	},
)
```

### Multilingual Topic Detection

```go
systemInstructions := `
You can identify topics in multiple languages (English, French, Spanish, German).
Detect the topic and respond in English only.

Topics: Technology, Health, Sports, Travel, Food, Business, Education

Respond in JSON: {"topic_discussion": "TopicName"}
`
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "hf.co/menlo/lucy-gguf:q4_k_m"
TEMPERATURE: 0.0  # Low for consistent classification

# Recommended models for orchestration
# - hf.co/menlo/lucy-gguf:q4_k_m (balanced)
# - hf.co/menlo/jan-nano-gguf:q4_k_m (fast)
# - ai/qwen2.5:0.5B-F16 (lightweight)
```

## Model Selection

### Best Models for Topic Detection

1. **Fast & Accurate**: `hf.co/menlo/lucy-gguf:q4_k_m`
   - Good balance of speed and accuracy
   - Recommended for most use cases

2. **Ultra-Fast**: `ai/qwen2.5:0.5B-F16`
   - Very fast classification
   - Good for high-volume routing

3. **Most Accurate**: `hf.co/menlo/jan-nano-gguf:q4_k_m`
   - Best for complex topic detection
   - Slightly slower but more reliable

## Telemetry & Monitoring

```go
// Get metrics from orchestrator
metadata := orchestratorAgent.GetLastResponseMetadata()
fmt.Printf("Tokens used: %d\n", metadata.Usage.TotalTokens)
fmt.Printf("Latency: %v\n", metadata.Latency)

// Set callback for real-time monitoring
orchestratorAgent.SetTelemetryCallback(func(event base.TelemetryEvent) {
	fmt.Printf("Topic detection: %dms\n", event.ResponseMetadata.Latency.Milliseconds())
})
```

## Important Notes

### DO:
- Use `orchestrator.NewAgent()` for topic detection agents
- Set `Temperature: 0.0` for deterministic classification
- Set `KeepConversationHistory: true` for context-aware routing
- Define specific topics in system instructions
- Use JSON format: `{"topic_discussion": "TopicName"}`
- Use `IdentifyTopicFromText()` for simple text-based detection
- Use `IdentifyIntent()` for message-based detection with history
- Choose small, fast models: `hf.co/menlo/lucy-gguf:q4_k_m` (recommended)
- Test topic detection with edge cases
- Monitor performance (should be < 100ms)

### DON'T:
- Don't use high temperature (> 0.1) for topic detection
- Don't skip defining allowed topics in system instructions
- Don't use large models (overkill for classification)
- Don't assume perfect classification - handle "unknown" topic
- Don't forget to integrate with crew using `SetOrchestratorAgent()`
- Don't use stateless mode (`KeepConversationHistory: false`) for crew integration
- Don't ignore the finish reason - check for errors

### API Methods:
```go
// Simple text-based detection
topic, err := orchestratorAgent.IdentifyTopicFromText(query)

// Message-based detection with history
intent, finishReason, err := orchestratorAgent.IdentifyIntent(messages)

// Get detected topic from intent
topicName := intent.TopicDiscussion
```

### Intent Structure:
```go
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}
```

## Common Patterns

### Pattern 1: Simple Classification
```go
topic, _ := orchestratorAgent.IdentifyTopicFromText(userQuery)
agentId := routingMap[topic]
```

### Pattern 2: With Confidence
```go
intent, _, _ := orchestratorAgent.IdentifyIntent(messages)
if intent.TopicDiscussion == "unknown" {
    // Use fallback agent
}
```

### Pattern 3: Dynamic Routing
```go
topic, _ := orchestratorAgent.IdentifyTopicFromText(query)
if agent, exists := agentCrew[topic]; exists {
    response, _ := agent.GenerateCompletion(messages)
}
```
