package server

import (
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/serverbase"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
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

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Step 1: Compress context if over limit
	agent.compressContextIfNeededCLI()

	// Step 1.5: Execute tasks plan if tasksAgent is configured
	if planExecuted, err := agent.executePlanCLI(question, callback); err != nil {
		return nil, err
	} else if planExecuted {
		// Call after completion hook if set
		if agent.afterCompletion != nil {
			agent.afterCompletion(agent)
		}
		return &chat.CompletionResult{}, nil
	}

	// Step 2: Handle tool calls if toolsAgent is configured
	if err := agent.handleToolCallsCLI(question, callback); err != nil {
		return nil, err
	}

	// Step 3: Generate completion only if tools weren't executed or user denied/quit
	if serverbase.ShouldGenerateCompletion(agent.Log, agent.ToolsAgent) {
		result, err := agent.generateCompletionCLI(question, callback)

		// Call after completion hook if set
		if agent.afterCompletion != nil {
			agent.afterCompletion(agent)
		}

		return result, err
	}

	// Clean up after tool execution
	agent.cleanupToolStateCLI()

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

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

	historyMessages := serverbase.BuildToolCallHistory(agent.ChatAgent, question)

	var toolCallsResult *tools.ToolCallResult
	var err error

	modelConfig := agent.ToolsAgent.GetModelConfig()
	isParallel := modelConfig.ParallelToolCalls != nil && *modelConfig.ParallelToolCalls

	if isParallel {
		if agent.ConfirmationPromptFn != nil {
			agent.Log.Info("🔄 Using DetectParallelToolCallsWithConfirmation (CLI)")
			toolCallsResult, err = agent.ToolsAgent.DetectParallelToolCallsWithConfirmation(
				historyMessages,
				agent.ExecuteFn,
				agent.ConfirmationPromptFn,
			)
		} else {
			agent.Log.Info("🔄 Using DetectParallelToolCalls (CLI)")
			toolCallsResult, err = agent.ToolsAgent.DetectParallelToolCalls(
				historyMessages,
				agent.ExecuteFn,
			)
		}
	} else {
		agent.Log.Info("🔄 Using DetectToolCallsLoopWithConfirmation (CLI)")
		toolCallsResult, err = agent.ToolsAgent.DetectToolCallsLoopWithConfirmation(
			historyMessages,
			agent.ExecuteFn,
			agent.ConfirmationPromptFn,
		)
	}

	if err != nil {
		return err
	}

	finishReason := agent.ToolsAgent.GetLastStateToolCalls().ExecutionResult.ExecFinishReason
	serverbase.LogToolExecutionStatus(agent.Log, finishReason)

	if serverbase.ToolsExecutedSuccessfully(toolCallsResult, finishReason) {
		serverbase.AddToolResultsToChat(agent.Log, agent.ChatAgent, toolCallsResult, callback)
	}

	return nil
}

// generateCompletionCLI generates the final streaming completion with RAG support
func (agent *ServerAgent) generateCompletionCLI(question string, callback chat.StreamCallback) (*chat.CompletionResult, error) {
	agent.Log.Info("No tool execution was performed.")

	serverbase.AddRAGContextToChat(agent.Log, agent.RagAgent, agent.ChatAgent, question, agent.SimilarityLimit, agent.MaxSimilarities)

	return serverbase.StreamChatResponse(agent.Log, agent.ChatAgent, question, callback)
}

// cleanupToolStateCLI resets tool agent state after completion
func (agent *ServerAgent) cleanupToolStateCLI() {
	if agent.ToolsAgent != nil {
		agent.ToolsAgent.ResetLastStateToolCalls()
		agent.ToolsAgent.ResetMessages()
	}
}
