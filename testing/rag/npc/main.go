package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()

	name := "male-dwarf-warrior"
	sheetFilePath := "./docs/" + name + ".md"

	//storePathFile := "./store/" + name + ".json"

	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	ragEmbeddingModel := "ai/mxbai-embed-large"
	//ragEmbeddingModel := "ai/granite-embedding-multilingual:latest"
	//ragEmbeddingModel := "huggingface.co/second-state/all-minilm-l6-v2-embedding-gguf:q4_k_m"

	ragAgent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL: engineURL,
		},
		models.Config{
			Name: ragEmbeddingModel,
		},
	)
	if err != nil {
		display.Errorf("❌ Error creating RAG agent: %v", err)
		panic(err)
	}

	// Read character sheet content
	characterSheetContent, err := files.ReadTextFile(sheetFilePath)
	if err != nil {
		display.Errorf("❌ Error reading character sheet: %v", err)
		panic(err)
	}

	contentPieces := chunks.SplitMarkdownBySections(characterSheetContent)
	display.Infof("📄 Split character sheet into %d sections", len(contentPieces))

	// Index each section
	for idx, piece := range contentPieces {
		err = ragAgent.SaveEmbedding(piece)
		if err != nil {
			display.Errorf("❌ Error embedding section %d: %v", idx, err)
		} else {
			display.Infof("✅ Indexed section %d/%d", idx+1, len(contentPieces))
			fmt.Println(piece)
		}
	}
	display.Infof("🎉 Completed indexing character sheet: %s", name)
	display.Separator()

	//search(ragAgent, "What is your favorite quote?")
	//search(ragAgent, "Tell me about your parents.")

	search(ragAgent, "Tell me more about your background.")

}

func expandQuery(question string) string {
	// Map de synonymes et expansions
	expansions := map[string]string{
		"background": "background story history adventures past experiences life events",
		"parents":    "father mother family grandfather relatives ancestors",
		"quote":      "favorite quote saying motto philosophy words",
	}

	lowerQ := strings.ToLower(question)
	for key, expansion := range expansions {
		if strings.Contains(lowerQ, key) {
			return question + " " + expansion
		}
	}
	return question
}

func search(ragAgent *rag.Agent, question string) {
	display.Title("🔍 " + question)
	expandedQuestion := expandQuery(question)
	display.Infof("🔎 Expanded Question: %s", expandedQuestion)

	/*
		📄 Context Piece 8 (Score: 0.3657):
		## Background Story
	*/

	// === SIMILARITY SEARCH IN RAG STORE ===
	similarRecords, err := ragAgent.SearchTopN(expandedQuestion, 0.45, 5)
	//similarRecords, err := ragAgent.SearchTopN(question, 0.0, 10)
	//similarRecords, err := ragAgent.SearchSimilar(question, 0.3)
	similarityContext := ""
	if err != nil {
		display.Errorf("failed to search RAG agent: %v", err)
	} else {
		if len(similarRecords) > 0 {
			display.Infof("📚 Retrieved %d relevant context pieces from RAG store.", len(similarRecords))
			for i, record := range similarRecords {
				display.Separator()
				display.Infof("📄 Context Piece %d (Score: %.4f):\n%s", i+1, record.Similarity, record.Prompt)
				similarityContext += "\n" + record.Prompt
			}
			display.Separator()
		} else {
			display.Infof("📚 No relevant context found in RAG store.")
		}
	}
}
