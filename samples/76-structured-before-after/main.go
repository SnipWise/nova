package main

import (
	"context"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

type Country struct {
	Name       string   `json:"name"`
	Capital    string   `json:"capital"`
	Population int      `json:"population"`
	Languages  []string `json:"languages"`
}

func main() {
	ctx := context.Background()

	agent, err := structured.NewAgent[Country](
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `Your name is Bob.
You are an assistant that answers questions about countries around the world.`,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.0),
		},
		// BeforeCompletion hook: called before each structured data generation
		structured.BeforeCompletion[Country](func(a *structured.Agent[Country]) {
			display.Info(">> [BeforeCompletion] Messages count: " + conversion.IntToString(len(a.GetMessages())))
			display.Info(">> [BeforeCompletion] Agent: " + a.GetName() + " (" + a.GetModelID() + ")")
		}),
		// AfterCompletion hook: called after each structured data generation
		structured.AfterCompletion[Country](func(a *structured.Agent[Country]) {
			display.Info("<< [AfterCompletion] Messages count: " + conversion.IntToString(len(a.GetMessages())))
		}),
	)
	if err != nil {
		panic(err)
	}

	// === Test 1: Generate structured data about Canada ===
	display.NewLine()
	display.Separator()
	display.Title("Structured data generation with BeforeCompletion / AfterCompletion hooks")
	display.Separator()

	response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: "Tell me about Canada."},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Name", response.Name)
	display.KeyValue("Capital", response.Capital)
	display.KeyValue("Population", conversion.IntToString(response.Population))
	display.KeyValue("Languages", strings.Join(response.Languages, ", "))
	display.KeyValue("Finish reason", finishReason)

	// === Test 2: Another country ===
	display.NewLine()
	display.Separator()
	display.Title("Another structured data generation")
	display.Separator()

	response2, finishReason2, err := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: "Tell me about Japan."},
	})
	if err != nil {
		panic(err)
	}

	display.KeyValue("Name", response2.Name)
	display.KeyValue("Capital", response2.Capital)
	display.KeyValue("Population", conversion.IntToString(response2.Population))
	display.KeyValue("Languages", strings.Join(response2.Languages, ", "))
	display.KeyValue("Finish reason", finishReason2)

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("Both structured data generations triggered the BeforeCompletion and AfterCompletion hooks.")
	display.Separator()
}
