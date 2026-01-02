package main

import (
	"context"
	"testing"

	"github.com/snipwise/nova/nova-sdk/ui/display"
)

/*
### Principes:
Une fiche de fait (sheet fact) est un document structuré
qui fournit des informations détaillées sur un sujet spécifique,
comme un produit ou un service.

Nous allons considérer que pour le découpage en "chunks",
une fiche (le contenu d'un document markdown) est un chunk unique.
Si la taille d'une fiche/chunk dépasse la limite autorisée (ce que supporte le modèle),
il faudra alors envisager un découpage plus fin
(par section, paragraphe, etc. ou méthode classique avec overlap).

*/

// go test -v -run TestVectorizeSheetFact
func TestVectorizeSheetFact(t *testing.T) {

	ctx := context.Background()

	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	ragEmbeddingModelId := "ai/embeddinggemma:latest"
	metadataModelId := "hf.co/menlo/jan-nano-gguf:q4_k_m"

	// === CREATE METADATA EXTRACTOR AGENT ===
	metadataExtractorAgent, err := GetMetaDataExtractorAgent(ctx, engineURL, metadataModelId)
	if err != nil {
		display.Errorf("❌ Error creating metadata extractor agent: %v", err)
		return
	}

	// === CREATE/LOAD RAG AGENT ===
	ragAgent, err := GetRagAgent(ctx, engineURL, ragEmbeddingModelId)
	if err != nil {
		display.Errorf("❌ Error creating/loading RAG agent: %v", err)
		return
	}

	// === LOAD DATA INTO RAG STORE ===
	//dataPath := "./docs-for-test"
	dataPath := "./docs"
	//storePathFile := "./store-for-test/support.json"
	storePathFile := "./store/support.json"

	err = LoadData(dataPath, storePathFile, ragAgent, metadataExtractorAgent)
	if err != nil {
		display.Errorf("❌ Error loading data into RAG store: %v", err)
		return
	}

	display.Infof("✅ Data loaded into RAG store successfully.")

	// === TEST VECTORIZE A SHEET FACT ===

}

// go test -v -run TestVectorizeSheetFact
