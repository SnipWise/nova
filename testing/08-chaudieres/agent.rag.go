package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// getRagAgent creates or loads a RAG agent with JSON store persistence
func GetRagAgent(ctx context.Context, engineURL, embeddingModelId string) (*rag.Agent, error) {

	// === CREATE RAG AGENT ===
	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL: engineURL,
		},
		models.Config{
			Name: embeddingModelId,
		},
	)
	if err != nil {
		display.Errorf("❌ Error creating RAG agent: %v", err)
		return nil, err
	}

	return ragAgent, nil
}
