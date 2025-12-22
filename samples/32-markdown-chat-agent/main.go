package main

import (
	"context"

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

	generatingSpinner := spinner.NewWithColor("").SetSuffix("generating...").SetFrames(spinner.FramesDots)
	generatingSpinner.SetSuffixColor(spinner.ColorPurple).SetFrameColor(spinner.ColorRed)

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

	generatingSpinner.Start()

	result, err := agent.GenerateCompletion([]messages.Message{
		{
			Role: roles.User,
			Content: `
			I need help writing a simple Go http server that responds with "Hello, World!".
			Can you provide a complete code example?
			Then Explain the Go code snippet in brief.

			`,
		},
	})
	if err != nil {
		generatingSpinner.Error("Failed!")
		panic(err)
	}
	generatingSpinner.Success("Done!")

	display.Markdown(result.Response)

	display.KeyValue("Finish reason", result.FinishReason)

}
