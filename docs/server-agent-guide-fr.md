# Guide du Server Agent

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent](#3-configuration-de-lagent)
4. [Options du serveur](#4-options-du-serveur)
5. [Démarrage du serveur](#5-démarrage-du-serveur)
6. [Pipeline de complétion](#6-pipeline-de-complétion)
7. [Complétion CLI (StreamCompletion)](#7-complétion-cli-streamcompletion)
8. [Méthodes de complétion directes](#8-méthodes-de-complétion-directes)
9. [Intégration des outils](#9-intégration-des-outils)
10. [Intégration RAG](#10-intégration-rag)
11. [Compression du contexte](#11-compression-du-contexte)
12. [Hooks de cycle de vie (BeforeCompletion / AfterCompletion)](#12-hooks-de-cycle-de-vie-beforecompletion--aftercompletion)
13. [Gestion de la conversation](#13-gestion-de-la-conversation)
14. [Référence API](#14-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Server Agent ?

Le `server.ServerAgent` est un agent de haut niveau fourni par le SDK Nova (`github.com/snipwise/nova`) qui encapsule un `chat.Agent` et l'expose comme serveur HTTP avec streaming SSE. Il orchestre les appels d'outils, l'injection de contexte RAG et la compression du contexte dans un seul pipeline.

### Quand utiliser un Server Agent

| Scénario | Agent recommandé |
|---|---|
| Serveur HTTP exposant un LLM via REST/SSE | `server.ServerAgent` |
| LLM avec outils, RAG et compression via HTTP | `server.ServerAgent` |
| Utilisation CLI avec le même pipeline (outils + RAG + compression) | `server.ServerAgent` (via `StreamCompletion`) |
| Accès direct simple au LLM | `chat.Agent`, `tools.Agent`, etc. |

### Capacités principales

- **Serveur HTTP avec streaming SSE** : Sert les complétions via `POST /completion` avec des Server-Sent Events.
- **Pipeline complet** : Compression, appels d'outils, injection de contexte RAG et complétion en streaming.
- **Double mode** : Fonctionne comme serveur HTTP (`StartServer`) et comme bibliothèque CLI (`StreamCompletion`).
- **Notifications d'appels d'outils** : Envoie des notifications via SSE pour le human-in-the-loop web.
- **Hooks de cycle de vie** : Exécute une logique personnalisée avant et après chaque complétion.
- **Pattern d'options fonctionnelles** : Configurable via des fonctions `ServerAgentOption`.

---

## 2. Démarrage rapide

### Exemple minimal

```go
package main

import (
    "context"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/server"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := server.NewAgent(
        ctx,
        agents.Config{
            Name:               "Mon Serveur",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "Tu es un assistant IA utile.",
        },
        models.Config{
            Name:        "my-model",
            Temperature: models.Float64(0.4),
        },
        server.WithPort(3500),
    )
    if err != nil {
        panic(err)
    }

    log.Fatal(agent.StartServer())
}
```

---

## 3. Configuration de l'agent

Le server agent nécessite un `agents.Config` et un `models.Config` :

```go
agents.Config{
    Name:               "Mon Serveur",          // Nom de l'agent
    EngineURL:          "http://localhost:...",  // URL du moteur LLM
    SystemInstructions: "Tu es...",             // Prompt système
}

models.Config{
    Name:        "model-name",        // Identifiant du modèle
    Temperature: models.Float64(0.4), // Température
}
```

---

## 4. Options du serveur

Les options sont passées en arguments variadiques à `NewAgent` :

```go
agent, err := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithPort(3500),
    server.WithToolsAgent(toolsAgent),
    server.WithExecuteFn(executeFn),
    server.WithRagAgent(ragAgent),
    server.WithCompressorAgent(compressorAgent),
    server.BeforeCompletion(beforeFn),
    server.AfterCompletion(afterFn),
)
```

| Option | Description |
|---|---|
| `WithPort(port int)` | Définit le port HTTP du serveur (défaut : 8080). |
| `WithExecuteFn(fn)` | Définit la fonction d'exécution des outils. |
| `WithConfirmationPromptFn(fn)` | Définit la fonction de confirmation pour le human-in-the-loop. |
| `WithToolsAgent(toolsAgent)` | Attache un agent d'outils pour les appels de fonctions. |
| `WithCompressorAgent(compressorAgent)` | Attache un agent compresseur pour la compression du contexte. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attache un agent compresseur avec une limite de taille de contexte. |
| `WithRagAgent(ragAgent)` | Attache un agent RAG pour la recherche de documents. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attache un agent RAG avec configuration de similarité. |
| `BeforeCompletion(fn)` | Définit un hook appelé avant chaque complétion. |
| `AfterCompletion(fn)` | Définit un hook appelé après chaque complétion. |

---

## 5. Démarrage du serveur

```go
log.Fatal(agent.StartServer())
```

### Points d'accès HTTP

| Méthode | Chemin | Description |
|---|---|---|
| `POST` | `/completion` | Envoie une requête de complétion (réponse en streaming SSE). |
| `POST` | `/completion/stop` | Arrête l'opération de streaming en cours. |
| `POST` | `/memory/reset` | Réinitialise l'historique de conversation. |
| `GET` | `/memory/messages/list` | Récupère les messages de conversation. |
| `GET` | `/memory/messages/context-size` | Récupère la taille du contexte en tokens. |
| `POST` | `/operation/validate` | Valide une opération d'appel d'outil en attente. |
| `POST` | `/operation/cancel` | Annule une opération d'appel d'outil en attente. |
| `POST` | `/operation/reset` | Réinitialise toutes les opérations en attente. |
| `GET` | `/models` | Informations sur les modèles. |
| `GET` | `/health` | Vérification de santé. |

---

## 6. Pipeline de complétion

Le handler `POST /completion` (`handleCompletion`) suit ce pipeline :

1. **Hook BeforeCompletion** (si défini)
2. **Compression du contexte** si l'agent compresseur est configuré et que le contexte dépasse la limite
3. **Analyse de la requête** (extraction de la question)
4. **Configuration du streaming SSE**
5. **Détection et exécution des appels d'outils** (si l'agent d'outils est configuré)
6. **Injection du contexte RAG** (si l'agent RAG est configuré)
7. **Génération de la complétion en streaming**
8. **Nettoyage de l'état des outils**
9. **Hook AfterCompletion** (si défini)

---

## 7. Complétion CLI (StreamCompletion)

La méthode `StreamCompletion` fournit le même pipeline pour une utilisation CLI :

```go
result, err := agent.StreamCompletion(
    "Combien font 2 + 2 ?",
    func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        return nil
    },
)
```

Le pipeline CLI reflète le pipeline HTTP :

1. **Hook BeforeCompletion** (si défini)
2. **Compression du contexte** si nécessaire
3. **Détection et exécution des appels d'outils**
4. **Injection du contexte RAG**
5. **Génération de la complétion en streaming**
6. **Hook AfterCompletion** (si défini)

---

## 8. Méthodes de complétion directes

Le server agent expose aussi des méthodes de complétion directes qui délèguent au `chat.Agent` interne :

```go
// Sans streaming
result, err := agent.GenerateCompletion(userMessages)

// Avec streaming
result, err := agent.GenerateStreamCompletion(userMessages, callback)

// Avec raisonnement
result, err := agent.GenerateCompletionWithReasoning(userMessages)
result, err := agent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCb, responseCb)
```

**Note :** Ces méthodes contournent le pipeline complet (pas de compression, pas d'outils, pas de RAG). Elles délèguent directement au `chat.Agent` sous-jacent. Les hooks de cycle de vie ne sont **pas** déclenchés par ces méthodes.

---

## 9. Intégration des outils

```go
toolsAgent, _ := tools.NewAgent(ctx, toolsConfig, toolsModelConfig,
    tools.WithTools(myTools),
)

agent, _ := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithToolsAgent(toolsAgent),
    server.WithExecuteFn(func(name string, args string) (string, error) {
        // Exécuter l'outil et retourner le résultat
        return `{"result": "ok"}`, nil
    }),
)
```

---

## 10. Intégration RAG

```go
ragAgent, _ := rag.NewAgent(ctx, ragConfig, ragModelConfig)

agent, _ := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithRagAgentAndSimilarityConfig(ragAgent, 0.3, 3),
)
```

---

## 11. Compression du contexte

```go
compressorAgent, _ := compressor.NewAgent(ctx, compressorConfig, compressorModelConfig)

agent, _ := server.NewAgent(ctx, agentConfig, modelConfig,
    server.WithCompressorAgentAndContextSize(compressorAgent, 4096),
)
```

---

## 12. Hooks de cycle de vie (BeforeCompletion / AfterCompletion)

Les hooks de cycle de vie permettent d'exécuter une logique personnalisée avant et après chaque complétion. Ils sont configurés comme des options fonctionnelles `ServerAgentOption`.

### BeforeCompletion

Appelé avant chaque complétion (HTTP `handleCompletion` et CLI `StreamCompletion`). Le hook reçoit une référence vers le server agent.

```go
server.BeforeCompletion(func(a *server.ServerAgent) {
    fmt.Printf("[AVANT] Agent : %s\n", a.GetName())
})
```

### AfterCompletion

Appelé après chaque complétion. Le hook reçoit une référence vers le server agent.

```go
server.AfterCompletion(func(a *server.ServerAgent) {
    fmt.Printf("[APRÈS] Agent : %s\n", a.GetName())
})
```

### Placement des hooks

| Méthode | Hooks déclenchés |
|---|---|
| `handleCompletion` (HTTP POST /completion) | Oui |
| `StreamCompletion` (CLI) | Oui |
| `GenerateCompletion` | Non (délègue au chat agent) |
| `GenerateStreamCompletion` | Non (délègue au chat agent) |
| `GenerateCompletionWithReasoning` | Non (délègue au chat agent) |
| `GenerateStreamCompletionWithReasoning` | Non (délègue au chat agent) |

Les hooks sont dans les méthodes du pipeline complet (`handleCompletion` et `StreamCompletion`) qui orchestrent la compression, les appels d'outils, le RAG et la complétion. Les méthodes `Generate*` délèguent directement au chat agent interne et ne déclenchent pas les hooks du serveur.

### Exemple complet

```go
callCount := 0

agent, err := server.NewAgent(
    ctx,
    agents.Config{
        Name:               "Mon Serveur",
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "Tu es un assistant utile.",
    },
    models.Config{
        Name:        "my-model",
        Temperature: models.Float64(0.4),
    },
    server.WithPort(3500),
    server.BeforeCompletion(func(a *server.ServerAgent) {
        callCount++
        fmt.Printf("[AVANT] Appel #%d\n", callCount)
    }),
    server.AfterCompletion(func(a *server.ServerAgent) {
        fmt.Printf("[APRÈS] Appel #%d\n", callCount)
    }),
)
```

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Le code existant sans hooks continue de fonctionner sans aucune modification.

---

## 13. Gestion de la conversation

```go
// Récupérer les messages
msgs := agent.GetMessages()

// Récupérer la taille du contexte
tokens := agent.GetContextSize()

// Réinitialiser la conversation
agent.ResetMessages()

// Ajouter un message
agent.AddMessage(roles.User, "Bonjour")

// Exporter en JSON
jsonStr, err := agent.ExportMessagesToJSON()

// Arrêter le streaming
agent.StopStream()
```

---

## 14. Référence API

### Constructeur

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    options ...ServerAgentOption,
) (*ServerAgent, error)
```

### Types

```go
type ServerAgentOption func(*ServerAgent) error
```

### Fonctions d'options

| Fonction | Description |
|---|---|
| `WithPort(port int)` | Définit le port HTTP du serveur. |
| `WithExecuteFn(fn)` | Définit la fonction d'exécution des outils. |
| `WithConfirmationPromptFn(fn)` | Définit la fonction de confirmation. |
| `WithToolsAgent(toolsAgent)` | Attache un agent d'outils. |
| `WithCompressorAgent(compressorAgent)` | Attache un agent compresseur. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attache un agent compresseur avec limite de contexte. |
| `WithRagAgent(ragAgent)` | Attache un agent RAG. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attache un agent RAG avec config de similarité. |
| `BeforeCompletion(fn func(*ServerAgent))` | Définit un hook appelé avant chaque complétion. |
| `AfterCompletion(fn func(*ServerAgent))` | Définit un hook appelé après chaque complétion. |

### Méthodes

| Méthode | Description |
|---|---|
| `StartServer() error` | Démarre le serveur HTTP avec toutes les routes. |
| `StreamCompletion(question, callback) (*chat.CompletionResult, error)` | Complétion pipeline complet pour utilisation CLI. |
| `GenerateCompletion(msgs) (*chat.CompletionResult, error)` | Complétion directe (délègue au chat agent). |
| `GenerateStreamCompletion(msgs, callback) (*chat.CompletionResult, error)` | Complétion streaming directe (délègue au chat agent). |
| `GenerateCompletionWithReasoning(msgs) (*chat.ReasoningResult, error)` | Complétion directe avec raisonnement. |
| `GenerateStreamCompletionWithReasoning(msgs, reasoningCb, responseCb) (*chat.ReasoningResult, error)` | Streaming direct avec raisonnement. |
| `StopStream()` | Arrête l'opération de streaming en cours. |
| `GetMessages() []messages.Message` | Récupère les messages de conversation. |
| `GetContextSize() int` | Récupère la taille du contexte en tokens. |
| `ResetMessages()` | Réinitialise l'historique de conversation. |
| `AddMessage(role, content)` | Ajoute un message à la conversation. |
| `ExportMessagesToJSON() (string, error)` | Exporte la conversation en JSON. |
| `Kind() agents.Kind` | Retourne `agents.ChatServer`. |
| `GetName() string` | Retourne le nom de l'agent. |
| `GetModelID() string` | Retourne l'identifiant du modèle. |
| `SetPort(port string)` | Définit le port HTTP. |
| `GetPort() string` | Retourne le port HTTP. |
| `SetToolsAgent(toolsAgent)` | Définit l'agent d'outils. |
| `GetToolsAgent() *tools.Agent` | Retourne l'agent d'outils. |
| `SetRagAgent(ragAgent)` | Définit l'agent RAG. |
| `GetRagAgent() *rag.Agent` | Retourne l'agent RAG. |
| `SetCompressorAgent(compressorAgent)` | Définit l'agent compresseur. |
| `GetCompressorAgent() *compressor.Agent` | Retourne l'agent compresseur. |
| `SetContextSizeLimit(limit)` | Définit la limite de taille de contexte pour la compression. |
| `GetContextSizeLimit() int` | Retourne la limite de taille de contexte. |
