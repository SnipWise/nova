package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/snipwise/nova/nova-sdk/toolbox/files"
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

// go test -v -run TestGenerateAllMetaData
func TestGenerateAllMetaData(t *testing.T) {

	ctx := context.Background()

	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	//ragEmbeddingModelId := "ai/embeddinggemma:latest"
	metadataModelId := "hf.co/menlo/jan-nano-gguf:q4_k_m"

	// === CREATE METADATA EXTRACTOR AGENT ===
	metadataExtractorAgent, err := GetMetaDataExtractorAgent(ctx, engineURL, metadataModelId)
	if err != nil {
		display.Errorf("❌ Error creating metadata extractor agent: %v", err)
		return
	}

	

	// === LOAD DATA INTO RAG STORE ===
	//dataPath := "./docs-for-test"
	dataPath := "./docs"
	//storePathFile := "./store-for-test/support.json"
	//storePathFile := "./store/support.json"

	filesToBeParsed, err := files.GetContentFilesAsMap(dataPath, ".md")
	if err != nil {
		display.Errorf("❌ Error reading content files from %s: %v", dataPath, err)
		return
	}
	fmt.Printf("📝 %d files to be parsed for metadata extraction.\n", len(filesToBeParsed))

	for fileName, fileContent := range filesToBeParsed {
		display.Infof("📄 Extracting metadate for: %s", fileName)
		sheetFactMetaData, err := ExtractMetaData(fileContent, metadataExtractorAgent)
		if err != nil {
			display.Errorf("❌ Error extracting metadata for file %s: %v", fileName, err)
			continue
		}

		sheetFactMetaData.FileName = fileName

		metaDataContent, err := MetaDataToString(sheetFactMetaData)
		if err != nil {
			display.Errorf("❌ Error converting metadata to string for file %s: %v", fileName, err)
			continue
		}

		display.Infof("✅ Metadata for %s:\n%s", fileName, metaDataContent)

		base := filepath.Base(fileName)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		err = files.WriteTextFile(dataPath+"/"+name+".metadata", metaDataContent)
		if err != nil {
			display.Errorf("❌ Error writing metadata file for %s: %v", fileName, err)
			continue
		}

		display.Infof("✅ Metadata file written for %s: %s", fileName, dataPath+"/"+name+".metadata")

	}

}

// go test -v -run TestGenerateAllMetaData
