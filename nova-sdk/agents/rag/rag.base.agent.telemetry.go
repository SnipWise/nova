package rag

import (
	"encoding/json"
	"time"

	"github.com/openai/openai-go/v3"
)

// EmbeddingRequestMetadata contains metadata about an embedding request
type EmbeddingRequestMetadata struct {
	Model          string
	InputLength    int
	EncodingFormat string
	Dimensions     int
	Timestamp      time.Time
}

// EmbeddingResponseMetadata contains metadata about an embedding response
type EmbeddingResponseMetadata struct {
	Model           string
	EmbeddingIndex  int
	VectorDimension int
	PromptTokens    int
	TotalTokens     int
	ResponseTime    int64 // in milliseconds
	Timestamp       time.Time
}

// captureEmbeddingRequest stores the embedding request parameters
func (agent *BaseAgent) captureEmbeddingRequest(params openai.EmbeddingNewParams) {
	agent.lastEmbeddingRequest = &params
	agent.lastEmbeddingRequestTime = time.Now()
}

// captureEmbeddingResponse stores the embedding response and calculates duration
func (agent *BaseAgent) captureEmbeddingResponse(response *openai.Embedding, startTime time.Time) {
	agent.lastEmbeddingResponse = response
	agent.lastEmbeddingResponseTime = time.Now()
	agent.lastEmbeddingResponseDuration = time.Since(startTime)
}

// GetLastEmbeddingRequestJSON returns the last embedding request as JSON string
func (agent *BaseAgent) GetLastEmbeddingRequestJSON() (string, error) {
	if agent.lastEmbeddingRequest == nil {
		return "", nil
	}

	jsonBytes, err := json.MarshalIndent(agent.lastEmbeddingRequest, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetLastEmbeddingResponseJSON returns the last embedding response as JSON string
func (agent *BaseAgent) GetLastEmbeddingResponseJSON() (string, error) {
	if agent.lastEmbeddingResponse == nil {
		return "", nil
	}

	jsonBytes, err := json.MarshalIndent(agent.lastEmbeddingResponse, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetLastEmbeddingRequestMetadata returns metadata about the last embedding request
func (agent *BaseAgent) GetLastEmbeddingRequestMetadata() EmbeddingRequestMetadata {
	if agent.lastEmbeddingRequest == nil {
		return EmbeddingRequestMetadata{}
	}

	// Simple metadata extraction - just get the model and timestamp
	// Input length calculation is complex due to union types
	return EmbeddingRequestMetadata{
		Model:     agent.lastEmbeddingRequest.Model,
		Timestamp: agent.lastEmbeddingRequestTime,
	}
}

// GetLastEmbeddingResponseMetadata returns metadata about the last embedding response
func (agent *BaseAgent) GetLastEmbeddingResponseMetadata() EmbeddingResponseMetadata {
	if agent.lastEmbeddingResponse == nil {
		return EmbeddingResponseMetadata{}
	}

	vectorDim := 0
	if len(agent.lastEmbeddingResponse.Embedding) > 0 {
		vectorDim = len(agent.lastEmbeddingResponse.Embedding)
	}

	return EmbeddingResponseMetadata{
		Model:           string(agent.lastEmbeddingResponse.Object),
		EmbeddingIndex:  int(agent.lastEmbeddingResponse.Index),
		VectorDimension: vectorDim,
		PromptTokens:    0, // Not provided in single embedding response
		TotalTokens:     0, // Not provided in single embedding response
		ResponseTime:    agent.lastEmbeddingResponseDuration.Milliseconds(),
		Timestamp:       agent.lastEmbeddingResponseTime,
	}
}
