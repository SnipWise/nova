# Guide de l'Agent Tools

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent](#3-configuration-de-lagent)
4. [Configuration du modèle](#4-configuration-du-modèle)
5. [Définition des outils](#5-définition-des-outils)
6. [Détection des appels d'outils](#6-détection-des-appels-doutils)
7. [Appels d'outils avec confirmation](#7-appels-doutils-avec-confirmation)
8. [Appels d'outils en streaming](#8-appels-doutils-en-streaming)
9. [Historique de conversation et messages](#9-historique-de-conversation-et-messages)
10. [Options : ToolAgentOption et ToolsAgentOption](#10-options--toolagentoption-et-toolsagentoption)
11. [Hooks de cycle de vie (ToolsAgentOption)](#11-hooks-de-cycle-de-vie-toolsagentoption)
12. [État des appels d'outils](#12-état-des-appels-doutils)
13. [Gestion du contexte et de l'état](#13-gestion-du-contexte-et-de-létat)
14. [Export JSON et débogage](#14-export-json-et-débogage)
15. [Référence API](#15-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Agent Tools ?

Le `tools.Agent` est un agent spécialisé fourni par le Nova SDK (`github.com/snipwise/nova`) qui permet l'appel de fonctions (tool use) avec les LLMs. Il envoie des messages au LLM avec les définitions d'outils, détecte quand le LLM souhaite appeler un outil, l'exécute via un callback, et renvoie le résultat au LLM.

### Quand utiliser un Agent Tools

| Scénario | Agent recommandé |
|---|---|
| Appel de fonctions / utilisation d'outils | `tools.Agent` |
| Boucles d'exécution d'outils multi-étapes | `tools.Agent` |
| Appels d'outils avec confirmation utilisateur | `tools.Agent` |
| IA conversationnelle en texte libre | `chat.Agent` |
| Extraction de données structurées | `structured.Agent[T]` |
| Détection d'intention et routage | `orchestrator.Agent` |
| Compression de contexte | `compressor.Agent` |
| Génération d'embeddings et recherche par similarité | `rag.Agent` |

### Capacités clés

- **Définition d'outils** : Définir des outils avec une API fluide ou utiliser les formats OpenAI/MCP directement.
- **Boucle de détection d'appels** : Détecter et exécuter automatiquement les appels d'outils en boucle jusqu'à l'arrêt du LLM.
- **Appels parallèles** : Support des LLMs capables d'appeler plusieurs outils simultanément.
- **Workflow de confirmation** : Demander optionnellement une confirmation utilisateur avant l'exécution (Confirmed/Denied/Quit).
- **Streaming** : Diffuser la réponse finale du LLM tout en traitant les appels d'outils.
- **Historique de conversation** : Maintenir optionnellement l'historique entre les appels.
- **Hooks de cycle de vie** : Exécuter de la logique personnalisée avant et après chaque détection d'appel d'outils.
- **Support MCP** : Utiliser des outils MCP (Model Context Protocol) aux côtés des outils natifs.

---

## 2. Démarrage rapide

### Exemple minimal

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := tools.NewAgent(
        ctx,
        agents.Config{
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a helpful assistant.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.0),
        },
        tools.WithTools([]*tools.Tool{
            tools.NewTool("calculate_sum").
                SetDescription("Calculate the sum of two numbers").
                AddParameter("a", "number", "The first number", true).
                AddParameter("b", "number", "The second number", true),
        }),
    )
    if err != nil {
        panic(err)
    }

    result, err := agent.DetectToolCallsLoop(
        []messages.Message{
            {Role: roles.User, Content: "What is 40 + 2?"},
        },
        func(functionName string, arguments string) (string, error) {
            var args struct {
                A float64 `json:"a"`
                B float64 `json:"b"`
            }
            json.Unmarshal([]byte(arguments), &args)
            return fmt.Sprintf(`{"result": %g}`, args.A+args.B), nil
        },
    )
    if err != nil {
        panic(err)
    }

    fmt.Println("Résultats :", result.Results)
    fmt.Println("Assistant :", result.LastAssistantMessage)
}
```

---

## 3. Configuration de l'agent

```go
agents.Config{
    Name:                    "tools-agent",
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
    APIKey:                  "your-api-key",
    SystemInstructions:      "You are a helpful assistant with tool access.",
    KeepConversationHistory: false,
}
```

| Champ | Type | Requis | Description |
|---|---|---|---|
| `Name` | `string` | Non | Identifiant de l'agent pour les logs. |
| `EngineURL` | `string` | Oui | URL du moteur LLM compatible OpenAI. |
| `APIKey` | `string` | Non | Clé API pour les moteurs authentifiés. |
| `SystemInstructions` | `string` | Recommandé | Prompt système. |
| `KeepConversationHistory` | `bool` | Non | Conserver les messages entre les appels. Par défaut : `false`. |

---

## 4. Configuration du modèle

```go
models.Config{
    Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
    Temperature:       models.Float64(0.0),
    ParallelToolCalls: models.Bool(false),
}
```

| Champ | Description |
|---|---|
| `ParallelToolCalls` | Activer/désactiver les appels d'outils parallèles. Tous les LLMs ne supportent pas cette fonctionnalité. |

---

## 5. Définition des outils

### API fluide (builder)

```go
tool := tools.NewTool("calculate_sum").
    SetDescription("Calculate the sum of two numbers").
    AddParameter("a", "number", "The first number", true).
    AddParameter("b", "number", "The second number", true)
```

Types de paramètres supportés : `"string"`, `"number"`, `"boolean"`, `"object"`, `"array"`.

### Passer les outils à l'agent

```go
agent, err := tools.NewAgent(ctx, agentConfig, modelConfig,
    tools.WithTools([]*tools.Tool{tool1, tool2}),
)
```

### Utiliser des outils OpenAI directement

```go
tools.WithOpenAITools([]openai.ChatCompletionToolUnionParam{...})
```

### Utiliser des outils MCP

```go
tools.WithMCPTools(mcpToolsList)
```

---

## 6. Détection des appels d'outils

### DetectToolCallsLoop

La méthode principale pour la détection d'appels d'outils. Exécute une boucle : envoie les messages au LLM, détecte les appels d'outils, les exécute, renvoie les résultats, jusqu'à ce que le LLM s'arrête.

```go
result, err := agent.DetectToolCallsLoop(
    []messages.Message{
        {Role: roles.User, Content: "What is 40 + 2?"},
    },
    func(functionName string, arguments string) (string, error) {
        // Exécuter l'outil et retourner le résultat en JSON
        return `{"result": 42}`, nil
    },
)
```

### DetectParallelToolCalls

Pour les LLMs supportant l'appel de plusieurs outils simultanément :

```go
result, err := agent.DetectParallelToolCalls(userMessages, toolCallback)
```

### ToolCallResult

Toutes les méthodes de détection retournent `*ToolCallResult` :

```go
type ToolCallResult struct {
    FinishReason         string
    Results              []string
    LastAssistantMessage string
}
```

---

## 7. Appels d'outils avec confirmation

### DetectToolCallsLoopWithConfirmation

Ajoute une étape de confirmation avant chaque exécution d'outil :

```go
result, err := agent.DetectToolCallsLoopWithConfirmation(
    userMessages,
    toolCallback,
    func(functionName string, arguments string) tools.ConfirmationResponse {
        fmt.Printf("Exécuter %s(%s) ? ", functionName, arguments)
        // Retourner tools.Confirmed, tools.Denied, ou tools.Quit
        return tools.Confirmed
    },
)
```

### Valeurs de ConfirmationResponse

| Valeur | Description |
|---|---|
| `tools.Confirmed` | Exécuter l'appel d'outil. |
| `tools.Denied` | Ignorer l'exécution mais continuer la boucle. |
| `tools.Quit` | Arrêter toute l'exécution de l'agent. |

### DetectParallelToolCallsWithConfirmation

Même chose en parallèle avec confirmation :

```go
result, err := agent.DetectParallelToolCallsWithConfirmation(
    userMessages, toolCallback, confirmationCallback,
)
```

---

## 8. Appels d'outils en streaming

### DetectToolCallsLoopStream

Diffuse la réponse finale du LLM tout en traitant les appels d'outils :

```go
result, err := agent.DetectToolCallsLoopStream(
    userMessages,
    toolCallback,
    func(chunk string) error {
        fmt.Print(chunk) // Afficher chaque morceau à mesure qu'il arrive
        return nil
    },
)
```

### DetectToolCallsLoopWithConfirmationStream

Combine streaming et confirmation :

```go
result, err := agent.DetectToolCallsLoopWithConfirmationStream(
    userMessages,
    toolCallback,
    confirmationCallback,
    streamCallback,
)
```

---

## 9. Historique de conversation et messages

### Gestion des messages

```go
msgs := agent.GetMessages()
agent.AddMessage(roles.User, "Un message")
agent.AddMessages([]messages.Message{...})
agent.ResetMessages()
agent.GetContextSize()
```

### Export en JSON

```go
jsonStr, err := agent.ExportMessagesToJSON()
```

---

## 10. Options : ToolAgentOption et ToolsAgentOption

L'agent tools supporte deux types d'options distincts, tous deux passés comme arguments variadiques `...any` à `NewAgent` :

### ToolAgentOption (niveau paramètres OpenAI)

`ToolAgentOption` opère sur `*openai.ChatCompletionNewParams` et configure les paramètres de requête LLM (outils, etc.) :

```go
tools.WithTools([]*tools.Tool{...})        // Outils via API fluide
tools.WithOpenAITools(openaiTools)          // Outils au format OpenAI
tools.WithMCPTools(mcpTools)                // Outils au format MCP
```

### ToolsAgentOption (niveau agent)

`ToolsAgentOption` opère sur l'`*Agent` de haut niveau et configure les hooks de cycle de vie et les callbacks par défaut :

```go
tools.BeforeCompletion(func(a *tools.Agent) { ... })       // Hook avant la détection d'appels d'outils
tools.AfterCompletion(func(a *tools.Agent) { ... })        // Hook après la détection d'appels d'outils
tools.WithExecuteFn(executeFunction)                        // Définir le callback d'exécution par défaut
tools.WithConfirmationPromptFn(confirmationPrompt)          // Définir le callback de confirmation par défaut
```

### Mixer les deux types d'options

```go
agent, err := tools.NewAgent(
    ctx, agentConfig, modelConfig,
    // ToolAgentOption
    tools.WithTools(myTools),
    // ToolsAgentOption
    tools.WithExecuteFn(executeFunction),
    tools.WithConfirmationPromptFn(confirmationPrompt),
    tools.BeforeCompletion(func(a *tools.Agent) {
        fmt.Println("Avant la détection d'appels d'outils...")
    }),
    tools.AfterCompletion(func(a *tools.Agent) {
        fmt.Println("Après la détection d'appels d'outils...")
    }),
)
```

**Avantages d'utiliser WithExecuteFn et WithConfirmationPromptFn :**
- Définir les callbacks une seule fois lors de la création de l'agent
- Omettre les paramètres de callback dans les méthodes de détection pour un code plus propre
- Conserver la flexibilité : les paramètres peuvent toujours remplacer les options si nécessaire

---

## 11. Hooks de cycle de vie (ToolsAgentOption)

Les hooks de cycle de vie permettent d'exécuter de la logique personnalisée avant et après chaque détection d'appel d'outils. Ils sont déclenchés dans les 6 méthodes de détection.

### ToolsAgentOption

```go
type ToolsAgentOption func(*Agent)
```

### BeforeCompletion

Appelé avant chaque détection d'appel d'outils. Le hook reçoit une référence vers l'agent.

```go
tools.BeforeCompletion(func(a *tools.Agent) {
    fmt.Printf("[AVANT] Agent : %s, Messages : %d\n",
        a.GetName(), len(a.GetMessages()))
})
```

### AfterCompletion

Appelé après chaque détection d'appel d'outils, une fois le résultat prêt. Le hook reçoit une référence vers l'agent.

```go
tools.AfterCompletion(func(a *tools.Agent) {
    fmt.Printf("[APRES] Agent : %s, Messages : %d\n",
        a.GetName(), len(a.GetMessages()))
})
```

### Les hooks sont déclenchés dans toutes les méthodes de détection

| Méthode | Hooks déclenchés |
|---|---|
| `DetectParallelToolCalls` | Oui |
| `DetectParallelToolCallsWithConfirmation` | Oui |
| `DetectToolCallsLoop` | Oui |
| `DetectToolCallsLoopWithConfirmation` | Oui |
| `DetectToolCallsLoopStream` | Oui |
| `DetectToolCallsLoopWithConfirmationStream` | Oui |

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Le code existant sans hooks continue de fonctionner sans aucune modification.

---

## 12. État des appels d'outils

### GetLastStateToolCalls

Accéder à l'état de la dernière exécution d'appel d'outils :

```go
state := agent.GetLastStateToolCalls()
// state.Confirmation : Confirmed, Denied, ou Quit
// state.ExecutionResult.Content : le résultat de l'outil
// state.ExecutionResult.ExecFinishReason : "function_executed", "user_denied", "user_quit", "error", "exit_loop"
```

### ResetLastStateToolCalls

```go
agent.ResetLastStateToolCalls()
```

---

## 13. Gestion du contexte et de l'état

### Obtenir et définir le contexte

```go
ctx := agent.GetContext()
agent.SetContext(newCtx)
```

### Obtenir et définir la configuration

```go
config := agent.GetConfig()
agent.SetConfig(newConfig)

modelConfig := agent.GetModelConfig()
agent.SetModelConfig(newModelConfig)
```

### Métadonnées de l'agent

```go
agent.Kind()       // Retourne agents.Tools
agent.GetName()    // Retourne le nom de l'agent
agent.GetModelID() // Retourne le nom du modèle
```

---

## 14. Export JSON et débogage

```go
rawReq := agent.GetLastRequestRawJSON()
rawResp := agent.GetLastResponseRawJSON()

prettyReq, err := agent.GetLastRequestSON()
prettyResp, err := agent.GetLastResponseJSON()
```

---

## 15. Référence API

### Constructeur

```go
func NewAgent(
    ctx context.Context,
    agentConfig agents.Config,
    modelConfig models.Config,
    options ...any,
) (*Agent, error)
```

Crée un nouvel agent tools. Le paramètre `options` accepte à la fois des `ToolAgentOption` (configuration des paramètres OpenAI) et des `ToolsAgentOption` (hooks de niveau agent). Le constructeur les sépare en interne par assertion de type.

---

### Types

```go
type ToolCallResult struct {
    FinishReason         string
    Results              []string
    LastAssistantMessage string
}

type ToolCallback func(functionName string, arguments string) (string, error)
type ConfirmationCallback func(functionName string, arguments string) ConfirmationResponse
type StreamCallback func(chunk string) error

type ConfirmationResponse int
const (
    Confirmed ConfirmationResponse = iota
    Denied
    Quit
)

// ToolAgentOption configure les paramètres OpenAI (outils, etc.)
type ToolAgentOption func(*openai.ChatCompletionNewParams)

// ToolsAgentOption configure l'Agent de haut niveau (hooks de cycle de vie)
type ToolsAgentOption func(*Agent)
```

---

### Fonctions d'options

| Fonction | Type | Description |
|---|---|---|
| `WithTools(tools []*Tool)` | `ToolAgentOption` | Définir les outils via l'API fluide. |
| `WithOpenAITools(tools []openai.ChatCompletionToolUnionParam)` | `ToolAgentOption` | Définir les outils au format OpenAI. |
| `WithMCPTools(tools []mcp.Tool)` | `ToolAgentOption` | Définir les outils au format MCP. |
| `WithExecuteFn(fn ToolCallback)` | `ToolsAgentOption` | Définir le callback d'exécution d'outils par défaut. Utilisé quand le paramètre callback est omis dans les méthodes de détection. |
| `WithConfirmationPromptFn(fn ConfirmationCallback)` | `ToolsAgentOption` | Définir le callback de confirmation par défaut. Utilisé quand le paramètre confirmation est omis dans les méthodes de confirmation. |
| `BeforeCompletion(fn func(*Agent))` | `ToolsAgentOption` | Hook appelé avant chaque détection d'appel d'outils. |
| `AfterCompletion(fn func(*Agent))` | `ToolsAgentOption` | Hook appelé après chaque détection d'appel d'outils. |

---

### Méthodes

| Méthode | Description |
|---|---|
| `DetectToolCallsLoop(msgs, callback...) (*ToolCallResult, error)` | Détecter et exécuter les appels d'outils en boucle. `callback` est optionnel si défini via `WithExecuteFn`. |
| `DetectToolCallsLoopWithConfirmation(msgs, callbacks...) (*ToolCallResult, error)` | Idem avec confirmation utilisateur. `callbacks` (toolCallback, confirmationCallback) sont optionnels si définis via options. L'ordre est important ! |
| `DetectToolCallsLoopStream(msgs, streamCallback, toolCallback...) (*ToolCallResult, error)` | Idem avec streaming. `streamCallback` est requis, `toolCallback` est optionnel. |
| `DetectToolCallsLoopWithConfirmationStream(msgs, streamCallback, callbacks...) (*ToolCallResult, error)` | Idem avec confirmation et streaming. `streamCallback` est requis, les autres callbacks sont optionnels. |
| `DetectParallelToolCalls(msgs, callback...) (*ToolCallResult, error)` | Détecter les appels d'outils parallèles. `callback` est optionnel si défini via `WithExecuteFn`. |
| `DetectParallelToolCallsWithConfirmation(msgs, callbacks...) (*ToolCallResult, error)` | Idem avec confirmation. `callbacks` sont optionnels si définis via options. |
| `GetMessages() []messages.Message` | Obtenir tous les messages de la conversation. |
| `AddMessage(role, content)` | Ajouter un message. |
| `AddMessages(msgs)` | Ajouter plusieurs messages. |
| `ResetMessages()` | Effacer les messages sauf l'instruction système. |
| `GetContextSize() int` | Obtenir la taille approximative du contexte. |
| `ExportMessagesToJSON() (string, error)` | Exporter la conversation en JSON. |
| `GetLastStateToolCalls() LastToolCallsState` | Obtenir l'état du dernier appel d'outils. |
| `ResetLastStateToolCalls()` | Réinitialiser l'état du dernier appel d'outils. |
| `GetConfig() agents.Config` | Obtient la configuration de l'agent. |
| `SetConfig(config)` | Met à jour la configuration de l'agent. |
| `GetModelConfig() models.Config` | Obtient la configuration du modèle. |
| `SetModelConfig(config)` | Met à jour la configuration du modèle. |
| `GetContext() context.Context` | Obtient le contexte de l'agent. |
| `SetContext(ctx)` | Met à jour le contexte de l'agent. |
| `GetLastRequestRawJSON() string` | JSON brut de la dernière requête. |
| `GetLastResponseRawJSON() string` | JSON brut de la dernière réponse. |
| `GetLastRequestSON() (string, error)` | JSON formaté de la dernière requête. |
| `GetLastResponseJSON() (string, error)` | JSON formaté de la dernière réponse. |
| `Kind() agents.Kind` | Retourne `agents.Tools`. |
| `GetName() string` | Retourne le nom de l'agent. |
| `GetModelID() string` | Retourne le nom du modèle. |
