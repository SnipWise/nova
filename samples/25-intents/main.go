package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"

	"github.com/snipwise/nova/nova-sdk/agents/structured"
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
	agent, err := structured.NewAgent[[]Intent](
		ctx,
		agents.Config{
			Name:        "DungeonMaster",
			Description: "...",
			EngineURL:   "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
			You are helping the dungeon master of a D&D game.
			Detect if the user wants to speak to one of the following NPCs:
			Thrain (dwarf blacksmith),
			Liora (elven mage),
			Galdor (human rogue),
			Elara (halfling ranger),
			Shesepankh (tiefling warlock).

			When identifying NPCs:
			- if the user wants to speak to a dwarf blacksmith, they mean Thrain.
			- if the user wants to speak to an elven mage, they mean Liora.
			- if the user wants to speak to a human rogue, they mean Galdor.
			- if the user wants to speak to a halfling ranger, they mean Elara.
			- if the user wants to speak to a tiefling warlock, they mean Shesepankh.

			For each intent, respond with:
			action: speak (or other action if relevant, ex meet, talk, etc)
			character: <NPC name>
			known: <true or false>

			Set known to true if:
			- The user explicitly mentions the NPC by name (Thrain, Liora, Galdor, Elara, or Shesepankh), OR
			- The user mentions the NPC by their role/description (dwarf blacksmith, elven mage, human rogue, halfling ranger, tiefling warlock), OR
			- The user mentions a topic clearly associated with one of the known NPCs (e.g., "spells and magic" = elven mage = Liora)

			Set known to false if:
			- The user wants to speak to someone who is NOT in the list of known NPCs
			`,
		},
		models.NewConfig("hf.co/menlo/jan-nano-gguf:q4_k_m").
			WithTemperature(0.7).WithTopP(0.9),
	)
	if err != nil {
		panic(err)
	}

	intents, _, err := agent.GenerateStructuredData([]messages.Message{
		{
			Role: roles.User,
			Content: `
				I want to chat with Thrain and learn about his blacksmith skills.
				I want to meet a dwarf blacksmith.
				I want to speak about spells and magic.
				I want to speak to Bob Morane.
				I want to talk to Galdor about stealth missions.
			`,
		},
	})
	if err != nil {
		panic(err)
	}

	for _, intent := range *intents {
		display.NewLine()
		display.Title("Intent Detection")

		display.KeyValue("Action", intent.Action)
		display.KeyValue("Character", intent.Character)
		display.KeyValue("Known", conversion.BoolToString(intent.Known))
		display.Separator()

	}

}
