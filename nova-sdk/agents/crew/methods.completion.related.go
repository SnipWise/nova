package crew

import (
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

// StreamCompletion processes a question through the crew pipeline:
// 1. Compresses context if needed
// 2. Executes tool calls if detected
// 3. Adds RAG context if available
// 4. Routes to appropriate agent if orchestrator is configured
// 5. Generates streaming completion
func (agent *CrewAgent) StreamCompletion(
	question string,
	callback chat.StreamCallback,
) (*chat.CompletionResult, error) {

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Step 1: Compress context if over limit
	agent.compressContextIfNeeded()

	// Step 2: Handle tool calls if toolsAgent is configured
	if err := agent.handleToolCalls(question, callback); err != nil {
		return nil, err
	}

	// Step 3: Generate completion only if tools weren't executed or user denied/quit
	if agent.shouldGenerateCompletion() {
		result, err := agent.generateCompletion(question, callback)

		// Call after completion hook if set
		if agent.afterCompletion != nil {
			agent.afterCompletion(agent)
		}

		return result, err
	}

	// Clean up after tool execution
	agent.cleanupToolState()

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return &chat.CompletionResult{}, nil
}

// compressContextIfNeeded compresses the chat context if compressor is configured and limit exceeded
func (agent *CrewAgent) compressContextIfNeeded() {
	if agent.compressorAgent == nil {
		return
	}

	newSize, err := agent.CompressChatAgentContextIfOverLimit()
	if err != nil {
		agent.log.Error("Error during context compression: %v", err)
		return
	}

	if newSize > 0 {
		agent.log.Info("🗜️  Chat agent context compressed to %d bytes", newSize)
	}
}

// handleToolCalls detects and executes tool calls if toolsAgent is configured
func (agent *CrewAgent) handleToolCalls(question string, callback chat.StreamCallback) error {
	if agent.toolsAgent == nil {
		return nil
	}

	agent.toolsAgent.ResetMessages()

	// Prepare message history including current question
	historyMessages := agent.buildToolCallHistory(question)

	// Detect and execute tool calls
	toolCallsResult, err := agent.toolsAgent.DetectToolCallsLoopWithConfirmation(
		historyMessages,
		agent.executeFn,
		agent.confirmationPromptFn,
	)
	if err != nil {
		return err
	}

	// Process tool execution results
	finishReason := agent.toolsAgent.GetLastStateToolCalls().ExecutionResult.ExecFinishReason
	agent.logToolExecutionStatus(finishReason)

	// Add tool results to chat context if execution succeeded
	if agent.toolsExecutedSuccessfully(toolCallsResult, finishReason) {
		agent.addToolResultsToContext(toolCallsResult, callback)
	}

	return nil
}

// buildToolCallHistory creates message history for tool detection
func (agent *CrewAgent) buildToolCallHistory(question string) []messages.Message {
	history := agent.currentChatAgent.GetMessages()
	return append(history, messages.Message{
		Role:    roles.User,
		Content: question,
	})
}

// logToolExecutionStatus logs the finish reason of tool execution
func (agent *CrewAgent) logToolExecutionStatus(finishReason string) {
	if finishReason == "" {
		agent.log.Info("1️⃣ finishReasonOfExecution: %s", "empty")
	} else {
		agent.log.Info("1️⃣ finishReasonOfExecution: %s", finishReason)
	}
}

// toolsExecutedSuccessfully checks if tools were executed successfully
func (agent *CrewAgent) toolsExecutedSuccessfully(result *tools.ToolCallResult, finishReason string) bool {
	return len(result.Results) > 0 && finishReason == "function_executed"
}

// addToolResultsToContext adds tool execution results to chat context
func (agent *CrewAgent) addToolResultsToContext(result *tools.ToolCallResult, callback chat.StreamCallback) {
	agent.log.Info("✅ Tool calls executed successfully.")
	agent.log.Info("📝 Tool calls results: %s", result.Results)
	agent.log.Info("😁 Last assistant message: %s", result.LastAssistantMessage)

	agent.currentChatAgent.AddMessage(roles.System, result.LastAssistantMessage)
	callback(result.LastAssistantMessage, "tool_calls_completed")
}

// shouldGenerateCompletion determines if we should generate a completion based on tool execution state
func (agent *CrewAgent) shouldGenerateCompletion() bool {
	// Always generate if no tools agent configured
	if agent.toolsAgent == nil {
		return true
	}

	state := agent.toolsAgent.GetLastStateToolCalls()
	confirmation := state.Confirmation
	finishReason := state.ExecutionResult.ExecFinishReason

	agent.log.Info("2️⃣ lastExecConfirmation: %v", confirmation)
	agent.log.Info("3️⃣ lastExecFinishReason: %v", finishReason)

	// Generate completion if:
	// - No confirmation needed AND
	// - User quit, denied, or no tool execution occurred
	return confirmation == 0 &&
		(finishReason == "user_quit" ||
		 finishReason == "user_denied" ||
		 finishReason == "")
}

// generateCompletion generates the final streaming completion with RAG and orchestrator support
func (agent *CrewAgent) generateCompletion(question string, callback chat.StreamCallback) (*chat.CompletionResult, error) {
	agent.log.Info("No tool execution was performed.")

	// Add RAG context if available
	agent.addRAGContext(question)

	// Switch agent based on topic if orchestrator is configured
	agent.routeToAppropriateAgent(question)

	// Generate streaming completion
	return agent.streamResponse(question, callback)
}

// addRAGContext performs similarity search and adds relevant context
func (agent *CrewAgent) addRAGContext(question string) {
	if agent.ragAgent == nil {
		return
	}

	similarities, err := agent.ragAgent.SearchTopN(question, agent.similarityLimit, agent.maxSimilarities)
	if err != nil {
		agent.log.Error("Error during similarity search: %v", err)
		return
	}

	if len(similarities) == 0 {
		agent.log.Info("No relevant contexts found for the query")
		return
	}

	// Build context from similarities
	relevantContext := ""
	for _, sim := range similarities {
		agent.log.Debug("Adding relevant context with similarity: %s", sim.Prompt)
		relevantContext += sim.Prompt + "\n---\n"
	}

	agent.log.Info("Added %d similar contexts from RAG agent", len(similarities))
	agent.currentChatAgent.AddMessage(
		roles.System,
		"Relevant information to help you answer the question:\n"+relevantContext,
	)
}

// routeToAppropriateAgent detects topic and switches to appropriate agent
func (agent *CrewAgent) routeToAppropriateAgent(question string) {
	if agent.orchestratorAgent == nil {
		return
	}

	detectedAgentId, err := agent.DetectTopicThenGetAgentId(question)
	if err != nil {
		agent.log.Error("Error during topic detection: %v", err)
		return
	}

	// Switch agent if different from current
	if detectedAgentId != "" && agent.chatAgents[detectedAgentId] != agent.currentChatAgent {
		agent.log.Info("💡 Switching to detected agent ID: %s", detectedAgentId)
		agent.currentChatAgent = agent.chatAgents[detectedAgentId]
		agent.selectedAgentId = detectedAgentId
	}
}

// streamResponse generates the final streaming completion
func (agent *CrewAgent) streamResponse(question string, callback chat.StreamCallback) (*chat.CompletionResult, error) {
	agent.log.Info("🚀 Generating streaming completion for question: %s", question)

	completionResult, err := agent.currentChatAgent.GenerateStreamCompletion(
		[]messages.Message{
			{Role: roles.User, Content: question},
		},
		callback,
	)

	if err != nil {
		agent.log.Error("Error during streaming completion: %v", err)
		return nil, err
	}

	return completionResult, nil
}

// cleanupToolState resets tool agent state after completion
func (agent *CrewAgent) cleanupToolState() {
	if agent.toolsAgent != nil {
		agent.toolsAgent.ResetLastStateToolCalls()
		agent.toolsAgent.ResetMessages()
	}
}
