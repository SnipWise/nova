package crew

import (
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
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

	// Step 1.5: Execute tasks plan if tasksAgent is configured
	if planExecuted, err := agent.executePlanCLI(question, callback); err != nil {
		return nil, err
	} else if planExecuted {
		if agent.afterCompletion != nil {
			agent.afterCompletion(agent)
		}
		return &chat.CompletionResult{}, nil
	}

	// Step 2: Handle tool calls if toolsAgent is configured
	if err := agent.handleToolCalls(question, callback); err != nil {
		return nil, err
	}

	// Step 3: Generate completion only if tools weren't executed or user denied/quit
	if serverbase.ShouldGenerateCompletion(agent.log, agent.toolsAgent) {
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

	historyMessages := serverbase.BuildToolCallHistory(agent.currentChatAgent, question)

	toolCallsResult, err := agent.toolsAgent.DetectToolCallsLoopWithConfirmation(
		historyMessages,
		agent.executeFn,
		agent.confirmationPromptFn,
	)
	if err != nil {
		return err
	}

	finishReason := agent.toolsAgent.GetLastStateToolCalls().ExecutionResult.ExecFinishReason
	serverbase.LogToolExecutionStatus(agent.log, finishReason)

	if serverbase.ToolsExecutedSuccessfully(toolCallsResult, finishReason) {
		serverbase.AddToolResultsToChat(agent.log, agent.currentChatAgent, toolCallsResult, callback)
	}

	return nil
}

// generateCompletion generates the final streaming completion with RAG and orchestrator support
func (agent *CrewAgent) generateCompletion(question string, callback chat.StreamCallback) (*chat.CompletionResult, error) {
	agent.log.Info("No tool execution was performed.")

	serverbase.AddRAGContextToChat(agent.log, agent.ragAgent, agent.currentChatAgent, question, agent.similarityLimit, agent.maxSimilarities)

	agent.routeToAppropriateAgent(question)

	return serverbase.StreamChatResponse(agent.log, agent.currentChatAgent, question, callback)
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

// cleanupToolState resets tool agent state after completion
func (agent *CrewAgent) cleanupToolState() {
	if agent.toolsAgent != nil {
		agent.toolsAgent.ResetLastStateToolCalls()
		agent.toolsAgent.ResetMessages()
	}
}
