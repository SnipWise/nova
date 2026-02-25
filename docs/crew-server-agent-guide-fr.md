# Guide du Crew Server Agent

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent (Options)](#3-configuration-de-lagent-options)
4. [Gestion de l'équipe (Crew)](#4-gestion-de-léquipe-crew)
5. [Pipeline HTTP de complétion (handleCompletion)](#5-pipeline-http-de-complétion-handlecompletion)
6. [Serveur HTTP et routes](#6-serveur-http-et-routes)
7. [Routage intelligent (Orchestrateur)](#7-routage-intelligent-orchestrateur)
8. [Intégration des appels de fonctions (Tool Calls)](#8-intégration-des-appels-de-fonctions-tool-calls)
9. [Intégration RAG](#9-intégration-rag)
10. [Compression du contexte](#10-compression-du-contexte)
11. [Hooks de cycle de vie (BeforeCompletion / AfterCompletion)](#11-hooks-de-cycle-de-vie-beforecompletion--aftercompletion)
12. [Méthodes de complétion directes](#12-méthodes-de-complétion-directes)
13. [Gestion de la conversation](#13-gestion-de-la-conversation)
14. [Commandes spéciales](#14-commandes-spéciales)
15. [Référence API](#15-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Crew Server Agent ?

Le `crewserver.CrewServerAgent` est un agent composite de haut niveau fourni par le SDK Nova (`github.com/snipwise/nova`) qui combine une **équipe de plusieurs agents de chat** avec un **serveur HTTP** exposant des endpoints SSE (Server-Sent Events) en streaming. Il étend le `BaseServerAgent` avec des fonctionnalités spécifiques aux équipes : routage multi-agents, gestion des appels de fonctions avec confirmation web, injection de contexte RAG, compression du contexte et routage intelligent.

### Quand utiliser un Crew Server Agent

| Scénario | Agent recommandé |
|---|---|
| Serveur HTTP avec plusieurs agents spécialisés et routage par sujet | `crewserver.CrewServerAgent` |
| Confirmation web des appels de fonctions avec streaming SSE | `crewserver.CrewServerAgent` |
| API HTTP avec pipeline complet : outils + RAG + compression + routage | `crewserver.CrewServerAgent` |
| Pipeline multi-agents en CLI uniquement (pas de HTTP) | `crew.CrewAgent` |
| Serveur HTTP simple avec un seul agent | `server.ServerAgent` |
| Accès direct simple au LLM | `chat.Agent` |

### Capacités clés

- **Serveur HTTP avec streaming SSE** : Serveur HTTP intégré avec support CORS et streaming temps réel.
- **Équipe multi-agents** : Gestion de plusieurs instances `chat.Agent`, chacune spécialisée pour un sujet.
- **Routage intelligent** : Routage automatique des questions vers l'agent le plus approprié via un orchestrateur.
- **Pipeline complet** : Compression du contexte, appels de fonctions (avec confirmation web), injection RAG et complétion en streaming.
- **Gestion dynamique de l'équipe** : Ajout ou suppression d'agents à la volée.
- **Hooks de cycle de vie** : Exécution de logique personnalisée avant et après chaque requête de complétion HTTP.
- **Appels de fonctions parallèles** : Support de l'exécution séquentielle et parallèle des appels de fonctions.
- **Pattern d'options fonctionnelles** : Configurable via les fonctions `CrewServerAgentOption`.

---

## 2. Démarrage rapide

### Exemple minimal avec un seul agent

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/crewserver"
    "github.com/snipwise/nova/nova-sdk/models"
)

func main() {
    ctx := context.Background()

    chatAgent, _ := chat.NewAgent(ctx,
        agents.Config{
            Name:               "assistant",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "You are a helpful assistant.",
        },
        models.Config{
            Name:        "my-model",
            Temperature: models.Float64(0.4),
        },
    )

    crewServerAgent, _ := crewserver.NewAgent(ctx,
        crewserver.WithSingleAgent(chatAgent),
        crewserver.WithPort(3500),
    )

    fmt.Printf("Démarrage du serveur sur http://localhost%s\n", crewServerAgent.GetPort())
    log.Fatal(crewServerAgent.StartServer())
}
```

### Exemple avec plusieurs agents

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "generic": genericAgent,
}

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithPort(9090),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
        switch strings.ToLower(topic) {
        case "coding", "programming":
            return "coder"
        case "philosophy", "thinking":
            return "thinker"
        default:
            return "generic"
        }
    }),
)

log.Fatal(crewServerAgent.StartServer())
```

---

## 3. Configuration de l'agent (Options)

Les options sont passées en arguments variadiques à `NewAgent` :

```go
crewServerAgent, err := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithPort(9090),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithRagAgent(ragAgent),
    crewserver.WithCompressorAgent(compressorAgent),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.BeforeCompletion(beforeFn),
    crewserver.AfterCompletion(afterFn),
)
```

| Option | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Définit l'équipe d'agents et l'agent initialement sélectionné. **Obligatoire** (ou `WithSingleAgent`). |
| `WithSingleAgent(chatAgent)` | Crée une équipe avec un seul agent (ID : `"single"`). **Obligatoire** (ou `WithAgentCrew`). |
| `WithPort(port)` | Définit le port du serveur HTTP en int (défaut : `3500`). |
| `WithMatchAgentIdToTopicFn(fn)` | Définit la fonction de correspondance entre sujets détectés et IDs d'agents. |
| `WithExecuteFn(fn)` | Définit la fonction d'exécution pour les appels de fonctions. |
| `WithConfirmationPromptFn(fn)` | Définit une fonction de confirmation personnalisée pour les appels de fonctions (remplace la confirmation web). |
| `WithTLSCert(certData, keyData []byte)` | Active HTTPS avec des données de certificat et clé PEM en mémoire. |
| `WithTLSCertFromFile(certPath, keyPath string)` | Active HTTPS avec les chemins vers les fichiers de certificat et clé. |
| `WithToolsAgent(toolsAgent)` | Attache un agent d'outils pour les appels de fonctions. |
| `WithTasksAgent(tasksAgent)` | Attache un agent de tâches pour la planification et l'orchestration de tâches. |
| `WithCompressorAgent(compressorAgent)` | Attache un agent de compression pour la compression du contexte. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attache un compresseur avec une limite de taille du contexte. |
| `WithRagAgent(ragAgent)` | Attache un agent RAG pour la recherche de documents. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attache un agent RAG avec configuration de similarité. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attache un agent orchestrateur pour la détection de sujet et le routage. |
| `BeforeCompletion(fn)` | Définit un hook appelé avant chaque appel `handleCompletion`. |
| `AfterCompletion(fn)` | Définit un hook appelé après chaque appel `handleCompletion`. |

### Valeurs par défaut

| Paramètre | Défaut |
|---|---|
| Port | `:3500` |
| `SimilarityLimit` | `0.6` (hérité de `BaseServerAgent`) |
| `MaxSimilarities` | `3` (hérité de `BaseServerAgent`) |
| `ContextSizeLimit` | `8000` (hérité de `BaseServerAgent`) |

### Support HTTPS

Le Crew Server Agent supporte HTTPS pour une communication sécurisée. Lorsque des certificats TLS sont fournis, le serveur utilisera automatiquement HTTPS au lieu de HTTP.

```go
// Option 1 : Utiliser des fichiers de certificats (recommandé)
crewServerAgent, err := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithPort(443),
    crewserver.WithTLSCertFromFile("server.crt", "server.key"),
)

// Option 2 : Utiliser des données de certificat en mémoire
certData, _ := os.ReadFile("server.crt")
keyData, _ := os.ReadFile("server.key")

crewServerAgent, err := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithPort(443),
    crewserver.WithTLSCert(certData, keyData),
)
```

**Notes importantes** :
- HTTPS est **optionnel** - sans certificats TLS, le serveur fonctionne en HTTP (rétrocompatible)
- Pour la production, utilisez des certificats d'une autorité de certification de confiance (ex : Let's Encrypt)
- Voir `/samples/90-https-server-example` pour un exemple complet

---

## 4. Gestion de l'équipe (Crew)

### Équipe statique (à la création)

```go
crewserver.WithAgentCrew(map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
}, "coder")
```

### Gestion dynamique de l'équipe

```go
// Ajouter un agent à la volée
err := crewServerAgent.AddChatAgentToCrew("cook", cookAgent)

// Supprimer un agent (impossible de supprimer l'agent actif)
err := crewServerAgent.RemoveChatAgentFromCrew("thinker")

// Obtenir tous les agents
agents := crewServerAgent.GetChatAgents()

// Remplacer toute l'équipe
crewServerAgent.SetChatAgents(newCrew)
```

### Changer d'agent manuellement

```go
// Obtenir l'agent actuellement sélectionné
id := crewServerAgent.GetSelectedAgentId()

// Basculer vers un autre agent
err := crewServerAgent.SetSelectedAgentId("coder")
```

**Note :** Un seul agent est actif à la fois. `GetName()`, `GetModelID()`, `GetMessages()`, etc. opèrent tous sur l'agent actuellement actif.

---

## 5. Pipeline HTTP de complétion (handleCompletion)

Le handler HTTP `handleCompletion` est le point d'entrée principal pour les requêtes de complétion. Il traite les requêtes `POST /completion` avec streaming SSE.

### Étapes du pipeline

1. **Hook BeforeCompletion** (si défini)
2. **Parsing de la requête** (extraction de la question du corps JSON)
3. **Configuration du streaming SSE** (headers et flusher)
4. **Compression du contexte** (si compresseur configuré et contexte dépassant la limite, avec notifications SSE)
5. **Configuration du canal de notifications** (pour les notifications d'appels de fonctions)
6. **Détection et exécution des appels de fonctions** (si agent d'outils configuré, avec confirmation web)
7. **Fermeture du canal de notifications**
8. **Génération de la complétion en streaming** (si nécessaire : RAG + routage + stream)
9. **Nettoyage de l'état des outils**
10. **Hook AfterCompletion** (si défini)

### Format de la requête

```json
POST /completion
{
    "data": {
        "message": "Votre question ici"
    }
}
```

### Format de la réponse SSE

```
data: {"message": "morceau de texte"}

data: {"message": "", "finish_reason": "stop"}
```

---

## 6. Serveur HTTP et routes

### Démarrage du serveur

```go
err := crewServerAgent.StartServer()
```

### Routes disponibles

| Méthode | Chemin | Description |
|---|---|---|
| `POST` | `/completion` | Générer une complétion en streaming (SSE) |
| `POST` | `/completion/stop` | Arrêter l'opération de streaming en cours |
| `POST` | `/memory/reset` | Réinitialiser l'historique de conversation |
| `GET` | `/memory/messages/list` | Lister tous les messages de la conversation |
| `GET` | `/memory/messages/context-size` | Obtenir la taille actuelle du contexte |
| `POST` | `/operation/validate` | Valider une opération d'appel de fonction en attente |
| `POST` | `/operation/cancel` | Annuler une opération d'appel de fonction en attente |
| `POST` | `/operation/reset` | Réinitialiser les opérations en attente |
| `GET` | `/models` | Obtenir les informations du modèle |
| `GET` | `/health` | Vérification de santé |
| `GET` | `/current-agent` | Obtenir les informations de l'agent actuel (ID, nom, modèle) |

### Routes personnalisées

Le champ `Mux` est exposé après l'initialisation par `StartServer`, permettant d'ajouter des routes personnalisées.

### CORS

Toutes les réponses incluent des headers CORS autorisant toutes les origines. Les requêtes de pré-vol `OPTIONS` sont gérées automatiquement.

---

## 7. Routage intelligent (Orchestrateur)

Lorsqu'un agent orchestrateur est attaché, le crew server agent peut automatiquement router les questions vers l'agent spécialisé le plus approprié.

### Configuration automatique du routage (Recommandé)

Lorsque vous utilisez `WithOrchestratorAgent`, le crew server agent **configure automatiquement le routage** en utilisant la méthode `GetAgentForTopic` de l'orchestrateur. Vous n'avez pas besoin de fournir `WithMatchAgentIdToTopicFn` sauf si vous avez une logique de routage personnalisée.

**Option 1 : Orchestrateur avec configuration de routage intégrée**

```go
// Charger la configuration de routage depuis un fichier JSON
routingConfig, _ := loadRoutingConfig("agent-routing.json")

orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:               "orchestrator",
        EngineURL:          engineURL,
        SystemInstructions: `Identifiez le sujet principal en un seul mot.
            Sujets possibles : Technology, Philosophy, Cooking, etc.
            Répondez en JSON avec le champ 'topic_discussion'.`,
    },
    models.Config{Name: "my-model", Temperature: models.Float64(0.0)},
    orchestrator.WithRoutingConfig(*routingConfig), // Configurer le routage dans l'orchestrateur
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithOrchestratorAgent(orchestratorAgent), // Configure automatiquement le routage avec GetAgentForTopic
)
```

**Format de la configuration de routage (agent-routing.json) :**

```json
{
    "routing": [
        {
            "topics": ["coding", "programming", "development"],
            "agent": "coder"
        },
        {
            "topics": ["cooking", "food", "recipe"],
            "agent": "cook"
        }
    ],
    "default_agent": "generic"
}
```

**Option 2 : Orchestrateur avec fonction de correspondance personnalisée**

Si vous avez besoin d'une logique de routage personnalisée au-delà d'une simple correspondance de sujets, vous pouvez toujours fournir `WithMatchAgentIdToTopicFn` :

```go
orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:               "orchestrator",
        EngineURL:          engineURL,
        SystemInstructions: `Identifiez le sujet principal en un seul mot...`,
    },
    models.Config{Name: "my-model", Temperature: models.Float64(0.0)},
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    // Remplacer la configuration automatique par une logique personnalisée
    crewserver.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
        // Logique de routage personnalisée avec conditions supplémentaires
        if currentAgentId == "coder" && strings.ToLower(topic) == "philosophy" {
            // Continuer avec coder si déjà en train de coder
            return currentAgentId
        }
        switch strings.ToLower(topic) {
        case "coding", "programming":
            return "coder"
        case "cooking", "food":
            return "cook"
        default:
            return "generic"
        }
    }),
)
```

### Fonctionnement

1. L'orchestrateur analyse la question de l'utilisateur et détecte le sujet en utilisant `IdentifyIntent` ou `IdentifyTopicFromText`.
2. **Si aucune `matchAgentIdToTopicFn` personnalisée n'est fournie** : Le crew server agent appelle automatiquement `orchestratorAgent.GetAgentForTopic(topic)` pour obtenir l'ID de l'agent.
3. **Si une `matchAgentIdToTopicFn` personnalisée est fournie** : Le crew server agent utilise votre fonction personnalisée à la place.
4. Le crew server agent bascule vers l'agent correspondant s'il est différent de l'agent actuel.
5. La complétion est générée par l'agent nouvellement sélectionné.

### Détection directe du sujet

```go
agentId, err := crewServerAgent.DetectTopicThenGetAgentId("Écris une fonction Python")
// agentId = "coder"
```

---

## 8. Intégration des appels de fonctions (Tool Calls)

Le crew server agent supporte deux modes de confirmation des appels de fonctions :

### Confirmation web (par défaut)

Lorsqu'aucune `confirmationPromptFn` personnalisée n'est fournie, les appels de fonctions déclenchent un flux de confirmation web via des notifications SSE :

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(func(name string, args string) (string, error) {
        return `{"result": "ok"}`, nil
    }),
)
```

Le client web reçoit les notifications d'appels de fonctions via SSE et peut valider ou annuler les opérations via :
- `POST /operation/validate` - Approuver l'appel de fonction
- `POST /operation/cancel` - Rejeter l'appel de fonction

### Fonction de confirmation personnalisée

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(executeFn),
    crewserver.WithConfirmationPromptFn(func(name string, args string) tools.ConfirmationResponse {
        return tools.Confirm // Toujours confirmer
    }),
)
```

### Appels de fonctions parallèles

Lorsque l'agent d'outils est configuré avec `ParallelToolCalls: true`, le crew server agent utilise automatiquement les méthodes de détection parallèle :

```go
toolsAgent, _ := tools.NewAgent(ctx, toolsConfig,
    models.Config{
        Name:              "my-model",
        ParallelToolCalls: models.Bool(true),
    },
    tools.WithTools(myTools),
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithToolsAgent(toolsAgent),
    crewserver.WithExecuteFn(executeFn),
)
```

---

## 9. Intégration RAG

```go
ragAgent, _ := rag.NewAgent(ctx, ragConfig, ragModelConfig)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 5),
)
```

Pendant le pipeline de complétion, le crew server agent effectue une recherche de similarité et injecte le contexte pertinent dans la conversation avant de générer la complétion.

---

## 10. Compression du contexte

```go
compressorAgent, _ := compressor.NewAgent(ctx, compressorConfig, compressorModelConfig,
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)
```

Au début de chaque requête de complétion, le contexte est compressé s'il dépasse la limite configurée. Le client reçoit des notifications SSE sur le processus de compression.

### Compression manuelle

```go
// Compresser uniquement si au-dessus de la limite
newSize, err := crewServerAgent.CompressChatAgentContextIfOverLimit()

// Forcer la compression
newSize, err := crewServerAgent.CompressChatAgentContext()
```

---

## 11. Hooks de cycle de vie (BeforeCompletion / AfterCompletion)

Les hooks de cycle de vie permettent d'exécuter de la logique personnalisée avant et après chaque requête de complétion HTTP (`POST /completion`). Ils sont configurés comme options fonctionnelles `CrewServerAgentOption`.

### BeforeCompletion

Appelé au tout début de chaque handler HTTP `handleCompletion`, avant le parsing de la requête. Le hook reçoit une référence vers le crew server agent.

```go
crewserver.BeforeCompletion(func(a *crewserver.CrewServerAgent) {
    fmt.Printf("[AVANT] Agent : %s\n", a.GetName())
})
```

### AfterCompletion

Appelé à la toute fin de chaque handler HTTP `handleCompletion`, après le nettoyage. Le hook reçoit une référence vers le crew server agent.

```go
crewserver.AfterCompletion(func(a *crewserver.CrewServerAgent) {
    fmt.Printf("[APRÈS] Agent : %s\n", a.GetName())
})
```

### Placement des hooks

| Méthode | Hooks déclenchés |
|---|---|
| `POST /completion` (handleCompletion) | Oui |
| `GenerateCompletion` | Non (délègue à l'agent de chat actif) |
| `GenerateStreamCompletion` | Non (délègue à l'agent de chat actif) |
| `GenerateCompletionWithReasoning` | Non (délègue à l'agent de chat actif) |
| `GenerateStreamCompletionWithReasoning` | Non (délègue à l'agent de chat actif) |

Les hooks sont dans `handleCompletion`, qui est le pipeline de complétion HTTP. Les méthodes `Generate*` délèguent directement à l'agent `chat.Agent` actif et ne déclenchent pas les hooks du niveau crew.

### Exemple complet

```go
callCount := 0

crewServerAgent, _ := crewserver.NewAgent(ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
    crewserver.BeforeCompletion(func(a *crewserver.CrewServerAgent) {
        callCount++
        fmt.Printf("[AVANT] Appel #%d - Agent : %s\n", callCount, a.GetName())
    }),
    crewserver.AfterCompletion(func(a *crewserver.CrewServerAgent) {
        fmt.Printf("[APRÈS] Appel #%d - Agent : %s\n", callCount, a.GetName())
    }),
)

log.Fatal(crewServerAgent.StartServer())
```

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Le code existant sans hooks continue de fonctionner sans aucune modification.

---

## 12. Méthodes de complétion directes

Le crew server agent expose des méthodes de complétion directes qui délèguent à l'agent `chat.Agent` actuellement actif :

```go
// Sans streaming
result, err := crewServerAgent.GenerateCompletion(userMessages)

// Avec streaming
result, err := crewServerAgent.GenerateStreamCompletion(userMessages, callback)

// Avec raisonnement
result, err := crewServerAgent.GenerateCompletionWithReasoning(userMessages)
result, err := crewServerAgent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCb, responseCb)
```

**Note :** Ces méthodes contournent le pipeline HTTP complet (pas de compression, pas d'appels de fonctions, pas de RAG, pas de routage, pas de SSE). Elles délèguent directement à l'agent de chat actif. Les hooks de cycle de vie ne sont **pas** déclenchés.

---

## 13. Gestion de la conversation

Toutes les méthodes de conversation opèrent sur l'agent de chat **actuellement actif** :

```go
// Obtenir les messages
msgs := crewServerAgent.GetMessages()

// Obtenir la taille du contexte
size := crewServerAgent.GetContextSize()

// Réinitialiser la conversation
crewServerAgent.ResetMessages()

// Ajouter un message
crewServerAgent.AddMessage(roles.User, "Bonjour")

// Exporter en JSON
jsonStr, err := crewServerAgent.ExportMessagesToJSON()

// Arrêter le streaming
crewServerAgent.StopStream()
```

---

## 14. Commandes spéciales

Le crew server agent supporte des préfixes de commandes spéciales dans la question :

### Sélectionner un agent

Envoyez un message préfixé par `[select-agent <id>]` pour basculer manuellement l'agent actif :

```
[select-agent coder]
```

### Lister les agents

Envoyez `[agent-list]` pour obtenir la liste de tous les agents disponibles :

```
[agent-list]
```

Ces commandes sont traitées avant le flux de complétion standard et retournent leurs résultats via SSE.

---

## 15. Référence API

### Constructeur

```go
func NewAgent(ctx context.Context, options ...CrewServerAgentOption) (*CrewServerAgent, error)
```

### Types

```go
type CrewServerAgentOption func(*CrewServerAgent) error
```

### Fonctions d'option

| Fonction | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Définit l'équipe et l'agent initial. |
| `WithSingleAgent(chatAgent)` | Crée une équipe à agent unique. |
| `WithPort(port)` | Définit le port du serveur HTTP (défaut : 3500). |
| `WithMatchAgentIdToTopicFn(fn)` | Définit la fonction de correspondance sujet-agent. |
| `WithExecuteFn(fn)` | Définit la fonction d'exécution des outils. |
| `WithConfirmationPromptFn(fn)` | Définit la fonction de confirmation personnalisée des outils. |
| `WithToolsAgent(toolsAgent)` | Attache un agent d'outils. |
| `WithTasksAgent(tasksAgent)` | Attache un agent de tâches pour la planification et l'orchestration. |
| `WithCompressorAgent(compressorAgent)` | Attache un agent de compression. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attache un compresseur avec limite. |
| `WithRagAgent(ragAgent)` | Attache un agent RAG. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attache un RAG avec configuration. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attache un agent orchestrateur. |
| `BeforeCompletion(fn func(*CrewServerAgent))` | Définit le hook avant chaque handleCompletion. |
| `AfterCompletion(fn func(*CrewServerAgent))` | Définit le hook après chaque handleCompletion. |

### Méthodes

| Méthode | Description |
|---|---|
| `StartServer() error` | Démarre le serveur HTTP avec toutes les routes. |
| `SetPort(port)` | Définit le port HTTP. |
| `GetPort() string` | Obtient le port HTTP. |
| `GenerateCompletion(msgs) (*chat.CompletionResult, error)` | Complétion directe (délègue à l'agent actif). |
| `GenerateStreamCompletion(msgs, callback) (*chat.CompletionResult, error)` | Streaming direct (délègue à l'agent actif). |
| `GenerateCompletionWithReasoning(msgs) (*chat.ReasoningResult, error)` | Complétion directe avec raisonnement. |
| `GenerateStreamCompletionWithReasoning(msgs, reasoningCb, responseCb) (*chat.ReasoningResult, error)` | Streaming direct avec raisonnement. |
| `StopStream()` | Arrête l'opération de streaming en cours. |
| `GetMessages() []messages.Message` | Obtient les messages de l'agent actif. |
| `GetContextSize() int` | Obtient la taille du contexte de l'agent actif. |
| `ResetMessages()` | Réinitialise la conversation de l'agent actif. |
| `AddMessage(role, content)` | Ajoute un message à l'agent actif. |
| `ExportMessagesToJSON() (string, error)` | Exporte la conversation de l'agent actif. |
| `GetChatAgents() map[string]*chat.Agent` | Obtient tous les agents de l'équipe. |
| `SetChatAgents(crew)` | Remplace toute l'équipe. |
| `AddChatAgentToCrew(id, agent) error` | Ajoute un agent à l'équipe. |
| `RemoveChatAgentFromCrew(id) error` | Supprime un agent de l'équipe. |
| `GetSelectedAgentId() string` | Obtient l'ID de l'agent actif. |
| `SetSelectedAgentId(id) error` | Change d'agent actif. |
| `DetectTopicThenGetAgentId(query) (string, error)` | Détecte le sujet et retourne l'ID de l'agent correspondant. |
| `SetOrchestratorAgent(orchestratorAgent)` | Définit l'agent orchestrateur. |
| `GetOrchestratorAgent() OrchestratorAgent` | Obtient l'agent orchestrateur. |
| `SetToolsAgent(toolsAgent)` | Définit l'agent d'outils. |
| `GetToolsAgent() *tools.Agent` | Obtient l'agent d'outils. |
| `SetExecuteFunction(fn)` | Définit la fonction d'exécution des outils. |
| `SetConfirmationPromptFn(fn)` | Définit la fonction de confirmation des outils. |
| `GetConfirmationPromptFn() func(...)` | Obtient la fonction de confirmation des outils. |
| `SetRagAgent(ragAgent)` | Définit l'agent RAG. |
| `GetRagAgent() *rag.Agent` | Obtient l'agent RAG. |
| `SetSimilarityLimit(limit)` | Définit le seuil de similarité. |
| `GetSimilarityLimit() float64` | Obtient le seuil de similarité. |
| `SetMaxSimilarities(n)` | Définit le nombre maximum de similarités. |
| `GetMaxSimilarities() int` | Obtient le nombre maximum de similarités. |
| `SetCompressorAgent(compressorAgent)` | Définit l'agent de compression. |
| `GetCompressorAgent() *compressor.Agent` | Obtient l'agent de compression. |
| `SetContextSizeLimit(limit)` | Définit la limite de taille du contexte. |
| `GetContextSizeLimit() int` | Obtient la limite de taille du contexte. |
| `CompressChatAgentContextIfOverLimit() (int, error)` | Compresse si au-dessus de la limite. |
| `CompressChatAgentContext() (int, error)` | Force la compression. |
| `Kind() agents.Kind` | Retourne `agents.ChatServer`. |
| `GetName() string` | Retourne le nom de l'agent actif. |
| `GetModelID() string` | Retourne l'ID du modèle de l'agent actif. |
