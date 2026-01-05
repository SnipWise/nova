# Crew Agent

## Description

Le **Crew Agent** est un agent composite qui orchestre plusieurs agents de chat (`chat.Agent`) pour répondre à des questions complexes. Il peut router intelligemment les requêtes vers l'agent le plus approprié et combiner plusieurs capacités (Tools, RAG, Compressor, Orchestrator).

## Fonctionnalités

- **Multi-agents** : Gère plusieurs agents de chat avec routage dynamique
- **Orchestration** : Détection de sujet/intention et routage automatique vers l'agent approprié
- **Tools Agent** : Exécution de fonctions (function calling) avec confirmation utilisateur
- **RAG Agent** : Recherche de similarité et enrichissement du contexte
- **Compressor Agent** : Compression automatique du contexte quand la limite est atteinte
- **Human-in-the-loop** : Validation personnalisable des appels de fonctions

## Création d'un Crew Agent

### Syntaxe avec options

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents/crew"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
)

// Création avec un seul agent
agent, err := crew.NewAgent(
    ctx,
    crew.WithSingleAgent(chatAgent),
)

// Création avec plusieurs agents
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "writer":  writerAgent,
}

agent, err := crew.NewAgent(
    ctx,
    crew.WithAgentCrew(agentCrew, "coder"), // "coder" est l'agent par défaut
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithToolsAgent(toolsAgent),
    crew.WithRagAgent(ragAgent),
    crew.WithCompressorAgentAndContextSize(compressorAgent, 8000),
    crew.WithExecuteFn(myCustomExecutor),
    crew.WithConfirmationPromptFn(myConfirmationPrompt),
)
```

### Options disponibles

| Option | Description |
|--------|-------------|
| `WithSingleAgent(chatAgent)` | Crée un crew avec un seul agent (ID: "single") |
| `WithAgentCrew(agentCrew, selectedAgentId)` | Définit plusieurs agents avec l'agent initial sélectionné |
| `WithMatchAgentIdToTopicFn(fn)` | Fonction personnalisée de mapping sujet → agent ID |
| `WithOrchestratorAgent(orchestratorAgent)` | Agent pour la détection de sujet/intention |
| `WithExecuteFn(fn)` | Fonction personnalisée d'exécution des tools |
| `WithConfirmationPromptFn(fn)` | Fonction personnalisée de confirmation des tool calls |
| `WithToolsAgent(toolsAgent)` | Ajoute un agent pour l'exécution de fonctions |
| `WithRagAgent(ragAgent)` | Ajoute un agent RAG pour la recherche de documents |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | RAG avec configuration de similarité |
| `WithCompressorAgent(compressorAgent)` | Ajoute un agent pour la compression du contexte |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Compressor avec limite de contexte |

## Méthodes principales

### Gestion du crew

```go
// Obtenir tous les agents
chatAgents := agent.GetChatAgents()

// Ajouter un agent au crew
err := agent.AddChatAgentToCrew("expert", expertAgent)

// Retirer un agent du crew (sauf l'agent actif)
err := agent.RemoveChatAgentFromCrew("expert")

// Obtenir/Définir l'agent actif
currentId := agent.GetSelectedAgentId()
err := agent.SetSelectedAgentId("thinker")
```

### Génération de complétion

```go
// Streaming avec callback
result, err := agent.StreamCompletion(question, func(chunk string, finishReason string) error {
    fmt.Print(chunk)
    return nil
})
```

### Gestion de l'orchestrateur

```go
// Définir l'agent orchestrateur
agent.SetOrchestratorAgent(orchestratorAgent)

// Détecter le sujet et obtenir l'ID de l'agent approprié
agentId, err := agent.DetectTopicThenGetAgentId("Comment cuisiner une pizza ?")
// → Retourne "cook" si la fonction matchAgentIdToTopicFn le mappe
```

### Gestion des agents auxiliaires

```go
// Tools Agent
agent.SetToolsAgent(toolsAgent)
agent.SetExecuteFunction(myExecutor)
agent.SetConfirmationPromptFunction(myConfirmationFn)

// RAG Agent
agent.SetRagAgent(ragAgent)

// Compressor Agent
agent.SetCompressorAgent(compressorAgent)
```

### Méthodes héritées de chat.Agent

```go
// Messages
agent.GetMessages()
agent.AddMessage(roles.User, "Question...")
agent.ResetMessages()

// Contexte
contextSize := agent.GetContextSize()

// Génération (délègue à l'agent actif)
agent.GenerateCompletion(messages)
agent.GenerateStreamCompletion(messages, callback)
agent.GenerateCompletionWithReasoning(messages)
agent.GenerateStreamCompletionWithReasoning(messages, reasoningCb, responseCb)

// Export
jsonData, err := agent.ExportMessagesToJSON()
```

## Pipeline de traitement (StreamCompletion)

```
1. Compression du contexte (si CompressorAgent configuré)
   ↓
2. Détection de tool calls (si ToolsAgent configuré)
   ↓
3. Demande de confirmation utilisateur (via confirmationPromptFn)
   ↓
4. Exécution des fonctions (si confirmées)
   ↓
5. Ajout du résultat au contexte
   ↓
6. Recherche RAG (si RagAgent configuré)
   ↓
7. Détection de sujet et routage (si OrchestratorAgent configuré)
   ↓
8. Génération de la réponse avec l'agent approprié
   ↓
9. Nettoyage de l'état
```

## Routage intelligent avec Orchestrator

L'orchestrateur détecte le sujet de la question et route vers l'agent approprié.

### Configuration du mapping sujet → agent

```go
// Fonction de mapping personnalisée
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

agent, err := crew.NewAgent(
    ctx,
    crew.WithAgentCrew(agentCrew, "generic"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithMatchAgentIdToTopicFn(matchAgentFn),
)
```

### Détection automatique lors de StreamCompletion

Quand `StreamCompletion` est appelé et qu'un orchestrateur est configuré :
1. L'orchestrateur détecte le sujet de la question
2. La fonction `matchAgentIdToTopicFn` mappe le sujet vers un agent ID
3. L'agent actif (`currentChatAgent`) est automatiquement basculé
4. La réponse est générée par le nouvel agent sélectionné

```go
// L'utilisateur pose une question sur la cuisine
result, err := agent.StreamCompletion("Comment faire une pizza ?", callback)
// → L'orchestrateur détecte "cooking" → route vers "cook"
```

## Exemple complet

```go
ctx := context.Background()

// Créer plusieurs agents spécialisés
coderAgent, _ := chat.NewAgent(ctx,
    agents.Config{Name: "Coder", Instructions: "Expert en programmation"},
    modelConfig,
)
thinkerAgent, _ := chat.NewAgent(ctx,
    agents.Config{Name: "Thinker", Instructions: "Expert en philosophie"},
    modelConfig,
)
cookAgent, _ := chat.NewAgent(ctx,
    agents.Config{Name: "Cook", Instructions: "Expert en cuisine"},
    modelConfig,
)

// Créer l'orchestrateur
orchestratorAgent, _ := orchestrator.NewAgent(ctx, agentConfig, modelConfig)

// Fonction de mapping sujet → agent
matchAgentFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "coding", "programming":
        return "coder"
    case "philosophy", "thinking":
        return "thinker"
    case "cooking", "food":
        return "cook"
    default:
        return "coder" // Agent par défaut
    }
}

// Créer le crew agent
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "cook":    cookAgent,
}

crewAgent, err := crew.NewAgent(
    ctx,
    crew.WithAgentCrew(agentCrew, "coder"),
    crew.WithOrchestratorAgent(orchestratorAgent),
    crew.WithMatchAgentIdToTopicFn(matchAgentFn),
    crew.WithToolsAgent(toolsAgent),
)

// Utilisation
result, err := crewAgent.StreamCompletion("Explique-moi le pattern Factory", func(chunk, reason string) error {
    fmt.Print(chunk)
    return nil
})
// → Routé vers "coder" automatiquement

result, err = crewAgent.StreamCompletion("Quelle est la recette de la carbonara ?", callback)
// → Routé vers "cook" automatiquement
```

## Gestion dynamique du crew

```go
// Ajouter un nouvel agent pendant l'exécution
expertAgent, _ := chat.NewAgent(ctx, expertConfig, modelConfig)
err := crewAgent.AddChatAgentToCrew("expert", expertAgent)

// Basculer manuellement vers un agent
err = crewAgent.SetSelectedAgentId("expert")

// Retirer un agent (impossible si c'est l'agent actif)
err = crewAgent.RemoveChatAgentFromCrew("expert")
```

## Notes

- **Au moins un agent requis** : `WithAgentCrew` ou `WithSingleAgent` est obligatoire
- **Agent actif** : Un seul agent est actif à la fois (`currentChatAgent`)
- **Routage automatique** : L'orchestrateur change automatiquement l'agent actif pendant `StreamCompletion`
- **Valeurs par défaut** :
  - `similarityLimit`: 0.6
  - `maxSimilarities`: 3
  - `contextSizeLimit`: 8000
- **Kind** : Retourne `agents.Composite`
- **Méthodes déléguées** : `GetName()`, `GetModelID()`, etc. sont déléguées à `currentChatAgent`

## Constructeur legacy

Un constructeur simplifié existe aussi (sans options) :

```go
agent, err := crew.NewSimpleAgent(ctx, agentCrew, "coder")
```

**Note** : Préférez `NewAgent` avec options pour plus de flexibilité.
