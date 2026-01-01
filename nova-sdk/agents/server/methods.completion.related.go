package server

import (
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// StreamCompletion processes a question through the server agent pipeline:
// 1. Compresses context if needed
// 2. Executes tool calls if detected
// 3. Adds RAG context if available
// 4. Generates streaming completion
// This method mirrors the crew agent's StreamCompletion for CLI usage
func (agent *ServerAgent) StreamCompletion(
	question string,
	callback chat.StreamCallback,
) (*chat.CompletionResult, error) {

	// Step 1: Compress context if over limit
	agent.compressContextIfNeededCLI()

	// Step 2: Handle tool calls if toolsAgent is configured
	if err := agent.handleToolCallsCLI(question, callback); err != nil {
		return nil, err
	}

	// Step 3: Generate completion only if tools weren't executed or user denied/quit
	if agent.shouldGenerateCompletionCLI() {
		return agent.generateCompletionCLI(question, callback)
	}

	// Clean up after tool execution
	agent.cleanupToolStateCLI()
	return &chat.CompletionResult{}, nil
}

// compressContextIfNeededCLI compresses the chat context if compressor is configured and limit exceeded
func (agent *ServerAgent) compressContextIfNeededCLI() {
	if agent.CompressorAgent == nil {
		return
	}

	newSize, err := agent.CompressChatAgentContextIfOverLimit()
	if err != nil {
		agent.Log.Error("Error during context compression: %v", err)
		return
	}

	if newSize > 0 {
		agent.Log.Info("🗜️  Chat agent context compressed to %d bytes", newSize)
	}
}

// handleToolCallsCLI detects and executes tool calls if toolsAgent is configured
func (agent *ServerAgent) handleToolCallsCLI(question string, callback chat.StreamCallback) error {
	if agent.ToolsAgent == nil {
		return nil
	}

	agent.ToolsAgent.ResetMessages()

	// Prepare message history including current question
	historyMessages := agent.buildToolCallHistoryCLI(question)

	// Detect and execute tool calls
	toolCallsResult, err := agent.ToolsAgent.DetectToolCallsLoopWithConfirmation(
		historyMessages,
		agent.ExecuteFn,
		agent.ConfirmationPromptFn,
	)
	if err != nil {
		return err
	}

	// Process tool execution results
	finishReason := agent.ToolsAgent.GetLastStateToolCalls().ExecutionResult.ExecFinishReason
	agent.logToolExecutionStatusCLI(finishReason)

	// Add tool results to chat context if execution succeeded
	if agent.toolsExecutedSuccessfullyCLI(toolCallsResult, finishReason) {
		agent.addToolResultsToContextCLI(toolCallsResult, callback)
	}

	return nil
}

// buildToolCallHistoryCLI creates message history for tool detection
func (agent *ServerAgent) buildToolCallHistoryCLI(question string) []messages.Message {
	history := agent.chatAgent.GetMessages()
	return append(history, messages.Message{
		Role:    roles.User,
		Content: question,
	})
}

// logToolExecutionStatusCLI logs the finish reason of tool execution
func (agent *ServerAgent) logToolExecutionStatusCLI(finishReason string) {
	if finishReason == "" {
		agent.Log.Info("1️⃣ finishReasonOfExecution: %s", "empty")
	} else {
		agent.Log.Info("1️⃣ finishReasonOfExecution: %s", finishReason)
	}
}

// toolsExecutedSuccessfullyCLI checks if tools were executed successfully
func (agent *ServerAgent) toolsExecutedSuccessfullyCLI(result *tools.ToolCallResult, finishReason string) bool {
	return len(result.Results) > 0 && finishReason == "function_executed"
}

// addToolResultsToContextCLI adds tool execution results to chat context
func (agent *ServerAgent) addToolResultsToContextCLI(result *tools.ToolCallResult, callback chat.StreamCallback) {
	agent.Log.Info("✅ Tool calls executed successfully.")
	agent.Log.Info("📝 Tool calls results: %s", result.Results)
	agent.Log.Info("😁 Last assistant message: %s", result.LastAssistantMessage)

	agent.chatAgent.AddMessage(roles.System, result.LastAssistantMessage)
	callback(result.LastAssistantMessage, "tool_calls_completed")
}

// shouldGenerateCompletionCLI determines if we should generate a completion based on tool execution state
func (agent *ServerAgent) shouldGenerateCompletionCLI() bool {
	// Always generate if no tools agent configured
	if agent.ToolsAgent == nil {
		return true
	}

	state := agent.ToolsAgent.GetLastStateToolCalls()
	confirmation := state.Confirmation
	finishReason := state.ExecutionResult.ExecFinishReason

	agent.Log.Info("2️⃣ lastExecConfirmation: %v", confirmation)
	agent.Log.Info("3️⃣ lastExecFinishReason: %v", finishReason)

	// Generate completion if:
	// - No confirmation needed AND
	// - User quit, denied, or no tool execution occurred
	return confirmation == 0 &&
		(finishReason == "user_quit" ||
			finishReason == "user_denied" ||
			finishReason == "")
}

// generateCompletionCLI generates the final streaming completion with RAG support
func (agent *ServerAgent) generateCompletionCLI(question string, callback chat.StreamCallback) (*chat.CompletionResult, error) {
	agent.Log.Info("No tool execution was performed.")

	// Add RAG context if available
	agent.addRAGContextCLI(question)

	// Generate streaming completion
	return agent.streamResponseCLI(question, callback)
}

// addRAGContextCLI performs similarity search and adds relevant context
func (agent *ServerAgent) addRAGContextCLI(question string) {
	if agent.RagAgent == nil {
		return
	}

	similarities, err := agent.RagAgent.SearchTopN(question, agent.SimilarityLimit, agent.MaxSimilarities)
	if err != nil {
		agent.Log.Error("Error during similarity search: %v", err)
		return
	}

	if len(similarities) == 0 {
		agent.Log.Info("No relevant contexts found for the query")
		return
	}

	// Build context from similarities
	relevantContext := ""
	for _, sim := range similarities {
		agent.Log.Debug("Adding relevant context with similarity: %s", sim.Prompt)
		relevantContext += sim.Prompt + "\n---\n"
	}

	agent.Log.Info("Added %d similar contexts from RAG agent", len(similarities))
	agent.chatAgent.AddMessage(
		roles.System,
		"Relevant information to help you answer the question:\n"+relevantContext,
	)
}

// streamResponseCLI generates the final streaming completion
func (agent *ServerAgent) streamResponseCLI(question string, callback chat.StreamCallback) (*chat.CompletionResult, error) {
	agent.Log.Info("🚀 Generating streaming completion for question: %s", question)

	completionResult, err := agent.chatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: question},
		},
		callback,
	)

	if err != nil {
		agent.Log.Error("Error during streaming completion: %v", err)
		return nil, err
	}

	return completionResult, nil
}

// cleanupToolStateCLI resets tool agent state after completion
func (agent *ServerAgent) cleanupToolStateCLI() {
	if agent.ToolsAgent != nil {
		agent.ToolsAgent.ResetLastStateToolCalls()
		agent.ToolsAgent.ResetMessages()
	}
}
