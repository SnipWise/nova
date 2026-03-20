package serverbase

import "github.com/snipwise/nova/nova-sdk/agents/rag"

// SetRagAgent sets the RAG agent
func (agent *BaseServerAgent) SetRagAgent(ragAgent *rag.Agent) {
	agent.RagAgent = ragAgent
}

// GetRagAgent returns the RAG agent
func (agent *BaseServerAgent) GetRagAgent() *rag.Agent {
	return agent.RagAgent
}

// SetSimilarityLimit sets the similarity limit for document retrieval
func (agent *BaseServerAgent) SetSimilarityLimit(limit float64) {
	agent.SimilarityLimit = limit
}

// GetSimilarityLimit returns the similarity limit for document retrieval
func (agent *BaseServerAgent) GetSimilarityLimit() float64 {
	return agent.SimilarityLimit
}

// SetMaxSimilarities sets the maximum number of similar documents to retrieve
func (agent *BaseServerAgent) SetMaxSimilarities(n int) {
	agent.MaxSimilarities = n
}

// GetMaxSimilarities returns the maximum number of similar documents to retrieve
func (agent *BaseServerAgent) GetMaxSimilarities() int {
	return agent.MaxSimilarities
}
