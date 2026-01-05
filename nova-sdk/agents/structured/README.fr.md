# Structured Agent

## Description

Le **Structured Agent** est un agent spécialisé dans la génération de données structurées au format JSON. Il garantit que la réponse du LLM respecte un schéma JSON strict défini par une structure Go, en utilisant la fonctionnalité de sortie structurée d'OpenAI.

## Fonctionnalités

- **Sortie structurée garantie** : Le LLM retourne toujours un JSON valide conforme au schéma
- **Typage fort** : Utilise les generics Go pour définir le type de sortie
- **Génération automatique de schéma** : Convertit automatiquement une struct Go en JSON Schema
- **Gestion de l'historique** : Support de la conversation contextuelle (optionnel)
- **Validation stricte** : Mode strict d'OpenAI pour garantir la conformité du schéma

## Cas d'usage

Le Structured Agent est utilisé pour :
- **Extraction d'informations** : Extraire des données structurées à partir de texte
- **Classification** : Catégoriser du contenu avec des champs définis
- **Parsing** : Convertir du langage naturel en structures de données
- **APIs** : Générer des réponses JSON pour des APIs
- **Validation de données** : Garantir un format de sortie cohérent

## Création d'un Structured Agent

### Syntaxe de base

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/structured"
    "github.com/snipwise/nova/nova-sdk/models"
)

// Définir la structure de sortie
type Country struct {
    Name       string   `json:"name"`
    Capital    string   `json:"capital"`
    Population int      `json:"population"`
    Languages  []string `json:"languages"`
}

ctx := context.Background()

// Créer l'agent avec le type de sortie
agent, err := structured.NewAgent[Country](
    ctx,
    agents.Config{
        Name:               "StructuredAgent",
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "You are an assistant that provides country information.",
    },
    models.Config{
        Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
        Temperature: models.Float64(0.0), // Déterministe pour données structurées
    },
)
if err != nil {
    log.Fatal(err)
}
```

## Méthodes principales

### GenerateStructuredData - Génération de données structurées

```go
import (
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
)

// Générer des données structurées
response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Tell me about Canada."},
})

if err != nil {
    log.Fatal(err)
}

// response est de type *Country
fmt.Printf("Name: %s\n", response.Name)
fmt.Printf("Capital: %s\n", response.Capital)
fmt.Printf("Population: %d\n", response.Population)
fmt.Printf("Languages: %v\n", response.Languages)
fmt.Printf("Finish reason: %s\n", finishReason)
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
kind := agent.Kind() // Retourne agents.Structured

// Contexte
ctx := agent.GetContext()
agent.SetContext(newCtx)

// Requêtes/Réponses (debugging)
rawRequest := agent.GetLastRequestRawJSON()
rawResponse := agent.GetLastResponseRawJSON()
prettyRequest, _ := agent.GetLastRequestJSON()
prettyResponse, _ := agent.GetLastResponseJSON()
```

## Structure StructuredResult

```go
type StructuredResult[Output any] struct {
    Data         *Output  // Les données structurées générées
    FinishReason string   // Raison de fin ("stop", "length", etc.)
}
```

## Types de structures supportés

### Structure simple

```go
type Person struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

agent, _ := structured.NewAgent[Person](ctx, agentConfig, modelConfig)
```

### Structure avec slices

```go
type Book struct {
    Title   string   `json:"title"`
    Authors []string `json:"authors"`
    Year    int      `json:"year"`
}

agent, _ := structured.NewAgent[Book](ctx, agentConfig, modelConfig)
```

### Slice de structures

```go
type Product struct {
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

// Génère un tableau de produits
agent, _ := structured.NewAgent[[]Product](ctx, agentConfig, modelConfig)

response, _, _ := agent.GenerateStructuredData(messages)
// response est de type *[]Product
for _, product := range *response {
    fmt.Printf("%s: $%.2f\n", product.Name, product.Price)
}
```

### Structures imbriquées

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type Company struct {
    Name    string  `json:"name"`
    Address Address `json:"address"`
    Revenue float64 `json:"revenue"`
}

agent, _ := structured.NewAgent[Company](ctx, agentConfig, modelConfig)
```

## Exemple complet

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/structured"
    "github.com/snipwise/nova/nova-sdk/messages"
    "github.com/snipwise/nova/nova-sdk/messages/roles"
    "github.com/snipwise/nova/nova-sdk/models"
)

// Définir la structure de sortie
type MovieInfo struct {
    Title    string   `json:"title"`
    Director string   `json:"director"`
    Year     int      `json:"year"`
    Genres   []string `json:"genres"`
    Rating   float64  `json:"rating"`
}

func main() {
    ctx := context.Background()

    // Créer l'agent
    agent, err := structured.NewAgent[MovieInfo](
        ctx,
        agents.Config{
            Name:               "MovieAgent",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You provide information about movies.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.0),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Générer des données structurées
    response, finishReason, err := agent.GenerateStructuredData([]messages.Message{
        {Role: roles.User, Content: "Tell me about the movie Inception."},
    })
    if err != nil {
        log.Fatal(err)
    }

    // Afficher les résultats
    fmt.Printf("Title: %s\n", response.Title)
    fmt.Printf("Director: %s\n", response.Director)
    fmt.Printf("Year: %d\n", response.Year)
    fmt.Printf("Genres: %v\n", response.Genres)
    fmt.Printf("Rating: %.1f/10\n", response.Rating)
    fmt.Printf("Finish reason: %s\n", finishReason)
}
```

**Sortie attendue** :
```
Title: Inception
Director: Christopher Nolan
Year: 2010
Genres: [Science Fiction Thriller Action]
Rating: 8.8/10
Finish reason: stop
```

## Génération automatique de JSON Schema

Le Structured Agent convertit automatiquement vos structs Go en JSON Schema :

```go
type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

// Génère automatiquement le schéma :
// {
//   "type": "object",
//   "properties": {
//     "name": {"type": "string"},
//     "age": {"type": "integer"},
//     "email": {"type": "string"}
//   },
//   "required": ["name", "age", "email"]
// }
```

**Tags JSON supportés** :
- `json:"fieldname"` : Définit le nom du champ dans le JSON
- `json:"-"` : Exclut le champ du schéma JSON

## Notes

- **Kind** : Retourne `agents.Structured`
- **Température** : Utiliser 0.0 pour des résultats déterministes
- **Modèles compatibles** : Tous les modèles supportant le format de réponse structuré OpenAI
- **Validation stricte** : Le mode strict garantit que la sortie respecte exactement le schéma
- **Types génériques** : Utilise les generics Go (Go 1.18+)
- **Historique** : Configurable via `KeepConversationHistory` dans `agents.Config`

## Recommandations

### Bonnes pratiques

1. **Température 0.0** : Utiliser une température basse pour des données structurées cohérentes
2. **Noms de champs clairs** : Utiliser des noms de champs descriptifs en anglais
3. **Tags JSON** : Toujours définir les tags JSON pour contrôler la sérialisation
4. **Validation** : Vérifier que la réponse n'est pas nil avant utilisation
5. **Types simples** : Préférer les types Go standards (string, int, float64, bool)

### Exemple de validation

```go
response, finishReason, err := agent.GenerateStructuredData(messages)
if err != nil {
    log.Printf("Error generating data: %v", err)
    return
}

if response == nil {
    log.Println("No data returned")
    return
}

// Utiliser les données
fmt.Printf("Result: %+v\n", response)
```

### Gestion des erreurs

```go
response, finishReason, err := agent.GenerateStructuredData(messages)
if err != nil {
    // Erreur de communication ou de parsing
    log.Printf("Generation failed: %v", err)
    return
}

if finishReason != "stop" {
    // La génération s'est arrêtée prématurément
    log.Printf("Unexpected finish reason: %s", finishReason)
}
```

## Configuration de l'historique

### Sans historique (par défaut)

```go
agent, _ := structured.NewAgent[Output](
    ctx,
    agents.Config{
        EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions:      "Instructions...",
        KeepConversationHistory: false, // Chaque appel est indépendant
    },
    modelConfig,
)
```

### Avec historique

```go
agent, _ := structured.NewAgent[Output](
    ctx,
    agents.Config{
        EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions:      "Instructions...",
        KeepConversationHistory: true, // Maintient le contexte
    },
    modelConfig,
)

// Premier appel
response1, _, _ := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Question 1"},
})

// Deuxième appel - a accès au contexte du premier
response2, _, _ := agent.GenerateStructuredData([]messages.Message{
    {Role: roles.User, Content: "Question 2 basée sur la réponse 1"},
})
```
