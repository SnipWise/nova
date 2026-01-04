---
id: intent-detection
name: Intent Detection Agent
category: structured
complexity: intermediate
sample_source: 24
description: Agent that detects user intent and extracts structured data using structured.NewAgent
---

# Intent Detection Agent

## Description

Creates a Nova structured agent specifically designed for intent detection, classification, and entity extraction. Uses `structured.NewAgent[T]` to return typed Go structs representing detected intents and their parameters.

## Use Cases

- Chat routing (detect which agent should handle request)
- Command classification
- Entity extraction from user input
- Game NPC selection
- Action detection for automation
- Topic categorization
- Sentiment analysis with structured output

## Complete Code

```go
package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// === INTENT STRUCTURE ===
// Define the structured output for intent detection
type Intent struct {
	Action    string `json:"action"`    // The detected action/intent
	Character string `json:"name"`      // Entity extracted from input
	Known     bool   `json:"known"`     // Whether entity is recognized
}

func main() {
	ctx := context.Background()

	// === CREATE INTENT DETECTION AGENT ===
	agent, err := structured.NewAgent[Intent](
		ctx,
		agents.Config{
			Name:        "DungeonMaster",
			Description: "Intent detection for D&D NPC conversations",
			EngineURL:   "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
You are helping the dungeon master of a D&D game.
Detect if the user wants to speak to one of the following NPCs:
Thrain (dwarf blacksmith),
Liora (elven mage),
Galdor (human rogue),
Elara (halfling ranger),
Shesepankh (tiefling warlock).

If the user's message does not explicitly mention wanting to speak to one of these NPCs, respond with:
action: speak
character: <NPC name>
known: false

Otherwise, respond with:
action: speak
character: <NPC name>
Where <NPC name> is the name of the NPC the user wants to speak to: Thrain, Liora, Galdor, Elara, or Shesepankh.
known: true
			`,
			KeepConversationHistory: false, // Stateless for classification
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0), // Deterministic classification
		},
	)
	if err != nil {
		panic(err)
	}

	// === TEST MESSAGES ===
	messagesList := []string{
		"I want to chat with Thrain and learn about his blacksmith skills.",
		"I want to meet a dwarf blacksmith.",
		"I want to speak about spells and magic.",
		"I want to speak to Bob Morane.",
	}

	// === PROCESS EACH MESSAGE ===
	for _, message := range messagesList {
		// Generate structured intent data
		response, finishReason, err := agent.GenerateStructuredData(
			[]messages.Message{
				{
					Role:    roles.User,
					Content: message,
				},
			},
		)
		if err != nil {
			panic(err)
		}

		display.NewLine()
		display.Title("Intent Detection")
		display.KeyValue("User Input", message)
		display.Separator()

		// Access structured fields directly
		display.KeyValue("Action", response.Action)
		display.KeyValue("Character", response.Character)
		display.KeyValue("Known", conversion.BoolToString(response.Known))

		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", finishReason)
		display.Separator()

		// Use intent for routing
		if response.Known {
			display.Success("Routing to known NPC: " + response.Character)
		} else {
			display.Warning("Unknown character, using generic response")
		}
	}
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TEMPERATURE: 0.0  # Deterministic classification
KEEP_CONVERSATION_HISTORY: false  # Stateless
```

## Key API

### structured.NewAgent[T]

```go
import "github.com/snipwise/nova/nova-sdk/agents/structured"

// Define intent structure
type Intent struct {
    Action    string `json:"action"`
    Entity    string `json:"entity"`
    Confidence float64 `json:"confidence"`
}

// Create typed agent
agent, err := structured.NewAgent[Intent](
    ctx,
    agents.Config{
        EngineURL:               engineURL,
        SystemInstructions:      classificationPrompt,
        KeepConversationHistory: false, // Stateless for classification
    },
    models.Config{
        Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature: models.Float64(0.0), // Deterministic
    },
)
```

### GenerateStructuredData

```go
// Returns typed struct directly
intent, finishReason, err := agent.GenerateStructuredData(
    []messages.Message{
        {Role: roles.User, Content: userInput},
    },
)

// Access typed fields
fmt.Println(intent.Action)
fmt.Println(intent.Entity)
```

## Customization

### Multi-Intent Detection

```go
type MultiIntent struct {
    Primary   string   `json:"primary"`   // Main intent
    Secondary []string `json:"secondary"` // Additional intents
    Entities  []Entity `json:"entities"`  // Extracted entities
    Sentiment string   `json:"sentiment"` // User sentiment
}

type Entity struct {
    Type  string `json:"type"`  // "person", "product", "location"
    Value string `json:"value"` // Extracted value
}

agent, err := structured.NewAgent[MultiIntent](
    ctx,
    agents.Config{
        SystemInstructions: `
Analyze user input and extract:
1. Primary intent (main purpose)
2. Secondary intents (other requests)
3. Named entities with types
4. Sentiment (positive/neutral/negative)
        `,
    },
    modelConfig,
)
```

### Confidence Scoring

```go
type IntentWithConfidence struct {
    Intent     string  `json:"intent"`
    Confidence float64 `json:"confidence"` // 0.0 to 1.0
    Fallback   bool    `json:"fallback"`   // True if uncertain
}

// Use in application
response, _, _ := agent.GenerateStructuredData(messages)
if response.Confidence < 0.7 {
    // Use fallback handler
    handleUncertainIntent(response)
} else {
    // Route with confidence
    routeToHandler(response.Intent)
}
```

### Intent Router

```go
type IntentRouter struct {
    agent    *structured.Agent[Intent]
    handlers map[string]func(Intent) error
}

func NewIntentRouter(agent *structured.Agent[Intent]) *IntentRouter {
    return &IntentRouter{
        agent:    agent,
        handlers: make(map[string]func(Intent) error),
    }
}

func (r *IntentRouter) RegisterHandler(action string, handler func(Intent) error) {
    r.handlers[action] = handler
}

func (r *IntentRouter) Route(userInput string) error {
    // Detect intent
    intent, _, err := r.agent.GenerateStructuredData(
        []messages.Message{
            {Role: roles.User, Content: userInput},
        },
    )
    if err != nil {
        return err
    }

    // Route to handler
    if handler, exists := r.handlers[intent.Action]; exists {
        return handler(intent)
    }

    return fmt.Errorf("no handler for action: %s", intent.Action)
}

// Usage
router := NewIntentRouter(agent)
router.RegisterHandler("speak", handleSpeakIntent)
router.RegisterHandler("buy", handleBuyIntent)
router.RegisterHandler("attack", handleAttackIntent)

router.Route("I want to speak to Thrain")
```

### Batch Classification

```go
func classifyBatch(agent *structured.Agent[Intent], inputs []string) []Intent {
    results := make([]Intent, len(inputs))
    var wg sync.WaitGroup

    for i, input := range inputs {
        wg.Add(1)
        go func(idx int, msg string) {
            defer wg.Done()
            intent, _, _ := agent.GenerateStructuredData(
                []messages.Message{
                    {Role: roles.User, Content: msg},
                },
            )
            results[idx] = intent
        }(i, input)
    }

    wg.Wait()
    return results
}
```

## Advanced Examples

### E-Commerce Intent Detection

```go
type ECommerceIntent struct {
    Action   string   `json:"action"`   // "search", "buy", "support", "return"
    Products []string `json:"products"` // Mentioned products
    Price    float64  `json:"price"`    // Price mentioned (0 if none)
    Urgent   bool     `json:"urgent"`   // Is request urgent?
}

agent, err := structured.NewAgent[ECommerceIntent](
    ctx,
    agents.Config{
        SystemInstructions: `
Analyze e-commerce customer messages and extract:
- Action: search, buy, support, return, track
- Products: list of mentioned product names
- Price: any price mentioned (0 if none)
- Urgent: true if message indicates urgency
        `,
    },
    models.Config{
        Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature: models.Float64(0.0),
    },
)
```

### Customer Support Routing

```go
type SupportIntent struct {
    Category    string   `json:"category"`    // "technical", "billing", "shipping"
    Priority    string   `json:"priority"`    // "low", "medium", "high", "urgent"
    Keywords    []string `json:"keywords"`    // Important keywords
    NeedsHuman  bool     `json:"needs_human"` // Escalate to human?
}

agent, err := structured.NewAgent[SupportIntent](
    ctx,
    agents.Config{
        SystemInstructions: `
Classify customer support tickets:
- Category: technical, billing, shipping, account, other
- Priority: low, medium, high, urgent
- Keywords: extract 3-5 important keywords
- NeedsHuman: true if complex/emotional/legal issue
        `,
    },
    modelConfig,
)

// Usage
ticket, _, _ := agent.GenerateStructuredData(messages)
if ticket.NeedsHuman || ticket.Priority == "urgent" {
    escalateToHuman(ticket)
} else {
    routeToBot(ticket.Category)
}
```

## Important Notes

### DO:
- Use `structured.NewAgent[YourStruct]` with custom Go struct
- Set `Temperature: 0.0` for deterministic classification
- Set `KeepConversationHistory: false` for intent detection (stateless)
- Use JSON tags on struct fields for consistent parsing
- Validate structured output before using in production
- Add confidence fields for uncertainty handling
- Use clear, specific system instructions
- Test with edge cases and ambiguous inputs

### DON'T:
- Don't use high temperature for classification (causes inconsistency)
- Don't skip validation of extracted entities
- Don't assume perfect classification - always handle fallbacks
- Don't use stateful history for routing (each request should be independent)
- Don't overcomplicate the output struct - keep it focused
- Don't ignore the finish reason - check for errors

## Model Recommendations

For intent detection:
- **Best**: `hf.co/menlo/jan-nano-gguf:q4_k_m` (optimized for structured output)
- **Alternative**: `hf.co/menlo/lucy-gguf:q4_k_m` (good for classification)
- **Avoid**: Large chat models (overkill for simple classification)

## Performance Tips

1. **Batch Processing**: Use goroutines for parallel classification
2. **Caching**: Cache common intents to reduce LLM calls
3. **Fallback Rules**: Use rule-based detection for simple intents
4. **Model Selection**: Use smallest model that meets accuracy requirements

```go
// Simple rule-based fallback
func quickDetect(input string) *Intent {
    lower := strings.ToLower(input)
    if strings.Contains(lower, "thrain") {
        return &Intent{Action: "speak", Character: "Thrain", Known: true}
    }
    return nil // Use LLM for complex cases
}
```

## Troubleshooting

### Inconsistent Results
- Set `Temperature: 0.0` (deterministic)
- Improve system instructions with examples
- Use smaller, specialized models

### Wrong Classifications
- Add more examples to system instructions
- Validate and correct in post-processing
- Use confidence scoring and fallbacks

### Performance Issues
- Use batch processing for multiple inputs
- Cache frequent classifications
- Consider rule-based preprocessing
