package remote

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// CompletionResult represents the result of a chat completion
type CompletionResult struct {
	Response     string
	FinishReason string
}

// ReasoningResult represents the result of a chat completion with reasoning
type ReasoningResult struct {
	Response     string
	Reasoning    string
	FinishReason string
}

// StreamCallback is a function called for each chunk of streaming response
type StreamCallback func(chunk string, finishReason string) error

// ToolCallCallback is a function called when a tool call is detected
type ToolCallCallback func(operationID string, message string) error

// Agent represents a remote chat agent that communicates with a server agent via HTTP
type Agent struct {
	ctx              context.Context
	config           agents.Config
	baseURL          string
	client           *http.Client
	log              logger.Logger
	toolCallCallback ToolCallCallback
}

// NewAgent creates a new remote chat agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	baseURL string,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	if baseURL == "" {
		return nil, errors.New("baseURL cannot be empty")
	}

	agent := &Agent{
		ctx:     ctx,
		config:  agentConfig,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{},
		log:     log,
	}

	return agent, nil
}

// SetToolCallCallback sets the callback for tool call notifications
func (agent *Agent) SetToolCallCallback(callback ToolCallCallback) {
	agent.toolCallCallback = callback
}

// Kind returns the agent type
func (agent *Agent) Kind() agents.Kind {
	return agents.Remote
}

// GetName returns the agent name
func (agent *Agent) GetName() string {
	return agent.config.Name
}

// ModelsInfo contains information about the models used by the server
type ModelsInfo struct {
	Status           string `json:"status"`
	ChatModel        string `json:"chat_model"`
	EmbeddingsModel  string `json:"embeddings_model"`
	ToolsModel       string `json:"tools_model"`
}

// HealthStatus contains the health status of the server
type HealthStatus struct {
	Status string `json:"status"`
}

// GetModelID returns the model ID from the server
func (agent *Agent) GetModelID() string {
	modelsInfo, err := agent.GetModelsInfo()
	if err != nil {
		agent.log.Error("Failed to get model info: %v", err)
		return "unknown"
	}
	return modelsInfo.ChatModel
}

// GetModelsInfo returns detailed information about all models used by the server
func (agent *Agent) GetModelsInfo() (*ModelsInfo, error) {
	resp, err := agent.client.Get(agent.baseURL + "/models")
	if err != nil {
		return nil, fmt.Errorf("failed to get models info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var info ModelsInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode models info: %w", err)
	}

	return &info, nil
}

// GetHealth checks if the server is healthy and reachable
func (agent *Agent) GetHealth() (*HealthStatus, error) {
	resp, err := agent.client.Get(agent.baseURL + "/health")
	if err != nil {
		return nil, fmt.Errorf("failed to get health status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var health HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health status: %w", err)
	}

	return &health, nil
}

// IsHealthy is a convenience method that returns true if the server is healthy
func (agent *Agent) IsHealthy() bool {
	health, err := agent.GetHealth()
	if err != nil {
		agent.log.Debug("Health check failed: %v", err)
		return false
	}
	return health.Status == "ok"
}

// GetMessages returns all conversation messages from the server
func (agent *Agent) GetMessages() []messages.Message {
	resp, err := agent.client.Get(agent.baseURL + "/memory/messages/list")
	if err != nil {
		agent.log.Error("Failed to get messages: %v", err)
		return []messages.Message{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		agent.log.Error("Failed to get messages: status %d", resp.StatusCode)
		return []messages.Message{}
	}

	var result struct {
		Messages []messages.Message `json:"messages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		agent.log.Error("Failed to decode messages: %v", err)
		return []messages.Message{}
	}

	return result.Messages
}

// GetContextSize returns the approximate size of the current context from the server
func (agent *Agent) GetContextSize() int {
	resp, err := agent.client.Get(agent.baseURL + "/memory/messages/context-size")
	if err != nil {
		agent.log.Error("Failed to get context size: %v", err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		agent.log.Error("Failed to get context size: status %d", resp.StatusCode)
		return 0
	}

	var result struct {
		Tokens int `json:"tokens"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		agent.log.Error("Failed to decode tokens: %v", err)
		return 0
	}

	return result.Tokens
}

// StopStream interrupts the current streaming operation
func (agent *Agent) StopStream() {
	req, err := http.NewRequestWithContext(agent.ctx, "POST", agent.baseURL+"/completion/stop", nil)
	if err != nil {
		agent.log.Error("Failed to create stop request: %v", err)
		return
	}

	resp, err := agent.client.Do(req)
	if err != nil {
		agent.log.Error("Failed to stop stream: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		agent.log.Error("Failed to stop stream: status %d", resp.StatusCode)
	}
}

// ResetMessages clears all messages except the system instruction
func (agent *Agent) ResetMessages() {
	req, err := http.NewRequestWithContext(agent.ctx, "POST", agent.baseURL+"/memory/reset", nil)
	if err != nil {
		agent.log.Error("Failed to create reset request: %v", err)
		return
	}

	resp, err := agent.client.Do(req)
	if err != nil {
		agent.log.Error("Failed to reset messages: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		agent.log.Error("Failed to reset messages: status %d", resp.StatusCode)
	}
}

// AddMessage adds a message to the conversation history
// Note: This is a no-op for remote agent as messages are managed server-side
func (agent *Agent) AddMessage(role roles.Role, content string) {
	// Remote agent doesn't maintain local message history
	// Messages are managed by the server
	agent.log.Debug("AddMessage called but not implemented for remote agent")
}

// AddMessages adds multiple messages to the conversation history
// Note: This is a no-op for remote agent as messages are managed server-side
func (agent *Agent) AddMessages(msgs []messages.Message) {
	// Remote agent doesn't maintain local message history
	// Messages are managed by the server
	agent.log.Debug("AddMessages called but not implemented for remote agent")
}

// GenerateCompletion sends messages and returns the completion result
func (agent *Agent) GenerateCompletion(userMessages []messages.Message) (*CompletionResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// For now, we'll use streaming under the hood and collect the result
	var fullResponse strings.Builder
	var lastFinishReason string

	_, err := agent.GenerateStreamCompletion(userMessages, func(chunk string, finishReason string) error {
		fullResponse.WriteString(chunk)
		if finishReason != "" {
			lastFinishReason = finishReason
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &CompletionResult{
		Response:     fullResponse.String(),
		FinishReason: lastFinishReason,
	}, nil
}

// GenerateCompletionWithReasoning sends messages and returns the completion result with reasoning
// Note: Reasoning is not yet supported by the server API
func (agent *Agent) GenerateCompletionWithReasoning(userMessages []messages.Message) (*ReasoningResult, error) {
	// The server doesn't expose reasoning separately, so we'll just call the regular completion
	result, err := agent.GenerateCompletion(userMessages)
	if err != nil {
		return nil, err
	}

	return &ReasoningResult{
		Response:     result.Response,
		Reasoning:    "",
		FinishReason: result.FinishReason,
	}, nil
}

// GenerateStreamCompletion sends messages and streams the response via callback
func (agent *Agent) GenerateStreamCompletion(
	userMessages []messages.Message,
	callback StreamCallback,
) (*CompletionResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Combine all user messages into one (server expects a single message)
	var messageContent strings.Builder
	for i, msg := range userMessages {
		if i > 0 {
			messageContent.WriteString("\n")
		}
		messageContent.WriteString(msg.Content)
	}

	requestBody := map[string]interface{}{
		"data": map[string]string{
			"message": messageContent.String(),
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		agent.ctx,
		"POST",
		agent.baseURL+"/completion",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := agent.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse SSE stream
	scanner := bufio.NewScanner(resp.Body)
	var fullResponse strings.Builder
	var lastFinishReason string

	for scanner.Scan() {
		line := scanner.Text()

		// SSE format: "data: {json}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "" {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			agent.log.Error("Failed to parse SSE event: %v", err)
			continue
		}

		// Handle different event types
		if kind, ok := event["kind"].(string); ok && kind == "tool_call" {
			// Tool call notification - display operation details
			operationID := ""
			if id, ok := event["operation_id"].(string); ok {
				operationID = id
			}
			message := ""
			if msg, ok := event["message"].(string); ok {
				message = msg
			}

			// Log the tool call notification
			agent.log.Info("\n🔔 Tool Call Detected: %s", message)
			agent.log.Info("📝 Operation ID: %s", operationID)
			agent.log.Info("✅ To validate: curl -X POST http://localhost:8080/operation/validate -d '{\"operation_id\":\"%s\"}'", operationID)
			agent.log.Info("⛔️ To cancel:   curl -X POST http://localhost:8080/operation/cancel -d '{\"operation_id\":\"%s\"}'\n", operationID)

			// Call the tool call callback if set
			if agent.toolCallCallback != nil {
				if err := agent.toolCallCallback(operationID, message); err != nil {
					return nil, err
				}
			}

			continue
		}

		// Handle message chunks
		if message, ok := event["message"].(string); ok {
			fullResponse.WriteString(message)
			if callback != nil {
				finishReason := ""
				if fr, ok := event["finish_reason"].(string); ok {
					finishReason = fr
					lastFinishReason = fr
				}
				if err := callback(message, finishReason); err != nil {
					return nil, err
				}
			}
		}

		// Handle finish reason
		if fr, ok := event["finish_reason"].(string); ok {
			lastFinishReason = fr
		}

		// Handle errors
		if errMsg, ok := event["error"].(string); ok {
			return nil, fmt.Errorf("server error: %s", errMsg)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %w", err)
	}

	return &CompletionResult{
		Response:     fullResponse.String(),
		FinishReason: lastFinishReason,
	}, nil
}

// GenerateStreamCompletionWithReasoning sends messages and streams both reasoning and response
// Note: The server doesn't support separate reasoning streams yet
func (agent *Agent) GenerateStreamCompletionWithReasoning(
	userMessages []messages.Message,
	reasoningCallback StreamCallback,
	responseCallback StreamCallback,
) (*ReasoningResult, error) {
	// The server doesn't expose reasoning separately, so we'll just call the regular streaming
	result, err := agent.GenerateStreamCompletion(userMessages, responseCallback)
	if err != nil {
		return nil, err
	}

	return &ReasoningResult{
		Response:     result.Response,
		Reasoning:    "",
		FinishReason: result.FinishReason,
	}, nil
}

// ExportMessagesToJSON exports the conversation history to JSON
func (agent *Agent) ExportMessagesToJSON() (string, error) {
	messagesList := agent.GetMessages()
	jsonData, err := json.MarshalIndent(messagesList, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ValidateOperation validates a pending tool call operation on the server
func (agent *Agent) ValidateOperation(operationID string) error {
	if operationID == "" {
		return errors.New("operation_id cannot be empty")
	}

	requestBody := map[string]string{
		"operation_id": operationID,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		agent.ctx,
		"POST",
		agent.baseURL+"/operation/validate",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := agent.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	agent.log.Info("✅ Operation %s validated successfully", operationID)
	return nil
}

// CancelOperation cancels a pending tool call operation on the server
func (agent *Agent) CancelOperation(operationID string) error {
	if operationID == "" {
		return errors.New("operation_id cannot be empty")
	}

	requestBody := map[string]string{
		"operation_id": operationID,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		agent.ctx,
		"POST",
		agent.baseURL+"/operation/cancel",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := agent.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	agent.log.Info("⛔️ Operation %s cancelled successfully", operationID)
	return nil
}

// ResetOperations cancels all pending tool call operations on the server
func (agent *Agent) ResetOperations() error {
	req, err := http.NewRequestWithContext(
		agent.ctx,
		"POST",
		agent.baseURL+"/operation/reset",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := agent.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	agent.log.Info("🔄 All pending operations reset successfully")
	return nil
}
