package gatewayserver

import (
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// SetRagAgent sets the RAG agent.
func (agent *GatewayServerAgent) SetRagAgent(ragAgent *rag.Agent) {
	agent.ragAgent = ragAgent
}

// GetRagAgent returns the RAG agent.
func (agent *GatewayServerAgent) GetRagAgent() *rag.Agent {
	return agent.ragAgent
}

// SetSimilarityLimit sets the similarity limit for document retrieval.
func (agent *GatewayServerAgent) SetSimilarityLimit(limit float64) {
	agent.similarityLimit = limit
}

// GetSimilarityLimit returns the similarity limit for document retrieval.
func (agent *GatewayServerAgent) GetSimilarityLimit() float64 {
	return agent.similarityLimit
}

// SetMaxSimilarities sets the maximum number of similar documents to retrieve.
func (agent *GatewayServerAgent) SetMaxSimilarities(n int) {
	agent.maxSimilarities = n
}

// GetMaxSimilarities returns the maximum number of similar documents to retrieve.
func (agent *GatewayServerAgent) GetMaxSimilarities() int {
	return agent.maxSimilarities
}

// addRAGContext performs similarity search and adds relevant context to the chat agent.
func (agent *GatewayServerAgent) addRAGContext(question string) {
	if agent.ragAgent == nil {
		return
	}

	similarities, err := agent.ragAgent.SearchTopN(question, agent.similarityLimit, agent.maxSimilarities)
	if err != nil {
		agent.log.Error("Error during similarity search: %v", err)
		return
	}

	if len(similarities) == 0 {
		return
	}

	relevantContext := ""
	for _, sim := range similarities {
		relevantContext += sim.Prompt + "\n---\n"
	}

	agent.log.Info("Added %d similar contexts from RAG agent", len(similarities))
	agent.currentChatAgent.AddMessage(
		roles.System,
		"Relevant information to help you answer the question:\n"+relevantContext,
	)
}
