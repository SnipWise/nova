package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/mcptools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"

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

func GetCookingAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	cookingModel := env.GetEnvOrDefault("COOKING_MODEL", "hf.co/menlo/lucy-gguf:q4_k_m")
	cookingModelTemperatureStr := env.GetEnvOrDefault("COOKING_MODEL_TEMPERATURE", "0.8")
	cookingModelTemperature := conversion.StringToFloat(cookingModelTemperatureStr)
	cookingAgentSystemInstructions := env.GetEnvOrDefault("COOKING_AGENT_SYSTEM_INSTRUCTIONS", "")

	cookingAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "cooking-agent",
			EngineURL:          engineURL,
			SystemInstructions: cookingAgentSystemInstructions,
		},
		models.NewConfig(cookingModel).
			WithTemperature(cookingModelTemperature),
	)
	if err != nil {
		return nil, err
	}

	return cookingAgent, nil
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

func GetToolsAgent(ctx context.Context, engineURL string, toolsCatalog []mcp.Tool) (*tools.Agent, error) {
	toolsAgentModel := env.GetEnvOrDefault("TOOLS_AGENT_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	toolsAgentSystemInstructions := env.GetEnvOrDefault("TOOLS_AGENT_SYSTEM_INSTRUCTIONS", "")

	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL:          engineURL,
			SystemInstructions: toolsAgentSystemInstructions,
		},

		models.NewConfig(toolsAgentModel).
			WithTemperature(0.0).
			WithParallelToolCalls(true),

		tools.WithMCPTools(toolsCatalog),
	)
	if err != nil {
		return nil, err
	}

	return toolsAgent, nil
}

type DuckDuckGoSearchInput struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results,omitempty"`
}

type DuckDuckGoFetchInput struct {
	Url string `json:"url"`
}

func GetRagAgent(ctx context.Context, engineURL string) (*rag.Agent, error) {

	generatingVectorsSpinner := spinner.NewWithColor("").SetSuffix("generating vectors...").SetFrames(spinner.FramesPulsingStar)
	generatingVectorsSpinner.SetSuffixColor(spinner.ColorBrightBlue).SetFrameColor(spinner.ColorBrightBlue)

	loadingVectorsSpinner := spinner.NewWithColor("").SetSuffix("loading vectors...").SetFrames(spinner.FramesPulsingStar)
	loadingVectorsSpinner.SetSuffixColor(spinner.ColorBrightBlue).SetFrameColor(spinner.ColorBrightBlue)

	ragModel := env.GetEnvOrDefault("RAG_MODEL", "ai/mxbai-embed-large")

	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:      "rag-agent",
			EngineURL: engineURL,
		},
		models.NewConfig(ragModel),
	)
	if err != nil {
		return nil, err
	}
	// Load or generate vector store
	ragStorePath := env.GetEnvOrDefault("RAG_STORE_PATH", "./store")
	ragStorePathFile := ragStorePath + "/" + ragAgent.GetName() + ".json"
	if ragAgent.StoreFileExists(ragStorePathFile) {
		loadingVectorsSpinner.Start()
		err := ragAgent.LoadStore(ragStorePathFile)
		if err != nil {
			loadingVectorsSpinner.Error("failed!")
			display.Errorf("failed to load vector store: %v\n", err)
			return nil, err
		}
		loadingVectorsSpinner.Success("Store loaded!")
	} else {

		// Read markdown files from data directory and generate embeddings
		ragDocumentsPath := env.GetEnvOrDefault("RAG_DOCUMENTS_PATH", "./data")
		generatingVectorsSpinner.Start()
		contents, err := files.GetContentFiles(ragDocumentsPath, ".md")
		if err != nil {
			return nil, err
		}
		for idx, content := range contents {
			piecesOfDoc := chunks.SplitMarkdownBySections(content)

			for chunkIdx, piece := range piecesOfDoc {

				progress := fmt.Sprintf("generating vectors... (docs %d/%d) [chunks: %d/%d]", idx+1, len(contents), chunkIdx+1, len(piecesOfDoc))
				generatingVectorsSpinner.SetSuffix(progress + "       ")

				err := ragAgent.SaveEmbedding(piece)
				if err != nil {
					generatingVectorsSpinner.Error("failed!")
					display.Errorf("failed to save embedding for document %d: %v\n", idx, err)
					return nil, err
				}
			}
		}

		err = ragAgent.PersistStore(ragStorePathFile)
		if err != nil {
			generatingVectorsSpinner.Error("failed!")
			display.Errorf("failed to persist vector store: %v\n", err)
			return nil, err
		}
		generatingVectorsSpinner.Success("Vectors generated and saved!")
	}

	return ragAgent, nil
}

func main() {

	ctx := context.Background()

	thinkingSpinner := spinner.NewWithColor("").SetSuffix("thinking...").SetFrames(spinner.FramesDots)
	thinkingSpinner.SetSuffixColor(spinner.ColorBrightYellow).SetFrameColor(spinner.ColorBrightYellow)

	compressingSpinner := spinner.NewWithColor("").SetSuffix("context compressing...").SetFrames(spinner.FramesDots)
	compressingSpinner.SetSuffixColor(spinner.ColorPurple).SetFrameColor(spinner.ColorRed)

	toolsSpinner := spinner.NewWithColor("").SetSuffix("tool detection...").SetFrames(spinner.FramesCircle)
	toolsSpinner.SetSuffixColor(spinner.ColorBrightGreen).SetFrameColor(spinner.ColorBrightGreen)

	ragSpinner := spinner.NewWithColor("").SetSuffix("retrieving relevant context...").SetFrames(spinner.FramesPulsingStar)
	ragSpinner.SetSuffixColor(spinner.ColorBrightMagenta).SetFrameColor(spinner.ColorBrightMagenta)

	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
	contextCompressingThreshold := env.GetEnvOrDefault("CONTEXT_COMPRESSING_THRESHOLD", "6000")
	contextCompressingThresholdInt := conversion.StringToInt(contextCompressingThreshold)
	displayReasoningSteps := env.GetEnvOrDefault("DISPLAY_REASONING_STEPS", "true") == "true"

	// toolDetectionEnabled := env.GetEnvOrDefault("TOOLS_DETECTION", "true") == "true"
	// if !toolDetectionEnabled {
	// 	display.Warningf("TOOLS_DETECTION is disabled. The agent will not be able to use tools.")
	// }

	// MCP:
	mcpGatewayURL := env.GetEnvOrDefault("MCP_GATEWAY_URL", "http://localhost:9011")

	mcpClient, err := mcptools.NewStreamableHttpMCPClient(ctx, mcpGatewayURL)
	if err != nil {
		display.Errorf("failed to create MCP client: %v", err)
		return
	}
	// Print available tools
	// for _, tool := range mcpClient.GetTools() {
	// 	display.KeyValue("Tool", tool.Name)
	// }

	ragAgent, err := GetRagAgent(ctx, engineURL)
	if err != nil {
		display.Errorf("failed to create RAG agent: %v", err)
		return
	}

	toolsAgent, err := GetToolsAgent(ctx, engineURL, mcpClient.GetTools())
	if err != nil {
		display.Errorf("failed to create tools agent: %v", err)
		return
	}

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

	cookingAgent, err := GetCookingAgent(ctx, engineURL)
	if err != nil {
		display.Errorf("failed to create cooking agent: %v", err)
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
		"cooking":    cookingAgent,
		"translator": tranlatorAgent,
		"generic":    genericAgent,
	}

	selectedAgent := teamAgents["generic"]

	for {
		markdownParser := display.NewMarkdownChunkParser()

		input := prompt.NewWithColor("🤖 Ask me something? [" + selectedAgent.GetName() + "]")
		question, err := input.RunWithEdit()

		// TODO: add a command to reduce context size manually
		// e.g., /compress

		if err != nil {
			display.Errorf("failed to get input: %v", err)
			return
		}
		if strings.HasPrefix(question, "/bye") {
			display.Infof("👋 Goodbye!")
			break
		}
		if strings.HasPrefix(question, "/reset") {
			display.Infof("🔄 Resetting %s context", selectedAgent.GetName())
			selectedAgent.ResetMessages()
			continue
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

		// --------------------------------------------------------
		// Select agent based on detected topic
		// --------------------------------------------------------

		switch strings.ToLower(response.TopicDiscussion) {
		case "coding", "programming", "development", "code", "software", "debugging", "technology":
			selectedAgent = teamAgents["coder"]
		case "philosophy", "thinking", "ideas", "thoughts", "psychology", "relationships", "math", "mathematics", "science":
			selectedAgent = teamAgents["thinker"]
		case "translation", "translate":
			selectedAgent = teamAgents["translator"]
		case "cooking", "recipe", "food", "culinary", "baking", "grilling", "meal":
			selectedAgent = teamAgents["cooking"]
		default:
			selectedAgent = teamAgents["generic"]
		}
		display.KeyValue("Selected agent", selectedAgent.GetName())
		display.NewLine()
		display.Separator()

		// --------------------------------------------------------
		// Retrieve relevant context from RAG agent
		// --------------------------------------------------------

		ragSpinner.Start()
		//similarities, err := ragAgent.SearchSimilar(question, 0.6)
		similarities, err := ragAgent.SearchTopN(question, 0.6, 3)
		if err != nil {
			ragSpinner.Error("Failed to retrieve relevant context!")
			display.Errorf("failed to search similar embeddings: %v", err)
			return
		} else {
			ragSpinner.Success("Relevant context retrieved.")
		}
		// --------------------------------------------------------
		// Display similarities
		// --------------------------------------------------------
		if len(similarities) == 0 {
			//display.Warningf("No relevant context found.")
			display.Color("No relevant context found.", display.ColorBrightRed)
		} else {
			display.Colorf(display.ColorGreen, "Similarities for query:\n")
			for _, sim := range similarities {

				displayResult := ""
				if len(sim.Prompt) >= 55 {
					displayResult = sim.Prompt[:55] + "..."
				}
				display.Colorf(display.ColorBrightBlue, "Content: %s\n", displayResult)
				display.Colorf(display.ColorYellow, "Score: %f\n", sim.Similarity)
			}
		}

		display.NewLine()
		display.Separator()
		relevantContext := ""
		for _, sim := range similarities {
			relevantContext += sim.Prompt + "\n---\n"
		}
		ragSpinner.Success("Done!")

		if len(relevantContext) > 0 {
			selectedAgent.AddMessage(
				roles.System,
				"Relevant information to help you answer the question:\n"+relevantContext,
			)
		}

		// --------------------------------------------------------
		// Detect tool calls
		// --------------------------------------------------------
		toolsSpinner.Start()

		resultOfToolCalls, err := toolsAgent.DetectParallelToolCalls(
			[]messages.Message{
				{
					Role:    roles.User,
					Content: question,
				},
			},
			// Tool execution function
			func(functionName, arguments string) (string, error) {
				display.NewLine()
				display.KeyValue("Calling tool", functionName)
				display.KeyValue("With arguments", arguments)
				display.NewLine()

				result, err := mcpClient.ExecToolWithString(functionName, arguments)
				if err != nil {
					return "", err
				}
				displayResult := ""
				if len(result) >= 150 {
					displayResult = result[:150] + "..."
				}
				display.NewLine()
				display.Colorln(displayResult, display.ColorBrightCyan)

				return result, err
				// switch functionName {
				// case "fetch_content":

				// 	param, err := conversion.FromJSON[DuckDuckGoFetchInput](arguments)
				// 	if err != nil {
				// 		return "", err
				// 	}
				// 	result, err := mcpClient.ExecToolWithAny(
				// 		"fetch_content",
				// 		param,
				// 	)

				// 	if err != nil {
				// 		return "", err
				// 	}
				// 	displayResult := ""
				// 	if len(result) >= 80 {
				// 		displayResult = result[:80] + "..."
				// 	}
				// 	display.NewLine()
				// 	display.Colorln(displayResult, display.ColorBrightPurple)

				// 	return result, nil

				// case "search":
				// 	param, err := conversion.FromJSON[DuckDuckGoSearchInput](arguments)
				// 	if err != nil {
				// 		return "", err
				// 	}
				// 	param.MaxResults = 3

				// 	result, err := mcpClient.ExecToolWithAny(
				// 		"search",
				// 		param,
				// 	)

				// 	if err != nil {
				// 		return "", err
				// 	}

				// 	displayResult := ""
				// 	if len(result) >= 80 {
				// 		displayResult = result[:80] + "..."
				// 	}
				// 	display.NewLine()
				// 	display.Colorln(displayResult, display.ColorBrightCyan)
				// 	return result, nil

				// default:
				// 	return "", fmt.Errorf("unknown function: %s", functionName)
				// }

			},
		)
		if err != nil {
			toolsSpinner.Error("Failed to detect tool calls!")
			display.Errorf("failed to detect tool calls: %v", err)
			return
		}
		if len(resultOfToolCalls.Results) == 0 {
			toolsSpinner.Success("No tool calls detected.")
		} else {
			toolsSpinner.Success("Tool calls detected and executed.")
		}
		//toolsSpinner.Success("Done!")
		display.KeyValue("Finish Reason", resultOfToolCalls.FinishReason)

		/* IMPORTANT: Avoid context size error
		✗ failed to detect tool calls: POST "http://localhost:12434/engines/llama.cpp/v1/chat/completions":
		400 Bad Request {"code":400,"message":"the request exceeds the available context size, try increasing it",
		"type":"exceed_context_size_error","n_prompt_tokens":4153,"n_ctx":4096}
		*/
		toolsAgent.ResetMessages()

		//display.NewLine()
		display.Separator()

		// for _, value := range resultOfToolCalls.Results {
		// 	display.KeyValue("Result for tool", value)
		// }
		// display.NewLine()
		// display.Separator()

		toolAssistantMessage := resultOfToolCalls.LastAssistantMessage

		var reasoningBuilder strings.Builder

		// Prepare messages for selected agent
		var messagesList []messages.Message
		if len(toolAssistantMessage) > 0 {
			messagesList = []messages.Message{
				{
					Role:    roles.System,
					Content: toolAssistantMessage,
				},
				{
					Role:    roles.User,
					Content: question,
				},
			}
		} else {
			messagesList = []messages.Message{
				{
					Role:    roles.User,
					Content: question,
				},
			}
		}

		// BEGIN:
		// --------------------------------------------------------
		// Context compressing if needed
		// --------------------------------------------------------
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
		// END:

		// --------------------------------------------------------
		// Generate response with reasoning and streaming
		// --------------------------------------------------------
		result, err := selectedAgent.GenerateStreamCompletionWithReasoning(
			messagesList,
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
			display.Errorf("[%s][%v]failed to get completion: %v", selectedAgent.GetName(), selectedAgent.GetContextSize(), err)
			return
		}
		display.NewLine()
		display.Separator()
		display.KeyValue("Finish reason", result.FinishReason)
		display.KeyValue("Context size", fmt.Sprintf("%d characters", selectedAgent.GetContextSize()))
		display.Separator()

		// --------------------------------------------------------
		// Context compressing if needed
		// --------------------------------------------------------
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
