package gatewayserver

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// SetOrchestratorAgent sets the orchestrator agent.
func (agent *GatewayServerAgent) SetOrchestratorAgent(orchestratorAgent agents.OrchestratorAgent) {
	agent.orchestratorAgent = orchestratorAgent
}

// GetOrchestratorAgent returns the orchestrator agent.
func (agent *GatewayServerAgent) GetOrchestratorAgent() agents.OrchestratorAgent {
	return agent.orchestratorAgent
}

// routeToAppropriateAgent detects topic and switches to the appropriate agent.
func (agent *GatewayServerAgent) routeToAppropriateAgent(question string) {
	if agent.orchestratorAgent == nil {
		return
	}

	detectedAgentId, err := agent.DetectTopicThenGetAgentId(question)
	if err != nil {
		agent.log.Error("Error during topic detection: %v", err)
		return
	}

	if detectedAgentId != "" && agent.chatAgents[detectedAgentId] != agent.currentChatAgent {
		agent.log.Info("🔀 Switching to detected agent ID: %s", detectedAgentId)
		agent.currentChatAgent = agent.chatAgents[detectedAgentId]
		agent.selectedAgentId = detectedAgentId
	}
}

// DetectTopicThenGetAgentId uses the orchestrator to detect the topic and returns the matching agent ID.
func (agent *GatewayServerAgent) DetectTopicThenGetAgentId(query string) (string, error) {
	agent.log.Info("🔍 Detecting topic for routing...")

	response, _, err := agent.orchestratorAgent.IdentifyIntent([]messages.Message{
		{
			Role:    roles.User,
			Content: query,
		},
	})
	if err != nil {
		return "", err
	}

	agent.log.Info("✅ Topic detected: %s", response.TopicDiscussion)

	agentId := agent.matchAgentIdToTopicFn(agent.selectedAgentId, response.TopicDiscussion)

	if _, exists := agent.chatAgents[agentId]; !exists {
		return "", fmt.Errorf("no chat agent found with ID: %s", agentId)
	}

	agent.log.Info("🔀 Should route to agent ID: %s", agentId)

	return agentId, nil
}
