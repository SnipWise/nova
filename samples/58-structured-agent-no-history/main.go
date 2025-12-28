package main

import (
	"context"
	"fmt"
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

	// Create agent with KeepConversationHistory set to false
	agent, err := structured.NewAgent[Country](
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
			Your name is Bob.
			You are an assistant that answers questions about countries around the world.
			`,
			KeepConversationHistory: false, // Disable conversation history
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	// First request about Canada
	display.NewLine()
	display.Separator()
	display.Title("First Request: Tell me about Canada")
	display.Separator()

	response1, finishReason1, err := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: "Tell me about Canada."},
	})

	if err != nil {
		panic(err)
	}

	display.KeyValue("Name", response1.Name)
	display.KeyValue("Capital", response1.Capital)
	display.KeyValue("Population", conversion.IntToString(response1.Population))
	display.KeyValue("Languages", strings.Join(response1.Languages, ", "))
	display.KeyValue("Finish reason", finishReason1)

	// Check messages after first request
	messages1 := agent.GetMessages()
	display.NewLine()
	display.KeyValue("Messages count after first request", conversion.IntToString(len(messages1)))
	display.Info("Expected: 1 (only system message, no user/assistant messages)")

	// Second request about France
	display.NewLine()
	display.Separator()
	display.Title("Second Request: Tell me about France")
	display.Separator()

	response2, finishReason2, err := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: "Tell me about France."},
	})

	if err != nil {
		panic(err)
	}

	display.KeyValue("Name", response2.Name)
	display.KeyValue("Capital", response2.Capital)
	display.KeyValue("Population", conversion.IntToString(response2.Population))
	display.KeyValue("Languages", strings.Join(response2.Languages, ", "))
	display.KeyValue("Finish reason", finishReason2)

	// Check messages after second request
	messages2 := agent.GetMessages()
	display.NewLine()
	display.KeyValue("Messages count after second request", conversion.IntToString(len(messages2)))
	display.Info("Expected: 1 (still only system message)")

	// Display all messages to verify
	display.NewLine()
	display.Separator()
	display.Title("All Messages in History")
	display.Separator()
	for i, msg := range messages2 {
		contentPreview := msg.Content
		if len(contentPreview) > 50 {
			contentPreview = contentPreview[:50] + "..."
		}
		display.KeyValue("Message "+conversion.IntToString(i+1), string(msg.Role)+": "+contentPreview)
	}

	display.NewLine()
	display.Separator()
	display.Success("Test completed!")
	display.Info("With KeepConversationHistory=false, only the system message should be kept.")
	display.Info("User and assistant messages should NOT be added to history.")
	display.Separator()

	fmt.Println(agent.ExportMessagesToJSON())
}
