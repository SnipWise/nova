package crew

import (
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
)

// SetOrchestratorAgent sets the orchestrator agent
func (agent *CrewAgent) SetOrchestratorAgent(orchestratorAgent agents.OrchestratorAgent) {
	agent.orchestratorAgent = orchestratorAgent
}

// GetOrchestratorAgent returns the orchestrator agent
func (agent *CrewAgent) GetOrchestratorAgent() agents.OrchestratorAgent {
	return agent.orchestratorAgent
}

// DetectTopicThenGetAgentId detects the topic of the query and returns the agent ID to route to.
func (ca *CrewAgent) DetectTopicThenGetAgentId(query string) (string, error) {
	return serverbase.DetectTopicAndGetAgentId(
		ca.log,
		ca.orchestratorAgent,
		ca.selectedAgentId,
		ca.matchAgentIdToTopicFn,
		func(id string) bool { _, ok := ca.chatAgents[id]; return ok },
		query,
	)
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
