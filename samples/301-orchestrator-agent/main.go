package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	agent, err := orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:      "orchestrator-agent",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
				You role is to identify the topic of a conversation.
				Possible topics: 
				- Coding,
				- Code,
				- Software,
				- Debugging,
				- Computing, 
				- Programming,
				- Development,
				- Technology,
				- Tools,
				- Read,
				- Write,
				- Command,
				- Run,
				- Bash
				Respond in JSON with the field 'topic_discussion'.			
			`,
		},
		models.Config{
			Name: "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
		},
		// BeforeCompletion hook: called before each intent identification
		orchestrator.BeforeCompletion(func(a *orchestrator.Agent) {
			display.Info(">> [BeforeCompletion] Messages count: " + conversion.IntToString(len(a.GetMessages())))
			display.Info(">> [BeforeCompletion] Agent: " + a.GetName() + " (" + a.GetModelID() + ")")
		}),
		// AfterCompletion hook: called after each intent identification
		orchestrator.AfterCompletion(func(a *orchestrator.Agent) {
			display.Info("<< [AfterCompletion] Messages count: " + conversion.IntToString(len(a.GetMessages())))
		}),
	)
	if err != nil {
		panic(err)
	}

	// === Test 1: IdentifyIntent ===
	display.NewLine()
	display.Separator()
	display.Title("IdentifyIntent with BeforeCompletion / AfterCompletion hooks")
	display.Separator()

	topic1, err := agent.IdentifyTopicFromText("I love playing football on weekends")
	if err != nil {
		panic(err)
	}
	display.KeyValue("Detected topic", topic1)

	// === Test 2: Another topic ===
	display.NewLine()
	display.Separator()
	display.Title("Another topic identification")
	display.Separator()

	topic2, err := agent.IdentifyTopicFromText("What are the best restaurants in Paris?")
	if err != nil {
		panic(err)
	}
	display.KeyValue("Detected topic", topic2)

	// === Test 3: Tech topic ===
	display.NewLine()
	display.Separator()
	display.Title("Tech topic identification")
	display.Separator()

	topic3, err := agent.IdentifyTopicFromText("How does quantum computing work?")
	if err != nil {
		panic(err)
	}
	display.KeyValue("Detected topic", topic3)

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("All intent identifications triggered the BeforeCompletion and AfterCompletion hooks.")
	display.Separator()
}
