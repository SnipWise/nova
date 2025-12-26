---
id: structured-validation
name: Structured Output with Advanced Validation
category: structured
complexity: advanced
sample_source: 25
description: Agent with automatic retry and error feedback for robust structured output
---

# Structured Output with Advanced Validation

## Description

Creates an agent that validates structured output against business rules and automatically retries with error feedback when validation fails, ensuring reliable data extraction.

## Use Cases

- Invoice data extraction
- Form processing with validation
- Data pipelines requiring accuracy
- Extraction with business rules
- Critical data processing

## Complete Code

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// === OUTPUT STRUCTURE ===
type Invoice struct {
	InvoiceNumber string        `json:"invoice_number"`
	Date          string        `json:"date"`
	Vendor        string        `json:"vendor"`
	Items         []InvoiceItem `json:"items"`
	Subtotal      float64       `json:"subtotal"`
	TaxRate       float64       `json:"tax_rate"`
	TaxAmount     float64       `json:"tax_amount"`
	Total         float64       `json:"total"`
	Currency      string        `json:"currency"`
}

type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	LineTotal   float64 `json:"line_total"`
}

// === VALIDATED AGENT ===
type ValidatedAgent struct {
	agent      *structured.Agent
	maxRetries int
}

func NewValidatedAgent(ctx context.Context, maxRetries int) (*ValidatedAgent, error) {
	agent, err := structured.NewAgent(
		ctx,
		agents.Config{
			Name:               "validated-extractor",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "Extract invoice data from text. Ensure all calculations are correct. Return valid JSON only.",
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
		},
		structured.WithOutputSchema(Invoice{}),
	)
	if err != nil {
		return nil, err
	}

	return &ValidatedAgent{
		agent:      agent,
		maxRetries: maxRetries,
	}, nil
}

func (va *ValidatedAgent) ExtractWithValidation(text string) (*Invoice, error) {
	var lastError error
	feedbackHistory := []string{}

	for attempt := 1; attempt <= va.maxRetries; attempt++ {
		fmt.Printf("🔄 Attempt %d/%d\n", attempt, va.maxRetries)

		// Build prompt with feedback if retry
		prompt := fmt.Sprintf("Extract invoice data from this text:\n%s", text)
		if len(feedbackHistory) > 0 {
			prompt += "\n\n⚠️ PREVIOUS ERRORS TO FIX:\n"
			for _, fb := range feedbackHistory {
				prompt += "- " + fb + "\n"
			}
			prompt += "\nPlease fix these issues in your response."
		}

		// Generate
		result, err := va.agent.GenerateCompletion([]messages.Message{
			{Role: roles.User, Content: prompt},
		})
		if err != nil {
			lastError = err
			feedbackHistory = append(feedbackHistory, fmt.Sprintf("Generation error: %v", err))
			continue
		}

		// Parse JSON
		var invoice Invoice
		if err := json.Unmarshal([]byte(result.Response), &invoice); err != nil {
			lastError = fmt.Errorf("JSON parsing error: %v", err)
			feedbackHistory = append(feedbackHistory, fmt.Sprintf("Invalid JSON: %v", err))
			continue
		}

		// Validate business rules
		validationErrors := validateInvoice(&invoice)
		if len(validationErrors) == 0 {
			fmt.Printf("✅ Validation passed on attempt %d\n", attempt)
			return &invoice, nil
		}

		// Add validation errors to feedback
		lastError = fmt.Errorf("validation failed: %v", validationErrors)
		feedbackHistory = append(feedbackHistory, validationErrors...)
		fmt.Printf("❌ Validation errors: %v\n", validationErrors)
	}

	return nil, fmt.Errorf("failed after %d attempts: %v", va.maxRetries, lastError)
}

// === BUSINESS RULE VALIDATION ===
func validateInvoice(inv *Invoice) []string {
	var errors []string

	// Check required fields
	if inv.InvoiceNumber == "" {
		errors = append(errors, "invoice_number is required")
	}
	if inv.Vendor == "" {
		errors = append(errors, "vendor is required")
	}
	if inv.Date == "" {
		errors = append(errors, "date is required")
	}

	// Validate date format
	if inv.Date != "" {
		if _, err := time.Parse("2006-01-02", inv.Date); err != nil {
			errors = append(errors, "date must be in YYYY-MM-DD format")
		}
	}

	// Validate currency
	validCurrencies := map[string]bool{"USD": true, "EUR": true, "GBP": true}
	if !validCurrencies[inv.Currency] {
		errors = append(errors, fmt.Sprintf("currency must be USD, EUR, or GBP, got '%s'", inv.Currency))
	}

	// Validate items
	if len(inv.Items) == 0 {
		errors = append(errors, "at least one item is required")
	}

	// Validate line totals
	calculatedSubtotal := 0.0
	for i, item := range inv.Items {
		expectedLineTotal := float64(item.Quantity) * item.UnitPrice
		if !floatEquals(item.LineTotal, expectedLineTotal) {
			errors = append(errors, fmt.Sprintf(
				"item %d: line_total should be %.2f (qty %d × price %.2f), got %.2f",
				i+1, expectedLineTotal, item.Quantity, item.UnitPrice, item.LineTotal))
		}
		calculatedSubtotal += item.LineTotal
	}

	// Validate subtotal
	if !floatEquals(inv.Subtotal, calculatedSubtotal) {
		errors = append(errors, fmt.Sprintf(
			"subtotal should be %.2f (sum of line totals), got %.2f",
			calculatedSubtotal, inv.Subtotal))
	}

	// Validate tax calculation
	expectedTax := inv.Subtotal * inv.TaxRate / 100
	if !floatEquals(inv.TaxAmount, expectedTax) {
		errors = append(errors, fmt.Sprintf(
			"tax_amount should be %.2f (%.2f × %.2f%%), got %.2f",
			expectedTax, inv.Subtotal, inv.TaxRate, inv.TaxAmount))
	}

	// Validate total
	expectedTotal := inv.Subtotal + inv.TaxAmount
	if !floatEquals(inv.Total, expectedTotal) {
		errors = append(errors, fmt.Sprintf(
			"total should be %.2f (subtotal + tax), got %.2f",
			expectedTotal, inv.Total))
	}

	return errors
}

func floatEquals(a, b float64) bool {
	const epsilon = 0.01
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

// === MAIN ===
func main() {
	ctx := context.Background()

	agent, err := NewValidatedAgent(ctx, 3) // Max 3 retries
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	invoiceText := `
	INVOICE #INV-2024-0042
	Date: January 15, 2024
	From: Acme Corp
	
	Items:
	1. Widget Pro (qty: 5) @ $29.99 each
	2. Gadget Plus (qty: 2) @ $49.99 each
	3. Service Fee (qty: 1) @ $25.00
	
	Tax Rate: 8.5%
	Currency: USD
	`

	fmt.Println("📄 Invoice text:")
	fmt.Println(invoiceText)
	fmt.Println(strings.Repeat("-", 40))

	invoice, err := agent.ExtractWithValidation(invoiceText)
	if err != nil {
		fmt.Printf("❌ Extraction failed: %v\n", err)
		return
	}

	fmt.Println("\n✅ Extracted and validated invoice:")
	prettyJSON, _ := json.MarshalIndent(invoice, "", "  ")
	fmt.Println(string(prettyJSON))
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "ai/qwen2.5:1.5B-F16"
TEMPERATURE: 0.0
MAX_RETRIES: 3
```

## Customization

### With Logging

```go
type ValidationLog struct {
    Attempt    int
    Errors     []string
    Response   string
    Timestamp  time.Time
}

var validationLogs []ValidationLog

func (va *ValidatedAgent) ExtractWithLogging(text string) (*Invoice, error) {
    // ... during each attempt:
    validationLogs = append(validationLogs, ValidationLog{
        Attempt:   attempt,
        Errors:    validationErrors,
        Response:  result.Response,
        Timestamp: time.Now(),
    })
    // ...
}

// Export logs for analysis
func exportLogs(filename string) error {
    data, _ := json.MarshalIndent(validationLogs, "", "  ")
    return os.WriteFile(filename, data, 0644)
}
```

### With Progressive Relaxation

```go
func (va *ValidatedAgent) ExtractWithRelaxation(text string) (*Invoice, error) {
    strictnessLevels := []string{"strict", "normal", "relaxed"}
    
    for _, level := range strictnessLevels {
        invoice, err := va.extractWithLevel(text, level)
        if err == nil {
            return invoice, nil
        }
        fmt.Printf("Failed at %s level, trying next...\n", level)
    }
    
    return nil, fmt.Errorf("failed at all strictness levels")
}

func validateAtLevel(inv *Invoice, level string) []string {
    var errors []string
    
    switch level {
    case "strict":
        // All validations
        errors = validateInvoice(inv)
    case "normal":
        // Skip calculation checks
        if inv.InvoiceNumber == "" {
            errors = append(errors, "invoice_number required")
        }
        // ...
    case "relaxed":
        // Only check required fields exist
        if inv.InvoiceNumber == "" {
            errors = append(errors, "invoice_number required")
        }
    }
    
    return errors
}
```

## Important Notes

- Retry with feedback significantly improves accuracy
- Limit retries to avoid infinite loops (3-5 recommended)
- Log all attempts for debugging and analysis
- Use specific error messages to guide the model
- Consider fallback strategies for persistent failures
