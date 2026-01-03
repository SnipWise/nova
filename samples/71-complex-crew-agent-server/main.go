package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/joho/godotenv"

	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/crewserver"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

var agentCrew map[string]*chat.Agent

// var crewServerAgent *crewserver.CrewServerAgent

func main() {

	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	envFile := ".env"
	// Load environment variables from env file (optional in Docker)
	if err := godotenv.Load(envFile); err != nil {
		log.Info("Note: .env file not found (using Docker environment variables)\n")
	}

	ctx := context.Background()

	storePath := env.GetEnvOrDefault("STORE_PATH", "./store")
	dataPath := env.GetEnvOrDefault("DATA_PATH", "./data")

	httpPort := env.GetEnvIntOrDefault("HTTP_PORT", 3500)

	engineURL := env.GetEnvOrDefault("ENGINE_URL", "http://localhost:12434/engines/llama.cpp/v1")
	ragModelId := env.GetEnvOrDefault("RAG_EMBEDDING_MODEL_ID", "ai/embeddinggemma:latest")
	metadataModelId := env.GetEnvOrDefault("METADATA_MODEL_ID", "hf.co/menlo/jan-nano-gguf:q4_k_m")

	ragAgent, err := CreateRagAgent(ctx, engineURL, ragModelId)
	if err != nil {
		return
	}

	metadataExtractorAgent, err := CreateMetadataExtractorAgent(ctx, engineURL, metadataModelId)
	if err != nil {
		return
	}

	err = LoadSnippetData(dataPath+"/snippets", storePath, ragAgent, metadataExtractorAgent)
	if err != nil {
		return
	}

	compressorAgent, err := CreateCompressorAgent(ctx, engineURL)
	if err != nil {
		return
	}

	coderAgent, err := CreateCoderAgent(ctx, engineURL)
	if err != nil {
		return
	}

	thinkerAgent, err := CreateThinkerAgent(ctx, engineURL)
	if err != nil {
		return
	}

	expertAgent, err := CreateExpertAgent(ctx, engineURL)
	if err != nil {
		return
	}

	// ------------------------------------------------
	// Create the agent crew
	// ------------------------------------------------
	agentCrew = map[string]*chat.Agent{
		"expert":  expertAgent,
		"thinker": thinkerAgent,
		"coder":   coderAgent,
	}

	orchestratorAgent, err := CreateOrchestratorAgent(ctx, engineURL)
	if err != nil {
		return
	}

	// ------------------------------------------------
	// Create the tools agent
	// ------------------------------------------------
	toolsAgent, err := CreateToolsAgent(ctx, engineURL)
	if err != nil {
		return
	}

	// ------------------------------------------------
	// Define the function to match agent ID to topic
	// ------------------------------------------------
	// TODO: this should be specified in an external configuration
	// old_matchAgentFunction := func(topic string) string {
	// 	display.Infof("🔵 Matching agent for topic: %s", topic)
	// 	var agentId string
	// 	switch strings.ToLower(topic) {
	// 	case "generate", "generating":
	// 		agentId = "coder"
	// 	case "think", "thinking", "ideas", "math", "mathematics", "science":
	// 		agentId = "thinker"
	// 	default:
	// 		agentId = "generic"
	// 	}
	// 	display.Infof("🟢 Matched agent ID: %s", agentId)
	// 	return agentId
	// }

	matchAgentFunction := func(currentAgentId, topic string) string {
		display.Infof("[current: %s] 🔵 Matching agent for topic: %s", currentAgentId, topic)
		var agentId string
		switch strings.ToLower(topic) {
		case "code_generation", "code generation", "generation", "write code", "create code":
			agentId = "coder"
		case "complex_thinking", "complex thinking", "thinking", "reasoning", "analysis", "design", "architecture":
			agentId = "thinker"
		case "code_question", "code question", "question", "how to", "explain":
			agentId = "expert"
		default:
			agentId = "expert"
		}
		display.Infof("🟢 Matched agent ID: %s", agentId)

		// Transfer conversation history if switching agents
		if agentId != currentAgentId {
			display.Infof("🔄 Switching agent from %s to %s", currentAgentId, agentId)
			history := agentCrew[currentAgentId].GetMessages()
			agentCrew[agentId].AddMessages(history)
			// agentCrew[currentAgentId].ResetMessages()
		} else {
			display.Infof("➡️ Continuing with the same agent: %s", agentId)
		}

		return agentId
	}

	// executeFunction := func(functionName string, arguments string) (string, error) {
	// 	display.Info("🔵 Executing function:" + functionName + "with arguments:" + arguments)
	// 	// For demonstration, we just echo back the function name and arguments
	// 	result := fmt.Sprintf("Executed function: %s with arguments: %s", functionName, arguments)
	// 	display.Info("🟢 Function execution result:" + result)
	// 	return result, nil
	// }

	// ------------------------------------------------
	// Create the server agent
	// ------------------------------------------------
	crewServerAgent, err := crewserver.NewAgent(
		ctx,
		crewserver.WithAgentCrew(agentCrew,"expert"),
		crewserver.WithPort(httpPort),
		crewserver.WithMatchAgentIdToTopicFn(matchAgentFunction),
		crewserver.WithExecuteFn(executeFunction),
		crewserver.WithToolsAgent(toolsAgent),
		crewserver.WithOrchestratorAgent(orchestratorAgent),
		crewserver.WithCompressorAgentAndContextSize(compressorAgent, 32000),
		crewserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 7),

	)
	if err != nil {
		display.Errorf("❌ Error creating crew server agent: %v", err)
		return
	}
	display.Infof("🚀 Crew server agent created and listening on port %s", crewServerAgent.GetPort())

	// Start the server agent
	if err := crewServerAgent.StartServer(); err != nil {
		display.Errorf("❌ Error starting crew server agent: %v", err)
		return
	}
}

func executeFunction(functionName string, arguments string) (string, error) {
	display.Colorf(display.ColorGreen, "🟢 Executing function: %s with arguments: %s\n", functionName, arguments)

	switch functionName {
	// case "say_hello":
	// 	var args struct {
	// 		Name string `json:"name"`
	// 	}
	// 	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
	// 		return `{"error": "Invalid arguments for say_hello"}`, nil
	// 	}
	// 	hello := fmt.Sprintf("👋 Hello, %s!🙂", args.Name)
	// 	return fmt.Sprintf(`{"message": "%s"}`, hello), nil

	case "save_snippet":
		var args struct {
			FilePath string `json:"file_path"`
			Content  string `json:"content"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for write_file"}`, nil
		}

		display.Colorln(args.Content, display.ColorBrightGreen)

		return fmt.Sprintf(`{"message": "%s"}`, "file is saved"), nil

	case "get_history_messages_of_agent_by_id":
		var args struct {
			AgentID string `json:"agent_id"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return `{"error": "Invalid arguments for get_history_messages_of_agent_by_id"}`, nil
		}
		agent, exists := agentCrew[args.AgentID]
		if !exists {
			return `{"error": "Agent not found"}`, errors.New("agent not found: " + args.AgentID)
		}
		historyMessages := agent.GetMessages()

		for _, msg := range historyMessages {
			display.Infof("🟠 %s: %s", msg.Role, msg.Content)
			display.Separator()
		}

		return fmt.Sprintf(`{"message": "%s"}`, "😂😉😁"), nil

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

	default:
		return `{"error": "Unknown function"}`, fmt.Errorf("unknown function: %s", functionName)
	}
}
