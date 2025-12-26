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

## Notes Importantes

- Le store JSON contient les embeddings et les textes originaux
- Créer le dossier `store/` avant la première exécution ou laisser le SDK le créer
- La taille du store dépend du nombre de documents indexés
- Pour de très gros volumes, considérer une base vectorielle dédiée
- Le seuil de similarité (0.7) est ajustable selon vos besoins
- `SplitMarkdownBySections` est préférable pour les fichiers .md bien structurés
