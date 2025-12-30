---
id: crew-server-agent
name: Crew Server Agent (Multi-Agent Server)
category: complex
complexity: advanced
sample_source: 56
description: HTTP server exposing an agent crew via REST API
interactive: true
---

# Crew Server Agent (Multi-Agent Server)

## Description

Creates an HTTP server that hosts an AI agent crew and exposes their capabilities via REST API. Allows remote clients to connect and use crew agents.

## Use Cases

- Microservices architecture with AI agents
- Centralizing agents for multiple applications
- Horizontal scaling of AI capabilities
- Agent API for web/mobile applications
- Sharing agents across teams/projects

## ⚠️ Interactive Mode

This snippet requires specific information. Answer the following questions:

### Configuration Questions

1. **Which agents in your crew server?**
   - Types: chat, rag, tools, structured, etc.
   - Number of agents

2. **Which API to expose?**
   - Simple REST
   - WebSocket for streaming
   - gRPC for performance

3. **Authentication required?**
   - API Key
   - JWT
   - OAuth2

4. **Which routes/endpoints?**
   - `/chat` - Conversation
   - `/complete` - Simple completion
   - `/tools` - Tool execution
   - etc.

5. **Scaling configuration?**
   - Connection pooling
   - Rate limiting
   - Load balancing

---

## Base Template

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

// === CONFIGURATION ===
const (
	ServerPort   = ":8080"
	EngineURL    = "http://localhost:12434/engines/llama.cpp/v1"
	APIKeyHeader = "X-API-Key"
	ValidAPIKey  = "your-secret-api-key" // Replace in production
)

// === STRUCTURES ===
type CrewServer struct {
	chatAgent    *chat.Agent
	ragAgent     *rag.Agent
	toolsAgent   *tools.Agent
	mu           sync.RWMutex
	requestCount int64
}

type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
}

type ChatResponse struct {
	Response     string `json:"response"`
	SessionID    string `json:"session_id"`
	FinishReason string `json:"finish_reason"`
	Timestamp    string `json:"timestamp"`
}

type RAGRequest struct {
	Query     string  `json:"query"`
	Threshold float64 `json:"threshold,omitempty"`
	TopK      int     `json:"top_k,omitempty"`
}

type RAGResponse struct {
	Results []RAGResult `json:"results"`
	Query   string      `json:"query"`
}

type RAGResult struct {
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}

type ToolsRequest struct {
	Message string `json:"message"`
}

type ToolsResponse struct {
	Results      []string `json:"results"`
	FinalMessage string   `json:"final_message"`
	FinishReason string   `json:"finish_reason"`
}

type HealthResponse struct {
	Status       string   `json:"status"`
	Agents       []string `json:"agents"`
	RequestCount int64    `json:"request_count"`
	Uptime       string   `json:"uptime"`
}

// === MAIN ===
func main() {
	ctx := context.Background()
	startTime := time.Now()

	// Initialize crew
	server, err := NewCrewServer(ctx)
	if err != nil {
		log.Fatalf("Crew initialization error: %v", err)
	}

	// API routes
	http.HandleFunc("/health", server.handleHealth(startTime))
	http.HandleFunc("/api/chat", server.authMiddleware(server.handleChat))
	http.HandleFunc("/api/rag/search", server.authMiddleware(server.handleRAGSearch))
	http.HandleFunc("/api/rag/index", server.authMiddleware(server.handleRAGIndex))
	http.HandleFunc("/api/tools", server.authMiddleware(server.handleTools))

	fmt.Printf("🚀 Crew Server started on %s\n", ServerPort)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /health          - Server status")
	fmt.Println("  POST /api/chat        - Chat agent")
	fmt.Println("  POST /api/rag/search  - RAG search")
	fmt.Println("  POST /api/rag/index   - RAG indexing")
	fmt.Println("  POST /api/tools       - Tools agent")

	log.Fatal(http.ListenAndServe(ServerPort, nil))
}

// === CREW INITIALIZATION ===
func NewCrewServer(ctx context.Context) (*CrewServer, error) {
	// Chat Agent
	chatAgent, err := chat.NewAgent(ctx,
		agents.Config{
			Name:                    "crew-chat",
			EngineURL:               EngineURL,
			SystemInstructions:      "You are a helpful AI assistant from the Crew Server.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        "ai/qwen2.5:1.5B-F16",
			Temperature: models.Float64(0.7),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("chat agent: %v", err)
	}

	// RAG Agent
	ragAgent, err := rag.NewAgent(ctx,
		agents.Config{
			Name:               "crew-rag",
			EngineURL:          EngineURL,
			SystemInstructions: "You are a semantic search agent.",
		},
		models.Config{
			Name: "ai/mxbai-embed-large",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("rag agent: %v", err)
	}

	// Tools Agent
	toolsAgent, err := tools.NewAgent(ctx,
		agents.Config{
			Name:                    "crew-tools",
			EngineURL:               EngineURL,
			SystemInstructions:      "You are an agent capable of using tools.",
			KeepConversationHistory: true,
		},
		models.Config{
			Name:              "hf.co/menlo/jan-nano-gguf:q4_k_m",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true),
		},
		tools.WithTools(getServerTools()),
	)
	if err != nil {
		return nil, fmt.Errorf("tools agent: %v", err)
	}

	return &CrewServer{
		chatAgent:  chatAgent,
		ragAgent:   ragAgent,
		toolsAgent: toolsAgent,
	}, nil
}

// === AUTH MIDDLEWARE ===
func (s *CrewServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(APIKeyHeader)
		if apiKey != ValidAPIKey {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		s.mu.Lock()
		s.requestCount++
		s.mu.Unlock()
		next(w, r)
	}
}

// === HANDLERS ===
func (s *CrewServer) handleHealth(startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		count := s.requestCount
		s.mu.RUnlock()

		resp := HealthResponse{
			Status:       "healthy",
			Agents:       []string{"chat", "rag", "tools"},
			RequestCount: count,
			Uptime:       time.Since(startTime).String(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func (s *CrewServer) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	result, err := s.chatAgent.GenerateCompletion([]messages.Message{
		{Role: roles.User, Content: req.Message},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	resp := ChatResponse{
		Response:     result.Response,
		SessionID:    req.SessionID,
		FinishReason: result.FinishReason,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *CrewServer) handleRAGSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req RAGRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	threshold := req.Threshold
	if threshold == 0 {
		threshold = 0.5
	}

	similarities, err := s.ragAgent.SearchSimilar(req.Query, threshold)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	var results []RAGResult
	for _, sim := range similarities {
		results = append(results, RAGResult{
			Content:    sim.Prompt,
			Similarity: sim.Similarity,
		})
	}

	resp := RAGResponse{Results: results, Query: req.Query}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *CrewServer) handleRAGIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Documents []string `json:"documents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	indexed := 0
	for _, doc := range req.Documents {
		if err := s.ragAgent.SaveEmbedding(doc); err == nil {
			indexed++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"indexed": indexed,
		"total":   len(req.Documents),
	})
}

func (s *CrewServer) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req ToolsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	result, err := s.toolsAgent.DetectToolCallsLoop(
		[]messages.Message{{Role: roles.User, Content: req.Message}},
		executeServerTool,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	resp := ToolsResponse{
		Results:      result.Results,
		FinalMessage: result.LastAssistantMessage,
		FinishReason: result.FinishReason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// === TOOLS ===
func getServerTools() []*tools.Tool {
	return []*tools.Tool{
		tools.NewTool("calculate").
			SetDescription("Perform a mathematical calculation").
			AddParameter("expression", "string", "Mathematical expression", true),
		tools.NewTool("get_time").
			SetDescription("Get current time").
			AddParameter("timezone", "string", "Timezone", false),
	}
}

func executeServerTool(name, args string) (string, error) {
	switch name {
	case "get_time":
		return fmt.Sprintf(`{"time": "%s"}`, time.Now().Format(time.RFC3339)), nil
	case "calculate":
		return `{"result": "calculation done"}`, nil
	default:
		return `{"error": "unknown tool"}`, nil
	}
}
```

## Configuration

```yaml
SERVER_PORT: ":8080"
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
API_KEY: "your-secret-api-key"

# Models
CHAT_MODEL: "ai/qwen2.5:1.5B-F16"
EMBEDDING_MODEL: "ai/mxbai-embed-large"
TOOLS_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
```

## Deployment

```bash
# Build
go build -o crew-server

# Run
./crew-server

# Docker
docker build -t crew-server .
docker run -p 8080:8080 crew-server
```

## Advanced: With Orchestrator for Auto-Routing

Add automatic topic detection and agent routing using the orchestrator agent:

```go
import (
    "github.com/snipwise/nova/nova-sdk/agents/crewserver"
    "github.com/snipwise/nova/nova-sdk/agents/orchestrator"
)

// Create orchestrator
orchestratorAgent, _ := orchestrator.NewAgent(
    ctx,
    agents.Config{
        Name:      "orchestrator",
        EngineURL: EngineURL,
        SystemInstructions: `
Identify query type: chat, search, tools, technical, general
Respond in JSON: {"topic_discussion": "QueryType"}`,
    },
    models.Config{
        Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
        Temperature: models.Float64(0.0),
    },
)

// Create crew server with specialized chat agents
chatAgents := map[string]*chat.Agent{
    "technical": technicalAgent,
    "general":   generalAgent,
    "support":   supportAgent,
}

// Define routing
matchAgentFn := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "technical", "tools":
        return "technical"
    case "support", "help":
        return "support"
    default:
        return "general"
    }
}

// Create crew server agent
crewServerAgent, _ := crewserver.NewAgent(
    ctx,
    chatAgents,
    "general",
    ":8080",
    matchAgentFn,
    executeToolFn,
)

// Attach orchestrator
crewServerAgent.SetOrchestratorAgent(orchestratorAgent)

// Start server - queries auto-route to appropriate agents
crewServerAgent.StartServer()
```

See `orchestrator/topic-detection` for more details.

## Important Notes

- Use HTTPS in production
- Implement proper authentication system
- Add metrics (Prometheus)
- Configure appropriate timeouts
- Implement graceful shutdown
- **Use orchestrator agent for intelligent request routing**
