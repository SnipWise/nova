package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
)

// CompositeAgent represents an agentic squad composed of multiple specialized agents.
type CompositeAgent struct {
	chatAgents map[string]*chat.Agent

	currentAgent *chat.Agent

	toolsAgent *tools.Agent
	ragAgent   *rag.Agent

	similarityLimit float64
	maxSimilarities int

	contextSizeLimit int
	compressorAgent  *compressor.Agent

	orchestratorAgent *structured.Agent[Intent]
}

// NewCompositeAgent creates and initializes a new CompositeAgent with the specified agents.
func NewCompositeAgent(ctx context.Context, engineURL string, toolsCatalog []mcp.Tool) (*CompositeAgent, error) {
	ca := &CompositeAgent{
		chatAgents:      map[string]*chat.Agent{},
		toolsAgent:      nil,
		ragAgent:        nil,
		compressorAgent: nil,
	}
	contextSizeLimit := env.GetEnvOrDefault("CONTEXT_SIZE_LIMIT", "8000")
	ca.contextSizeLimit = conversion.StringToInt(contextSizeLimit)

	similarityLimit := env.GetEnvOrDefault("SIMILARITY_LIMIT", "0.6")
	ca.similarityLimit = conversion.StringToFloat(similarityLimit)

	maxSimilarities := env.GetEnvOrDefault("MAX_SIMILARITIES", "3")
	ca.maxSimilarities = conversion.StringToInt(maxSimilarities)

	// Initialize agentic squad members
	if err := ca.initializeAgenticSquad(ctx, engineURL); err != nil {
		return nil, err
	}

	// Initialiser l'agent d'orchestration
	if err := ca.initializeOrchestratorAgent(ctx, engineURL); err != nil {
		return nil, err
	}

	// Initialiser l'agent de compression
	if err := ca.initializeCompressorAgent(ctx, engineURL); err != nil {
		return nil, err
	}

	// Initialiser l'agent RAG
	if err := ca.initializeRAGAgent(ctx, engineURL); err != nil {
		return nil, err
	}

	// Initialiser l'agent d'outils
	if err := ca.initializeToolsAgent(ctx, engineURL, toolsCatalog); err != nil {
		return nil, err
	}

	return ca, nil
}

// SetCurrentAgent sets the current active chat agent by its name.
func (ca *CompositeAgent) SetCurrentAgent(agentName string) error {
	agent, exists := ca.chatAgents[agentName]
	if !exists {
		return fmt.Errorf("agent with name %s does not exist", agentName)
	}
	ca.currentAgent = agent
	return nil
}

// GetCurrentAgent retrieves the current active chat agent.
func (ca *CompositeAgent) GetCurrentAgent() (*chat.Agent, error) {
	if ca.currentAgent == nil {
		return nil, fmt.Errorf("current agent is not set")
	}
	return ca.currentAgent, nil
}

func (ca *CompositeAgent) GetCurrentAgentMessages() []messages.Message {
	return ca.currentAgent.GetMessages()
}

func (ca *CompositeAgent) GetCurrentAgentModelId() string {
	return ca.currentAgent.GetModelID()
}

func (ca *CompositeAgent) GetSimilarityLimit() float64 {
	return ca.similarityLimit
}

func (ca *CompositeAgent) GetMaxSimilarities() int {
	return ca.maxSimilarities
}

func (ca *CompositeAgent) GetContextSizeLimit() int {
	return ca.contextSizeLimit
}

// TODO: tool call handling
