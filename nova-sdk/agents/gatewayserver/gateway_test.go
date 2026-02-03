package gatewayserver_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/gatewayserver"
	"github.com/snipwise/nova/nova-sdk/models"
)

// --- Fake LLM Server ---

// newFakeLLMServer creates an httptest.Server that mimics an OpenAI-compatible LLM backend.
// It responds to POST /chat/completions with a simple response and GET /models with a model list.
func newFakeLLMServer(responseContent string) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /chat/completions", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Stream bool              `json:"stream"`
			Tools  []json.RawMessage `json:"tools"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		if len(req.Tools) > 0 {
			// Simulate a tool_calls response
			handleFakeToolCallResponse(w, r, req.Stream)
			return
		}

		if req.Stream {
			handleFakeStreamResponse(w, responseContent)
		} else {
			handleFakeNonStreamResponse(w, responseContent)
		}
	})

	mux.HandleFunc("GET /models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []map[string]any{
				{"id": "test-model", "object": "model", "created": time.Now().Unix(), "owned_by": "test"},
			},
		})
	})

	return httptest.NewServer(mux)
}

func handleFakeNonStreamResponse(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"id":      "chatcmpl-fake",
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   "test-model",
		"choices": []map[string]any{
			{
				"index":         0,
				"message":       map[string]any{"role": "assistant", "content": content},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]any{
			"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15,
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleFakeStreamResponse(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "no flusher", 500)
		return
	}

	// Chunk 1: role
	writeSSEChunk(w, flusher, map[string]any{
		"id": "chatcmpl-fake", "object": "chat.completion.chunk", "model": "test-model",
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{"role": "assistant"}, "finish_reason": nil},
		},
	})

	// Chunk 2: content (word by word)
	words := strings.Fields(content)
	for _, word := range words {
		writeSSEChunk(w, flusher, map[string]any{
			"id": "chatcmpl-fake", "object": "chat.completion.chunk", "model": "test-model",
			"choices": []map[string]any{
				{"index": 0, "delta": map[string]any{"content": word + " "}, "finish_reason": nil},
			},
		})
	}

	// Chunk 3: finish
	writeSSEChunk(w, flusher, map[string]any{
		"id": "chatcmpl-fake", "object": "chat.completion.chunk", "model": "test-model",
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{}, "finish_reason": "stop"},
		},
	})

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func handleFakeToolCallResponse(w http.ResponseWriter, r *http.Request, stream bool) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"id":      "chatcmpl-fake-tool",
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   "test-model",
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": nil,
					"tool_calls": []map[string]any{
						{
							"id":   "call_test123",
							"type": "function",
							"function": map[string]any{
								"name":      "get_weather",
								"arguments": `{"city":"Paris"}`,
							},
						},
					},
				},
				"finish_reason": "tool_calls",
			},
		},
		"usage": map[string]any{
			"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15,
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func writeSSEChunk(w http.ResponseWriter, flusher http.Flusher, data map[string]any) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	flusher.Flush()
}

// (helpers moved inline to each test)

// --- Tests ---

func TestNonStreamingCompletion_RequestParsing(t *testing.T) {
	// Verify that a well-formed OpenAI request can be parsed correctly
	reqBody := gatewayserver.ChatCompletionRequest{
		Model: "test-model",
		Messages: []gatewayserver.ChatCompletionMessage{
			{Role: "system", Content: gatewayserver.NewMessageContent("You are helpful.")},
			{Role: "user", Content: gatewayserver.NewMessageContent("What is 1+1?")},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var parsed gatewayserver.ChatCompletionRequest
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed.Model != "test-model" {
		t.Errorf("expected model test-model, got %s", parsed.Model)
	}
	if len(parsed.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(parsed.Messages))
	}
	if parsed.Stream {
		t.Error("expected stream=false")
	}
}

func TestTypes_ChatCompletionRequest(t *testing.T) {
	// Test that our request type correctly deserializes an OpenAI-compatible request
	input := `{
		"model": "gpt-4",
		"messages": [
			{"role": "system", "content": "You are helpful."},
			{"role": "user", "content": "Hello!"},
			{"role": "assistant", "content": "Hi there!", "tool_calls": [
				{"id": "call_1", "type": "function", "function": {"name": "get_time", "arguments": "{}"}}
			]},
			{"role": "tool", "tool_call_id": "call_1", "content": "{\"time\": \"12:00\"}"}
		],
		"stream": true,
		"temperature": 0.7,
		"tools": [
			{"type": "function", "function": {"name": "get_time", "description": "Get current time", "parameters": {"type": "object", "properties": {}}}}
		]
	}`

	var req gatewayserver.ChatCompletionRequest
	err := json.Unmarshal([]byte(input), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if req.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", req.Model)
	}
	if !req.Stream {
		t.Error("expected stream=true")
	}
	if len(req.Messages) != 4 {
		t.Errorf("expected 4 messages, got %d", len(req.Messages))
	}
	if req.Messages[0].Role != "system" {
		t.Errorf("expected role system, got %s", req.Messages[0].Role)
	}
	if req.Messages[1].Role != "user" && req.Messages[1].Content.String() != "Hello!" {
		t.Error("unexpected user message")
	}
	if len(req.Messages[2].ToolCalls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(req.Messages[2].ToolCalls))
	}
	if req.Messages[2].ToolCalls[0].ID != "call_1" {
		t.Errorf("expected tool call ID call_1, got %s", req.Messages[2].ToolCalls[0].ID)
	}
	if req.Messages[2].ToolCalls[0].Function.Name != "get_time" {
		t.Errorf("expected function name get_time, got %s", req.Messages[2].ToolCalls[0].Function.Name)
	}
	if req.Messages[3].Role != "tool" {
		t.Errorf("expected role tool, got %s", req.Messages[3].Role)
	}
	if req.Messages[3].ToolCallID != "call_1" {
		t.Errorf("expected tool_call_id call_1, got %s", req.Messages[3].ToolCallID)
	}
	if len(req.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(req.Tools))
	}
	if req.Tools[0].Function.Name != "get_time" {
		t.Errorf("expected tool name get_time, got %s", req.Tools[0].Function.Name)
	}
	if req.Temperature == nil || *req.Temperature != 0.7 {
		t.Error("expected temperature 0.7")
	}
}

func TestTypes_ChatCompletionResponse(t *testing.T) {
	// Test serialization of a non-streaming response
	content := "Hello!"
	finishReason := "stop"
	resp := gatewayserver.ChatCompletionResponse{
		ID:      "chatcmpl-test",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "test-model",
		Choices: []gatewayserver.ChatCompletionChoice{
			{
				Index: 0,
				Message: gatewayserver.ChatCompletionMessage{
					Role:    "assistant",
					Content: gatewayserver.NewMessageContent(content),
				},
				FinishReason: &finishReason,
			},
		},
		Usage: &gatewayserver.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify it round-trips correctly
	var parsed map[string]any
	_ = json.Unmarshal(jsonData, &parsed)

	if parsed["id"] != "chatcmpl-test" {
		t.Errorf("expected id chatcmpl-test, got %v", parsed["id"])
	}
	if parsed["object"] != "chat.completion" {
		t.Errorf("expected object chat.completion, got %v", parsed["object"])
	}

	choices := parsed["choices"].([]any)
	if len(choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(choices))
	}

	choice := choices[0].(map[string]any)
	msg := choice["message"].(map[string]any)
	if msg["role"] != "assistant" {
		t.Errorf("expected role assistant, got %v", msg["role"])
	}
	if msg["content"] != "Hello!" {
		t.Errorf("expected content Hello!, got %v", msg["content"])
	}
}

func TestTypes_StreamingChunk(t *testing.T) {
	// Test serialization of a streaming chunk with tool_calls
	finishReason := "tool_calls"
	idx := 0
	chunk := gatewayserver.ChatCompletionChunk{
		ID:      "chatcmpl-stream",
		Object:  "chat.completion.chunk",
		Created: 1234567890,
		Model:   "test-model",
		Choices: []gatewayserver.ChatCompletionChunkChoice{
			{
				Index: 0,
				Delta: gatewayserver.ChatCompletionDelta{
					ToolCalls: []gatewayserver.ToolCall{
						{
							Index: &idx,
							ID:    "call_abc",
							Type:  "function",
							Function: gatewayserver.FunctionCall{
								Name:      "search",
								Arguments: `{"query":"test"}`,
							},
						},
					},
				},
				FinishReason: &finishReason,
			},
		},
	}

	jsonData, err := json.Marshal(chunk)
	if err != nil {
		t.Fatalf("Failed to marshal chunk: %v", err)
	}

	// Parse it back as a generic map to verify structure
	var parsed map[string]any
	_ = json.Unmarshal(jsonData, &parsed)

	if parsed["object"] != "chat.completion.chunk" {
		t.Errorf("expected object chat.completion.chunk, got %v", parsed["object"])
	}

	choices := parsed["choices"].([]any)
	choice := choices[0].(map[string]any)
	if choice["finish_reason"] != "tool_calls" {
		t.Errorf("expected finish_reason tool_calls, got %v", choice["finish_reason"])
	}

	delta := choice["delta"].(map[string]any)
	toolCalls := delta["tool_calls"].([]any)
	if len(toolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(toolCalls))
	}

	tc := toolCalls[0].(map[string]any)
	if tc["id"] != "call_abc" {
		t.Errorf("expected tool call id call_abc, got %v", tc["id"])
	}
	fn := tc["function"].(map[string]any)
	if fn["name"] != "search" {
		t.Errorf("expected function name search, got %v", fn["name"])
	}
}

func TestTypes_ErrorResponse(t *testing.T) {
	apiErr := gatewayserver.APIError{
		Error: gatewayserver.APIErrorDetail{
			Message: "Something went wrong",
			Type:    "server_error",
		},
	}

	jsonData, err := json.Marshal(apiErr)
	if err != nil {
		t.Fatalf("Failed to marshal error: %v", err)
	}

	var parsed map[string]any
	_ = json.Unmarshal(jsonData, &parsed)

	errObj := parsed["error"].(map[string]any)
	if errObj["message"] != "Something went wrong" {
		t.Errorf("unexpected error message: %v", errObj["message"])
	}
	if errObj["type"] != "server_error" {
		t.Errorf("unexpected error type: %v", errObj["type"])
	}
}

func TestIntegration_NonStreamingCompletion(t *testing.T) {
	fakeLLM := newFakeLLMServer("Hello from the gateway!")
	defer fakeLLM.Close()

	ctx := context.Background()
	chatAgent, err := chat.NewAgent(ctx, agents.Config{
		Name: "test", EngineURL: fakeLLM.URL, SystemInstructions: "test", KeepConversationHistory: true,
	}, models.Config{Name: "test-model", Temperature: models.Float64(0.0)})
	if err != nil {
		t.Fatalf("Failed to create chat agent: %v", err)
	}

	gateway, err := gatewayserver.NewAgent(ctx,
		gatewayserver.WithSingleAgent(chatAgent),
		gatewayserver.WithPort(0),
	)
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	// Start a test HTTP server with the gateway's routes
	testMux := http.NewServeMux()
	testMux.HandleFunc("POST /v1/chat/completions", gateway.HandleChatCompletionsForTest)
	testMux.HandleFunc("GET /v1/models", gateway.HandleListModelsForTest)
	testMux.HandleFunc("GET /health", gateway.HandleHealthForTest)
	ts := httptest.NewServer(testMux)
	defer ts.Close()

	// Test non-streaming completion
	reqBody := `{"model":"test","messages":[{"role":"user","content":"Hello!"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var result gatewayserver.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Object != "chat.completion" {
		t.Errorf("expected object chat.completion, got %s", result.Object)
	}
	if len(result.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(result.Choices))
	}
	if result.Choices[0].Message.Role != "assistant" {
		t.Errorf("expected role assistant, got %s", result.Choices[0].Message.Role)
	}
	if result.Choices[0].Message.Content == nil || result.Choices[0].Message.Content.String() == "" {
		t.Error("expected non-empty content")
	}
	if result.Choices[0].FinishReason == nil || *result.Choices[0].FinishReason != "stop" {
		t.Errorf("expected finish_reason stop, got %v", result.Choices[0].FinishReason)
	}
}

func TestIntegration_StreamingCompletion(t *testing.T) {
	fakeLLM := newFakeLLMServer("Hello world")
	defer fakeLLM.Close()

	ctx := context.Background()
	chatAgent, err := chat.NewAgent(ctx, agents.Config{
		Name: "test", EngineURL: fakeLLM.URL, SystemInstructions: "test", KeepConversationHistory: true,
	}, models.Config{Name: "test-model", Temperature: models.Float64(0.0)})
	if err != nil {
		t.Fatalf("Failed to create chat agent: %v", err)
	}

	gateway, err := gatewayserver.NewAgent(ctx,
		gatewayserver.WithSingleAgent(chatAgent),
		gatewayserver.WithPort(0),
	)
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	testMux := http.NewServeMux()
	testMux.HandleFunc("POST /v1/chat/completions", gateway.HandleChatCompletionsForTest)
	ts := httptest.NewServer(testMux)
	defer ts.Close()

	reqBody := `{"model":"test","messages":[{"role":"user","content":"Hello!"}],"stream":true}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("expected Content-Type text/event-stream, got %s", ct)
	}

	// Parse SSE stream
	scanner := bufio.NewScanner(resp.Body)
	var chunks []gatewayserver.ChatCompletionChunk
	gotDone := false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if line == "data: [DONE]" {
			gotDone = true
			break
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		jsonStr := strings.TrimPrefix(line, "data: ")
		var chunk gatewayserver.ChatCompletionChunk
		if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
			t.Logf("Skipping unparseable chunk: %s", jsonStr)
			continue
		}
		chunks = append(chunks, chunk)
	}

	if !gotDone {
		t.Error("expected [DONE] marker in stream")
	}
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	// First chunk should have role
	if chunks[0].Choices[0].Delta.Role != "assistant" {
		t.Errorf("expected first chunk delta.role=assistant, got %s", chunks[0].Choices[0].Delta.Role)
	}

	// Verify object type
	for _, chunk := range chunks {
		if chunk.Object != "chat.completion.chunk" {
			t.Errorf("expected object chat.completion.chunk, got %s", chunk.Object)
		}
	}
}

func TestIntegration_ModelsEndpoint(t *testing.T) {
	fakeLLM := newFakeLLMServer("test")
	defer fakeLLM.Close()

	ctx := context.Background()
	chatAgent, err := chat.NewAgent(ctx, agents.Config{
		Name: "test", EngineURL: fakeLLM.URL, SystemInstructions: "test", KeepConversationHistory: true,
	}, models.Config{Name: "test-model", Temperature: models.Float64(0.0)})
	if err != nil {
		t.Fatalf("Failed to create chat agent: %v", err)
	}

	gateway, err := gatewayserver.NewAgent(ctx,
		gatewayserver.WithSingleAgent(chatAgent),
		gatewayserver.WithPort(0),
	)
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	testMux := http.NewServeMux()
	testMux.HandleFunc("GET /v1/models", gateway.HandleListModelsForTest)
	ts := httptest.NewServer(testMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/models")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	var result gatewayserver.ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if result.Object != "list" {
		t.Errorf("expected object list, got %s", result.Object)
	}
	if len(result.Data) == 0 {
		t.Error("expected at least one model")
	}
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	fakeLLM := newFakeLLMServer("test")
	defer fakeLLM.Close()

	ctx := context.Background()
	chatAgent, err := chat.NewAgent(ctx, agents.Config{
		Name: "test", EngineURL: fakeLLM.URL, SystemInstructions: "test", KeepConversationHistory: true,
	}, models.Config{Name: "test-model", Temperature: models.Float64(0.0)})
	if err != nil {
		t.Fatalf("Failed to create chat agent: %v", err)
	}

	gateway, err := gatewayserver.NewAgent(ctx,
		gatewayserver.WithSingleAgent(chatAgent),
	)
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	testMux := http.NewServeMux()
	testMux.HandleFunc("GET /health", gateway.HandleHealthForTest)
	ts := httptest.NewServer(testMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %s", result["status"])
	}
}

// --- helpers ---

func strPtr(s string) *string {
	return &s
}
