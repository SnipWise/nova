# Server Agent

## Description

Le **Server Agent** est un agent de chat qui expose une API HTTP/REST avec streaming SSE (Server-Sent Events). Il encapsule un `chat.Agent` et peut être enrichi avec des agents auxiliaires (Tools, RAG, Compressor) pour des fonctionnalités avancées.

## Fonctionnalités

- **API HTTP/REST** : Expose des endpoints pour interagir avec l'agent via HTTP
- **Streaming SSE** : Réponses en temps réel via Server-Sent Events
- **Tools Agent** : Exécution de fonctions (function calling) avec confirmation utilisateur
- **RAG Agent** : Recherche de similarité et enrichissement du contexte
- **Compressor Agent** : Compression automatique du contexte quand la limite est atteinte
- **Human-in-the-loop** : Validation des appels de fonctions via l'interface web

## Création d'un Server Agent

### Syntaxe avec options

```go
import (
    "context"
    "github.com/snipwise/nova/nova-sdk/agents"
    "github.com/snipwise/nova/nova-sdk/agents/server"
    "github.com/snipwise/nova/nova-sdk/models"
)

// Création d'un server agent simple
agent, err := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithPort(8080),
)

// Création d'un server agent complet
agent, err := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithPort(8080),
    server.WithToolsAgent(toolsAgent),
    server.WithRagAgent(ragAgent),
    server.WithCompressorAgentAndContextSize(compressorAgent, 8000),
    server.WithExecuteFn(myCustomExecutor),
)
```

### Options disponibles

| Option | Description |
|--------|-------------|
| `WithPort(port int)` | Définit le port HTTP (défaut: 8080) |
| `WithExecuteFn(fn)` | Fonction personnalisée d'exécution des tools |
| `WithToolsAgent(toolsAgent)` | Ajoute un agent pour l'exécution de fonctions |
| `WithRagAgent(ragAgent)` | Ajoute un agent RAG pour la recherche de documents |
| `WithRagAgentAndSimilarityConfig(ragAgent, limit, max)` | RAG avec configuration de similarité |
| `WithCompressorAgent(compressorAgent)` | Ajoute un agent pour la compression du contexte |
| `WithCompressorAgentAndContextSize(compressorAgent, limit)` | Compressor avec limite de contexte |

## API HTTP Routes

### Routes principales

#### `POST /completion`
Génère une complétion avec streaming SSE.

**Request Body:**
```json
{
  "data": {
    "message": "Votre question ici"
  }
}
```

**Response:** Server-Sent Events (SSE)
```
data: {"message": "Chunk de réponse..."}
data: {"message": "", "finish_reason": "stop"}
```

**Processus de traitement:**
1. Compression du contexte si nécessaire (CompressorAgent)
2. Détection et exécution des appels de fonctions (ToolsAgent)
3. Recherche de contexte pertinent (RagAgent)
4. Génération de la réponse avec streaming

#### `POST /completion/stop`
Arrête le streaming en cours.

**Response:**
```json
{
  "status": "ok",
  "message": "Stream stopped"
}
```

### Routes de gestion de la mémoire

#### `POST /memory/reset`
Réinitialise l'historique de conversation.

**Response:**
```json
{
  "status": "ok",
  "message": "Memory reset successfully"
}
```

#### `GET /memory/messages/list`
Récupère tous les messages de la conversation.

**Response:**
```json
{
  "messages": [
    {
      "role": "user",
      "content": "Message..."
    }
  ]
}
```

#### `GET /memory/messages/context-size`
Obtient la taille du contexte actuel.

**Response:**
```json
{
  "messages_count": 10,
  "characters_count": 1500,
  "limit": 8000
}
```

### Routes de gestion des opérations (Tools)

Ces routes sont utilisées pour la validation des appels de fonctions (human-in-the-loop).

#### `POST /operation/validate`
Valide une opération de tool call en attente.

**Request Body:**
```json
{
  "operation_id": "op_12345"
}
```

**Response:** SSE
```
data: {"message": "✅ Operation op_12345 validated<br>"}
```

#### `POST /operation/cancel`
Annule une opération de tool call en attente.

**Request Body:**
```json
{
  "operation_id": "op_12345"
}
```

**Response:** SSE
```
data: {"message": "⛔️ Operation op_12345 cancelled<br>"}
```

#### `POST /operation/reset`
Annule toutes les opérations en attente.

**Response:** SSE
```
data: {"message": "🔄 All pending operations cancelled (3 operations)"}
```

### Routes d'information

#### `GET /models`
Retourne les informations sur les modèles utilisés.

**Response:**
```json
{
  "status": "ok",
  "chat_model": "qwen2.5:1.5b",
  "embeddings_model": "mxbai-embed-large",
  "tools_model": "jan-nano"
}
```

#### `GET /health`
Vérifie l'état de santé du serveur.

**Response:**
```json
{
  "status": "ok"
}
```

## Démarrage du serveur

```go
// Démarrer le serveur (bloquant)
if err := agent.StartServer(); err != nil {
    log.Fatal(err)
}
```

Le serveur démarre sur `http://localhost:8080` (ou le port configuré).

## Modes d'utilisation

### 1. Mode HTTP/API
Pour une utilisation via API REST avec interface web.
- Les tool calls nécessitent une validation via `/operation/validate`
- Streaming SSE pour les réponses en temps réel

### 2. Mode CLI
Pour une utilisation en ligne de commande directe.
```go
result, err := agent.StreamCompletion(question, callback)
```
- Les tool calls sont auto-confirmés
- Streaming via callback

## Notifications de Tool Calls

Quand un tool call est détecté, une notification SSE est envoyée:

```json
{
  "kind": "tool_call",
  "status": "pending",
  "operation_id": "op_12345",
  "message": "Tool call detected: calculate"
}
```

L'utilisateur peut alors valider ou annuler l'opération via les routes `/operation/*`.

## Exemple complet

```go
ctx := context.Background()

// Configuration
agentConfig := agents.Config{
    Name: "Assistant",
    Instructions: "Tu es un assistant utile.",
}
modelConfig := models.Config{
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    ModelID: "qwen2.5:1.5b",
}

// Créer l'agent
agent, err := server.NewAgent(
    ctx,
    agentConfig,
    modelConfig,
    server.WithPort(8080),
    server.WithToolsAgent(toolsAgent),
)
if err != nil {
    log.Fatal(err)
}

// Démarrer le serveur
log.Println("🚀 Starting server on :8080")
if err := agent.StartServer(); err != nil {
    log.Fatal(err)
}
```

## Pipeline de traitement (POST /completion)

```
1. Compression du contexte (si CompressorAgent configuré)
   ↓
2. Détection de tool calls (si ToolsAgent configuré)
   ↓
3. Notification SSE des tool calls détectés
   ↓
4. Validation utilisateur via /operation/validate ou /cancel
   ↓
5. Exécution des fonctions (si validées)
   ↓
6. Ajout du résultat au contexte
   ↓
7. Recherche RAG (si RagAgent configuré)
   ↓
8. Génération de la réponse avec streaming SSE
   ↓
9. Nettoyage de l'état
```

## Notes

- Port par défaut: **8080**
- Format de streaming: **Server-Sent Events (SSE)**
- CORS: **Activé** (`Access-Control-Allow-Origin: *`)
- Le serveur utilise `http.ServeMux` standard de Go
- Les opérations de tool calls sont gérées avec des channels Go pour la concurrence
