---
id: parallel-tools
name: Parallel Tools Agent
category: tools
complexity: intermediate
sample_source: 19
description: Agent that can execute multiple tools in parallel for better performance
---

# Parallel Tools Agent

## Description

Creates an agent capable of detecting and executing multiple tools simultaneously, reducing total execution time when tools are independent.

## Use Cases

- Requests requiring multiple pieces of information
- Independent API calls
- Performance optimization
- Dashboard data aggregation

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
	"sync"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	// === DEFINE TOOLS ===
	availableTools := []*tools.Tool{
		tools.NewTool("get_stock_price").
			SetDescription("Get current stock price").
			AddParameter("symbol", "string", "Stock symbol (e.g., AAPL, GOOGL)", true),

		tools.NewTool("get_company_info").
			SetDescription("Get company information").
			AddParameter("symbol", "string", "Stock symbol", true),

		tools.NewTool("get_market_news").
			SetDescription("Get latest market news").
			AddParameter("category", "string", "News category: 'tech', 'finance', 'general'", false),
	}

	// === CREATE AGENT WITH PARALLEL CALLS ===
	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "parallel-tools-assistant",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a financial assistant. Use multiple tools simultaneously when needed to gather comprehensive information.",
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true), // Enable parallel calls
		},
		tools.WithTools(availableTools),
	)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	fmt.Println("⚡ Parallel Tools Agent")
	fmt.Println("Try: 'Give me info about AAPL including price and news'")
	fmt.Println("Type 'quit' to exit")
	fmt.Println(strings.Repeat("-", 40))

	scanner := bufio.NewScanner(os.Stdin)

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

		startTime := time.Now()

		// Use parallel executor
		result, err := agent.DetectToolCallsLoop(
			[]messages.Message{{Role: roles.User, Content: input}},
			executeToolsParallel,
		)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		elapsed := time.Since(startTime)

		// Display results
		if len(result.Results) > 0 {
			fmt.Printf("🔧 Tools executed (%d in parallel):\n", len(result.Results))
			for _, r := range result.Results {
				fmt.Printf("   - %s\n", r)
			}
		}

		fmt.Printf("🤖 Assistant: %s\n", result.LastAssistantMessage)
		fmt.Printf("⏱️ Total time: %v\n", elapsed)
	}
}

// === PARALLEL EXECUTION ===
func executeToolsParallel(name string, argsJSON string) (string, error) {
	// This function is called for each tool
	// Parallelism is handled by DetectToolCallsLoop
	return executeSingleTool(name, argsJSON)
}

// Alternative: Manual parallel execution
func executeToolsManually(toolCalls []tools.ToolCall) []string {
	results := make([]string, len(toolCalls))
	var wg sync.WaitGroup

	for i, call := range toolCalls {
		wg.Add(1)
		go func(idx int, tc tools.ToolCall) {
			defer wg.Done()
			result, _ := executeSingleTool(tc.Name, tc.Arguments)
			results[idx] = result
		}(i, call)
	}

	wg.Wait()
	return results
}

func executeSingleTool(name string, argsJSON string) (string, error) {
	var args map[string]interface{}
	json.Unmarshal([]byte(argsJSON), &args)

	// Simulate latency (API calls take time)
	time.Sleep(100 * time.Millisecond)

	switch name {
	case "get_stock_price":
		symbol, _ := args["symbol"].(string)
		// Simulated data
		prices := map[string]float64{
			"AAPL": 178.50, "GOOGL": 141.25, "MSFT": 378.90,
		}
		price := prices[symbol]
		if price == 0 {
			price = 100.00
		}
		return fmt.Sprintf(`{"symbol": "%s", "price": %.2f, "currency": "USD"}`,
			symbol, price), nil

	case "get_company_info":
		symbol, _ := args["symbol"].(string)
		return fmt.Sprintf(`{"symbol": "%s", "name": "Company Inc.", "sector": "Technology", "employees": 150000}`,
			symbol), nil

	case "get_market_news":
		category := "general"
		if v, ok := args["category"].(string); ok {
			category = v
		}
		return fmt.Sprintf(`{"category": "%s", "headlines": ["Markets rally on economic data", "Tech sector leads gains"]}`,
			category), nil

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
PARALLEL_TOOL_CALLS: true  # Key setting
```

## Customization

### With Timeout and Error Handling

```go
func executeWithTimeout(name, args string, timeout time.Duration) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    resultCh := make(chan string, 1)
    errCh := make(chan error, 1)

    go func() {
        result, err := executeSingleTool(name, args)
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- result
    }()

    select {
    case result := <-resultCh:
        return result, nil
    case err := <-errCh:
        return "", err
    case <-ctx.Done():
        return `{"error": "timeout"}`, nil
    }
}
```

### With Result Aggregation

```go
type AggregatedResult struct {
    Stock   *StockPrice   `json:"stock,omitempty"`
    Company *CompanyInfo  `json:"company,omitempty"`
    News    *MarketNews   `json:"news,omitempty"`
}

func aggregateResults(results []string) string {
    aggregated := AggregatedResult{}
    
    for _, r := range results {
        // Parse and aggregate based on content
        if strings.Contains(r, "price") {
            json.Unmarshal([]byte(r), &aggregated.Stock)
        }
        // ... etc
    }
    
    output, _ := json.Marshal(aggregated)
    return string(output)
}
```

## Important Notes

- `ParallelToolCalls: true` must be set in model config
- Model must support parallel function calling
- Tools should be independent (no dependencies between them)
- Parallel execution significantly reduces total time
- Handle individual tool failures gracefully
