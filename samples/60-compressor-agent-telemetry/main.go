package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// Create compressor agent
	agent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a text compression assistant.",
		},
		models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.3).
			WithMaxTokens(3000),
	)
	if err != nil {
		panic(err)
	}

	display.Title("Compressor Agent with Telemetry")
	display.Separator()

	// Long conversation to compress
	longConversation := []messages.Message{
		{Role: roles.User, Content: "What is the capital of France?"},
		{Role: roles.Assistant, Content: "The capital of France is Paris. Paris is not only the capital but also the largest city in France. It's known for the Eiffel Tower, the Louvre Museum, and many other famous landmarks."},
		{Role: roles.User, Content: "What about Germany?"},
		{Role: roles.Assistant, Content: "The capital of Germany is Berlin. Berlin is the largest city in Germany and has a rich history. It was divided during the Cold War and reunified in 1990."},
		{Role: roles.User, Content: "And Italy?"},
		{Role: roles.Assistant, Content: "The capital of Italy is Rome. Rome is one of the oldest continuously occupied cities in Europe and is known as the 'Eternal City'. It contains many historical sites including the Colosseum and the Vatican."},
	}

	// Calculate original size
	originalSize := 0
	for _, msg := range longConversation {
		originalSize += len(msg.Content)
	}

	display.Info("Original Conversation:")
	display.KeyValue("  Messages", fmt.Sprintf("%d", len(longConversation)))
	display.KeyValue("  Total Characters", fmt.Sprintf("%d", originalSize))
	display.NewLine()

	// Compress the context
	display.Info("Compressing conversation...")
	result, err := agent.CompressContext(longConversation)
	if err != nil {
		panic(err)
	}

	compressedSize := len(result.CompressedText)
	compressionRatio := float64(originalSize-compressedSize) / float64(originalSize) * 100

	display.Success("Compression Complete!")
	display.KeyValue("  Compressed Size", fmt.Sprintf("%d characters", compressedSize))
	display.KeyValue("  Compression Ratio", fmt.Sprintf("%.1f%%", compressionRatio))
	display.KeyValue("  Finish Reason", result.FinishReason)
	display.NewLine()

	display.Info("Compressed Text:")
	// Show first 200 chars
	preview := result.CompressedText
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	fmt.Println(strings.ReplaceAll(preview, "\n", " "))

	// Display telemetry
	display.NewLine()
	display.Separator()
	display.Title("📊 Compression Telemetry")
	display.Separator()

	// Request metadata
	reqMeta := agent.GetLastRequestMetadata()
	display.Info("Compression Request:")
	display.KeyValue("  Model", reqMeta.Model)
	display.KeyValue("  Input Context Length", fmt.Sprintf("%d bytes", reqMeta.ContextLength))
	display.KeyValue("  Temperature", fmt.Sprintf("%.2f", reqMeta.Temperature))
	display.KeyValue("  Max Tokens", fmt.Sprintf("%d", reqMeta.MaxTokens))

	// Response metadata
	display.NewLine()
	respMeta := agent.GetLastResponseMetadata()
	display.Info("Compression Response:")
	display.KeyValue("  Response ID", respMeta.ID)
	display.KeyValue("  Input Tokens", fmt.Sprintf("%d", respMeta.PromptTokens))
	display.KeyValue("  Output Tokens", fmt.Sprintf("%d", respMeta.CompletionTokens))
	display.KeyValue("  Total Tokens", fmt.Sprintf("%d", respMeta.TotalTokens))
	display.KeyValue("  Processing Time", fmt.Sprintf("%d ms", respMeta.ResponseTime))

	// Compression efficiency metrics
	display.NewLine()
	display.Info("Efficiency Metrics:")
	display.KeyValue("  Characters Saved", fmt.Sprintf("%d", originalSize-compressedSize))
	display.KeyValue("  Tokens Used", fmt.Sprintf("%d", respMeta.TotalTokens))
	display.KeyValue("  Characters per Token", fmt.Sprintf("%.2f", float64(originalSize)/float64(respMeta.TotalTokens)))

	// Token usage tracking
	display.NewLine()
	display.Info("Session Statistics:")
	display.KeyValue("  Total Tokens Used", fmt.Sprintf("%d", agent.GetTotalTokensUsed()))

	// Export full request JSON
	display.NewLine()
	display.Separator()
	display.Title("📄 Full Request JSON")
	display.Separator()
	reqJSON, _ := agent.GetLastRequestJSON()
	fmt.Println(reqJSON)

	// Export full response JSON
	display.NewLine()
	display.Separator()
	display.Title("📄 Full Response JSON")
	display.Separator()
	respJSON, _ := agent.GetLastResponseJSON()
	fmt.Println(respJSON)

	display.NewLine()
	display.Success("Compressor agent telemetry example completed!")
}
