# Guide du Structured Agent

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Définir les types de sortie](#3-définir-les-types-de-sortie)
4. [Configuration de l'agent](#4-configuration-de-lagent)
5. [Configuration du modèle](#5-configuration-du-modèle)
6. [Générer des données structurées](#6-générer-des-données-structurées)
7. [Historique de conversation et messages](#7-historique-de-conversation-et-messages)
8. [Hooks de cycle de vie (StructuredAgentOption)](#8-hooks-de-cycle-de-vie-structuredagentoption)
9. [Gestion du contexte et de l'état](#9-gestion-du-contexte-et-de-létat)
10. [Export JSON et débogage](#10-export-json-et-débogage)
11. [Référence API](#11-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Structured Agent ?

Le `structured.Agent[Output]` est un agent générique fourni par le SDK Nova (`github.com/snipwise/nova`) qui génère une sortie JSON structurée conforme à un type struct Go. Il utilise la génération de JSON Schema à partir de votre struct Go pour instruire le LLM de retourner des données dans un format précis et typé.

Contrairement à un chat agent qui retourne du texte libre, le structured agent retourne toujours un struct Go parsé. Cela le rend idéal pour l'extraction de données, la classification, la reconnaissance d'entités, et toute tâche où vous avez besoin d'une sortie structurée et typée depuis un LLM.

### Quand utiliser un Structured Agent

| Scénario | Agent recommandé |
|---|---|
| Extraire des données structurées depuis du texte (entités, faits, etc.) | `structured.Agent[VotreType]` |
| Classifier du texte en catégories prédéfinies avec métadonnées | `structured.Agent[Classification]` |
| Parser du langage naturel en structs Go typés | `structured.Agent[VotreType]` |
| Détection de topic/intention pour le routage | `orchestrator.Agent` (encapsule le structured agent) |
| IA conversationnelle en texte libre | `chat.Agent` |
| Appels de fonctions / utilisation d'outils | `tools.Agent` |

### Capacités principales

- **Sortie typée générique** : Définissez n'importe quel struct Go comme type de sortie ; l'agent garantit que les réponses du LLM y sont conformes.
- **Application du JSON Schema** : Génère automatiquement un JSON Schema depuis votre struct Go et utilise le mode strict pour une sortie valide garantie.
- **Historique de conversation** : Maintenez optionnellement l'historique entre les appels.
- **Hooks de cycle de vie** : Exécutez une logique personnalisée avant et après chaque génération de données.
- **Export JSON** : Exportez l'historique de conversation pour le débogage ou la persistance.

---

## 2. Démarrage rapide

### Exemple minimal

```go
package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/structured"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

type Country struct {
    Name       string   `json:"name"`
    Capital    string   `json:"capital"`
    Population int      `json:"population"`
    Languages  []string `json:"languages"`
}

func main() {
    ctx := context.Background()

    agent, err := structured.NewAgent[Country](
        ctx,
        agents.Config{
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are an assistant that answers questions about countries.",
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.0),
        },
    )
    if err != nil {
        panic(err)
    }

    response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
        {Role: roles.User, Content: "Parle-moi du Canada."},
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("Nom :", response.Name)
    fmt.Println("Capitale :", response.Capital)
    fmt.Println("Population :", response.Population)
    fmt.Println("Langues :", strings.Join(response.Languages, ", "))
    fmt.Println("Raison de fin :", finishReason)
}
```

---

## 3. Définir les types de sortie

Le structured agent utilise les génériques Go pour imposer le type de sortie. N'importe quel struct Go avec des tags JSON peut être utilisé.

### Struct simple

```go
type Country struct {
    Name       string   `json:"name"`
    Capital    string   `json:"capital"`
    Population int      `json:"population"`
    Languages  []string `json:"languages"`
}

agent, _ := structured.NewAgent[Country](ctx, agentConfig, modelConfig)
```

### Structs imbriqués

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type Person struct {
    Name    string  `json:"name"`
    Age     int     `json:"age"`
    Address Address `json:"address"`
}

agent, _ := structured.NewAgent[Person](ctx, agentConfig, modelConfig)
```

### Sortie de type slice

Vous pouvez utiliser un slice comme type de sortie pour générer des listes :

```go
type Item struct {
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

agent, _ := structured.NewAgent[[]Item](ctx, agentConfig, modelConfig)
```

### Génération du JSON Schema

L'agent génère automatiquement un JSON Schema depuis votre struct Go par réflexion. Le schéma est passé au LLM avec `strict: true` pour garantir que la réponse correspond toujours au format attendu. Les types Go supportés sont mappés aux types JSON Schema :

| Type Go | Type JSON Schema |
|---|---|
| `string` | `string` |
| `int`, `int64` | `integer` |
| `float64` | `number` |
| `bool` | `boolean` |
| `[]T` | `array` de T |
| `struct` | `object` avec propriétés |

---

## 4. Configuration de l'agent

```go
agents.Config{
    Name:                    "structured-agent",   // Nom de l'agent (optionnel)
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1", // URL du moteur LLM (requis)
    APIKey:                  "votre-clé-api",       // Clé API (optionnel)
    SystemInstructions:      "Tu es un assistant qui extrait des informations sur les pays.", // Prompt système (recommandé)
    KeepConversationHistory: false,                 // Habituellement false pour l'extraction
}
```

| Champ | Type | Requis | Description |
|---|---|---|---|
| `Name` | `string` | Non | Identifiant de l'agent pour le logging. |
| `EngineURL` | `string` | Oui | URL du moteur LLM compatible OpenAI. |
| `APIKey` | `string` | Non | Clé API pour les moteurs authentifiés. |
| `SystemInstructions` | `string` | Recommandé | Prompt système définissant la tâche d'extraction/génération. |
| `KeepConversationHistory` | `bool` | Non | Habituellement `false` pour une extraction sans état. Défaut : `false`. |

---

## 5. Configuration du modèle

```go
models.Config{
    Name:        "ai/qwen2.5:1.5B-F16",    // ID du modèle (requis)
    Temperature: models.Float64(0.0),        // 0.0 pour une extraction déterministe
    MaxTokens:   models.Int(2000),            // Longueur maximale de la réponse
}
```

### Paramètres recommandés

- **Temperature** : `0.0` pour une extraction déterministe et factuelle. Des valeurs plus élevées pour la génération créative.
- **Modèle** : Les modèles avec de bonnes capacités de suivi d'instructions et JSON fonctionnent le mieux (Qwen, Llama, etc.).

---

## 6. Générer des données structurées

### GenerateStructuredData

La méthode principale pour générer une sortie typée :

```go
response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Parle-moi de la France."},
})
if err != nil {
    // gérer l'erreur
}

// response est *Country (typé)
fmt.Println(response.Name)       // "France"
fmt.Println(response.Capital)    // "Paris"
fmt.Println(response.Population) // 67390000
fmt.Println(response.Languages)  // ["French"]
fmt.Println(finishReason)        // "stop"
```

**Valeurs de retour :**
- `response *Output` : Un pointeur vers le struct de sortie parsé.
- `finishReason string` : Raison de l'arrêt de la génération (`"stop"`, `"length"`, etc.).
- `err error` : Erreur si la génération ou le parsing a échoué.

### Envoi de plusieurs messages

Vous pouvez fournir un contexte de conversation :

```go
response, _, err := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Je m'intéresse aux pays européens."},
    {Role: roles.Assistant, Content: `{"name":"","capital":"","population":0,"languages":[]}`},
    {Role: roles.User, Content: "Parle-moi de l'Allemagne."},
})
```

---

## 7. Historique de conversation et messages

### Gestion des messages

```go
// Obtenir tous les messages de l'historique
msgs := agent.GetMessages()

// Ajouter un message
agent.AddMessage(roles.User, "Un message manuel")

// Ajouter plusieurs messages d'un coup
agent.AddMessages([]messages.Message{
    {Role: roles.User, Content: "Premier message"},
    {Role: roles.Assistant, Content: "Première réponse"},
})

// Effacer tous les messages sauf l'instruction système
agent.ResetMessages()
```

### Exporter la conversation en JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
if err != nil {
    // gérer l'erreur
}
fmt.Println(jsonStr)
```

---

## 8. Hooks de cycle de vie (StructuredAgentOption)

Les hooks de cycle de vie permettent d'exécuter une logique personnalisée avant et après chaque génération de données structurées. Ils sont configurés comme des options fonctionnelles lors de la création de l'agent.

### StructuredAgentOption

Le type d'option est générique, correspondant au type de sortie de l'agent :

```go
type StructuredAgentOption[Output any] func(*Agent[Output])
```

Les options sont passées en arguments variadiques à `NewAgent` :

```go
agent, err := structured.NewAgent[Country](ctx, agentConfig, modelConfig,
    structured.BeforeCompletion[Country](fn),
    structured.AfterCompletion[Country](fn),
)
```

**Note :** Go peut souvent inférer le paramètre de type, vous pouvez donc l'omettre :

```go
agent, err := structured.NewAgent[Country](ctx, agentConfig, modelConfig,
    structured.BeforeCompletion(fn),
    structured.AfterCompletion(fn),
)
```

### BeforeCompletion

Appelé avant chaque génération de données structurées. Le hook reçoit une référence à l'agent typé.

```go
structured.BeforeCompletion[Country](func(a *structured.Agent[Country]) {
    fmt.Println("Génération de données structurées en cours...")
    fmt.Printf("Nombre de messages : %d\n", len(a.GetMessages()))
})
```

**Cas d'usage :**
- Logging et monitoring
- Collecte de métriques
- Inspection de l'état pré-génération

### AfterCompletion

Appelé après chaque génération de données structurées, une fois le résultat parsé. Le hook reçoit une référence à l'agent typé.

```go
structured.AfterCompletion[Country](func(a *structured.Agent[Country]) {
    fmt.Println("Génération de données structurées terminée.")
    fmt.Printf("Nombre de messages : %d\n", len(a.GetMessages()))
})
```

**Cas d'usage :**
- Logging des résultats de génération
- Métriques post-génération
- Déclenchement d'actions en aval
- Audit/suivi

### Exemple complet avec hooks

```go
agent, err := structured.NewAgent[Country](
    ctx,
    agents.Config{
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "You are an assistant that answers questions about countries.",
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
    },
    structured.BeforeCompletion[Country](func(a *structured.Agent[Country]) {
        fmt.Printf("[AVANT] Agent : %s, Messages : %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
    structured.AfterCompletion[Country](func(a *structured.Agent[Country]) {
        fmt.Printf("[APRÈS] Agent : %s, Messages : %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
)
```

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Le paramètre `...StructuredAgentOption[Output]` est variadique, donc le code existant sans hooks continue de fonctionner sans aucune modification.

---

## 9. Gestion du contexte et de l'état

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
agent.Kind()       // Retourne agents.Structured
agent.GetName()    // Retourne le nom de l'agent depuis la config
agent.GetModelID() // Retourne le nom du modèle depuis la config modèle
```

---

## 10. Export JSON et débogage

### Exporter la conversation en JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
if err != nil {
    // gérer l'erreur
}
fmt.Println(jsonStr)
```

### JSON brut requête/réponse

```go
// JSON brut (non formaté)
rawReq := agent.GetLastRequestRawJSON()
rawResp := agent.GetLastResponseRawJSON()

// JSON formaté (pretty-print)
prettyReq, err := agent.GetLastRequestJSON()
prettyResp, err := agent.GetLastResponseJSON()
```

---

## 11. Référence API

### Constructeur

```go
func NewAgent[Output any](
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    opts ...StructuredAgentOption[Output],
) (*Agent[Output], error)
```

Crée un nouvel agent de données structurées. Le paramètre de type `Output` définit le struct de sortie attendu. Le paramètre `opts` accepte zéro ou plusieurs options fonctionnelles `StructuredAgentOption[Output]`.

---

### Types

```go
// StructuredAgentOption est une option fonctionnelle pour configurer un Agent lors de sa création
type StructuredAgentOption[Output any] func(*Agent[Output])

// StructuredResult représente le résultat de la génération de données structurées
type StructuredResult[Output any] struct {
    Data         *Output
    FinishReason string
}
```

---

### Fonctions d'options

| Fonction | Description |
|---|---|
| `BeforeCompletion[Output any](fn func(*Agent[Output]))` | Définit un hook appelé avant chaque génération de données structurées. |
| `AfterCompletion[Output any](fn func(*Agent[Output]))` | Définit un hook appelé après chaque génération de données structurées. |

---

### Méthodes

| Méthode | Description |
|---|---|
| `GenerateStructuredData(msgs []messages.Message) (*Output, string, error)` | Génère des données structurées à partir de messages. Retourne la sortie typée, la raison de fin et l'erreur. |
| `GetMessages() []messages.Message` | Obtient tous les messages de conversation. |
| `AddMessage(role roles.Role, content string)` | Ajoute un message à l'historique. |
| `AddMessages(msgs []messages.Message)` | Ajoute plusieurs messages à l'historique. |
| `ResetMessages()` | Efface tous les messages sauf l'instruction système. |
| `ExportMessagesToJSON() (string, error)` | Exporte l'historique de conversation en JSON. |
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
| `Kind() agents.Kind` | Retourne `agents.Structured`. |
| `GetName() string` | Retourne le nom de l'agent. |
| `GetModelID() string` | Retourne le nom du modèle. |
