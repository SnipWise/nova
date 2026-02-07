# Guide du Gateway Server Agent

## Table des matières

1. [Introduction](#1-introduction)
2. [Démarrage rapide](#2-démarrage-rapide)
3. [Configuration de l'agent (Options)](#3-configuration-de-lagent-options)
4. [Gestion de l'équipe (Crew)](#4-gestion-de-léquipe-crew)
5. [Pipeline HTTP de complétion (handleChatCompletions)](#5-pipeline-http-de-complétion-handlechatcompletions)
6. [Serveur HTTP et routes](#6-serveur-http-et-routes)
7. [Modes d'exécution des outils (Tool Modes)](#7-modes-dexécution-des-outils-tool-modes)
8. [Routage intelligent (Orchestrateur)](#8-routage-intelligent-orchestrateur)
9. [Intégration RAG](#9-intégration-rag)
10. [Compression du contexte](#10-compression-du-contexte)
11. [Hooks de cycle de vie (BeforeCompletion / AfterCompletion)](#11-hooks-de-cycle-de-vie-beforecompletion--aftercompletion)
12. [Gestion de la conversation](#12-gestion-de-la-conversation)
13. [Types compatibles OpenAI](#13-types-compatibles-openai)
14. [Tests](#14-tests)
15. [Référence API](#15-référence-api)

---

## 1. Introduction

### Qu'est-ce qu'un Gateway Server Agent ?

Le `gatewayserver.GatewayServerAgent` est un agent composite de haut niveau fourni par le SDK Nova (`github.com/snipwise/nova`) qui expose une **API HTTP compatible OpenAI** (`POST /v1/chat/completions`) adossée à une **équipe d'agents N.O.V.A.**. Les clients externes (comme `qwen-code`, `aider`, `continue.dev`, ou tout SDK OpenAI) voient un seul "modèle", tandis qu'en interne la gateway route les requêtes vers des agents spécialisés.

Contrairement au `crewserver.CrewServerAgent` qui utilise un protocole SSE personnalisé, le Gateway Server Agent parle le **format standard de l'API OpenAI Chat Completions**, ce qui en fait un remplacement direct de l'API OpenAI.

### Quand utiliser un Gateway Server Agent

| Scénario | Agent recommandé |
|---|---|
| API compatible OpenAI pour outils externes (qwen-code, aider, etc.) | `gatewayserver.GatewayServerAgent` |
| Passthrough des tool_calls au client (le client gère l'exécution) | `gatewayserver.GatewayServerAgent` avec `ToolModePassthrough` |
| Exécution des outils côté serveur avec format API OpenAI | `gatewayserver.GatewayServerAgent` avec `ToolModeAutoExecute` |
| Protocole SSE personnalisé avec confirmation web des outils | `crewserver.CrewServerAgent` |
| Pipeline multi-agents en CLI uniquement (pas de HTTP) | `crew.CrewAgent` |
| Accès direct simple au LLM | `chat.Agent` |

### Capacités clés

- **API compatible OpenAI** : Support complet de `POST /v1/chat/completions` (streaming SSE + JSON non-streaming).
- **Deux modes d'outils** : Passthrough (le client exécute les outils) et auto-execute (le serveur exécute les outils).
- **Équipe multi-agents** : Gestion de plusieurs instances `chat.Agent`, chacune spécialisée pour un sujet.
- **Routage intelligent** : Routage automatique des questions vers l'agent le plus approprié via un orchestrateur.
- **Pipeline complet** : Compression du contexte, appels de fonctions, injection RAG et complétion en streaming.
- **Streaming SSE standard** : Chunks `data: {json}\n\n` + terminateur `data: [DONE]\n\n`.
- **Endpoint des modèles** : `GET /v1/models` liste tous les agents de l'équipe comme modèles disponibles.
- **Hooks de cycle de vie** : Exécution de logique personnalisée avant et après chaque requête de complétion.
- **Pattern d'options fonctionnelles** : Configurable via les fonctions `GatewayServerAgentOption`.

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
    "github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
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
            Temperature: models.Float64(0.7),
        },
    )

    gateway, _ := gatewayserver.NewAgent(ctx,
        gatewayserver.WithSingleAgent(chatAgent),
        gatewayserver.WithPort(8080),
    )

    fmt.Println("Démarrage de la gateway sur http://localhost:8080")
    log.Fatal(gateway.StartServer())
}
```

**Utilisation avec curl :**

```bash
# Non-streaming
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"assistant","messages":[{"role":"user","content":"Bonjour !"}]}'

# Streaming
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"assistant","messages":[{"role":"user","content":"Bonjour !"}],"stream":true}'
```

**Utilisation avec qwen-code :**

```bash
OPENAI_BASE_URL=http://localhost:8080/v1 OPENAI_API_KEY=none qwen-code
```

### Exemple avec plusieurs agents

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "generic": genericAgent,
}

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithAgentCrew(agentCrew, "generic"),
    gatewayserver.WithPort(8080),
    gatewayserver.WithOrchestratorAgent(orchestratorAgent),
    gatewayserver.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
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

log.Fatal(gateway.StartServer())
```

---

## 3. Configuration de l'agent (Options)

Les options sont passées en arguments variadiques à `NewAgent` :

```go
gateway, err := gatewayserver.NewAgent(ctx,
    gatewayserver.WithAgentCrew(agentCrew, "generic"),
    gatewayserver.WithPort(8080),
    gatewayserver.WithToolsAgent(toolsAgent),
    gatewayserver.WithToolMode(gatewayserver.ToolModeAutoExecute),
    gatewayserver.WithExecuteFn(executeFn),
    gatewayserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 5),
    gatewayserver.WithCompressorAgentAndContextSize(compressorAgent, 7000),
    gatewayserver.WithOrchestratorAgent(orchestratorAgent),
    gatewayserver.WithMatchAgentIdToTopicFn(matchFn),
    gatewayserver.BeforeCompletion(beforeFn),
    gatewayserver.AfterCompletion(afterFn),
)
```

| Option | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Définit l'équipe d'agents et l'agent initialement sélectionné. **Obligatoire** (ou `WithSingleAgent`). |
| `WithSingleAgent(chatAgent)` | Crée une équipe avec un seul agent (ID : `"single"`). **Obligatoire** (ou `WithAgentCrew`). |
| `WithPort(port)` | Définit le port du serveur HTTP en int (défaut : `8080`). |
| `WithToolsAgent(toolsAgent)` | Attache un agent d'outils pour les appels de fonctions. |
| `WithToolMode(mode)` | Définit le mode d'exécution des outils : `ToolModePassthrough` (défaut) ou `ToolModeAutoExecute`. |
| `WithExecuteFn(fn)` | Définit la fonction d'exécution pour l'exécution côté serveur des outils. |
| `WithConfirmationPromptFn(fn)` | Définit une fonction de confirmation personnalisée pour les appels de fonctions. |
| `WithTLSCert(certData, keyData []byte)` | Active HTTPS avec des données de certificat et clé PEM en mémoire. |
| `WithTLSCertFromFile(certPath, keyPath string)` | Active HTTPS avec les chemins vers les fichiers de certificat et clé. |
| `WithMatchAgentIdToTopicFn(fn)` | Définit la fonction de correspondance entre sujets détectés et IDs d'agents. |
| `WithRagAgent(ragAgent)` | Attache un agent RAG pour la recherche de documents. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attache un agent RAG avec configuration de similarité. |
| `WithCompressorAgent(compressorAgent)` | Attache un agent de compression pour la compression du contexte. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attache un compresseur avec une limite de taille du contexte. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attache un agent orchestrateur pour la détection de sujet et le routage. |
| `BeforeCompletion(fn)` | Définit un hook appelé avant chaque requête de complétion. |
| `AfterCompletion(fn)` | Définit un hook appelé après chaque requête de complétion. |

### Valeurs par défaut

| Paramètre | Défaut |
|---|---|
| Port | `:8080` |
| ToolMode | `ToolModePassthrough` |
| `SimilarityLimit` | `0.6` |
| `MaxSimilarities` | `3` |
| `ContextSizeLimit` | `8000` |

### Support HTTPS

Le Gateway Server Agent supporte HTTPS pour une communication sécurisée. Lorsque des certificats TLS sont fournis, le serveur utilisera automatiquement HTTPS au lieu de HTTP.

```go
// Option 1 : Utiliser des fichiers de certificats (recommandé)
gateway, err := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithPort(443),
    gatewayserver.WithTLSCertFromFile("server.crt", "server.key"),
)

// Option 2 : Utiliser des données de certificat en mémoire
certData, _ := os.ReadFile("server.crt")
keyData, _ := os.ReadFile("server.key")

gateway, err := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithPort(443),
    gatewayserver.WithTLSCert(certData, keyData),
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
gatewayserver.WithAgentCrew(map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
}, "coder")
```

### Gestion dynamique de l'équipe

```go
// Ajouter un agent à la volée
err := gateway.AddChatAgentToCrew("cook", cookAgent)

// Supprimer un agent (impossible de supprimer l'agent actif)
err := gateway.RemoveChatAgentFromCrew("thinker")

// Obtenir tous les agents
agents := gateway.GetChatAgents()

// Remplacer toute l'équipe
gateway.SetChatAgents(newCrew)
```

### Changer d'agent manuellement

```go
// Obtenir l'agent actuellement sélectionné
id := gateway.GetSelectedAgentId()

// Basculer vers un autre agent
err := gateway.SetSelectedAgentId("coder")
```

**Note :** Un seul agent est actif à la fois. `GetName()`, `GetModelID()`, `GetMessages()`, etc. opèrent tous sur l'agent actuellement actif.

---

## 5. Pipeline HTTP de complétion (handleChatCompletions)

Le handler HTTP `handleChatCompletions` est le point d'entrée principal pour les requêtes de complétion. Il traite les requêtes `POST /v1/chat/completions`.

### Étapes du pipeline

1. **Hook BeforeCompletion** (si défini)
2. **Parsing de la requête** (décodage du corps JSON au format OpenAI)
3. **Résolution du modèle** (correspondance du champ `model` avec un agent de l'équipe ou utilisation de l'agent courant)
4. **Synchronisation des messages** (import de l'historique de conversation depuis la requête)
5. **Compression du contexte** (si compresseur configuré et contexte dépassant la limite)
6. **Injection du contexte RAG** (si agent RAG configuré)
7. **Routage intelligent** (si orchestrateur configuré, détection du sujet et changement d'agent)
8. **Gestion des outils** (dispatch selon le mode d'outils) :
   - **Passthrough** : Transmet les définitions d'outils au LLM, retourne les tool_calls au client
   - **Auto-execute** : Exécute les outils côté serveur, boucle jusqu'à la réponse finale
9. **Génération de la complétion** (streaming SSE ou JSON non-streaming)
10. **Nettoyage de l'état des outils**
11. **Hook AfterCompletion** (si défini)

### Format de la requête (compatible OpenAI)

```json
POST /v1/chat/completions
{
    "model": "assistant",
    "messages": [
        {"role": "system", "content": "Tu es un assistant utile."},
        {"role": "user", "content": "Bonjour !"}
    ],
    "stream": false,
    "temperature": 0.7,
    "tools": [...]
}
```

### Format de la réponse non-streaming

```json
{
    "id": "chatcmpl-abc123",
    "object": "chat.completion",
    "created": 1700000000,
    "model": "assistant",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "Bonjour ! Comment puis-je vous aider ?"
            },
            "finish_reason": "stop"
        }
    ],
    "usage": {
        "prompt_tokens": 10,
        "completion_tokens": 8,
        "total_tokens": 18
    }
}
```

### Format de la réponse streaming (SSE)

```
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"assistant","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"assistant","choices":[{"index":0,"delta":{"content":"Bonjour"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"assistant","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

---

## 6. Serveur HTTP et routes

### Démarrage du serveur

```go
err := gateway.StartServer()
```

### Routes disponibles

| Méthode | Chemin | Description |
|---|---|---|
| `POST` | `/v1/chat/completions` | Générer une complétion (streaming ou non-streaming) |
| `GET` | `/v1/models` | Lister les modèles disponibles (un par agent de l'équipe) |
| `GET` | `/health` | Vérification de santé |

### Endpoint des modèles

`GET /v1/models` retourne tous les agents de l'équipe comme modèles disponibles :

```json
{
    "object": "list",
    "data": [
        {"id": "coder", "object": "model", "created": 1700000000, "owned_by": "nova-gateway"},
        {"id": "thinker", "object": "model", "created": 1700000000, "owned_by": "nova-gateway"},
        {"id": "generic", "object": "model", "created": 1700000000, "owned_by": "nova-gateway"}
    ]
}
```

### CORS

Toutes les réponses incluent des headers CORS autorisant toutes les origines. Les requêtes de pré-vol `OPTIONS` sont gérées automatiquement.

---

## 7. Modes d'exécution des outils (Tool Modes)

Le Gateway Server Agent supporte deux modes distincts d'exécution des outils, contrôlés via `WithToolMode` :

### ToolModePassthrough (par défaut)

En mode passthrough, la gateway transmet les définitions d'outils au backend LLM et retourne les `tool_calls` au client. Le client est responsable de l'exécution des outils et de l'envoi des résultats dans les requêtes suivantes. C'est le mode utilisé par des outils comme `qwen-code` et `aider`.

```go
gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithPort(8080),
    // ToolModePassthrough est le mode par défaut
)
```

**Flux d'appel de fonctions côté client :**

1. Le client envoie une requête avec un tableau `tools`
2. La gateway transmet les outils au LLM
3. Le LLM retourne des `tool_calls` dans la réponse (ou le flux SSE)
4. Le client exécute les outils localement
5. Le client envoie une nouvelle requête incluant des messages avec le rôle `tool` contenant les résultats
6. La gateway complète avec la réponse finale

**Réponse non-streaming avec tool_calls :**

```json
{
    "id": "chatcmpl-abc123",
    "choices": [{
        "index": 0,
        "message": {
            "role": "assistant",
            "tool_calls": [{
                "id": "call_xyz",
                "type": "function",
                "function": {"name": "calculate_sum", "arguments": "{\"a\":3,\"b\":5}"}
            }]
        },
        "finish_reason": "tool_calls"
    }]
}
```

**Réponse streaming avec tool_calls :**

```
data: {"choices":[{"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"call_xyz","type":"function","function":{"name":"calculate_sum","arguments":""}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"a\":3,\"b\":5}"}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]
```

### ToolModeAutoExecute

En mode auto-execute, la gateway gère l'exécution des outils côté serveur en utilisant la fonction `ExecuteFn` configurée. Le client ne voit que la réponse finale et n'a pas connaissance des appels de fonctions.

```go
gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithToolsAgent(toolsAgent),
    gatewayserver.WithToolMode(gatewayserver.ToolModeAutoExecute),
    gatewayserver.WithExecuteFn(func(name string, args string) (string, error) {
        switch name {
        case "calculate_sum":
            // Exécuter la fonction
            return `{"result": 8}`, nil
        default:
            return `{"error": "fonction inconnue"}`, fmt.Errorf("inconnue : %s", name)
        }
    }),
)
```

**Flux d'exécution des outils côté serveur :**

1. Le client envoie une requête (pas besoin de tableau `tools`)
2. La gateway détecte les appels de fonctions via le `tools.Agent`
3. La gateway exécute chaque outil via `ExecuteFn`
4. La gateway renvoie les résultats au LLM
5. Les étapes 2-4 se répètent jusqu'à ce que le LLM produise une réponse finale
6. Le client ne reçoit que la réponse finale

---

## 7.3. Architecture Passthrough-First (Avancée)

Ce pattern avancé garantit que toutes les requêtes d'outils sont toujours traitées par un agent capable de gérer les outils en premier, empêchant les erreurs avec les modèles qui ne supportent pas les outils.

### Pourquoi Passthrough-First ?

Lorsqu'on utilise des équipes multi-agents, certains agents peuvent utiliser des modèles qui ne supportent pas les appels de fonctions. Si un client (comme `pi`, `qwen-code`, ou `aider`) envoie une requête avec des outils à un tel agent, cela échouera. L'architecture passthrough-first résout ce problème en routant toutes les requêtes d'outils via un agent "passthrough" désigné qui utilise un modèle capable de gérer les outils.

### Fonctionnement

```
Requête Client + tools[]
    ↓
🔀 AGENT PASSTHROUGH (toujours en premier)
    ├─ Phase 1 : Détection rapide (non-streaming)
    │  ├─ Détecte des tool_calls nécessaires ?
    │  │  ├─ OUI → Phase 2 : Stream la réponse au client
    │  │  └─ NON → Redirige vers l'agent approprié
    ↓
💬 Agent Sélectionné (coder/generic/etc.)
    └─ Répond sans outils
```

### Configuration

1. **Créer un agent passthrough** avec un modèle capable de gérer les outils :

```go
passthroughAgent, err := chat.NewAgent(ctx,
    agents.Config{
        Name:                    "passthrough",  // ← L'ID doit être "passthrough"
        EngineURL:               engineURL,
        SystemInstructions:      "Vous répondez de manière appropriée aux requêtes d'outils.",
        KeepConversationHistory: true,
    },
    models.Config{
        Name:        "hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m",  // ← Modèle supportant les outils
        Temperature: models.Float64(0.0),
    },
)
```

2. **Ajouter à l'équipe** avec l'ID `"passthrough"` :

```go
agentCrew := map[string]*chat.Agent{
    "coder":       coderAgent,
    "generic":     genericAgent,
    "passthrough": passthroughAgent,  // ← Doit avoir exactement cet ID
}
```

3. **Créer la gateway** (aucune configuration supplémentaire nécessaire) :

```go
gateway, err := gatewayserver.NewAgent(ctx,
    gatewayserver.WithAgentCrew(agentCrew, "generic"),
    // ToolModePassthrough est le mode par défaut
)
```

### Validation

La gateway valide au démarrage qu'un agent `"passthrough"` existe lorsqu'elle est en mode `ToolModePassthrough` :

```
Erreur : le mode passthrough nécessite un agent avec l'ID 'passthrough' dans l'équipe.
Veuillez ajouter un agent avec l'ID 'passthrough' qui supporte les appels de fonctions.
```

### Détection en Deux Phases

**Phase 1 : Détection (Toujours Non-Streaming)**
- Appel API rapide pour déterminer si des outils sont nécessaires
- Analyse le `finish_reason` et les `tool_calls` dans la réponse
- Coût : 1 appel API

**Phase 2 : Réponse (Conditionnelle)**
- Si `tool_calls` détectés + streaming demandé → Fait un appel streaming
- Si `tool_calls` détectés + non-streaming → Utilise la réponse de détection
- Si pas de `tool_calls` → Redirige vers l'agent approprié
- Coût : 0-1 appel API supplémentaire

### Exemple de Flux

**Requête d'outil détectée :**
```
Client : "Quelle est la météo à Paris ?" + tools[]
    ↓
🔀 PASSTHROUGH : Phase 1 détecte des tool_calls nécessaires
    ↓
🔀 PASSTHROUGH : Phase 2 stream les tool_calls au client
    ↓
Client exécute get_weather() localement
    ↓
Réponse finale
```

**Pas de requête d'outil :**
```
Client : "Combien font 2+2 ?" + tools[]
    ↓
🔀 PASSTHROUGH : Phase 1 détecte AUCUN tool_calls nécessaire
    ↓
💬 GENERIC : Répond directement "4"
```

### Considération de Coût

- **Meilleur cas** (pas d'outils nécessaires) : 1 appel API (détection seulement)
- **Pire cas** (outils + streaming) : 2 appels API (détection + réponse streaming)
- **Cas moyen** (outils + non-streaming) : 1 appel API (réponse de détection réutilisée)

### Notes Importantes

1. **L'ID de l'agent est critique** : L'agent doit avoir exactement l'ID `"passthrough"` pour que la validation fonctionne.
2. **Modèle capable d'outils requis** : Utilisez des modèles comme Qwen2.5-Coder, GPT-4, Claude, ou similaires qui supportent les appels de fonctions.
3. **Activation automatique** : Aucune configuration spéciale nécessaire au-delà de l'ajout de l'agent à l'équipe.
4. **Support du streaming** : Supporte pleinement les clients streaming et non-streaming.

### Exemple d'Implémentation

Voir [samples/89-gateway-compose-cagent](../../samples/89-gateway-compose-cagent) pour un exemple complet fonctionnel avec :
- Routage d'agents basé sur la configuration
- Gestion passthrough-first des outils
- Hooks BeforeCompletion pour le traçage
- Script de test pour valider le comportement

---

## 8. Routage intelligent (Orchestrateur)

Lorsqu'un agent orchestrateur est attaché, la gateway peut automatiquement router les questions vers l'agent spécialisé le plus approprié.

### Configuration

```go
orchestratorAgent, _ := orchestrator.NewAgent(ctx,
    agents.Config{
        Name:               "orchestrator",
        EngineURL:          engineURL,
        SystemInstructions: `Identifiez le sujet principal en un seul mot.
            Sujets possibles : Technology, Philosophy, Cooking, etc.
            Répondez en JSON avec le champ 'topic_discussion'.`,
    },
    models.Config{Name: "my-model", Temperature: models.Float64(0.0)},
)

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithAgentCrew(agentCrew, "generic"),
    gatewayserver.WithOrchestratorAgent(orchestratorAgent),
    gatewayserver.WithMatchAgentIdToTopicFn(func(currentAgentId, topic string) string {
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

1. L'orchestrateur analyse la question de l'utilisateur et détecte le sujet.
2. La fonction `matchAgentIdToTopicFn` fait correspondre le sujet à un ID d'agent.
3. La gateway bascule vers l'agent correspondant s'il est différent de l'agent actuel.
4. La complétion est générée par l'agent nouvellement sélectionné.

### Détection directe du sujet

```go
agentId, err := gateway.DetectTopicThenGetAgentId("Écris une fonction Python")
// agentId = "coder"
```

---

## 9. Intégration RAG

```go
ragAgent, _ := rag.NewAgent(ctx, ragConfig, ragModelConfig)

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.4, 5),
)
```

Pendant le pipeline de complétion, la gateway effectue une recherche de similarité et injecte le contexte pertinent dans la conversation avant de générer la complétion.

---

## 10. Compression du contexte

```go
compressorAgent, _ := compressor.NewAgent(ctx, compressorConfig, compressorModelConfig,
    compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
)

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithCompressorAgentAndContextSize(compressorAgent, 8000),
)
```

Au début de chaque requête de complétion, le contexte est compressé s'il dépasse la limite configurée.

### Compression manuelle

```go
// Compresser uniquement si au-dessus de la limite
newSize, err := gateway.CompressChatAgentContextIfOverLimit()
```

---

## 11. Hooks de cycle de vie (BeforeCompletion / AfterCompletion)

Les hooks de cycle de vie permettent d'exécuter de la logique personnalisée avant et après chaque requête de complétion HTTP (`POST /v1/chat/completions`). Ils sont configurés comme options fonctionnelles `GatewayServerAgentOption`.

### BeforeCompletion

Appelé au tout début de chaque handler HTTP `handleChatCompletions`, avant le parsing de la requête. Le hook reçoit une référence vers le gateway server agent.

```go
gatewayserver.BeforeCompletion(func(a *gatewayserver.GatewayServerAgent) {
    fmt.Printf("[AVANT] Agent : %s\n", a.GetName())
})
```

### AfterCompletion

Appelé à la toute fin de chaque handler HTTP `handleChatCompletions`, après le nettoyage. Le hook reçoit une référence vers le gateway server agent.

```go
gatewayserver.AfterCompletion(func(a *gatewayserver.GatewayServerAgent) {
    fmt.Printf("[APRÈS] Agent : %s\n", a.GetName())
})
```

### Exemple complet

```go
callCount := 0

gateway, _ := gatewayserver.NewAgent(ctx,
    gatewayserver.WithSingleAgent(chatAgent),
    gatewayserver.WithPort(8080),
    gatewayserver.BeforeCompletion(func(a *gatewayserver.GatewayServerAgent) {
        callCount++
        fmt.Printf("[AVANT] Appel #%d - Agent : %s\n", callCount, a.GetName())
    }),
    gatewayserver.AfterCompletion(func(a *gatewayserver.GatewayServerAgent) {
        fmt.Printf("[APRÈS] Appel #%d - Agent : %s\n", callCount, a.GetName())
    }),
)

log.Fatal(gateway.StartServer())
```

### Les hooks sont optionnels

Si aucun hook n'est fourni, l'agent se comporte exactement comme avant. Le code existant sans hooks continue de fonctionner sans aucune modification.

---

## 12. Gestion de la conversation

Toutes les méthodes de conversation opèrent sur l'agent de chat **actuellement actif** :

```go
// Obtenir les messages
msgs := gateway.GetMessages()

// Obtenir la taille du contexte
size := gateway.GetContextSize()

// Réinitialiser la conversation
gateway.ResetMessages()

// Ajouter un message
gateway.AddMessage(roles.User, "Bonjour")

// Arrêter le streaming
gateway.StopStream()
```

---

## 13. Types compatibles OpenAI

Le package `gatewayserver` exporte tous les types compatibles OpenAI pour utilisation dans les tests et les intégrations personnalisées :

### Types de requête

| Type | Description |
|---|---|
| `ChatCompletionRequest` | Le corps principal de la requête `POST /v1/chat/completions` |
| `ChatCompletionMessage` | Un message dans la conversation (role, content, tool_calls, tool_call_id) |
| `MessageContent` | Contenu d'un message, supporte les formats string et array (multi-modal) |
| `ToolDefinition` | Un outil disponible pour le modèle |
| `FunctionDefinition` | Décrit une fonction appelable (name, description, parameters) |
| `ToolCall` | Un appel de fonction effectué par l'assistant |
| `FunctionCall` | Contient le nom de la fonction et les arguments JSON |

Le type `MessageContent` gère automatiquement les différents formats de contenu de l'API OpenAI :
- String simple : `"Bonjour"`
- Tableau de strings : `["Bonjour", "monde"]`
- Tableau de parties multi-modales : `[{"type": "text", "text": "Bonjour"}]`

Utiliser `NewMessageContent("texte")` pour créer un nouveau contenu de message.

### Types de réponse (non-streaming)

| Type | Description |
|---|---|
| `ChatCompletionResponse` | Réponse complète avec id, object, model, choices, usage |
| `ChatCompletionChoice` | Un choix avec message et finish_reason |
| `Usage` | Statistiques d'utilisation des tokens |

### Types de réponse (streaming)

| Type | Description |
|---|---|
| `ChatCompletionChunk` | Un chunk SSE en mode streaming |
| `ChatCompletionChunkChoice` | Un choix avec delta et finish_reason |
| `ChatCompletionDelta` | Contenu incrémental (role, content, tool_calls) |

### Autres types

| Type | Description |
|---|---|
| `ModelsResponse` | Réponse pour `GET /v1/models` |
| `ModelEntry` | Une entrée de modèle |
| `APIError` | Réponse d'erreur compatible OpenAI |
| `APIErrorDetail` | Détails de l'erreur (message, type, code) |

---

## 14. Tests

### Tests unitaires

Le package inclut des tests unitaires complets dans `gateway_test.go` avec un faux backend LLM. Lancez-les avec :

```bash
go test ./nova-sdk/agents/gatewayserver/ -v
```

La suite de tests couvre :
- Sérialisation des requêtes/réponses
- Aller-retour HTTP complet (non-streaming et streaming)
- Parsing SSE et terminaison `data: [DONE]`
- Endpoint des modèles
- Endpoint de santé
- Types d'appels de fonctions

### Helpers publics pour les tests

Pour les tests d'intégration, le package expose des wrappers publics autour des handlers HTTP privés :

```go
// Créez un gateway agent et utilisez ces méthodes dans vos tests :
gateway.HandleChatCompletionsForTest(w, r)
gateway.HandleListModelsForTest(w, r)
gateway.HandleHealthForTest(w, r)
```

### Tests manuels avec curl

Voir `samples/84-gateway-server-agent/test.sh` (agent unique) et `samples/85-gateway-server-agent-crew/test.sh` (équipe) pour des scripts de tests curl complets.

---

## 15. Référence API

### Constructeur

```go
func NewAgent(ctx context.Context, options ...GatewayServerAgentOption) (*GatewayServerAgent, error)
```

### Types

```go
type GatewayServerAgentOption func(*GatewayServerAgent) error

type ToolMode int
const (
    ToolModePassthrough ToolMode = iota
    ToolModeAutoExecute
)
```

### Fonctions d'option

| Fonction | Description |
|---|---|
| `WithAgentCrew(crew, selectedId)` | Définit l'équipe et l'agent initial. |
| `WithSingleAgent(chatAgent)` | Crée une équipe à agent unique. |
| `WithPort(port)` | Définit le port du serveur HTTP (défaut : 8080). |
| `WithToolsAgent(toolsAgent)` | Attache un agent d'outils. |
| `WithToolMode(mode)` | Définit le mode d'exécution des outils. |
| `WithExecuteFn(fn)` | Définit la fonction d'exécution des outils. |
| `WithConfirmationPromptFn(fn)` | Définit la fonction de confirmation personnalisée des outils. |
| `WithMatchAgentIdToTopicFn(fn)` | Définit la fonction de correspondance sujet-agent. |
| `WithRagAgent(ragAgent)` | Attache un agent RAG. |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | Attache un RAG avec configuration. |
| `WithCompressorAgent(compressorAgent)` | Attache un agent de compression. |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Attache un compresseur avec limite. |
| `WithOrchestratorAgent(orchestratorAgent)` | Attache un agent orchestrateur. |
| `BeforeCompletion(fn func(*GatewayServerAgent))` | Définit le hook avant chaque complétion. |
| `AfterCompletion(fn func(*GatewayServerAgent))` | Définit le hook après chaque complétion. |

### Méthodes

| Méthode | Description |
|---|---|
| `StartServer() error` | Démarre le serveur HTTP avec toutes les routes. |
| `GetPort() string` | Obtient le port HTTP. |
| `GetToolMode() ToolMode` | Obtient le mode d'exécution des outils. |
| `SetToolMode(mode)` | Définit le mode d'exécution des outils. |
| `StopStream()` | Arrête l'opération de streaming en cours. |
| `GetMessages() []messages.Message` | Obtient les messages de l'agent actif. |
| `GetContextSize() int` | Obtient la taille du contexte de l'agent actif. |
| `ResetMessages()` | Réinitialise la conversation de l'agent actif. |
| `AddMessage(role, content)` | Ajoute un message à l'agent actif. |
| `GetChatAgents() map[string]*chat.Agent` | Obtient tous les agents de l'équipe. |
| `SetChatAgents(crew)` | Remplace toute l'équipe. |
| `AddChatAgentToCrew(id, agent) error` | Ajoute un agent à l'équipe. |
| `RemoveChatAgentFromCrew(id) error` | Supprime un agent de l'équipe. |
| `GetSelectedAgentId() string` | Obtient l'ID de l'agent actif. |
| `SetSelectedAgentId(id) error` | Change d'agent actif. |
| `DetectTopicThenGetAgentId(query) (string, error)` | Détecte le sujet et retourne l'ID de l'agent correspondant. |
| `SetOrchestratorAgent(orchestratorAgent)` | Définit l'agent orchestrateur. |
| `GetOrchestratorAgent() OrchestratorAgent` | Obtient l'agent orchestrateur. |
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
| `Kind() agents.Kind` | Retourne `agents.ChatServer`. |
| `GetName() string` | Retourne le nom de l'agent actif. |
| `GetModelID() string` | Retourne l'ID du modèle de l'agent actif. |
| `HandleChatCompletionsForTest(w, r)` | Helper de test : expose le handler de complétion. |
| `HandleListModelsForTest(w, r)` | Helper de test : expose le handler des modèles. |
| `HandleHealthForTest(w, r)` | Helper de test : expose le handler de santé. |
