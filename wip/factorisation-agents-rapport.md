# Rapport d'Analyse - Factorisation du Package Agents

**Date:** 2025-12-23
**Package:** `/nova-sdk/agents`
**Lignes totales:** ~6,767 lignes
**Potentiel de factorisation:** ~2,500 lignes (37%)

---

## 📊 Vue d'ensemble

Le package agents contient 43 fichiers Go répartis sur 9 types d'agents:
- **chat** - Agent de chat basique (653 lignes)
- **tools** - Agent avec appels d'outils (1,380 lignes)
- **rag** - Agent RAG/embedding (733 lignes)
- **compressor** - Agent de compression de contexte (378 lignes)
- **structured** - Agent avec sortie structurée (391 lignes)
- **crew** - Orchestration multi-agents (631 lignes)
- **server** - Wrapper HTTP pour chat agent (766 lignes)
- **crewserver** - Wrapper HTTP pour crew agent (1,058 lignes)
- **remote** - Client HTTP distant (586 lignes)

---

## 🔴 PRIORITÉ HAUTE

### 1. Structure BaseAgent dupliquée (~400 lignes économisées)

**Problème:** Struct BaseAgent quasi-identique dans 4+ agents.

**Fichiers concernés:**
- `/nova-sdk/agents/chat/chat.base.agent.go:16-24`
- `/nova-sdk/agents/tools/tools.base.agent.go:15-22`
- `/nova-sdk/agents/compressor/compressor.base.agent.go:16-25`
- `/nova-sdk/agents/structured/structured.base.agent.go:17-24`

**Code dupliqué:**
```go
type BaseAgent struct {
    ctx                  context.Context
    config               agents.Config
    chatCompletionParams openai.ChatCompletionNewParams
    openaiClient         openai.Client
    log                  logger.Logger
    // +champs optionnels par type d'agent
}
```

**Solution proposée:**
```go
// Fichier: /nova-sdk/agents/base/base.agent.go
package base

type Agent struct {
    Ctx                  context.Context
    Config               agents.Config
    ChatCompletionParams openai.ChatCompletionNewParams
    OpenaiClient         openai.Client
    Log                  logger.Logger
}
```

---

### 2. Initialisation NewBaseAgent dupliquée (~150 lignes économisées)

**Problème:** Constructeurs NewBaseAgent identiques dans 4 agents.

**Fichiers concernés:**
- `/nova-sdk/agents/chat/chat.base.agent.go:28-60`
- `/nova-sdk/agents/tools/tools.base.agent.go:27-59`
- `/nova-sdk/agents/compressor/compressor.base.agent.go:37-71`
- `/nova-sdk/agents/structured/structured.base.agent.go:28-96`

**Pattern dupliqué:**
```go
func NewBaseAgent(...) (*BaseAgent, error) {
    client, log, err := agents.InitializeConnection(ctx, agentConfig, models.Config{
        Name: modelConfig.Model,
    })
    if err != nil {
        return nil, err
    }

    agent := &BaseAgent{
        ctx:                  ctx,
        config:               agentConfig,
        chatCompletionParams: modelConfig,
        openaiClient:         client,
        log:                  log,
    }

    agent.chatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{}
    agent.chatCompletionParams.Messages = append(..., openai.SystemMessage(agentConfig.SystemInstructions))

    // Apply options...
    return agent, nil
}
```

**Solution:** Créer une fonction d'initialisation partagée dans le package base.

---

### 3. Méthodes de gestion des messages (~200 lignes économisées)

**Problème:** Méthodes identiques de gestion des messages dans tous les BaseAgent.

**Fichiers concernés:**
- `/nova-sdk/agents/chat/chat.base.agent.go:66-164`
- `/nova-sdk/agents/tools/tools.base.agent.go:65-102`
- `/nova-sdk/agents/compressor/compressor.base.agent.go:74-84`
- `/nova-sdk/agents/structured/structured.base.agent.go:102-130`

**Méthodes dupliquées:**
- `GetMessages()` - Retourne la liste des messages
- `AddMessage()` - Ajoute un message à l'historique
- `GetStringMessages()` - Convertit en string messages
- `GetCurrentContextSize()` - Calcule la taille du contexte
- `ResetMessages()` - Reset au message système uniquement
- `SetSystemInstructions()` - Met à jour le message système

**Exemple de duplication:**
```go
// Apparaît dans chat, tools, structured agents
func (agent *BaseAgent) ResetMessages() {
    if len(agent.chatCompletionParams.Messages) > 0 {
        firstMsg := agent.chatCompletionParams.Messages[0]
        if firstMsg.OfSystem != nil {
            agent.chatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{firstMsg}
        } else {
            agent.chatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{}
        }
    }
}
```

**Solution:** Déplacer toutes ces méthodes dans le BaseAgent partagé.

---

### 4. Server vs CrewServer (~700 lignes économisées!)

**Problème:** Duplication MASSIVE entre ServerAgent et CrewServerAgent (95% identique).

**Fichiers concernés:**
- `/nova-sdk/agents/server/server.agent.go` (227 lignes)
- `/nova-sdk/agents/crewserver/crew.server.agent.go` (302 lignes)

**Code dupliqué:**

**A) Structures (lignes 20-79 dans les deux fichiers):**
```go
type ServerAgent struct {
    chatAgent / currentChatAgent *chat.Agent
    toolsAgent       *tools.Agent
    ragAgent         *rag.Agent
    similarityLimit  float64
    maxSimilarities  int
    contextSizeLimit int
    compressorAgent  *compressor.Agent
    port             string
    ctx              context.Context
    log              logger.Logger
    pendingOperations map[string]*PendingOperation
    operationsMutex   sync.RWMutex
    stopStreamChan    chan bool
    currentNotificationChan chan ToolCallNotification
    notificationChanMutex   sync.Mutex
    executeFn        func(string, string) (string, error)
}

// Types de support identiques
type ToolCallNotification struct { ... }
type PendingOperation struct { ... }
type CompletionRequest struct { ... }
type OperationRequest struct { ... }
type MemoryResponse struct { ... }
type TokensResponse struct { ... }
```

**B) Gestion des routes (lignes 202-219 server, 277-294 crewserver):**
```go
func (agent *ServerAgent) StartServer() error {
    mux := http.NewServeMux()
    mux.HandleFunc("POST /completion", agent.handleCompletion)
    mux.HandleFunc("POST /completion/stop", agent.handleCompletionStop)
    mux.HandleFunc("POST /memory/reset", agent.handleMemoryReset)
    mux.HandleFunc("GET /memory/messages/list", agent.handleMessagesList)
    mux.HandleFunc("GET /memory/messages/tokens", agent.handleTokensCount)
    mux.HandleFunc("POST /operation/validate", agent.handleOperationValidate)
    mux.HandleFunc("POST /operation/cancel", agent.handleOperationCancel)
    mux.HandleFunc("POST /operation/reset", agent.handleOperationReset)
    mux.HandleFunc("GET /models", agent.handleModelsInformation)
    mux.HandleFunc("GET /health", agent.handleHealth)
    return http.ListenAndServe(agent.port, mux)
}
```

**C) Méthodes de délégation (lignes 135-199 server, 210-274 crewserver):**
```go
func (agent *ServerAgent) GetMessages() []messages.Message {
    return agent.chatAgent.GetMessages()
}
func (agent *ServerAgent) GetContextSize() int {
    return agent.chatAgent.GetContextSize()
}
// ... 10+ méthodes identiques
```

**Solution:** Créer un package `agents/serverbase` avec l'infrastructure HTTP commune.

---

### 5. Handlers HTTP dupliqués (~300 lignes économisées)

**Problème:** Handlers HTTP identiques entre server et crewserver.

**Fichiers concernés:**
- `/nova-sdk/agents/server/handlers.completion.go` (226 lignes)
- `/nova-sdk/agents/crewserver/handlers.completion.go` (245 lignes)
- `/nova-sdk/agents/server/handlers.operations.go` (142 lignes)
- `/nova-sdk/agents/crewserver/handlers.operations.go` (142 lignes)
- `/nova-sdk/agents/server/handlers.go` (80 lignes)
- `/nova-sdk/agents/crewserver/handlers.go` (80 lignes)

**Handlers 100% identiques:**
- `handleHealth` - Health check endpoint
- `handleModelsInformation` - Model info endpoint
- `handleMemoryReset` - Reset memory endpoint
- `handleMessagesList` - List messages endpoint
- `handleTokensCount` - Token count endpoint
- `handleOperationValidate` - Validate operation endpoint
- `handleOperationCancel` - Cancel operation endpoint
- `handleOperationReset` - Reset operations endpoint

**Solution:** Extraire les handlers dans un package partagé avec des interfaces.

---

## 🟡 PRIORITÉ MOYENNE

### 6. Pattern Wrapper Agent (~200 lignes économisées)

**Problème:** Tous les agents high-level (chat.Agent, tools.Agent, rag.Agent, etc.) suivent le même pattern.

**Pattern commun:**
```go
type Agent struct {
    ctx           context.Context
    config        agents.Config
    modelConfig   models.Config
    internalAgent *BaseAgent
    log           logger.Logger
}

func NewAgent(ctx, agentConfig, modelConfig) (*Agent, error) {
    log := logger.GetLoggerFromEnv()
    openaiModelConfig := models.ConvertToOpenAIModelConfig(modelConfig)
    internalAgent, err := NewBaseAgent(ctx, agentConfig, openaiModelConfig)
    if err != nil {
        return nil, err
    }
    agent := &Agent{...}
    return agent, nil
}

func (agent *Agent) Kind() agents.Kind { return agents.XXX }
func (agent *Agent) GetName() string { return agent.config.Name }
func (agent *Agent) GetModelID() string { return agent.modelConfig.Name }
```

**Fichiers:** chat.agent.go, tools.agent.go, compressor.agent.go, structured.agent.go

---

### 7. Méthodes "related" dupliquées (~200 lignes économisées)

**Problème:** Fichiers methods.*.related.go dupliqués entre crew, server et crewserver.

**Fichiers identiques:**
- Compression: `crew/methods.compression.related.go`, `server/methods.compression.related.go`, `crewserver/methods.compression.related.go`
- RAG: `crew/methods.rag.related.go`, `server/methods.rag.related.go`, `crewserver/methods.rag.related.go`
- Tools: `crew/methods.tools.related.go`, `server/methods.tools.related.go`, `crewserver/methods.tools.related.go`

**Exemple (identique partout):**
```go
func (agent *XXXAgent) SetRagAgent(ragAgent *rag.Agent) {
    agent.ragAgent = ragAgent
}
func (agent *XXXAgent) GetRagAgent() *rag.Agent {
    return agent.ragAgent
}
```

---

### 8. Logique Tool Calls (~600 lignes économisées)

**Problème:** 6 méthodes dans tools.base.agent.go avec 80% de code commun.

**Méthodes:**
- `DetectParallelToolCalls`
- `DetectParallelToolCallsWithConfirmation`
- `DetectToolCallsLoop`
- `DetectToolCallsLoopWithConfirmation`
- `DetectToolCallsLoopStream`
- `DetectToolCallsLoopWithConfirmationStream`

**Pattern commun:**
1. Création des paramètres de tool calls
2. Exécution des tool calls
3. Gestion des finish reasons (tool_calls, stop, error)

---

### 9. Logique Streaming (~150 lignes économisées)

**Problème:** Pattern de streaming dupliqué.

**Fichiers:**
- `/nova-sdk/agents/chat/chat.base.agent.go:235-396`
- `/nova-sdk/agents/compressor/compressor.base.agent.go:133-207`

**Pattern commun:**
```go
stream := agent.openaiClient.Chat.Completions.NewStreaming(ctx, params)
var callBackError error
finalFinishReason := ""

for stream.Next() {
    chunk := stream.Current()

    if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
        finalFinishReason = chunk.Choices[0].FinishReason
    }

    if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
        callBackError = callBack(chunk.Choices[0].Delta.Content, finalFinishReason)
        response += chunk.Choices[0].Delta.Content
    }

    if callBackError != nil {
        break
    }
}
```

---

### 10. Crew vs CrewServer orchestration (~250 lignes économisées)

**Problème:** CrewAgent et CrewServerAgent partagent 80% de leur logique d'orchestration.

**Fonctionnalités partagées:**
- Gestion de la map de chat agents
- Sélection de l'agent courant
- Gestion des tools/RAG/Compressor agents
- Logique de routing
- Détection de topics
- Méthodes de délégation

**Différence:** CrewServerAgent ajoute simplement une couche HTTP.

---

## 📈 RÉCAPITULATIF DES GAINS

| Catégorie | Lignes économisées | Priorité |
|-----------|-------------------|----------|
| BaseAgent consolidation | ~400 | 🔴 Haute |
| Server/CrewServer unification | ~700 | 🔴 Haute |
| Tool call processing | ~600 | 🔴 Haute |
| Wrapper pattern | ~200 | 🟡 Moyenne |
| Methods files | ~200 | 🟡 Moyenne |
| Stream processing | ~150 | 🟡 Moyenne |
| Crew orchestration | ~250 | 🟡 Moyenne |
| **TOTAL** | **~2,500** | **37% du code** |

---

## 🎯 PLAN DE REFACTORING RECOMMANDÉ

### Phase 1: Fondation (BaseAgent)
1. ✅ Créer `/nova-sdk/agents/base` package
2. ✅ Implémenter BaseAgent partagé avec tous les champs communs
3. ✅ Migrer les méthodes de gestion des messages
4. ✅ Créer fonction d'initialisation commune
5. ✅ Migrer chat, tools, compressor, structured vers la nouvelle base

### Phase 2: Infrastructure Serveur
6. ✅ Créer `/nova-sdk/agents/serverbase` package
7. ✅ Extraire structures communes (ServerAgent, types de support)
8. ✅ Extraire handlers HTTP partagés
9. ✅ Implémenter interface pour délégation chat/crew
10. ✅ Refactoriser server et crewserver

### Phase 3: Optimisations Tools
11. ⏳ Extraire helpers de création de tool call params
12. ⏳ Consolider logique d'exécution des tools
13. ⏳ Unifier gestion des finish reasons

### Phase 4: Utilitaires
14. ⏳ Extraire utilitaires de streaming
15. ⏳ Créer mixins pour RAG/Tools/Compression
16. ⏳ Refactoriser crew/crewserver orchestration

---

## ✅ BÉNÉFICES ATTENDUS

1. **Maintenabilité**: Source unique de vérité (Single Source of Truth)
2. **Corrections de bugs**: Fix une fois, s'applique partout
3. **Tests**: Tester le code partagé une seule fois
4. **Cohérence**: Même comportement dans tous les agents
5. **Extensibilité**: Plus facile d'ajouter de nouveaux types d'agents
6. **Réduction de la dette technique**: -37% de code dupliqué

---

## 🔄 COMPATIBILITÉ

Toutes les modifications doivent maintenir la compatibilité ascendante (backward compatibility) via:
- Interfaces bien définies
- Méthodes de délégation
- Composition plutôt qu'héritage
- Migration progressive package par package
