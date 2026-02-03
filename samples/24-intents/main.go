package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"

	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/ui/display"
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
		agents.Config{
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
		models.Config{
			//Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Name:        "ai/qwen2.5:1.5B-F16",

			Temperature: models.Float64(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	messagesList := []string{
		"I want to chat with Thrain and learn about his blacksmith skills.",
		"I want to meet a dwarf blacksmith.",
		"I want to speak about spells and magic.",
		"I want to speak to Bob Morane.",
	}

	for _, message := range messagesList {
		response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
			{
				Role:    roles.User,
				Content: message,
			},
		})
		if err != nil {
			panic(err)
		}

		display.NewLine()
		display.Title("Intent Detection")

		display.KeyValue("Action", response.Action)
		display.KeyValue("Character", response.Character)
		display.KeyValue("Known", conversion.BoolToString(response.Known))
		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", finishReason)
		display.Separator()

	}

}
