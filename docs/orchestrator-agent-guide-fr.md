# Guide de l'Orchestrator Agent

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent](#3-configuration-de-lagent)
4. [Configuration du modèle](#4-configuration-du-modèle)
5. [Méthodes de détection d'intent](#5-méthodes-de-détection-dintent)
6. [Historique de conversation et messages](#6-historique-de-conversation-et-messages)
7. [Hooks de cycle de vie (OrchestratorAgentOption)](#7-hooks-de-cycle-de-vie-orchestratoragentoption)
8. [Utilisation avec le Crew Agent](#8-utilisation-avec-le-crew-agent)
9. [Gestion du contexte et de l'état](#9-gestion-du-contexte-et-de-létat)
10. [Débogage JSON](#10-débogage-json)
11. [Référence API](#11-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Orchestrator Agent ?

L'`orchestrator.Agent` est un agent spécialisé fourni par le SDK Nova (`github.com/snipwise/nova`) pour la détection de topics et d'intentions à partir des messages utilisateur. Il encapsule en interne un `structured.Agent[agents.Intent]` pour générer une sortie JSON structurée contenant le sujet de discussion identifié.

Contrairement à un chat agent qui génère des réponses en texte libre, l'orchestrator agent retourne toujours un objet `Intent` structuré avec un champ `TopicDiscussion`. Cela le rend idéal pour le routage, la classification et la prise de décision dans les systèmes multi-agents.

### Quand utiliser un Orchestrator Agent

| Scénario | Agent recommandé |
|---|---|
| Classification de topic/intention depuis une entrée utilisateur | `orchestrator.Agent` |
| Routage de requêtes vers des agents spécialisés dans un crew | `orchestrator.Agent` avec `CrewServerAgent` |
| Classification simple de questions | `orchestrator.Agent` avec `IdentifyTopicFromText` |
| IA conversationnelle en texte libre | `chat.Agent` |
| Appels de fonctions / utilisation d'outils | `tools.Agent` |

### Capacités principales

- **Détection de topic** : Identifie le sujet principal d'une conversation à partir des messages utilisateur.
- **Sortie structurée** : Retourne toujours un objet `Intent` avec un champ `TopicDiscussion` au format JSON.
- **Routage d'intentions** : Conçu pour fonctionner avec les Crew Agents pour le routage automatique des requêtes vers des agents spécialisés.
- **Hooks de cycle de vie** : Exécutez une logique personnalisée avant et après chaque identification d'intention.
- **Léger** : Utilise des modèles rapides et petits pour une classification rapide sans nécessiter de grands modèles de langage.

---

## 2. Démarrage rapide

### Exemple minimal

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := orchestrator.NewAgent(
        ctx,
        agents.Config{
            Name:      "détecteur-de-topic",
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: `You are good at identifying the topic of a conversation.
Given a user's input, identify the main topic of discussion in only one word.
The possible topics are: Technology, Health, Sports, Entertainment, Science.
Respond in JSON format with the field 'topic_discussion'.`,
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.0),
        },
    )
    if err != nil {
        panic(err)
    }

    topic, err := agent.IdentifyTopicFromText("Comment fonctionne l'informatique quantique ?")
    if err != nil {
        panic(err)
    }

    fmt.Println("Topic détecté :", topic)
    // Sortie : Topic détecté : Technology
}
```

---

## 3. Configuration de l'agent

Le struct `agents.Config` contrôle l'identité et le comportement de l'agent :

```go
agents.Config{
    Name:                    "orchestrator-agent",  // Nom de l'agent (optionnel)
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1", // URL du moteur LLM (requis)
    APIKey:                  "votre-clé-api",       // Clé API (optionnel)
    SystemInstructions:      "...",                 // Instructions de classification (critique)
    KeepConversationHistory: false,                 // Habituellement false pour la classification
}
```

| Champ | Type | Requis | Description |
|---|---|---|---|
| `Name` | `string` | Non | Identifiant de l'agent pour le logging et les configurations multi-agents. |
| `EngineURL` | `string` | Oui | URL du moteur LLM compatible OpenAI. |
| `APIKey` | `string` | Non | Clé API pour les moteurs authentifiés. |
| `SystemInstructions` | `string` | Critique | Instructions qui définissent les catégories de classification et le format. C'est le champ le plus important -- la qualité de la détection de topic dépend directement de ces instructions. |
| `KeepConversationHistory` | `bool` | Non | Habituellement `false` pour une classification sans état. Défaut : `false`. |

### Rédiger des instructions système efficaces

Les instructions système doivent clairement définir :
1. Les catégories de topics possibles
2. Le format de sortie attendu (JSON avec le champ `topic_discussion`)
3. Toute règle ou priorité de classification

```go
SystemInstructions: `You are a topic classification assistant.
Analyze the user's message and identify the main topic of discussion.

Categories:
- coding: programming, software development, debugging
- cooking: recipes, food, ingredients
- philosophy: ideas, concepts, thinking
- science: physics, chemistry, biology
- generic: everything else

Return only the topic category in JSON format with the field 'topic_discussion'.`
```

---

## 4. Configuration du modèle

Le struct `models.Config` contrôle les paramètres de génération du modèle :

```go
models.Config{
    Name:        "ai/qwen2.5:1.5B-F16",    // ID du modèle (requis)
    Temperature: models.Float64(0.0),        // Utiliser 0.0 pour une classification déterministe
}
```

### Paramètres recommandés pour la classification

- **Temperature** : `0.0` -- Une sortie déterministe est essentielle pour une classification cohérente.
- **Taille du modèle** : Utilisez des modèles petits et rapides (1-3B paramètres). La détection de topic ne nécessite pas de grands modèles.
- **Modèles recommandés** : `qwen2.5:1.5b`, `lucy`, `jan-nano` -- rapides et suffisants pour la classification.

---

## 5. Méthodes de détection d'intent

### IdentifyIntent

La méthode principale pour détecter les topics à partir de messages. Retourne l'objet `Intent` complet ainsi que la raison de fin.

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

intent, finishReason, err := agent.IdentifyIntent([]messages.Message{
    {Role: roles.User, Content: "Comment faire une pizza napolitaine ?"},
})
if err != nil {
    // gérer l'erreur
}

fmt.Println("Topic :", intent.TopicDiscussion) // "cooking" ou "food"
fmt.Println("Raison de fin :", finishReason)   // "stop"
```

**Type de retour :** `*agents.Intent`

```go
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}
```

### IdentifyTopicFromText

Une méthode de commodité qui prend une chaîne de texte brut et retourne uniquement le topic détecté. Elle crée en interne un message utilisateur et appelle `IdentifyIntent`.

```go
topic, err := agent.IdentifyTopicFromText("Expliquer le pattern Factory en Go")
if err != nil {
    // gérer l'erreur
}

fmt.Println("Topic détecté :", topic) // "coding"
```

### Interface OrchestratorAgent

L'orchestrator agent implémente l'interface `agents.OrchestratorAgent` :

```go
type OrchestratorAgent interface {
    IdentifyIntent(userMessages []messages.Message) (intent *Intent, finishReason string, err error)
    IdentifyTopicFromText(text string) (string, error)
}
```

Cette interface est utilisée par `CrewServerAgent` pour le routage automatique des requêtes.

---

## 6. Historique de conversation et messages

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

### Quand utiliser l'historique de conversation

Pour la plupart des cas d'utilisation de classification, `KeepConversationHistory` devrait être `false`. Chaque appel de classification est typiquement indépendant. Cependant, dans certains scénarios vous pourriez vouloir activer l'historique -- par exemple, si l'orchestrateur doit considérer les messages précédents pour classifier des requêtes ambiguës.

---

## 7. Hooks de cycle de vie (OrchestratorAgentOption)

Les hooks de cycle de vie permettent d'exécuter une logique personnalisée avant et après chaque identification d'intention. Ils sont configurés comme des options fonctionnelles lors de la création de l'agent.

### OrchestratorAgentOption

```go
type OrchestratorAgentOption func(*Agent)
```

Les options sont passées en arguments variadiques à `NewAgent` :

```go
agent, err := orchestrator.NewAgent(ctx, agentConfig, modelConfig,
    orchestrator.BeforeCompletion(fn),
    orchestrator.AfterCompletion(fn),
)
```

### BeforeCompletion

Appelé avant chaque identification d'intention (`IdentifyIntent` et, par extension, `IdentifyTopicFromText`). Le hook reçoit une référence à l'agent.

```go
orchestrator.BeforeCompletion(func(a *orchestrator.Agent) {
    fmt.Println("Identification d'intention en cours...")
    fmt.Printf("Agent : %s (%s)\n", a.GetName(), a.GetModelID())
    fmt.Printf("Nombre de messages : %d\n", len(a.GetMessages()))
})
```

**Cas d'usage :**
- Logging et monitoring
- Collecte de métriques (compter les requêtes de classification)
- Inspection de l'état pré-classification

### AfterCompletion

Appelé après chaque identification d'intention, une fois le résultat reçu. Le hook reçoit une référence à l'agent.

```go
orchestrator.AfterCompletion(func(a *orchestrator.Agent) {
    fmt.Println("Identification d'intention terminée.")
    fmt.Printf("Nombre de messages : %d\n", len(a.GetMessages()))
})
```

**Cas d'usage :**
- Logging des résultats de classification
- Métriques post-classification
- Déclenchement d'actions en aval basées sur la classification
- Audit/suivi de la distribution des topics

### Exemple complet avec hooks

```go
agent, err := orchestrator.NewAgent(
    ctx,
    agents.Config{
        Name:      "orchestrator-agent",
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: `You are good at identifying the topic of a conversation.
Given a user's input, identify the main topic of discussion in only one word.
Respond in JSON format with the field 'topic_discussion'.`,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
    },
    orchestrator.BeforeCompletion(func(a *orchestrator.Agent) {
        fmt.Printf("[AVANT] Agent : %s, Messages : %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
    orchestrator.AfterCompletion(func(a *orchestrator.Agent) {
        fmt.Printf("[APRÈS] Agent : %s, Messages : %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
)
```

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Les hooks ne sont appelés que lorsqu'ils ont été définis. Le paramètre `...OrchestratorAgentOption` est variadique, donc le code existant sans hooks continue de fonctionner sans aucune modification.

### Les hooks s'appliquent aux deux méthodes de détection

Puisque `IdentifyTopicFromText` appelle `IdentifyIntent` en interne, les hooks sont déclenchés pour les deux méthodes :

| Méthode | BeforeCompletion | AfterCompletion |
|---|---|---|
| `IdentifyIntent` | Oui | Oui |
| `IdentifyTopicFromText` | Oui (via IdentifyIntent) | Oui (via IdentifyIntent) |

---

## 8. Utilisation avec le Crew Agent

L'orchestrator agent est conçu pour fonctionner avec `CrewServerAgent` pour le routage automatique des requêtes dans les systèmes multi-agents.

### Routage basique avec crew

```go
import (
    "strings"
    "github.com/snipwise/nova/nova-sdk/agents/crewserver"
)

// Créer des chat agents spécialisés
agentCrew := map[string]*chat.Agent{
    "expert":  expertAgent,
    "coder":   coderAgent,
    "thinker": thinkerAgent,
}

// Créer l'orchestrator agent
orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:      "orchestrator-agent",
        EngineURL: engineURL,
        SystemInstructions: "Classify queries into: code_generation, complex_thinking, code_question.",
    },
    models.Config{
        Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
        Temperature: models.Float64(0.0),
    },
)

// Définir la fonction de routage
matchFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "code_generation", "write code":
        return "coder"
    case "complex_thinking", "reasoning":
        return "thinker"
    default:
        return "expert"
    }
}

// Assembler le crew server
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "expert"),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.WithMatchAgentIdToTopicFn(matchFn),
    crewserver.WithPort(3500),
)

crewServerAgent.StartServer()
```

### Comment fonctionne le routage

1. Un utilisateur envoie un message au crew server.
2. L'orchestrator agent classifie le topic du message.
3. La fonction de correspondance mappe le topic à un ID d'agent.
4. Le crew server route la requête vers l'agent correspondant.
5. L'agent correspondant génère la réponse.

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
agent.Kind()       // Retourne agents.Orchestrator
agent.GetName()    // Retourne le nom de l'agent depuis la config
agent.GetModelID() // Retourne le nom du modèle depuis la config modèle
```

---

## 10. Débogage JSON

Pour le débogage, vous pouvez accéder au JSON brut envoyé et reçu du moteur LLM :

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
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    opts ...OrchestratorAgentOption,
) (*Agent, error)
```

Crée un nouvel orchestrator agent. Le paramètre `opts` accepte zéro ou plusieurs options fonctionnelles `OrchestratorAgentOption`.

---

### Types

```go
// OrchestratorAgentOption est une option fonctionnelle pour configurer un Agent lors de sa création
type OrchestratorAgentOption func(*Agent)

// Intent représente la sortie structurée de l'orchestrateur
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}

// OrchestratorAgent est l'interface implémentée par l'orchestrator agent
type OrchestratorAgent interface {
    IdentifyIntent(userMessages []messages.Message) (intent *Intent, finishReason string, err error)
    IdentifyTopicFromText(text string) (string, error)
}
```

---

### Fonctions d'options

| Fonction | Description |
|---|---|
| `BeforeCompletion(fn func(*Agent))` | Définit un hook appelé avant chaque identification d'intention. |
| `AfterCompletion(fn func(*Agent))` | Définit un hook appelé après chaque identification d'intention. |

---

### Méthodes

| Méthode | Description |
|---|---|
| `IdentifyIntent(msgs []messages.Message) (*agents.Intent, string, error)` | Identifie l'intention à partir de messages. Retourne l'intention, la raison de fin et l'erreur. |
| `IdentifyTopicFromText(text string) (string, error)` | Méthode de commodité pour détecter le topic à partir d'une chaîne de texte. |
| `GetMessages() []messages.Message` | Obtient tous les messages de conversation. |
| `AddMessage(role roles.Role, content string)` | Ajoute un message à l'historique. |
| `AddMessages(msgs []messages.Message)` | Ajoute plusieurs messages à l'historique. |
| `ResetMessages()` | Efface tous les messages sauf l'instruction système. |
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
| `Kind() agents.Kind` | Retourne `agents.Orchestrator`. |
| `GetName() string` | Retourne le nom de l'agent. |
| `GetModelID() string` | Retourne le nom du modèle. |
