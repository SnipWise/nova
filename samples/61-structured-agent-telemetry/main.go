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
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// Person represents a structured person data
type Person struct {
	Name       string   `json:"name" jsonschema:"description=Full name of the person"`
	Age        int      `json:"age" jsonschema:"description=Age in years"`
	Occupation string   `json:"occupation" jsonschema:"description=Current job or profession"`
	Hobbies    []string `json:"hobbies" jsonschema:"description=List of hobbies"`
	Location   Location `json:"location" jsonschema:"description=Where the person lives"`
}

// Location represents geographical location
type Location struct {
	City    string `json:"city" jsonschema:"description=City name"`
	Country string `json:"country" jsonschema:"description=Country name"`
}

func main() {
	ctx := context.Background()

	// Create structured agent for Person type
	agent, err := structured.NewAgent[Person](
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "Extract structured information about people from text.",
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.2).
			WithMaxTokens(1000),
	)
	if err != nil {
		panic(err)
	}

	display.Title("Structured Agent with Telemetry")
	display.Separator()

	// Generate structured data
	userPrompt := `Extract information about: John Smith is a 35-year-old software engineer
	living in San Francisco, USA. He enjoys hiking, photography, and playing guitar in his free time.`

	display.Info("Input Text:")
	fmt.Println(userPrompt)
	display.NewLine()

	display.Info("Generating structured data...")
	person, finishReason, err := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: userPrompt},
	})
	if err != nil {
		panic(err)
	}

	display.Success("Structured Data Generated!")
	display.KeyValue("  Finish Reason", finishReason)
	display.NewLine()

	// Display extracted data
	display.Info("Extracted Person Data:")
	personJSON, _ := json.MarshalIndent(person, "  ", "  ")
	fmt.Println(string(personJSON))

	// Display telemetry
	display.NewLine()
	display.Separator()
	display.Title("📊 Structured Generation Telemetry")
	display.Separator()

	// Request metadata
	reqMeta := agent.GetLastRequestMetadata()
	display.Info("Request Metadata:")
	display.KeyValue("  Model", reqMeta.Model)
	display.KeyValue("  Context Length", fmt.Sprintf("%d bytes", reqMeta.ContextLength))
	display.KeyValue("  Temperature", fmt.Sprintf("%.2f", reqMeta.Temperature))
	display.KeyValue("  Max Tokens", fmt.Sprintf("%d", reqMeta.MaxTokens))
	display.NewLine()

	// Response metadata
	respMeta := agent.GetLastResponseMetadata()
	display.Info("Response Metadata:")
	display.KeyValue("  Response ID", respMeta.ID)
	display.KeyValue("  Finish Reason", respMeta.FinishReason)
	display.KeyValue("  Prompt Tokens", fmt.Sprintf("%d", respMeta.PromptTokens))
	display.KeyValue("  Completion Tokens", fmt.Sprintf("%d", respMeta.CompletionTokens))
	display.KeyValue("  Total Tokens", fmt.Sprintf("%d", respMeta.TotalTokens))
	display.KeyValue("  Response Time", fmt.Sprintf("%d ms", respMeta.ResponseTime))

	// Structured output metrics
	display.NewLine()
	display.Info("Structured Output Metrics:")
	display.KeyValue("  Fields Extracted", "5 (name, age, occupation, hobbies, location)")
	display.KeyValue("  Nested Objects", "1 (location)")
	display.KeyValue("  Array Fields", "1 (hobbies)")

	// Calculate tokens per field
	tokensPerField := float64(respMeta.CompletionTokens) / 5.0
	display.KeyValue("  Avg Tokens/Field", fmt.Sprintf("%.1f", tokensPerField))

	// Session statistics
	display.NewLine()
	display.Info("Session Statistics:")
	display.KeyValue("  Total Tokens Used", fmt.Sprintf("%d", agent.GetTotalTokensUsed()))

	// Show JSON schema in request
	display.NewLine()
	display.Separator()
	display.Title("📋 Request with JSON Schema")
	display.Separator()
	reqJSON, _ := agent.GetLastRequestJSON()
	// Parse and pretty print just the schema part
	var reqData map[string]interface{}
	json.Unmarshal([]byte(reqJSON), &reqData)
	if schema, ok := reqData["response_format"]; ok {
		schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
		fmt.Println(string(schemaJSON))
	}

	// Full response
	display.NewLine()
	display.Separator()
	display.Title("📄 Full Response JSON")
	display.Separator()
	respJSON, _ := agent.GetLastResponseJSON()
	fmt.Println(respJSON)

	display.NewLine()
	display.Success("Structured agent telemetry example completed!")
}
