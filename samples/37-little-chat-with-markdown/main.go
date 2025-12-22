package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {
	ctx := context.Background()

	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:      "bob-agent",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
			You are B.O.B, an AI assistant created by SnipWise.
			`,
		},
		models.NewConfig("hf.co/menlo/jan-nano-gguf:q4_k_m").
			//models.NewConfig("ai/qwen2.5:1.5B-F16").
			WithTemperature(0.8),
	)
	if err != nil {
		panic(err)
	}

	// Create markdown chunk parser for streaming display

	for {

		markdownParser := display.NewMarkdownChunkParser()

		input := prompt.NewWithColor("🤖 Ask me something?").
			SetMessageColor(prompt.ColorBrightCyan).
			SetInputColor(prompt.ColorBrightWhite)

		question, err := input.Run()
		if err != nil {
			log.Fatal(err)
		}

		if strings.HasPrefix(question, "/bye") {
			fmt.Println("Goodbye!")
			break
		}

		result, err := agent.GenerateStreamCompletion(
			[]messages.Message{
				{
					Role:    roles.User,
					Content: question,
				},
			},
			func(chunk string, finishReason string) error {

				// Use markdown chunk parser for colorized streaming output
				if chunk != "" {
					display.MarkdownChunk(markdownParser, chunk)
				}
				if finishReason == "stop" {
					markdownParser.Flush()
					markdownParser.Reset()
					//markdownParser.Flush()
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

}

// create a golang hello world program and explain it
// [Brief] who is james t kirk
// tell me a story about jean-luc picard
// create an http server in golang and explain it
