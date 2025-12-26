---
id: structured-output
name: Structured Output Agent
category: structured
complexity: intermediate
sample_source: 23
description: Agent that returns structured JSON data based on a Go struct definition
---

# Structured Output Agent

## Description

Creates an agent that returns structured data in JSON format, automatically validated against a Go struct schema. Ideal for extracting structured information from unstructured text.

## Use Cases

- Entity extraction from text
- Form data parsing
- Structured API responses
- Data normalization
- Document analysis

## Complete Code

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// === OUTPUT STRUCTURE - CUSTOMIZE HERE ===
// ⚠️ IMPORTANT: Use string for numeric IDs and counters instead of int
// LLMs often generate numbers as strings, which will cause JSON unmarshaling errors
type PersonInfo struct {
	Name       string   `json:"name" description:"Full name of the person"`
	Age        string   `json:"age" description:"Age in years"`
	Email      string   `json:"email" description:"Email address"`
	Occupation string   `json:"occupation" description:"Current job or profession"`
	Skills     []string `json:"skills" description:"List of skills or competencies"`
}

func main() {
	ctx := context.Background()

	// === CREATE STRUCTURED AGENT ===
	agent, err := structured.NewAgent(
		ctx,
		agents.Config{
			Name:               "structured-extractor",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "Extract structured information from the provided text. Return only valid JSON matching the schema.",
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m", // Recommended for structured output
			Temperature: models.Float64(0.0),                 // 0 for determinism
		},
		structured.WithOutputSchema(PersonInfo{}), // Pass struct type
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	// === TEST INPUT ===
	text := `
	I met John Smith yesterday at the tech conference. He's 32 years old and 
	works as a Senior Software Engineer at Google. You can reach him at 
	john.smith@gmail.com. He mentioned he's skilled in Go, Python, Kubernetes,
	and machine learning.
	`

	fmt.Println("📝 Input text:")
	fmt.Println(text)
	fmt.Println("\n🔄 Extracting structured data...")

	// Generate structured output
	result, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: fmt.Sprintf("Extract person information from this text:\n%s", text)},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Parse JSON response
	var person PersonInfo
	if err := json.Unmarshal([]byte(result.Response), &person); err != nil {
		fmt.Printf("JSON parsing error: %v\n", err)
		fmt.Printf("Raw response: %s\n", result.Response)
		return
	}

	// Display extracted data
	fmt.Println("\n✅ Extracted data:")
	fmt.Printf("   Name: %s\n", person.Name)
	fmt.Printf("   Age: %s\n", person.Age)
	fmt.Printf("   Email: %s\n", person.Email)
	fmt.Printf("   Occupation: %s\n", person.Occupation)
	fmt.Printf("   Skills: %v\n", person.Skills)
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"  # Recommended for structured output
TEMPERATURE: 0.0  # Critical for structured output

# Alternative models:
# - ai/qwen2.5:1.5B-F16 (balanced)
# - hf.co/menlo/lucy-gguf:q4_k_m (fast)
```

## ⚠️ Important Best Practices

### Use Strings for Numeric Fields

**Problem:** LLMs often generate numbers as JSON strings, causing unmarshaling errors:
```
Error: json: cannot unmarshal string into Go struct field Action.id of type int
```

**Solution:** Use `string` instead of `int` for numeric fields:

```go
// ❌ BAD - Will cause errors
type Task struct {
    ID       int    `json:"id"`
    Priority int    `json:"priority"`
    Count    int    `json:"count"`
}

// ✅ GOOD - Works reliably
type Task struct {
    ID       string `json:"id"`        // "1", "2", "3"
    Priority string `json:"priority"`  // "1", "2", "3" or "high", "medium", "low"
    Count    string `json:"count"`     // "5", "10"
}
```

**Why?**
- LLMs are text models and naturally output everything as strings
- JSON schema inference is not always precise
- Strings are more flexible: can be "1", "A1", "TASK-001", etc.
- You can convert to int later if needed: `strconv.Atoi(task.ID)`

### Converting Strings to Numbers (if needed)

```go
import "strconv"

// After extraction, convert if needed
age, err := strconv.Atoi(person.Age)
if err != nil {
    // Handle invalid number
}

count, err := strconv.ParseFloat(product.Price, 64)
if err != nil {
    // Handle invalid number
}
```

## Customization

### Complex Structure with Nested Types

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
    ZipCode string `json:"zip_code"`
}

type Company struct {
    Name      string   `json:"name"`
    Industry  string   `json:"industry"`
    Employees int      `json:"employees"`
    Address   Address  `json:"address"`
    Founded   string   `json:"founded" description:"Year founded (YYYY)"`
}

// Use with WithOutputSchema(Company{})
```

### With Multiple Entities Extraction

```go
type ExtractedEntities struct {
    People    []PersonInfo  `json:"people"`
    Companies []CompanyInfo `json:"companies"`
    Dates     []string      `json:"dates"`
    Locations []string      `json:"locations"`
}

agent, _ := structured.NewAgent(
    ctx,
    config,
    modelConfig,
    structured.WithOutputSchema(ExtractedEntities{}),
)
```

### With Optional Fields

```go
type ProductInfo struct {
    Name        string   `json:"name"`
    Price       float64  `json:"price"`
    Currency    string   `json:"currency"`
    Description *string  `json:"description,omitempty"` // Optional
    Tags        []string `json:"tags,omitempty"`        // Optional
}
```

### With Validation After Extraction

```go
func validatePerson(p PersonInfo) error {
    if p.Name == "" {
        return fmt.Errorf("name is required")
    }
    if p.Age < 0 || p.Age > 150 {
        return fmt.Errorf("invalid age: %d", p.Age)
    }
    if !strings.Contains(p.Email, "@") {
        return fmt.Errorf("invalid email: %s", p.Email)
    }
    return nil
}

// After extraction
if err := validatePerson(person); err != nil {
    fmt.Printf("Validation failed: %v\n", err)
    // Optionally retry with feedback
}
```

## Important Notes

- Temperature 0.0 is essential for consistent JSON output
- Larger models (7B+) produce better structured output
- Use `description` tags to guide the model
- Always validate JSON parsing before using data
- Consider retry logic for parsing failures (see structured-validation)
