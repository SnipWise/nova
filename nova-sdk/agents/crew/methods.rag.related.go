package crew

import "github.com/snipwise/nova/nova-sdk/agents/rag"

// SetRagAgent sets the RAG agent
func (agent *CrewAgent) SetRagAgent(ragAgent *rag.Agent) {
	agent.ragAgent = ragAgent
}

// GetRagAgent returns the RAG agent
func (agent *CrewAgent) GetRagAgent() *rag.Agent {
	return agent.ragAgent
}

// SetSimilarityLimit sets the similarity limit for document retrieval
func (agent *CrewAgent) SetSimilarityLimit(limit float64) {
	agent.similarityLimit = limit
}

// GetSimilarityLimit returns the similarity limit for document retrieval
func (agent *CrewAgent) GetSimilarityLimit() float64 {
	return agent.similarityLimit
}

// SetMaxSimilarities sets the maximum number of similar documents to retrieve
func (agent *CrewAgent) SetMaxSimilarities(n int) {
	agent.maxSimilarities = n
}

// GetMaxSimilarities returns the maximum number of similar documents to retrieve
func (agent *CrewAgent) GetMaxSimilarities() int {
	return agent.maxSimilarities
}
