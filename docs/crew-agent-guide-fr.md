# Guide du Crew Agent

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent (Options)](#3-configuration-de-lagent-options)
4. [Gestion de l'équipe](#4-gestion-de-léquipe)
5. [Pipeline StreamCompletion](#5-pipeline-streamcompletion)
6. [Routage intelligent (Orchestrateur)](#6-routage-intelligent-orchestrateur)
7. [Intégration des outils](#7-intégration-des-outils)
8. [Intégration RAG](#8-intégration-rag)
9. [Compression du contexte](#9-compression-du-contexte)
10. [Hooks de cycle de vie (BeforeCompletion / AfterCompletion)](#10-hooks-de-cycle-de-vie-beforecompletion--aftercompletion)
11. [Méthodes de complétion directes](#11-méthodes-de-complétion-directes)
12. [Gestion de la conversation](#12-gestion-de-la-conversation)
13. [Gestion du contexte](#13-gestion-du-contexte)
14. [Constructeur legacy (NewSimpleAgent)](#14-constructeur-legacy-newsimpleagent)
15. [Référence API](#15-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Crew Agent ?

Le `crew.CrewAgent` est un agent composite de haut niveau fourni par le SDK Nova (`github.com/snipwise/nova`) qui gère une **équipe de plusieurs agents chat** et route entre eux en fonction des sujets. Il orchestre les appels d'outils, l'injection de contexte RAG, la compression du contexte et le routage intelligent dans un seul pipeline.

### Quand utiliser un Crew Agent

| Scénario | Agent recommandé |
|---|---|
| Plusieurs agents spécialisés avec routage par sujet | `crew.CrewAgent` |
| Pipeline complet : outils + RAG + compression + routage | `crew.CrewAgent` |
| Agent unique avec outils, RAG et compression | `crew.CrewAgent` (via `WithSingleAgent`) |
| Accès direct simple au LLM | `chat.Agent` |

### Capacités principales

- **Équipe multi-agents** : Gère plusieurs instances de `chat.Agent`, chacune spécialisée pour un sujet.
- **Routage intelligent** : Route automatiquement les questions vers l'agent le plus approprié via un orchestrateur.
- **Pipeline complet** : Compression du contexte, appels d'outils, injection RAG et complétion en streaming.
- **Gestion dynamique** : Ajouter ou supprimer des agents à l'exécution.
- **Hooks de cycle de vie** : Exécute une logique personnalisée avant et après chaque complétion.
- **Pattern d'options fonctionnelles** : Configurable via des fonctions `CrewAgentOption`.

---

## 2. Démarrage rapide

### Exemple minimal avec un seul agent

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/crew"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    chatAgent, _ := chat.NewAgent(ctx,
        agents.Config{
            Name:               "assistant",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "Tu es un assistant utile.",
        },
        models.Config{
            Name:        "my-model",
            Temperature: models.Float64(0.4),
        },
    )

    crewAgent, _ := crew.NewAgent(ctx,
        crew.WithSingleAgent(chatAgent),
    )

    result, _ := crewAgent.StreamCompletion("Bonjour !", func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        return nil
    })

    fmt.Println("\nRaison de fin :", result.FinishReason)
}
```

### Exemple avec plusieurs agents

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "generic": genericAgent,
}

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
        switch strings.ToLower(topic) {
        case "coding", "programming":
            return "coder"
        case "philosophy", "thinking":
            return "thinker"
        default:
            return "generic"
        }
    }),
)
```

---

## 3. Configuration de l'agent (Options)

Les options sont passées en arguments variadiques à `NewAgent` :

```go
crewAgent, err := crew.NewAgent(ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithToolsAgent(toolsAgent),
    crew.WithRagAgent(ragAgent),
    crew.WithCompressorAgent(compressorAgent),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.BeforeCompletion(beforeFn),
    crew.AfterCompletion(afterFn),
)
```

| Option | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Définit l'équipe d'agents et l'agent sélectionné initialement. **Obligatoire** (ou `WithSingleAgent`). |
| `WithSingleAgent(chatAgent)` | Crée une équipe avec un seul agent (ID : `"single"`). **Obligatoire** (ou `WithAgentCrew`). |
| `WithMatchAgentIdToTopicFn(fn)` | Définit la fonction de correspondance sujet → ID d'agent. |
| `WithExecuteFn(fn)` | Définit la fonction d'exécution des outils. |
| `WithConfirmationPromptFn(fn)` | Définit la fonction de confirmation pour le human-in-the-loop. |
| `WithToolsAgent(toolsAgent)` | Attache un agent d'outils pour les appels de fonctions. |
| `WithCompressorAgent(compressorAgent)` | Attache un agent compresseur pour la compression du contexte. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attache un compresseur avec une limite de taille de contexte. |
| `WithRagAgent(ragAgent)` | Attache un agent RAG pour la recherche de documents. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attache un agent RAG avec configuration de similarité. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attache un orchestrateur pour la détection de sujets et le routage. |
| `BeforeCompletion(fn)` | Définit un hook appelé avant chaque `StreamCompletion`. |
| `AfterCompletion(fn)` | Définit un hook appelé après chaque `StreamCompletion`. |

### Valeurs par défaut

| Paramètre | Valeur par défaut |
|---|---|
| `similarityLimit` | `0.6` |
| `maxSimilarities` | `3` |
| `contextSizeLimit` | `8000` |

---

## 4. Gestion de l'équipe

### Équipe statique (à la création)

```go
crew.WithAgentCrew(map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
}, "coder")
```

### Gestion dynamique

```go
// Ajouter un agent à l'exécution
err := crewAgent.AddChatAgentToCrew("cook", cookAgent)

// Supprimer un agent (impossible de supprimer l'agent actif)
err := crewAgent.RemoveChatAgentFromCrew("thinker")

// Récupérer tous les agents
agents := crewAgent.GetChatAgents()

// Remplacer toute l'équipe
crewAgent.SetChatAgents(newCrew)
```

### Changer d'agent manuellement

```go
// Récupérer l'agent actuellement sélectionné
id := crewAgent.GetSelectedAgentId()

// Basculer vers un autre agent
err := crewAgent.SetSelectedAgentId("coder")
```

**Note :** Un seul agent est actif à la fois. `GetName()`, `GetModelID()`, `GetMessages()`, etc. opèrent tous sur l'agent actuellement actif.

---

## 5. Pipeline StreamCompletion

La méthode `StreamCompletion` est le point d'entrée principal du crew agent. Elle orchestre le pipeline complet :

```go
result, err := crewAgent.StreamCompletion(question, func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
})
```

### Étapes du pipeline

1. **Hook BeforeCompletion** (si défini)
2. **Compression du contexte** (si l'agent compresseur est configuré et que le contexte dépasse la limite)
3. **Détection et exécution des appels d'outils** (si l'agent d'outils est configuré)
4. **Injection du contexte RAG** (si l'agent RAG est configuré)
5. **Détection du sujet et routage vers l'agent approprié** (si l'orchestrateur est configuré)
6. **Complétion en streaming** avec l'agent actif
7. **Hook AfterCompletion** (si défini)

---

## 6. Routage intelligent (Orchestrateur)

Quand un orchestrateur est attaché, le crew agent peut automatiquement router les questions vers l'agent spécialisé le plus approprié.

### Configuration

```go
orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:               "orchestrator",
        EngineURL:          engineURL,
        SystemInstructions: `Identifie le sujet principal en un mot.
            Sujets possibles : Technology, Philosophy, Cooking, etc.
            Réponds en JSON avec 'topic_discussion'.`,
    },
    models.Config{Name: "my-model", Temperature: models.Float64(0.0)},
)

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
        switch strings.ToLower(topic) {
        case "coding", "programming":
            return "coder"
        case "cooking", "food":
            return "cook"
        default:
            return "generic"
        }
    }),
)
```

### Fonctionnement

1. L'orchestrateur analyse la question de l'utilisateur et détecte le sujet.
2. La `matchAgentIdToTopicFn` fait correspondre le sujet à un ID d'agent.
3. Le crew agent bascule vers l'agent correspondant s'il est différent de l'actuel.
4. La complétion est générée par l'agent nouvellement sélectionné.

### Détection directe de sujet

```go
agentId, err := crewAgent.DetectTopicThenGetAgentId("Écris une fonction Python")
// agentId = "coder"
```

---

## 7. Intégration des outils

```go
toolsAgent, _ := tools.NewAgent(ctx, toolsConfig, toolsModelConfig,
    tools.WithTools(myTools),
)

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithToolsAgent(toolsAgent),
    crew.WithExecuteFn(func(name string, args string) (string, error) {
        return `{"result": "ok"}`, nil
    }),
    crew.WithConfirmationPromptFn(func(name string, args string) tools.ConfirmationResponse {
        return tools.Confirm
    }),
)
```

Les appels d'outils sont détectés et exécutés pendant le pipeline `StreamCompletion`. Les résultats sont injectés dans le contexte du chat avant de générer la complétion finale.

---

## 8. Intégration RAG

```go
ragAgent, _ := rag.NewAgent(ctx, ragConfig, ragModelConfig)

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 5),
)
```

Pendant `StreamCompletion`, le crew agent effectue une recherche de similarité et injecte le contexte pertinent dans la conversation avant de générer la complétion.

---

## 9. Compression du contexte

```go
compressorAgent, _ := compressor.NewAgent(ctx, compressorConfig, compressorModelConfig,
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
)

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)
```

Au début de chaque `StreamCompletion`, le contexte est compressé s'il dépasse la limite configurée.

### Compression manuelle

```go
// Compresser seulement si au-dessus de la limite
newSize, err := crewAgent.CompressChatAgentContextIfOverLimit()

// Forcer la compression
newSize, err := crewAgent.CompressChatAgentContext()
```

---

## 10. Hooks de cycle de vie (BeforeCompletion / AfterCompletion)

Les hooks de cycle de vie permettent d'exécuter une logique personnalisée avant et après chaque appel à `StreamCompletion`. Ils sont configurés comme des options fonctionnelles `CrewAgentOption`.

### BeforeCompletion

Appelé avant chaque `StreamCompletion`. Le hook reçoit une référence vers le crew agent.

```go
crew.BeforeCompletion(func(a *crew.CrewAgent) {
    fmt.Printf("[AVANT] Agent : %s\n", a.GetName())
})
```

### AfterCompletion

Appelé après chaque `StreamCompletion`. Le hook reçoit une référence vers le crew agent.

```go
crew.AfterCompletion(func(a *crew.CrewAgent) {
    fmt.Printf("[APRÈS] Agent : %s\n", a.GetName())
})
```

### Placement des hooks

| Méthode | Hooks déclenchés |
|---|---|
| `StreamCompletion` | Oui |
| `GenerateCompletion` | Non (délègue au chat agent actif) |
| `GenerateStreamCompletion` | Non (délègue au chat agent actif) |
| `GenerateCompletionWithReasoning` | Non (délègue au chat agent actif) |
| `GenerateStreamCompletionWithReasoning` | Non (délègue au chat agent actif) |

Les hooks sont dans `StreamCompletion`, la méthode du pipeline complet. Les méthodes `Generate*` délèguent directement au `chat.Agent` actif et ne déclenchent pas les hooks du crew.

### Exemple complet

```go
callCount := 0

crewAgent, _ := crew.NewAgent(ctx,
    crew.WithSingleAgent(chatAgent),
    crew.BeforeCompletion(func(a *crew.CrewAgent) {
        callCount++
        fmt.Printf("[AVANT] Appel #%d - Agent : %s\n", callCount, a.GetName())
    }),
    crew.AfterCompletion(func(a *crew.CrewAgent) {
        fmt.Printf("[APRÈS] Appel #%d - Agent : %s\n", callCount, a.GetName())
    }),
)
```

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Le code existant sans hooks continue de fonctionner sans aucune modification.

---

## 11. Méthodes de complétion directes

Le crew agent expose des méthodes de complétion directes qui délèguent au `chat.Agent` actuellement actif :

```go
// Sans streaming
result, err := crewAgent.GenerateCompletion(userMessages)

// Avec streaming
result, err := crewAgent.GenerateStreamCompletion(userMessages, callback)

// Avec raisonnement
result, err := crewAgent.GenerateCompletionWithReasoning(userMessages)
result, err := crewAgent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCb, responseCb)
```

**Note :** Ces méthodes contournent le pipeline complet (pas de compression, pas d'outils, pas de RAG, pas de routage). Elles délèguent directement au chat agent actif. Les hooks de cycle de vie ne sont **pas** déclenchés.

---

## 12. Gestion de la conversation

Toutes les méthodes de conversation opèrent sur l'agent chat **actuellement actif** :

```go
// Récupérer les messages
msgs := crewAgent.GetMessages()

// Récupérer la taille du contexte
tokens := crewAgent.GetContextSize()

// Réinitialiser la conversation
crewAgent.ResetMessages()

// Ajouter un message
crewAgent.AddMessage(roles.User, "Bonjour")

// Exporter en JSON
jsonStr, err := crewAgent.ExportMessagesToJSON()

// Arrêter le streaming
crewAgent.StopStream()
```

---

## 13. Gestion du contexte

```go
ctx := crewAgent.GetContext()
crewAgent.SetContext(newCtx)
```

---

## 14. Constructeur legacy (NewSimpleAgent)

Un constructeur simplifié est disponible pour la rétrocompatibilité :

```go
crewAgent, err := crew.NewSimpleAgent(ctx, agentCrew, "generic")
```

**Note :** `NewSimpleAgent` ne supporte pas les options (outils, RAG, compresseur, orchestrateur, hooks). Utilisez `NewAgent` avec des options pour toutes les fonctionnalités.

---

## 15. Référence API

### Constructeur

```go
func NewAgent(ctx context.Context, options ...CrewAgentOption) (*CrewAgent, error)
func NewSimpleAgent(ctx context.Context, agentCrew map[string]*chat.Agent, selectedAgentId string) (*CrewAgent, error)
```

### Types

```go
type CrewAgentOption func(*CrewAgent) error
```

### Fonctions d'options

| Fonction | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Définit l'équipe et l'agent initial. |
| `WithSingleAgent(chatAgent)` | Crée une équipe à agent unique. |
| `WithMatchAgentIdToTopicFn(fn)` | Définit la fonction de correspondance sujet → agent. |
| `WithExecuteFn(fn)` | Définit la fonction d'exécution des outils. |
| `WithConfirmationPromptFn(fn)` | Définit la fonction de confirmation des outils. |
| `WithToolsAgent(toolsAgent)` | Attache un agent d'outils. |
| `WithCompressorAgent(compressorAgent)` | Attache un agent compresseur. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attache un compresseur avec limite. |
| `WithRagAgent(ragAgent)` | Attache un agent RAG. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attache un RAG avec config. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attache un orchestrateur. |
| `BeforeCompletion(fn func(*CrewAgent))` | Définit un hook avant chaque StreamCompletion. |
| `AfterCompletion(fn func(*CrewAgent))` | Définit un hook après chaque StreamCompletion. |

### Méthodes

| Méthode | Description |
|---|---|
| `StreamCompletion(question, callback) (*chat.CompletionResult, error)` | Complétion pipeline complet avec streaming. |
| `GenerateCompletion(msgs) (*chat.CompletionResult, error)` | Complétion directe (délègue à l'agent actif). |
| `GenerateStreamCompletion(msgs, callback) (*chat.CompletionResult, error)` | Streaming direct (délègue à l'agent actif). |
| `GenerateCompletionWithReasoning(msgs) (*chat.ReasoningResult, error)` | Complétion directe avec raisonnement. |
| `GenerateStreamCompletionWithReasoning(msgs, reasoningCb, responseCb) (*chat.ReasoningResult, error)` | Streaming direct avec raisonnement. |
| `StopStream()` | Arrête l'opération de streaming en cours. |
| `GetMessages() []messages.Message` | Récupère les messages de l'agent actif. |
| `GetContextSize() int` | Récupère la taille du contexte de l'agent actif. |
| `ResetMessages()` | Réinitialise la conversation de l'agent actif. |
| `AddMessage(role, content)` | Ajoute un message à l'agent actif. |
| `ExportMessagesToJSON() (string, error)` | Exporte la conversation de l'agent actif. |
| `GetChatAgents() map[string]*chat.Agent` | Récupère tous les agents de l'équipe. |
| `SetChatAgents(crew)` | Remplace toute l'équipe. |
| `AddChatAgentToCrew(id, agent) error` | Ajoute un agent à l'équipe. |
| `RemoveChatAgentFromCrew(id) error` | Supprime un agent de l'équipe. |
| `GetSelectedAgentId() string` | Récupère l'ID de l'agent actif. |
| `SetSelectedAgentId(id) error` | Change l'agent actif. |
| `DetectTopicThenGetAgentId(query) (string, error)` | Détecte le sujet et retourne l'ID de l'agent correspondant. |
| `SetOrchestratorAgent(orchestratorAgent)` | Définit l'orchestrateur. |
| `GetOrchestratorAgent() OrchestratorAgent` | Retourne l'orchestrateur. |
| `SetToolsAgent(toolsAgent)` | Définit l'agent d'outils. |
| `GetToolsAgent() *tools.Agent` | Retourne l'agent d'outils. |
| `SetExecuteFunction(fn)` | Définit la fonction d'exécution. |
| `SetConfirmationPromptFunction(fn)` | Définit la fonction de confirmation. |
| `SetRagAgent(ragAgent)` | Définit l'agent RAG. |
| `GetRagAgent() *rag.Agent` | Retourne l'agent RAG. |
| `SetSimilarityLimit(limit)` | Définit le seuil de similarité. |
| `GetSimilarityLimit() float64` | Retourne le seuil de similarité. |
| `SetMaxSimilarities(n)` | Définit le nombre max de similarités. |
| `GetMaxSimilarities() int` | Retourne le nombre max de similarités. |
| `SetCompressorAgent(compressorAgent)` | Définit l'agent compresseur. |
| `GetCompressorAgent() *compressor.Agent` | Retourne l'agent compresseur. |
| `SetContextSizeLimit(limit)` | Définit la limite de taille de contexte. |
| `GetContextSizeLimit() int` | Retourne la limite de taille de contexte. |
| `CompressChatAgentContextIfOverLimit() (int, error)` | Compresse si au-dessus de la limite. |
| `CompressChatAgentContext() (int, error)` | Force la compression. |
| `Kind() agents.Kind` | Retourne `agents.Composite`. |
| `GetName() string` | Retourne le nom de l'agent actif. |
| `GetModelID() string` | Retourne l'ID du modèle de l'agent actif. |
| `GetContext() context.Context` | Retourne le contexte de l'agent. |
| `SetContext(ctx)` | Définit le contexte de l'agent. |
