# Remote Agent

## Description

Le **Remote Agent** est un client HTTP qui se connecte à un Server Agent ou Crew Server Agent distant. Il permet d'interagir avec des agents hébergés via une API REST avec support SSE (Server-Sent Events) pour le streaming en temps réel.

## Fonctionnalités

- **Client HTTP léger** : Connexion à des agents distants via HTTP/REST
- **Streaming SSE** : Support du streaming en temps réel avec Server-Sent Events
- **Gestion des tool calls** : Validation et annulation d'appels de fonctions à distance
- **Health checks** : Vérification de la disponibilité du serveur
- **Gestion de l'historique** : Accès à l'historique de conversation géré côté serveur
- **Callbacks** : Notifications personnalisables pour les appels de fonctions

## Cas d'usage

Le Remote Agent est utilisé pour :
- **Applications client-serveur** : Frontend se connectant à un backend AI
- **Microservices distribués** : Communication entre services via HTTP
- **Interfaces utilisateur** : Web apps, applications mobiles, CLIs distants
- **Load balancing** : Connexion à plusieurs instances de serveur
- **Architecture découplée** : Séparation entre client et logique AI

## Création d'un Remote Agent

### Syntaxe de base

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/remote"
)

ctx := context.Background()

// Créer l'agent distant
agent, err := remote.NewAgent(
    ctx,
    agents.Config{
        Name: "Remote Client",
    },
    "http://localhost:8080", // URL du serveur
)
if err != nil {
    log.Fatal(err)
}
```

## Méthodes principales

### GenerateCompletion - Complétion non-streaming

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

// Envoyer une question et recevoir la réponse complète
result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "What is the capital of France?"},
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Response: %s\n", result.Response)
fmt.Printf("Finish Reason: %s\n", result.FinishReason)
```

### GenerateStreamCompletion - Complétion avec streaming

```go
// Streaming en temps réel
_, err := agent.GenerateStreamCompletion(
    []messages.Message{
        {Role: roles.User, Content: "Tell me a story."},
    },
    func(chunk string, finishReason string) error {
        if chunk != "" {
            fmt.Print(chunk) // Afficher au fur et à mesure
        }
        if finishReason == "stop" {
            fmt.Println()
        }
        return nil
    },
)
```

### Health et informations serveur

```go
// Vérifier la santé du serveur
if agent.IsHealthy() {
    fmt.Println("✅ Server is healthy")
} else {
    fmt.Println("❌ Server is not available")
    return
}

// Obtenir le statut détaillé
health, err := agent.GetHealth()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %s\n", health.Status)

// Obtenir les informations sur les modèles
modelsInfo, err := agent.GetModelsInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Chat Model: %s\n", modelsInfo.ChatModel)
fmt.Printf("Embeddings Model: %s\n", modelsInfo.EmbeddingsModel)
fmt.Printf("Tools Model: %s\n", modelsInfo.ToolsModel)
```

### Gestion de l'historique

```go
// Obtenir tous les messages de la conversation (côté serveur)
messages := agent.GetMessages()
for _, msg := range messages {
    fmt.Printf("%s: %s\n", msg.Role, msg.Content)
}

// Obtenir la taille du contexte en tokens
contextSize := agent.GetContextSize()
fmt.Printf("Context: %d tokens\n", contextSize)

// Réinitialiser la conversation
agent.ResetMessages()

// Exporter en JSON
jsonData, err := agent.ExportMessagesToJSON()
if err == nil {
    fmt.Println(jsonData)
}
```

### Gestion des tool calls

```go
// Valider un appel de fonction en attente
err := agent.ValidateOperation("operation-id-12345")
if err != nil {
    log.Fatal(err)
}

// Annuler un appel de fonction en attente
err = agent.CancelOperation("operation-id-12345")
if err != nil {
    log.Fatal(err)
}

// Réinitialiser toutes les opérations en attente
err = agent.ResetOperations()
if err != nil {
    log.Fatal(err)
}
```

### Callback pour les tool calls

```go
// Définir un callback pour les notifications de tool calls
agent.SetToolCallCallback(func(operationID string, message string) error {
    fmt.Printf("🔔 Tool call detected: %s\n", message)
    fmt.Printf("📝 Operation ID: %s\n", operationID)

    // Valider automatiquement (ou demander confirmation à l'utilisateur)
    return agent.ValidateOperation(operationID)

    // Ou annuler
    // return agent.CancelOperation(operationID)
})

// Le callback sera appelé automatiquement lors du streaming
agent.GenerateStreamCompletion(messages, streamCallback)
```

### Contrôle du streaming

```go
// Arrêter le streaming en cours
agent.StopStream()
```

### Getters

```go
// Informations de base
name := agent.GetName()
modelID := agent.GetModelID() // Récupéré depuis le serveur
kind := agent.Kind() // Retourne agents.Remote

// Contexte
ctx := agent.GetContext()
agent.SetContext(newCtx)
```

## Structures de résultat

### CompletionResult

```go
type CompletionResult struct {
    Response     string // Réponse complète
    FinishReason string // "stop", "length", etc.
}
```

### ReasoningResult

```go
type ReasoningResult struct {
    Response     string // Réponse complète
    Reasoning    string // Raisonnement (non supporté actuellement)
    FinishReason string // "stop", "length", etc.
}
```

### ModelsInfo

```go
type ModelsInfo struct {
    Status           string // "ok"
    ChatModel        string // Modèle de chat utilisé
    EmbeddingsModel  string // Modèle d'embeddings
    ToolsModel       string // Modèle pour les tools
}
```

### HealthStatus

```go
type HealthStatus struct {
    Status string // "ok" si sain
}
```

## Exemple complet

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/remote"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

func main() {
    ctx := context.Background()

    // Créer le client distant
    agent, err := remote.NewAgent(
        ctx,
        agents.Config{
            Name: "Remote Client",
        },
        "http://localhost:8080",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Vérifier la santé du serveur
    if !agent.IsHealthy() {
        log.Fatal("Server is not available")
    }
    fmt.Println("✅ Connected to server")

    // Obtenir les informations du serveur
    modelsInfo, _ := agent.GetModelsInfo()
    fmt.Printf("Using model: %s\n\n", modelsInfo.ChatModel)

    // Définir un callback pour les tool calls
    agent.SetToolCallCallback(func(operationID string, message string) error {
        fmt.Printf("\n🔔 Tool call: %s\n", message)
        fmt.Printf("📝 Validating operation: %s\n\n", operationID)
        return agent.ValidateOperation(operationID)
    })

    // Exemple 1: Complétion simple
    fmt.Println("=== Simple Question ===")
    result, err := agent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: "What is 2 + 2?"},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Answer: %s\n\n", result.Response)

    // Exemple 2: Streaming avec tool calls
    fmt.Println("=== Question with Tools ===")
    fmt.Print("Response: ")
    _, err = agent.GenerateStreamCompletion(
        []messages.Message{
            {Role: roles.User, Content: "Calculate 40 + 2 and say hello to Alice"},
        },
        func(chunk string, finishReason string) error {
            if chunk != "" {
                fmt.Print(chunk)
            }
            if finishReason == "stop" {
                fmt.Println()
            }
            return nil
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Afficher l'historique
    fmt.Println("\n=== Conversation History ===")
    messages := agent.GetMessages()
    for i, msg := range messages {
        fmt.Printf("[%d] %s: %s\n", i+1, msg.Role, msg.Content)
    }

    // Afficher la taille du contexte
    fmt.Printf("\nContext size: %d tokens\n", agent.GetContextSize())

    // Réinitialiser la conversation
    agent.ResetMessages()
    fmt.Println("Conversation reset")
}
```

## Endpoints HTTP utilisés

Le Remote Agent communique avec les endpoints suivants du serveur :

### Complétion
- `POST /completion` - Streaming SSE de la complétion

### Mémoire
- `GET /memory/messages/list` - Liste des messages
- `GET /memory/messages/context-size` - Taille du contexte
- `POST /memory/reset` - Réinitialiser l'historique

### Opérations (Tool calls)
- `POST /operation/validate` - Valider un appel de fonction
- `POST /operation/cancel` - Annuler un appel de fonction
- `POST /operation/reset` - Réinitialiser toutes les opérations

### Informations
- `GET /models` - Informations sur les modèles
- `GET /health` - Vérification de santé

### Contrôle
- `POST /completion/stop` - Arrêter le streaming

## Format SSE (Server-Sent Events)

Le Remote Agent parse les événements SSE dans ce format :

```
data: {"message": "chunk de texte", "finish_reason": ""}
data: {"message": "", "finish_reason": "stop"}
data: {"kind": "tool_call", "operation_id": "abc123", "message": "Tool detected"}
data: {"error": "message d'erreur"}
```

## Notes importantes

- **Kind** : Retourne `agents.Remote`
- **Historique local** : Le Remote Agent ne maintient PAS d'historique local
  - `AddMessage()` et `AddMessages()` sont des no-ops
  - L'historique est géré entièrement côté serveur
  - `GetMessages()` récupère l'historique depuis le serveur
- **Streaming** : Utilise Server-Sent Events (SSE) pour le temps réel
- **Tool calls** : Nécessitent une validation manuelle ou via callback
- **Connexion** : Utilise un client HTTP standard
- **Timeouts** : Géré par le contexte Go standard

## Recommandations

### Bonnes pratiques

1. **Health checks** : Toujours vérifier `IsHealthy()` avant utilisation
2. **Gestion d'erreurs** : Vérifier les erreurs réseau et serveur
3. **Callbacks** : Utiliser `SetToolCallCallback` pour gérer les tool calls automatiquement
4. **Contexte** : Utiliser un contexte avec timeout pour éviter les blocages
5. **Reconnexion** : Implémenter une logique de retry en cas de perte de connexion

### Exemple avec timeout

```go
import "time"

// Créer un contexte avec timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

agent, err := remote.NewAgent(ctx, agentConfig, serverURL)
if err != nil {
    log.Fatal(err)
}
```

### Exemple avec retry

```go
func connectWithRetry(baseURL string, maxRetries int) (*remote.Agent, error) {
    for i := 0; i < maxRetries; i++ {
        agent, err := remote.NewAgent(context.Background(), agents.Config{
            Name: "Client",
        }, baseURL)

        if err == nil && agent.IsHealthy() {
            return agent, nil
        }

        fmt.Printf("Attempt %d failed, retrying...\n", i+1)
        time.Sleep(2 * time.Second)
    }
    return nil, fmt.Errorf("failed to connect after %d attempts", maxRetries)
}
```

### Validation automatique des tool calls

```go
// Auto-valider tous les tool calls
agent.SetToolCallCallback(func(operationID string, message string) error {
    fmt.Printf("Auto-validating: %s\n", message)
    return agent.ValidateOperation(operationID)
})
```

### Validation manuelle avec confirmation

```go
// Demander confirmation utilisateur
agent.SetToolCallCallback(func(operationID string, message string) error {
    fmt.Printf("Tool call: %s\n", message)
    fmt.Printf("Validate? (y/n): ")

    var response string
    fmt.Scanln(&response)

    if response == "y" {
        return agent.ValidateOperation(operationID)
    }
    return agent.CancelOperation(operationID)
})
```

## Différences avec les agents locaux

| Fonctionnalité | Remote Agent | Agents locaux |
|----------------|--------------|---------------|
| Historique | Géré côté serveur | Géré localement |
| AddMessage() | No-op | Fonctionne |
| Streaming | Via SSE | Direct |
| Tool calls | Validation requise | Exécution directe |
| Configuration | Côté serveur | Côté client |
| Latence | Réseau | Minimale |

## Troubleshooting

### Serveur non disponible

```go
if !agent.IsHealthy() {
    fmt.Println("Server is down. Check:")
    fmt.Println("1. Server is running")
    fmt.Println("2. URL is correct")
    fmt.Println("3. Firewall allows connection")
}
```

### Tool calls ne s'exécutent pas

```go
// Vérifier que le callback est défini
agent.SetToolCallCallback(func(operationID, message string) error {
    // IMPORTANT: Valider l'opération
    return agent.ValidateOperation(operationID)
})
```

### Timeout lors du streaming

```go
// Utiliser un contexte avec timeout plus long
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

agent.SetContext(ctx)
```
