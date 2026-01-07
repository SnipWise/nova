## Chat agent - stream completion with reasoning

```go
package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are a helpful AI assistant that thinks step by step.",
		},
		models.Config{
			Name: "hf.co/menlo/lucy-gguf:q4_k_m",
			Temperature:        models.Float64(0.7),
			TopP:               models.Float64(0.9),
			ReasoningEffort: models.String(models.ReasoningEffortMedium),
		},			
	)
	if err != nil {
		panic(err)
	}

	_, err = agent.GenerateStreamCompletionWithReasoning(
		[]messages.Message{
			{Role: roles.User, Content: "What is 15 * 24?"},
		},
		func(reasoningChunk string, finishReason string) error {
			fmt.Print(reasoningChunk)
			if finishReason != "" {
				fmt.Println()
				fmt.Println("Finish reason", finishReason)
			}
			return nil
		},
		func(responseChunk string, finishReason string) error {
			fmt.Print(responseChunk)
			if finishReason != "" {
				fmt.Println()
				fmt.Println("Finish reason", finishReason)
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
}
```