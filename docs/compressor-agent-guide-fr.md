# Guide de l'Agent Compressor

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent](#3-configuration-de-lagent)
4. [Configuration du modèle](#4-configuration-du-modèle)
5. [Instructions et prompts intégrés](#5-instructions-et-prompts-intégrés)
6. [Compression du contexte](#6-compression-du-contexte)
7. [Compression en streaming](#7-compression-en-streaming)
8. [Options : AgentOption et CompressorAgentOption](#8-options--agentoption-et-compressoragentoption)
9. [Hooks de cycle de vie (CompressorAgentOption)](#9-hooks-de-cycle-de-vie-compressoragentoption)
10. [Gestion du contexte et de l'état](#10-gestion-du-contexte-et-de-létat)
11. [Export JSON et débogage](#11-export-json-et-débogage)
12. [Référence API](#12-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Agent Compressor ?

Le `compressor.Agent` est un agent spécialisé fourni par le Nova SDK (`github.com/snipwise/nova`) qui compresse le contexte de conversation. Il prend une liste de messages (typiquement provenant d'un agent chat) et produit un résumé concis qui préserve les faits clés, les décisions et le contexte nécessaire pour continuer la conversation.

C'est essentiel pour gérer les limites de tokens dans les conversations longues : au lieu d'envoyer l'historique complet au LLM, vous le compressez et utilisez le résumé comme nouveau message système.

### Quand utiliser un Agent Compressor

| Scénario | Agent recommandé |
|---|---|
| Compresser le contexte de conversation pour réduire l'utilisation de tokens | `compressor.Agent` |
| Résumer de longues conversations pour la persistance | `compressor.Agent` |
| IA conversationnelle en texte libre | `chat.Agent` |
| Extraction de données structurées | `structured.Agent[T]` |
| Appel de fonctions / utilisation d'outils | `tools.Agent` |
| Détection d'intention et routage | `orchestrator.Agent` |

### Capacités clés

- **Compression standard et en streaming** : Compressez le contexte en un seul appel ou diffusez le résultat morceau par morceau.
- **Instructions et prompts intégrés** : Instructions système prédéfinies (Minimalist, Expert, Effective) et prompts de compression (Minimalist, Structured, UltraShort, ContinuityFocus).
- **Prompts de compression personnalisés** : Remplacez le prompt de compression par défaut à la création ou à l'exécution.
- **Deux types d'options** : `AgentOption` pour la configuration de base (ex : prompt de compression) et `CompressorAgentOption` pour les hooks de l'agent de haut niveau.
- **Hooks de cycle de vie** : Exécutez de la logique personnalisée avant et après chaque compression.
- **Débogage JSON** : Inspectez les payloads bruts de requête/réponse.

---

## 2. Démarrage rapide

### Exemple minimal

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := compressor.NewAgent(
        ctx,
        agents.Config{
            Name:               "Compressor",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: compressor.Instructions.Expert,
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.0),
        },
        compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
    )
    if err != nil {
        panic(err)
    }

    // Messages à compresser (typiquement depuis un agent chat via chatAgent.GetMessages())
    messagesToCompress := []messages.Message{
        {Role: roles.System, Content: "You are a helpful assistant."},
        {Role: roles.User, Content: "Who is James T Kirk?"},
        {Role: roles.Assistant, Content: "James T. Kirk is a fictional character in the Star Trek franchise. He is the captain of the USS Enterprise."},
        {Role: roles.User, Content: "Who is his best friend?"},
        {Role: roles.Assistant, Content: "His best friend is Spock, a half-Vulcan, half-human science officer aboard the Enterprise."},
    }

    result, err := agent.CompressContext(messagesToCompress)
    if err != nil {
        panic(err)
    }

    fmt.Println("Compressé :", result.CompressedText)
    fmt.Println("Raison de fin :", result.FinishReason)
}
```

---

## 3. Configuration de l'agent

```go
agents.Config{
    Name:               "Compressor",                                      // Nom de l'agent (optionnel)
    EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",      // URL du moteur LLM (requis)
    APIKey:             "your-api-key",                                     // Clé API (optionnel)
    SystemInstructions: compressor.Instructions.Expert,                     // Prompt système (recommandé)
}
```

| Champ | Type | Requis | Description |
|---|---|---|---|
| `Name` | `string` | Non | Identifiant de l'agent pour les logs. |
| `EngineURL` | `string` | Oui | URL du moteur LLM compatible OpenAI. |
| `APIKey` | `string` | Non | Clé API pour les moteurs authentifiés. |
| `SystemInstructions` | `string` | Recommandé | Prompt système définissant le comportement de compression. Utilisez l'une des `Instructions` intégrées ou fournissez la vôtre. |

---

## 4. Configuration du modèle

```go
models.Config{
    Name:        "ai/qwen2.5:1.5B-F16",    // ID du modèle (requis)
    Temperature: models.Float64(0.0),        // 0.0 pour une compression déterministe
    MaxTokens:   models.Int(2000),           // Longueur max de la réponse
}
```

### Paramètres recommandés

- **Temperature** : `0.0` pour une compression déterministe et cohérente.
- **Modèle** : Les modèles plus petits (0.5B-1.5B) fonctionnent bien pour les tâches de compression et sont plus rapides.

---

## 5. Instructions et prompts intégrés

Le package compressor fournit des instructions système et des prompts de compression prédéfinis.

### Instructions système (`compressor.Instructions`)

| Instruction | Description |
|---|---|
| `Instructions.Minimalist` | Instruction simple et concise pour un résumé basique. |
| `Instructions.Expert` | Instruction détaillée avec des règles de formatage, de compression et de structure de sortie. |
| `Instructions.Effective` | Instruction équilibrée qui préserve les informations clés, décisions, préférences et contexte émotionnel. |

```go
// Utiliser une instruction intégrée comme SystemInstructions
agents.Config{
    SystemInstructions: compressor.Instructions.Expert,
}
```

### Prompts de compression (`compressor.Prompts`)

Le prompt de compression est l'instruction de niveau utilisateur envoyée avec la conversation à compresser. Il est défini via `WithCompressionPrompt` ou `SetCompressionPrompt`.

| Prompt | Description |
|---|---|
| `Prompts.Minimalist` | (Par défaut) Instruction concise de résumé préservant les faits clés. |
| `Prompts.Structured` | Demande un résumé structuré avec sujets, décisions et contexte (moins de 200 mots). |
| `Prompts.UltraShort` | Extrait uniquement les faits clés, décisions et contexte essentiel. |
| `Prompts.ContinuityFocus` | Se concentre sur la préservation des informations nécessaires pour continuer la discussion naturellement. |

```go
// Définir à la création
compressor.WithCompressionPrompt(compressor.Prompts.Structured)

// Ou changer à l'exécution
agent.SetCompressionPrompt(compressor.Prompts.UltraShort)
```

---

## 6. Compression du contexte

### CompressContext

La méthode principale pour compresser une liste de messages :

```go
result, err := agent.CompressContext(messagesToCompress)
if err != nil {
    // gérer l'erreur
}

fmt.Println(result.CompressedText) // Le résumé compressé
fmt.Println(result.FinishReason)   // "stop", "length", etc.
```

**Valeurs de retour :**
- `result *CompressionResult` : Contient le texte compressé et la raison de fin.
- `err error` : Erreur si la compression a échoué.

### Flux de travail typique avec un agent chat

```go
// 1. Récupérer les messages depuis un agent chat
msgs := chatAgent.GetMessages()

// 2. Compresser le contexte
result, err := compressorAgent.CompressContext(msgs)

// 3. Réinitialiser l'agent chat et utiliser le contexte compressé
chatAgent.ResetMessages()
chatAgent.AddMessage(roles.System, result.CompressedText)

// 4. Continuer la conversation avec une utilisation réduite de tokens
```

---

## 7. Compression en streaming

### CompressContextStream

Pour une sortie en temps réel, utilisez la compression en streaming avec un callback :

```go
result, err := agent.CompressContextStream(
    messagesToCompress,
    func(chunk string, finishReason string) error {
        fmt.Print(chunk) // Afficher chaque morceau à mesure qu'il arrive
        return nil
    },
)
if err != nil {
    // gérer l'erreur
}

fmt.Println(result.CompressedText) // Texte compressé complet
fmt.Println(result.FinishReason)   // "stop", "length", etc.
```

**Paramètres :**
- `messagesList []messages.Message` : Les messages à compresser.
- `callback StreamCallback` : Fonction appelée pour chaque morceau. Retournez une erreur non-nil pour arrêter le streaming.

**Valeurs de retour :**
- `result *CompressionResult` : Le résultat compressé complet (accumulé depuis tous les morceaux).
- `err error` : Erreur si la compression ou le streaming a échoué.

---

## 8. Options : AgentOption et CompressorAgentOption

L'agent compressor supporte deux types d'options distincts, tous deux passés comme arguments variadiques `...any` à `NewAgent` :

### AgentOption (niveau de base)

`AgentOption` opère sur le `*BaseAgent` interne et configure le comportement de bas niveau :

```go
// Définir un prompt de compression personnalisé à la création
compressor.WithCompressionPrompt(compressor.Prompts.Structured)
```

### CompressorAgentOption (niveau agent)

`CompressorAgentOption` opère sur l'`*Agent` de haut niveau et configure les hooks de cycle de vie :

```go
// Définir les hooks de cycle de vie à la création
compressor.BeforeCompletion(func(a *compressor.Agent) { ... })
compressor.AfterCompletion(func(a *compressor.Agent) { ... })
```

### Mixer les deux types d'options

Les deux types d'options peuvent être passés ensemble à `NewAgent`. Le constructeur utilise l'assertion de type pour les séparer :

```go
agent, err := compressor.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    // AgentOption (niveau de base)
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
    // CompressorAgentOption (niveau agent)
    compressor.BeforeCompletion(func(a *compressor.Agent) {
        fmt.Println("Avant la compression...")
    }),
    compressor.AfterCompletion(func(a *compressor.Agent) {
        fmt.Println("Après la compression...")
    }),
)
```

---

## 9. Hooks de cycle de vie (CompressorAgentOption)

Les hooks de cycle de vie permettent d'exécuter de la logique personnalisée avant et après chaque compression (standard et streaming). Ils sont configurés comme options fonctionnelles lors de la création de l'agent.

### CompressorAgentOption

```go
type CompressorAgentOption func(*Agent)
```

Les options sont passées comme arguments variadiques à `NewAgent` aux côtés des `AgentOption` :

```go
agent, err := compressor.NewAgent(ctx, agentConfig, modelConfig,
    compressor.BeforeCompletion(fn),
    compressor.AfterCompletion(fn),
)
```

### BeforeCompletion

Appelé avant chaque compression (standard ou streaming). Le hook reçoit une référence vers l'agent.

```go
compressor.BeforeCompletion(func(a *compressor.Agent) {
    fmt.Println("Compression du contexte en cours...")
    fmt.Printf("Agent : %s (%s)\n", a.GetName(), a.GetModelID())
})
```

**Cas d'utilisation :**
- Logging et monitoring
- Collecte de métriques (ex : compter les compressions)
- Inspection de l'état avant compression

### AfterCompletion

Appelé après chaque compression, une fois le résultat prêt. Le hook reçoit une référence vers l'agent.

```go
compressor.AfterCompletion(func(a *compressor.Agent) {
    fmt.Println("Compression terminée.")
    fmt.Printf("Agent : %s (%s)\n", a.GetName(), a.GetModelID())
})
```

**Cas d'utilisation :**
- Logging des résultats de compression
- Métriques post-compression
- Déclenchement d'actions en aval
- Audit/traçabilité

### Exemple complet avec hooks

```go
agent, err := compressor.NewAgent(
    ctx,
    agents.Config{
        Name:               "Compressor",
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: compressor.Instructions.Expert,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
    },
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
    compressor.BeforeCompletion(func(a *compressor.Agent) {
        fmt.Printf("[AVANT] Agent : %s (%s)\n", a.GetName(), a.GetModelID())
    }),
    compressor.AfterCompletion(func(a *compressor.Agent) {
        fmt.Printf("[APRES] Agent : %s (%s)\n", a.GetName(), a.GetModelID())
    }),
)
```

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Le paramètre `...any` est variadique, donc le code existant sans hooks continue de fonctionner sans aucune modification.

---

## 10. Gestion du contexte et de l'état

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

### Changer le prompt de compression à l'exécution

```go
agent.SetCompressionPrompt(compressor.Prompts.Structured)
```

### Métadonnées de l'agent

```go
agent.GetKind()    // Retourne agents.Compressor
agent.GetName()    // Retourne le nom de l'agent depuis la config
agent.GetModelID() // Retourne le nom du modèle depuis la config modèle
```

---

## 11. Export JSON et débogage

### JSON brut de requête/réponse

```go
// JSON brut (non formaté)
rawReq := agent.GetLastRequestRawJSON()
rawResp := agent.GetLastResponseRawJSON()

// JSON formaté (pretty-printed)
prettyReq, err := agent.GetLastRequestJSON()
prettyResp, err := agent.GetLastResponseJSON()
```

---

## 12. Référence API

### Constructeur

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    options ...any,
) (*Agent, error)
```

Crée un nouvel agent compressor. Le paramètre `options` accepte à la fois des `AgentOption` (niveau de base) et des `CompressorAgentOption` (hooks de niveau agent). Le constructeur les sépare en interne par assertion de type.

---

### Types

```go
// CompressionResult représente le résultat d'une compression de contexte
type CompressionResult struct {
    CompressedText string
    FinishReason   string
}

// StreamCallback est une fonction appelée pour chaque morceau de réponse en streaming
type StreamCallback func(chunk string, finishReason string) error

// AgentOption configure le BaseAgent interne (ex : prompt de compression)
type AgentOption func(*BaseAgent)

// CompressorAgentOption configure l'Agent de haut niveau (ex : hooks de cycle de vie)
type CompressorAgentOption func(*Agent)
```

---

### Constantes intégrées

```go
// Instructions système
compressor.Instructions.Minimalist   // Instruction simple de résumé
compressor.Instructions.Expert       // Instruction détaillée de spécialiste en compression
compressor.Instructions.Effective    // Instruction équilibrée avec sortie structurée

// Prompts de compression
compressor.Prompts.Minimalist        // (Par défaut) Prompt concis de résumé
compressor.Prompts.Structured        // Résumé structuré avec sujets et décisions
compressor.Prompts.UltraShort        // Faits clés et décisions uniquement
compressor.Prompts.ContinuityFocus   // Focus sur la continuité de la conversation
```

---

### Fonctions d'options

| Fonction | Type | Description |
|---|---|---|
| `WithCompressionPrompt(prompt string)` | `AgentOption` | Définit le prompt de compression utilisé lors de la compression du contexte. |
| `BeforeCompletion(fn func(*Agent))` | `CompressorAgentOption` | Définit un hook appelé avant chaque compression (standard et streaming). |
| `AfterCompletion(fn func(*Agent))` | `CompressorAgentOption` | Définit un hook appelé après chaque compression (standard et streaming). |

---

### Méthodes

| Méthode | Description |
|---|---|
| `CompressContext(msgs []messages.Message) (*CompressionResult, error)` | Compresse les messages et retourne le résultat. |
| `CompressContextStream(msgs []messages.Message, cb StreamCallback) (*CompressionResult, error)` | Compresse les messages avec sortie en streaming via callback. |
| `SetCompressionPrompt(prompt string)` | Change le prompt de compression à l'exécution. |
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
| `GetKind() agents.Kind` | Retourne `agents.Compressor`. |
| `GetName() string` | Retourne le nom de l'agent. |
| `GetModelID() string` | Retourne le nom du modèle. |
