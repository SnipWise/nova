# Crew Server Agent

## Description

Le **Crew Server Agent** est un agent HTTP qui permet d'exposer un ou plusieurs agents de chat via une API REST avec streaming SSE (Server-Sent Events). Il peut gérer plusieurs agents simultanément et router les requêtes vers l'agent approprié en fonction du contexte.

## Fonctionnalités principales

- **Multi-agents** : Gestion d'une collection d'agents de chat
- **Routage intelligent** : Sélection automatique de l'agent approprié via un orchestrateur
- **API HTTP/REST** : Exposition d'une API avec streaming SSE
- **Appels d'outils** : Support des fonction calling avec notifications en temps réel
- **RAG** : Recherche de contexte pertinent via un agent RAG
- **Compression** : Compression automatique du contexte quand la limite est atteinte
- **Confirmation humaine** : Validation des appels d'outils critiques

## Configuration

### Création basique avec un seul agent

```go
crewAgent, err := crewserver.NewAgent(
    ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(3500),
)
```

### Création avec plusieurs agents

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "cook":    cookAgent,
}

crewAgent, err := crewserver.NewAgent(
    ctx,
    crewserver.WithAgentCrew(agentCrew, "coder"), // "coder" est l'agent par défaut
    crewserver.WithPort(3500),
)
```

### Options disponibles

- `WithSingleAgent(chatAgent)` - Crée un crew avec un seul agent
- `WithAgentCrew(agentCrew, selectedAgentId)` - Définit une collection d'agents et l'agent par défaut
- `WithPort(port)` - Définit le port HTTP (défaut: 3500)
- `WithToolsAgent(toolsAgent)` - Ajoute un agent d'outils pour le function calling
- `WithTasksAgent(tasksAgent)` - Ajoute un agent de tâches pour la planification et l'orchestration
- `WithRagAgent(ragAgent)` - Ajoute un agent RAG pour la recherche de contexte
- `WithRagAgentAndSimilarityConfig(ragAgent, similarityLimit, maxSimilarities)` - RAG avec configuration de similarité
- `WithCompressorAgent(compressorAgent)` - Ajoute un agent de compression
- `WithCompressorAgentAndContextSize(compressorAgent, contextSizeLimit)` - Compresseur avec limite de taille
- `WithOrchestratorAgent(orchestratorAgent)` - Ajoute un orchestrateur pour le routage automatique
- `WithMatchAgentIdToTopicFn(fn)` - Définit la fonction de mapping topic -> agent ID
- `WithExecuteFn(fn)` - Définit la fonction d'exécution des outils

## API REST

### Routes disponibles

#### POST /completion

Génère une complétion avec streaming SSE.

**Request:**
```json
{
  "data": {
    "message": "Votre question ici"
  }
}
```

**Response:** Stream SSE avec événements:
```
data: {"message": "chunk de réponse"}
data: {"message": "autre chunk"}
data: {"message": "", "finish_reason": "stop"}
```

**Notifications d'outils:**
```
data: {"kind": "tool_call", "status": "pending", "operation_id": "123", "message": "Appel de la fonction X"}
```

#### POST /completion/stop

Arrête le streaming en cours.

**Response:**
```json
{
  "status": "ok",
  "message": "Stream stopped"
}
```

#### POST /memory/reset

Réinitialise l'historique de conversation (garde l'instruction système).

**Response:**
```json
{
  "status": "ok",
  "message": "Memory reset successfully"
}
```

#### GET /memory/messages/list

Liste tous les messages de la conversation.

**Response:**
```json
{
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ]
}
```

#### GET /memory/messages/context-size

Retourne la taille approximative du contexte.

**Response:**
```json
{
  "context_size": 1234
}
```

#### POST /operation/validate

Valide une opération en attente (confirmation humaine).

**Request:**
```json
{
  "operation_id": "123"
}
```

**Response:**
```json
{
  "status": "ok",
  "message": "Operation validated"
}
```

#### POST /operation/cancel

Annule une opération en attente.

**Request:**
```json
{
  "operation_id": "123"
}
```

**Response:**
```json
{
  "status": "ok",
  "message": "Operation cancelled"
}
```

#### POST /operation/reset

Réinitialise toutes les opérations en attente.

**Response:**
```json
{
  "status": "ok",
  "message": "Operations reset successfully"
}
```

#### GET /models

Retourne les informations sur les modèles utilisés.

**Response:**
```json
{
  "chat_model": "ai/qwen2.5:1.5B-F16",
  "tools_model": "hf.co/menlo/jan-nano-gguf:q4_k_m",
  "rag_model": "ai/mxbai-embed-large",
  "compressor_model": "ai/qwen2.5:1.5B-F16"
}
```

#### GET /health

Vérification de l'état du serveur.

**Response:**
```json
{
  "status": "ok",
  "message": "Server is healthy"
}
```

## Flux de traitement d'une requête

1. **Compression du contexte** (si configuré et limite atteinte)
2. **Détection et exécution des appels d'outils** (si agent d'outils configuré)
3. **Ajout du contexte RAG** (si agent RAG configuré)
4. **Routage vers l'agent approprié** (si orchestrateur configuré)
5. **Génération de la complétion** avec streaming SSE
6. **Nettoyage** de l'état des outils

## Routage intelligent

Avec un orchestrateur, le crew peut détecter automatiquement le sujet et router vers l'agent spécialisé:

```go
// Fonction de mapping topic -> agent ID
matchAgentFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "coding", "programming", "development":
        return "coder"
    case "philosophy", "thinking", "ideas":
        return "thinker"
    case "cooking", "recipe", "food":
        return "cook"
    default:
        return "generic"
    }
}

crewAgent, err := crewserver.NewAgent(
    ctx,
    crewserver.WithAgentCrew(agentCrew, "generic"),
    crewserver.WithOrchestratorAgent(orchestratorAgent),
    crewserver.WithMatchAgentIdToTopicFn(matchAgentFn),
)
```

## Commandes spéciales

Le crew server supporte des commandes internes:

- `[agent-list]` - Liste tous les agents disponibles
- `[select-agent <id>]` - Sélectionne manuellement un agent

## Démarrage du serveur

```go
if err := crewAgent.StartServer(); err != nil {
    log.Fatal(err)
}
```

Le serveur démarre sur `http://localhost:3500` (par défaut).
