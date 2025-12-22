package main

import (
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// SetCurrentAgentByTopic sets the current active chat agent based on the detected topic from the query.
func (ca *CompositeAgent) SetCurrentAgentByTopic(query string) (*chat.Agent, error) {

	fmt.Println("🔍 Detecting topic for routing...")
	fmt.Println("📝 Query: " + query)

	// Topic detection via orchestrator agent
	response, _, err := ca.orchestratorAgent.GenerateStructuredData([]messages.Message{
		{
			Role:    roles.User,
			Content: query,
		},
	})
	if err != nil {
		return nil, err
	}
	
	fmt.Println("✅ Topic detected: " + response.TopicDiscussion)

	// --------------------------------------------------------
	// Select agent based on detected topic
	// --------------------------------------------------------
	// TODO: make this configurable via env vars or config file
	// TODO: improve topic matching with more advanced techniques (e.g., embeddings, fuzzy matching), contains words...
	switch strings.ToLower(response.TopicDiscussion) {
	case "coding", "programming", "development", "code", "software", "debugging", "technology", "software sevelopment":
		ca.currentAgent = ca.chatAgents["coder"]
	case "philosophy", "thinking", "ideas", "thoughts", "psychology", "relationships", "math", "mathematics", "science":
		ca.currentAgent = ca.chatAgents["thinker"]
	case "translation", "translate":
		ca.currentAgent = ca.chatAgents["generic"]
	case "cooking", "recipe", "food", "culinary", "baking", "grilling", "meal":
		ca.currentAgent = ca.chatAgents["generic"]
	default:
		ca.currentAgent = ca.chatAgents["generic"]
	}

	return ca.currentAgent, nil
}
