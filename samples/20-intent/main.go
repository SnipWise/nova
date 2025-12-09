package main

import (
	"context"

	"github.com/snipwise/nova/nova/agents"
	"github.com/snipwise/nova/nova/models"
	"github.com/snipwise/nova/nova/structured"
	"github.com/snipwise/nova/nova/toolbox/conversion"
	"github.com/snipwise/nova/nova/ui/display"
)

type Intent struct {
	//Action string `json:"intent"`
	Action    string `json:"action"`
	Character string `json:"name"`
	Known     bool   `json:"known"`
}

func main() {
	ctx := context.Background()
	agent, err := structured.NewAgent[Intent](
		ctx,
		agents.AgentConfig{
			Name:        "DungeonMaster",
			Description: "...",
			EngineURL:   "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
			You are helping the dungeon master of a D&D game.
			Detect if the user want to speak to one of the following NPCs: 
			Thrain (dwarf blacksmith), 
			Liora (elven mage), 
			Galdor (human rogue), 
			Elara (halfling ranger), 
			Shesepankh (tiefling warlock).

			If the user's message does not explicitly mention wanting to speak to one of these NPCs, respond with:
			action: speak
			character: <NPC name>
			known: false

			Otherwise, respond with:
			action: speak
			character: <NPC name> 
			Where <NPC name> is the name of the NPC the user wants to speak to: Thrain, Liora, Galdor, Elara, or Shesepankh.
			known: true			
			`,
		},
		models.NewConfig("hf.co/menlo/jan-nano-gguf:q4_k_m").
			WithTemperature(0.0),
	)
	if err != nil {
		panic(err)
	}

	messages := []string{
		"I want to chat with Thrain and learn about his blacksmith skills.",
		"I want to meet a dwarf blacksmith.",
		"I want to speak about spells and magic.",
		"I want to speak to Bob Morane.",
	}

	for _, message := range messages {
		response, finishReason, err := agent.GenerateStructuredData([]structured.Message{
			{
				Role:    "user",
				Content: message,
			},
		})
		if err != nil {
			panic(err)
		}

		display.NewLine()
		display.Title("Intant Detection")

		display.KeyValue("Action", response.Action)
		display.KeyValue("Character", response.Character)
		display.KeyValue("Known", conversion.BoolToString(response.Known))
		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", finishReason)
		display.Separator()

	}

}
