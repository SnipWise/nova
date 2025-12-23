package server

import "github.com/snipwise/nova/nova-sdk/agents/rag"

// SetRagAgent sets the RAG agent
func (agent *ServerAgent) SetRagAgent(ragAgent *rag.Agent) {
	agent.RagAgent = ragAgent
}

// GetRagAgent returns the RAG agent
func (agent *ServerAgent) GetRagAgent() *rag.Agent {
	return agent.RagAgent
}

// SetSimilarityLimit sets the similarity limit for document retrieval
func (agent *ServerAgent) SetSimilarityLimit(limit float64) {
	agent.SimilarityLimit = limit
}

// GetSimilarityLimit returns the similarity limit for document retrieval
func (agent *ServerAgent) GetSimilarityLimit() float64 {
	return agent.SimilarityLimit
}

// SetMaxSimilarities sets the maximum number of similar documents to retrieve
func (agent *ServerAgent) SetMaxSimilarities(n int) {
	agent.MaxSimilarities = n
}

// GetMaxSimilarities returns the maximum number of similar documents to retrieve
func (agent *ServerAgent) GetMaxSimilarities() int {
	return agent.MaxSimilarities
}
