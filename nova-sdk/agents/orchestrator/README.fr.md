# Orchestrator Agent

## Description

L'**Orchestrator Agent** est un agent spécialisé dans la détection de sujet/intention à partir de messages utilisateur. Il utilise un agent structuré en interne pour générer une sortie au format `agents.Intent` contenant le sujet de discussion identifié.

## Fonctionnalités

- **Détection de sujet** : Identifie le sujet principal d'une conversation
- **Détection d'intention** : Extrait l'intention de l'utilisateur à partir de messages
- **Sortie structurée** : Retourne un objet `Intent` avec le champ `TopicDiscussion`
- **Routage intelligent** : Utilisé par les Crew Agents pour router les requêtes vers l'agent approprié

## Cas d'usage

L'Orchestrator Agent est principalement utilisé pour :
- **Router les requêtes** dans un système multi-agents (Crew Agent)
- **Classifier les questions** par sujet
- **Détecter l'intention** de l'utilisateur pour déclencher des actions spécifiques

## Création d'un Orchestrator Agent

### Syntaxe de base

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

// Configuration de l'agent
agentConfig := agents.Config{
    Name: "Orchestrator",
    Instructions: `You are a topic detection assistant. Analyze user messages
and identify the main topic of discussion. Return only the topic category.`,
}

// Configuration du modèle
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    ModelID: "qwen2.5:1.5b", // Modèle rapide pour classification
}

// Créer l'agent
agent, err := orchestrator.NewAgent(ctx, agentConfig, modelConfig)
if err != nil {
    log.Fatal(err)
}
```

## Structure Intent

L'agent retourne un objet `agents.Intent` :

```go
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}
```

## Méthodes principales

### IdentifyIntent - Détection à partir de messages

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

// Identifier l'intention à partir de messages
userMessages := []messages.Message{
    {
        Role:    roles.User,
        Content: "Comment faire une pizza napolitaine ?",
    },
}

intent, finishReason, err := agent.IdentifyIntent(userMessages)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Topic:", intent.TopicDiscussion) // "cooking" ou "food"
fmt.Println("Finish reason:", finishReason)   // "stop"
```

### IdentifyTopicFromText - Détection à partir de texte

```go
// Méthode pratique pour détecter le sujet à partir d'un simple texte
topic, err := agent.IdentifyTopicFromText("Explique-moi le pattern Factory en Go")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Detected topic:", topic) // "programming" ou "coding"
```

### Gestion des messages

```go
// Ajouter un message
agent.AddMessage(roles.User, "Question...")

// Ajouter plusieurs messages
messages := []messages.Message{
    {Role: roles.User, Content: "Question 1"},
    {Role: roles.Assistant, Content: "Réponse 1"},
}
agent.AddMessages(messages)

// Récupérer tous les messages
allMessages := agent.GetMessages()

// Réinitialiser les messages
agent.ResetMessages()
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
kind := agent.Kind() // Retourne agents.Orchestrator

// Contexte
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requêtes/Réponses (debugging)
lastRequestJSON, _ := agent.GetLastRequestJSON()
lastResponseJSON, _ := agent.GetLastResponseJSON()
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
```

## Utilisation avec Crew Agent

L'Orchestrator Agent est conçu pour être utilisé avec le Crew Agent pour router automatiquement les requêtes :

```go
// Créer l'orchestrateur
orchestratorAgent, _ := orchestrator.NewAgent(ctx, orchestratorConfig, modelConfig)

// Fonction de mapping sujet → agent ID
matchAgentFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "coding", "programming", "development":
        return "coder"
    case "philosophy", "thinking", "psychology":
        return "thinker"
    case "cooking", "recipe", "food":
        return "cook"
    default:
        return "generic"
    }
}

// Créer le crew avec l'orchestrateur
crewAgent, _ := crew.NewAgent(
    ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithMatchAgentIdToTopicFn(matchAgentFn),
)

// L'orchestrateur détecte automatiquement le sujet et route vers l'agent approprié
result, _ := crewAgent.StreamCompletion("Comment faire une carbonara ?", callback)
// → L'orchestrateur détecte "cooking" → route vers l'agent "cook"
```

## Exemple complet

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Configuration avec instructions de détection de sujet
    agentConfig := agents.Config{
        Name: "TopicDetector",
        Instructions: `You are a topic classification assistant.
Analyze the user's message and identify the main topic category.

Categories:
- coding/programming: Questions about software development, code, debugging
- cooking/food: Questions about recipes, cooking techniques, ingredients
- philosophy/thinking: Questions about ideas, concepts, psychology
- generic: Everything else

Return only the topic category.`,
    }

    modelConfig := models.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        ModelID:   "qwen2.5:1.5b",
    }

    // Créer l'orchestrateur
    agent, err := orchestrator.NewAgent(ctx, agentConfig, modelConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Exemples de détection
    questions := []string{
        "Comment implémenter un singleton en Go ?",
        "Quelle est la recette de la pizza margherita ?",
        "Qu'est-ce que le libre arbitre ?",
    }

    for _, question := range questions {
        topic, err := agent.IdentifyTopicFromText(question)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        fmt.Printf("Question: %s\n", question)
        fmt.Printf("→ Topic: %s\n\n", topic)
    }

    // Sortie attendue:
    // Question: Comment implémenter un singleton en Go ?
    // → Topic: coding
    //
    // Question: Quelle est la recette de la pizza margherita ?
    // → Topic: cooking
    //
    // Question: Qu'est-ce que le libre arbitre ?
    // → Topic: philosophy
}
```

## Interface OrchestratorAgent

L'Orchestrator Agent implémente l'interface `agents.OrchestratorAgent` :

```go
type OrchestratorAgent interface {
    // IdentifyIntent sends messages and returns the identified intent
    IdentifyIntent(userMessages []messages.Message) (intent *Intent, finishReason string, err error)

    // IdentifyTopicFromText is a convenience method that takes a text string and returns the topic
    IdentifyTopicFromText(text string) (string, error)
}
```

## Architecture interne

L'Orchestrator Agent utilise en interne un `structured.Agent[agents.Intent]` qui :
1. Prend les messages utilisateur
2. Utilise le modèle LLM pour générer une sortie structurée
3. Parse la sortie JSON en objet `agents.Intent`
4. Retourne le champ `TopicDiscussion`

## Notes

- **Kind** : Retourne `agents.Orchestrator`
- **Basé sur Structured Agent** : Utilise `structured.Agent[agents.Intent]` en interne
- **Sortie structurée** : Garantit un format JSON cohérent avec le champ `topic_discussion`
- **Erreur si vide** : Retourne une erreur si `userMessages` est vide
- **Modèle recommandé** : Utilisez un modèle rapide (ex: `qwen2.5:1.5b`) pour une classification rapide
- **Instructions critiques** : Les instructions de l'agent doivent guider le modèle à identifier correctement les sujets

## Recommandations

### Instructions efficaces

```go
Instructions: `You are a topic classifier. Analyze the message and return ONE topic category.

Categories:
- coding: programming, software, code
- cooking: recipes, food, ingredients
- philosophy: ideas, concepts, thinking
- science: physics, chemistry, biology
- generic: everything else

Return only the category name.`
```

### Modèle approprié

- **Modèle rapide** : `qwen2.5:1.5b`, `lucy`, `jan-nano` pour classification rapide
- **Éviter les gros modèles** : La détection de sujet ne nécessite pas de modèles lourds

### Utilisation optimale

- **Crew routing** : Utilisez avec `crew.Agent` pour router automatiquement
- **Instructions claires** : Définissez clairement les catégories de sujets
- **Mapping explicite** : Utilisez `WithMatchAgentIdToTopicFn` pour mapper les sujets aux agents
