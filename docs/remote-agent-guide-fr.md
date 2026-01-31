# Guide du Remote Agent

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent](#3-configuration-de-lagent)
4. [Santé du serveur et modèles](#4-santé-du-serveur-et-modèles)
5. [Génération de complétions](#5-génération-de-complétions)
6. [Complétions en streaming](#6-complétions-en-streaming)
7. [Complétions avec raisonnement](#7-complétions-avec-raisonnement)
8. [Historique de conversation et messages](#8-historique-de-conversation-et-messages)
9. [Opérations d'appel d'outils](#9-opérations-dappel-doutils)
10. [Hooks de cycle de vie (RemoteAgentOption)](#10-hooks-de-cycle-de-vie-remoteagentoption)
11. [Gestion du contexte](#11-gestion-du-contexte)
12. [Export JSON](#12-export-json)
13. [Référence API](#13-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Remote Agent ?

Le `remote.Agent` est un agent spécialisé fourni par le SDK Nova (`github.com/snipwise/nova`) qui communique avec un serveur Nova via HTTP. Au lieu d'appeler le LLM directement, il envoie des requêtes à un serveur distant qui exécute l'agent LLM, et retransmet les réponses via des Server-Sent Events (SSE).

C'est utile pour les architectures client-serveur où le LLM tourne sur un serveur dédié et où plusieurs clients s'y connectent.

### Quand utiliser un Remote Agent

| Scénario | Agent recommandé |
|---|---|
| Architecture client-serveur avec LLM partagé | `remote.Agent` |
| Frontend web/mobile connecté à un backend LLM | `remote.Agent` |
| Appels d'outils avec validation côté serveur | `remote.Agent` |
| Accès direct au LLM local | `chat.Agent`, `tools.Agent`, etc. |

### Capacités principales

- **Communication HTTP** : Se connecte à un serveur Nova via REST/SSE.
- **Support du streaming** : Reçoit les réponses sous forme de Server-Sent Events pour un affichage en temps réel.
- **Appels d'outils côté serveur** : Supporte la détection d'appels d'outils avec validation/annulation.
- **Vérification de santé** : Vérifie la disponibilité du serveur avant d'envoyer des requêtes.
- **Découverte des modèles** : Interroge les modèles utilisés par le serveur.
- **Hooks de cycle de vie** : Exécute une logique personnalisée avant et après chaque complétion.

---

## 2. Démarrage rapide

### Exemple minimal

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/remote"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

func main() {
    ctx := context.Background()

    agent, err := remote.NewAgent(
        ctx,
        agents.Config{
            Name: "Remote Client",
        },
        "http://localhost:8080",
    )
    if err != nil {
        panic(err)
    }

    if !agent.IsHealthy() {
        fmt.Println("Serveur non disponible")
        return
    }

    result, err := agent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: "Quelle est la capitale de la France ?"},
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("Réponse :", result.Response)
}
```

---

## 3. Configuration de l'agent

```go
agents.Config{
    Name: "Remote Client",    // Nom de l'agent (optionnel)
}
```

Le remote agent nécessite un paramètre `baseURL` (l'adresse du serveur) au lieu d'un `EngineURL`. Il n'a pas besoin de `SystemInstructions` ni de configuration de modèle puisque celles-ci sont gérées côté serveur.

```go
agent, err := remote.NewAgent(
    ctx,
    agents.Config{Name: "Mon Client"},
    "http://localhost:8080",       // URL du serveur (requis)
)
```

---

## 4. Santé du serveur et modèles

### Vérification de santé

```go
if agent.IsHealthy() {
    fmt.Println("Le serveur est en bonne santé")
}

// Ou obtenir un statut détaillé
health, err := agent.GetHealth()
fmt.Println(health.Status) // "ok"
```

### Informations sur les modèles

```go
modelsInfo, err := agent.GetModelsInfo()
fmt.Println("Modèle de chat :", modelsInfo.ChatModel)
fmt.Println("Modèle d'embeddings :", modelsInfo.EmbeddingsModel)
fmt.Println("Modèle d'outils :", modelsInfo.ToolsModel)

// Ou juste le modèle de chat
modelID := agent.GetModelID()
```

---

## 5. Génération de complétions

### GenerateCompletion

Envoie des messages et obtient la réponse complète :

```go
result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Combien font 2 + 2 ?"},
})
if err != nil {
    // gérer l'erreur
}

fmt.Println(result.Response)     // "4"
fmt.Println(result.FinishReason) // "stop"
```

**Note :** Cette méthode utilise le streaming en interne et collecte la réponse complète.

---

## 6. Complétions en streaming

### GenerateStreamCompletion

Diffuse la réponse morceau par morceau :

```go
result, err := agent.GenerateStreamCompletion(
    []messages.Message{
        {Role: roles.User, Content: "Raconte-moi une histoire."},
    },
    func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        return nil
    },
)
```

### Arrêter un stream

```go
agent.StopStream()
```

---

## 7. Complétions avec raisonnement

### GenerateCompletionWithReasoning

```go
result, err := agent.GenerateCompletionWithReasoning(userMessages)
fmt.Println(result.Response)
fmt.Println(result.Reasoning)   // Actuellement vide (pas encore supporté par le serveur)
fmt.Println(result.FinishReason)
```

### GenerateStreamCompletionWithReasoning

```go
result, err := agent.GenerateStreamCompletionWithReasoning(
    userMessages,
    reasoningCallback,  // Actuellement non utilisé
    responseCallback,
)
```

**Note :** Le raisonnement n'est pas encore supporté par l'API du serveur. Ces méthodes délèguent aux méthodes de complétion standard.

---

## 8. Historique de conversation et messages

### Récupérer les messages du serveur

```go
msgs := agent.GetMessages()
for _, msg := range msgs {
    fmt.Printf("[%s] %s\n", msg.Role, msg.Content)
}
```

### Obtenir la taille du contexte

```go
tokens := agent.GetContextSize()
fmt.Printf("Taille du contexte : %d tokens\n", tokens)
```

### Réinitialiser la conversation

```go
agent.ResetMessages()
```

### Exporter en JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
```

**Note :** Les messages sont gérés côté serveur. `AddMessage` et `AddMessages` sont des no-ops pour le remote agent.

---

## 9. Opérations d'appel d'outils

Lorsque le serveur détecte des appels d'outils, il envoie des notifications via SSE. Vous pouvez définir un callback pour les gérer et valider/annuler des opérations.

### Définir le callback d'appel d'outils

```go
agent.SetToolCallCallback(func(operationID string, message string) error {
    fmt.Printf("Appel d'outil : %s (opération : %s)\n", message, operationID)
    return nil
})
```

### Valider une opération

```go
err := agent.ValidateOperation(operationID)
```

### Annuler une opération

```go
err := agent.CancelOperation(operationID)
```

### Réinitialiser toutes les opérations

```go
err := agent.ResetOperations()
```

---

## 10. Hooks de cycle de vie (RemoteAgentOption)

Les hooks de cycle de vie permettent d'exécuter une logique personnalisée avant et après chaque complétion. Ils sont configurés comme des options fonctionnelles lors de la création de l'agent.

### RemoteAgentOption

```go
type RemoteAgentOption func(*Agent)
```

Les options sont passées en arguments variadiques à `NewAgent` :

```go
agent, err := remote.NewAgent(ctx, agentConfig, baseURL,
    remote.BeforeCompletion(fn),
    remote.AfterCompletion(fn),
)
```

### BeforeCompletion

Appelé avant chaque complétion. Le hook reçoit une référence vers l'agent.

```go
remote.BeforeCompletion(func(a *remote.Agent) {
    fmt.Printf("[AVANT] Agent : %s\n", a.GetName())
})
```

### AfterCompletion

Appelé après chaque complétion. Le hook reçoit une référence vers l'agent.

```go
remote.AfterCompletion(func(a *remote.Agent) {
    fmt.Printf("[APRÈS] Agent : %s\n", a.GetName())
})
```

### Placement des hooks

Les hooks sont dans `GenerateStreamCompletion`, qui est la méthode de base à laquelle toutes les autres méthodes de complétion délèguent :

| Méthode | Hooks déclenchés |
|---|---|
| `GenerateStreamCompletion` | Oui (directement) |
| `GenerateCompletion` | Oui (via `GenerateStreamCompletion`) |
| `GenerateCompletionWithReasoning` | Oui (via `GenerateCompletion` -> `GenerateStreamCompletion`) |
| `GenerateStreamCompletionWithReasoning` | Oui (via `GenerateStreamCompletion`) |

Cela garantit exactement un before/after hook par appel, quelle que soit la méthode utilisée.

### Exemple complet

```go
callCount := 0

agent, err := remote.NewAgent(
    ctx,
    agents.Config{Name: "Remote Client"},
    "http://localhost:8080",
    remote.BeforeCompletion(func(a *remote.Agent) {
        callCount++
        fmt.Printf("[AVANT] Appel #%d\n", callCount)
    }),
    remote.AfterCompletion(func(a *remote.Agent) {
        fmt.Printf("[APRÈS] Appel #%d\n", callCount)
    }),
)
```

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Le code existant sans hooks continue de fonctionner sans aucune modification.

---

## 11. Gestion du contexte

### Obtenir et définir le contexte

```go
ctx := agent.GetContext()
agent.SetContext(newCtx)
```

### Métadonnées de l'agent

```go
agent.Kind()       // Retourne agents.Remote
agent.GetName()    // Retourne le nom de l'agent
agent.GetModelID() // Retourne le modèle de chat du serveur
```

---

## 12. Export JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
```

---

## 13. Référence API

### Constructeur

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    baseURL string,
    opts ...RemoteAgentOption,
) (*Agent, error)
```

Crée un nouveau remote agent. Le `baseURL` est l'adresse du serveur. Le paramètre `opts` accepte zéro ou plusieurs options fonctionnelles `RemoteAgentOption`.

---

### Types

```go
type CompletionResult struct {
    Response     string
    FinishReason string
}

type ReasoningResult struct {
    Response     string
    Reasoning    string
    FinishReason string
}

type StreamCallback func(chunk string, finishReason string) error
type ToolCallCallback func(operationID string, message string) error

type ModelsInfo struct {
    Status          string
    ChatModel       string
    EmbeddingsModel string
    ToolsModel      string
}

type HealthStatus struct {
    Status string
}

type RemoteAgentOption func(*Agent)
```

---

### Fonctions d'options

| Fonction | Type | Description |
|---|---|---|
| `BeforeCompletion(fn func(*Agent))` | `RemoteAgentOption` | Définit un hook appelé avant chaque complétion. |
| `AfterCompletion(fn func(*Agent))` | `RemoteAgentOption` | Définit un hook appelé après chaque complétion. |

---

### Méthodes

| Méthode | Description |
|---|---|
| `GenerateCompletion(msgs) (*CompletionResult, error)` | Envoie des messages et obtient la réponse complète. |
| `GenerateStreamCompletion(msgs, callback) (*CompletionResult, error)` | Diffuse la réponse via callback. |
| `GenerateCompletionWithReasoning(msgs) (*ReasoningResult, error)` | Complétion avec raisonnement (pas encore supporté côté serveur). |
| `GenerateStreamCompletionWithReasoning(msgs, reasoningCb, responseCb) (*ReasoningResult, error)` | Streaming avec raisonnement. |
| `StopStream()` | Arrête l'opération de streaming en cours. |
| `SetToolCallCallback(callback)` | Définit le callback pour les notifications d'appels d'outils. |
| `ValidateOperation(operationID) error` | Valide un appel d'outil en attente. |
| `CancelOperation(operationID) error` | Annule un appel d'outil en attente. |
| `ResetOperations() error` | Annule toutes les opérations en attente. |
| `GetMessages() []messages.Message` | Récupère les messages de conversation du serveur. |
| `GetContextSize() int` | Récupère la taille du contexte en tokens du serveur. |
| `ResetMessages()` | Réinitialise la conversation sur le serveur. |
| `ExportMessagesToJSON() (string, error)` | Exporte la conversation en JSON. |
| `IsHealthy() bool` | Vérifie si le serveur est en bonne santé. |
| `GetHealth() (*HealthStatus, error)` | Obtient le statut de santé détaillé. |
| `GetModelsInfo() (*ModelsInfo, error)` | Obtient les informations sur les modèles du serveur. |
| `GetContext() context.Context` | Retourne le contexte de l'agent. |
| `SetContext(ctx)` | Met à jour le contexte de l'agent. |
| `Kind() agents.Kind` | Retourne `agents.Remote`. |
| `GetName() string` | Retourne le nom de l'agent. |
| `GetModelID() string` | Retourne le modèle de chat du serveur. |
