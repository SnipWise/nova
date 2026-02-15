package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	var config orchestrator.AgentRoutingConfig
	if err := json.Unmarshal([]byte(`
		{
		"routing": [
			{
			"topics": ["coding", "programming", "development", "code", "software", "debugging", "technology", "computing"],
			"agent": "coder"
			},
			{
			"topics": ["health", "science", "mathematics", "philosophy", "food", "education", "travel", "sports"],
			"agent": "generic"
			},
			{
			"topics": ["read", "list", "write", "command"],
			"agent": "thinker"
			}
		],
		"default_agent": "generic"
		}
	`), &config); err != nil {
		panic(fmt.Errorf("failed to parse routing config: %w", err))
	}

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
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
		},
		orchestrator.WithRoutingConfig(config),
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
	display.KeyValue("Chosen agent for topic", agent.GetAgentForTopic(topic1))

	// === Test 2: Another topic ===
	display.NewLine()
	display.Separator()
	display.Title("Another topic identification")
	display.Separator()

	topic2, err := agent.IdentifyTopicFromText("I want to read a book about life philosophy")
	if err != nil {
		panic(err)
	}
	display.KeyValue("Detected topic", topic2)
	display.KeyValue("Chosen agent for topic", agent.GetAgentForTopic(topic2))

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
	display.KeyValue("Chosen agent for topic", agent.GetAgentForTopic(topic3))

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("All intent identifications triggered the BeforeCompletion and AfterCompletion hooks.")
	display.Separator()
}
