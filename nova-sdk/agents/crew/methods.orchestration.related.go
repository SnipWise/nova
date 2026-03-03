package crew

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// SetOrchestratorAgent sets the orchestrator agent
func (agent *CrewAgent) SetOrchestratorAgent(orchestratorAgent agents.OrchestratorAgent) {
	agent.orchestratorAgent = orchestratorAgent
}

// GetOrchestratorAgent returns the orchestrator agent
func (agent *CrewAgent) GetOrchestratorAgent() agents.OrchestratorAgent {
	return agent.orchestratorAgent
}

// DetectTopicThenSetCurrentAgent sets the current active chat agent based on the detected topic from the query.
// query -> user message
func (ca *CrewAgent) DetectTopicThenGetAgentId(query string) (string, error) {

	ca.log.Info("🔍 Detecting topic for routing...")
	ca.log.Info("📝 Query: " + query)
	// Topic detection via orchestrator agent
	response, _, err := ca.orchestratorAgent.IdentifyIntent([]messages.Message{
		{
			Role:    roles.User,
			Content: query,
		},
	})
	if err != nil {
		return "", err
	}

	ca.log.Info("✅ Topic detected: " + response.TopicDiscussion)

	// --------------------------------------------------------
	// Get agent ID based on detected topic
	// --------------------------------------------------------
	agentId := ca.matchAgentIdToTopicFn(ca.selectedAgentId, response.TopicDiscussion)

	if _, exists := ca.chatAgents[agentId]; !exists {
		return "", fmt.Errorf("no chat agent found with ID: %s", agentId)
	}

	ca.log.Info("🔀 You should route to agent ID: " + agentId)

	return agentId, nil
}

/*
	// ------------------------------------------------
	// Define the function to match agent ID to topic
	// ------------------------------------------------
	matchAgentFunction := func(topic string) string {
		fmt.Println("🔵 Matching agent for topic:", topic)
		var agentId string
		switch strings.ToLower(topic) {
		case "coding", "programming", "development", "code", "software", "debugging", "technology", "software sevelopment":
			agentId = "coder"
		case "philosophy", "thinking", "ideas", "thoughts", "psychology", "relationships", "math", "mathematics", "science":
			agentId = "thinker"
		case "cooking", "recipe", "food", "culinary", "baking", "grilling", "meal":
			agentId = "cook"
		default:
			agentId = "generic"
		}
		fmt.Println("🟢 Matched agent ID:", agentId)
		return agentId
	}

*/
