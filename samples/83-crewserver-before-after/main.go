package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/crewserver"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

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

	// Create the crew server agent with lifecycle hooks
	crewServerAgent, err := crewserver.NewAgent(
		ctx,
		crewserver.WithAgentCrew(agentCrew, "generic"),
		crewserver.WithPort(3500),
		crewserver.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
			switch strings.ToLower(topic) {
			case "coding", "programming", "code", "software":
				return "coder"
			default:
				return "generic"
			}
		}),
		// BeforeCompletion hook: called before each HTTP completion request
		crewserver.BeforeCompletion(func(a *crewserver.CrewServerAgent) {
			callCount++
			display.Info(">> [BeforeCompletion] Agent: " + a.GetName() + " - Call #" + conversion.IntToString(callCount))
		}),
		// AfterCompletion hook: called after each HTTP completion request
		crewserver.AfterCompletion(func(a *crewserver.CrewServerAgent) {
			display.Info("<< [AfterCompletion] Agent: " + a.GetName() + " - Call #" + conversion.IntToString(callCount))
		}),
	)
	if err != nil {
		panic(err)
	}

	// Start the HTTP server
	fmt.Printf("Starting crew server agent on http://localhost%s\n", crewServerAgent.GetPort())
	display.Info("Hooks will be triggered on each POST /completion request")
	log.Fatal(crewServerAgent.StartServer())
}
