package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/agents/server"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func main() {
	// Enable logging
	os.Setenv("NOVA_LOG_LEVEL", "INFO")

	ctx := context.Background()

	// Create the server agent (can be used in both CLI and server modes)
	agent, err := server.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-cli-agent",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: models.Float64(0.4),
		},
		":3500", // Port (not used in CLI mode)
		executeFunction,
	)
	if err != nil {
		panic(err)
	}

	// Create and set the tools agent
	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-tools-agent",
			EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: "You are Bob, a helpful AI assistant.",
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true),
		},
		tools.WithTools(GetToolsIndex()),
	)
	if err != nil {
		panic(err)
	}

	// Set the tools agent
	agent.SetToolsAgent(toolsAgent)

	// Optional: Set custom confirmation prompt for tool execution
	// By default, it auto-confirms in CLI mode
	// agent.SetConfirmationPromptFunction(customConfirmationPrompt)

	// Create the RAG agent
	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:      "rag-agent",
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
		},
		models.Config{
			Name: "ai/mxbai-embed-large",
		},
	)
	if err != nil {
		panic(err)
	}

	// Add data to the RAG agent
	contents, err := files.GetContentFiles("./data", ".md")
	if err != nil {
		panic(err)
	}
	for idx, content := range contents {
		piecesOfDoc := chunks.SplitMarkdownBySections(content)

		for chunkIdx, piece := range piecesOfDoc {

			display.Colorf(display.ColorYellow, "generating vectors... (docs %d/%d) [chunks: %d/%d]\n", idx+1, len(contents), chunkIdx+1, len(piecesOfDoc))

			err := ragAgent.SaveEmbedding(piece)
			if err != nil {
				display.Errorf("failed to save embedding for document %d: %v\n", idx, err)

			}
		}
	}

	agent.SetRagAgent(ragAgent)

	fmt.Println("🤖 Server Agent in CLI Mode with StreamCompletion")
	fmt.Println("Type 'exit' to quit")
	fmt.Println("---")

	// Interactive loop
	for {
		input := prompt.NewWithColor("🧑 You: ")
		question, err := input.RunWithEdit()
		if err != nil {
			display.Errorf("Error reading input: %v", err)
			continue
		}

		if question == "exit" {
			display.Infof("👋 Goodbye!")
			break
		}

		if question == "" {
			continue
		}

		// Use StreamCompletion method (same as crew agent)
		fmt.Print("🤖 Bob: ")
		_, err = agent.StreamCompletion(question, func(chunk string, finishReason string) error {
			if chunk != "" {
				fmt.Print(chunk)
			}
			if finishReason == "stop" {
				fmt.Println()
			}
			return nil
		})

		if err != nil {
			display.Errorf("❌ Error: %v", err)
		}
	}
}

func GetToolsIndex() []*tools.Tool {
	calculateSumTool := tools.NewTool("calculate_sum").
		SetDescription("Calculate the sum of two numbers").
		AddParameter("a", "number", "The first number", true).
		AddParameter("b", "number", "The second number", true)

	sayHelloTool := tools.NewTool("say_hello").
		SetDescription("Say hello to the given name").
		AddParameter("name", "string", "The name to greet", true)

	getCurrentTimeTool := tools.NewTool("get_current_time").
		SetDescription("Get the current time")

	return []*tools.Tool{
		calculateSumTool,
		sayHelloTool,
		getCurrentTimeTool,
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	fmt.Printf("\n🔧 Executing: %s\n", functionName)

	switch functionName {
	case "say_hello":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for say_hello"}`, nil
		}
		hello := fmt.Sprintf("👋 Hello, %s! Nice to meet you!", args.Name)
		return fmt.Sprintf(`{"message": "%s"}`, hello), nil

	case "calculate_sum":
		var args struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for calculate_sum"}`, nil
		}
		sum := args.A + args.B
		return fmt.Sprintf(`{"result": %g}`, sum), nil

	case "get_current_time":
		return `{"time": "2025-01-01 12:00:00 UTC"}`, nil

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}

// Optional: Custom confirmation prompt (commented out by default)
/*
func customConfirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
	display.Colorf(display.ColorYellow, "⚠️  Tool call detected: %s\n", functionName)
	display.Infof("Arguments: %s", arguments)

	choice := prompt.HumanConfirmation(fmt.Sprintf("Execute %s?", functionName))
	return choice
}
*/
