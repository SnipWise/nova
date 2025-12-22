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
	"github.com/snipwise/nova/nova-sdk/ui/spinner"
)

func main() {
	ctx := context.Background()

	loadingSpinner := spinner.NewWithColor("").SetSuffix("loading model...").SetFrames(spinner.FramesPulsingStar)
	loadingSpinner.SetSuffixColor(spinner.ColorPurple).SetFrameColor(spinner.ColorRed)

	loadingSpinner.Start()

	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:      "bob-agent",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
			You are B.O.B, an AI assistant created by SnipWise.
			You are friendly, concise, and very helpful.
			You are a Golang expert and love to help with Go programming questions.
			`,
		},
		models.NewConfig("hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m").
			WithTemperature(0.0),
	)
	if err != nil {
		panic(err)
	}

	// Create markdown chunk parser for streaming display
	markdownParser := display.NewMarkdownChunkParser()

	result, err := agent.GenerateStreamCompletion(
		[]messages.Message{
			{
				Role: roles.User,
				Content: `
			I need help writing a simple Go http server that responds with "Hello, World!".
			Can you provide a complete code example?
			Then Explain the Go code snippet in brief.
			`,
			},
		},
		func(chunk string, finishReason string) error {
			if loadingSpinner.IsRunning() && finishReason == "" {
				loadingSpinner.Success("Model loaded!")
				loadingSpinner.Stop()
			}
			// Use markdown chunk parser for colorized streaming output
			if chunk != "" {
				display.MarkdownChunk(markdownParser, chunk)
			}
			if finishReason == "stop" {
				markdownParser.Flush()
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
