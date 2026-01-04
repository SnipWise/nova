# Chat Agent

Agent de conversation simplifié qui masque les détails du SDK OpenAI.

## Création

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/models"
)

agent, err := chat.NewAgent(ctx, agentConfig, modelConfig)
```

## Fonctionnalités principales

### 1. Génération de réponses

#### Completion simple
```go
result, err := agent.GenerateCompletion(userMessages)
// result.Response (string)
// result.FinishReason (string)
```

#### Completion avec streaming
```go
callback := func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
}

result, err := agent.GenerateStreamCompletion(userMessages, callback)
```

#### Completion avec raisonnement
```go
result, err := agent.GenerateCompletionWithReasoning(userMessages)
// result.Response (string)
// result.Reasoning (string)
// result.FinishReason (string)
```

#### Streaming avec raisonnement
```go
reasoningCallback := func(chunk string, finishReason string) error {
    fmt.Print("[Reasoning] ", chunk)
    return nil
}

responseCallback := func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
}

result, err := agent.GenerateStreamCompletionWithReasoning(
    userMessages,
    reasoningCallback,
    responseCallback,
)
```

### 2. Gestion des messages

```go
// Ajouter un message
agent.AddMessage(roles.User, "Hello")

// Ajouter plusieurs messages
agent.AddMessages([]messages.Message{...})

// Récupérer tous les messages
msgs := agent.GetMessages()

// Supprimer les N derniers messages (sauf le système)
agent.RemoveLastNMessages(3)

// Réinitialiser (garde uniquement le message système)
agent.ResetMessages()
```

### 3. Instructions système

```go
// Définir/mettre à jour les instructions système
agent.SetSystemInstructions("You are a helpful assistant...")
```

### 4. Directives pré/post pour messages utilisateur

**Cas d'usage**: Encadrer systématiquement tous les messages utilisateur avec du contexte ou des instructions supplémentaires.

```go
// Ajouter du contexte AVANT le dernier message utilisateur
agent.SetUserMessagePreDirectives("Context: You are a technical support agent.")

// Ajouter des instructions APRÈS le dernier message utilisateur
agent.SetUserMessagePostDirectives("Always respond in French.")

// Récupérer les directives
pre := agent.GetUserMessagePreDirectives()
post := agent.GetUserMessagePostDirectives()
```

**Fonctionnement**: Les directives sont automatiquement ajoutées au dernier message utilisateur lors de chaque appel à `GenerateCompletion`, `GenerateStreamCompletion`, etc.

**Exemple**:
```go
agent.SetUserMessagePreDirectives("You are an expert in Go programming.")
agent.SetUserMessagePostDirectives("Keep your answer under 100 words.")

// Message utilisateur: "How do I use goroutines?"
// Message réel envoyé au modèle:
// "You are an expert in Go programming.\n\nHow do I use goroutines?\n\nKeep your answer under 100 words."
```

### 5. Contrôle du streaming

```go
// Interrompre le streaming en cours
agent.StopStream()
```

### 6. Contexte et métadonnées

```go
// Taille approximative du contexte actuel
size := agent.GetContextSize()

// Type de l'agent
kind := agent.Kind() // agents.Chat

// Nom de l'agent
name := agent.GetName()

// ID du modèle
modelID := agent.GetModelID()
```

### 7. Export et inspection

```go
// Export JSON de la conversation
jsonStr, err := agent.ExportMessagesToJSON()

// Dernière requête (raw JSON)
rawReq := agent.GetLastRequestRawJSON()

// Dernière réponse (raw JSON)
rawResp := agent.GetLastResponseRawJSON()

// Dernière requête (formatted JSON)
reqJSON, err := agent.GetLastRequestJSON()

// Dernière réponse (formatted JSON)
respJSON, err := agent.GetLastResponseJSON()
```

### 8. Configuration

```go
// Récupérer la configuration de l'agent
config := agent.GetConfig()

// Mettre à jour la configuration de l'agent
agent.SetConfig(newConfig)

// Récupérer la configuration du modèle
modelConfig := agent.GetModelConfig()

// Mettre à jour la configuration du modèle
agent.SetModelConfig(newModelConfig)
```

### 9. Contexte Go

```go
// Récupérer le context.Context
ctx := agent.GetContext()

// Définir un nouveau context.Context
agent.SetContext(newCtx)
```

## Types

### CompletionResult
```go
type CompletionResult struct {
    Response     string
    FinishReason string
}
```

### ReasoningResult
```go
type ReasoningResult struct {
    Response     string
    Reasoning    string
    FinishReason string
}
```

### StreamCallback
```go
type StreamCallback func(chunk string, finishReason string) error
```

## Notes importantes

- L'historique de conversation est géré automatiquement par le `BaseAgent` selon `KeepConversationHistory`
- Le message système est préservé lors de `ResetMessages()`
- Les directives pré/post sont appliquées automatiquement au dernier message utilisateur
- Le streaming peut être interrompu avec `StopStream()`
