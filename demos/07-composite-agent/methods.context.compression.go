package main

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
)

// GetCurrentChatAgentContextSize retrieves the context size of the current active chat agent.
func (ca *CompositeAgent) GetCurrentAgentContextSize() int {

	return ca.currentAgent.GetContextSize()
}

// GetChatAgentContextSizeByName retrieves the context size of a chat agent by its name.
func (ca *CompositeAgent) GetChatAgentContextSizeByName(name string) (int, error) {
	agent, exists := ca.chatAgents[name]
	if !exists {
		return 0, fmt.Errorf("agent with name %s does not exist", name)
	}
	return agent.GetContextSize(), nil
}

// CompressCurrentChatAgentContextIfOverLimit compresses the context of the current active chat agent if it exceeds a predefined limit.
func (ca *CompositeAgent) CompressCurrentChatAgentContextIfOverLimit() error {

	contextCompressingThreshold := env.GetEnvOrDefault("CONTEXT_COMPRESSING_THRESHOLD", "6000")
	contextCompressingThresholdInt := conversion.StringToInt(contextCompressingThreshold)

	if ca.currentAgent == nil {
		return fmt.Errorf("current agent is not set")
	}
	if ca.currentAgent.GetContextSize() > contextCompressingThresholdInt {
		newContext, err := ca.compressorAgent.CompressContext(ca.currentAgent.GetMessages())
		if err != nil {
			return err
		}
		ca.currentAgent.ResetMessages()
		ca.currentAgent.AddMessage(
			roles.System,
			newContext.CompressedText,
		)
	}

	return nil
}

// CompressCurrentChatAgentContext compresses the context of the current active chat agent.
func (ca *CompositeAgent) CompressCurrentChatAgentContext() error {

	newContext, err := ca.compressorAgent.CompressContext(ca.currentAgent.GetMessages())
	if err != nil {
		return err
	}
	ca.currentAgent.ResetMessages()
	ca.currentAgent.AddMessage(
		roles.System,
		newContext.CompressedText,
	)

	return nil
}

// ResetCurrentChatAgentMemory resets the message history of the current active chat agent.
func (ca *CompositeAgent) ResetCurrentChatAgentMemory() error {
	if ca.currentAgent == nil {
		return fmt.Errorf("current agent is not set")
	}
	ca.currentAgent.ResetMessages()
	return nil
}

// ResetChatAgentMemory resets the message history of a chat agent by its name.
func (ca *CompositeAgent) ResetChatAgentMemory(agentName string) error {
	agent, exists := ca.chatAgents[agentName]
	if !exists {
		return fmt.Errorf("agent with name %s does not exist", agentName)
	}
	agent.ResetMessages()
	return nil
}

// ResetToolsAgentMemory resets the message history of the tools agent.
func (ca *CompositeAgent) ResetToolsAgentMemory() error {
	if ca.toolsAgent == nil {
		return fmt.Errorf("tools agent is not initialized")
	}
	ca.toolsAgent.ResetMessages()
	return nil
}
