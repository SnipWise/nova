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
	if agent.compressorAgent != nil {
		newSize, err := agent.CompressChatAgentContextIfOverLimit()
		if err != nil {
			agent.log.Error("Error during context compression: %v", err)
		} else if newSize > 0 {
			agent.log.Info("🗜️  Chat agent context compressed to %d bytes", newSize)
		}
	}

	// ------------------------------------------------------------
	// NOTE: Tool calls detection and execution if toolsAgent is set
	// ------------------------------------------------------------
	// Tool calls detection and execution if toolsAgent is set
	//toolExecution := false
	if agent.toolsAgent != nil {

		agent.toolsAgent.ResetMessages()

		historyMessagesForToolsAgent := agent.currentChatAgent.GetMessages()
		historyMessagesForToolsAgent = append(historyMessagesForToolsAgent, messages.Message{
			Role:    roles.User,
			Content: question,
		})

		toolCallsResult, err := agent.toolsAgent.DetectToolCallsLoopWithConfirmation(
			// []messages.Message{
			// 	{Role: roles.User, Content: question},
			// },
			historyMessagesForToolsAgent,
			agent.executeFn,
			agent.confirmationPromptFn,
		)
		if err != nil {
			return nil, err
		}

		finishReasonOfExecution := agent.toolsAgent.GetLastStateToolCalls().ExecutionResult.ExecFinishReason

		if finishReasonOfExecution == "" {
			agent.log.Info("1️⃣ finishReasonOfExecution: %s", "empty")
		} else {
			agent.log.Info("1️⃣ finishReasonOfExecution: %s", finishReasonOfExecution)
		}

		// Stream the final response after tool calls

		// Add tool results to chat agent context
		if len(toolCallsResult.Results) > 0 && finishReasonOfExecution == "function_executed" {

			agent.log.Info("✅ Tool calls executed successfully.")
			agent.log.Info("📝 Tool calls results: %s", toolCallsResult.Results)
			agent.log.Info("😁 Last assistant message: %s", toolCallsResult.LastAssistantMessage)

			agent.currentChatAgent.AddMessage(roles.System, toolCallsResult.LastAssistantMessage)

			callback(toolCallsResult.LastAssistantMessage, "tool_calls_completed")

			//agent.toolsAgent.ResetMessages()
		}
	} else {
		// Do nothing
	}

	lastExecConfirmation := agent.toolsAgent.GetLastStateToolCalls().Confirmation
	lastExecFinishReason := agent.toolsAgent.GetLastStateToolCalls().ExecutionResult.ExecFinishReason

	agent.log.Info("2️⃣ lastExecConfirmation: %v", lastExecConfirmation)
	agent.log.Info("3️⃣ lastExecFinishReason: %v", lastExecFinishReason)

	// TODO: check about lastExecConfirmation value == 0???
	if (lastExecConfirmation == 0) &&
		(lastExecFinishReason == "user_quit" ||
			lastExecFinishReason == "user_denied" ||
			lastExecFinishReason == "") {

		// IMPORTANT: only generate completion if no tool execution was done

		agent.log.Info("No tool execution was performed.")

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
				if relevantContext != "" {
					agent.currentChatAgent.AddMessage(
						roles.System,
						"Relevant information to help you answer the question:\n"+relevantContext,
					)
				}
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
				agent.selectedAgentId = detectedAgentId
			}
		}

		// ------------------------------------------------------------
		// NOTE: Generate streaming completion
		// ------------------------------------------------------------

		agent.log.Info("🚀 Generating streaming completion for question: %s", question)
		//stopped := false
		completionResult, errCompletion := agent.currentChatAgent.GenerateStreamCompletion(
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
	// reset last tool calls state
	agent.toolsAgent.ResetLastStateToolCalls()
	agent.toolsAgent.ResetMessages()

	return &chat.CompletionResult{}, nil

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
