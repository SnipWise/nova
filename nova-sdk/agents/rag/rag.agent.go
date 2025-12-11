package rag

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// VectorRecord represents a vector record with prompt and embedding
type VectorRecord struct {
	ID        string
	Prompt    string
	Embedding []float64
	Metadata  map[string]any
	// CosineSimilarity
	Similarity float64
}

// Agent represents a simplified RAG agent that hides OpenAI SDK details
type Agent struct {
	ctx           context.Context
	config        agents.AgentConfig
	modelConfig   models.Config
	internalAgent *BaseAgent
	log           logger.Logger
}

// NewAgent creates a new simplified RAG agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.AgentConfig,
	modelConfig models.Config,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	// Create internal OpenAI-based agent with converted parameters
	openaiModelConfig := openai.EmbeddingNewParams{
		Model: modelConfig.Name,
	}

	// // Set optional parameters if provided
	// if modelConfig.EncodingFormat != nil {
	// 	openaiModelConfig.EncodingFormat = openai.F(*modelConfig.EncodingFormat)
	// }
	// if modelConfig.Dimensions != nil {
	// 	openaiModelConfig.Dimensions = openai.Int(*modelConfig.Dimensions)
	// }

	internalAgent, err := NewBaseAgent(ctx, agentConfig, openaiModelConfig)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		ctx:           ctx,
		config:        agentConfig,
		modelConfig:   modelConfig,
		internalAgent: internalAgent,
		log:           log,
	}

	return agent, nil
}

// Kind returns the agent type
func (agent *Agent) Kind() agents.Kind {
	return agents.Rag
}

// GenerateEmbedding creates a vector embedding for the given text content
func (agent *Agent) GenerateEmbedding(content string) ([]float64, error) {
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	return agent.internalAgent.GenerateEmbeddingVector(content)
}

// SaveEmbedding generates and saves an embedding for the given content
func (agent *Agent) SaveEmbedding(content string) error {
	if content == "" {
		return errors.New("content cannot be empty")
	}

	return agent.internalAgent.GenerateThenSaveEmbeddingVector(content)
}

// SearchSimilar searches for similar records based on content
// limit is the minimum cosine similarity threshold (1.0 = exact match, 0.0 = no similarity)
func (agent *Agent) SearchSimilar(content string, limit float64) ([]VectorRecord, error) {
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	results, err := agent.internalAgent.SearchSimilarities(content, limit)
	if err != nil {
		return nil, err
	}

	// Convert internal VectorRecord to public VectorRecord
	publicResults := make([]VectorRecord, len(results))
	for i, result := range results {
		publicResults[i] = VectorRecord{
			ID:        result.Id,
			Prompt:    result.Prompt,
			Embedding: result.Embedding,
			//Metadata:   result,
			Similarity: result.CosineSimilarity,
		}
	}

	return publicResults, nil
}

// SearchTopN searches for top N similar records based on content
// limit is the minimum cosine similarity threshold (1.0 = exact match, 0.0 = no similarity)
// n is the maximum number of results to return
func (agent *Agent) SearchTopN(content string, limit float64, n int) ([]VectorRecord, error) {
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	if n <= 0 {
		return nil, errors.New("n must be greater than 0")
	}

	results, err := agent.internalAgent.SearchTopNSimilarities(content, limit, n)
	if err != nil {
		return nil, err
	}

	// Convert internal VectorRecord to public VectorRecord
	publicResults := make([]VectorRecord, len(results))
	for i, result := range results {
		publicResults[i] = VectorRecord{
			ID:        result.Id,
			Prompt:    result.Prompt,
			Embedding: result.Embedding,
			//Metadata:   result.Metadata,
			Similarity: result.CosineSimilarity,
		}
	}

	return publicResults, nil
}
