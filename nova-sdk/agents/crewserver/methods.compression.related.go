package crewserver

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// SetCompressorAgent sets the compressor agent
func (agent *CrewServerAgent) SetCompressorAgent(compressorAgent *compressor.Agent) {
	agent.compressorAgent = compressorAgent
}

// GetCompressorAgent returns the compressor agent
func (agent *CrewServerAgent) GetCompressorAgent() *compressor.Agent {
	return agent.compressorAgent
}

// SetContextSizeLimit sets the context size limit for compression
func (agent *CrewServerAgent) SetContextSizeLimit(limit int) {
	agent.contextSizeLimit = limit
}

// GetContextSizeLimit returns the context size limit for compression
func (agent *CrewServerAgent) GetContextSizeLimit() int {
	return agent.contextSizeLimit
}

// CompressChatAgentContextIfOverLimit compresses the chat agent context if it exceeds the size limit.
func (agent *CrewServerAgent) CompressChatAgentContextIfOverLimit() (int, error) {

	if agent.contextSizeLimit == 0 {
		agent.log.Debug("No context size limit set; skipping compression")
		return 0, nil // No limit set
	}

	if agent.compressorAgent == nil {
		return 0, fmt.Errorf("compressor agent is not set")
	}

	if agent.currentChatAgent.GetContextSize() > agent.contextSizeLimit {
		agent.log.Info("Chat agent context size %d exceeds limit of %d; compressing...", agent.currentChatAgent.GetContextSize(), agent.contextSizeLimit)

		newContext, err := agent.compressorAgent.CompressContext(agent.currentChatAgent.GetMessages())
		if err != nil {
			return 0, err
		}
		agent.currentChatAgent.ResetMessages()
		agent.currentChatAgent.AddMessage(
			roles.System,
			newContext.CompressedText,
		)
		// IMPORTANT: if the new context is still over the limit, we might need to handle that case
		// if the new size is arround 80% of the limit, return an error
		if len(newContext.CompressedText) > int(0.8*float64(agent.contextSizeLimit)) {
			return len(newContext.CompressedText), fmt.Errorf("compressed context size %d still exceeds 80%% of limit %d", len(newContext.CompressedText), agent.contextSizeLimit)
		}
		if len(newContext.CompressedText) > int(0.9*float64(agent.contextSizeLimit)) {
			agent.log.Warn("Compressed context size %d exceeds 90%% of limit %d; resetting chat agent messages", len(newContext.CompressedText), agent.contextSizeLimit)
			agent.currentChatAgent.ResetMessages()
			return len(newContext.CompressedText), fmt.Errorf("compressed context size %d still exceeds 90%% of limit %d", len(newContext.CompressedText), agent.contextSizeLimit)
		}
		return len(newContext.CompressedText), nil
	}

	return 0, nil
}

// CompressChatAgentContext compresses the chat agent context.
func (agent *CrewServerAgent) CompressChatAgentContext() (int, error) {

	if agent.contextSizeLimit == 0 {
		return 0, nil // No limit set
	}

	if agent.compressorAgent == nil {
		return 0, fmt.Errorf("compressor agent is not set")
	}

	newContext, err := agent.compressorAgent.CompressContext(agent.currentChatAgent.GetMessages())
	if err != nil {
		return 0, err
	}
	agent.currentChatAgent.ResetMessages()
	agent.currentChatAgent.AddMessage(
		roles.System,
		newContext.CompressedText,
	)
	return len(newContext.CompressedText), nil
}
