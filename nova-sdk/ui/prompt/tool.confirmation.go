package prompt

import (
	"log"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
)

func HumanConfirmation(text string) tools.ConfirmationResponse {
	choices := []Choice{
		{Label: "yes", Value: "y"},
		{Label: "no", Value: "n"},
		{Label: "quit", Value: "q"},
	}

	selectPrompt := NewColorSelectKey(text, choices).
		SetDefault("y").
		SetColors(
			ColorBrightCyan,   // message color
			ColorWhite,        // choice color
			ColorBrightYellow, // default color
			ColorGray,         // key color
			ColorRed,          // error color
		).
		SetSymbols("❯", "●", "✗")

	selected, err := selectPrompt.Run()
	if err != nil {
		log.Fatal(err)
	}

	switch selected {
	case "q":
		return tools.Quit
	case "n":
		return tools.Denied
	case "y":
		return tools.Confirmed
	default:
		return tools.Denied
	}
}
