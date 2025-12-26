---
id: structured-schema
name: Structured Output with JSON Schema
category: structured
complexity: intermediate
sample_source: 24
description: Agent using explicit JSON Schema for fine-grained control over output format
---

# Structured Output with JSON Schema

## Description

Creates an agent that uses an explicit JSON Schema for maximum control over output format, including constraints like min/max values, enums, patterns, and required fields.

## Use Cases

- Strict API contracts
- Form validation
- Data with specific constraints
- Enum-based classifications
- Pattern-matched fields (dates, emails, etc.)

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

func main() {
	ctx := context.Background()

	// === DEFINE JSON SCHEMA ===
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"product_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the product",
				"minLength":   1,
				"maxLength":   100,
			},
			"price": map[string]interface{}{
				"type":        "number",
				"description": "Product price",
				"minimum":     0,
				"maximum":     1000000,
			},
			"currency": map[string]interface{}{
				"type":        "string",
				"description": "Currency code",
				"enum":        []string{"USD", "EUR", "GBP", "JPY"},
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Product category",
				"enum":        []string{"electronics", "clothing", "food", "books", "other"},
			},
			"in_stock": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether product is in stock",
			},
			"rating": map[string]interface{}{
				"type":        "number",
				"description": "Product rating",
				"minimum":     0,
				"maximum":     5,
			},
			"tags": map[string]interface{}{
				"type":        "array",
				"description": "Product tags",
				"items": map[string]interface{}{
					"type": "string",
				},
				"maxItems": 10,
			},
		},
		"required": []string{"product_name", "price", "currency", "category", "in_stock"},
	}

	// === CREATE AGENT WITH SCHEMA ===
	agent, err := structured.NewAgent(
		ctx,
		agents.Config{
			Name:               "schema-extractor",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "Extract product information from text. Follow the JSON schema exactly.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
		},
		structured.WithJSONSchema(schema),
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	// === TEST INPUT ===
	text := `
	New Apple iPhone 15 Pro now available! Priced at $999 USD. 
	This electronics product is currently in stock with a 4.8 star rating.
	Tags: smartphone, apple, premium, 5G
	`

	fmt.Println("📝 Input text:")
	fmt.Println(text)
	fmt.Println("\n🔄 Extracting with JSON Schema...")

	result, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: fmt.Sprintf("Extract product info:\n%s", text)},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Parse and display
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(result.Response), &product); err != nil {
		fmt.Printf("JSON parsing error: %v\n", err)
		return
	}

	fmt.Println("\n✅ Extracted product:")
	prettyJSON, _ := json.MarshalIndent(product, "", "  ")
	fmt.Println(string(prettyJSON))

	// Validate against schema
	if err := validateAgainstSchema(product, schema); err != nil {
		fmt.Printf("⚠️ Schema validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Schema validation passed")
	}
}

// === SCHEMA VALIDATION ===
func validateAgainstSchema(data map[string]interface{}, schema map[string]interface{}) error {
	// Check required fields
	if required, ok := schema["required"].([]string); ok {
		for _, field := range required {
			if _, exists := data[field]; !exists {
				return fmt.Errorf("missing required field: %s", field)
			}
		}
	}

	// Check properties
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		for field, value := range data {
			if propSchema, exists := properties[field].(map[string]interface{}); exists {
				if err := validateField(field, value, propSchema); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateField(name string, value interface{}, schema map[string]interface{}) error {
	expectedType, _ := schema["type"].(string)

	switch expectedType {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("field %s: expected string", name)
		}
		if enum, ok := schema["enum"].([]string); ok {
			valid := false
			for _, e := range enum {
				if str == e {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("field %s: value '%s' not in enum", name, str)
			}
		}

	case "number":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("field %s: expected number", name)
		}
		if min, ok := schema["minimum"].(float64); ok && num < min {
			return fmt.Errorf("field %s: value %f below minimum %f", name, num, min)
		}
		if max, ok := schema["maximum"].(float64); ok && num > max {
			return fmt.Errorf("field %s: value %f above maximum %f", name, num, max)
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s: expected boolean", name)
		}
	}

	return nil
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "ai/qwen2.5:1.5B-F16"
TEMPERATURE: 0.0
```

## Customization

### Schema with Patterns (Regex)

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "email": map[string]interface{}{
            "type":    "string",
            "pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
        },
        "phone": map[string]interface{}{
            "type":    "string",
            "pattern": "^\\+?[1-9]\\d{1,14}$",
        },
        "date": map[string]interface{}{
            "type":   "string",
            "format": "date", // ISO 8601 date format
        },
    },
}
```

### Dynamic Schema Generation

```go
func generateSchemaFromStruct(v interface{}) map[string]interface{} {
    t := reflect.TypeOf(v)
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    
    properties := make(map[string]interface{})
    required := []string{}
    
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        jsonTag := field.Tag.Get("json")
        if jsonTag == "" || jsonTag == "-" {
            continue
        }
        
        // Parse json tag
        jsonName := strings.Split(jsonTag, ",")[0]
        
        // Determine type
        fieldSchema := map[string]interface{}{
            "type": goTypeToJSONType(field.Type.Kind()),
        }
        
        // Add description if present
        if desc := field.Tag.Get("description"); desc != "" {
            fieldSchema["description"] = desc
        }
        
        properties[jsonName] = fieldSchema
        
        // Check if required (not omitempty)
        if !strings.Contains(jsonTag, "omitempty") {
            required = append(required, jsonName)
        }
    }
    
    return map[string]interface{}{
        "type":       "object",
        "properties": properties,
        "required":   required,
    }
}
```

### Using gojsonschema Library

```go
import "github.com/xeipuuv/gojsonschema"

func validateWithLibrary(data, schema map[string]interface{}) error {
    schemaLoader := gojsonschema.NewGoLoader(schema)
    documentLoader := gojsonschema.NewGoLoader(data)
    
    result, err := gojsonschema.Validate(schemaLoader, documentLoader)
    if err != nil {
        return err
    }
    
    if !result.Valid() {
        var errors []string
        for _, err := range result.Errors() {
            errors = append(errors, err.String())
        }
        return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
    }
    
    return nil
}
```

## Important Notes

- JSON Schema provides fine-grained control over output format
- Use enums for classification tasks
- Pattern validation helps enforce formats (dates, emails, etc.)
- Consider using a validation library for production
- Temperature 0.0 is essential for schema compliance
