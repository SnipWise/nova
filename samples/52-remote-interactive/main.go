package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/remote"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

var pendingOperations []string

func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// Create a remote agent that connects to the server
	agent, err := remote.NewAgent(
		ctx,
		agents.Config{
			Name: "Interactive Remote Bob Client",
		},
		"http://localhost:8080",
	)
	if err != nil {
		panic(err)
	}

	display.Colorf(display.ColorCyan, "🌐 Connected to remote agent at %s\n", "http://localhost:8080")
	display.Colorf(display.ColorCyan, "Agent: %s\n", agent.GetName())
	display.Colorf(display.ColorCyan, "Model: %s\n\n", agent.GetModelID())

	display.Colorf(display.ColorYellow, "=== Interactive Remote Agent Demo ===\n")
	display.Colorf(display.ColorMagenta, "This example demonstrates operation management with the remote agent.\n\n")

	// Example question that triggers tool calls
	display.Colorf(display.ColorYellow, "Sending question that triggers multiple tool calls...\n\n")

	question := []messages.Message{
		{
			Content: `
			Make the sum of 40 and 2,
			then say hello to Bob and to Sam,
			make the sum of 5 and 37,
			Say hello to Alice
			`,
			Role: roles.User,
		},
	}

	// Start streaming in a goroutine
	responseChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		_, err := agent.GenerateStreamCompletion(question, func(chunk string, finishReason string) error {
			if chunk != "" {
				fmt.Print(chunk)
			}
			if finishReason == "stop" {
				fmt.Println()
			}
			return nil
		})
		if err != nil {
			errorChan <- err
		} else {
			responseChan <- "done"
		}
	}()

	// Interactive loop for handling operations
	display.Colorf(display.ColorYellow, "\n\n=== Operation Management ===\n")
	display.Colorf(display.ColorMagenta, "Commands:\n")
	display.Colorf(display.ColorMagenta, "  v <operation_id>  - Validate an operation\n")
	display.Colorf(display.ColorMagenta, "  c <operation_id>  - Cancel an operation\n")
	display.Colorf(display.ColorMagenta, "  va                - Validate all pending operations\n")
	display.Colorf(display.ColorMagenta, "  r                 - Reset (cancel) all pending operations\n")
	display.Colorf(display.ColorMagenta, "  q                 - Quit\n\n")

	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-responseChan:
			display.Colorf(display.ColorGreen, "\n✅ Response completed!\n")
			return
		case err := <-errorChan:
			display.Colorf(display.ColorRed, "\n❌ Error: %v\n", err)
			return
		default:
			fmt.Print("> ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input == "" {
				continue
			}

			parts := strings.Fields(input)
			if len(parts) == 0 {
				continue
			}

			command := parts[0]

			switch command {
			case "v":
				if len(parts) < 2 {
					display.Colorf(display.ColorRed, "Usage: v <operation_id>\n")
					continue
				}
				operationID := parts[1]
				if err := agent.ValidateOperation(operationID); err != nil {
					display.Colorf(display.ColorRed, "Error: %v\n", err)
				}

			case "c":
				if len(parts) < 2 {
					display.Colorf(display.ColorRed, "Usage: c <operation_id>\n")
					continue
				}
				operationID := parts[1]
				if err := agent.CancelOperation(operationID); err != nil {
					display.Colorf(display.ColorRed, "Error: %v\n", err)
				}

			case "va":
				display.Colorf(display.ColorYellow, "Validating all pending operations...\n")
				for _, opID := range pendingOperations {
					if err := agent.ValidateOperation(opID); err != nil {
						display.Colorf(display.ColorRed, "Error validating %s: %v\n", opID, err)
					}
				}
				pendingOperations = []string{}

			case "r":
				if err := agent.ResetOperations(); err != nil {
					display.Colorf(display.ColorRed, "Error: %v\n", err)
				}
				pendingOperations = []string{}

			case "q":
				display.Colorf(display.ColorYellow, "Quitting...\n")
				return

			default:
				display.Colorf(display.ColorRed, "Unknown command: %s\n", command)
			}
		}
	}
}
