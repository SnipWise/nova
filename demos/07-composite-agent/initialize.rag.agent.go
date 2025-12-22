package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
)

func (ca *CompositeAgent) initializeRAGAgent(ctx context.Context, engineURL string) error {
	ragModel := env.GetEnvOrDefault("RAG_MODEL", "ai/mxbai-embed-large")

	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			Name:      "rag-agent",
			EngineURL: engineURL,
		},
		models.Config{
			Name: ragModel,
		},
	)
	if err != nil {
		return err
	}

	// Load or generate vector store
	ragStorePath := env.GetEnvOrDefault("RAG_STORE_PATH", "./store")
	ragStorePathFile := ragStorePath + "/" + ragAgent.GetName() + ".json"
	if ragAgent.StoreFileExists(ragStorePathFile) {
		err := ragAgent.LoadStore(ragStorePathFile)
		if err != nil {
			return err
		}
	} else {

		// Read markdown files from data directory and generate embeddings
		ragDocumentsPath := env.GetEnvOrDefault("RAG_DOCUMENTS_PATH", "./data")
		contents, err := files.GetContentFiles(ragDocumentsPath, ".md")
		if err != nil {
			return err
		}
		for _, content := range contents {
			piecesOfDoc := chunks.SplitMarkdownBySections(content)

			for _, piece := range piecesOfDoc {
				// TODO: add logging info about progress

				err := ragAgent.SaveEmbedding(piece)
				if err != nil {
					return err
				}
			}
		}

		err = ragAgent.PersistStore(ragStorePathFile)
		if err != nil {
			return err

		}
	}

	ca.ragAgent = ragAgent
	return nil
}
