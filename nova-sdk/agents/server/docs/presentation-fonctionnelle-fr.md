# Server Agent - Présentation Fonctionnelle

## Vue d'ensemble

Le **Server Agent** est un agent HTTP/REST API du SDK N.O.V.A. qui encapsule un agent conversationnel (Chat Agent) et expose ses fonctionnalités via une interface web moderne avec support du streaming en temps réel.

Il s'agit d'un composant d'orchestration sophistiqué capable de coordonner plusieurs agents spécialisés pour fournir une API conversationnelle complète avec appel d'outils, récupération de documents (RAG) et gestion intelligente du contexte.

## Caractéristiques principales

### 1. API REST avec Streaming SSE

- **Endpoints HTTP RESTful** pour les complétions de chat
- **Streaming en temps réel** via Server-Sent Events (SSE)
- **Réponses progressives** pour une expérience utilisateur fluide
- **Gestion des erreurs** en temps réel via SSE

### 2. Appel d'outils avec workflow de confirmation

- **Détection automatique** des appels d'outils dans les requêtes utilisateur
- **Appels parallèles** de multiples outils
- **Workflow de validation web** : les outils requièrent une approbation avant exécution
- **Système de notifications** via SSE pour informer le client des opérations en attente
- **Gestion des opérations** : validation, annulation, réinitialisation

### 3. RAG (Retrieval-Augmented Generation)

- **Recherche de similarité** dans une base de documents vectoriels
- **Injection automatique** de contexte pertinent avant génération
- **Configuration flexible** : seuil de similarité, nombre maximal de documents
- **Amélioration des réponses** grâce à des informations externes

### 4. Compression intelligente du contexte

- **Gestion automatique** de la limite de tokens
- **Compression à la demande** ou automatique au-delà d'un seuil
- **Sauvegardes** : avertissement à 80%, réinitialisation à 90%
- **Préservation du contexte** essentiel tout en respectant les limites du modèle

### 5. Gestion de la mémoire conversationnelle

- **Historique des messages** persistant durant la session
- **Comptage de tokens** pour surveiller l'utilisation
- **Réinitialisation** de l'historique
- **Export JSON** des conversations
- **Listage** de tous les messages

## Architecture

### Structure principale

```go
type ServerAgent struct {
    chatAgent       *chat.Agent      // Agent conversationnel principal
    toolsAgent      *tools.Agent     // Optionnel : détection/exécution d'outils
    ragAgent        *rag.Agent       // Optionnel : récupération de documents
    compressorAgent *compressor.Agent // Optionnel : compression du contexte

    // Configuration RAG
    similarityLimit float64  // Par défaut : 0.6
    maxSimilarities int      // Par défaut : 3

    // Configuration compression
    contextSizeLimit int     // Par défaut : 8000

    // Configuration serveur
    port string
    ctx  context.Context
    log  logger.Logger

    // Gestion des opérations
    pendingOperations       map[string]*PendingOperation
    operationsMutex         sync.RWMutex
    stopStreamChan          chan bool
    currentNotificationChan chan ToolCallNotification
    notificationChanMutex   sync.Mutex

    // Exécuteur d'outils personnalisé
    executeFn func(string, string) (string, error)
}
```

### Agents spécialisés

Le Server Agent orchestre quatre types d'agents :

1. **Chat Agent** (obligatoire) : génération de texte conversationnel
2. **Tools Agent** (optionnel) : détection et exécution de fonctions
3. **RAG Agent** (optionnel) : recherche de documents pertinents
4. **Compressor Agent** (optionnel) : compression du contexte

## Endpoints HTTP

### Complétion

#### `POST /completion`
Stream une complétion de chat avec SSE.

**Requête :**
```json
{
  "data": {
    "message": "Votre question ici"
  }
}
```

**Réponse :** Flux SSE avec chunks de texte
```
data: {"message": "chunk de texte"}
data: {"message": "suite du texte"}
data: {"message": "", "finish_reason": "stop"}
```

**Notifications d'outils (si tools agent configuré) :**
```json
data: {
  "kind": "tool_call",
  "status": "pending",
  "operation_id": "op_0x140003dcbe0",
  "message": "Appel d'outil détecté: nom_fonction"
}
```

#### `POST /completion/stop`
Arrête le streaming en cours.

### Gestion de la mémoire

#### `POST /memory/reset`
Efface l'historique de la conversation.

**Réponse :**
```json
{
  "status": "ok",
  "message": "Memory reset"
}
```

#### `GET /memory/messages/list`
Récupère tous les messages de l'historique.

**Réponse :**
```json
{
  "messages": [
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ]
}
```

#### `GET /memory/messages/tokens`
Obtient le nombre de tokens et statistiques.

**Réponse :**
```json
{
  "context_size": 1234,
  "message": "Current context size: 1234 tokens"
}
```

### Gestion des opérations d'outils

#### `POST /operation/validate`
Approuve un appel d'outil en attente.

**Requête :**
```json
{
  "operation_id": "op_0x140003dcbe0"
}
```

**Réponse :**
```json
{
  "status": "validated",
  "operation_id": "op_0x140003dcbe0"
}
```

#### `POST /operation/cancel`
Rejette un appel d'outil en attente.

**Requête :**
```json
{
  "operation_id": "op_0x140003dcbe0"
}
```

**Réponse :**
```json
{
  "status": "cancelled",
  "operation_id": "op_0x140003dcbe0"
}
```

#### `POST /operation/reset`
Annule toutes les opérations en attente.

**Réponse :**
```json
{
  "status": "ok",
  "message": "All pending operations cancelled"
}
```

### Informations

#### `GET /models`
Informations sur les modèles utilisés.

**Réponse :**
```json
{
  "chat_model": "hf.co/menlo/jan-nano-gguf:q4_k_m",
  "tools_model": "hf.co/menlo/jan-nano-gguf:q4_k_m",
  "embeddings_model": "all-minilm:l6-v2"
}
```

#### `GET /health`
Vérification de santé du serveur.

**Réponse :**
```json
{
  "status": "ok",
  "message": "Server is running"
}
```

## Flux d'exécution d'une requête

### Traitement complet d'une requête `/completion`

1. **Compression du contexte** (si compressor agent configuré)
   - Vérification de la taille du contexte
   - Compression si dépassement du seuil

2. **Analyse de la requête**
   - Décodage JSON du message

3. **Configuration du streaming SSE**
   - Headers `text/event-stream`
   - Création du canal de notifications

4. **Détection d'outils** (si tools agent configuré)
   - Appel de `DetectParallelToolCallsWithConfirmation()`
   - Envoi de notifications SSE pour opérations en attente
   - Attente de validation/annulation
   - Exécution des outils approuvés
   - Ajout des résultats au contexte

5. **Recherche de similarité RAG** (si RAG agent configuré)
   - Recherche de documents pertinents
   - Injection en tant que message système

6. **Génération de la complétion en streaming**
   - Streaming des chunks via SSE
   - Gestion des signaux d'arrêt
   - Envoi de la raison de fin

7. **Gestion des erreurs**
   - Streaming des erreurs via SSE

## Exemples d'utilisation

### 1. Server Agent basique (Chat uniquement)

```go
import (
    "context"
    "log"

    "nova-sdk/agents"
    "nova-sdk/agents/server"
    "nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Le paramètre executeFunction est maintenant OPTIONNEL
    // Vous pouvez :
    // 1. L'omettre complètement (utilise executeFunction par défaut)
    // 2. Passer nil (utilise executeFunction par défaut)
    // 3. Passer une fonction personnalisée
    agent, err := server.NewAgent(
        ctx,
        agents.Config{
            Name:               "bob-server-agent",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "Vous êtes Bob, un assistant IA serviable.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.4),
        },
        ":3500",
        // executeFunction est optionnel - omis ici, utilisera la fonction par défaut
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Fatal(agent.StartServer())
}
```

**Utilisation :**
```bash
curl -X POST http://localhost:3500/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "Bonjour !"}}'
```

### 2. Server Agent avec outils

```go
func executeFunction(functionName, arguments string) (string, error) {
    // Implémentation de vos outils
    switch functionName {
    case "sayHello":
        return "Hello " + arguments, nil
    default:
        return "", fmt.Errorf("unknown function: %s", functionName)
    }
}

func main() {
    ctx := context.Background()

    // Créer le server agent
    serverAgent, err := server.NewAgent(
        ctx,
        agentConfig,
        modelConfig,
        ":8080",
        executeFunction,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Créer le tools agent
    toolsAgent, err := tools.NewAgent(
        ctx,
        agentConfig,
        models.Config{
            Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature:       models.Float64(0.0),
            ParallelToolCalls: models.Bool(true),
        },
        tools.WithTools(GetToolsIndex()),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Attacher le tools agent
    serverAgent.SetToolsAgent(toolsAgent)

    log.Fatal(serverAgent.StartServer())
}
```

**Workflow d'appel d'outil :**

```bash
# 1. Déclencher un appel d'outil
curl -X POST http://localhost:8080/completion \
  -d '{"data":{"message":"Dis bonjour à Alice"}}'

# Réponse SSE contient :
# data: {"kind":"tool_call","status":"pending","operation_id":"op_0x140003dcbe0",...}

# 2. Valider l'opération
curl -X POST http://localhost:8080/operation/validate \
  -d '{"operation_id":"op_0x140003dcbe0"}'

# Ou 3. Annuler l'opération
curl -X POST http://localhost:8080/operation/cancel \
  -d '{"operation_id":"op_0x140003dcbe0"}'
```

### 3. Server Agent complet (Tools + RAG + Compression)

```go
func main() {
    ctx := context.Background()

    // Créer le server agent
    serverAgent, err := server.NewAgent(
        ctx,
        agentConfig,
        modelConfig,
        ":8080",
        executeFunction,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Ajouter le tools agent
    toolsAgent, err := tools.NewAgent(
        ctx,
        toolsConfig,
        toolsModelConfig,
        tools.WithTools(GetToolsIndex()),
    )
    serverAgent.SetToolsAgent(toolsAgent)

    // Ajouter le RAG agent
    ragAgent, err := rag.NewAgent(ctx, ragConfig, embeddingsModelConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Charger des documents
    documents := []string{
        "Contenu du document 1...",
        "Contenu du document 2...",
    }

    for _, content := range documents {
        chunks := chunks.SplitMarkdownBySections(content)
        for _, chunk := range chunks {
            ragAgent.SaveEmbedding(chunk)
        }
    }
    serverAgent.SetRagAgent(ragAgent)
    serverAgent.SetSimilarityLimit(0.6)
    serverAgent.SetMaxSimilarities(3)

    // Ajouter le compressor agent
    compressorAgent, err := compressor.NewAgent(
        ctx,
        compressorConfig,
        compressorModelConfig,
        compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
    )
    if err != nil {
        log.Fatal(err)
    }
    serverAgent.SetCompressorAgent(compressorAgent)
    serverAgent.SetContextSizeLimit(3000)

    log.Fatal(serverAgent.StartServer())
}
```

## Méthodes de configuration

### Configuration du serveur
```go
agent.SetPort(":8080")
port := agent.GetPort()
```

### Configuration des agents
```go
agent.SetToolsAgent(toolsAgent)
agent.SetRagAgent(ragAgent)
agent.SetCompressorAgent(compressorAgent)
```

### Configuration RAG
```go
agent.SetSimilarityLimit(0.6)      // Seuil de similarité (0.0 - 1.0)
agent.SetMaxSimilarities(3)        // Nombre max de documents à récupérer
```

### Configuration de la compression
```go
agent.SetContextSizeLimit(8000)    // Limite en tokens
agent.CompressChatAgentContext()   // Compression forcée
agent.CompressChatAgentContextIfOverLimit() // Compression conditionnelle
```

### Gestion de la mémoire
```go
agent.ResetMessages()              // Effacer l'historique
agent.AddMessage(role, content)    // Ajouter un message
messages := agent.GetMessages()    // Récupérer tous les messages
size := agent.GetContextSize()     // Obtenir la taille du contexte
json := agent.ExportMessagesToJSON() // Exporter en JSON
```

### Fonction d'exécution personnalisée
```go
agent.SetExecuteFunction(func(functionName, arguments string) (string, error) {
    // Votre logique d'exécution
    return result, nil
})
```

## Format des messages SSE

### Chunk de texte normal
```
data: {"message": "chunk de texte"}
```

### Notification d'appel d'outil
```json
data: {
  "kind": "tool_call",
  "status": "pending",
  "operation_id": "op_0x140003dcbe0",
  "function_name": "sayHello",
  "arguments": "{\"name\":\"Alice\"}",
  "message": "Appel d'outil détecté : sayHello"
}
```

### Fin de réponse
```
data: {"message": "", "finish_reason": "stop"}
```

### Erreur
```
data: {"error": "message d'erreur"}
```

## Sécurité et concurrence

### Thread-safety
- **Map des opérations protégée par mutex** pour les appels d'outils concurrents
- **Communication par canaux** pour les réponses d'opération
- **Verrouillage du canal de notifications** pour éviter les race conditions

### Gestion des opérations
- Chaque appel d'outil reçoit un ID unique
- Les opérations en attente sont stockées dans une map thread-safe
- Les canaux de réponse permettent une communication asynchrone
- Timeout et annulations gérés proprement

## Logging

Le Server Agent utilise le système de logging du SDK N.O.V.A. :

```bash
# Contrôle via variable d'environnement
export NOVA_LOG_LEVEL=DEBUG  # DEBUG, INFO, WARN, ERROR
```

## Valeurs par défaut

| Paramètre | Valeur par défaut | Description |
|-----------|-------------------|-------------|
| Port | Défini à la création | Port HTTP du serveur |
| Similarity Limit | 0.6 | Seuil de similarité RAG (0.0-1.0) |
| Max Similarities | 3 | Nombre max de documents RAG |
| Context Size Limit | 8000 | Limite de tokens avant compression |
| Compression Warning | 80% | Avertissement de limite proche |
| Compression Reset | 90% | Réinitialisation forcée |

## Scripts de test

Le répertoire contient plusieurs scripts de test :

- **`stream.sh`** - Test de complétion en streaming basique
- **`call-tool.sh`** - Déclenchement d'appels d'outils
- **`validate.sh`** - Approbation d'opération en attente
- **`cancel.sh`** - Rejet d'opération en attente
- **`reset.sh`** - Annulation de toutes les opérations
- **`test-api.sh`** - Test complet de tous les endpoints

## Exemples de répertoire

Des exemples complets sont disponibles dans `/samples` :

- **`49-server-agent-stream`** - Server agent basique avec chat uniquement
- **`50-server-agent-with-tools`** - Server agent avec appels d'outils
- **`54-server-agent-tools-rag-compress`** - Configuration complète avec tous les agents

## Fonctionnalités futures

Basé sur les fichiers de tâches dans `/nova-sdk/agents/server/tasks/` :

- **Capacité multi-chat agents** - Orchestration de plusieurs agents de chat
- **Ajout d'endpoints personnalisés** - Extension de l'API avec des endpoints custom
- Améliorations supplémentaires du RAG et du compresseur

## Conclusion

Le **Server Agent** est un composant puissant et flexible du SDK N.O.V.A. qui transforme n'importe quel agent conversationnel en une API REST complète avec :

- ✅ Streaming en temps réel via SSE
- ✅ Appels d'outils avec workflow de validation
- ✅ Récupération de documents pertinents (RAG)
- ✅ Compression intelligente du contexte
- ✅ Gestion complète de la mémoire conversationnelle
- ✅ Architecture modulaire et extensible

Il permet de déployer rapidement des assistants IA conversationnels avec des capacités avancées, tout en maintenant un contrôle fin sur les opérations sensibles comme l'exécution d'outils.
