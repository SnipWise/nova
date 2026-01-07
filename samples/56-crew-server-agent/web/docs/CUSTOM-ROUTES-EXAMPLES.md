# Custom Routes Examples

The SDK now exposes the HTTP multiplexer via `agent.Mux`, allowing easy addition of custom routes.

## Basic Usage

```go
// Create the agent
crewAgent, err := crewserver.NewAgent(
    ctx,
    crewserver.WithSingleAgent(chatAgent),
    crewserver.WithPort(8080),
)

// Add custom routes BEFORE starting the server
crewAgent.Mux.HandleFunc("GET /custom/endpoint", myHandler)

// Start the server (CORS will be applied automatically)
crewAgent.StartServer()
```

## ⚠️ Important

- Custom routes must be added **BEFORE** `StartServer()`
- CORS middleware is applied automatically to all routes
- Use standard HTTP methods: GET, POST, PUT, DELETE
- Go 1.22+ supports methods in patterns: `"POST /endpoint"`

## Practical Examples

### 1. Server Statistics

```go
package main

import (
    "encoding/json"
    "net/http"
    "sync/atomic"
    "time"
)

var (
    startTime     = time.Now()
    requestCount  int64
    completionCount int64
)

func main() {
    // ... agent creation ...

    // Stats route
    crewAgent.Mux.HandleFunc("GET /stats", func(w http.ResponseWriter, r *http.Request) {
        stats := map[string]interface{}{
            "uptime_seconds":     time.Since(startTime).Seconds(),
            "total_requests":     atomic.LoadInt64(&requestCount),
            "total_completions":  atomic.LoadInt64(&completionCount),
            "current_agent":      crewAgent.GetSelectedAgentId(),
            "context_size":       crewAgent.GetContextSize(),
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(stats)
    })

    crewAgent.StartServer()
}
```

**Test:**
```bash
curl http://localhost:8080/stats
```

**Response:**
```json
{
  "uptime_seconds": 123.45,
  "total_requests": 42,
  "total_completions": 15,
  "current_agent": "coder",
  "context_size": 2456
}
```

### 2. Dynamic Agent Switch

```go
crewAgent.Mux.HandleFunc("POST /agent/switch", func(w http.ResponseWriter, r *http.Request) {
    var req struct {
        AgentID string `json:"agent_id"`
    }

    // Parse request body
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Switch to the requested agent
    if err := crewAgent.SetSelectedAgentId(req.AgentID); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Return success
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":   "ok",
        "agent_id": req.AgentID,
    })
})
```

**Test:**
```bash
curl -X POST http://localhost:8080/agent/switch \
  -H "Content-Type: application/json" \
  -d '{"agent_id": "thinker"}'
```

**Response:**
```json
{
  "status": "ok",
  "agent_id": "thinker"
}
```

### 3. List Available Agents

```go
crewAgent.Mux.HandleFunc("GET /agents", func(w http.ResponseWriter, r *http.Request) {
    agents := crewAgent.GetChatAgents()

    agentList := []map[string]string{}
    for id, agent := range agents {
        agentList = append(agentList, map[string]string{
            "id":       id,
            "name":     agent.GetName(),
            "model_id": agent.GetModelID(),
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "current_agent": crewAgent.GetSelectedAgentId(),
        "agents":        agentList,
    })
})
```

**Test:**
```bash
curl http://localhost:8080/agents
```

**Response:**
```json
{
  "current_agent": "generic",
  "agents": [
    {
      "id": "coder",
      "name": "coder",
      "model_id": "hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m"
    },
    {
      "id": "thinker",
      "name": "thinker",
      "model_id": "hf.co/menlo/lucy-gguf:q4_k_m"
    },
    {
      "id": "cook",
      "name": "cook",
      "model_id": "ai/qwen2.5:1.5B-F16"
    },
    {
      "id": "generic",
      "name": "generic",
      "model_id": "hf.co/menlo/jan-nano-gguf:q4_k_m"
    }
  ]
}
```

### 4. Export Conversation (JSON)

```go
crewAgent.Mux.HandleFunc("GET /export/conversation", func(w http.ResponseWriter, r *http.Request) {
    messagesJSON, err := crewAgent.ExportMessagesToJSON()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Set headers for file download
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition", "attachment; filename=conversation.json")
    w.Write([]byte(messagesJSON))
})
```

**Test:**
```bash
curl http://localhost:8080/export/conversation > conversation.json
```

### 5. Detailed Health Check

```go
crewAgent.Mux.HandleFunc("GET /health/detailed", func(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status":         "healthy",
        "timestamp":      time.Now().Format(time.RFC3339),
        "uptime_seconds": time.Since(startTime).Seconds(),
        "agent": map[string]interface{}{
            "current_id":   crewAgent.GetSelectedAgentId(),
            "context_size": crewAgent.GetContextSize(),
        },
        "components": map[string]bool{
            "tools_agent":      crewAgent.ToolsAgent != nil,
            "rag_agent":        crewAgent.RagAgent != nil,
            "compressor_agent": crewAgent.CompressorAgent != nil,
        },
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
})
```

**Test:**
```bash
curl http://localhost:8080/health/detailed
```

### 6. Add New Agent to Crew

```go
crewAgent.Mux.HandleFunc("POST /agents/add", func(w http.ResponseWriter, r *http.Request) {
    var req struct {
        AgentID            string `json:"agent_id"`
        ModelID            string `json:"model_id"`
        SystemInstructions string `json:"system_instructions"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Create new agent
    newAgent, err := chat.NewAgent(
        r.Context(),
        agents.Config{
            Name:                    req.AgentID,
            EngineURL:               "http://localhost:12434/engines/llama.cpp/v1",
            SystemInstructions:      req.SystemInstructions,
            KeepConversationHistory: true,
        },
        models.Config{
            Name:        req.ModelID,
            Temperature: models.Float64(0.8),
        },
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Add to crew
    if err := crewAgent.AddChatAgentToCrew(req.AgentID, newAgent); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":   "ok",
        "agent_id": req.AgentID,
    })
})
```

**Test:**
```bash
curl -X POST http://localhost:8080/agents/add \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "translator",
    "model_id": "ai/qwen2.5:1.5B-F16",
    "system_instructions": "You are a professional translator."
  }'
```

### 7. Upload Documents for RAG

```go
import (
    "io"
    "github.com/snipwise/nova/nova-sdk/agents/rag/chunks"
)

crewAgent.Mux.HandleFunc("POST /rag/upload", func(w http.ResponseWriter, r *http.Request) {
    if crewAgent.RagAgent == nil {
        http.Error(w, "RAG agent not configured", http.StatusBadRequest)
        return
    }

    // Parse multipart form
    if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    file, _, err := r.FormFile("document")
    if err != nil {
        http.Error(w, "No file uploaded", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Read file content
    content, err := io.ReadAll(file)
    if err != nil {
        http.Error(w, "Failed to read file", http.StatusInternalServerError)
        return
    }

    // Split into chunks
    piecesOfDoc := chunks.SplitMarkdownBySections(string(content))

    // Save embeddings
    for _, piece := range piecesOfDoc {
        if err := crewAgent.RagAgent.SaveEmbedding(piece); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":      "ok",
        "chunks_added": len(piecesOfDoc),
    })
})
```

**Test:**
```bash
curl -X POST http://localhost:8080/rag/upload \
  -F "document=@README.md"
```

### 8. Direct RAG Search

```go
crewAgent.Mux.HandleFunc("POST /rag/search", func(w http.ResponseWriter, r *http.Request) {
    if crewAgent.RagAgent == nil {
        http.Error(w, "RAG agent not configured", http.StatusBadRequest)
        return
    }

    var req struct {
        Query           string  `json:"query"`
        SimilarityLimit float64 `json:"similarity_limit"`
        MaxResults      int     `json:"max_results"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Set defaults
    if req.SimilarityLimit == 0 {
        req.SimilarityLimit = 0.4
    }
    if req.MaxResults == 0 {
        req.MaxResults = 5
    }

    // Search embeddings
    results, err := crewAgent.RagAgent.GetSimilarities(
        req.Query,
        req.SimilarityLimit,
        req.MaxResults,
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "query":   req.Query,
        "results": results,
    })
})
```

**Test:**
```bash
curl -X POST http://localhost:8080/rag/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to install Nova SDK?",
    "similarity_limit": 0.4,
    "max_results": 5
  }'
```

## Complete Example in main.go

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"
    "sync/atomic"
    "time"

    "github.com/snipwise/nova/nova-sdk/agents/chat"
    "github.com/snipwise/nova/nova-sdk/agents/crewserver"
)

var (
    startTime    = time.Now()
    requestCount int64
)

func main() {
    ctx := context.Background()
    engineURL := "http://localhost:12434/engines/llama.cpp/v1"

    // Create chat agent
    chatAgent, err := chat.NewAgent(ctx, /* config */)
    if err != nil {
        panic(err)
    }

    // Create crew server agent
    crewAgent, err := crewserver.NewAgent(
        ctx,
        crewserver.WithSingleAgent(chatAgent),
        crewserver.WithPort(8080),
    )
    if err != nil {
        panic(err)
    }

    // Add custom routes
    addCustomRoutes(crewAgent)

    // Start server
    if err := crewAgent.StartServer(); err != nil {
        panic(err)
    }
}

func addCustomRoutes(crewAgent *crewserver.CrewServerAgent) {
    // Stats endpoint
    crewAgent.Mux.HandleFunc("GET /stats", func(w http.ResponseWriter, r *http.Request) {
        atomic.AddInt64(&requestCount, 1)

        stats := map[string]interface{}{
            "uptime_seconds": time.Since(startTime).Seconds(),
            "total_requests": atomic.LoadInt64(&requestCount),
            "current_agent":  crewAgent.GetSelectedAgentId(),
            "context_size":   crewAgent.GetContextSize(),
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(stats)
    })

    // Agent list
    crewAgent.Mux.HandleFunc("GET /agents", func(w http.ResponseWriter, r *http.Request) {
        agents := crewAgent.GetChatAgents()
        agentList := []map[string]string{}

        for id, agent := range agents {
            agentList = append(agentList, map[string]string{
                "id":       id,
                "name":     agent.GetName(),
                "model_id": agent.GetModelID(),
            })
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "current_agent": crewAgent.GetSelectedAgentId(),
            "agents":        agentList,
        })
    })

    // Agent switch
    crewAgent.Mux.HandleFunc("POST /agent/switch", func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            AgentID string `json:"agent_id"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        if err := crewAgent.SetSelectedAgentId(req.AgentID); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "ok",
            "agent_id": req.AgentID,
        })
    })

    // Export conversation
    crewAgent.Mux.HandleFunc("GET /export/conversation", func(w http.ResponseWriter, r *http.Request) {
        messagesJSON, err := crewAgent.ExportMessagesToJSON()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Content-Disposition", "attachment; filename=conversation.json")
        w.Write([]byte(messagesJSON))
    })
}
```

## Important Notes

### Automatic CORS

All custom routes automatically benefit from CORS middleware. You don't need to add CORS headers manually.

### Go 1.22+ Syntax

If you're using Go 1.22+, you can specify the HTTP method in the pattern:

```go
// ✅ Go 1.22+
crewAgent.Mux.HandleFunc("GET /stats", handler)
crewAgent.Mux.HandleFunc("POST /agent/switch", handler)

// ✅ Go < 1.22 (compatible)
crewAgent.Mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    // handler logic
})
```

### Route Order

Routes must be added **before** `StartServer()`:

```go
// ✅ Correct
crewAgent.Mux.HandleFunc("GET /stats", handler)
crewAgent.StartServer()

// ❌ Incorrect - too late!
crewAgent.StartServer()
crewAgent.Mux.HandleFunc("GET /stats", handler)
```

### Testing

You can test your routes with curl, Postman, or directly from the browser.

**Example of automated tests:**

```bash
#!/bin/bash

# Test stats
echo "Testing /stats..."
curl -s http://localhost:8080/stats | jq

# Test agents list
echo "Testing /agents..."
curl -s http://localhost:8080/agents | jq

# Test agent switch
echo "Testing /agent/switch..."
curl -s -X POST http://localhost:8080/agent/switch \
  -H "Content-Type: application/json" \
  -d '{"agent_id": "thinker"}' | jq

echo "All tests completed!"
```

## Conclusion

The Nova SDK now allows easy server extension with custom routes while automatically benefiting from CORS middleware. This opens up many possibilities for creating rich APIs around your agents!
