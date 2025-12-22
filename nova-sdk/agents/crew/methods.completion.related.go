package crew

import (
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// StreamCompletion

func (agent *CrewAgent) StreamCompletion(
	question string,
	callback chat.StreamCallback,
) (*chat.CompletionResult, error) {

	// ------------------------------------------------------------
	// NOTE: Context packing
	// ------------------------------------------------------------
	newSize, err := agent.CompressChatAgentContextIfOverLimit()
	if err != nil {
		agent.log.Error("Error during context compression: %v", err)
	} else if newSize > 0 {
		agent.log.Info("🗜️  Chat agent context compressed to %d bytes", newSize)
	}

	// ------------------------------------------------------------
	// NOTE: Tool calls detection and execution if toolsAgent is set
	// ------------------------------------------------------------
	// Tool calls detection and execution if toolsAgent is set
	if agent.toolsAgent != nil {
		toolCallsResult, err := agent.toolsAgent.DetectParallelToolCallsWithConfirmation(
			[]messages.Message{
				{Role: roles.User, Content: question},
			},
			agent.executeFn,
			agent.confirmationPromptFn,
		)
		if err != nil {
			return nil, err
		}

		// Stream the final response after tool calls

		// Add tool results to chat agent context
		if len(toolCallsResult.Results) > 0 {
			agent.currentChatAgent.AddMessage(roles.System, toolCallsResult.LastAssistantMessage)
			agent.toolsAgent.ResetMessages()
		}
	} else {
		// Do nothing
	}

	// ------------------------------------------------------------
	// NOTE: Similarity search and add to context if RAG agent is set
	// ------------------------------------------------------------
	if agent.ragAgent != nil {
		relevantContext := ""
		similarities, err := agent.ragAgent.SearchTopN(question, agent.similarityLimit, agent.maxSimilarities)
		if err == nil && len(similarities) > 0 {
			for _, sim := range similarities {
				agent.log.Debug("Adding relevant context with similarity: %s", sim.Prompt)
				relevantContext += sim.Prompt + "\n---\n"
			}
			agent.log.Info("Added %d similar contexts from RAG agent", len(similarities))
			agent.currentChatAgent.AddMessage(
				roles.System,
				"Relevant information to help you answer the question:\n"+relevantContext,
			)
		} else {
			if err != nil {
				agent.log.Error("Error during similarity search: %v", err)
			} else {
				agent.log.Info("No relevant contexts found for the query")
			}
		}

	}

	// ------------------------------------------------------------
	// NOTE: Detect if we need to select another agent based on topic
	// ------------------------------------------------------------
	if agent.orchestratorAgent != nil {
		detectedAgentId, err := agent.DetectTopicThenGetAgentId(question)
		if err != nil {
			agent.log.Error("Error during topic detection: %v", err)
		} else if detectedAgentId != "" && agent.chatAgents[detectedAgentId] != agent.currentChatAgent {
			agent.log.Info("💡 Switching to detected agent ID: %s", detectedAgentId)
			agent.currentChatAgent = agent.chatAgents[detectedAgentId]
		}
	}

	// ------------------------------------------------------------
	// NOTE: Generate streaming completion
	// ------------------------------------------------------------

	agent.log.Info("🚀 Generating streaming completion for question: %s", question)
	//stopped := false
	completionResult , errCompletion := agent.currentChatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: question},
		},
		callback,
	)

	if errCompletion != nil {
		agent.log.Error("Error during streaming completion: %v", errCompletion)
		return nil, errCompletion
	}

	return completionResult, nil
}

/*
// GenerateStreamCompletion sends messages and streams the response via callback
func (agent *CrewAgent) GenerateStreamCompletion(
	userMessages []messages.Message,
	callback chat.StreamCallback,
) (*chat.CompletionResult, error) {
	return agent.currentChatAgent.GenerateStreamCompletion(userMessages, callback)
}

// GenerateStreamCompletionWithReasoning sends messages and streams both reasoning and response
func (agent *CrewAgent) GenerateStreamCompletionWithReasoning(
	userMessages []messages.Message,
	reasoningCallback chat.StreamCallback,
	responseCallback chat.StreamCallback,
) (*chat.ReasoningResult, error) {
	return agent.currentChatAgent.GenerateStreamCompletionWithReasoning(userMessages, reasoningCallback, responseCallback)
}

*/
