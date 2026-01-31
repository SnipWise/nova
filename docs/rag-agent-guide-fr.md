# Guide de l'Agent RAG

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent](#3-configuration-de-lagent)
4. [Configuration du modèle](#4-configuration-du-modèle)
5. [Génération d'embeddings](#5-génération-dembeddings)
6. [Sauvegarde d'embeddings](#6-sauvegarde-dembeddings)
7. [Recherche de contenu similaire](#7-recherche-de-contenu-similaire)
8. [Persistance du store](#8-persistance-du-store)
9. [Utilitaires de chunking](#9-utilitaires-de-chunking)
10. [Options : AgentOption et RagAgentOption](#10-options--agentoption-et-ragagentoption)
11. [Hooks de cycle de vie (RagAgentOption)](#11-hooks-de-cycle-de-vie-ragagentoption)
12. [Gestion du contexte et de l'état](#12-gestion-du-contexte-et-de-létat)
13. [Export JSON et débogage](#13-export-json-et-débogage)
14. [Référence API](#14-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Agent RAG ?

Le `rag.Agent` est un agent spécialisé fourni par le Nova SDK (`github.com/snipwise/nova`) qui gère les workflows de Retrieval-Augmented Generation (RAG). Il génère des embeddings vectoriels à partir de contenu textuel et fournit une recherche par similarité sur un vector store en mémoire.

Contrairement aux agents chat ou structured qui utilisent l'API Chat Completions, l'agent RAG utilise l'**API Embeddings** pour convertir du texte en vecteurs numériques, puis utilise la similarité cosinus pour trouver du contenu sémantiquement similaire.

### Quand utiliser un Agent RAG

| Scénario | Agent recommandé |
|---|---|
| Générer des embeddings vectoriels à partir de texte | `rag.Agent` |
| Recherche par similarité sémantique | `rag.Agent` |
| Construire une base de connaissances pour la récupération contextuelle | `rag.Agent` |
| IA conversationnelle en texte libre | `chat.Agent` |
| Extraction de données structurées | `structured.Agent[T]` |
| Appel de fonctions / utilisation d'outils | `tools.Agent` |
| Détection d'intention et routage | `orchestrator.Agent` |
| Compression de contexte | `compressor.Agent` |

### Capacités clés

- **Génération d'embeddings** : Convertir du contenu textuel en embeddings vectoriels avec n'importe quel modèle d'embedding compatible OpenAI.
- **Vector store en mémoire** : Sauvegarder et gérer les embeddings avec génération automatique d'identifiants.
- **Recherche par similarité** : Trouver du contenu sémantiquement similaire par similarité cosinus avec des seuils configurables.
- **Recherche Top-N** : Récupérer les N résultats les plus similaires au-dessus d'un seuil.
- **Persistance du store** : Sauvegarder et charger le vector store depuis/vers des fichiers JSON.
- **Utilitaires de chunking** : Helpers intégrés pour découper les documents avant l'embedding.
- **Hooks de cycle de vie** : Exécuter de la logique personnalisée avant et après chaque génération d'embedding.

---

## 2. Démarrage rapide

### Exemple minimal

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

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

    // Générer un embedding
    embedding, err := agent.GenerateEmbedding("James T Kirk est le capitaine de l'USS Enterprise.")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Dimension de l'embedding : %d\n", len(embedding))

    // Sauvegarder des documents dans le vector store
    agent.SaveEmbedding("Spock est l'officier scientifique à bord de l'Enterprise.")
    agent.SaveEmbedding("Leonard McCoy est le médecin en chef.")

    // Rechercher du contenu similaire
    results, err := agent.SearchSimilar("Qui est le médecin ?", 0.5)
    if err != nil {
        panic(err)
    }

    for _, r := range results {
        fmt.Printf("Résultat : %s (similarité : %.4f)\n", r.Prompt, r.Similarity)
    }
}
```

---

## 3. Configuration de l'agent

```go
agents.Config{
    Name:      "RAG",                                              // Nom de l'agent (optionnel)
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",      // URL du moteur LLM (requis)
    APIKey:    "your-api-key",                                     // Clé API (optionnel)
}
```

| Champ | Type | Requis | Description |
|---|---|---|---|
| `Name` | `string` | Non | Identifiant de l'agent pour les logs. |
| `EngineURL` | `string` | Oui | URL du moteur LLM compatible OpenAI. |
| `APIKey` | `string` | Non | Clé API pour les moteurs authentifiés. |

**Note :** Contrairement aux agents chat ou structured, l'agent RAG n'utilise pas `SystemInstructions` car il travaille avec l'API Embeddings, pas Chat Completions.

---

## 4. Configuration du modèle

```go
models.Config{
    Name: "ai/mxbai-embed-large",    // ID du modèle d'embedding (requis)
}
```

### Modèles recommandés

- **mxbai-embed-large** : Bon modèle d'embedding généraliste avec 1024 dimensions.
- Choisissez un modèle adapté à vos besoins de recherche sémantique et aux ressources disponibles.

---

## 5. Génération d'embeddings

### GenerateEmbedding

Générer un embedding vectoriel pour un texte donné :

```go
embedding, err := agent.GenerateEmbedding("Du contenu textuel")
if err != nil {
    // gérer l'erreur
}

fmt.Printf("Dimension : %d\n", len(embedding)) // ex : 1024
fmt.Printf("Première valeur : %f\n", embedding[0])
```

**Valeurs de retour :**
- `[]float64` : Le vecteur d'embedding.
- `error` : Erreur si la génération a échoué.

### GetEmbeddingDimension

Obtenir la dimension des vecteurs d'embedding produits par le modèle :

```go
dimension := agent.GetEmbeddingDimension()
fmt.Printf("Dimension de l'embedding : %d\n", dimension) // ex : 1024
```

**Note :** Cette méthode effectue un appel API de test pour déterminer la dimension.

---

## 6. Sauvegarde d'embeddings

### SaveEmbedding / SaveEmbeddingIntoMemoryVectorStore

Générer un embedding et le sauvegarder dans le vector store en mémoire :

```go
err := agent.SaveEmbedding("Spock est un officier scientifique mi-Vulcain.")
if err != nil {
    // gérer l'erreur
}
```

Chaque embedding sauvegardé reçoit automatiquement un identifiant unique. Le store associe le contenu à sa représentation vectorielle pour la recherche de similarité ultérieure.

### Sauvegarder plusieurs documents

```go
documents := []string{
    "James T Kirk est le capitaine de l'Enterprise.",
    "Spock est l'officier scientifique.",
    "Leonard McCoy est le médecin en chef.",
}

for _, doc := range documents {
    err := agent.SaveEmbedding(doc)
    if err != nil {
        fmt.Printf("Échec de la sauvegarde : %v\n", err)
    }
}
```

---

## 7. Recherche de contenu similaire

### SearchSimilar

Rechercher tous les documents au-dessus d'un seuil de similarité :

```go
results, err := agent.SearchSimilar("Qui est le médecin ?", 0.5)
if err != nil {
    // gérer l'erreur
}

for _, r := range results {
    fmt.Printf("Contenu : %s\n", r.Prompt)
    fmt.Printf("Similarité : %.4f\n", r.Similarity)
}
```

**Paramètres :**
- `content string` : Le texte de la requête.
- `limit float64` : Seuil minimum de similarité cosinus (1.0 = correspondance exacte, 0.0 = aucune similarité).

### SearchTopN

Rechercher les N documents les plus similaires au-dessus d'un seuil :

```go
results, err := agent.SearchTopN("Qui est le capitaine ?", 0.5, 3)
if err != nil {
    // gérer l'erreur
}
```

**Paramètres :**
- `content string` : Le texte de la requête.
- `limit float64` : Seuil minimum de similarité cosinus.
- `n int` : Nombre maximum de résultats à retourner.

### VectorRecord

Les résultats de recherche sont retournés sous forme de `[]VectorRecord` :

```go
type VectorRecord struct {
    ID         string
    Prompt     string
    Embedding  []float64
    Metadata   map[string]any
    Similarity float64
}
```

---

## 8. Persistance du store

### Sauvegarder le store sur disque

```go
err := agent.PersistStore("./store/connaissances.json")
if err != nil {
    // gérer l'erreur
}
```

### Charger le store depuis le disque

```go
err := agent.LoadStore("./store/connaissances.json")
if err != nil {
    // gérer l'erreur
}
```

### Vérifier si un fichier de store existe

```go
if agent.StoreFileExists("./store/connaissances.json") {
    agent.LoadStore("./store/connaissances.json")
} else {
    // Construire le store depuis zéro
}
```

### Flux de travail typique de persistance

```go
storeFile := "./store/data.json"

if agent.StoreFileExists(storeFile) {
    agent.LoadStore(storeFile)
} else {
    // Sauvegarder les documents
    for _, doc := range documents {
        agent.SaveEmbedding(doc)
    }
    // Persister pour la prochaine exécution
    agent.PersistStore(storeFile)
}
```

---

## 9. Utilitaires de chunking

Le sous-package `chunks` fournit des utilitaires pour découper les documents avant l'embedding.

### ChunkText

Découper du texte en morceaux de taille fixe avec chevauchement :

```go
import "github.com/snipwise/nova/nova-sdk/agents/rag/chunks"

pieces := chunks.ChunkText(longText, 512, 64) // taille=512, chevauchement=64
for _, piece := range pieces {
    agent.SaveEmbedding(piece)
}
```

### SplitMarkdownBySections

Découper du contenu Markdown par sections (en-têtes) :

```go
sections := chunks.SplitMarkdownBySections(contenuMarkdown)
for _, section := range sections {
    agent.SaveEmbedding(section)
}
```

---

## 10. Options : AgentOption et RagAgentOption

L'agent RAG supporte deux types d'options distincts, tous deux passés comme arguments variadiques `...any` à `NewAgent` :

### AgentOption (niveau de base)

`AgentOption` opère sur le `*BaseAgent` interne et configure le comportement de bas niveau :

```go
// Actuellement disponible pour l'extensibilité
```

### RagAgentOption (niveau agent)

`RagAgentOption` opère sur l'`*Agent` de haut niveau et configure les hooks de cycle de vie :

```go
rag.BeforeCompletion(func(a *rag.Agent) { ... })
rag.AfterCompletion(func(a *rag.Agent) { ... })
```

### Mixer les deux types d'options

Les deux types d'options peuvent être passés ensemble à `NewAgent` :

```go
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    // RagAgentOption (niveau agent)
    rag.BeforeCompletion(func(a *rag.Agent) {
        fmt.Println("Avant la génération d'embedding...")
    }),
    rag.AfterCompletion(func(a *rag.Agent) {
        fmt.Println("Après la génération d'embedding...")
    }),
)
```

---

## 11. Hooks de cycle de vie (RagAgentOption)

Les hooks de cycle de vie permettent d'exécuter de la logique personnalisée avant et après chaque génération d'embedding via la méthode `GenerateEmbedding`. Ils sont configurés comme options fonctionnelles lors de la création de l'agent.

### RagAgentOption

```go
type RagAgentOption func(*Agent)
```

### BeforeCompletion

Appelé avant chaque génération d'embedding dans `GenerateEmbedding`. Le hook reçoit une référence vers l'agent.

```go
rag.BeforeCompletion(func(a *rag.Agent) {
    fmt.Printf("Génération d'embedding en cours... Agent : %s (%s)\n",
        a.GetName(), a.GetModelID())
})
```

**Cas d'utilisation :**
- Logging et monitoring
- Collecte de métriques (ex : compter les générations d'embeddings)
- Limitation de débit ou throttling

### AfterCompletion

Appelé après chaque génération d'embedding dans `GenerateEmbedding`. Le hook reçoit une référence vers l'agent.

```go
rag.AfterCompletion(func(a *rag.Agent) {
    fmt.Printf("Embedding généré. Agent : %s (%s)\n",
        a.GetName(), a.GetModelID())
})
```

**Cas d'utilisation :**
- Logging des résultats
- Métriques post-génération
- Déclenchement d'actions en aval
- Audit/traçabilité

### Exemple complet avec hooks

```go
embeddingCount := 0

agent, err := rag.NewAgent(
    ctx,
    agents.Config{
        Name:      "RAG",
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    },
    models.Config{
        Name: "ai/mxbai-embed-large",
    },
    rag.BeforeCompletion(func(a *rag.Agent) {
        embeddingCount++
        fmt.Printf("[AVANT] Agent : %s, Embedding #%d\n", a.GetName(), embeddingCount)
    }),
    rag.AfterCompletion(func(a *rag.Agent) {
        fmt.Printf("[APRES] Agent : %s, Embedding #%d\n", a.GetName(), embeddingCount)
    }),
)
```

### Note importante sur la portée des hooks

Les hooks sont déclenchés uniquement par les appels directs à `GenerateEmbedding`. Les autres méthodes comme `SaveEmbedding`, `SearchSimilar` et `SearchTopN` utilisent directement `BaseAgent.GenerateEmbeddingVector` en interne et ne déclenchent **pas** les hooks.

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Le paramètre `...any` est variadique, donc le code existant sans hooks continue de fonctionner sans aucune modification.

---

## 12. Gestion du contexte et de l'état

### Obtenir et définir le contexte

```go
ctx := agent.GetContext()
agent.SetContext(newCtx)
```

### Obtenir et définir la configuration

```go
// Configuration de l'agent
config := agent.GetConfig()
agent.SetConfig(newConfig)

// Configuration du modèle
modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)
```

### Métadonnées de l'agent

```go
agent.Kind()       // Retourne agents.Rag
agent.GetName()    // Retourne le nom de l'agent depuis la config
agent.GetModelID() // Retourne le nom du modèle depuis la config modèle
```

---

## 13. Export JSON et débogage

### JSON brut de requête/réponse

```go
// JSON brut (non formaté) de la dernière requête/réponse d'embedding
rawReq := agent.GetLastRequestRawJSON()
rawResp := agent.GetLastResponseRawJSON()

// JSON formaté (pretty-printed)
prettyReq, err := agent.GetLastRequestJSON()
prettyResp, err := agent.GetLastResponseJSON()
```

---

## 14. Référence API

### Constructeur

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    options ...any,
) (*Agent, error)
```

Crée un nouvel agent RAG. Le paramètre `options` accepte à la fois des `AgentOption` (niveau de base) et des `RagAgentOption` (hooks de niveau agent). Le constructeur les sépare en interne par assertion de type.

---

### Types

```go
// VectorRecord représente un enregistrement vectoriel avec prompt et embedding
type VectorRecord struct {
    ID         string
    Prompt     string
    Embedding  []float64
    Metadata   map[string]any
    Similarity float64
}

// RagAgentOption configure l'Agent de haut niveau (ex : hooks de cycle de vie)
type RagAgentOption func(*Agent)

// AgentOption configure le BaseAgent interne
type AgentOption func(*BaseAgent)
```

---

### Fonctions d'options

| Fonction | Type | Description |
|---|---|---|
| `BeforeCompletion(fn func(*Agent))` | `RagAgentOption` | Définit un hook appelé avant chaque génération d'embedding dans `GenerateEmbedding`. |
| `AfterCompletion(fn func(*Agent))` | `RagAgentOption` | Définit un hook appelé après chaque génération d'embedding dans `GenerateEmbedding`. |

---

### Méthodes

| Méthode | Description |
|---|---|
| `GenerateEmbedding(content string) ([]float64, error)` | Générer un embedding vectoriel pour le texte donné. Déclenche les hooks de cycle de vie. |
| `GetEmbeddingDimension() int` | Obtenir la dimension des vecteurs d'embedding produits par le modèle. |
| `SaveEmbedding(content string) error` | Générer et sauvegarder un embedding dans le vector store en mémoire. |
| `SaveEmbeddingIntoMemoryVectorStore(content string) error` | Alias pour `SaveEmbedding`. |
| `SearchSimilar(content string, limit float64) ([]VectorRecord, error)` | Rechercher les enregistrements similaires au-dessus d'un seuil de similarité. |
| `SearchTopN(content string, limit float64, n int) ([]VectorRecord, error)` | Rechercher les N enregistrements les plus similaires au-dessus d'un seuil. |
| `LoadStore(path string) error` | Charger le vector store depuis un fichier JSON. |
| `PersistStore(path string) error` | Sauvegarder le vector store dans un fichier JSON. |
| `StoreFileExists(path string) bool` | Vérifier si un fichier de store existe au chemin donné. |
| `GetConfig() agents.Config` | Obtient la configuration de l'agent. |
| `SetConfig(config agents.Config)` | Met à jour la configuration de l'agent. |
| `GetModelConfig() models.Config` | Obtient la configuration du modèle. |
| `SetModelConfig(config models.Config)` | Met à jour la configuration du modèle. |
| `GetContext() context.Context` | Obtient le contexte de l'agent. |
| `SetContext(ctx context.Context)` | Met à jour le contexte de l'agent. |
| `GetLastRequestRawJSON() string` | Obtient le JSON brut de la dernière requête. |
| `GetLastResponseRawJSON() string` | Obtient le JSON brut de la dernière réponse. |
| `GetLastRequestJSON() (string, error)` | Obtient le JSON formaté de la dernière requête. |
| `GetLastResponseJSON() (string, error)` | Obtient le JSON formaté de la dernière réponse. |
| `Kind() agents.Kind` | Retourne `agents.Rag`. |
| `GetName() string` | Retourne le nom de l'agent. |
| `GetModelID() string` | Retourne le nom du modèle. |
