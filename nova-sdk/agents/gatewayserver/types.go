package gatewayserver

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OpenAI-compatible Chat Completions API types.
// These types mirror the OpenAI API specification to ensure compatibility
// with clients like qwen-code, aider, continue.dev, etc.

// MessageContent represents the content field of a message.
// It supports both simple string format and array format (for multi-modal content).
// OpenAI API allows content to be either:
//   - A simple string: "Hello world"
//   - An array of strings: ["Hello", "world"]
//   - An array of content parts: [{"type": "text", "text": "Hello"}]
type MessageContent struct {
	text string
}

// UnmarshalJSON implements custom unmarshaling to support both string and array formats.
func (mc *MessageContent) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first (most common case)
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		mc.text = str
		return nil
	}

	// Try to unmarshal as an array of strings
	var strArray []string
	if err := json.Unmarshal(data, &strArray); err == nil {
		// Concatenate all strings in the array
		mc.text = strings.Join(strArray, " ")
		return nil
	}

	// Try to unmarshal as an array of content parts (multi-modal format)
	var parts []map[string]any
	if err := json.Unmarshal(data, &parts); err == nil {
		var texts []string
		for _, part := range parts {
			if text, ok := part["text"].(string); ok {
				texts = append(texts, text)
			}
		}
		mc.text = strings.Join(texts, " ")
		return nil
	}

	return fmt.Errorf("content must be a string or an array")
}

// MarshalJSON implements JSON marshaling, always outputting as a simple string.
func (mc MessageContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(mc.text)
}

// String returns the text content as a string.
func (mc *MessageContent) String() string {
	if mc == nil {
		return ""
	}
	return mc.text
}

// IsEmpty returns true if the content is empty or nil.
func (mc *MessageContent) IsEmpty() bool {
	return mc == nil || mc.text == ""
}

// NewMessageContent creates a new MessageContent from a string.
func NewMessageContent(text string) *MessageContent {
	return &MessageContent{text: text}
}

// --- Request types ---

// ChatCompletionRequest represents a POST /v1/chat/completions request.
type ChatCompletionRequest struct {
	Model            string                  `json:"model"`
	Messages         []ChatCompletionMessage `json:"messages"`
	Stream           bool                    `json:"stream,omitempty"`
	Temperature      *float64                `json:"temperature,omitempty"`
	TopP             *float64                `json:"top_p,omitempty"`
	MaxTokens        *int64                  `json:"max_tokens,omitempty"`
	Stop             []string                `json:"stop,omitempty"`
	Tools            []ToolDefinition        `json:"tools,omitempty"`
	ToolChoice       any                     `json:"tool_choice,omitempty"`
	FrequencyPenalty *float64                `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64                `json:"presence_penalty,omitempty"`
	N                *int                    `json:"n,omitempty"`
}

// ChatCompletionMessage represents a message in the conversation.
type ChatCompletionMessage struct {
	Role       string          `json:"role"`
	Content    *MessageContent `json:"content"`
	Name       string          `json:"name,omitempty"`
	ToolCalls  []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

// ToolDefinition represents a tool available for the model to call.
type ToolDefinition struct {
	Type     string           `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition describes a function the model can call.
type FunctionDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  any         `json:"parameters,omitempty"`
}

// ToolCall represents a tool call made by the assistant.
type ToolCall struct {
	Index    *int             `json:"index,omitempty"`
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function FunctionCall     `json:"function"`
}

// FunctionCall contains the function name and arguments.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// --- Non-streaming response types ---

// ChatCompletionResponse represents a complete (non-streaming) response.
type ChatCompletionResponse struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []ChatCompletionChoice   `json:"choices"`
	Usage   *Usage                   `json:"usage,omitempty"`
}

// ChatCompletionChoice represents one choice in a non-streaming response.
type ChatCompletionChoice struct {
	Index        int                    `json:"index"`
	Message      ChatCompletionMessage  `json:"message"`
	FinishReason *string                `json:"finish_reason"`
}

// Usage reports token usage for the request.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// --- Streaming response types ---

// ChatCompletionChunk represents a single SSE chunk in streaming mode.
type ChatCompletionChunk struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionChunkChoice  `json:"choices"`
	Usage   *Usage                       `json:"usage,omitempty"`
}

// ChatCompletionChunkChoice represents one choice in a streaming chunk.
type ChatCompletionChunkChoice struct {
	Index        int                   `json:"index"`
	Delta        ChatCompletionDelta   `json:"delta"`
	FinishReason *string               `json:"finish_reason"`
}

// ChatCompletionDelta represents the incremental content in a streaming chunk.
type ChatCompletionDelta struct {
	Role      string          `json:"role,omitempty"`
	Content   *MessageContent `json:"content,omitempty"`
	ToolCalls []ToolCall      `json:"tool_calls,omitempty"`
}

// --- Models endpoint types ---

// ModelsResponse represents the GET /v1/models response.
type ModelsResponse struct {
	Object string       `json:"object"`
	Data   []ModelEntry `json:"data"`
}

// ModelEntry represents a single model in the models list.
type ModelEntry struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// --- Error types ---

// APIError represents an OpenAI-compatible error response.
type APIError struct {
	Error APIErrorDetail `json:"error"`
}

// APIErrorDetail contains the error details.
type APIErrorDetail struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Code    *string `json:"code"`
}

// --- Agent Execution Order types ---

// AgentExecutionType defines the type of agent processing step
type AgentExecutionType string

const (
	// AgentExecutionClientSideTools processes requests with client-side tool execution
	// The gateway detects tool calls and returns them to the client for execution
	AgentExecutionClientSideTools AgentExecutionType = "client_side_tools"

	// AgentExecutionServerSideTools processes requests with server-side tool execution
	// The gateway executes tools internally and continues the completion loop
	AgentExecutionServerSideTools AgentExecutionType = "server_side_tools"

	// AgentExecutionOrchestrator processes requests through the orchestrator
	// The orchestrator detects topics and routes to appropriate agents
	AgentExecutionOrchestrator AgentExecutionType = "orchestrator"
)

// DefaultAgentExecutionOrder defines the default order of agent processing
var DefaultAgentExecutionOrder = []AgentExecutionType{
	AgentExecutionClientSideTools,
	AgentExecutionServerSideTools,
	AgentExecutionOrchestrator,
}
