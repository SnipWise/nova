---
id: structured-output
name: Structured Output Agent
category: structured
complexity: intermediate
sample_source: 23
description: Agent that returns structured data validated against a Go struct using generics
---

# Structured Output Agent

## Description

Creates an agent that returns structured data in JSON format, automatically validated against a Go struct schema using Go generics. The agent ensures type-safe extraction of structured information from unstructured text.

## Use Cases

- Entity extraction from text
- Information normalization
- Data classification
- Form data parsing
- Structured API responses
- Knowledge extraction

## Complete Code

```go
package main

import (
	"context"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// === DEFINE YOUR OUTPUT STRUCTURE ===
// This is the schema the LLM will follow
type Country struct {
	Name       string   `json:"name"`
	Capital    string   `json:"capital"`
	Population int      `json:"population"`
	Languages  []string `json:"languages"`
}

func main() {
	ctx := context.Background()

	// === CREATE STRUCTURED AGENT WITH GENERIC TYPE ===
	agent, err := structured.NewAgent[Country](
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
			Your name is Bob.
			You are an assistant that answers questions about countries around the world.
			`,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),  // Deterministic for structured output
		},
	)
	if err != nil {
		panic(err)
	}

	// === GENERATE STRUCTURED DATA ===
	response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: "Tell me about Canada."},
	})

	if err != nil {
		panic(err)
	}

	// === DISPLAY RESULTS ===
	// response is already a *Country struct
	display.NewLine()
	display.Separator()
	display.Title("Response")
	display.KeyValue("Name", response.Name)
	display.KeyValue("Capital", response.Capital)
	display.KeyValue("Population", conversion.IntToString(response.Population))
	display.KeyValue("Languages", strings.Join(response.Languages, ", "))
	display.NewLine()
	display.Separator()
	display.KeyValue("Finish reason", finishReason)
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TEMPERATURE: 0.0  # Always use 0.0 for structured output
```

## How It Works

### 1. Define Your Struct

```go
type Country struct {
	Name       string   `json:"name"`
	Capital    string   `json:"capital"`
	Population int      `json:"population"`
	Languages  []string `json:"languages"`
}
```

**JSON tags are mandatory** - they define the field names in the output.

### 2. Create Agent with Generic Type

```go
agent, err := structured.NewAgent[Country](ctx, config, models)
```

The `[Country]` generic parameter tells the agent what structure to return.

### 3. Generate Structured Data

```go
response, finishReason, err := agent.GenerateStructuredData(messages)
```

Returns a `*Country` struct (already parsed and validated).

## Customization

### Different Data Structures

#### Person Information
```go
type Person struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Age       int      `json:"age"`
	Email     string   `json:"email"`
	Skills    []string `json:"skills"`
}

agent, _ := structured.NewAgent[Person](ctx, config, models)
response, _, _ := agent.GenerateStructuredData(messages)
fmt.Printf("Name: %s %s\n", response.FirstName, response.LastName)
```

#### Product Information
```go
type Product struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	InStock     bool    `json:"in_stock"`
	Features    []string `json:"features"`
}

agent, _ := structured.NewAgent[Product](ctx, config, models)
```

#### Event Information
```go
type Event struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Attendees   int    `json:"attendees"`
	Description string `json:"description"`
}

agent, _ := structured.NewAgent[Event](ctx, config, models)
```

### Complex Nested Structures

```go
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type Company struct {
	Name      string   `json:"name"`
	Address   Address  `json:"address"`
	Employees int      `json:"employees"`
	Industries []string `json:"industries"`
}

agent, _ := structured.NewAgent[Company](ctx, config, models)
```

### Better System Instructions

```go
agents.Config{
	SystemInstructions: `
	Extract country information from the user's question.

	Rules:
	- name: Full official name of the country
	- capital: Name of the capital city
	- population: Population in millions (integer)
	- languages: List of official languages

	Be accurate and provide real data.
	`,
}
```

## Important Notes

- **Generic Type**: Must specify `[YourStruct]` when creating agent
- **JSON Tags**: Required on all struct fields
- **Temperature**: Always use `0.0` for deterministic structured output
- **Return Type**: `GenerateStructuredData` returns `*YourStruct`, not JSON string
- **Validation**: Automatic JSON schema validation
- **Type Safety**: Compile-time type checking with generics

## Field Types

Supported Go types in your struct:

```go
type Example struct {
	StringField  string   `json:"string_field"`   // Text
	IntField     int      `json:"int_field"`      // Integer
	FloatField   float64  `json:"float_field"`    // Decimal
	BoolField    bool     `json:"bool_field"`     // True/False
	ArrayField   []string `json:"array_field"`    // List
	NestedField  Address  `json:"nested_field"`   // Nested struct
}
```

## Error Handling

```go
response, finishReason, err := agent.GenerateStructuredData(messages)
if err != nil {
	log.Printf("Generation error: %v", err)
	return
}

// response is guaranteed to match your struct if no error
fmt.Printf("Country: %s\n", response.Name)
```

## Processing Multiple Items

```go
countries := []string{"France", "Germany", "Spain"}

for _, country := range countries {
	response, _, _ := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: fmt.Sprintf("Tell me about %s", country)},
	})

	fmt.Printf("%s - Capital: %s\n", response.Name, response.Capital)
}
```

## Best Practices

1. **Use descriptive field names**: Clear JSON tag names help the LLM
2. **Keep structures simple**: Fewer fields = better accuracy
3. **Always use temperature 0.0**: Ensures consistent formatting
4. **Validate business logic**: Check if data makes sense for your use case
5. **Handle edge cases**: LLM might return unexpected values

## Related Patterns

- For array output: See `structured-schema.md`
- For validation: See `structured-validation.md`
- For intent detection: See `orchestrator/topic-detection.md`
