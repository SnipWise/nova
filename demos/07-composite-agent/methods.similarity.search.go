package main

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// SearchSimilarities searches for similar documents using the RAG agent.
func (ca *CompositeAgent) SearchSimilarities(query string, limit float64, n int) (string, error) {
	similarities, err := ca.ragAgent.SearchTopN(query, limit, n)
	if err != nil {
		return "", err
	}
	relevantContext := ""
	if len(similarities) > 0 {
		for _, sim := range similarities {
			relevantContext += sim.Prompt + "\n---\n"
		}
	}
	return relevantContext, nil
}

// SearchSimilaritiesAndAddToCurrentAgentContext searches for similar documents and adds them to the current agent's context.
func (ca *CompositeAgent) SearchSimilaritiesAndAddToCurrentAgentContext(query string, limit float64, n int) (string, error) {
	similarities, err := ca.ragAgent.SearchTopN(query, limit, n)
	if err != nil {
		return "", err
	}
	relevantContext := ""
	if len(similarities) > 0 {
		for _, sim := range similarities {
			relevantContext += sim.Prompt + "\n---\n"
		}

		fmt.Println("🗂️  Adding relevant context to current agent:")
		fmt.Println(relevantContext)

		ca.currentAgent.AddMessage(
			roles.System,
			"Relevant information to help you answer the question:\n"+relevantContext,
		)
	}
	return relevantContext, nil
}
