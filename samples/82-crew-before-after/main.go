package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/crew"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {
	ctx := context.Background()
	engineURL := "http://localhost:12434/engines/llama.cpp/v1"

	callCount := 0

	// Create two chat agents for the crew
	coderAgent, err := chat.NewAgent(ctx,
		agents.Config{
			Name:                    "coder",
			EngineURL:               engineURL,
			SystemInstructions:      "You are an expert programming assistant.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.4),
		},
	)
	if err != nil {
		panic(err)
	}

	genericAgent, err := chat.NewAgent(ctx,
		agents.Config{
			Name:                    "generic",
			EngineURL:               engineURL,
			SystemInstructions:      "You are a helpful AI assistant.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.4),
		},
	)
	if err != nil {
		panic(err)
	}

	agentCrew := map[string]*chat.Agent{
		"coder":   coderAgent,
		"generic": genericAgent,
	}

	// Create the crew agent with lifecycle hooks
	crewAgent, err := crew.NewAgent(
		ctx,
		crew.WithAgentCrew(agentCrew, "generic"),
		crew.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
			switch strings.ToLower(topic) {
			case "coding", "programming", "code", "software":
				return "coder"
			default:
				return "generic"
			}
		}),
		// BeforeCompletion hook: called before each StreamCompletion
		crew.BeforeCompletion(func(a *crew.CrewAgent) {
			callCount++
			display.Info(">> [BeforeCompletion] Agent: " + a.GetName() + " - Call #" + conversion.IntToString(callCount))
		}),
		// AfterCompletion hook: called after each StreamCompletion
		crew.AfterCompletion(func(a *crew.CrewAgent) {
			display.Info("<< [AfterCompletion] Agent: " + a.GetName() + " - Call #" + conversion.IntToString(callCount))
		}),
	)
	if err != nil {
		panic(err)
	}

	display.Success("Crew agent created with BeforeCompletion / AfterCompletion hooks")
	display.Info("Available agents: coder, generic")
	display.Separator()

	for {
		input := prompt.NewWithColor("Ask me something? [" + crewAgent.GetName() + "]")
		question, err := input.RunWithEdit()
		if err != nil {
			display.Errorf("failed to get input: %v", err)
			return
		}
		if strings.HasPrefix(question, "/bye") {
			display.Info("Goodbye!")
			break
		}

		display.NewLine()

		result, err := crewAgent.StreamCompletion(question, func(chunk string, finishReason string) error {
			if chunk != "" {
				fmt.Print(chunk)
			}
			return nil
		})
		if err != nil {
			display.Errorf("failed to get completion: %v", err)
			return
		}

		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", result.FinishReason)
		display.KeyValue("Total calls", conversion.IntToString(callCount))
		display.Separator()
	}
}
