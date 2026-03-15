package serverbase

import (
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// LogToolExecutionStatus logs the finish reason of tool execution.
func LogToolExecutionStatus(log logger.Logger, finishReason string) {
	if finishReason == "" {
		log.Info("1️⃣ finishReasonOfExecution: %s", "empty")
	} else {
		log.Info("1️⃣ finishReasonOfExecution: %s", finishReason)
	}
}

// ToolsExecutedSuccessfully returns true if tools were executed and completed.
func ToolsExecutedSuccessfully(result *tools.ToolCallResult, finishReason string) bool {
	return len(result.Results) > 0 && finishReason == "function_executed"
}

// ShouldGenerateCompletion determines whether to generate a completion based on tool execution state.
// Returns true if no tools agent is configured, or if tool execution did not happen / was denied / quit.
func ShouldGenerateCompletion(log logger.Logger, toolsAgent *tools.Agent) bool {
	if toolsAgent == nil {
		return true
	}
	state := toolsAgent.GetLastStateToolCalls()
	confirmation := state.Confirmation
	finishReason := state.ExecutionResult.ExecFinishReason
	log.Info("2️⃣ lastExecConfirmation: %v", confirmation)
	log.Info("3️⃣ lastExecFinishReason: %v", finishReason)
	return confirmation == 0 &&
		(finishReason == "user_quit" ||
			finishReason == "user_denied" ||
			finishReason == "")
}

// AddRAGContextToChat performs a similarity search and injects matching context into the chat agent.
func AddRAGContextToChat(log logger.Logger, ragAgent *rag.Agent, chatAgent ChatAgent, question string, similarityLimit float64, maxSimilarities int) {
	if ragAgent == nil {
		return
	}
	similarities, err := ragAgent.SearchTopN(question, similarityLimit, maxSimilarities)
	if err != nil {
		log.Error("Error during similarity search: %v", err)
		return
	}
	if len(similarities) == 0 {
		log.Info("No relevant contexts found for the query")
		return
	}
	relevantContext := ""
	for _, sim := range similarities {
		log.Debug("Adding relevant context with similarity: %s", sim.Prompt)
		relevantContext += sim.Prompt + "\n---\n"
	}
	log.Info("Added %d similar contexts from RAG agent", len(similarities))
	chatAgent.AddMessage(roles.System, "Relevant information to help you answer the question:\n"+relevantContext)
}

// AddToolResultsToChat injects the tool execution results into the chat agent context.
func AddToolResultsToChat(log logger.Logger, chatAgent ChatAgent, result *tools.ToolCallResult, callback chat.StreamCallback) {
	log.Info("✅ Tool calls executed successfully.")
	log.Info("📝 Tool calls results: %s", result.Results)
	log.Info("😁 Last assistant message: %s", result.LastAssistantMessage)
	chatAgent.AddMessage(roles.System, result.LastAssistantMessage)
	callback(result.LastAssistantMessage, "tool_calls_completed")
}

// BuildToolCallHistory returns the current chat history appended with the new user question.
func BuildToolCallHistory(chatAgent ChatAgent, question string) []messages.Message {
	return append(chatAgent.GetMessages(), messages.Message{
		Role:    roles.User,
		Content: question,
	})
}

// StreamChatResponse generates the final streaming completion for a question.
func StreamChatResponse(log logger.Logger, chatAgent ChatAgent, question string, callback chat.StreamCallback) (*chat.CompletionResult, error) {
	log.Info("🚀 Generating streaming completion for question: %s", question)
	completionResult, err := chatAgent.GenerateStreamCompletion(
		[]messages.Message{{Role: roles.User, Content: question}},
		callback,
	)
	if err != nil {
		log.Error("Error during streaming completion: %v", err)
		return nil, err
	}
	return completionResult, nil
}
