package main

import (
	"context"
	"testing"

	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// go test -v -run TestExtractMetaData
func TestExtractMetaData(t *testing.T) {

	ctx := context.Background()

	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	metadataModelId := "hf.co/menlo/jan-nano-gguf:q4_k_m"

	// === CREATE METADATA EXTRACTOR AGENT ===
	metadataExtractorAgent, err := GetMetaDataExtractorAgent(ctx, engineURL, metadataModelId)
	if err != nil {
		display.Errorf("❌ Error creating metadata extractor agent: %v", err)
		return
	}
	sampleContent, err := files.ReadTextFile(`./docs/FACT-CHAUD-001_Chaudière_ne_démarre_pas___Erreur_flamme_absente.md`)
	if err != nil {
		display.Errorf("❌ Error reading sample content file: %v", err)
		return
	}

	// === EXTRACT METADATA ===
	metaData, err := ExtractMetaData(sampleContent, metadataExtractorAgent)
	if err != nil {
		display.Errorf("❌ Error extracting metadata: %v", err)
		return
	}
	//display.Infof("✅ Extracted Metadata: %+v", metaData)

	//display.Table("ID", metaData.FactSheetID)
	display.Table("Title", metaData.Title)
	display.Table("Category", metaData.Category)
	display.Table("Brand", metaData.Brand)

	display.ObjectStart("Symptoms")
	for _, symptom := range metaData.Symptoms {
		display.Field("symptom", symptom)
	}
	display.ObjectEnd()

	display.ObjectStart("Keywords")
	for _, keyword := range metaData.Keywords {
		display.Field("keyword", keyword)
	}
	display.ObjectEnd()

	// display.Table("Short Description", metaData.ShortDescription)
	// display.Table("Language", metaData.Language)

	display.Separator()

	txtMetatData, err := MetaDataToString(metaData)
	if err != nil {
		display.Errorf("❌ Error converting metadata to string: %v", err)
		return
	}
	display.Infof("✅ Metadata as text:\n%s", txtMetatData)

	display.Separator()
	xmlMetaData, err := MetaDataToXMLString(metaData)
	if err != nil {
		display.Errorf("❌ Error converting metadata to XML string: %v", err)
		return
	}
	display.Infof("✅ Metadata as XML:\n%s", xmlMetaData)
}

// go test -v -run TestExtractMetaData
