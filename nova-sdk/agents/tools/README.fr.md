# Tools Agent

## Description

Le **Tools Agent** est un agent spécialisé dans la détection et l'exécution d'appels de fonctions (function calling). Il permet à un LLM de décider quand et comment utiliser des outils/fonctions externes pour accomplir des tâches.

## Fonctionnalités

- **Détection de tool calls** : Identifie quand le LLM veut appeler une fonction
- **Exécution de fonctions** : Execute les fonctions via des callbacks
- **Appels parallèles** : Support des appels de fonctions en parallèle (si le modèle le supporte)
- **Boucle d'exécution** : Exécute plusieurs appels successifs automatiquement
- **Confirmation utilisateur** : Human-in-the-loop pour valider les appels de fonctions
- **Streaming** : Support du streaming pendant la détection et l'exécution
- **Support MCP** : Intégration avec MCP (Model Context Protocol) tools

## Cas d'usage

Le Tools Agent est utilisé pour :
- **Appels de fonctions** : Calculateur, météo, API externes, base de données
- **Actions** : Envoyer des emails, créer des fichiers, effectuer des requêtes HTTP
- **Human-in-the-loop** : Demander confirmation avant exécution
- **Automatisation** : Chaîner plusieurs appels de fonctions automatiquement

## Création d'un Tools Agent

### Syntaxe de base

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/models"
)

ctx := context.Background()

// Configuration de l'agent
agentConfig := agents.Config{
    Name: "ToolsAgent",
    Instructions: "You are a helpful assistant with access to tools.",
}

// Configuration du modèle (doit supporter function calling)
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    Name:      "hf.co/menlo/jan-nano-gguf:q4_k_m", // Supporte function calling
}

// Définir les tools disponibles
myTools := []*tools.Tool{
    tools.NewTool("calculate").
        SetDescription("Perform a mathematical calculation").
        AddParameter("expression", "string", "The mathematical expression to evaluate", true),

    tools.NewTool("get_weather").
        SetDescription("Get current weather for a location").
        AddParameter("location", "string", "City name", true).
        AddParameter("unit", "string", "Temperature unit (celsius/fahrenheit)", false),
}

// Créer l'agent avec tools
agent, err := tools.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    tools.WithTools(myTools),
)
if err != nil {
    log.Fatal(err)
}
```

### Options de création

| Option | Description |
|--------|-------------|
| `WithTools(tools)` | Définit les tools Nova SDK |
| `WithOpenAITools(tools)` | Définit les tools au format OpenAI brut |
| `WithMCPTools(tools)` | Définit les tools MCP (Model Context Protocol) |

## Définition de Tools

### API Fluent

```go
// Créer un tool avec l'API fluent
calculateTool := tools.NewTool("calculate").
    SetDescription("Perform mathematical calculations").
    AddParameter("expression", "string", "Expression to evaluate (e.g., '2 + 2')", true).
    AddParameter("precision", "number", "Number of decimal places", false)

emailTool := tools.NewTool("send_email").
    SetDescription("Send an email to a recipient").
    AddParameter("to", "string", "Recipient email address", true).
    AddParameter("subject", "string", "Email subject", true).
    AddParameter("body", "string", "Email body content", true)
```

### Structure Tool

```go
type Tool struct {
    Name        string
    Description string
    Parameters  map[string]Parameter
    Required    []string
}

type Parameter struct {
    Type        string // "string", "number", "boolean", "object", "array"
    Description string
}
```

### Types de paramètres

- `"string"` : Texte
- `"number"` : Nombre (entier ou décimal)
- `"boolean"` : Booléen (true/false)
- `"object"` : Objet JSON
- `"array"` : Tableau

## Méthodes principales

### DetectToolCallsLoop - Boucle d'exécution (recommandé)

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

// Callback pour exécuter les fonctions
executeFunction := func(functionName string, arguments string) (string, error) {
    switch functionName {
    case "calculate":
        // Parse arguments et exécuter
        return `{"result": 4}`, nil
    case "get_weather":
        return `{"temperature": 22, "condition": "sunny"}`, nil
    default:
        return "", fmt.Errorf("unknown function: %s", functionName)
    }
}

// Détecter et exécuter les tool calls en boucle
userMessages := []messages.Message{
    {Role: roles.User, Content: "Quelle est la météo à Paris et combien font 2 + 2 ?"},
}

result, err := agent.DetectToolCallsLoop(userMessages, executeFunction)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Finish reason:", result.FinishReason)         // "function_executed" ou "stop"
fmt.Println("Tool results:", result.Results)               // ["{"result": 4}", "{"temperature": 22, ...}"]
fmt.Println("Assistant message:", result.LastAssistantMessage)
```

### DetectToolCallsLoopWithConfirmation - Avec confirmation utilisateur

```go
// Callback de confirmation (human-in-the-loop)
confirmationPrompt := func(functionName string, arguments string) tools.ConfirmationResponse {
    fmt.Printf("Execute %s with args %s? (y/n/q): ", functionName, arguments)
    var response string
    fmt.Scanln(&response)

    switch response {
    case "y":
        return tools.Confirmed
    case "n":
        return tools.Denied
    case "q":
        return tools.Quit
    default:
        return tools.Denied
    }
}

// Détecter et exécuter avec confirmation
result, err := agent.DetectToolCallsLoopWithConfirmation(
    userMessages,
    executeFunction,
    confirmationPrompt,
)
```

**Réponses de confirmation** :
- `tools.Confirmed` : Exécuter la fonction
- `tools.Denied` : Ne pas exécuter, mais continuer
- `tools.Quit` : Arrêter toute l'exécution

### DetectParallelToolCalls - Appels parallèles

**Note** : Tous les LLMs ne supportent pas les appels parallèles.

```go
// Exécuter plusieurs tool calls en parallèle
result, err := agent.DetectParallelToolCalls(userMessages, executeFunction)

// Avec confirmation
result, err := agent.DetectParallelToolCallsWithConfirmation(
    userMessages,
    executeFunction,
    confirmationPrompt,
)
```

### Streaming

```go
// Streaming pendant la détection de tool calls
streamCallback := func(chunk string) error {
    fmt.Print(chunk)
    return nil
}

result, err := agent.DetectToolCallsLoopStream(
    userMessages,
    executeFunction,
    streamCallback,
)

// Avec confirmation
result, err := agent.DetectToolCallsLoopWithConfirmationStream(
    userMessages,
    executeFunction,
    confirmationPrompt,
    streamCallback,
)
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

// Exporter en JSON
jsonData, err := agent.ExportMessagesToJSON()

// Taille du contexte
contextSize := agent.GetContextSize()
```

### État des tool calls

```go
// Récupérer l'état du dernier appel de fonction
state := agent.GetLastStateToolCalls()

// State contient:
// - Confirmation: tools.ConfirmationResponse
// - ExecutionResult.Content: Résultat de l'exécution
// - ExecutionResult.ExecFinishReason: "function_executed", "user_denied", "user_quit", etc.
// - ExecutionResult.ShouldStop: Si l'exécution doit s'arrêter

fmt.Println("Confirmation:", state.Confirmation)
fmt.Println("Finish reason:", state.ExecutionResult.ExecFinishReason)

// Réinitialiser l'état
agent.ResetLastStateToolCalls()
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
kind := agent.Kind() // Retourne agents.Tools

// Contexte
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requêtes/Réponses (debugging)
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
prettyRequest, _ := agent.GetLastRequestSON()
prettyResponse, _ := agent.GetLastResponseJSON()
```

## Structure ToolCallResult

```go
type ToolCallResult struct {
    FinishReason         string   // "function_executed", "stop", "user_denied", "user_quit"
    Results              []string // Résultats JSON des fonctions exécutées
    LastAssistantMessage string   // Dernier message de l'assistant
}
```

## Utilisation avec d'autres agents

Le Tools Agent est généralement utilisé avec Server ou Crew agents :

```go
// Créer le tools agent
toolsAgent, _ := tools.NewAgent(ctx, agentConfig, modelConfig, tools.WithTools(myTools))

// Fonction d'exécution
executeFn := func(functionName string, arguments string) (string, error) {
    // Implementation...
    return result, nil
}

// Utiliser avec Server Agent
serverAgent, _ := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithToolsAgent(toolsAgent),
    server.WithExecuteFn(executeFn),
)

// Utiliser avec Crew Agent
crewAgent, _ := crew.NewAgent(
    ctx,
    crew.WithSingleAgent(chatAgent),
    crew.WithToolsAgent(toolsAgent),
    crew.WithExecuteFn(executeFn),
    crew.WithConfirmationPromptFn(confirmationPrompt),
)
```

## Exemple complet

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/tools"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Configuration
    agentConfig := agents.Config{
        Name:         "Calculator",
        Instructions: "You are a helpful calculator assistant.",
    }
    modelConfig := models.Config{
        EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
        Name:      "hf.co/menlo/jan-nano-gguf:q4_k_m",
    }

    // Définir les tools
    myTools := []*tools.Tool{
        tools.NewTool("add").
            SetDescription("Add two numbers").
            AddParameter("a", "number", "First number", true).
            AddParameter("b", "number", "Second number", true),

        tools.NewTool("multiply").
            SetDescription("Multiply two numbers").
            AddParameter("a", "number", "First number", true).
            AddParameter("b", "number", "Second number", true),
    }

    // Créer l'agent
    agent, err := tools.NewAgent(ctx, agentConfig, modelConfig, tools.WithTools(myTools))
    if err != nil {
        log.Fatal(err)
    }

    // Fonction d'exécution
    executeFunction := func(functionName string, arguments string) (string, error) {
        var args map[string]float64
        if err := json.Unmarshal([]byte(arguments), &args); err != nil {
            return "", err
        }

        var result float64
        switch functionName {
        case "add":
            result = args["a"] + args["b"]
        case "multiply":
            result = args["a"] * args["b"]
        default:
            return "", fmt.Errorf("unknown function: %s", functionName)
        }

        return fmt.Sprintf(`{"result": %f}`, result), nil
    }

    // Détecter et exécuter
    userMessages := []messages.Message{
        {Role: roles.User, Content: "Combien font 5 + 3 ?"},
    }

    result, err := agent.DetectToolCallsLoop(userMessages, executeFunction)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Finish reason: %s\n", result.FinishReason)
    fmt.Printf("Results: %v\n", result.Results)
    fmt.Printf("Assistant: %s\n", result.LastAssistantMessage)
}
```

## MCP (Model Context Protocol) Support

Le Tools Agent supporte les tools MCP :

```go
import "github.com/mark3labs/mcp-go/mcp"

// Tools MCP
mcpTools := []mcp.Tool{
    // Vos tools MCP
}

// Créer l'agent avec tools MCP
agent, err := tools.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    tools.WithMCPTools(mcpTools),
)
```

## Notes

- **Kind** : Retourne `agents.Tools`
- **Modèles requis** : Seuls les modèles supportant function calling fonctionnent
- **Appels parallèles** : Tous les LLMs ne supportent pas cette fonctionnalité
- **Boucle automatique** : `DetectToolCallsLoop` exécute plusieurs appels successifs
- **Confirmation** : `WithConfirmation` ajoute human-in-the-loop
- **Streaming** : Compatible avec le streaming de réponses
- **État persistant** : `GetLastStateToolCalls()` permet de maintenir l'état entre invocations

## Recommandations

### Modèles recommandés pour function calling

- **hf.co/menlo/jan-nano-gguf:q4_k_m** : Petit, rapide, bon support
- **qwen2.5:1.5b** : Équilibre taille/performance
- **Éviter** : Modèles sans support natif de function calling

### Bonnes pratiques

1. **Descriptions claires** : Décrivez précisément ce que fait chaque tool
2. **Paramètres explicites** : Indiquez les types et les contraintes
3. **Gestion d'erreurs** : Retournez des erreurs JSON claires
4. **Confirmation** : Utilisez `WithConfirmation` pour les actions sensibles
5. **Validation** : Validez les arguments avant exécution

```go
executeFunction := func(functionName string, arguments string) (string, error) {
    // Parse
    var args map[string]any
    if err := json.Unmarshal([]byte(arguments), &args); err != nil {
        return `{"error": "Invalid JSON arguments"}`, err
    }

    // Validate
    if functionName == "delete_file" {
        if args["path"] == "" {
            return `{"error": "path is required"}`, fmt.Errorf("missing path")
        }
    }

    // Execute
    // ...

    return `{"success": true}`, nil
}
```
