package main

import (
	"context"
	"text/template"
	"strings"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

type SheetFactMetadata struct {
	FactSheetID string   `json:"fact_sheet_id"`
	Title       string   `json:"title"`
	Category    string   `json:"category"`
	Brand       string   `json:"brand"`
	Symptoms    []string `json:"symptoms"`

	Keywords         []string `json:"keywords"`
	//ShortDescription string   `json:"short_description"`
	//Language         string   `json:"language"`
	//FileName string `json:"file_name,omitempty"`
}

/*
=> KeepConversationHistory est mis à false
*/
func GetMetaDataExtractorAgent(ctx context.Context, engineURL, dataExtractorModelId string) (*structured.Agent[SheetFactMetadata], error) {
	// Structured Extractor agent for sheet fact (to fill meta data fields)
	metaDataExtractionInstructions, err := files.ReadTextFile("./config/meta.data.extraction.instructions.md")
	if err != nil {
		panic(err)
	}

	metaDataExtractorAgent, err := structured.NewAgent[SheetFactMetadata](
		ctx,
		agents.Config{
			Name:                    "metadata-extractor-agent",
			EngineURL:               engineURL,
			SystemInstructions:      metaDataExtractionInstructions,
			KeepConversationHistory: false,
		},
		models.Config{
			Name: dataExtractorModelId,
		},
	)
	if err != nil {
		return nil, err
	}

	return metaDataExtractorAgent, nil
}

func ExtractMetaData(content string, metadataExtractorAgent *structured.Agent[SheetFactMetadata]) (*SheetFactMetadata, error) {
	// Créer un contexte avec timeout de 30 secondes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Canal pour recevoir le résultat
	type result struct {
		metaData *SheetFactMetadata
		err      error
	}
	resultChan := make(chan result, 1)

	// Exécuter l'extraction dans une goroutine
	go func() {
		metaData, _, err := metadataExtractorAgent.GenerateStructuredData([]messages.Message{
			{Role: roles.User, Content: content},
		})
		resultChan <- result{metaData: metaData, err: err}
	}()

	// Attendre soit le résultat, soit le timeout
	select {
	case <-ctx.Done():
		display.Errorf("⏱️  Timeout extracting metadata after 30s")
		return nil, ctx.Err()
	case res := <-resultChan:
		if res.err != nil {
			display.Errorf("❌ Error extracting metadata: %v", res.err)
			return nil, res.err
		}
		return res.metaData, nil
	}
}

var metaDataTemplate = `[METADATA]
ID: {{.FactSheetID}}
Title: {{.Title}}
Category: {{.Category}}
Brand: {{.Brand}}
Symptoms:
{{- range .Symptoms}}
  - {{.}}
{{- end}}
Keywords:
{{- range .Keywords}}
  - {{.}}
{{- end}}
[/METADATA]
`

func MetaDataToString(metaData *SheetFactMetadata) (string, error) {

	tmpl, err := template.New("mets-data-report").Parse(metaDataTemplate)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	err = tmpl.Execute(&builder, metaData)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

var metaDataTemplateXML = `<?xml version="1.0" encoding="UTF-8"?>
<metadata>
    <id>{{.FactSheetID}}</id>
    <title>{{.Title}}</title>
    <category>{{.Category}}</category>
    <brand>{{.Brand}}</brand>
    <symptoms>
{{- range .Symptoms}}
        <symptom>{{.}}</symptom>
{{- end}}
    </symptoms>
    <keywords>
{{- range .Keywords}}
        <keyword>{{.}}</keyword>
{{- end}}
    </keywords>
</metadata>
`

func MetaDataToXMLString(metaData *SheetFactMetadata) (string, error) {

	tmpl, err := template.New("mets-data-xml-report").Parse(metaDataTemplateXML)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	err = tmpl.Execute(&builder, metaData)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}