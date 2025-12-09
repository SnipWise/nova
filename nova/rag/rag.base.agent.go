package rag

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/snipwise/nova/nova/agents"
	"github.com/snipwise/nova/nova/toolbox/logger"

	"github.com/snipwise/nova/nova/rag/stores"
)

type BaseAgent struct {
	ctx             context.Context
	config          agents.AgentConfig
	EmbeddingParams openai.EmbeddingNewParams
	openaiClient    openai.Client
	log             logger.Logger

	store stores.MemoryVectorStore
}

type AgentOption func(*BaseAgent)

func NewBaseAgent(
	ctx context.Context,
	agentConfig agents.AgentConfig,
	modelConfig openai.EmbeddingNewParams,
	options ...AgentOption,
) (ragAgent *BaseAgent, err error) {
	// export SNIP_LOG_LEVEL=debug  # Shows all logs
	// export SNIP_LOG_LEVEL=info   # Shows info, warn, error
	// export SNIP_LOG_LEVEL=warn   # Shows warn, error only
	// export SNIP_LOG_LEVEL=error  # Shows errors only
	// export SNIP_LOG_LEVEL=none   # No logging (default)

	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	// Create a vector store
	// store := stores.MemoryVectorStore{
	// 	Records: make(map[string]stores.VectorRecord),
	// }

	client := openai.NewClient(
		option.WithBaseURL(agentConfig.EngineURL),
		option.WithAPIKey("I💙DockerModelRunner"),
	)

	_, err = client.Models.Get(ctx, modelConfig.Model)
	if err != nil {
		log.Error("Model not available:", err)
		return nil, errors.New("model not available on the specified engine URL")
	}
	log.Info("Model %s is available on %s", modelConfig.Model, agentConfig.EngineURL)

	ragAgent = &BaseAgent{
		ctx:             ctx,
		config:          agentConfig,
		EmbeddingParams: modelConfig,
		openaiClient:    client,

		store: stores.MemoryVectorStore{
			Records: make(map[string]stores.VectorRecord),
		},

		log: log,
	}

	return ragAgent, nil
}

// GenerateEmbeddingVector creates a vector embedding for the given text content using the agent's embedding model
func (agent *BaseAgent) GenerateEmbeddingVector(content string) (embeddingVector []float64, err error) {
	// Create embedding parameters using the agent's embedding parameters

	agent.EmbeddingParams.Input = openai.EmbeddingNewParamsInputUnion{
		OfString: openai.String(content),
	}
	// Use the client to create embeddings
	embeddingResponse, err := agent.openaiClient.Embeddings.New(agent.ctx, agent.EmbeddingParams)
	if err != nil {
		return nil, err
	}

	return embeddingResponse.Data[0].Embedding, nil
}

func (agent *BaseAgent) GenerateThenSaveEmbeddingVector(content string) (err error) {
	embeddingVector, err := agent.GenerateEmbeddingVector(content)
	if err != nil {
		return err
	}

	_, errSave := agent.store.Save(stores.VectorRecord{
		Prompt:    content,
		Embedding: embeddingVector,
	})

	if errSave != nil {
		return errSave
	}

	return nil
}

// SearchSimilarities searches the vector store for similar records based on the embedding of the provided content
// and returns the top results up to the specified limit
// Parameters:
//   - content: the text content to generate an embedding for searching.
//   - limit: the minimum cosine distance similarity threshold. 1.0 means exact match, 0.0 means no similarity.
func (agent *BaseAgent) SearchSimilarities(content string, limit float64) (results []stores.VectorRecord, err error) {
	embeddingVector, err := agent.GenerateEmbeddingVector(content)
	if err != nil {
		return nil, err
	}

	vectorRecord := stores.VectorRecord{
		Prompt:    content,
		Embedding: embeddingVector,
	}

	results, err = agent.store.SearchSimilarities(vectorRecord, limit)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// SearchTopNSimilarities searches the vector store for similar records based on the embedding of the provided content
// and returns the top N results above the specified similarity limit
// Parameters:
//   - content: the text content to generate an embedding for searching.
//   - limit: the minimum cosine distance similarity threshold. 1.0 means exact match, 0.0 means no similarity.
//   - n: the maximum number of top similar records to return.
func (agent *BaseAgent) SearchTopNSimilarities(content string, limit float64, n int) (results []stores.VectorRecord, err error) {
	embeddingVector, err := agent.GenerateEmbeddingVector(content)
	if err != nil {
		return nil, err
	}

	vectorRecord := stores.VectorRecord{
		Prompt:    content,
		Embedding: embeddingVector,
	}

	results, err = agent.store.SearchTopNSimilarities(vectorRecord, limit, n)
	if err != nil {
		return nil, err
	}

	return results, nil
}