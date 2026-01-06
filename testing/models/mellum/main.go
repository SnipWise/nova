package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	// Create a simple agent without exposing OpenAI SDK types
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions:      "You are Bob, a helpful AI assistant.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "huggingface.co/jetbrains/mellum-4b-dpo-all-gguf:q8_0",
			Temperature: models.Float64(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	display.Info("Streaming response:")
	display.NewLine()

	text := `<filename>Utils.kt\npackage utils\n\nfun multiply(x: Int, y: Int): Int {\n    return x * y\n}\n\n<filename>Config.kt\npackage config\n\nobject Config {\n    const val DEBUG = true\n    const val MAX_VALUE = 100\n}\n\n<filename>Example.kt\n<fim_suffix>\nfun main() {\n    val result = calculateSum(5, 10)\n    println(result)\n}\n<fim_prefix>fun calculateSum(a: Int, b: Int): Int {\n<fim_middle>'
`

	// Chat with streaming - no OpenAI types exposed
	result, err := agent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: text},
		},
		func(chunk string, finishReason string) error {
			// Simple callback that receives strings only
			if chunk != "" {
				fmt.Print(chunk)
			}
			if finishReason == "stop" {
				fmt.Println()
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Separator()
	display.KeyValue("Finish reason", result.FinishReason)
	display.KeyValue("Context size", fmt.Sprintf("%d characters", agent.GetContextSize()))
	display.Separator()

}
