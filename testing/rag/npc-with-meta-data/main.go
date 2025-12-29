package main

import (
	"context"
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

// KeywordMetadata represents extracted keywords from content
type KeywordMetadata struct {
	Keywords  []string `json:"keywords"`
	MainTopic string   `json:"main_topic"`
	Category  string   `json:"category"`
}

func main() {
	ctx := context.Background()

	name := "male-dwarf-warrior"
	sheetFilePath := "./docs/" + name + ".md"

	//storePathFile := "./store/" + name + ".json"

	engineURL := "http://localhost:12434/engines/llama.cpp/v1"
	//ragEmbeddingModel := "ai/mxbai-embed-large"
	ragEmbeddingModel := "ai/embeddinggemma:latest"

	metadataModel := "hf.co/menlo/jan-nano-gguf:q4_k_m"

	// huggingface.co/qwen/qwen2.5-3b-instruct-gguf:q4_k_m

	// Create structured agent for keyword extraction
	structuredAgent, err := structured.NewAgent[KeywordMetadata](
		ctx,
		agents.Config{
			EngineURL: engineURL,
		},
		models.Config{
			Name: metadataModel,
		},
	)
	if err != nil {
		display.Errorf("❌ Error creating structured agent: %v", err)
		panic(err)
	}
	display.Infof("✅ Structured agent created for keyword extraction")

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

	// Index each section with keyword extraction
	for idx, piece := range contentPieces {
		display.Infof("📝 Processing section %d/%d", idx+1, len(contentPieces))

		// Extract keywords and metadata using structured agent
		extractionPrompt := fmt.Sprintf(`Analyze the following content and extract relevant metadata.
			Content:
			%s

			Extract:
			- Keywords: only 4 keywords, important terms and concepts from the markdown section title then from the content
			- Main topic: the primary subject (use the markdown section title)
			- Category: type of content
			`,
			piece,
		)

		metadata, _, err := structuredAgent.GenerateStructuredData([]messages.Message{
			{Role: roles.User, Content: extractionPrompt},
		})
		if err != nil {
			display.Errorf("❌ Error extracting keywords from section %d: %v", idx, err)
			// Continue with embedding even if keyword extraction fails
		} else {
			display.Infof("🏷️  Keywords: %v", metadata.Keywords)
			display.Infof("📌 Topic: %s | Category: %s",
				metadata.MainTopic, metadata.Category)

			// Enrich the chunk with metadata
			enrichedPiece := fmt.Sprintf("[METADATA]\nKeywords: %v\nTopic: %s\nCategory: %s\n\nContent:\n%s",
				metadata.Keywords, metadata.MainTopic, metadata.Category, piece,
			)

			piece = enrichedPiece

		}

		// Save embedding with enriched content
		err = ragAgent.SaveEmbedding(piece)
		if err != nil {
			display.Errorf("❌ Error embedding section %d: %v", idx, err)
		} else {
			display.Infof("✅ Indexed section %d/%d", idx+1, len(contentPieces))
			fmt.Println("---- Embedded Content ----")
			fmt.Println(piece)
			display.Separator()
		}
	}
	display.Infof("🎉 Completed indexing character sheet: %s", name)
	display.Separator()

	search(ragAgent, "What is your favorite quote?")
	search(ragAgent, "Tell me about your parents.")

	//search(ragAgent, "Tell me more about your background.")
	search(ragAgent, "Tell me about your background story.")
	//search(ragAgent, "Tell me about your background story and adventures.")
	//search(ragAgent, "Tell me about your background story and adventures in Erebor.")

}

func search(ragAgent *rag.Agent, question string) {
	display.Title("🔍 " + question)

	// === SIMILARITY SEARCH IN RAG STORE ===
	similarRecords, err := ragAgent.SearchTopN(question, 0.4, 5)

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
