# RAG Agent

## Description

Le **RAG Agent** (Retrieval-Augmented Generation) est un agent spécialisé dans la génération d'embeddings vectoriels et la recherche de similarité. Il permet de stocker des contenus textuels sous forme de vecteurs et de rechercher les contenus les plus similaires à une requête donnée.

## Fonctionnalités

- **Génération d'embeddings** : Convertit du texte en vecteurs numériques
- **Stockage vectoriel** : Sauvegarde les embeddings en mémoire
- **Recherche de similarité** : Trouve les contenus les plus similaires via similarité cosinus
- **Persistance** : Sauvegarde et charge le vector store depuis un fichier JSON
- **Top-N Search** : Récupère les N meilleurs résultats similaires

## Cas d'usage

Le RAG Agent est utilisé pour :
- **Enrichir le contexte** des agents de chat avec des informations pertinentes
- **Créer une base de connaissances** interrogeable par similarité sémantique
- **Recherche sémantique** dans des documents, FAQs, documentations
- **Recommandation** de contenus similaires

## Création d'un RAG Agent

### Syntaxe de base

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

// Configuration de l'agent
agentConfig := agents.Config{
    Name: "RAG",
}

// Configuration du modèle d'embeddings
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    Name:      "mxbai-embed-large", // Modèle d'embeddings
}

// Créer l'agent
agent, err := rag.NewAgent(ctx, agentConfig, modelConfig)
if err != nil {
    log.Fatal(err)
}
```

## Structure VectorRecord

Les résultats de recherche retournent des objets `VectorRecord` :

```go
type VectorRecord struct {
    ID         string         // Identifiant unique du record
    Prompt     string         // Le contenu textuel original
    Embedding  []float64      // Le vecteur d'embedding
    Metadata   map[string]any // Métadonnées optionnelles
    Similarity float64        // Score de similarité cosinus (0.0 - 1.0)
}
```

## Méthodes principales

### Génération d'embeddings

```go
// Générer un embedding pour du texte
embedding, err := agent.GenerateEmbedding("Comment faire une pizza ?")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Embedding vector: %d dimensions\n", len(embedding))

// Obtenir la dimension des embeddings du modèle
dimension := agent.GetEmbeddingDimension()
fmt.Printf("Model dimension: %d\n", dimension) // ex: 1024
```

### Sauvegarde d'embeddings

```go
// Sauvegarder un embedding dans le vector store en mémoire
err := agent.SaveEmbedding("La pizza napolitaine se prépare avec de la farine tipo 00.")
if err != nil {
    log.Fatal(err)
}

// Alternative (même fonction)
err = agent.SaveEmbeddingIntoMemoryVectorStore("La pâte doit lever pendant 24 heures.")
```

### Recherche de similarité

```go
// Rechercher tous les contenus similaires avec un seuil de similarité
results, err := agent.SearchSimilar("Comment préparer la pâte à pizza ?", 0.6)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Similarity: %.2f - Content: %s\n", result.Similarity, result.Prompt)
}
```

**Paramètres** :
- `content` : Le texte de recherche
- `limit` : Seuil minimum de similarité cosinus (0.0 - 1.0)
  - 1.0 = correspondance exacte
  - 0.8-1.0 = très similaire
  - 0.6-0.8 = similaire
  - 0.0-0.6 = peu similaire

### Recherche Top-N

```go
// Rechercher les 3 meilleurs résultats avec un seuil de 0.6
results, err := agent.SearchTopN("Comment faire lever la pâte ?", 0.6, 3)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d results:\n", len(results))
for i, result := range results {
    fmt.Printf("%d. [%.2f] %s\n", i+1, result.Similarity, result.Prompt)
}
```

**Paramètres** :
- `content` : Le texte de recherche
- `limit` : Seuil minimum de similarité (0.0 - 1.0)
- `n` : Nombre maximum de résultats à retourner

### Persistance du vector store

```go
// Sauvegarder le vector store dans un fichier JSON
err := agent.PersistStore("./data/knowledge.json")
if err != nil {
    log.Fatal(err)
}

// Vérifier si le fichier existe
exists := agent.StoreFileExists("./data/knowledge.json")
fmt.Printf("Store file exists: %v\n", exists)

// Charger le vector store depuis un fichier JSON
err = agent.LoadStore("./data/knowledge.json")
if err != nil {
    log.Fatal(err)
}
```

### Getters et Setters

```go
// Configuration
config := agent.GetConfig()
agent.SetConfig(newConfig)

modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig) // Note: Nécessite de recréer l'agent

// Informations
name := agent.GetName()
modelID := agent.GetModelID()
kind := agent.Kind() // Retourne agents.Rag

// Contexte
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requêtes/Réponses (debugging)
lastRequestJSON, _ := agent.GetLastRequestJSON()
lastResponseJSON, _ := agent.GetLastResponseJSON()
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
```

## Utilisation avec d'autres agents

Le RAG Agent est généralement utilisé avec Server ou Crew agents pour enrichir automatiquement le contexte :

```go
// Créer le RAG agent
ragAgent, _ := rag.NewAgent(ctx, agentConfig, modelConfig)

// Peupler la base de connaissances
ragAgent.SaveEmbedding("La pizza napolitaine se cuit à 450°C pendant 90 secondes.")
ragAgent.SaveEmbedding("La farine tipo 00 est idéale pour la pizza.")
ragAgent.SaveEmbedding("La mozzarella di bufala est traditionnellement utilisée.")

// Utiliser avec Server Agent
serverAgent, _ := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithRagAgentAndSimilarityConfig(ragAgent, 0.6, 3),
)

// Utiliser avec Crew Agent
crewAgent, _ := crew.NewAgent(
    ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithRagAgentAndSimilarityConfig(ragAgent, 0.6, 3),
)

// Lors d'une requête, le contexte est automatiquement enrichi
// avec les 3 contenus les plus similaires (seuil 0.6)
```

## Exemple complet

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Configuration
    agentConfig := agents.Config{
        Name: "PizzaKnowledge",
    }
    modelConfig := models.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        Name:      "mxbai-embed-large",
    }

    // Créer le RAG agent
    agent, err := rag.NewAgent(ctx, agentConfig, modelConfig)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Embedding dimension: %d\n", agent.GetEmbeddingDimension())

    // Peupler la base de connaissances
    knowledge := []string{
        "La pizza napolitaine se cuit à 450°C pendant 90 secondes dans un four à bois.",
        "La farine tipo 00 est la meilleure pour la pâte à pizza napolitaine.",
        "La mozzarella di bufala campana DOP est traditionnellement utilisée.",
        "La pâte doit lever pendant au moins 8 heures, idéalement 24-48 heures.",
        "La sauce tomate est faite avec des tomates San Marzano DOP.",
        "L'huile d'olive extra vierge est ajoutée après la cuisson.",
    }

    for _, content := range knowledge {
        if err := agent.SaveEmbedding(content); err != nil {
            log.Printf("Error saving: %v", err)
        }
    }

    // Sauvegarder dans un fichier
    if err := agent.PersistStore("./pizza-knowledge.json"); err != nil {
        log.Fatal(err)
    }
    fmt.Println("✅ Knowledge base saved")

    // Recherche de similarité
    query := "Quelle température pour cuire la pizza ?"
    results, err := agent.SearchTopN(query, 0.5, 2)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n🔍 Query: %s\n", query)
    fmt.Printf("Found %d results:\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. [Similarity: %.2f]\n   %s\n\n",
            i+1, result.Similarity, result.Prompt)
    }
}
```

**Sortie attendue** :
```
Embedding dimension: 1024
✅ Knowledge base saved

🔍 Query: Quelle température pour cuire la pizza ?
Found 2 results:
1. [Similarity: 0.87]
   La pizza napolitaine se cuit à 450°C pendant 90 secondes dans un four à bois.

2. [Similarity: 0.62]
   La sauce tomate est faite avec des tomates San Marzano DOP.
```

## Similarité cosinus

Le RAG Agent utilise la **similarité cosinus** pour comparer les vecteurs :

- **1.0** : Vecteurs identiques (parfaite correspondance)
- **0.8-1.0** : Très similaires
- **0.6-0.8** : Modérément similaires
- **0.4-0.6** : Peu similaires
- **0.0-0.4** : Très peu similaires
- **0.0** : Aucune similarité

**Recommandation de seuils** :
- `0.7-0.8` : Pour des correspondances précises
- `0.6` : Bon équilibre (recommandé)
- `0.5` : Pour plus de résultats, moins précis

## Notes

- **Kind** : Retourne `agents.Rag`
- **Vector Store** : Stockage en mémoire avec persistance JSON
- **Dimension** : Dépend du modèle (ex: `mxbai-embed-large` = 1024 dimensions)
- **Erreur si vide** : Retourne une erreur si `content` est vide
- **Top-N** : Retourne au maximum N résultats, triés par similarité décroissante
- **Persistance** : Format JSON, peut être partagé entre instances

## Recommandations

### Modèles d'embeddings recommandés

- **mxbai-embed-large** : 1024 dimensions, excellent équilibre qualité/vitesse
- **nomic-embed-text** : 768 dimensions, rapide et efficace
- **all-minilm** : 384 dimensions, très rapide, moins précis

### Bonnes pratiques

1. **Chunking** : Divisez les longs documents en chunks de 200-500 mots
2. **Seuil de similarité** : Commencez avec 0.6, ajustez selon vos besoins
3. **Top-N** : Limitez à 3-5 résultats pour éviter le bruit
4. **Persistance** : Sauvegardez régulièrement le vector store
5. **Chargement initial** : Vérifiez si le fichier existe avant de peupler

```go
// Charger ou créer
if agent.StoreFileExists("./knowledge.json") {
    agent.LoadStore("./knowledge.json")
} else {
    // Peupler la base
    for _, content := range knowledge {
        agent.SaveEmbedding(content)
    }
    agent.PersistStore("./knowledge.json")
}
```
