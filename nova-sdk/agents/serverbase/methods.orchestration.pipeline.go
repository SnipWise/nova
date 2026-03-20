package serverbase

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// DetectTopicAndGetAgentId detects the topic of a query and returns the agent ID to route to.
// matchFn maps (currentAgentId, detectedTopic) → target agentId.
// hasAgent checks whether an agent with the given ID exists.
func DetectTopicAndGetAgentId(
	log logger.Logger,
	orchestratorAgent agents.OrchestratorAgent,
	selectedAgentId string,
	matchFn func(selectedID string, topic string) string,
	hasAgent func(agentId string) bool,
	query string,
) (string, error) {
	log.Info("🔍 Detecting topic for routing...")
	log.Info("📝 Query: " + query)

	response, _, err := orchestratorAgent.IdentifyIntent([]messages.Message{
		{Role: roles.User, Content: query},
	})
	if err != nil {
		return "", err
	}

	log.Info("✅ Topic detected: " + response.TopicDiscussion)

	agentId := matchFn(selectedAgentId, response.TopicDiscussion)

	if !hasAgent(agentId) {
		return "", fmt.Errorf("no chat agent found with ID: %s", agentId)
	}

	log.Info("🔀 You should route to agent ID: " + agentId)

	return agentId, nil
}
