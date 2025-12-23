package serverbase

import (
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// ChatAgent defines the interface for the underlying agent (chat or crew)
// that will handle the actual message processing
type ChatAgent interface {
	Kind() agents.Kind
	GetName() string
	GetModelID() string
	GetMessages() []messages.Message
	GetContextSize() int
	StopStream()
	ResetMessages()
	AddMessage(role roles.Role, content string)
	GenerateCompletion(userMessages []messages.Message) (*chat.CompletionResult, error)
	GenerateCompletionWithReasoning(userMessages []messages.Message) (*chat.ReasoningResult, error)
	GenerateStreamCompletion(userMessages []messages.Message, callback chat.StreamCallback) (*chat.CompletionResult, error)
	GenerateStreamCompletionWithReasoning(userMessages []messages.Message, reasoningCallback chat.StreamCallback, responseCallback chat.StreamCallback) (*chat.ReasoningResult, error)
	ExportMessagesToJSON() (string, error)
}

// ServerAgentConfig contains common configuration for server agents
type ServerAgentConfig struct {
	ChatAgent        ChatAgent
	ToolsAgent       *tools.Agent
	RagAgent         *rag.Agent
	SimilarityLimit  float64
	MaxSimilarities  int
	ContextSizeLimit int
	CompressorAgent  *compressor.Agent
	ExecuteFn        func(string, string) (string, error)
}
