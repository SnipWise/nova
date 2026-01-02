package main

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func LoadData(dataPath, storePathFile string, ragAgent *rag.Agent, metadataExtractorAgent *structured.Agent[SheetFactMetadata]) error {

	// dataPath := "./docs"
	// storePathFile := "./store/support.json"

	display.Infof("📦 RAG Store path: %s", storePathFile)

	// === LOAD OR CREATE STORE ===
	if ragAgent.StoreFileExists(storePathFile) {
		// Load existing store
		err := ragAgent.LoadStore(storePathFile)
		if err != nil {
			display.Errorf("❌ Error loading store %s: %v", storePathFile, err)
			return err
		}
		display.Infof("✅ RAG store loaded from %s", storePathFile)
	} else {
		display.Infof("📝 Store not found. Creating new store and indexing character sheet...")

		chunkSize := 1024 // <> Embedding model dimension limit, <> not tokens
		chunkOverlap := 128

		filesToChunk, err := files.GetContentFilesAsMap(dataPath, ".md")

		if err != nil {
			display.Errorf("❌ Error reading content files from %s: %v", dataPath, err)
			return err
		}

		for fileName, fileContent := range filesToChunk {
			display.Infof("📄 Indexing: %s", fileName)

			sheetFactMetaData, err := ExtractMetaData(fileContent, metadataExtractorAgent)

			metaDataContent := ""

			if err != nil {
				display.Errorf("❌ Error extracting metadata for file %s: %v", fileName, err)
				metaDataContent = fmt.Sprintf("[METADATA]file name: %s[/METADATA]", fileName)
			} else {

				metaDataContent, err = MetaDataToString(sheetFactMetaData)
				if err != nil {
					display.Errorf("❌ Error converting metadata to string for file %s: %v", fileName, err)
					metaDataContent = fmt.Sprintf("[METADATA]file name: %s[/METADATA]", fileName)
				}
			}

			display.Separator()
			display.Title("MetaData")
			display.Println(metaDataContent)
			display.Separator()
			display.Infof("✋  metaDataContent size (%v)", len(metaDataContent))
			display.Infof("📝  fileContent size (%v)", len(fileContent))
			display.Separator()

			docChunks := chunks.ChunkText(fileContent, chunkSize, chunkOverlap)

			// Index each new chunk
			for cidx, chunk := range docChunks {
				chunkWithMeta := fmt.Sprintf("%s\n\nContent:\n%s",
					metaDataContent, chunk,
				)
				display.Infof("🤖 chunkWithMeta size (%v)", len(chunkWithMeta))

				err = ragAgent.SaveEmbedding(chunkWithMeta)
				if err != nil {
					display.Errorf("❌ Error embedding file %s chunk %d: %v", fileName, cidx, err)

					rechunks := chunks.ChunkText(fileContent, chunkSize/3, chunkOverlap)
					for rcidx, rchunk := range rechunks {
						rechunkWithMeta := fmt.Sprintf("%s\n\nContent:\n%s",
							metaDataContent, rchunk,
						)
						display.Infof("🤖 REchunkWithMeta size (%v)", len(rechunkWithMeta))
						err = ragAgent.SaveEmbedding(rechunkWithMeta)
						if err != nil {
							display.Errorf("❌ Error RE-embedding file %s chunk %d: %v", fileName, rcidx, err)
							panic(err)
						} else {
							display.Infof("✅ RE-Indexed file: %s chunk %d/%d", fileName, rcidx+1, len(rechunks))
						}
					}

				} else {
					display.Infof("✅ Indexed file: %s chunk %d/%d", fileName, cidx+1, len(docChunks))
				}
				//time.Sleep(500 * time.Millisecond) // To avoid rate limits

			}

		}

		// === PERSIST STORE TO DISK ===
		err = ragAgent.PersistStore(storePathFile)
		if err != nil {
			display.Errorf("❌ Error persisting store: %v", err)
			return err
		}
		display.Infof("💾 RAG store saved to %s", storePathFile)
	}
	return nil
}
