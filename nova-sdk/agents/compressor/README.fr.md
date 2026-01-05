# Compressor Agent

## Description

Le **Compressor Agent** est un agent spécialisé dans la compression de contexte de conversation. Il prend une liste de messages et génère un résumé concis qui préserve les informations essentielles tout en réduisant la taille du contexte.

## Fonctionnalités

- **Compression de contexte** : Résume des conversations longues en préservant les faits clés
- **Streaming** : Génération du résumé en streaming ou en une seule fois
- **Prompts personnalisables** : Plusieurs prompts de compression prédéfinis et possibilité de créer des prompts personnalisés
- **Instructions configurables** : Instructions système prédéfinies pour différents styles de compression

## Création d'un Compressor Agent

### Syntaxe de base

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

// Configuration de l'agent
agentConfig := agents.Config{
    Name: "Compressor",
    Instructions: compressor.Instructions.Minimalist,
}

// Configuration du modèle
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    ModelID: "qwen2.5:1.5b",
}

// Créer l'agent avec prompt par défaut (Minimalist)
agent, err := compressor.NewAgent(ctx, agentConfig, modelConfig)

// Créer l'agent avec un prompt personnalisé
agent, err := compressor.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    compressor.WithCompressionPrompt(compressor.Prompts.Structured),
)
```

### Options disponibles

| Option | Description |
|--------|-------------|
| `WithCompressionPrompt(prompt)` | Définit le prompt de compression à utiliser |

## Instructions système prédéfinies

Le package fournit trois instructions système prédéfinies :

### `Instructions.Minimalist` (recommandé par défaut)
```
You are a context compression assistant. Your task is to summarize
conversations concisely, preserving key facts, decisions, and context
needed for continuation.
```

### `Instructions.Expert`
Instructions détaillées avec :
- Préservation des informations critiques
- Élimination des redondances
- Maintien de la chronologie
- Format de sortie structuré
- Directives de compression spécifiques

### `Instructions.Effective`
Format structuré avec sections :
- Conversation Summary
- Key Points
- To Remember

## Prompts de compression prédéfinis

Le package fournit quatre prompts de compression :

| Prompt | Description | Cas d'usage |
|--------|-------------|-------------|
| `Prompts.Minimalist` ⭐ | Résumé concis préservant faits clés, décisions et contexte | **Recommandé** - Usage général |
| `Prompts.Structured` | Format structuré avec topics, décisions, contexte (< 200 mots) | Résumés organisés |
| `Prompts.UltraShort` | Extraction des faits, décisions et contexte essentiel uniquement | Compression maximale |
| `Prompts.ContinuityFocus` | Préserve toute l'information nécessaire pour continuer naturellement | Continuité de conversation |

**Prompt par défaut** : `Prompts.Minimalist`

## Méthodes principales

### Compression sans streaming

```go
// Compresser une liste de messages
result, err := agent.CompressContext(messagesList)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Compressed text:", result.CompressedText)
fmt.Println("Finish reason:", result.FinishReason)
```

**Retour** : `*CompressionResult`
- `CompressedText` : Le texte compressé
- `FinishReason` : La raison de fin (généralement "stop")

### Compression avec streaming

```go
// Compresser avec streaming
result, err := agent.CompressContextStream(messagesList, func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
})
if err != nil {
    log.Fatal(err)
}

fmt.Println("\nFinal compressed text:", result.CompressedText)
```

### Changer le prompt de compression

```go
// Changer le prompt après création
agent.SetCompressionPrompt(compressor.Prompts.UltraShort)

// Ou utiliser un prompt personnalisé
customPrompt := "Résume cette conversation en 3 phrases maximum."
agent.SetCompressionPrompt(customPrompt)
```

### Getters et Setters

```go
// Configuration
config := agent.GetConfig()
agent.SetConfig(newConfig)

modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)

// Informations
name := agent.GetName()
modelID := agent.GetModelID()
kind := agent.GetKind() // Retourne agents.Compressor

// Contexte
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requêtes/Réponses (debugging)
lastRequestJSON, _ := agent.GetLastRequestJSON()
lastResponseJSON, _ := agent.GetLastResponseJSON()
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
```

## Exemple complet

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/compressor"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Configuration
    agentConfig := agents.Config{
        Name:         "Compressor",
        Instructions: compressor.Instructions.Minimalist,
    }
    modelConfig := models.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        ModelID:   "qwen2.5:1.5b",
    }

    // Créer l'agent avec prompt structuré
    agent, err := compressor.NewAgent(
        ctx,
        agentConfig,
        modelConfig,
        compressor.WithCompressionPrompt(compressor.Prompts.Structured),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Messages à compresser
    messagesList := []messages.Message{
        {Role: roles.User, Content: "Bonjour, je voudrais créer une API REST."},
        {Role: roles.Assistant, Content: "Bien sûr ! Quel langage préférez-vous ?"},
        {Role: roles.User, Content: "J'aimerais utiliser Go."},
        {Role: roles.Assistant, Content: "Excellent choix. Voici comment créer une API REST en Go..."},
        // ... beaucoup plus de messages
    }

    // Compression avec streaming
    fmt.Println("🗜️  Compressing context...")
    result, err := agent.CompressContextStream(messagesList, func(chunk string, finishReason string) error {
        fmt.Print(chunk)
        if finishReason != "" {
            fmt.Printf("\n[Finish: %s]\n", finishReason)
        }
        return nil
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n✅ Compression complete. Original: %d messages → Compressed: %d chars\n",
        len(messagesList), len(result.CompressedText))
}
```

## Utilisation avec d'autres agents

Le Compressor Agent est généralement utilisé avec les agents Server, Crew ou Chat pour gérer automatiquement la compression du contexte :

```go
// Avec Server Agent
serverAgent, _ := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)

// Avec Crew Agent
crewAgent, _ := crew.NewAgent(
    ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)

// La compression se fait automatiquement quand la limite est atteinte
```

## Format de compression

Le Compressor Agent :
1. Convertit les messages en format texte :
   ```
   user: Message de l'utilisateur
   assistant: Réponse de l'assistant
   system: Message système
   ```
2. Envoie le texte avec le prompt de compression
3. Retourne le résumé généré par le modèle

## Notes

- **Kind** : Retourne `agents.Compressor`
- **Streaming** : Utilise OpenAI SDK en interne pour le streaming
- **Prompt par défaut** : `Prompts.Minimalist`
- **Instructions par défaut** : Aucune instruction par défaut - doit être définie dans `agentConfig.Instructions`
- **Erreur si vide** : Retourne une erreur si `messagesList` est vide
- **Conversion automatique** : Les messages sont automatiquement convertis au format OpenAI en interne

## Recommandations

- **Prompt recommandé** : `Prompts.Minimalist` pour la plupart des cas
- **Instructions recommandées** : `Instructions.Minimalist` pour usage général, `Instructions.Expert` pour compression avancée
- **Streaming** : Utilisez `CompressContextStream` pour voir la progression en temps réel
- **Taille de contexte** : Configurez une limite appropriée (ex: 8000 caractères) lors de l'utilisation avec Server/Crew agents
