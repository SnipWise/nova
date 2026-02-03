package gatewayserver

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// SetCompressorAgent sets the compressor agent.
func (agent *GatewayServerAgent) SetCompressorAgent(compressorAgent *compressor.Agent) {
	agent.compressorAgent = compressorAgent
}

// GetCompressorAgent returns the compressor agent.
func (agent *GatewayServerAgent) GetCompressorAgent() *compressor.Agent {
	return agent.compressorAgent
}

// SetContextSizeLimit sets the context size limit for compression.
func (agent *GatewayServerAgent) SetContextSizeLimit(limit int) {
	agent.contextSizeLimit = limit
}

// GetContextSizeLimit returns the context size limit for compression.
func (agent *GatewayServerAgent) GetContextSizeLimit() int {
	return agent.contextSizeLimit
}

// compressContextIfNeeded compresses the chat agent context if it exceeds the size limit.
func (agent *GatewayServerAgent) compressContextIfNeeded() {
	if agent.compressorAgent == nil {
		return
	}

	contextSize := agent.currentChatAgent.GetContextSize()
	if contextSize <= agent.contextSizeLimit {
		return
	}

	agent.log.Info("🗜️ Context size %d exceeds limit %d, compressing...", contextSize, agent.contextSizeLimit)

	newSize, err := agent.CompressChatAgentContextIfOverLimit()
	if err != nil {
		agent.log.Error("Compression failed: %v", err)
		return
	}

	if newSize > 0 {
		agent.log.Info("🗜️ Context compressed from %d to %d bytes", contextSize, newSize)
	}
}

// CompressChatAgentContextIfOverLimit compresses the chat agent context if it exceeds the size limit.
func (agent *GatewayServerAgent) CompressChatAgentContextIfOverLimit() (int, error) {
	if agent.contextSizeLimit == 0 {
		return 0, nil
	}

	if agent.compressorAgent == nil {
		return 0, fmt.Errorf("compressor agent is not set")
	}

	if agent.currentChatAgent.GetContextSize() > agent.contextSizeLimit {
		newContext, err := agent.compressorAgent.CompressContext(agent.currentChatAgent.GetMessages())
		if err != nil {
			return 0, err
		}
		agent.currentChatAgent.ResetMessages()
		agent.currentChatAgent.AddMessage(roles.System, newContext.CompressedText)

		if len(newContext.CompressedText) > int(0.8*float64(agent.contextSizeLimit)) {
			return len(newContext.CompressedText), fmt.Errorf("compressed context size %d still exceeds 80%% of limit %d", len(newContext.CompressedText), agent.contextSizeLimit)
		}
		if len(newContext.CompressedText) > int(0.9*float64(agent.contextSizeLimit)) {
			agent.log.Warn("Compressed context size %d exceeds 90%% of limit %d; resetting", len(newContext.CompressedText), agent.contextSizeLimit)
			agent.currentChatAgent.ResetMessages()
			return len(newContext.CompressedText), fmt.Errorf("compressed context size %d still exceeds 90%% of limit %d", len(newContext.CompressedText), agent.contextSizeLimit)
		}
		return len(newContext.CompressedText), nil
	}

	return 0, nil
}
