# Crew Server Agent - Présentation Fonctionnelle

## Vue d'ensemble

Le **Crew Server Agent** est un système avancé d'orchestration multi-agents du SDK N.O.V.A. qui gère plusieurs agents conversationnels spécialisés et route intelligemment les requêtes vers l'agent le plus approprié en fonction de la détection du sujet.

Il s'agit d'un composant HTTP/REST API sophistiqué qui étend les capacités du Server Agent classique en ajoutant une orchestration intelligente d'agents, un routage dynamique et une commutation transparente du contexte entre des experts de domaine spécialisés.

## Caractéristiques principales

### 1. Gestion de crew multi-agents

- **Plusieurs agents conversationnels spécialisés** gérés dans un crew
- **Commutation dynamique d'agents** basée sur les sujets de conversation
- **Gestion du crew à l'exécution** : ajout ou suppression d'agents à la volée
- **Expertise de domaine spécialisé** : chaque agent peut être optimisé pour des sujets spécifiques
- **Délégation transparente** : apparaît comme un seul assistant cohérent pour les utilisateurs

### 2. Détection intelligente de sujets et routage

- **Classification automatique de sujets** utilisant un agent orchestrateur structuré
- **Logique de routage personnalisée** via fonction de matching configurable
- **Commutation d'agents transparente** tout en maintenant le contexte de conversation
- **Mécanisme de secours** pour les sujets non reconnus
- **Délégation basée sur l'intention** pour une qualité de réponse optimale

### 3. API REST avec Streaming SSE

- **Endpoints HTTP RESTful** pour les complétions de chat
- **Streaming en temps réel** via Server-Sent Events (SSE)
- **Réponses progressives** pour une expérience utilisateur fluide
- **Gestion des erreurs** en temps réel via SSE

### 4. Appel d'outils avec workflow de confirmation

- **Détection automatique** des appels d'outils dans les requêtes utilisateur
- **Appels parallèles** de multiples outils
- **Workflow de validation web** : les outils requièrent une approbation avant exécution
- **Système de notifications** via SSE pour informer le client des opérations en attente
- **Gestion des opérations** : validation, annulation, réinitialisation

### 5. RAG (Retrieval-Augmented Generation)

- **Recherche de similarité** dans une base de documents vectoriels
- **Injection automatique** de contexte pertinent avant génération
- **Configuration flexible** : seuil de similarité, nombre maximal de documents
- **Amélioration des réponses** grâce à des informations externes

### 6. Compression intelligente du contexte

- **Gestion automatique** de la limite de tokens
- **Compression à la demande** ou automatique au-delà d'un seuil
- **Sauvegardes** : avertissement à 80%, réinitialisation à 90%
- **Préservation du contexte** essentiel tout en respectant les limites du modèle

### 7. Gestion de la mémoire conversationnelle

- **Historique des messages** persistant durant la session
- **Comptage de tokens** pour surveiller l'utilisation
- **Réinitialisation** de l'historique
- **Export JSON** des conversations
- **Listage** de tous les messages

## Architecture

### Structure principale

```go
type CrewServerAgent struct {
    // Gestion du crew
    chatAgents       map[string]*chat.Agent // Map des agents spécialisés
    currentChatAgent *chat.Agent            // Agent actuellement actif

    // Agents spécialisés
    toolsAgent        *tools.Agent           // Optionnel : détection/exécution d'outils
    ragAgent          *rag.Agent             // Optionnel : récupération de documents
    compressorAgent   *compressor.Agent      // Optionnel : compression du contexte
    orchestratorAgent *structured.Agent[Intent] // Optionnel : détection de sujets

    // Configuration de l'orchestration
    matchAgentIdToTopicFn func(string) string // Logique de routage personnalisée

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

### Structure de détection d'intention

```go
type Intent struct {
    TopicDiscussion string `json:"topic_discussion"`
}
```

### Agents spécialisés

Le Crew Server Agent orchestre cinq types d'agents :

1. **Agents de Chat** (obligatoire) : plusieurs agents conversationnels spécialisés, chacun avec une expertise de domaine
2. **Agent Orchestrateur** (optionnel) : agent structuré pour la détection de sujets et le routage
3. **Agent d'Outils** (optionnel) : détection et exécution de fonctions
4. **Agent RAG** (optionnel) : recherche de documents pertinents
5. **Agent Compresseur** (optionnel) : compression du contexte

## Endpoints HTTP

### Complétion

#### `POST /completion`

Stream une complétion de chat avec routage intelligent d'agents et SSE.

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
    { "role": "user", "content": "..." },
    { "role": "assistant", "content": "..." }
  ]
}
```

#### `GET /memory/messages/context-size`

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
  "chat_models": {
    "coder": "hf.co/menlo/jan-nano-gguf:q4_k_m",
    "cook": "hf.co/menlo/jan-nano-gguf:q4_k_m",
    "thinker": "hf.co/menlo/jan-nano-gguf:q4_k_m"
  },
  "tools_model": "hf.co/menlo/jan-nano-gguf:q4_k_m",
  "embeddings_model": "all-minilm:l6-v2",
  "orchestrator_model": "hf.co/menlo/jan-nano-gguf:q4_k_m"
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

6. **Détection de sujet et routage d'agent** (si orchestrator agent configuré)
   - Appel de `DetectTopicThenGetAgentId()`
   - Analyse de la requête utilisateur avec l'agent orchestrateur
   - Extraction du sujet depuis la réponse structurée
   - Matching du sujet vers un ID d'agent via `matchAgentIdToTopicFn`
   - Commutation du `currentChatAgent` si un agent différent est détecté

7. **Génération de la complétion en streaming**
   - Streaming des chunks via SSE en utilisant l'agent de chat sélectionné
   - Gestion des signaux d'arrêt
   - Envoi de la raison de fin

8. **Gestion des erreurs**
   - Streaming des erreurs via SSE

## Exemples d'utilisation

### 1. Crew Server Agent basique (Plusieurs agents spécialisés)

```go
import (
    "context"
    "log"
    "strings"

    "nova-sdk/agents"
    "nova-sdk/agents/chat"
    "nova-sdk/agents/crewserver"
    "nova-sdk/models"
)

func main() {
    ctx := context.Background()

    // Créer des agents de chat spécialisés
    coderAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "coder",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "Vous êtes un expert en programmation spécialisé en Go, Python et développement web.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.3),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    cookAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "cook",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "Vous êtes un chef professionnel spécialisé dans la cuisine mondiale et les techniques culinaires.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.7),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    thinkerAgent, err := chat.NewAgent(
        ctx,
        agents.Config{
            Name:               "thinker",
            EngineURL:          "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: "Vous êtes un philosophe et psychologue spécialisé dans la pensée critique et le développement personnel.",
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.6),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Créer le crew d'agents
    agentCrew := map[string]*chat.Agent{
        "coder":   coderAgent,
        "cook":    cookAgent,
        "thinker": thinkerAgent,
    }

    // Définir la logique de routage
    matchAgentFunction := func(topic string) string {
        switch strings.ToLower(topic) {
        case "coding", "programming", "development", "software":
            return "coder"
        case "cooking", "recipe", "food", "cuisine":
            return "cook"
        case "philosophy", "psychology", "thinking", "mindfulness":
            return "thinker"
        default:
            return "coder" // Agent par défaut
        }
    }

    // Créer le crew server agent
    crewAgent, err := crewserver.NewAgent(
        ctx,
        agentCrew,
        "coder", // Agent actif initial
        ":8080",
        matchAgentFunction,
        nil, // Pas de fonction d'exécution personnalisée
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Fatal(crewAgent.StartServer())
}
```

**Utilisation :**

```bash
# Question de programmation - routée vers l'agent coder
curl -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "Comment utiliser switch case en Golang ?"}}'

# Question de cuisine - routée vers l'agent cook
curl -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "Quelle est la recette de la pizza hawaïenne ?"}}'

# Question de psychologie - routée vers l'agent thinker
curl -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "Comment gérer l'anxiété ?"}}'
```

### 2. Crew Server Agent avec orchestration intelligente

```go
import (
    "nova-sdk/agents/structured"
)

func main() {
    ctx := context.Background()

    // Créer le crew d'agents (comme ci-dessus)
    agentCrew := map[string]*chat.Agent{
        "coder":   coderAgent,
        "cook":    cookAgent,
        "thinker": thinkerAgent,
    }

    // Créer le crew server agent
    crewAgent, err := crewserver.NewAgent(
        ctx,
        agentCrew,
        "coder",
        ":8080",
        matchAgentFunction,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Créer l'agent orchestrateur pour la détection de sujets
    orchestratorAgent, err := structured.NewAgent[crewserver.Intent](
        ctx,
        agents.Config{
            Name:      "orchestrator",
            EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions: `Vous êtes un classificateur de sujets.
Analysez la requête utilisateur et déterminez le sujet principal.
Sujets possibles : programming, cooking, philosophy, psychology, general.`,
        },
        models.Config{
            Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
            Temperature: models.Float64(0.0),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Attacher l'orchestrateur pour le routage intelligent
    crewAgent.SetOrchestratorAgent(orchestratorAgent)

    log.Fatal(crewAgent.StartServer())
}
```

**Comment ça fonctionne :**

1. L'utilisateur envoie : "Quelle est la meilleure façon d'implémenter une API REST ?"
2. L'orchestrateur détecte le sujet : "programming"
3. `matchAgentFunction` mappe "programming" vers "coder"
4. La requête est routée vers l'agent coder
5. L'agent coder génère la réponse

### 3. Crew Server Agent complet (Orchestration + Outils + RAG + Compression)

```go
import (
    "nova-sdk/agents/tools"
    "nova-sdk/agents/rag"
    "nova-sdk/agents/compressor"
    "nova-sdk/chunks"
)

func executeFunction(functionName, arguments string) (string, error) {
    // Implémentation de vos outils
    switch functionName {
    case "calculator":
        // Implémenter la logique de calcul
        return "42", nil
    default:
        return "", fmt.Errorf("unknown function: %s", functionName)
    }
}

func main() {
    ctx := context.Background()

    // Créer le crew (comme ci-dessus)
    agentCrew := map[string]*chat.Agent{
        "coder":   coderAgent,
        "cook":    cookAgent,
        "thinker": thinkerAgent,
    }

    // Créer le crew server agent
    crewAgent, err := crewserver.NewAgent(
        ctx,
        agentCrew,
        "coder",
        ":8080",
        matchAgentFunction,
        executeFunction,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Ajouter l'orchestrateur
    orchestratorAgent, err := structured.NewAgent[crewserver.Intent](
        ctx,
        orchestratorConfig,
        orchestratorModelConfig,
    )
    crewAgent.SetOrchestratorAgent(orchestratorAgent)

    // Ajouter le tools agent
    toolsAgent, err := tools.NewAgent(
        ctx,
        toolsConfig,
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
    crewAgent.SetToolsAgent(toolsAgent)

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
    crewAgent.SetRagAgent(ragAgent)
    crewAgent.SetSimilarityLimit(0.6)
    crewAgent.SetMaxSimilarities(3)

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
    crewAgent.SetCompressorAgent(compressorAgent)
    crewAgent.SetContextSizeLimit(3000)

    log.Fatal(crewAgent.StartServer())
}
```

## Méthodes de configuration

### Gestion du crew

```go
agent.AddChatAgentToCrew("expert", expertAgent)    // Ajouter un nouvel agent
agent.RemoveChatAgentFromCrew("expert")            // Supprimer un agent
agents := agent.GetChatAgents()                    // Obtenir tous les agents
agent.SetChatAgents(newAgentMap)                   // Remplacer tous les agents
```

### Configuration du serveur

```go
agent.SetPort(":8080")
port := agent.GetPort()
```

### Configuration des agents

```go
agent.SetOrchestratorAgent(orchestratorAgent)  // Activer le routage intelligent
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
  "function_name": "calculator",
  "arguments": "{\"a\":40,\"b\":2}",
  "message": "Appel d'outil détecté : calculator"
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

## Détection de sujets et routage

### Comment ça fonctionne

Le système d'orchestration utilise un processus en deux étapes :

1. **Détection de sujet** : L'agent orchestrateur analyse la requête utilisateur et détermine le sujet principal
2. **Matching d'agent** : La fonction `matchAgentIdToTopicFn` mappe le sujet détecté vers un ID d'agent spécifique

### Exemple de configuration de routage

```go
matchAgentFunction := func(topic string) string {
    topicLower := strings.ToLower(topic)

    // Sujets liés à la programmation
    if strings.Contains(topicLower, "coding") ||
       strings.Contains(topicLower, "programming") ||
       strings.Contains(topicLower, "software") ||
       strings.Contains(topicLower, "development") {
        return "coder"
    }

    // Sujets liés à la cuisine
    if strings.Contains(topicLower, "cooking") ||
       strings.Contains(topicLower, "recipe") ||
       strings.Contains(topicLower, "food") ||
       strings.Contains(topicLower, "cuisine") {
        return "cook"
    }

    // Sujets de philosophie/psychologie
    if strings.Contains(topicLower, "philosophy") ||
       strings.Contains(topicLower, "psychology") ||
       strings.Contains(topicLower, "thinking") ||
       strings.Contains(topicLower, "mindfulness") {
        return "thinker"
    }

    // Secours par défaut
    return "coder"
}
```

### Règles de commutation d'agents

- Le système peut changer d'agent entre les requêtes
- La commutation d'agent est transparente et maintient le contexte de conversation
- Impossible de supprimer l'agent actuellement actif du crew
- Chaque agent a sa propre personnalité et instructions système

## Sécurité et concurrence

### Thread-safety

- **Map des opérations protégée par mutex** pour les appels d'outils concurrents
- **Communication par canaux** pour les réponses d'opération
- **Verrouillage du canal de notifications** pour éviter les race conditions
- **Gestion sécurisée du crew** avec traitement concurrent des requêtes

### Gestion des opérations

- Chaque appel d'outil reçoit un ID unique
- Les opérations en attente sont stockées dans une map thread-safe
- Les canaux de réponse permettent une communication asynchrone
- Timeout et annulations gérés proprement

## Logging

Le Crew Server Agent utilise le système de logging du SDK N.O.V.A. :

```bash
# Contrôle via variable d'environnement
export NOVA_LOG_LEVEL=DEBUG  # DEBUG, INFO, WARN, ERROR
```

## Valeurs par défaut

| Paramètre           | Valeur par défaut    | Description                        |
| ------------------- | -------------------- | ---------------------------------- |
| Port                | Défini à la création | Port HTTP du serveur               |
| Similarity Limit    | 0.6                  | Seuil de similarité RAG (0.0-1.0)  |
| Max Similarities    | 3                    | Nombre max de documents RAG        |
| Context Size Limit  | 8000                 | Limite de tokens avant compression |
| Compression Warning | 80%                  | Avertissement de limite proche     |
| Compression Reset   | 90%                  | Réinitialisation forcée            |

## Scripts de test

Le répertoire contient plusieurs scripts de test :

- **`01-programming-stream.sh`** - Test de question de programmation (routée vers coder)
- **`02-cooking-stream.sh`** - Test de question de cuisine (routée vers cook)
- **`03-psychology-stream.sh`** - Test de question de psychologie (routée vers thinker)
- **`call-tool.sh`** - Déclenchement d'appels d'outils
- **`validate.sh`** - Approbation d'opération en attente
- **`cancel.sh`** - Rejet d'opération en attente
- **`reset.sh`** - Annulation de toutes les opérations

## Exemples de répertoire

Des exemples complets sont disponibles dans `/samples` :

- **`55-crew-server-agent`** - Crew server agent complet avec orchestration

## Comparaison : Server Agent vs Crew Server Agent

| Fonctionnalité           | Server Agent           | Crew Server Agent                    |
| ------------------------ | ---------------------- | ------------------------------------ |
| **Agents de Chat**       | Agent unique           | Plusieurs agents spécialisés         |
| **Commutation d'agents** | Non                    | Oui - dynamique selon le sujet       |
| **Orchestration**        | Non                    | Oui - avec agent structuré           |
| **Détection de sujets**  | Non                    | Oui - routage automatique            |
| **Gestion des agents**   | Statique               | Dynamique ajout/suppression d'agents |
| **Logique de routage**   | N/A                    | Fonction de matching personnalisable |
| **Cas d'usage**          | Chatbot à usage unique | Assistant intelligent multi-domaines |
| **Appel d'outils**       | Oui                    | Oui                                  |
| **RAG**                  | Oui                    | Oui                                  |
| **Compression**          | Oui                    | Oui                                  |
| **Streaming SSE**        | Oui                    | Oui                                  |

## Avantages de l'orchestration en crew

1. **Expertise spécialisée** : Chaque agent peut être optimisé pour des domaines spécifiques
2. **Meilleure qualité de réponse** : Des instructions système spécifiques au domaine conduisent à des réponses plus précises
3. **Scalabilité** : Facile d'ajouter de nouveaux agents spécialisés
4. **Flexibilité** : Routage dynamique basé sur le contexte de conversation
5. **Maintenabilité** : Séparation des préoccupations par domaine
6. **Performance** : Différents paramètres de température pour différentes tâches

## Conclusion

Le **Crew Server Agent** est un système puissant d'orchestration multi-agents du SDK N.O.V.A. qui transforme plusieurs agents conversationnels spécialisés en un assistant intelligent unifié avec :

- ✅ Détection intelligente de sujets et routage
- ✅ Plusieurs experts de domaine spécialisés
- ✅ Streaming en temps réel via SSE
- ✅ Appels d'outils avec workflow de validation
- ✅ Récupération de documents pertinents (RAG)
- ✅ Compression intelligente du contexte
- ✅ Gestion complète de la mémoire conversationnelle
- ✅ Gestion dynamique du crew
- ✅ Architecture modulaire et extensible

Il permet de déployer rapidement des assistants IA multi-domaines sophistiqués capables de gérer diverses requêtes utilisateur en déléguant automatiquement à des agents spécialisés, tout en maintenant un contrôle fin sur les opérations sensibles comme l'exécution d'outils.
