---
id: confirmation-tools
name: Tools Agent with Confirmation
category: tools
complexity: advanced
sample_source: 47
description: Agent that requests human confirmation before executing sensitive tools
---

# Tools Agent with Confirmation (Human-in-the-Loop)

## Description

Creates an agent that requests user confirmation before executing sensitive or irreversible actions, implementing a human-in-the-loop pattern.

## Use Cases

- Financial transactions
- Data deletion operations
- Email/message sending
- System configuration changes
- Any irreversible action

## Complete Code

```go
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// Tools requiring confirmation
var sensitiveTools = map[string]bool{
	"send_email":     true,
	"delete_file":    true,
	"make_payment":   true,
	"update_config":  true,
}

var scanner *bufio.Scanner

func main() {
	ctx := context.Background()
	scanner = bufio.NewScanner(os.Stdin)

	// === DEFINE TOOLS ===
	availableTools := []*tools.Tool{
		tools.NewTool("get_balance").
			SetDescription("Get current account balance").
			AddParameter("account_id", "string", "Account identifier", true),

		tools.NewTool("make_payment").
			SetDescription("Make a payment transfer").
			AddParameter("to", "string", "Recipient account", true).
			AddParameter("amount", "number", "Amount to transfer", true).
			AddParameter("currency", "string", "Currency code", true),

		tools.NewTool("send_email").
			SetDescription("Send an email").
			AddParameter("to", "string", "Recipient email", true).
			AddParameter("subject", "string", "Email subject", true).
			AddParameter("body", "string", "Email body", true),

		tools.NewTool("delete_file").
			SetDescription("Delete a file").
			AddParameter("path", "string", "File path to delete", true),
	}

	// === CREATE AGENT ===
	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "confirmation-assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful assistant. Some actions require user confirmation before execution.",
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
		},
		tools.WithTools(availableTools),
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	fmt.Println("🔐 Tools Agent with Confirmation")
	fmt.Println("Sensitive actions will require your approval")
	fmt.Println("Type 'quit' to exit")
	fmt.Println(strings.Repeat("-", 40))

	for {
		fmt.Print("\n👤 You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "quit" {
			break
		}

		result, err := agent.DetectToolCallsLoop(
			[]messages.Message{{Role: roles.User, Content: input}},
			executeWithConfirmation,
		)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("🤖 Assistant: %s\n", result.LastAssistantMessage)
	}
}

// === EXECUTION WITH CONFIRMATION ===
func executeWithConfirmation(name string, argsJSON string) (string, error) {
	var args map[string]interface{}
	json.Unmarshal([]byte(argsJSON), &args)

	// Check if confirmation is required
	if sensitiveTools[name] {
		if !requestConfirmation(name, args) {
			return `{"status": "cancelled", "reason": "user declined"}`, nil
		}
	}

	// Execute the tool
	return executeTool(name, args)
}

func requestConfirmation(toolName string, args map[string]interface{}) bool {
	fmt.Println("\n" + strings.Repeat("=", 40))
	fmt.Println("⚠️  CONFIRMATION REQUIRED")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Action: %s\n", toolName)
	fmt.Println("Parameters:")
	
	for k, v := range args {
		fmt.Printf("  - %s: %v\n", k, v)
	}
	
	fmt.Println(strings.Repeat("-", 40))
	fmt.Print("Do you approve this action? (yes/no): ")
	
	if !scanner.Scan() {
		return false
	}
	
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	approved := response == "yes" || response == "y"
	
	if approved {
		fmt.Println("✅ Action approved")
	} else {
		fmt.Println("❌ Action cancelled")
	}
	
	return approved
}

func executeTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "get_balance":
		accountID, _ := args["account_id"].(string)
		// Simulated response
		return fmt.Sprintf(`{"account_id": "%s", "balance": 5420.50, "currency": "USD"}`,
			accountID), nil

	case "make_payment":
		to, _ := args["to"].(string)
		amount, _ := args["amount"].(float64)
		currency, _ := args["currency"].(string)
		// Simulated payment
		return fmt.Sprintf(`{"status": "success", "to": "%s", "amount": %.2f, "currency": "%s", "reference": "PAY-12345"}`,
			to, amount, currency), nil

	case "send_email":
		to, _ := args["to"].(string)
		subject, _ := args["subject"].(string)
		// Simulated email
		return fmt.Sprintf(`{"status": "sent", "to": "%s", "subject": "%s", "message_id": "MSG-67890"}`,
			to, subject), nil

	case "delete_file":
		path, _ := args["path"].(string)
		// Simulated deletion
		return fmt.Sprintf(`{"status": "deleted", "path": "%s"}`, path), nil

	default:
		return `{"error": "unknown tool"}`, nil
	}
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
TEMPERATURE: 0.0

# Tools requiring confirmation
SENSITIVE_TOOLS:
  - send_email
  - delete_file
  - make_payment
  - update_config
```

## Customization

### With Confirmation Levels

```go
type ConfirmationLevel int

const (
    NoConfirmation ConfirmationLevel = iota
    SimpleConfirmation
    DetailedConfirmation
    MultiFactorConfirmation
)

var toolConfirmationLevels = map[string]ConfirmationLevel{
    "get_balance":   NoConfirmation,
    "send_email":    SimpleConfirmation,
    "make_payment":  DetailedConfirmation,
    "delete_all":    MultiFactorConfirmation,
}

func requestConfirmationByLevel(level ConfirmationLevel, toolName string, args map[string]interface{}) bool {
    switch level {
    case NoConfirmation:
        return true
    case SimpleConfirmation:
        fmt.Printf("Confirm %s? (y/n): ", toolName)
        // Simple yes/no
    case DetailedConfirmation:
        // Show all details + yes/no
    case MultiFactorConfirmation:
        // Require code or additional verification
    }
    // ...
}
```

### With Audit Logging

```go
type AuditLog struct {
    Timestamp   time.Time
    Tool        string
    Args        map[string]interface{}
    Approved    bool
    ApprovedBy  string
}

var auditLogs []AuditLog

func logAction(toolName string, args map[string]interface{}, approved bool) {
    auditLogs = append(auditLogs, AuditLog{
        Timestamp:  time.Now(),
        Tool:       toolName,
        Args:       args,
        Approved:   approved,
        ApprovedBy: "user", // Could be user ID in production
    })
    
    // Persist to file/database
    saveAuditLogs()
}
```

### With Timeout for Confirmation

```go
func requestConfirmationWithTimeout(toolName string, args map[string]interface{}, timeout time.Duration) bool {
    fmt.Printf("⚠️ Confirm %s? (y/n) [%v timeout]: ", toolName, timeout)
    
    responseCh := make(chan bool, 1)
    
    go func() {
        scanner.Scan()
        response := strings.ToLower(strings.TrimSpace(scanner.Text()))
        responseCh <- (response == "yes" || response == "y")
    }()
    
    select {
    case approved := <-responseCh:
        return approved
    case <-time.After(timeout):
        fmt.Println("\n⏰ Confirmation timeout - action cancelled")
        return false
    }
}
```

## Important Notes

- Always require confirmation for irreversible actions
- Log all confirmations for audit purposes
- Consider timeout for automated environments
- Display clear information about the action before confirmation
- Handle network/system errors gracefully during confirmation
