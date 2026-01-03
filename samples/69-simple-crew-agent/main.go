package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/crew"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
)

func getCoderAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	coderAgentSystemInstructionsContent := `
        You are an expert programming assistant. You write clean, efficient, and well-documented code. Always:
        - Provide complete, working code
        - Include error handling
        - Add helpful comments
        - Follow best practices for the language
        - Explain your approach briefly
	`
	coderAgentModel := "hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m"

	coderAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:                    "coder",
			EngineURL:               engineURL,
			SystemInstructions:      coderAgentSystemInstructionsContent,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        coderAgentModel,
			Temperature: models.Float64(0.8),
		},
	)
	if err != nil {
		return nil, err
	}

	return coderAgent, nil
}

var agentCrew = make(map[string]*chat.Agent)

func main() {
	// Enable logging
	if err := os.Setenv("NOVA_LOG_LEVEL", "INFO"); err != nil {
		panic(err)
	}

	engineURL := "http://localhost:12434/engines/llama.cpp/v1"

	ctx := context.Background()

	// ------------------------------------------------
	// Create the agent crew
	// ------------------------------------------------
	//agentCrew := make(map[string]*chat.Agent)

	coderAgent, err := getCoderAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}

	// Create the RAG agent
	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:      "rag-agent",
			EngineURL: engineURL,
		},
		models.Config{
			Name: "ai/mxbai-embed-large:latest",
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

	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "compressor-agent",
			EngineURL:          engineURL,
			SystemInstructions: compressor.Instructions.Minimalist,
		},
		models.Config{
			Name:        "ai/qwen2.5:0.5B-F16",
			Temperature: models.Float64(0.0),
		},
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
	)
	if err != nil {
		panic(err)
	}


	// Create the server agent
	crewAgent, err := crew.NewAgent(
		ctx,
		crew.WithSingleAgent(coderAgent),
		crew.WithCompressorAgentAndContextSize(compressorAgent, 8500),
		crew.WithRagAgent(ragAgent),
	)
	if err != nil {
		panic(err)
	}

	for {

		markdownParser := display.NewMarkdownChunkParser()

		input := prompt.NewWithColor("🤖 Ask me something? [" + crewAgent.GetName() + "]")
		question, err := input.RunWithEdit()

		if err != nil {
			display.Errorf("failed to get input: %v", err)
			return
		}
		if strings.HasPrefix(question, "/bye") {
			display.Infof("👋 Goodbye!")
			break
		}

		if strings.HasPrefix(question, "/messages") {
			display.Infof("💬 Current conversation messages:")
			for i, msg := range crewAgent.GetMessages() {
				display.Infof("Message %d - Role: %s, Content: \n%s", i, msg.Role, msg.Content)
				display.Separator()
			}
			continue
		}

		if strings.HasPrefix(question, "/reset") {
			display.Infof("🔄 Resetting %s context", crewAgent.GetName())
			crewAgent.ResetMessages()
			continue
		}

		display.NewLine()

		result, err := crewAgent.StreamCompletion(question, func(chunk string, finishReason string) error {

			// Use markdown chunk parser for colorized streaming output
			if chunk != "" {
				display.MarkdownChunk(markdownParser, chunk)
			}
			if finishReason == "stop" {
				markdownParser.Flush()
				markdownParser.Reset()
				//markdownParser.Flush()
				display.NewLine()
			}
			return nil
		})

		if err != nil {
			display.Errorf("[%s][%v]failed to get completion: %v", crewAgent.GetName(), crewAgent.GetContextSize(), err)
			return
		}

		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", result.FinishReason)
		display.KeyValue("Context size", fmt.Sprintf("%d characters", crewAgent.GetContextSize()))
		display.Separator()
	}

}
