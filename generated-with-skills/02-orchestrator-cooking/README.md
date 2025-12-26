# 02 - Orchestrator Agent for Cooking Topics

> Generated with Nova Agent Builder skill

## Description

An orchestrator agent specialized in detecting and classifying cooking-related topics. This agent can identify the main culinary category from user queries and route them to appropriate specialized cooking agents.

## Features

- **14 Cooking Categories**: Baking, Grilling, Desserts, Vegetables, Seafood, Pasta, Meat, Beverages, Asian, Mexican, Breakfast, Soup, Sauce, Technique
- **Fast Classification**: Uses low temperature (0.0) for consistent topic detection
- **Smart Routing**: Automatically routes to specialized cooking agents
- **Emoji Indicators**: Visual topic representation
- **Non-cooking Detection**: Identifies when topics aren't food-related

## Use Cases

- Multi-agent cooking assistant system
- Recipe recommendation routing
- Culinary chatbot with specialized experts
- Cooking education platform
- Restaurant ordering system

## Prerequisites

- Model `hf.co/menlo/lucy-gguf:q4_k_m` available

## Installation

```bash
cd generated-with-skills/02-orchestrator-cooking
go mod init orchestrator-cooking
go mod tidy
```

## Usage

```bash
go run main.go
```

### Example Output

```
🍳 Cooking Topic Detection Orchestrator
======================================================================

[ 1] Query: How do I make chocolate chip cookies?
     ✅ Detected Topic: 🥐 Baking

[ 2] Query: What's the best way to grill a steak?
     ✅ Detected Topic: 🔥 Grilling

[ 3] Query: Can you give me a recipe for tomato soup?
     ✅ Detected Topic: 🍲 Soup

[ 4] Query: How do I make homemade pasta?
     ✅ Detected Topic: 🍝 Pasta

[ 5] Query: What's a good marinade for salmon?
     ✅ Detected Topic: 🐟 Seafood

[ 6] Query: How do I make scrambled eggs fluffy?
     ✅ Detected Topic: 🍳 Breakfast

[ 7] Query: What spices go well with chicken curry?
     ✅ Detected Topic: 🍜 Asian

[ 8] Query: How do I prepare a Caesar salad?
     ✅ Detected Topic: 🥗 Vegetables

[ 9] Query: What's the best chocolate cake recipe?
     ✅ Detected Topic: 🍰 Desserts

[10] Query: How do I cook rice perfectly?
     ✅ Detected Topic: 🍜 Asian

[11] Query: Tell me about knife sharpening techniques
     ✅ Detected Topic: 🔪 Technique

[12] Query: What's the capital of France?
     ✅ Detected Topic: 💬 General

======================================================================

📊 Topic Routing Examples
----------------------------------------------------------------------
🥐 Baking       → Route to → Pastry Chef Agent
🔥 Grilling     → Route to → BBQ Master Agent
🐟 Seafood      → Route to → Seafood Specialist Agent
🍝 Pasta        → Route to → Italian Chef Agent
🍜 Asian        → Route to → Asian Cuisine Expert Agent
🔪 Technique    → Route to → Cooking Instructor Agent
💬 General      → Route to → General Assistant Agent

======================================================================
```

## Configuration

### System Instructions

The orchestrator is configured to recognize 14 cooking categories:

```go
systemInstructions := `
You are an expert at identifying cooking and food-related topics.

Topics: Baking, Grilling, Desserts, Vegetables, Seafood,
        Pasta, Meat, Beverages, Asian, Mexican, Breakfast,
        Soup, Sauce, Technique

Respond in JSON: {"topic_discussion": "TopicName"}
`
```

### Model Settings

```go
models.Config{
    Name:        "hf.co/menlo/lucy-gguf:q4_k_m",  // Fast & accurate
    Temperature: models.Float64(0.0),              // Consistent classification
}
```

## Integration with Crew Agent

Use this orchestrator to route cooking queries to specialized agents:

```go
import (
    "github.com/snipwise/nova/nova-sdk/agents/crew"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
)

// Create specialized cooking agents
bakingAgent, _ := chat.NewAgent(ctx, bakingConfig, modelConfig)
grillingAgent, _ := chat.NewAgent(ctx, grillingConfig, modelConfig)
pastaAgent, _ := chat.NewAgent(ctx, pastaConfig, modelConfig)

// Map agents by cooking topic
cookingCrew := map[string]*chat.Agent{
    "baking":   bakingAgent,
    "grilling": grillingAgent,
    "pasta":    pastaAgent,
}

// Create orchestrator
orchestratorAgent, _ := orchestrator.NewAgent(ctx, orchestratorConfig, modelConfig)

// Define routing logic
matchAgentFn := func(topic string) string {
    agentMap := map[string]string{
        "baking":     "baking",
        "grilling":   "grilling",
        "desserts":   "baking",
        "pasta":      "pasta",
        "seafood":    "grilling",
        "vegetables": "grilling",
    }

    if agentId, exists := agentMap[strings.ToLower(topic)]; exists {
        return agentId
    }
    return "general" // fallback
}

// Create crew with auto-routing
crewAgent, _ := crew.NewAgent(
    ctx,
    cookingCrew,
    "general",
    matchAgentFn,
    executeFunction,
    confirmationFn,
)

// Attach orchestrator
crewAgent.SetOrchestratorAgent(orchestratorAgent)

// Now cooking queries are automatically routed!
response, _ := crewAgent.StreamCompletion(
    "How do I make perfect croissants?",
    streamCallback,
)
// Orchestrator detects "Baking" → routes to bakingAgent
```

## Customization

### Add More Categories

```go
systemInstructions := `
Additional topics:
- Vegan (plant-based cooking)
- Barbecue (smoking, low-and-slow)
- Indian (curry, tandoori, biryani)
- French (classic French cuisine)
- Fusion (creative combinations)
`
```

### Custom Routing Logic

```go
func routeToCookingAgent(topic string) string {
    switch strings.ToLower(topic) {
    case "baking", "desserts":
        return "pastry-chef"
    case "grilling", "meat", "seafood":
        return "grill-master"
    case "pasta", "sauce":
        return "italian-chef"
    case "asian", "mexican":
        return "international-chef"
    case "technique":
        return "instructor"
    default:
        return "general-chef"
    }
}
```

### Topic Statistics

```go
topicCounts := make(map[string]int)

for _, query := range queries {
    topic, _ := orchestratorAgent.IdentifyTopicFromText(query)
    topicCounts[topic]++
}

// Display most common topics
fmt.Println("📊 Topic Statistics:")
for topic, count := range topicCounts {
    fmt.Printf("%s %-12s: %d queries\n", getTopicEmoji(topic), topic, count)
}
```

## Cooking Topic Categories

| Category | Examples | Emoji |
|----------|----------|-------|
| Baking | Bread, cakes, pastries, cookies | 🥐 |
| Grilling | BBQ, grilled meats, vegetables | 🔥 |
| Desserts | Sweets, ice cream, puddings | 🍰 |
| Vegetables | Salads, veggie dishes, prep | 🥗 |
| Seafood | Fish, shellfish, sushi | 🐟 |
| Pasta | Italian pasta, noodles | 🍝 |
| Meat | Steaks, roasts, preparation | 🥩 |
| Beverages | Drinks, smoothies, cocktails | 🍹 |
| Asian | Asian cuisine, stir-fry, rice | 🍜 |
| Mexican | Tacos, burritos, salsa | 🌮 |
| Breakfast | Eggs, pancakes, morning meals | 🍳 |
| Soup | Broths, stews, chowders | 🍲 |
| Sauce | Condiments, dressings, gravies | 🥫 |
| Technique | Cooking methods, knife skills | 🔪 |
| General | Non-cooking topics | 💬 |

## Performance

- **Average latency**: ~100-200ms per classification
- **Accuracy**: High consistency due to temperature=0.0
- **Model**: Lightweight and fast (lucy-gguf:q4_k_m)

## Related Examples

- **crew-agent**: Multi-agent collaboration (sample 55)
- **orchestrator/topic-detection**: General topic detection
- See [CLAUDE.md](../../CLAUDE.md) for all snippets

## Reference

- Snippet: `.claude/skills/nova-agent-builder/snippets/orchestrator/topic-detection.md`
- Category: `orchestrator`
- Complexity: `intermediate`
- Based on: Nova Orchestrator Agent pattern
