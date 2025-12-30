package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/crewserver"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
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
	coderAgentModel := "hf.co/quantfactory/deepseek-coder-7b-instruct-v1.5-gguf:q4_k_m"

	coderAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "coder",
			EngineURL:          engineURL,
			SystemInstructions: coderAgentSystemInstructionsContent,
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

func getThinkerAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	thinkerAgentSystemInstructionsContent := `
        You are a thoughtful conversational assistant. 
        - Listen carefully to the user
        - Think before responding
        - Ask clarifying questions when needed
        - Discuss topics with curiosity and respect
        - Admit when you don't know something
        Keep responses natural and conversational.	
	`
	thinkerModel := "hf.co/menlo/lucy-gguf:q4_k_m"

	thinkerAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "thinker",
			EngineURL:          engineURL,
			SystemInstructions: thinkerAgentSystemInstructionsContent,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        thinkerModel,
			Temperature: models.Float64(0.8),
		},
	)
	if err != nil {
		return nil, err
	}

	return thinkerAgent, nil
}

func getCookAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	cookAgentSystemInstructionsContent := `
		You are a culinary expert assistant. You provide:
		- Creative recipes
		- Cooking tips and techniques
		- Ingredient substitutions
		- Meal planning ideas
		Keep responses engaging and appetizing.
	`
	cookModel := "ai/qwen2.5:1.5B-F16"

	cookAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "cook",
			EngineURL:          engineURL,
			SystemInstructions: cookAgentSystemInstructionsContent,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        cookModel,
			Temperature: models.Float64(0.8),
		},
	)
	if err != nil {
		return nil, err
	}

	return cookAgent, nil
}

func getGenericAgent(ctx context.Context, engineURL string) (*chat.Agent, error) {

	genericAgentSystemInstructionsContent := `
        You respond appropriately to different types of questions.
        For factual questions: Give direct answers with key facts
        For how-to questions: Provide step-by-step guidance
        For opinion questions: Present balanced perspectives
        For complex topics: Break into digestible parts

        Always start with the most important information.	
	`
	genericModel := "hf.co/menlo/jan-nano-gguf:q4_k_m"

	genericAgent, err := chat.NewAgent(
		ctx,
		agents.Config{
			Name:               "generic",
			EngineURL:          engineURL,
			SystemInstructions: genericAgentSystemInstructionsContent,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        genericModel,
			Temperature: models.Float64(0.8),
		},
	)
	if err != nil {
		return nil, err
	}

	return genericAgent, nil
}

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
	agentCrew := make(map[string]*chat.Agent)

	coderAgent, err := getCoderAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}
	agentCrew["coder"] = coderAgent

	thinkerAgent, err := getThinkerAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}
	agentCrew["thinker"] = thinkerAgent

	cookAgent, err := getCookAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}
	agentCrew["cook"] = cookAgent

	genericAgent, err := getGenericAgent(ctx, engineURL)
	if err != nil {
		panic(err)
	}
	agentCrew["generic"] = genericAgent

	// ------------------------------------------------
	// Define the function to match agent ID to topic
	// ------------------------------------------------
	matchAgentFunction := func(currentAgentId, topic string) string {
		fmt.Println("🔵 Matching agent for topic:", topic)
		var agentId string
		switch strings.ToLower(topic) {
		case "coding", "programming", "development", "code", "software", "debugging", "technology", "software sevelopment":
			agentId = "coder"
		case "philosophy", "thinking", "ideas", "thoughts", "psychology", "relationships", "math", "mathematics", "science":
			agentId = "thinker"
		case "cooking", "recipe", "food", "culinary", "baking", "grilling", "meal":
			agentId = "cook"
		default:
			agentId = "generic"
		}
		fmt.Println("🟢 Matched agent ID:", agentId)
		return agentId
	}

	// Create the server agent
	crewServerAgent, err := crewserver.NewAgent(
		ctx,
		agentCrew,
		"generic",
		":8080",
		matchAgentFunction,
		executeFunction,
	)
	if err != nil {
		panic(err)
	}

	// Create the tools agent
	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			Name:      "tools-agent",
			EngineURL: engineURL,
			SystemInstructions: `
			You are an AI assistant that can call tools to help answer user queries effectively.
			- Always decide when to use a tool based on the user's request.
			- Choose the most appropriate tool for the task.
			- Provide clear and concise arguments to the tool.
			- After calling a tool, use the result to formulate your final response to the user.
			`,
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
	// Attach the tools agent to the server agent
	crewServerAgent.SetToolsAgent(toolsAgent)

	// Create the RAG agent
	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:      "rag-agent",
			EngineURL: engineURL,
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

	// Attach the RAG agent to the server agent
	crewServerAgent.SetRagAgent(ragAgent)

	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "compressor-agent",
			EngineURL:          engineURL,
			SystemInstructions: compressor.Instructions.Minimalist,
		},
		models.Config{
			//Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Name:        "ai/qwen2.5:0.5B-F16",
			Temperature: models.Float64(0.0),
		},
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
	)
	if err != nil {
		panic(err)
	}

	// Attach the compressor agent to the server agent
	crewServerAgent.SetCompressorAgent(compressorAgent)

	crewServerAgent.SetContextSizeLimit(8500)

	orchestratorAgentSystemInstructions := `
        You are good at identifying the topic of a conversation.
        Given a user's input, identify the main topic of discussion in only one word.
        The possible topics are: Technology, Health, Sports, Entertainment, Politics, Science, Mathematics,
        Travel, Food, Education, Finance, Environment, Fashion, History, Literature, Art,
        Music, Psychology, Relationships, Philosophy, Religion, Automotive, Gaming, Translation.
        Respond in JSON format with the field 'topic_discussion'.
	`

	orchestratorAgent, err := orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:               "orchestrator-agent",
			EngineURL:          engineURL,
			SystemInstructions: orchestratorAgentSystemInstructions,
		},
		models.Config{
			Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
			Temperature: models.Float64(0.0),
		},
	)
	if err != nil {
		panic(err)
	}

	// Attach the orchestrator agent to the server agent
	crewServerAgent.SetOrchestratorAgent(orchestratorAgent)

	// Display server start message

	display.Colorf(display.ColorCyan, "🚀 Server starting on http://localhost%s\n", crewServerAgent.GetPort())

	// Start the server
	if err := crewServerAgent.StartServer(); err != nil {
		panic(err)
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

	sayExit := tools.NewTool("say_exit").
		SetDescription("Say exit")

	return []*tools.Tool{
		calculateSumTool,
		sayHelloTool,
		sayExit,
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	display.Colorf(display.ColorGreen, "🟢 Executing function: %s with arguments: %s\n", functionName, arguments)

	switch functionName {
	case "say_hello":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for say_hello"}`, nil
		}
		hello := fmt.Sprintf("👋 Hello, %s!🙂", args.Name)
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

	case "say_exit":
		// NOTE: Returning a message and an error to stop further processing
		return fmt.Sprintf(`{"message": "%s"}`, "❌ EXIT"), errors.New("exit_loop")

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
