package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// Define tools using the builder pattern
	weatherTool := tools.NewTool("get_weather").
		SetDescription("Get current weather for a location").
		AddParameter("location", "string", "The city name", true)

	calculatorTool := tools.NewTool("calculate").
		SetDescription("Perform mathematical calculations").
		AddParameter("expression", "string", "Mathematical expression to evaluate", true)

	// Create tools agent
	agent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful assistant with access to tools.",
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.5).
			WithMaxTokens(2000),
		tools.WithTools([]*tools.Tool{weatherTool, calculatorTool}),
	)
	if err != nil {
		panic(err)
	}

	display.Title("Tools Agent with Telemetry")
	display.Separator()

	// Tool callback
	toolCallback := func(functionName string, arguments string) (string, error) {
		display.Info(fmt.Sprintf("🔧 Tool called: %s", functionName))
		display.KeyValue("  Arguments", arguments)

		switch functionName {
		case "get_weather":
			return `{"temperature": 22, "condition": "sunny", "humidity": 65}`, nil
		case "calculate":
			return `{"result": 42}`, nil
		default:
			return "", fmt.Errorf("unknown function: %s", functionName)
		}
	}

	// Execute with tool calls
	display.Info("Asking: What's the weather in Paris and what is 21 * 2?")
	display.NewLine()

	result, err := agent.DetectParallelToolCalls(
		[]messages.Message{
			{Role: roles.User, Content: "What's the weather in Paris and what is 21 * 2?"},
		},
		toolCallback,
	)
	if err != nil {
		panic(err)
	}

	display.KeyValue("Finish Reason", result.FinishReason)
	display.KeyValue("Tool Results", fmt.Sprintf("%d tools called", len(result.Results)))
	display.KeyValue("Final Message", result.LastAssistantMessage)

	// Display telemetry
	display.NewLine()
	display.Separator()
	display.Title("📊 Telemetry Data")
	display.Separator()

	// Request metadata
	reqMeta := agent.GetLastRequestMetadata()
	display.Info("Last Request:")
	display.KeyValue("  Model", reqMeta.Model)
	display.KeyValue("  Context Length", fmt.Sprintf("%d bytes", reqMeta.ContextLength))
	display.KeyValue("  Temperature", fmt.Sprintf("%.2f", reqMeta.Temperature))
	display.KeyValue("  Timestamp", reqMeta.Timestamp.Format("15:04:05"))

	// Response metadata
	display.NewLine()
	respMeta := agent.GetLastResponseMetadata()
	display.Info("Last Response:")
	display.KeyValue("  Response ID", respMeta.ID)
	display.KeyValue("  Finish Reason", respMeta.FinishReason)
	display.KeyValue("  Prompt Tokens", fmt.Sprintf("%d", respMeta.PromptTokens))
	display.KeyValue("  Completion Tokens", fmt.Sprintf("%d", respMeta.CompletionTokens))
	display.KeyValue("  Total Tokens", fmt.Sprintf("%d", respMeta.TotalTokens))
	display.KeyValue("  Response Time", fmt.Sprintf("%d ms", respMeta.ResponseTime))

	// Session stats
	display.NewLine()
	display.Info("Session Statistics:")
	display.KeyValue("  Total Tokens Used", fmt.Sprintf("%d", agent.GetTotalTokensUsed()))

	// Full JSON exports
	display.NewLine()
	display.Separator()
	display.Title("📄 Full Request JSON")
	display.Separator()
	reqJSON, _ := agent.GetLastRequestJSON()
	fmt.Println(reqJSON)

	display.NewLine()
	display.Separator()
	display.Title("📄 Full Response JSON")
	display.Separator()
	respJSON, _ := agent.GetLastResponseJSON()
	fmt.Println(respJSON)

	display.NewLine()
	display.Success("Tools agent telemetry example completed!")
}
