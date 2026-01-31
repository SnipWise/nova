# Guide du Chat Agent

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent](#3-configuration-de-lagent)
4. [Configuration du modèle](#4-configuration-du-modèle)
5. [Méthodes de complétion](#5-méthodes-de-complétion)
6. [Complétion en streaming](#6-complétion-en-streaming)
7. [Support du raisonnement](#7-support-du-raisonnement)
8. [Historique de conversation](#8-historique-de-conversation)
9. [Directives de messages utilisateur](#9-directives-de-messages-utilisateur)
10. [Hooks de cycle de vie (ChatAgentOption)](#10-hooks-de-cycle-de-vie-chatagentoption)
11. [Gestion du contexte et de l'état](#11-gestion-du-contexte-et-de-létat)
12. [Export JSON et débogage](#12-export-json-et-débogage)
13. [Référence API](#13-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Chat Agent ?

Le `chat.Agent` est l'agent conversationnel principal fourni par le SDK Nova (`github.com/snipwise/nova`). Il encapsule l'API compatible OpenAI derrière une interface Go simplifiée, gérant de manière transparente le formatage des messages, l'historique de conversation, le streaming et la configuration du modèle.

Un `chat.Agent` est conçu pour une utilisation programmatique directe -- vous appelez des méthodes dans votre code Go et recevez les réponses directement, sans couche HTTP. Pour une utilisation basée sur HTTP, consultez les guides `ServerAgent` ou `CrewServerAgent`.

### Quand utiliser un Chat Agent

| Scénario | Agent recommandé |
|---|---|
| IA conversationnelle simple dans une application Go | `chat.Agent` |
| Conversations multi-tours avec historique | `chat.Agent` avec `KeepConversationHistory: true` |
| Question/réponse en un seul tour | `chat.Agent` avec `KeepConversationHistory: false` |
| Réponses en streaming vers un terminal ou une UI | `chat.Agent` avec `GenerateStreamCompletion` |
| Modèles de raisonnement/chaîne de pensée | `chat.Agent` avec `GenerateCompletionWithReasoning` |
| Appels de fonctions / utilisation d'outils | `tools.Agent` |
| API HTTP avec streaming SSE | `ServerAgent` ou `CrewServerAgent` |

### Capacités principales

- **Complétion standard** : Envoyez des messages et recevez une réponse complète.
- **Complétion en streaming** : Recevez les fragments de réponse en temps réel via des callbacks.
- **Support du raisonnement** : Récupérez à la fois la chaîne de raisonnement et la réponse finale des modèles compatibles.
- **Historique de conversation** : Maintenez optionnellement l'historique complet de conversation entre les appels.
- **Directives de messages** : Injectez des directives pré/post dans les messages utilisateur pour un cadrage cohérent.
- **Hooks de cycle de vie** : Exécutez une logique personnalisée avant et après chaque appel de complétion.
- **Export JSON** : Exportez l'historique de conversation en JSON pour le débogage ou la persistance.

---

## 2. Démarrage rapide

### Exemple minimal

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := chat.NewAgent(
        ctx,
        agents.Config{
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a helpful assistant.",
        },
        models.Config{
            Name:        "ai/qwen2.5:1.5B-F16",
            Temperature: models.Float64(0.7),
            MaxTokens:   models.Int(2000),
        },
    )
    if err != nil {
        panic(err)
    }

    result, err := agent.GenerateCompletion([]messages.Message{
        {Role: roles.User, Content: "Quelle est la capitale de la France ?"},
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Response)
    // Sortie : Paris est la capitale de la France.
}
```

---

## 3. Configuration de l'agent

Le struct `agents.Config` contrôle l'identité et le comportement de l'agent :

```go
agents.Config{
    Name:                    "mon-agent",            // Nom de l'agent (optionnel, pour identification)
    Description:             "Un assistant utile",   // Description de l'agent (optionnel)
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1", // URL du moteur LLM (requis)
    APIKey:                  "votre-clé-api",        // Clé API (optionnel, dépend du moteur)
    SystemInstructions:      "Tu es un assistant utile.", // Prompt système (recommandé)
    KeepConversationHistory: true,                   // Maintenir l'historique (défaut : false)
}
```

| Champ | Type | Requis | Description |
|---|---|---|---|
| `Name` | `string` | Non | Identifiant de l'agent, utilisé pour le logging et les configurations multi-agents. |
| `Description` | `string` | Non | Description lisible par un humain. |
| `EngineURL` | `string` | Oui | URL du moteur LLM compatible OpenAI. |
| `APIKey` | `string` | Non | Clé API pour les moteurs authentifiés. |
| `SystemInstructions` | `string` | Recommandé | Prompt système qui définit la personnalité et le comportement de l'agent. |
| `KeepConversationHistory` | `bool` | Non | Quand `true`, tous les messages utilisateur et assistant sont conservés entre les appels. Défaut : `false`. |

---

## 4. Configuration du modèle

Le struct `models.Config` contrôle les paramètres de génération du modèle :

```go
models.Config{
    Name:             "ai/qwen2.5:1.5B-F16",    // ID du modèle (requis)
    Temperature:      models.Float64(0.7),        // Créativité (0.0 = déterministe, 1.0+ = créatif)
    MaxTokens:        models.Int(2000),            // Longueur maximale de la réponse
    TopP:             models.Float64(0.9),         // Échantillonnage nucleus
    FrequencyPenalty: models.Float64(0.0),         // Pénalité pour les tokens répétés
    PresencePenalty:  models.Float64(0.0),         // Pénalité pour les tokens déjà présents
    Stop:             []string{"\n\n"},            // Séquences d'arrêt
    ReasoningEffort:  models.String("medium"),     // Effort de raisonnement pour les modèles compatibles
}
```

Tous les paramètres sauf `Name` sont optionnels et utilisent des types pointeurs (`*float64`, `*int64`) de sorte que `nil` signifie « utiliser la valeur par défaut du modèle ». Des fonctions utilitaires sont fournies :

- `models.Float64(v)` retourne `*float64`
- `models.Int(v)` retourne `*int64`
- `models.String(v)` retourne `*string`
- `models.Bool(v)` retourne `*bool`

---

## 5. Méthodes de complétion

### GenerateCompletion

La manière la plus simple d'obtenir une réponse de l'agent :

```go
result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Bonjour, qui êtes-vous ?"},
})
if err != nil {
    // gérer l'erreur
}

fmt.Println(result.Response)     // Le texte de réponse de l'agent
fmt.Println(result.FinishReason) // "stop", "length", etc.
```

**Type de retour :** `*CompletionResult`

```go
type CompletionResult struct {
    Response     string // Le texte de réponse généré
    FinishReason string // Raison de l'arrêt de la génération ("stop", "length", etc.)
}
```

### Envoi de plusieurs messages

Vous pouvez envoyer plusieurs messages en un seul appel pour fournir du contexte :

```go
result, err := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Je m'appelle Alice."},
    {Role: roles.Assistant, Content: "Enchanté, Alice !"},
    {Role: roles.User, Content: "Quel est mon prénom ?"},
})
```

---

## 6. Complétion en streaming

### GenerateStreamCompletion

Le streaming vous permet de recevoir les fragments de réponse au fur et à mesure de leur génération, offrant une expérience en temps réel :

```go
result, err := agent.GenerateStreamCompletion(
    []messages.Message{
        {Role: roles.User, Content: "Raconte-moi une histoire de chat."},
    },
    func(chunk string, finishReason string) error {
        fmt.Print(chunk) // Affiche chaque fragment à mesure qu'il arrive
        return nil
    },
)
if err != nil {
    // gérer l'erreur
}
fmt.Println()
fmt.Println("Raison de fin :", result.FinishReason)
```

La fonction callback reçoit :
- `chunk` : Un morceau du texte de réponse (peut être vide lors du dernier appel).
- `finishReason` : Vide pour les fragments intermédiaires, défini à `"stop"`, `"length"`, etc. pour le fragment final.

La méthode retourne également le `*CompletionResult` complet une fois le streaming terminé.

### Arrêter un stream

Vous pouvez interrompre un stream en cours depuis une autre goroutine :

```go
agent.StopStream()
```

---

## 7. Support du raisonnement

Pour les modèles qui supportent le raisonnement en chaîne de pensée (ex. : DeepSeek-R1, QwQ), l'agent fournit des méthodes spécifiques qui retournent à la fois la chaîne de raisonnement et la réponse finale.

### GenerateCompletionWithReasoning

```go
result, err := agent.GenerateCompletionWithReasoning([]messages.Message{
    {Role: roles.User, Content: "Combien font 15% de 240 ?"},
})
if err != nil {
    // gérer l'erreur
}

fmt.Println("Raisonnement :", result.Reasoning)  // La chaîne de raisonnement du modèle
fmt.Println("Réponse :", result.Response)         // La réponse finale
fmt.Println("Fin :", result.FinishReason)
```

**Type de retour :** `*ReasoningResult`

```go
type ReasoningResult struct {
    Response     string // La réponse finale
    Reasoning    string // La chaîne de raisonnement/réflexion
    FinishReason string // Raison de l'arrêt de la génération
}
```

### GenerateStreamCompletionWithReasoning

Variante streaming avec des callbacks séparés pour le raisonnement et la réponse :

```go
result, err := agent.GenerateStreamCompletionWithReasoning(
    []messages.Message{
        {Role: roles.User, Content: "Explique l'informatique quantique."},
    },
    // Callback de raisonnement
    func(chunk string, finishReason string) error {
        fmt.Print(chunk) // Stream des fragments de raisonnement
        return nil
    },
    // Callback de réponse
    func(chunk string, finishReason string) error {
        fmt.Print(chunk) // Stream des fragments de réponse
        return nil
    },
)
```

---

## 8. Historique de conversation

### Activer l'historique de conversation

Définissez `KeepConversationHistory: true` dans la configuration de l'agent pour maintenir le contexte entre les appels :

```go
agent, err := chat.NewAgent(ctx,
    agents.Config{
        EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions:      "Tu es Bob, un assistant utile.",
        KeepConversationHistory: true,
    },
    models.Config{
        Name: "ai/qwen2.5:1.5B-F16",
    },
)
```

Avec l'historique activé, l'agent se souvient des interactions précédentes :

```go
// Premier appel
agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Qui est James T Kirk ?"},
})

// Deuxième appel - l'agent connaît le contexte du premier appel
result, _ := agent.GenerateCompletion([]messages.Message{
    {Role: roles.User, Content: "Qui est son meilleur ami ?"},
})
// L'agent peut répondre Spock car il se souvient de Kirk
```

### Sans historique de conversation

Quand `KeepConversationHistory` est `false` (par défaut), chaque appel est indépendant. L'agent ne se souvient pas des interactions précédentes.

### Gestion des messages

```go
// Obtenir tous les messages de l'historique
msgs := agent.GetMessages()

// Obtenir la taille approximative du contexte (en caractères)
size := agent.GetContextSize()

// Effacer tous les messages sauf l'instruction système
agent.ResetMessages()

// Supprimer les N derniers messages
agent.RemoveLastNMessages(2)

// Ajouter manuellement un message à l'historique
agent.AddMessage(roles.User, "Un message manuel")

// Ajouter plusieurs messages d'un coup
agent.AddMessages([]messages.Message{
    {Role: roles.User, Content: "Premier message"},
    {Role: roles.Assistant, Content: "Première réponse"},
})
```

### Mettre à jour les instructions système

```go
agent.SetSystemInstructions("Tu es maintenant un assistant pirate. Arrr !")
```

---

## 9. Directives de messages utilisateur

Les directives de messages utilisateur permettent d'ajouter automatiquement du texte avant ou après chaque message utilisateur. C'est utile pour cadrer de manière cohérente les entrées utilisateur avec du contexte ou des instructions supplémentaires.

### Définir les directives

```go
// Ajouter un préfixe à chaque message utilisateur
agent.SetUserMessagePreDirectives("Réponds toujours en français formel.")

// Ajouter un suffixe à chaque message utilisateur
agent.SetUserMessagePostDirectives("Limite ta réponse à 100 mots.")
```

### Fonctionnement des directives

Quand un utilisateur envoie "Qu'est-ce que Go ?", le message réellement envoyé au modèle devient :

```
Réponds toujours en français formel.

Qu'est-ce que Go ?

Limite ta réponse à 100 mots.
```

### Obtenir les directives actuelles

```go
pre := agent.GetUserMessagePreDirectives()
post := agent.GetUserMessagePostDirectives()
```

---

## 10. Hooks de cycle de vie (ChatAgentOption)

Les hooks de cycle de vie permettent d'exécuter une logique personnalisée avant et après chaque appel de complétion. Ils sont configurés comme des options fonctionnelles lors de la création de l'agent.

### ChatAgentOption

```go
type ChatAgentOption func(*Agent)
```

Les options sont passées en arguments variadiques à `NewAgent` :

```go
agent, err := chat.NewAgent(ctx, agentConfig, modelConfig,
    chat.BeforeCompletion(fn),
    chat.AfterCompletion(fn),
)
```

### BeforeCompletion

Appelé avant chaque complétion (standard et streaming). Le hook reçoit une référence à l'agent, vous permettant d'inspecter ou modifier son état.

```go
chat.BeforeCompletion(func(a *chat.Agent) {
    fmt.Println("Appel du LLM en cours...")
    fmt.Printf("Taille du contexte actuel : %d\n", a.GetContextSize())
    fmt.Printf("Nombre de messages : %d\n", len(a.GetMessages()))
})
```

**Cas d'usage :**
- Logging et monitoring
- Collecte de métriques (mesurer la fréquence des requêtes)
- Vérification de la taille du contexte avant chaque appel
- Mise à jour dynamique des instructions système

### AfterCompletion

Appelé après chaque complétion (standard et streaming), une fois la réponse complète reçue. Le hook reçoit une référence à l'agent.

```go
chat.AfterCompletion(func(a *chat.Agent) {
    fmt.Println("Appel LLM terminé.")
    fmt.Printf("Taille du contexte mise à jour : %d\n", a.GetContextSize())
    fmt.Printf("Nombre de messages : %d\n", len(a.GetMessages()))
})
```

**Cas d'usage :**
- Logging des métriques de réponse
- Post-traitement (ex. : sauvegarder la conversation en base de données)
- Monitoring de la taille du contexte après la réponse
- Déclenchement d'actions en aval

### Exemple complet avec hooks

```go
agent, err := chat.NewAgent(
    ctx,
    agents.Config{
        EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions:      "Tu es Bob, un assistant IA utile.",
        KeepConversationHistory: true,
    },
    models.Config{
        Name:        "ai/qwen2.5:1.5B-F16",
        Temperature: models.Float64(0.0),
        MaxTokens:   models.Int(2000),
    },
    chat.BeforeCompletion(func(a *chat.Agent) {
        fmt.Printf("[AVANT] Contexte : %d chars, Messages : %d\n",
            a.GetContextSize(), len(a.GetMessages()))
    }),
    chat.AfterCompletion(func(a *chat.Agent) {
        fmt.Printf("[APRÈS] Contexte : %d chars, Messages : %d\n",
            a.GetContextSize(), len(a.GetMessages()))
    }),
)
```

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Les hooks ne sont appelés que lorsqu'ils ont été définis. Le paramètre `...ChatAgentOption` est variadique, donc le code existant sans hooks continue de fonctionner sans aucune modification.

### Les hooks s'appliquent à toutes les méthodes de complétion

Les hooks `BeforeCompletion` et `AfterCompletion` sont déclenchés par les quatre méthodes de complétion :

| Méthode | BeforeCompletion | AfterCompletion |
|---|---|---|
| `GenerateCompletion` | Oui | Oui |
| `GenerateCompletionWithReasoning` | Oui | Oui |
| `GenerateStreamCompletion` | Oui | Oui |
| `GenerateStreamCompletionWithReasoning` | Oui | Oui |

---

## 11. Gestion du contexte et de l'état

### Obtenir et définir le contexte

L'agent porte un `context.Context` pour l'annulation et la propagation de valeurs :

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
agent.Kind()       // Retourne agents.Chat
agent.GetName()    // Retourne le nom de l'agent depuis la config
agent.GetModelID() // Retourne le nom du modèle depuis la config modèle
```

---

## 12. Export JSON et débogage

### Exporter la conversation en JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
if err != nil {
    // gérer l'erreur
}
fmt.Println(jsonStr)
```

Sortie :

```json
[
  {
    "role": "system",
    "content": "Tu es un assistant utile."
  },
  {
    "role": "user",
    "content": "Bonjour"
  },
  {
    "role": "assistant",
    "content": "Bonjour ! Comment puis-je vous aider ?"
  }
]
```

### JSON brut requête/réponse

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

## 13. Référence API

### Constructeur

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    opts ...ChatAgentOption,
) (*Agent, error)
```

Crée un nouvel agent de chat. Le paramètre `opts` accepte zéro ou plusieurs options fonctionnelles `ChatAgentOption`.

---

### Types

```go
// ChatAgentOption est une option fonctionnelle pour configurer un Agent lors de sa création
type ChatAgentOption func(*Agent)

// CompletionResult représente le résultat d'une complétion de chat
type CompletionResult struct {
    Response     string
    FinishReason string
}

// ReasoningResult représente le résultat d'une complétion de chat avec raisonnement
type ReasoningResult struct {
    Response     string
    Reasoning    string
    FinishReason string
}

// StreamCallback est une fonction appelée pour chaque fragment de réponse en streaming
type StreamCallback func(chunk string, finishReason string) error
```

---

### Fonctions d'options

| Fonction | Description |
|---|---|
| `BeforeCompletion(fn func(*Agent))` | Définit un hook appelé avant chaque complétion (standard et streaming). |
| `AfterCompletion(fn func(*Agent))` | Définit un hook appelé après chaque complétion (standard et streaming). |

---

### Méthodes

| Méthode | Description |
|---|---|
| `GenerateCompletion(msgs []messages.Message) (*CompletionResult, error)` | Envoie des messages et obtient une réponse complète. |
| `GenerateCompletionWithReasoning(msgs []messages.Message) (*ReasoningResult, error)` | Envoie des messages et obtient une réponse avec chaîne de raisonnement. |
| `GenerateStreamCompletion(msgs []messages.Message, cb StreamCallback) (*CompletionResult, error)` | Envoie des messages et streame la réponse via callback. |
| `GenerateStreamCompletionWithReasoning(msgs []messages.Message, reasoningCb StreamCallback, responseCb StreamCallback) (*ReasoningResult, error)` | Envoie des messages et streame le raisonnement et la réponse. |
| `GetMessages() []messages.Message` | Obtient tous les messages de conversation. |
| `GetContextSize() int` | Obtient la taille approximative du contexte en caractères. |
| `ResetMessages()` | Efface tous les messages sauf l'instruction système. |
| `RemoveLastNMessages(n int)` | Supprime les N derniers messages de l'historique. |
| `AddMessage(role roles.Role, content string)` | Ajoute un message à l'historique. |
| `AddMessages(msgs []messages.Message)` | Ajoute plusieurs messages à l'historique. |
| `SetSystemInstructions(instructions string)` | Met à jour les instructions système. |
| `SetUserMessagePreDirectives(directives string)` | Définit le texte ajouté avant chaque message utilisateur. |
| `GetUserMessagePreDirectives() string` | Obtient les pré-directives actuelles. |
| `SetUserMessagePostDirectives(directives string)` | Définit le texte ajouté après chaque message utilisateur. |
| `GetUserMessagePostDirectives() string` | Obtient les post-directives actuelles. |
| `StopStream()` | Interrompt l'opération de streaming en cours. |
| `ExportMessagesToJSON() (string, error)` | Exporte l'historique de conversation en JSON. |
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
| `Kind() agents.Kind` | Retourne `agents.Chat`. |
| `GetName() string` | Retourne le nom de l'agent. |
| `GetModelID() string` | Retourne le nom du modèle. |
