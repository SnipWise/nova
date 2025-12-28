---
id: structured-validation
name: Structured Agent with Array Output
category: structured
complexity: advanced
sample_source: 25
description: Agent that returns an array of structured objects for multi-entity extraction
---

# Structured Agent with Array Output

## Description

Creates an agent that returns an array of structured objects, perfect for detecting multiple entities, intents, or items in a single message. Uses Go generics with slice types.

## Use Cases

- Multiple intent detection
- Multi-entity extraction
- Batch classification
- List extraction
- Multiple action detection
- Conversation analysis

## Complete Code

```go
package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"

	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// === DEFINE YOUR INTENT STRUCTURE ===
type Intent struct {
	Action    string `json:"action"`
	Character string `json:"name"`
	Known     bool   `json:"known"`
}

func main() {
	ctx := context.Background()

	// === CREATE STRUCTURED AGENT WITH ARRAY TYPE ===
	// Note: []Intent means the agent returns a slice of Intent structs
	agent, err := structured.NewAgent[[]Intent](
		ctx,
		agents.Config{
			Name:                    "DungeonMaster",
			Description:             "...",
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			KeepConversationHistory: true,
			SystemInstructions: `
			You are helping the dungeon master of a D&D game.
			Detect if the user wants to speak to one of the following NPCs:
			Thrain (dwarf blacksmith),
			Liora (elven mage),
			Galdor (human rogue),
			Elara (halfling ranger),
			Shesepankh (tiefling warlock).

			When identifying NPCs:
			- if the user wants to speak to a dwarf blacksmith, they mean Thrain.
			- if the user wants to speak to an elven mage, they mean Liora.
			- if the user wants to speak to a human rogue, they mean Galdor.
			- if the user wants to speak to a halfling ranger, they mean Elara.
			- if the user wants to speak to a tiefling warlock, they mean Shesepankh.

			For each intent, respond with:
			action: speak (or other action if relevant, ex meet, talk, etc)
			character: <NPC name>
			known: <true or false>

			Set known to true if:
			- The user explicitly mentions the NPC by name (Thrain, Liora, Galdor, Elara, or Shesepankh), OR
			- The user mentions the NPC by their role/description (dwarf blacksmith, elven mage, human rogue, halfling ranger, tiefling warlock), OR
			- The user mentions a topic clearly associated with one of the known NPCs (e.g., "spells and magic" = elven mage = Liora)

			Set known to false if:
			- The user wants to speak to someone who is NOT in the list of known NPCs
			`,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.7),
			TopP:        models.Float64(0.9),
		},
	)
	if err != nil {
		panic(err)
	}

	// === GENERATE ARRAY OF INTENTS ===
	intents, _, err := agent.GenerateStructuredData([]messages.Message{
		{
			Role: roles.User,
			Content: `
				I want to chat with Thrain and learn about his blacksmith skills.
				I want to meet a dwarf blacksmith.
				I want to speak about spells and magic.
				I want to speak to Bob Morane.
				I want to talk to Galdor about stealth missions.
			`,
		},
	})
	if err != nil {
		panic(err)
	}

	// === PROCESS EACH INTENT ===
	// intents is a *[]Intent (pointer to slice)
	for _, intent := range *intents {
		display.NewLine()
		display.Title("Intent Detection")

		display.KeyValue("Action", intent.Action)
		display.KeyValue("Character", intent.Character)
		display.KeyValue("Known", conversion.BoolToString(intent.Known))
		display.Separator()
	}
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TEMPERATURE: 0.7
TOP_P: 0.9
```

## How It Works

### 1. Define Struct

```go
type Intent struct {
	Action    string `json:"action"`
	Character string `json:"name"`
	Known     bool   `json:"known"`
}
```

### 2. Create Agent with Array Type

```go
// IMPORTANT: Use []Intent not Intent
agent, err := structured.NewAgent[[]Intent](ctx, config, models)
```

The `[[]Intent]` generic parameter tells the agent to return a slice.

### 3. Generate Array of Structs

```go
intents, finishReason, err := agent.GenerateStructuredData(messages)

// intents is a *[]Intent (pointer to slice)
for _, intent := range *intents {
	fmt.Printf("%s -> %s\n", intent.Action, intent.Character)
}
```

## Customization

### Different Array Types

#### Multiple Products
```go
type Product struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

agent, _ := structured.NewAgent[[]Product](ctx, config, models)

products, _, _ := agent.GenerateStructuredData(messages)
for _, product := range *products {
	fmt.Printf("%s: $%.2f\n", product.Name, product.Price)
}
```

#### Multiple Tasks
```go
type Task struct {
	Title    string `json:"title"`
	Priority string `json:"priority"`
	Assignee string `json:"assignee"`
	DueDate  string `json:"due_date"`
}

agent, _ := structured.NewAgent[[]Task](ctx, config, models)
```

#### Multiple Entities
```go
type Entity struct {
	Type  string `json:"type"`   // "person", "organization", "location"
	Name  string `json:"name"`
	Context string `json:"context"`
}

agent, _ := structured.NewAgent[[]Entity](ctx, config, models)
```

### Processing Results

```go
intents, _, err := agent.GenerateStructuredData(messages)
if err != nil {
	panic(err)
}

// Count known vs unknown
knownCount := 0
unknownCount := 0

for _, intent := range *intents {
	if intent.Known {
		knownCount++
		fmt.Printf("✅ %s -> %s\n", intent.Action, intent.Character)
	} else {
		unknownCount++
		fmt.Printf("❌ %s -> %s (unknown)\n", intent.Action, intent.Character)
	}
}

fmt.Printf("\nKnown: %d, Unknown: %d\n", knownCount, unknownCount)
```

### Filtering Results

```go
intents, _, _ := agent.GenerateStructuredData(messages)

// Get only known NPCs
var knownIntents []Intent
for _, intent := range *intents {
	if intent.Known {
		knownIntents = append(knownIntents, intent)
	}
}

// Process known intents
for _, intent := range knownIntents {
	handleNPCInteraction(intent.Character, intent.Action)
}
```

### Grouping by Action

```go
intents, _, _ := agent.GenerateStructuredData(messages)

// Group by action
actionMap := make(map[string][]Intent)
for _, intent := range *intents {
	actionMap[intent.Action] = append(actionMap[intent.Action], intent)
}

// Process each action group
for action, group := range actionMap {
	fmt.Printf("Action: %s (%d intents)\n", action, len(group))
	for _, intent := range group {
		fmt.Printf("  - %s\n", intent.Character)
	}
}
```

## Important Notes

- **Array Type**: Use `[[]YourStruct]` not `[YourStruct]`
- **Return Type**: `GenerateStructuredData` returns `*[]YourStruct`
- **Iteration**: Use `for _, item := range *result`
- **Empty Array**: If no items found, returns empty slice `[]`
- **Order**: Items are returned in detection order

## Validation

```go
intents, _, err := agent.GenerateStructuredData(messages)
if err != nil {
	log.Fatalf("Error: %v", err)
}

// Check if any intents were detected
if len(*intents) == 0 {
	fmt.Println("No intents detected")
	return
}

// Validate each intent
for i, intent := range *intents {
	if intent.Action == "" {
		log.Printf("Warning: Intent %d has empty action\n", i)
	}
	if intent.Character == "" {
		log.Printf("Warning: Intent %d has empty character\n", i)
	}
}
```

## Batch Processing with Arrays

```go
messages := []string{
	"Talk to Thrain and Liora about their skills",
	"Meet Galdor for a quest",
	"I want to see Bob and Alice",
}

for _, msg := range messages {
	intents, _, _ := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: msg},
	})

	fmt.Printf("Message: %s\n", msg)
	fmt.Printf("Detected %d intents:\n", len(*intents))
	for _, intent := range *intents {
		fmt.Printf("  - %s: %s\n", intent.Action, intent.Character)
	}
	fmt.Println()
}
```

## Comparison: Single vs Array

### Single Object
```go
// Returns ONE intent
agent, _ := structured.NewAgent[Intent](ctx, config, models)
intent, _, _ := agent.GenerateStructuredData(messages)

fmt.Printf("%s -> %s\n", intent.Action, intent.Character)
```

### Array of Objects
```go
// Returns MULTIPLE intents
agent, _ := structured.NewAgent[[]Intent](ctx, config, models)
intents, _, _ := agent.GenerateStructuredData(messages)

for _, intent := range *intents {
	fmt.Printf("%s -> %s\n", intent.Action, intent.Character)
}
```

## Best Practices

1. **Clear instructions**: Specify when to create multiple vs single entries
2. **Validate length**: Check if array is empty before processing
3. **Handle duplicates**: LLM might return same entity multiple times
4. **Set limits**: Consider adding "maximum 10 items" in instructions
5. **Temperature**: Use 0.7-0.9 for arrays (more creative)

## Related Patterns

- For single object: See `structured-schema.md`
- For simple extraction: See `structured-output.md`
- For intent routing: See `orchestrator/topic-detection.md`
