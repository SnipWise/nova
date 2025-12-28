package main

import (
	"context"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

type Country struct {
	Name       string   `json:"name"`
	Capital    string   `json:"capital"`
	Population int      `json:"population"`
	Languages  []string `json:"languages"`
}

func main() {

	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	envFile := ".env"
	// Load environment variables from env file
	if err := godotenv.Load(envFile); err != nil {
		log.Error("Warning: Error loading env file: %v\n", err)
	}

	ctx := context.Background()
	agent, err := structured.NewBaseAgent[Country](
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
			SystemInstructions: `
			Your name is Bob. 
			You are an assistant that answers questions about countries around the world.
			`,
		},
		openai.ChatCompletionNewParams{
			Model:       "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature: openai.Opt(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	response, finishReason, err := agent.GenerateStructuredData([]openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Tell me about Canada."),
	})

	if err != nil {
		panic(err)
	}

	display.NewLine()
	display.Separator()
	display.Title("Response")
	display.KeyValue("Name", response.Name)
	display.KeyValue("Capital", response.Capital)
	display.KeyValue("Population", conversion.IntToString(response.Population))
	display.KeyValue("Languages", strings.Join(response.Languages, ", "))
	display.NewLine()
	display.Separator()
	display.KeyValue("Finish reason", finishReason)
}
