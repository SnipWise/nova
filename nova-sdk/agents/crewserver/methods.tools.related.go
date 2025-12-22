package crewserver

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
)

// SetToolsAgent sets the tools agent
func (agent *CrewServerAgent) SetToolsAgent(toolsAgent *tools.Agent) {
	agent.toolsAgent = toolsAgent
}

// GetToolsAgent returns the tools agent
func (agent *CrewServerAgent) GetToolsAgent() *tools.Agent {
	return agent.toolsAgent
}

// webConfirmationPrompt sends a confirmation prompt via web interface and waits for user response
func (agent *CrewServerAgent) webConfirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
	operationID := fmt.Sprintf("op_%p", &arguments)

	agent.log.Info("🟡 Tool call detected: %s with args: %s (ID: %s)", functionName, arguments, operationID)

	// Create a response channel
	responseChan := make(chan tools.ConfirmationResponse)

	// Register the pending operation
	agent.operationsMutex.Lock()
	agent.pendingOperations[operationID] = &PendingOperation{
		ID:           operationID,
		FunctionName: functionName,
		Arguments:    arguments,
		Response:     responseChan,
	}
	agent.operationsMutex.Unlock()

	// Send notification via web interface
	message := fmt.Sprintf("Tool call detected: %s", functionName)
	agent.notificationChanMutex.Lock()
	if agent.currentNotificationChan != nil {
		agent.currentNotificationChan <- ToolCallNotification{
			OperationID:  operationID,
			FunctionName: functionName,
			Arguments:    arguments,
			Message:      message,
		}
	}
	agent.notificationChanMutex.Unlock()

	agent.log.Info("⏳ Waiting for validation of operation %s", operationID)

	// Wait for user response
	response := <-responseChan

	agent.log.Info("✅ Operation %s resolved with response: %v", operationID, response)

	return response
}

// executeFunction is a placeholder that should be overridden by the user
func (agent *CrewServerAgent) executeFunction(functionName string, arguments string) (string, error) {
	return fmt.Sprintf(`{"error": "executeFunction not implemented for %s"}`, functionName),
		fmt.Errorf("executeFunction not implemented")
}

// SetExecuteFunction allows the user to set a custom execute function
func (agent *CrewServerAgent) SetExecuteFunction(fn func(string, string) (string, error)) {
	agent.executeFn = fn
}
