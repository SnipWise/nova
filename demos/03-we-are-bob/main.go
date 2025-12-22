package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"

	"github.com/snipwise/nova/nova-sdk/ui/display"
	"github.com/snipwise/nova/nova-sdk/ui/prompt"
	"github.com/snipwise/nova/nova-sdk/ui/spinner"
)

func GetCoderAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	coderAgentModel := env.GetEnvOrDefault("CODER_AGENT_MODEL", "hf.co/quantfactory/deepseek-coder-7b-instruct-v1.5-gguf:q4_k_m")
	coderAgentModelTemperatureStr := env.GetEnvOrDefault("CODER_AGENT_MODEL_TEMPERATURE", "0.8")
	coderAgentModelTemperature := conversion.StringToFloat(coderAgentModelTemperatureStr)
	coderAgentSystemInstructions := env.GetEnvOrDefault("CODER_AGENT_SYSTEM_INSTRUCTIONS", "")

	coderAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "coder-agent",
			EngineURL:          engineURL,
			SystemInstructions: coderAgentSystemInstructions,
		},
		models.NewConfig(coderAgentModel).
			WithTemperature(coderAgentModelTemperature),
	)
	if err != nil {
		return nil, err
	}

	return coderAgent, nil
}

func GetThinkerAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	thinkerModel := env.GetEnvOrDefault("THINKER_MODEL", "hf.co/menlo/lucy-gguf:q4_k_m")
	thinkerModelTemperatureStr := env.GetEnvOrDefault("THINKER_MODEL_TEMPERATURE", "0.8")
	thinkerModelTemperature := conversion.StringToFloat(thinkerModelTemperatureStr)
	thinkerAgentSystemInstructions := env.GetEnvOrDefault("THINKER_AGENT_SYSTEM_INSTRUCTIONS", "")

	thinkerAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "thinker-agent",
			EngineURL:          engineURL,
			SystemInstructions: thinkerAgentSystemInstructions,
		},
		models.NewConfig(thinkerModel).
			WithTemperature(thinkerModelTemperature),
	)
	if err != nil {
		return nil, err
	}

	return thinkerAgent, nil
}

func GetTranslatorAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	translatorModel := env.GetEnvOrDefault("TRANSLATOR_MODEL", "huggingface.co/tensorblock/nvidia_riva-translate-4b-instruct-gguf:q4_k_m")
	translatorModelTemperatureStr := env.GetEnvOrDefault("TRANSLATOR_MODEL_TEMPERATURE", "0.8")
	translatorModelTemperature := conversion.StringToFloat(translatorModelTemperatureStr)
	translatorAgentSystemInstructions := env.GetEnvOrDefault("TRANSLATOR_AGENT_SYSTEM_INSTRUCTIONS", "")

	translatorAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "translator-agent",
			EngineURL:          engineURL,
			SystemInstructions: translatorAgentSystemInstructions,
		},
		models.NewConfig(translatorModel).
			WithTemperature(translatorModelTemperature),
	)
	if err != nil {
		return nil, err
	}

	return translatorAgent, nil
}

func GetGenericAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	genericModel := env.GetEnvOrDefault("GENERIC_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	genericModelTemperatureStr := env.GetEnvOrDefault("GENERIC_MODEL_TEMPERATURE", "0.8")
	genericModelTemperature := conversion.StringToFloat(genericModelTemperatureStr)
	genericAgentSystemInstructions := env.GetEnvOrDefault("GENERIC_MODEL_SYSTEM_INSTRUCTIONS", "")

	genericAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "generic-agent",
			EngineURL:          engineURL,
			SystemInstructions: genericAgentSystemInstructions,
		},
		models.NewConfig(genericModel).
			WithTemperature(genericModelTemperature),
	)
	if err != nil {
		return nil, err
	}

	return genericAgent, nil
}

func GetCompressorAgent(ctx context.Context, engineURL string) (*compressor.Agent, error) {

	compressorModel := env.GetEnvOrDefault("COMPRESSOR_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")

	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "compressor-agent",
			EngineURL:          engineURL,
			SystemInstructions: compressor.Instructions.Expert,
		},
		models.NewConfig(compressorModel).
			WithTemperature(0.0),
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
	)
	if err != nil {
		return nil, err
	}

	return compressorAgent, nil
}

type Intent struct {
	TopicDiscussion string `json:"topic_discussion"`
}

func GetOrchestratorAgent(ctx context.Context, engineURL string) (*structured.Agent[Intent], error) {

	orchestratorModel := env.GetEnvOrDefault("ORCHESTRATOR_MODEL", "hf.co/menlo/lucy-gguf:q4_k_m")
	orchestratorAgentSystemInstructions := env.GetEnvOrDefault("ORCHESTRATOR_AGENT_SYSTEM_INSTRUCTIONS", "")

	agent, err := structured.NewAgent[Intent](
		ctx,
		agents.Config{
			Name:               "orchestrator-agent",
			EngineURL:          engineURL,
			SystemInstructions: orchestratorAgentSystemInstructions,
		},
		models.NewConfig(orchestratorModel).
			WithTemperature(0.0),
	)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

func main() {

	ctx := context.Background()

	thinkingSpinner := spinner.NewWithColor("").SetSuffix("thinking...").SetFrames(spinner.FramesDots)
	thinkingSpinner.SetSuffixColor(spinner.ColorBrightYellow).SetFrameColor(spinner.ColorBrightYellow)

	compressingSpinner := spinner.NewWithColor("").SetSuffix("context compressing...").SetFrames(spinner.FramesDots)
	compressingSpinner.SetSuffixColor(spinner.ColorPurple).SetFrameColor(spinner.ColorRed)

	err := godotenv.Load()
	if err != nil {
		display.Warningf("No .env file found or error loading it: %v", err)
		return
	}
	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
	contextCompressingThreshold := env.GetEnvOrDefault("CONTEXT_COMPRESSING_THRESHOLD", "6000")
	contextCompressingThresholdInt := conversion.StringToInt(contextCompressingThreshold)
	displayReasoningSteps := env.GetEnvOrDefault("DISPLAY_REASONING_STEPS", "true") == "true"

	compressorAgent, err := GetCompressorAgent(ctx, engineURL)
	if err != nil {
		display.Errorf("failed to create compressor agent: %v", err)
		return
	}

	orchestratorAgent, err := GetOrchestratorAgent(ctx, engineURL)
	if err != nil {
		display.Errorf("failed to create orchestrator agent: %v", err)
		return
	}

	coderAgent, err := GetCoderAgent(ctx, engineURL)
	if err != nil {
		display.Errorf("failed to create coder agent: %v", err)
		return
	}
	thinkerAgent, err := GetThinkerAgent(ctx, engineURL)
	if err != nil {
		display.Errorf("failed to create thinker agent: %v", err)
		return
	}
	tranlatorAgent, err := GetTranslatorAgent(ctx, engineURL)
	if err != nil {
		display.Errorf("failed to create translator agent: %v", err)
		return
	}
	genericAgent, err := GetGenericAgent(ctx, engineURL)
	if err != nil {
		display.Errorf("failed to create generic agent: %v", err)
		return
	}

	teamAgents := map[string]*chat.Agent{
		"coder":      coderAgent,
		"thinker":    thinkerAgent,
		"translator": tranlatorAgent,
		"generic":    genericAgent,
	}

	selectedAgent := teamAgents["generic"]

	for {
		markdownParser := display.NewMarkdownChunkParser()

		input := prompt.NewWithColor("🤖 Ask me something? [" + selectedAgent.GetName() + "]")
		question, err := input.RunWithEdit()

		if err != nil {
			display.Errorf("failed to get input: %v", err)
			return
		}
		if strings.HasPrefix(question, "/bye") {
			display.Infof("👋 Goodbye!")
			break
		}

		// Determine the dialogue topic
		thinkingSpinner.Start()
		response, _, err := orchestratorAgent.GenerateStructuredData([]messages.Message{
			{
				Role:    roles.User,
				Content: question,
			},
		})
		if err != nil {
			thinkingSpinner.Error("Failed to detect the conversation topic!")
			response.TopicDiscussion = "default"
		} else {
			thinkingSpinner.Success("Detected topic: " + response.TopicDiscussion)
		}
		display.NewLine()
		// Select agent based on detected topic
		switch strings.ToLower(response.TopicDiscussion) {
		case "coding", "programming", "development", "code", "software", "debugging", "technology":
			selectedAgent = teamAgents["coder"]
		case "philosophy", "thinking", "ideas", "thoughts", "psychology", "relationships":
			selectedAgent = teamAgents["thinker"]
		case "translation", "translate":
			selectedAgent = teamAgents["translator"]
		default:
			selectedAgent = teamAgents["generic"]
		}
		display.KeyValue("Selected agent", selectedAgent.GetName())
		display.NewLine()
		display.Separator()

		var reasoningBuilder strings.Builder

		result, err := selectedAgent.GenerateStreamCompletionWithReasoning(
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
		display.KeyValue("Context size", fmt.Sprintf("%d characters", selectedAgent.GetContextSize()))
		display.Separator()

		if selectedAgent.GetContextSize() > contextCompressingThresholdInt {
			compressingSpinner.Start()
			// newContext
			newContext, err := compressorAgent.CompressContext(selectedAgent.GetMessages())

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

			selectedAgent.ResetMessages()
			selectedAgent.AddMessage(
				roles.System,
				newContext.CompressedText,
			)
			display.KeyValue("New context size", conversion.IntToString(selectedAgent.GetContextSize()))
			display.Separator()

		}

	}

}
