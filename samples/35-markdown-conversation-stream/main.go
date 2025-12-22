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

	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:      "bob-agent",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
			You are B.O.B, an AI assistant created by SnipWise.
			You are friendly, concise, and very helpful.
			You are a Star Trek expert.
			Your answers should be in markdown format.
			`,
		},
		models.NewConfig("hf.co/menlo/lucy-gguf:q4_k_m").
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
				Role:    roles.User,
				Content: `Who is James T. Kirk? Provide a biography.`,
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

	markdownParser = display.NewMarkdownChunkParser()

	result, err = agent.GenerateStreamCompletion(
		[]messages.Message{
			{
				Role:    roles.User,
				Content: `Who is his best friend.`,
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

	markdownParser = display.NewMarkdownChunkParser()

	result, err = agent.GenerateStreamCompletion(
		[]messages.Message{
			{
				Role:    roles.User,
				Content: `What is the name of his ship.`,
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
