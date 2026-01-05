# Structured Agent

## Description

The **Structured Agent** is a specialized agent for generating structured data in JSON format. It guarantees that the LLM response follows a strict JSON schema defined by a Go struct, using OpenAI's structured output functionality.

## Features

- **Guaranteed structured output** : LLM always returns valid JSON conforming to the schema
- **Strong typing** : Uses Go generics to define output type
- **Automatic schema generation** : Automatically converts Go struct to JSON Schema
- **History management** : Support for contextual conversation (optional)
- **Strict validation** : OpenAI strict mode to guarantee schema compliance

## Use cases

The Structured Agent is used for:
- **Information extraction** : Extract structured data from text
- **Classification** : Categorize content with defined fields
- **Parsing** : Convert natural language to data structures
- **APIs** : Generate JSON responses for APIs
- **Data validation** : Guarantee consistent output format

## Creating a Structured Agent

### Basic syntax

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/structured"
    "github.com/snipwise/nova/nova-sdk/models"
)

// Define the output structure
type Country struct {
    Name       string   `json:"name"`
    Capital    string   `json:"capital"`
    Population int      `json:"population"`
    Languages  []string `json:"languages"`
}

ctx := context.Background()

// Create agent with output type
agent, err := structured.NewAgent[Country](
    ctx,
    agents.Config{
        Name:               "StructuredAgent",
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "You are an assistant that provides country information.",
    },
    models.Config{
        Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature: models.Float64(0.0), // Deterministic for structured data
    },
)
if err != nil {
    log.Fatal(err)
}
```

## Main methods

### GenerateStructuredData - Generate structured data

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

// Generate structured data
response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Tell me about Canada."},
})

if err != nil {
    log.Fatal(err)
}

// response is of type *Country
fmt.Printf("Name: %s\n", response.Name)
fmt.Printf("Capital: %s\n", response.Capital)
fmt.Printf("Population: %d\n", response.Population)
fmt.Printf("Languages: %v\n", response.Languages)
fmt.Printf("Finish reason: %s\n", finishReason)
```

### Message management

```go
// Add a message
agent.AddMessage(roles.User, "Question...")

// Add multiple messages
messages := []messages.Message{
    {Role: roles.User, Content: "Question 1"},
    {Role: roles.Assistant, Content: "Answer 1"},
}
agent.AddMessages(messages)

// Get all messages
allMessages := agent.GetMessages()

// Reset messages
agent.ResetMessages()

// Export to JSON
jsonData, err := agent.ExportMessagesToJSON()
```

### Getters and Setters

```go
// Configuration
config := agent.GetConfig()
agent.SetConfig(newConfig)

modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)

// Information
name := agent.GetName()
modelID := agent.GetModelID()
kind := agent.Kind() // Returns agents.Structured

// Context
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requests/Responses (debugging)
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
prettyRequest, _ := agent.GetLastRequestJSON()
prettyResponse, _ := agent.GetLastResponseJSON()
```

## StructuredResult structure

```go
type StructuredResult[Output any] struct {
    Data         *Output  // The generated structured data
    FinishReason string   // Finish reason ("stop", "length", etc.)
}
```

## Supported structure types

### Simple structure

```go
type Person struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

agent, _ := structured.NewAgent[Person](ctx, agentConfig, modelConfig)
```

### Structure with slices

```go
type Book struct {
    Title   string   `json:"title"`
    Authors []string `json:"authors"`
    Year    int      `json:"year"`
}

agent, _ := structured.NewAgent[Book](ctx, agentConfig, modelConfig)
```

### Slice of structures

```go
type Product struct {
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

// Generates an array of products
agent, _ := structured.NewAgent[[]Product](ctx, agentConfig, modelConfig)

response, _, _ := agent.GenerateStructuredData(messages)
// response is of type *[]Product
for _, product := range *response {
    fmt.Printf("%s: $%.2f\n", product.Name, product.Price)
}
```

### Nested structures

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type Company struct {
    Name    string  `json:"name"`
    Address Address `json:"address"`
    Revenue float64 `json:"revenue"`
}

agent, _ := structured.NewAgent[Company](ctx, agentConfig, modelConfig)
```

## Complete example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/structured"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

// Define output structure
type MovieInfo struct {
    Title    string   `json:"title"`
    Director string   `json:"director"`
    Year     int      `json:"year"`
    Genres   []string `json:"genres"`
    Rating   float64  `json:"rating"`
}

func main() {
    ctx := context.Background()

    // Create agent
    agent, err := structured.NewAgent[MovieInfo](
        ctx,
        agents.Config{
            Name:               "MovieAgent",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You provide information about movies.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.0),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Generate structured data
    response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
        {Role: roles.User, Content: "Tell me about the movie Inception."},
    })
    if err != nil {
        log.Fatal(err)
    }

    // Display results
    fmt.Printf("Title: %s\n", response.Title)
    fmt.Printf("Director: %s\n", response.Director)
    fmt.Printf("Year: %d\n", response.Year)
    fmt.Printf("Genres: %v\n", response.Genres)
    fmt.Printf("Rating: %.1f/10\n", response.Rating)
    fmt.Printf("Finish reason: %s\n", finishReason)
}
```

**Expected output** :
```
Title: Inception
Director: Christopher Nolan
Year: 2010
Genres: [Science Fiction Thriller Action]
Rating: 8.8/10
Finish reason: stop
```

## Automatic JSON Schema generation

The Structured Agent automatically converts your Go structs to JSON Schema:

```go
type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

// Automatically generates the schema:
// {
//   "type": "object",
//   "properties": {
//     "name": {"type": "string"},
//     "age": {"type": "integer"},
//     "email": {"type": "string"}
//   },
//   "required": ["name", "age", "email"]
// }
```

**Supported JSON tags** :
- `json:"fieldname"` : Defines the field name in JSON
- `json:"-"` : Excludes field from JSON schema

## Notes

- **Kind** : Returns `agents.Structured`
- **Temperature** : Use 0.0 for deterministic results
- **Compatible models** : All models supporting OpenAI structured response format
- **Strict validation** : Strict mode guarantees output exactly matches schema
- **Generic types** : Uses Go generics (Go 1.18+)
- **History** : Configurable via `KeepConversationHistory` in `agents.Config`

## Recommendations

### Best practices

1. **Temperature 0.0** : Use low temperature for consistent structured data
2. **Clear field names** : Use descriptive field names in English
3. **JSON tags** : Always define JSON tags to control serialization
4. **Validation** : Check that response is not nil before use
5. **Simple types** : Prefer standard Go types (string, int, float64, bool)

### Validation example

```go
response, finishReason, err := agent.GenerateStructuredData(messages)
if err != nil {
    log.Printf("Error generating data: %v", err)
    return
}

if response == nil {
    log.Println("No data returned")
    return
}

// Use the data
fmt.Printf("Result: %+v\n", response)
```

### Error handling

```go
response, finishReason, err := agent.GenerateStructuredData(messages)
if err != nil {
    // Communication or parsing error
    log.Printf("Generation failed: %v", err)
    return
}

if finishReason != "stop" {
    // Generation stopped prematurely
    log.Printf("Unexpected finish reason: %s", finishReason)
}
```

## History configuration

### Without history (default)

```go
agent, _ := structured.NewAgent[Output](
    ctx,
    agents.Config{
        EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions:      "Instructions...",
        KeepConversationHistory: false, // Each call is independent
    },
    modelConfig,
)
```

### With history

```go
agent, _ := structured.NewAgent[Output](
    ctx,
    agents.Config{
        EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions:      "Instructions...",
        KeepConversationHistory: true, // Maintains context
    },
    modelConfig,
)

// First call
response1, _, _ := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Question 1"},
})

// Second call - has access to first call context
response2, _, _ := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Question 2 based on answer 1"},
})
```
