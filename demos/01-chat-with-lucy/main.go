package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"

	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
	"github.com/snipwise/nova/nova-sdk/ui/spinner"
)

func main() {

	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		display.Warningf("No .env file found or error loading it: %v", err)
	}
	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
	chatModel := env.GetEnvOrDefault("CHAT_MODEL", "hf.co/menlo/lucy-gguf:q4_k_m")
	compressorModel := env.GetEnvOrDefault("COMPRESSOR_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")

	chatModelTemperatureStr := env.GetEnvOrDefault("CHAT_MODEL_TEMPERATURE", "0.8")
	chatModelTemperature := conversion.StringToFloat(chatModelTemperatureStr)
	chatAgentSystemInstructions := env.GetEnvOrDefault("CHAT_AGENT_SYSTEM_INSTRUCTIONS", "")

	displayReasoningSteps := env.GetEnvOrDefault("DISPLAY_REASONING_STEPS", "true") == "true"

	contextCompressingThreshold := env.GetEnvOrDefault("CONTEXT_COMPRESSING_THRESHOLD", "6000")
	contextCompressingThresholdInt := conversion.StringToInt(contextCompressingThreshold)

	compressingSpinner := spinner.NewWithColor("").SetSuffix("context compressing...").SetFrames(spinner.FramesDots)
	compressingSpinner.SetSuffixColor(spinner.ColorPurple).SetFrameColor(spinner.ColorRed)

	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "Compressor",
			EngineURL:          engineURL,
			SystemInstructions: compressor.Instructions.Expert,
		},
		models.NewConfig(compressorModel).
			WithTemperature(0.0),
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
	)
	if err != nil {
		display.Errorf("failed to create agent: %v", err)
		return
	}

	chatAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "bob-agent",
			EngineURL:          engineURL,
			SystemInstructions: chatAgentSystemInstructions,
		},
		models.NewConfig(chatModel).
			WithTemperature(chatModelTemperature),
	)
	if err != nil {
		display.Errorf("failed to create agent: %v", err)
		return
	}

	for {
		markdownParser := display.NewMarkdownChunkParser()

		input := prompt.NewWithColor("🤖 Ask me something?")
		question, err := input.RunWithEdit()

		if err != nil {
			display.Errorf("failed to get input: %v", err)
			return
		}
		if strings.HasPrefix(question, "/bye") {
			display.Infof("👋 Goodbye!")
			break
		}

		var reasoningBuilder strings.Builder

		result, err := chatAgent.GenerateStreamCompletionWithReasoning(
			[]messages.Message{
				{
					Role:    roles.User,
					Content: question,
				},
			},
			func(reasoningChunk string, finishReason string) error {

				if displayReasoningSteps {
					reasoningBuilder.WriteString(reasoningChunk)
					display.Color(reasoningChunk, display.ColorYellow)
					if finishReason != "" {
						display.NewLine()
						display.KeyValue("Reasoning finish reason", finishReason)
						display.NewLine()
					}
				}
				return nil
			},
			func(chunk string, finishReason string) error {

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
			},
		)
		if err != nil {
			display.Errorf("failed to get completion: %v", err)
			return
		}
		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", result.FinishReason)
		display.KeyValue("Context size", fmt.Sprintf("%d characters", chatAgent.GetContextSize()))
		display.Separator()

		if chatAgent.GetContextSize() > contextCompressingThresholdInt {
			compressingSpinner.Start()
			// newContext
			newContext, err := compressorAgent.CompressContext(chatAgent.GetMessages())

			// newContext, err := compressorAgent.CompressContextStream(
			// 	chatAgent.GetMessages(),
			// 	func(partialResponse string, finishReason string) error {
			// 		display.Color(partialResponse, display.ColorCyan)
			// 		return nil
			// 	},
			// )
			if err != nil {
				compressingSpinner.Error("Failed!")
				display.Errorf("failed to compress context: %v", err)
				return
			}
			compressingSpinner.Success("Done!")

			chatAgent.ResetMessages()
			chatAgent.AddMessage(
				roles.System,
				newContext.CompressedText,
			)
			display.KeyValue("New context size", conversion.IntToString(chatAgent.GetContextSize()))
			display.Separator()

		}

	}

}
