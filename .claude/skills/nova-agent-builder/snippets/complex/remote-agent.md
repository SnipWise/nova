---
id: remote-agent
name: Remote Agent (Crew Server Client)
category: complex
complexity: intermediate
sample_source: 51
description: Client agent that connects to a remote Crew Server to use its capabilities
interactive: true
---

# Remote Agent (Crew Server Client)

## Description

Creates a client agent that connects to a remote Crew Server via HTTP/WebSocket. Allows using a centralized agent server's capabilities from any application.

## Use Cases

- Client applications using remote agents
- Microservices consuming an agent API
- Mobile/web apps with centralized AI backend
- Integrating agents into existing systems
- Scaling clients without duplicating models

## ⚠️ Interactive Mode

This snippet requires specific information. Answer the following questions:

### Configuration Questions

1. **Crew Server URL?**
   - e.g., `http://localhost:8080` or `https://agents.example.com`

2. **Which endpoints to use?**
   - Chat, RAG, Tools, or all?

3. **Connection mode?**
   - HTTP (request/response)
   - WebSocket (streaming)
   - Both

4. **Error handling?**
   - Automatic retry
   - Local fallback
   - Circuit breaker

5. **Response caching?**
   - Yes/No
   - Cache TTL

---

## Base Template

```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// === CONFIGURATION ===
type RemoteAgentConfig struct {
	ServerURL  string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

// === REMOTE AGENT CLIENT ===
type RemoteAgent struct {
	config     RemoteAgentConfig
	httpClient *http.Client
}

// === REQUEST/RESPONSE STRUCTURES ===
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
	Error        string `json:"error,omitempty"`
}

type RAGSearchRequest struct {
	Query     string  `json:"query"`
	Threshold float64 `json:"threshold,omitempty"`
	TopK      int     `json:"top_k,omitempty"`
}

type RAGSearchResponse struct {
	Results []RAGResult `json:"results"`
	Query   string      `json:"query"`
	Error   string      `json:"error,omitempty"`
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
	Error        string   `json:"error,omitempty"`
}

type HealthResponse struct {
	Status       string   `json:"status"`
	Agents       []string `json:"agents"`
	RequestCount int64    `json:"request_count"`
	Uptime       string   `json:"uptime"`
}

// === CONSTRUCTOR ===
func NewRemoteAgent(config RemoteAgentConfig) *RemoteAgent {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	return &RemoteAgent{
		config:     config,
		httpClient: &http.Client{Timeout: config.Timeout},
	}
}

// === MAIN METHODS ===

// Chat sends a message and receives a response
func (ra *RemoteAgent) Chat(ctx context.Context, message string) (*ChatResponse, error) {
	req := ChatRequest{
		Message:   message,
		SessionID: fmt.Sprintf("session-%d", time.Now().UnixNano()),
	}

	var resp ChatResponse
	err := ra.doRequest(ctx, "POST", "/api/chat", req, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("server error: %s", resp.Error)
	}

	return &resp, nil
}

// RAGSearch performs a semantic search
func (ra *RemoteAgent) RAGSearch(ctx context.Context, query string, threshold float64) (*RAGSearchResponse, error) {
	req := RAGSearchRequest{Query: query, Threshold: threshold}

	var resp RAGSearchResponse
	err := ra.doRequest(ctx, "POST", "/api/rag/search", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// RAGIndex indexes documents
func (ra *RemoteAgent) RAGIndex(ctx context.Context, documents []string) (int, error) {
	req := struct {
		Documents []string `json:"documents"`
	}{Documents: documents}

	var resp struct {
		Indexed int `json:"indexed"`
		Total   int `json:"total"`
	}
	err := ra.doRequest(ctx, "POST", "/api/rag/index", req, &resp)
	if err != nil {
		return 0, err
	}

	return resp.Indexed, nil
}

// Tools executes a request with tools
func (ra *RemoteAgent) Tools(ctx context.Context, message string) (*ToolsResponse, error) {
	req := ToolsRequest{Message: message}

	var resp ToolsResponse
	err := ra.doRequest(ctx, "POST", "/api/tools", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Health checks server status
func (ra *RemoteAgent) Health(ctx context.Context) (*HealthResponse, error) {
	var resp HealthResponse
	err := ra.doRequestNoAuth(ctx, "GET", "/health", nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// IsAvailable checks if server is available
func (ra *RemoteAgent) IsAvailable(ctx context.Context) bool {
	health, err := ra.Health(ctx)
	return err == nil && health.Status == "healthy"
}

// === INTERNAL METHODS ===

func (ra *RemoteAgent) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	return ra.doRequestWithRetry(ctx, method, path, body, result, true)
}

func (ra *RemoteAgent) doRequestNoAuth(ctx context.Context, method, path string, body, result interface{}) error {
	return ra.doRequestWithRetry(ctx, method, path, body, result, false)
}

func (ra *RemoteAgent) doRequestWithRetry(ctx context.Context, method, path string, body, result interface{}, auth bool) error {
	var lastErr error

	for attempt := 0; attempt <= ra.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(ra.config.RetryDelay * time.Duration(attempt)):
			}
			fmt.Printf("⚠️ Retry %d/%d\n", attempt, ra.config.MaxRetries)
		}

		err := ra.doSingleRequest(ctx, method, path, body, result, auth)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed after %d attempts: %v", ra.config.MaxRetries+1, lastErr)
}

func (ra *RemoteAgent) doSingleRequest(ctx context.Context, method, path string, body, result interface{}, auth bool) error {
	url := ra.config.ServerURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal error: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("request creation error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if auth {
		req.Header.Set("X-API-Key", ra.config.APIKey)
	}

	resp, err := ra.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode error: %v", err)
		}
	}

	return nil
}

// === MAIN - USAGE EXAMPLE ===
func main() {
	ctx := context.Background()

	// Create client
	agent := NewRemoteAgent(RemoteAgentConfig{
		ServerURL:  "http://localhost:8080",
		APIKey:     "your-secret-api-key",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	})

	// Check availability
	fmt.Println("🔍 Checking server...")
	if !agent.IsAvailable(ctx) {
		fmt.Println("❌ Server unavailable")
		return
	}
	fmt.Println("✅ Server available")

	// Test Chat
	fmt.Println("\n=== Chat Test ===")
	chatResp, err := agent.Chat(ctx, "Hello, how are you?")
	if err != nil {
		fmt.Printf("❌ Chat error: %v\n", err)
	} else {
		fmt.Printf("🤖 Response: %s\n", chatResp.Response)
	}

	// Test RAG
	fmt.Println("\n=== RAG Test ===")
	docs := []string{
		"Go is a programming language created by Google",
		"Python is popular for machine learning",
	}
	indexed, err := agent.RAGIndex(ctx, docs)
	if err != nil {
		fmt.Printf("❌ Index error: %v\n", err)
	} else {
		fmt.Printf("📚 Documents indexed: %d\n", indexed)
	}

	// Test Tools
	fmt.Println("\n=== Tools Test ===")
	toolsResp, err := agent.Tools(ctx, "What time is it?")
	if err != nil {
		fmt.Printf("❌ Tools error: %v\n", err)
	} else {
		fmt.Printf("🔧 Result: %s\n", toolsResp.FinalMessage)
	}

	fmt.Println("\n✅ Tests completed")
}
```

## Configuration

```yaml
CREW_SERVER_URL: "http://localhost:8080"
API_KEY: "your-secret-api-key"
TIMEOUT: "30s"
MAX_RETRIES: 3
RETRY_DELAY: "1s"
```

## Customization

### With Local Cache

```go
import "github.com/patrickmn/go-cache"

type CachedRemoteAgent struct {
    *RemoteAgent
    cache *cache.Cache
}

func (cra *CachedRemoteAgent) Chat(ctx context.Context, message string) (*ChatResponse, error) {
    cacheKey := fmt.Sprintf("chat:%s", message)
    if cached, found := cra.cache.Get(cacheKey); found {
        return cached.(*ChatResponse), nil
    }
    
    resp, err := cra.RemoteAgent.Chat(ctx, message)
    if err != nil {
        return nil, err
    }
    
    cra.cache.Set(cacheKey, resp, cache.DefaultExpiration)
    return resp, nil
}
```

### With Circuit Breaker

```go
import "github.com/sony/gobreaker"

type ResilientRemoteAgent struct {
    *RemoteAgent
    breaker *gobreaker.CircuitBreaker
}

func (rra *ResilientRemoteAgent) Chat(ctx context.Context, message string) (*ChatResponse, error) {
    result, err := rra.breaker.Execute(func() (interface{}, error) {
        return rra.RemoteAgent.Chat(ctx, message)
    })
    
    if err != nil {
        return nil, err
    }
    
    return result.(*ChatResponse), nil
}
```

## Important Notes

- Always check availability with `IsAvailable()` before use
- Implement appropriate timeout for your use case
- Use circuit breaker for critical systems
- Cache reduces load but watch data freshness
- Consider local fallback for high availability
