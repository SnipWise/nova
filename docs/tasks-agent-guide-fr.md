# Guide du Tasks Agent

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Comprendre le type de sortie Plan](#3-comprendre-le-type-de-sortie-plan)
4. [Configuration de l'agent](#4-configuration-de-lagent)
5. [Configuration du modèle](#5-configuration-du-modèle)
6. [Identifier des plans](#6-identifier-des-plans)
7. [Historique de conversation et messages](#7-historique-de-conversation-et-messages)
8. [Hooks de cycle de vie (TasksAgentOption)](#8-hooks-de-cycle-de-vie-tasksagentoption)
9. [Gestion du contexte et de l'état](#9-gestion-du-contexte-et-de-létat)
10. [Export JSON et débogage](#10-export-json-et-débogage)
11. [Référence API](#11-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Tasks Agent ?

Le `tasks.Agent` est un agent structuré spécialisé fourni par le SDK Nova (`github.com/snipwise/nova`) qui identifie et extrait des plans de tâches structurés à partir d'une entrée en langage naturel. Il utilise le type `agents.Plan` comme sortie, convertissant les descriptions utilisateur en listes de tâches organisées avec des responsabilités claires.

Contrairement à un agent de chat général, le tasks agent se concentre sur la compréhension des exigences de projet et la génération de plans structurés et actionnables. Il est construit sur le framework d'agent structuré, garantissant une sortie type-safe et prévisible.

### Quand utiliser un Tasks Agent

| Scénario | Agent recommandé |
|---|---|
| Décomposer des descriptions de projet en listes de tâches | `tasks.Agent` |
| Extraire des plans structurés depuis des notes de réunion ou des exigences | `tasks.Agent` |
| Convertir des objectifs en langage naturel en tâches actionnables | `tasks.Agent` |
| Générer des listes de tâches organisées pour les projets | `tasks.Agent` |
| Extraction ou classification de texte général | `structured.Agent[VotreType]` |
| IA conversationnelle en texte libre | `chat.Agent` |
| Appels de fonctions / utilisation d'outils | `tools.Agent` |

### Capacités principales

- **Identification de plans** : Convertit automatiquement les descriptions en langage naturel en plans structurés avec tâches et responsabilités.
- **Organisation séquentielle** : Génère des listes de tâches ordonnées pour une exécution claire du projet.
- **Sortie type-safe** : Retourne toujours un objet `agents.Plan` correctement structuré.
- **Historique de conversation** : Maintenez optionnellement l'historique pour un raffinement itératif du plan.
- **Hooks de cycle de vie** : Exécutez une logique personnalisée avant et après l'identification du plan.
- **Export JSON** : Exportez l'historique de conversation et les plans pour le débogage ou la persistance.

---

## 2. Démarrage rapide

### Exemple minimal

```go
package main

import (
    "context"
    "fmt"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/tasks"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    agent, err := tasks.NewAgent(
        ctx,
        agents.Config{
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "Tu es un assistant de planification de projet qui décompose les descriptions de projet en plans de tâches structurés.",
        },
        models.Config{
            Name:        "ai/qwen2.5:7B-Q8_0",
            Temperature: models.Float64(0.3),
        },
    )
    if err != nil {
        panic(err)
    }

    plan, err := agent.IdentifyPlanFromText(
        "Créer une application web de gestion de tâches avec authentification utilisateur et tableau de bord.",
    )
    if err != nil {
        panic(err)
    }

    fmt.Println("Plan identifié avec succès !")
    fmt.Printf("Nombre de tâches principales : %d\n", len(plan.Tasks))
}
```

---

## 3. Comprendre le type de sortie Plan

Le tasks agent retourne une structure `agents.Plan`, qui contient une liste de tâches :

### Structure du Plan

```go
type Plan struct {
    Tasks []Task `json:"tasks"`
}

type Task struct {
    ID          string `json:"id"`
    Description string `json:"description"`
    Responsible string `json:"responsible"`
}
```

### Exemple de sortie de plan

```json
{
  "tasks": [
    {
      "id": "1",
      "description": "Configurer l'infrastructure du projet et initialiser le dépôt Git",
      "responsible": "Équipe DevOps"
    },
    {
      "id": "2",
      "description": "Configurer le pipeline CI/CD avec tests automatisés",
      "responsible": "Ingénieur DevOps"
    },
    {
      "id": "3",
      "description": "Implémenter le système d'authentification et d'autorisation utilisateur",
      "responsible": "Équipe Backend"
    },
    {
      "id": "4",
      "description": "Concevoir et développer le tableau de bord frontend",
      "responsible": "Équipe Frontend"
    }
  ]
}
```

### Propriétés des tâches

- **ID** : Identifiant unique pour la tâche (séquentiel : "1", "2", "3", etc.)
- **Description** : Description claire et actionnable de ce qui doit être fait
- **Responsible** : Qui est responsable de compléter la tâche (équipe, rôle ou personne)

---

## 4. Configuration de l'agent

```go
agents.Config{
    Name:                    "tasks-agent",        // Nom de l'agent (optionnel)
    EngineURL:               "http://localhost:12434/engines/llama.cpp/v1", // URL du moteur LLM (requis)
    APIKey:                  "votre-clé-api",       // Clé API (optionnel)
    SystemInstructions:      "Tu es un assistant de planification de projet...", // Prompt système (recommandé)
    KeepConversationHistory: false,                 // Habituellement false pour extraction sans état
}
```

| Champ | Type | Requis | Description |
|---|---|---|---|
| `Name` | `string` | Non | Identifiant de l'agent pour le logging. |
| `EngineURL` | `string` | Oui | URL du moteur LLM compatible OpenAI. |
| `APIKey` | `string` | Non | Clé API pour les moteurs authentifiés. |
| `SystemInstructions` | `string` | Recommandé | Prompt système définissant la tâche de planification. Une bonne valeur par défaut est fournie si omis. |
| `KeepConversationHistory` | `bool` | Non | Habituellement `false` pour extraction de plan sans état. Défaut : `false`. |

### Instructions système recommandées

```go
SystemInstructions: `Tu es un expert en planification de projet. Décompose les descriptions de projet en tâches claires et actionnables avec :
- Des IDs séquentiels uniques (1, 2, 3, etc.)
- Des descriptions spécifiques et actionnables
- Des affectations de responsabilité appropriées (équipe/rôle/personne)
- Une organisation et un ordonnancement logiques des tâches`
```

---

## 5. Configuration du modèle

```go
models.Config{
    Name:        "ai/qwen2.5:7B-Q8_0",    // ID du modèle (requis)
    Temperature: models.Float64(0.3),      // 0.3 pour planification structurée mais légèrement créative
    MaxTokens:   models.Int(4000),          // Permettre des plans plus larges
}
```

### Paramètres recommandés

- **Temperature** : `0.3` - assez structuré pour un formatage cohérent, mais permet une décomposition créative des tâches
- **Modèle** : Utilisez des modèles avec de fortes capacités de suivi d'instructions et JSON (Qwen 2.5 7B+ recommandé)
- **MaxTokens** : 3000-4000+ pour des projets complexes avec de nombreuses tâches

---

## 6. Identifier des plans

### IdentifyPlanFromText

La méthode la plus simple pour extraire un plan depuis du texte :

```go
plan, err := agent.IdentifyPlanFromText(
    "Construire une API REST avec authentification, gestion des utilisateurs, et notifications en temps réel",
)
if err != nil {
    // gérer l'erreur
}

// Accéder aux tâches
for _, task := range plan.Tasks {
    fmt.Printf("[%s] %s\n", task.ID, task.Description)
    fmt.Printf("     Responsable : %s\n\n", task.Responsible)
}
```

### IdentifyPlan

Pour plus de contrôle, utilisez la méthode complète avec des messages :

```go
userMessages := []messages.Message{
    {
        Role:    roles.User,
        Content: "Créer une application mobile de suivi des dépenses",
    },
}

plan, finishReason, err := agent.IdentifyPlan(userMessages)
if err != nil {
    // gérer l'erreur
}

fmt.Println("Raison de fin :", finishReason) // "stop"
```

**Valeurs de retour :**
- `plan *agents.Plan` : Le plan de tâches extrait
- `finishReason string` : Raison de l'arrêt de la génération (`"stop"`, `"length"`, etc.)
- `err error` : Erreur si la génération ou le parsing a échoué

### Raffinement multi-tour du plan

```go
// Planification initiale
plan, _, err := agent.IdentifyPlan([]messages.Message{
    {Role: roles.User, Content: "Construire une plateforme e-commerce"},
})

// Ajouter des retours et raffiner
agent.AddMessage(roles.User, "Ajouter des tâches pour l'intégration de paiement et la gestion des stocks")

// Générer un plan raffiné
refinedPlan, _, err := agent.IdentifyPlan([]messages.Message{
    {Role: roles.User, Content: "Raffine le plan avec les nouvelles exigences"},
})
```

---

## 7. Historique de conversation et messages

### Gestion des messages

```go
// Obtenir tous les messages de l'historique
msgs := agent.GetMessages()

// Ajouter un message
agent.AddMessage(roles.User, "Ajouter des tâches d'audit de sécurité")

// Ajouter plusieurs messages d'un coup
agent.AddMessages([]messages.Message{
    {Role: roles.User, Content: "Considérer la scalabilité"},
    {Role: roles.Assistant, Content: "..."},
})

// Effacer tous les messages sauf l'instruction système
agent.ResetMessages()
```

---

## 8. Hooks de cycle de vie (TasksAgentOption)

Les hooks de cycle de vie permettent d'exécuter une logique personnalisée avant et après l'identification du plan.

### TasksAgentOption

```go
type TasksAgentOption func(*Agent)
```

Les options sont passées en arguments variadiques à `NewAgent` :

```go
agent, err := tasks.NewAgent(ctx, agentConfig, modelConfig,
    tasks.BeforeCompletion(fn),
    tasks.AfterCompletion(fn),
)
```

### BeforeCompletion

Appelé avant chaque identification de plan :

```go
tasks.BeforeCompletion(func(a *tasks.Agent) {
    fmt.Println("Sur le point d'identifier le plan...")
    fmt.Printf("Nombre de messages : %d\n", len(a.GetMessages()))
})
```

**Cas d'usage :**
- Logging et monitoring
- Collecte de métriques
- Inspection de l'état pré-identification

### AfterCompletion

Appelé après chaque identification de plan :

```go
tasks.AfterCompletion(func(a *tasks.Agent) {
    fmt.Println("Identification du plan terminée.")
    fmt.Printf("Nombre de messages : %d\n", len(a.GetMessages()))
})
```

**Cas d'usage :**
- Logging des résultats d'identification
- Métriques post-identification
- Déclenchement d'actions en aval (ex: stockage des plans en base de données)
- Audit/suivi

### Exemple complet avec hooks

```go
agent, err := tasks.NewAgent(
    ctx,
    agents.Config{
        EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
        SystemInstructions: "Tu es un assistant de planification de projet.",
    },
    models.Config{
        Name:        "ai/qwen2.5:7B-Q8_0",
        Temperature: models.Float64(0.3),
    },
    tasks.BeforeCompletion(func(a *tasks.Agent) {
        fmt.Printf("[AVANT] Agent : %s, Messages : %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
    tasks.AfterCompletion(func(a *tasks.Agent) {
        fmt.Printf("[APRÈS] Agent : %s, Messages : %d\n",
            a.GetName(), len(a.GetMessages()))
    }),
)
```

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
agent.Kind()       // Retourne agents.Tasks
agent.GetName()    // Retourne le nom de l'agent depuis la config
agent.GetModelID() // Retourne le nom du modèle depuis la config modèle
```

---

## 10. Export JSON et débogage

### JSON brut requête/réponse

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
    opts ...TasksAgentOption,
) (*Agent, error)
```

Crée un nouvel agent de tâches. Le paramètre `opts` accepte zéro ou plusieurs options fonctionnelles `TasksAgentOption`.

---

### Types

```go
// TasksAgentOption est une option fonctionnelle pour configurer un Agent lors de sa création
type TasksAgentOption func(*Agent)
```

---

### Fonctions d'options

| Fonction | Description |
|---|---|
| `BeforeCompletion(fn func(*Agent))` | Définit un hook appelé avant chaque identification de plan. |
| `AfterCompletion(fn func(*Agent))` | Définit un hook appelé après chaque identification de plan. |

---

### Méthodes

| Méthode | Description |
|---|---|
| `IdentifyPlanFromText(text string) (*agents.Plan, error)` | Identifie un plan depuis une simple description texte. |
| `IdentifyPlan(userMessages []messages.Message) (*agents.Plan, string, error)` | Identifie un plan depuis des messages. Retourne le plan, la raison de fin et l'erreur. |
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
| `Kind() agents.Kind` | Retourne `agents.Tasks`. |
| `GetName() string` | Retourne le nom de l'agent. |
| `GetModelID() string` | Retourne le nom du modèle. |
