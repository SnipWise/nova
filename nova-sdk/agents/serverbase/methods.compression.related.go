package serverbase

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// SetCompressorAgent sets the compressor agent
func (agent *BaseServerAgent) SetCompressorAgent(compressorAgent *compressor.Agent) {
	agent.CompressorAgent = compressorAgent
}

// GetCompressorAgent returns the compressor agent
func (agent *BaseServerAgent) GetCompressorAgent() *compressor.Agent {
	return agent.CompressorAgent
}

// SetContextSizeLimit sets the context size limit for compression
func (agent *BaseServerAgent) SetContextSizeLimit(limit int) {
	agent.ContextSizeLimit = limit
}

// GetContextSizeLimit returns the context size limit for compression
func (agent *BaseServerAgent) GetContextSizeLimit() int {
	return agent.ContextSizeLimit
}

// CompressChatAgentContextIfOverLimit compresses the chat agent context if it exceeds the size limit.
func (agent *BaseServerAgent) CompressChatAgentContextIfOverLimit(chatAgent ChatAgent) (int, error) {
	if agent.ContextSizeLimit == 0 {
		agent.Log.Debug("No context size limit set; skipping compression")
		return 0, nil
	}

	if agent.CompressorAgent == nil {
		return 0, fmt.Errorf("compressor agent is not set")
	}

	if chatAgent.GetContextSize() <= agent.ContextSizeLimit {
		return 0, nil
	}

	agent.Log.Info("Chat agent context size %d exceeds limit of %d; compressing...", chatAgent.GetContextSize(), agent.ContextSizeLimit)

	newContext, err := agent.CompressorAgent.CompressContext(chatAgent.GetMessages())
	if err != nil {
		return 0, err
	}
	chatAgent.ResetMessages()
	if newContext.CompressedText != "" {
		chatAgent.AddMessage(roles.System, newContext.CompressedText)
	}
	if len(newContext.CompressedText) > int(0.8*float64(agent.ContextSizeLimit)) {
		return len(newContext.CompressedText), fmt.Errorf("compressed context size %d still exceeds 80%% of limit %d", len(newContext.CompressedText), agent.ContextSizeLimit)
	}
	if len(newContext.CompressedText) > int(0.9*float64(agent.ContextSizeLimit)) {
		agent.Log.Warn("Compressed context size %d exceeds 90%% of limit %d; resetting chat agent messages", len(newContext.CompressedText), agent.ContextSizeLimit)
		chatAgent.ResetMessages()
		return len(newContext.CompressedText), fmt.Errorf("compressed context size %d still exceeds 90%% of limit %d", len(newContext.CompressedText), agent.ContextSizeLimit)
	}
	return len(newContext.CompressedText), nil
}

// CompressChatAgentContext compresses the chat agent context unconditionally.
func (agent *BaseServerAgent) CompressChatAgentContext(chatAgent ChatAgent) (int, error) {
	if agent.ContextSizeLimit == 0 {
		return 0, nil
	}

	if agent.CompressorAgent == nil {
		return 0, fmt.Errorf("compressor agent is not set")
	}

	newContext, err := agent.CompressorAgent.CompressContext(chatAgent.GetMessages())
	if err != nil {
		return 0, err
	}
	chatAgent.ResetMessages()
	if newContext.CompressedText != "" {
		chatAgent.AddMessage(roles.System, newContext.CompressedText)
	}
	return len(newContext.CompressedText), nil
}
