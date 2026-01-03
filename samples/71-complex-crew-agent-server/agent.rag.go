package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func CreateRagAgent(ctx context.Context, engineURL, embeddingModelId string) (*rag.Agent, error) {

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
	display.Infof("✅ Rag agent created for knowledge retrieval")


	return ragAgent, nil

}
