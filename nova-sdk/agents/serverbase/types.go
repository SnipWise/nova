package serverbase

import (
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
)

// ToolCallNotification represents a notification about a pending tool call
type ToolCallNotification struct {
	OperationID  string
	FunctionName string
	Arguments    string
	Message      string
}

// PendingOperation represents a tool call operation awaiting user confirmation
type PendingOperation struct {
	ID           string
	FunctionName string
	Arguments    string
	Response     chan tools.ConfirmationResponse
}

// CompletionRequest represents an HTTP request for chat completion
type CompletionRequest struct {
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

// OperationRequest represents an HTTP request for operation management
type OperationRequest struct {
	OperationID string `json:"operation_id"`
}

// MemoryResponse represents the response containing conversation history
type MemoryResponse struct {
	Messages []messages.Message `json:"messages"`
}

// ContextSizeResponse represents the response containing token count information
type ContextSizeResponse struct {
	MessagesCount   int `json:"messages_count"`
	CharactersCount int `json:"characters_count"`
	Limit           int `json:"limit"`
}
