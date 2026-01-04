---
id: jsonstore-rag
name: Agent RAG avec Store JSON Persistant
category: rag
complexity: intermediate
sample_source: 69
description: Agent RAG avec persistance des embeddings dans un fichier JSON pour réutilisation entre sessions
---

# Agent RAG avec Store JSON Persistant

## Description

Crée un agent RAG Nova qui persiste les embeddings dans un fichier JSON. Permet de charger les embeddings existants au démarrage au lieu de les recalculer à chaque exécution, économisant ainsi du temps et des ressources.

## Cas d'utilisation

- Base de connaissances persistante
- Documentation d'entreprise avec mises à jour incrémentielles
- Chatbots avec mémoire long-terme
- Indexation de gros volumes de documents (one-time indexing)
- Applications RAG en production

## Prérequis

- Go 1.21+
- Nova SDK installé (`go get github.com/snipwise/nova@latest`)
- Modèle d'embedding disponible (ex: mxbai-embed-large)

## Code

```go
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
)

func main() {
	ctx := context.Background()

	// === CONFIGURATION ===
	storePathFile := "./store/knowledge.json"  // Fichier de persistance
	dataPath := "./data"                        // Dossier des documents source

	// === CRÉATION DE L'AGENT RAG ===
	agent, err := rag.NewAgent(
		ctx,
		agents.Config{
			EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
		},
		models.Config{
			Name: "ai/mxbai-embed-large",
		},
	)
	if err != nil {
		panic(err)
	}

	// === CHARGEMENT OU CRÉATION DU STORE ===
	if agent.StoreFileExists(storePathFile) {
		// Charger le store existant
		err := agent.LoadStore(storePathFile)
		if err != nil {
			fmt.Printf("Erreur chargement store %s: %v\n", storePathFile, err)
		}
		fmt.Printf("✅ Store chargé depuis %s\n", storePathFile)
	} else {
		fmt.Printf("📝 Store inexistant. Création d'un nouveau store...\n")

		// Récupérer les fichiers markdown du dossier data
		filesContent, err := files.GetContentFilesWithNames(dataPath, ".md")
		if err != nil {
			fmt.Printf("Erreur lecture fichiers: %v\n", err)
		}

		// Indexer chaque fichier
		for idx, content := range filesContent {
			// Option 1: Découpage par sections markdown
			contentPieces := chunks.SplitMarkdownBySections(content.Content)
			
			// Option 2: Découpage par taille fixe (commenté)
			// contentPieces := chunks.ChunkText(content.Content, 512, 64)

			for _, piece := range contentPieces {
				err = agent.SaveEmbedding(piece)
				if err != nil {
					fmt.Printf("Erreur embedding doc %d: %v\n", idx, err)
				} else {
					fmt.Printf("✅ Indexé: %s\n", content.FileName)
				}
			}
		}

		// Persister le store
		err = agent.PersistStore(storePathFile)
		if err != nil {
			fmt.Printf("Erreur persistance store: %v\n", err)
		}
		fmt.Printf("💾 Store sauvegardé dans %s\n", storePathFile)
	}

	// === RECHERCHE SÉMANTIQUE ===
	queries := []string{
		"Comment configurer l'authentification?",
		"Quelles sont les bonnes pratiques de sécurité?",
	}

	for _, query := range queries {
		fmt.Println(strings.Repeat("=", 50))
		fmt.Printf("🔍 Recherche: %s\n", query)

		similarities, err := agent.SearchSimilar(query, 0.7)
		if err != nil {
			panic(err)
		}

		if len(similarities) == 0 {
			fmt.Println("Aucun résultat trouvé")
		} else {
			for _, sim := range similarities {
				fmt.Println(strings.Repeat("-", 30))
				fmt.Printf("📄 Contenu: %s\n", truncate(sim.Prompt, 200))
				fmt.Printf("📊 Score: %.2f\n", sim.Similarity)
			}
		}
	}
}

// Fonction utilitaire pour tronquer le texte
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
```

## Configuration

```yaml
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
EMBEDDING_MODEL: "ai/mxbai-embed-large"

# Chemins
STORE_PATH: "./store/knowledge.json"
DATA_PATH: "./data"

# Autres modèles d'embedding compatibles:
# - "nomic-embed-text"
# - "all-minilm"
# - "bge-base-en"
```

## Structure du Projet

```
my-rag-app/
├── main.go
├── data/                  # Documents source à indexer
│   ├── doc1.md
│   ├── doc2.md
│   └── ...
└── store/                 # Dossier de persistance (créé automatiquement)
    └── knowledge.json     # Store des embeddings
```

## API de Persistance

### Vérifier l'existence du store

```go
if agent.StoreFileExists(storePathFile) {
    // Le store existe, on peut le charger
}
```

### Charger un store existant

```go
err := agent.LoadStore(storePathFile)
if err != nil {
    // Gérer l'erreur
}
```

### Persister le store

```go
err := agent.PersistStore(storePathFile)
if err != nil {
    // Gérer l'erreur
}
```

## Stratégies de Chunking

### Par sections Markdown (recommandé pour .md)

```go
// Découpe intelligente par titres/sections
contentPieces := chunks.SplitMarkdownBySections(content.Content)
```

### Par taille fixe

```go
// chunkSize: taille max d'un chunk (en caractères)
// overlap: chevauchement entre chunks
contentPieces := chunks.ChunkText(content.Content, 512, 64)
```

## Personnalisation

### Indexation incrémentielle

```go
func addNewDocuments(agent *rag.Agent, storePath string, newFiles []string) error {
    // Charger le store existant
    if agent.StoreFileExists(storePath) {
        agent.LoadStore(storePath)
    }

    // Ajouter les nouveaux documents
    for _, filePath := range newFiles {
        content, _ := os.ReadFile(filePath)
        pieces := chunks.SplitMarkdownBySections(string(content))
        for _, piece := range pieces {
            agent.SaveEmbedding(piece)
        }
    }

    // Sauvegarder le store mis à jour
    return agent.PersistStore(storePath)
}
```

### Multi-format avec toolbox/files

```go
// Markdown
mdFiles, _ := files.GetContentFilesWithNames(dataPath, ".md")

// Texte brut
txtFiles, _ := files.GetContentFilesWithNames(dataPath, ".txt")

// JSON
jsonFiles, _ := files.GetContentFilesWithNames(dataPath, ".json")
```

### Intégration avec Chat Agent

```go
func answerWithRAG(ragAgent *rag.Agent, chatAgent *chat.Agent, question string) string {
    // 1. Récupérer le contexte depuis le store
    similarities, _ := ragAgent.SearchSimilar(question, 0.6)
    
    // 2. Construire le contexte
    var context strings.Builder
    for _, sim := range similarities {
        context.WriteString(sim.Prompt + "\n\n")
    }
    
    // 3. Générer la réponse
    prompt := fmt.Sprintf(`Contexte:
%s

Question: %s

Réponds en utilisant uniquement le contexte fourni.`, 
        context.String(), question)
    
    result, _ := chatAgent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: prompt},
    })
    
    return result.Response
}
```

## API Persistence Complete

### Key Methods

```go
// Check if store file exists
exists := agent.StoreFileExists("./store/knowledge.json")

// Load existing store (fast - no re-indexing)
err := agent.LoadStore("./store/knowledge.json")

// Persist store to file
err := agent.PersistStore("./store/knowledge.json")

// Search similar (works with loaded or in-memory store)
similarities, err := agent.SearchSimilar(query, 0.7)
```

### Complete Persistence Workflow

```go
func setupRAGAgent(ctx context.Context, storeFile, dataDir string) (*rag.Agent, error) {
    agent, err := rag.NewAgent(
        ctx,
        agents.Config{
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        },
        models.Config{
            Name: "ai/mxbai-embed-large",
        },
    )
    if err != nil {
        return nil, err
    }

    // Load or create store
    if agent.StoreFileExists(storeFile) {
        // Store exists - load it (fast)
        err := agent.LoadStore(storeFile)
        if err != nil {
            return nil, fmt.Errorf("failed to load store: %w", err)
        }
        fmt.Printf("Loaded existing store from %s\n", storeFile)
    } else {
        // Store doesn't exist - index documents
        fmt.Printf("Creating new store...\n")

        filesContent, err := files.GetContentFilesWithNames(dataDir, ".md")
        if err != nil {
            return nil, fmt.Errorf("failed to read files: %w", err)
        }

        for idx, content := range filesContent {
            contentPieces := chunks.SplitMarkdownBySections(content.Content)

            for _, piece := range contentPieces {
                if err := agent.SaveEmbedding(piece); err != nil {
                    fmt.Printf("Error indexing doc %d: %v\n", idx, err)
                } else {
                    fmt.Printf("Indexed: %s\n", content.FileName)
                }
            }
        }

        // Save store for future use
        if err := agent.PersistStore(storeFile); err != nil {
            return nil, fmt.Errorf("failed to persist store: %w", err)
        }
        fmt.Printf("Store saved to %s\n", storeFile)
    }

    return agent, nil
}
```

## Notes Importantes

### DO:
- Use `agent.StoreFileExists()` to check before loading
- Use `agent.LoadStore()` on startup to avoid re-indexing
- Use `agent.PersistStore()` after indexing to save embeddings
- Set appropriate similarity threshold (0.6-0.8 recommended)
- Use `chunks.SplitMarkdownBySections()` for markdown files
- Use `files.GetContentFilesWithNames()` to track filenames
- Create `store/` directory or let SDK create it automatically

### DON'T:
- Don't re-index documents every time (use LoadStore instead)
- Don't forget to persist after initial indexing
- Don't use very low thresholds (< 0.5) - too many irrelevant results
- Don't use very high thresholds (> 0.9) - too few results
- Don't index without chunking for large documents
- Don't forget error handling for LoadStore/PersistStore

### Store File Format:
- **Format**: JSON file containing embeddings + original text
- **Size**: Depends on number of chunks and embedding dimensions
- **Location**: Can be anywhere (relative or absolute path)
- **Portability**: Can be copied between machines (model-dependent)

### Performance:
- **First run** (no store): Slow - indexes all documents
- **Subsequent runs** (with store): Fast - loads pre-computed embeddings
- **Search time**: O(n) where n = number of chunks (fast with ~1000s of chunks)
- **Benefit**: 100x faster startup for pre-indexed documents

### Production Considerations:
- For < 10,000 documents: JSON store is fine
- For > 10,000 documents: Consider vector database (Pinecone, Qdrant, Weaviate)
- For distributed systems: Use shared vector store
- For updates: Incremental indexing (load + add + persist)
