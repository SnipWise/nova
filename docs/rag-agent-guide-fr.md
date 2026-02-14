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
9. [Store JSON avec WithJsonStore](#9-store-json-avec-withjsonstore)
10. [Initialisation de documents avec WithDocuments](#10-initialisation-de-documents-avec-withdocuments)
11. [Redis Vector Store](#11-redis-vector-store)
12. [Utilitaires de chunking](#12-utilitaires-de-chunking)
13. [Options : AgentOption et RagAgentOption](#13-options--agentoption-et-ragagentoption)
14. [Hooks de cycle de vie (RagAgentOption)](#14-hooks-de-cycle-de-vie-ragagentoption)
15. [Gestion du contexte et de l'état](#15-gestion-du-contexte-et-de-létat)
16. [Export JSON et débogage](#16-export-json-et-débogage)
17. [Référence API](#17-référence-api)

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
- **Vector store Redis** : Utiliser Redis comme backend persistant avec indexation HNSW pour une recherche ultra-rapide et scalable.
- **Recherche par similarité** : Trouver du contenu sémantiquement similaire par similarité cosinus avec des seuils configurables.
- **Recherche Top-N** : Récupérer les N résultats les plus similaires au-dessus d'un seuil.
- **Persistance du store** : Sauvegarder et charger le vector store depuis/vers des fichiers JSON (Memory) ou Redis.
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

## 9. Store JSON avec WithJsonStore

### Introduction

L'option `WithJsonStore` offre un moyen pratique de charger et persister automatiquement votre vector store depuis un fichier JSON lors de la création de l'agent. Cela élimine le besoin d'appels manuels à `LoadStore`/`PersistStore` dans de nombreux scénarios courants.

### Utilisation de base

```go
agent, err := rag.NewAgent(
    ctx,
    agents.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    },
    models.Config{
        Name: "ai/mxbai-embed-large:latest",
    },
    rag.WithJsonStore("./store/embeddings.json"),
)
```

### Fonctionnement

1. **À la création** : L'agent tente de charger les embeddings existants depuis le fichier JSON spécifié
2. **Fichier existant** : Les données sont chargées en mémoire automatiquement
3. **Fichier manquant** : Un store en mémoire vide est créé
4. **Persistance automatique** : Lorsque combiné avec `WithDocuments`, le store est automatiquement persisté si de nouveaux documents sont ajoutés
5. **Persistance manuelle** : Vous pouvez toujours appeler `agent.PersistStore(filePath)` manuellement pour sauvegarder à tout moment

### ✨ Persistance automatique (Nouveau !)

Lorsque vous utilisez `WithJsonStore` avec `WithDocuments`, l'agent **persiste automatiquement** le store si de nouveaux documents sont ajoutés lors de l'initialisation :

```go
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithJsonStore("./store/embeddings.json"),
    rag.WithDocuments(documents),  // Persiste automatiquement si des documents sont ajoutés !
)
// Plus besoin d'appeler agent.PersistStore() manuellement !
```

**Comment ça fonctionne :**
- Si le store est vide ou que de nouveaux documents sont ajoutés → **persistance automatique**
- Si vous utilisez `DocumentLoadModeSkip` et que le store contient déjà des données → pas de persistance (aucun nouveau document ajouté)
- Si vous utilisez `DocumentLoadModeSkipDuplicates` → persiste uniquement si des documents non-dupliqués sont ajoutés
- Le répertoire parent est automatiquement créé s'il n'existe pas

### Exemple complet

```go
package main

import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()
    storeFile := "./store/knowledge.json"

    // Créer l'agent avec store JSON - charge automatiquement si le fichier existe
    agent, err := rag.NewAgent(
        ctx,
        agents.Config{
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        },
        models.Config{
            Name: "ai/mxbai-embed-large:latest",
        },
        rag.WithJsonStore(storeFile),
    )
    if err != nil {
        panic(err)
    }

    // Ajouter de nouveaux documents
    agent.SaveEmbedding("Nouveau document à ajouter")

    // Sauvegarder les changements sur disque
    agent.PersistStore(storeFile)
}
```

### Avantages

- ✅ Chargement automatique à la création de l'agent
- ✅ Configuration déclarative et propre
- ✅ Pas besoin de vérifier si le fichier existe
- ✅ Repli transparent vers un store vide
- ✅ Contrôle total sur quand persister

### Quand utiliser WithJsonStore

| Scénario | Approche recommandée |
|----------|---------------------|
| Persistance JSON simple | `WithJsonStore` ✅ |
| Contrôle manuel load/persist | `LoadStore`/`PersistStore` |
| Production, grands datasets | `WithRedisStore` (voir section suivante) |
| Temporaire/test | Store en mémoire par défaut |

---

## 10. Initialisation de documents avec WithDocuments

### Introduction

L'option `WithDocuments` permet d'initialiser votre agent RAG avec une liste prédéfinie de documents. C'est parfait pour :
- Pré-charger une base de connaissances
- Amorcer l'agent avec des données initiales
- Simplifier la configuration de l'agent avec du contenu connu

### Utilisation de base

```go
documents := []string{
    "Les écureuils courent dans la forêt",
    "Les oiseaux volent dans le ciel",
    "Les grenouilles nagent dans l'étang",
}

agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithDocuments(documents),
)
```

### Modes de chargement de documents

Lors de l'utilisation de `WithDocuments`, vous pouvez spécifier comment gérer les données existantes dans le store :

```go
type DocumentLoadMode string

const (
    DocumentLoadModeOverwrite  // Effacer les données existantes et charger les nouveaux documents
    DocumentLoadModeMerge      // Ajouter les documents aux données existantes (défaut)
    DocumentLoadModeSkip       // Ignorer le chargement si le store contient déjà des données
    DocumentLoadModeError      // Logger une erreur si le store n'est pas vide
)
```

### Utilisation avec différents modes

#### Mode Merge (par défaut)

Ajoute les nouveaux documents aux documents existants :

```go
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithJsonStore(storeFile),
    rag.WithDocuments(documents, rag.DocumentLoadModeMerge), // ou juste rag.WithDocuments(documents)
)
```

#### Mode Overwrite

Remplace toutes les données existantes :

```go
rag.WithDocuments(documents, rag.DocumentLoadModeOverwrite)
```

#### Mode Skip

Préserve les données existantes, ignore le chargement si le store a du contenu :

```go
rag.WithDocuments(documents, rag.DocumentLoadModeSkip)
```

#### Mode Error

Empêche les écrasements accidentels en loggant une erreur :

```go
rag.WithDocuments(documents, rag.DocumentLoadModeError)
```

### Exemple complet avec Store JSON

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
    storeFile := "./store/animals.json"

    // Base de connaissances initiale
    documents := []string{
        "Les écureuils courent dans la forêt",
        "Les oiseaux volent dans le ciel",
        "Les grenouilles nagent dans l'étang",
        "Les poissons nagent dans la mer",
        "Les lions rugissent dans la savane",
    }

    // Créer l'agent avec store JSON et documents initiaux
    agent, err := rag.NewAgent(
        ctx,
        agents.Config{
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        },
        models.Config{
            Name: "ai/mxbai-embed-large:latest",
        },
        rag.WithJsonStore(storeFile),  // Charger le store existant (s'il existe)
        rag.WithDocuments(documents, rag.DocumentLoadModeSkipDuplicates), // Ajouter les documents, ignorer les doublons
    )
    if err != nil {
        panic(err)
    }

    // ✨ Plus besoin d'appeler agent.PersistStore() - c'est automatique !
    // Le store a été automatiquement persisté lorsque les documents ont été ajoutés

    // Recherche
    results, _ := agent.SearchSimilar("Quels animaux vivent dans l'eau ?", 0.6)
    for _, r := range results {
        fmt.Printf("Résultat : %s (%.3f)\n", r.Prompt, r.Similarity)
    }
}
```

### ⚠️ Important : Ordre des options

Toujours appliquer `WithDocuments` **APRÈS** `WithJsonStore` ou `WithRedisStore` :

```go
// ✅ Ordre correct
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithJsonStore(storeFile),      // 1. Configurer le store d'abord
    rag.WithDocuments(documents),      // 2. Puis charger les documents
)

// ❌ Mauvais ordre - les documents seront chargés dans le store par défaut, puis écrasés par JsonStore
agent, err := rag.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    rag.WithDocuments(documents),      // ❌ Chargé dans le store par défaut
    rag.WithJsonStore(storeFile),      // ❌ Remplace le store, perdant les documents
)
```

### Comportement de persistance

#### Avec Store JSON

✨ **Persistance automatique** lorsque combiné avec `WithDocuments` :

```go
agent, err := rag.NewAgent(
    ctx, agentConfig, modelConfig,
    rag.WithJsonStore(storeFile),
    rag.WithDocuments(documents),
)
// ✅ Automatiquement persisté si de nouveaux documents ont été ajoutés !
```

**Quand la persistance automatique se produit :**
- ✅ Nouveaux documents ajoutés (première exécution ou modes Merge/Overwrite)
- ✅ Documents non-dupliqués ajoutés (mode SkipDuplicates)
- ❌ Aucun nouveau document ajouté (mode Skip avec données existantes)

**Persistance manuelle toujours disponible :**
```go
// Ajouter des documents après la création de l'agent
agent.SaveEmbedding("Nouveau document")
// Persister manuellement les changements
agent.PersistStore(storeFile)
```

#### Avec Store Redis

Les documents sont automatiquement persistés dans Redis :

```go
agent, err := rag.NewAgent(
    ctx, agentConfig, modelConfig,
    rag.WithRedisStore(redisConfig, dimension),
    rag.WithDocuments(documents),
)
// Les documents sont automatiquement sauvegardés dans Redis
```

### Cas d'usage

| Cas d'usage | Mode recommandé |
|----------|-----------------|
| Configuration initiale | `Overwrite` ou `Merge` |
| Mises à jour quotidiennes | `Merge` |
| Garder les données existantes inchangées | `Skip` |
| Empêcher les changements accidentels | `Error` |
| Tests avec ardoise vierge | `Overwrite` |

### Exemple de code

Voir `samples/110-rag-agent-with-json-store/` pour un exemple complet fonctionnel démontrant `WithJsonStore` et `WithDocuments` avec différents modes.

---

## 11. Redis Vector Store

### Introduction : Redis vs In-Memory

Par défaut, le RAG Agent utilise un **vector store en mémoire** qui stocke les embeddings dans la RAM. C'est parfait pour le prototypage et les petits datasets, mais les données sont perdues au redémarrage de l'application.

Le **Redis Vector Store** offre une alternative persistante et scalable :
- 💾 **Persistance** : Les données survivent aux redémarrages
- 🔄 **Partage** : Plusieurs applications peuvent accéder aux mêmes données
- 📈 **Scalabilité** : Support de millions de vecteurs
- ⚡ **Performance** : Indexation HNSW pour une recherche ultra-rapide

### Quand utiliser Redis vs In-Memory

| Critère | In-Memory | Redis |
|---------|-----------|-------|
| **Persistance** | ❌ Perdu au redémarrage | ✅ Survit aux redémarrages |
| **Partage multi-process** | ❌ Un seul process | ✅ Plusieurs applications |
| **Scalabilité** | Limité par la RAM | Millions de vecteurs |
| **Vitesse** | Très rapide | Très rapide (HNSW) |
| **Setup** | Aucun | Nécessite Redis |
| **Cas d'usage** | Prototypage, petits datasets | Production, datasets larges |

### Configuration Redis

Pour utiliser Redis comme backend, vous devez configurer la connexion via `RedisConfig` :

```go
type RedisConfig struct {
    Address   string // Adresse du serveur Redis (ex: "localhost:6379")
    Password  string // Mot de passe Redis (chaîne vide si aucun)
    DB        int    // Numéro de base de données Redis (défaut: 0)
    IndexName string // Nom de l'index de recherche Redis (défaut: "nova_rag_index")
}
```

### Utilisation avec WithRedisStore

Pour créer un agent RAG avec Redis comme backend, utilisez l'option `WithRedisStore` :

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/agents/rag/stores"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

agent, err := rag.NewAgent(
    ctx,
    agents.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    },
    models.Config{
        Name: "ai/mxbai-embed-large", // 1024 dimensions
    },
    // Option Redis
    rag.WithRedisStore(stores.RedisConfig{
        Address:   "localhost:6379",
        Password:  "",                    // Vide si pas de mot de passe
        DB:        0,                     // Base de données par défaut
        IndexName: "my_knowledge_base",   // Nom personnalisé de l'index
    }, 1024), // ⚠️ La dimension DOIT correspondre au modèle d'embedding
)
if err != nil {
    panic(err)
}

// Utilisation identique au store en mémoire
agent.SaveEmbedding("James T Kirk est le capitaine de l'Enterprise.")
agent.SaveEmbedding("Spock est l'officier scientifique.")

// Recherche
results, _ := agent.SearchSimilar("Qui est le capitaine ?", 0.5)
```

### ⚠️ Important : Dimension des embeddings

Le paramètre `dimension` dans `WithRedisStore` **DOIT** correspondre à la dimension des vecteurs produits par votre modèle d'embedding :

| Modèle | Dimension |
|--------|-----------|
| `ai/mxbai-embed-large` | 1024 |
| `text-embedding-3-small` | 1536 |
| `text-embedding-3-large` | 3072 |
| `text-embedding-ada-002` | 1536 |

Vous pouvez vérifier la dimension avec :
```go
dimension := agent.GetEmbeddingDimension()
fmt.Printf("Dimension : %d\n", dimension)
```

### Exemple complet

```go
package main

import (
    "context"
    "fmt"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/rag"
    "github.com/snipwise/nova/nova-sdk/agents/rag/stores"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Créer un agent avec Redis
    agent, err := rag.NewAgent(
        ctx,
        agents.Config{
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        },
        models.Config{
            Name: "ai/mxbai-embed-large",
        },
        rag.WithRedisStore(stores.RedisConfig{
            Address:   "localhost:6379",
            Password:  "",
            DB:        0,
            IndexName: "star_trek_knowledge",
        }, 1024),
    )
    if err != nil {
        panic(err)
    }

    // Sauvegarder des documents
    documents := []string{
        "James T Kirk est le capitaine de l'Enterprise.",
        "Spock est l'officier scientifique mi-Vulcain.",
        "Leonard McCoy est le médecin en chef.",
        "Montgomery Scott est l'ingénieur en chef.",
    }

    for _, doc := range documents {
        err := agent.SaveEmbedding(doc)
        if err != nil {
            fmt.Printf("Erreur : %v\n", err)
        }
    }

    // Rechercher
    results, err := agent.SearchSimilar("Qui est le docteur ?", 0.5)
    if err != nil {
        panic(err)
    }

    for _, r := range results {
        fmt.Printf("Résultat : %s (similarité : %.4f)\n", r.Prompt, r.Similarity)
    }
}
```

### Prérequis : Démarrer Redis

Redis doit être en cours d'exécution avec le support de recherche vectorielle (Redis Stack ou module RediSearch) :

```bash
# Avec Docker
docker run -d \
  --name redis-vector-store \
  -p 6379:6379 \
  redis/redis-stack-server:latest

# Vérifier que Redis fonctionne
docker exec -it redis-vector-store redis-cli ping
# Devrait retourner : PONG
```

### Inspection des données dans Redis

Vous pouvez inspecter les données stockées avec Redis CLI :

```bash
# Accéder à Redis CLI
docker exec -it redis-vector-store redis-cli

# Lister tous les index
FT._LIST

# Voir les détails d'un index
FT.INFO my_knowledge_base

# Lister toutes les clés de documents
KEYS doc:*

# Voir un document spécifique
HGETALL doc:<uuid>

# Compter les documents
DBSIZE
```

### Persistance et redémarrage

L'avantage principal de Redis est la **persistance automatique** :

```bash
# Premier lancement - sauvegarde des données
go run main.go

# Arrêt du programme (Ctrl+C)

# Relancement - les données sont toujours là !
go run main.go
# Les embeddings précédemment sauvegardés sont accessibles
```

Pour repartir de zéro :
```bash
# Supprimer l'index et toutes les données
docker exec -it redis-vector-store redis-cli
FT.DROPINDEX my_knowledge_base DD  # DD = delete documents
```

### Troubleshooting

#### Erreur de connexion Redis

```
❌ Failed to create RAG agent: failed to connect to Redis: dial tcp [::1]:6379: connect: connection refused
```

**Solution** : Démarrez Redis avec la commande Docker ci-dessus.

#### Erreur de dimension

```
Error: vector dimension mismatch
```

**Solution** : Vérifiez que le paramètre `dimension` dans `WithRedisStore` correspond à votre modèle :
```go
dimension := agent.GetEmbeddingDimension()
fmt.Printf("Dimension du modèle : %d\n", dimension)
```

#### Index déjà existant

Redis réutilise les index existants. Si vous voulez créer un index frais :
```bash
docker exec -it redis-vector-store redis-cli
FT.DROPINDEX my_knowledge_base DD
```

### Performance et scalabilité

Le Redis Vector Store utilise l'**algorithme HNSW** (Hierarchical Navigable Small World) pour une recherche de similarité ultra-rapide :

- ⚡ Recherche en temps constant O(log n)
- 📊 Support de millions de vecteurs
- 🎯 Précision élevée avec cosine similarity
- 🔄 Mises à jour en temps réel

**Recommandations :**
- Utilisez Redis pour des datasets > 10 000 documents
- Indexez par batches pour de meilleures performances
- Configurez la persistance Redis (RDB ou AOF) selon vos besoins

---

## 10. Utilitaires de chunking

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

### ChunkXML

Découper du contenu XML en chunks basés sur un tag cible spécifié :

```go
xml := `<menu>
  <item id="1">
    <name>Margherita Pizza</name>
    <price currency="USD">12.99</price>
  </item>
  <item id="2">
    <name>Caesar Salad</name>
    <price currency="USD">8.50</price>
  </item>
</menu>`

chunks := chunks.ChunkXML(xml, "item")
for _, chunk := range chunks {
    agent.SaveEmbedding(chunk)
}
// Chaque chunk contient : <item id="1">...</item>, <item id="2">...</item>, etc.
```

**Fonctionnalités :**
- Extrait tous les éléments correspondant au nom du tag cible
- Préserve automatiquement tous les attributs XML
- Support des tags auto-fermants (`<item ... />`)
- Support des tags avec contenu (`<item>...</item>`)
- Gère correctement les éléments imbriqués

---

## 11. Options : AgentOption et RagAgentOption

L'agent RAG supporte deux types d'options distincts, tous deux passés comme arguments variadiques `...any` à `NewAgent` :

### AgentOption (niveau de base)

`AgentOption` opère sur le `*BaseAgent` interne et configure le comportement de bas niveau tel que le backend de stockage :

```go
// Options de configuration du store
rag.WithInMemoryStore()
rag.WithJsonStore(storeFilePath)
rag.WithRedisStore(stores.RedisConfig{...}, dimension)
rag.WithDocuments(documents, mode)
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
    // Utiliser Redis comme backend (optionnel)
    rag.WithRedisStore(stores.RedisConfig{
        Address:   "localhost:6379",
        Password:  "",
        DB:        0,
        IndexName: "my_index",
    }, 1024),
)
```

---

## 12. Hooks de cycle de vie (RagAgentOption)

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

## 13. Gestion du contexte et de l'état

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

## 14. Export JSON et débogage

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

## 15. Référence API

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

// RedisConfig configure la connexion Redis pour le vector store
type RedisConfig struct {
    Address   string // Adresse du serveur Redis (ex: "localhost:6379")
    Password  string // Mot de passe Redis (chaîne vide si aucun)
    DB        int    // Numéro de base de données Redis (défaut: 0)
    IndexName string // Nom de l'index de recherche Redis (défaut: "nova_rag_index")
}

// DocumentLoadMode définit comment les documents doivent être chargés quand le store contient des données
type DocumentLoadMode string

const (
    DocumentLoadModeOverwrite  // Effacer les données existantes et charger les nouveaux documents
    DocumentLoadModeMerge      // Ajouter les documents aux données existantes (défaut)
    DocumentLoadModeSkip       // Ignorer le chargement si le store contient déjà des données
    DocumentLoadModeError      // Logger une erreur si le store n'est pas vide
)
```

---

### Fonctions d'options

| Fonction | Type | Description |
|---|---|---|
| `BeforeCompletion(fn func(*Agent))` | `RagAgentOption` | Définit un hook appelé avant chaque génération d'embedding dans `GenerateEmbedding`. |
| `AfterCompletion(fn func(*Agent))` | `RagAgentOption` | Définit un hook appelé après chaque génération d'embedding dans `GenerateEmbedding`. |
| `WithInMemoryStore()` | `AgentOption` | Configure l'agent pour utiliser le stockage vectoriel en mémoire (comportement par défaut). |
| `WithJsonStore(storePathFile string)` | `AgentOption` | Configure l'agent pour utiliser le stockage basé sur fichier JSON. Charge automatiquement les données existantes depuis le fichier si il existe. |
| `WithRedisStore(config RedisConfig, dimension int)` | `AgentOption` | Configure Redis comme backend de vector store. Le paramètre `dimension` doit correspondre à la dimension du modèle d'embedding. |
| `WithDocuments(documents []string, mode ...DocumentLoadMode)` | `AgentOption` | Initialise l'agent avec des documents prédéfinis. Le paramètre mode optionnel contrôle le comportement quand le store contient des données (défaut: `DocumentLoadModeMerge`). Doit être appliqué APRÈS les options de configuration du store. |

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
| `PersistStore(path string) error` | Sauvegarder le vector store dans un fichier JSON. Note : Appelé automatiquement lors de l'utilisation de `WithJsonStore` + `WithDocuments` si de nouveaux documents sont ajoutés. |
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
