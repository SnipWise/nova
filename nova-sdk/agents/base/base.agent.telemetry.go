package base

import (
	"encoding/json"
	"time"

	"github.com/openai/openai-go/v3"
)

// RequestMetadata contains metadata about the last request sent to the LLM
type RequestMetadata struct {
	ContextLength int       `json:"context_length"`
	Model         string    `json:"model"`
	Stream        bool      `json:"stream"`
	Temperature   float64   `json:"temperature"`
	MaxTokens     int64     `json:"max_tokens"`
	TopP          float64   `json:"top_p"`
	Timestamp     time.Time `json:"timestamp"`
}

// ResponseMetadata contains metadata about the last response received from the LLM
type ResponseMetadata struct {
	ID               string    `json:"id"`
	Created          int64     `json:"created"`
	Model            string    `json:"model"`
	FinishReason     string    `json:"finish_reason"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	ResponseTime     int64     `json:"response_time_ms"` // en millisecondes
	Timestamp        time.Time `json:"timestamp"`
}

// TelemetryCallback is an interface for receiving telemetry events
type TelemetryCallback interface {
	OnRequestSent(metadata RequestMetadata, requestJSON string)
	OnResponseReceived(metadata ResponseMetadata, responseJSON string)
	OnStreamChunk(chunk string, index int)
	OnError(err error, context string)
}

// GetLastRequestJSON returns the last request sent to the LLM as JSON
func (agent *Agent) GetLastRequestJSON() (string, error) {
	if agent.lastRequest == nil {
		return "", nil
	}

	jsonBytes, err := json.MarshalIndent(agent.lastRequest, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetLastRequestContextLength returns the context length of the last request
func (agent *Agent) GetLastRequestContextLength() int {
	if agent.lastRequest == nil {
		return 0
	}

	// Calculate the total size of all messages
	contextSize := 0
	for _, msg := range agent.lastRequest.Messages {
		// Convert to JSON to estimate size
		jsonBytes, err := json.Marshal(msg)
		if err == nil {
			contextSize += len(jsonBytes)
		}
	}

	return contextSize
}

// GetLastRequestMetadata returns metadata about the last request
func (agent *Agent) GetLastRequestMetadata() RequestMetadata {
	if agent.lastRequest == nil {
		return RequestMetadata{}
	}

	modelName := ""
	if modelJSON, err := json.Marshal(agent.lastRequest.Model); err == nil {
		modelName = string(modelJSON)
	}

	metadata := RequestMetadata{
		ContextLength: agent.GetLastRequestContextLength(),
		Model:         modelName,
		Stream:        false, // Will be updated below if stream is used
		Timestamp:     agent.lastRequestTime,
	}

	// Extract values from the request params JSON
	// Since the OpenAI SDK uses Opt types, we parse them via JSON
	if reqBytes, err := json.Marshal(agent.lastRequest); err == nil {
		var rawParams map[string]interface{}
		if err := json.Unmarshal(reqBytes, &rawParams); err == nil {
			if temp, ok := rawParams["temperature"].(float64); ok {
				metadata.Temperature = temp
			}
			if maxTokens, ok := rawParams["max_tokens"].(float64); ok {
				metadata.MaxTokens = int64(maxTokens)
			}
			if topP, ok := rawParams["top_p"].(float64); ok {
				metadata.TopP = topP
			}
		}
	}

	return metadata
}

// GetLastResponseJSON returns the last response received from the LLM as JSON
func (agent *Agent) GetLastResponseJSON() (string, error) {
	if agent.lastResponse == nil {
		return "", nil
	}

	jsonBytes, err := json.MarshalIndent(agent.lastResponse, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetLastResponseMetadata returns metadata about the last response
func (agent *Agent) GetLastResponseMetadata() ResponseMetadata {
	if agent.lastResponse == nil {
		return ResponseMetadata{}
	}

	metadata := ResponseMetadata{
		ID:           agent.lastResponse.ID,
		Created:      agent.lastResponse.Created,
		Model:        agent.lastResponse.Model,
		ResponseTime: agent.lastResponseDuration.Milliseconds(),
		Timestamp:    agent.lastResponseTime,
	}

	// Extract finish reason from first choice if available
	if len(agent.lastResponse.Choices) > 0 {
		metadata.FinishReason = agent.lastResponse.Choices[0].FinishReason
	}

	// Extract token usage if available
	if agent.lastResponse.Usage.PromptTokens > 0 {
		metadata.PromptTokens = int(agent.lastResponse.Usage.PromptTokens)
	}
	if agent.lastResponse.Usage.CompletionTokens > 0 {
		metadata.CompletionTokens = int(agent.lastResponse.Usage.CompletionTokens)
	}
	if agent.lastResponse.Usage.TotalTokens > 0 {
		metadata.TotalTokens = int(agent.lastResponse.Usage.TotalTokens)
	}

	return metadata
}

// GetConversationHistoryJSON returns the entire conversation history as JSON
func (agent *Agent) GetConversationHistoryJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(agent.ChatCompletionParams.Messages, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetTotalTokensUsed returns the total number of tokens used since the agent was created
func (agent *Agent) GetTotalTokensUsed() int {
	return agent.totalTokensUsed
}

// ResetTelemetry resets all telemetry counters and stored data
func (agent *Agent) ResetTelemetry() {
	agent.lastRequest = nil
	agent.lastRequestTime = time.Time{}
	agent.lastResponse = nil
	agent.lastResponseTime = time.Time{}
	agent.lastResponseDuration = 0
	agent.totalTokensUsed = 0
}

// SetTelemetryCallback sets a callback for receiving telemetry events in real-time
func (agent *Agent) SetTelemetryCallback(callback TelemetryCallback) {
	agent.telemetryCallback = callback
}

// CaptureRequest captures request data for telemetry (can be called by derived agents)
func (agent *Agent) CaptureRequest(params openai.ChatCompletionNewParams) {
	agent.captureRequest(params)
}

// CaptureResponse captures response data for telemetry (can be called by derived agents)
func (agent *Agent) CaptureResponse(response *openai.ChatCompletion, startTime time.Time) {
	agent.captureResponse(response, startTime)
}

// CaptureError captures error data for telemetry (can be called by derived agents)
func (agent *Agent) CaptureError(err error, context string) {
	agent.captureError(err, context)
}

// captureRequest captures request data for telemetry (internal method)
func (agent *Agent) captureRequest(params openai.ChatCompletionNewParams) {
	// Deep copy the request params
	agent.lastRequest = &params
	agent.lastRequestTime = time.Now()

	// Trigger callback if set
	if agent.telemetryCallback != nil {
		metadata := agent.GetLastRequestMetadata()
		requestJSON, _ := agent.GetLastRequestJSON()
		agent.telemetryCallback.OnRequestSent(metadata, requestJSON)
	}
}

// captureResponse captures response data for telemetry (internal method)
func (agent *Agent) captureResponse(response *openai.ChatCompletion, startTime time.Time) {
	agent.lastResponse = response
	agent.lastResponseTime = time.Now()
	agent.lastResponseDuration = time.Since(startTime)

	// Update total tokens
	if response.Usage.TotalTokens > 0 {
		agent.totalTokensUsed += int(response.Usage.TotalTokens)
	}

	// Trigger callback if set
	if agent.telemetryCallback != nil {
		metadata := agent.GetLastResponseMetadata()
		responseJSON, _ := agent.GetLastResponseJSON()
		agent.telemetryCallback.OnResponseReceived(metadata, responseJSON)
	}
}

// captureError captures error data for telemetry (internal method)
func (agent *Agent) captureError(err error, context string) {
	if agent.telemetryCallback != nil {
		agent.telemetryCallback.OnError(err, context)
	}
}
