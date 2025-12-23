package main

import (
	"context"
	"fmt"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/base"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// CustomTelemetryLogger implements the TelemetryCallback interface
// This allows you to log, monitor, or send telemetry data to external systems
type CustomTelemetryLogger struct {
	RequestCount  int
	ResponseCount int
	ErrorCount    int
	TotalTokens   int
	StartTime     time.Time
}

func NewCustomTelemetryLogger() *CustomTelemetryLogger {
	return &CustomTelemetryLogger{
		StartTime: time.Now(),
	}
}

func (logger *CustomTelemetryLogger) OnRequestSent(metadata base.RequestMetadata, requestJSON string) {
	logger.RequestCount++
	display.Info(fmt.Sprintf("📤 Request #%d sent", logger.RequestCount))
	display.KeyValue("  Model", metadata.Model)
	display.KeyValue("  Context Length", fmt.Sprintf("%d bytes", metadata.ContextLength))
	display.KeyValue("  Temperature", fmt.Sprintf("%.2f", metadata.Temperature))
	display.KeyValue("  Time", metadata.Timestamp.Format("15:04:05"))
	display.NewLine()
}

func (logger *CustomTelemetryLogger) OnResponseReceived(metadata base.ResponseMetadata, responseJSON string) {
	logger.ResponseCount++
	logger.TotalTokens += metadata.TotalTokens

	display.Success(fmt.Sprintf("📥 Response #%d received", logger.ResponseCount))
	display.KeyValue("  Response ID", metadata.ID)
	display.KeyValue("  Finish Reason", metadata.FinishReason)
	display.KeyValue("  Prompt Tokens", fmt.Sprintf("%d", metadata.PromptTokens))
	display.KeyValue("  Completion Tokens", fmt.Sprintf("%d", metadata.CompletionTokens))
	display.KeyValue("  Total Tokens", fmt.Sprintf("%d", metadata.TotalTokens))
	display.KeyValue("  Response Time", fmt.Sprintf("%d ms", metadata.ResponseTime))
	display.KeyValue("  Cumulative Tokens", fmt.Sprintf("%d", logger.TotalTokens))
	display.NewLine()
}

func (logger *CustomTelemetryLogger) OnStreamChunk(chunk string, index int) {
	// For streaming: log chunk information
	display.Info(fmt.Sprintf("📦 Stream chunk #%d: %d bytes", index, len(chunk)))
}

func (logger *CustomTelemetryLogger) OnError(err error, context string) {
	logger.ErrorCount++
	display.Error(fmt.Sprintf("❌ Error #%d in %s: %v", logger.ErrorCount, context, err))
	display.NewLine()
}

func (logger *CustomTelemetryLogger) PrintSummary() {
	duration := time.Since(logger.StartTime)
	display.Separator()
	display.Title("📊 Telemetry Summary")
	display.Separator()
	display.KeyValue("Session Duration", duration.String())
	display.KeyValue("Total Requests", fmt.Sprintf("%d", logger.RequestCount))
	display.KeyValue("Total Responses", fmt.Sprintf("%d", logger.ResponseCount))
	display.KeyValue("Total Errors", fmt.Sprintf("%d", logger.ErrorCount))
	display.KeyValue("Total Tokens Used", fmt.Sprintf("%d", logger.TotalTokens))
	if logger.ResponseCount > 0 {
		avgTokens := float64(logger.TotalTokens) / float64(logger.ResponseCount)
		display.KeyValue("Average Tokens/Response", fmt.Sprintf("%.1f", avgTokens))
	}
}

func main() {
	ctx := context.Background()

	display.Title("Chat Agent with Telemetry Callback")
	display.Separator()

	// Create custom telemetry logger
	telemetryLogger := NewCustomTelemetryLogger()

	// Create a chat agent
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.8).
			WithMaxTokens(2000),
	)
	if err != nil {
		panic(err)
	}

	// Register the telemetry callback
	// This will trigger the callback methods for every request/response
	agent.SetTelemetryCallback(telemetryLogger)

	display.Info("Telemetry callback registered!")
	display.Info("All requests and responses will be logged automatically.")
	display.Separator()
	display.NewLine()

	// First question
	display.Title("Question 1: Who is James T Kirk?")
	display.Separator()
	result, err := agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Who is James T Kirk?"},
	})
	if err != nil {
		panic(err)
	}
	display.KeyValue("Answer", result.Response)
	display.NewLine()

	// Second question
	display.Title("Question 2: What is his ship called?")
	display.Separator()
	result, err = agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "What is his ship called?"},
	})
	if err != nil {
		panic(err)
	}
	display.KeyValue("Answer", result.Response)
	display.NewLine()

	// Third question
	display.Title("Question 3: Who is his best friend?")
	display.Separator()
	result, err = agent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: "Who is his best friend?"},
	})
	if err != nil {
		panic(err)
	}
	display.KeyValue("Answer", result.Response)
	display.NewLine()

	// Print telemetry summary
	telemetryLogger.PrintSummary()

	display.NewLine()
	display.Success("Telemetry callback example completed!")
}
