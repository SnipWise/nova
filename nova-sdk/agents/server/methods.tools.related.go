package server

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
)

// SetToolsAgent sets the tools agent
func (agent *ServerAgent) SetToolsAgent(toolsAgent *tools.Agent) {
	agent.ToolsAgent = toolsAgent
}

// GetToolsAgent returns the tools agent
func (agent *ServerAgent) GetToolsAgent() *tools.Agent {
	return agent.ToolsAgent
}

// webConfirmationPrompt sends a confirmation prompt via web interface and waits for user response
func (agent *ServerAgent) webConfirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
	operationID := fmt.Sprintf("op_%p", &arguments)

	agent.Log.Info("🟡 Tool call detected: %s with args: %s (ID: %s)", functionName, arguments, operationID)

	// Create a response channel
	responseChan := make(chan tools.ConfirmationResponse)

	// Register the pending operation
	agent.OperationsMutex.Lock()
	agent.PendingOperations[operationID] = &PendingOperation{
		ID:           operationID,
		FunctionName: functionName,
		Arguments:    arguments,
		Response:     responseChan,
	}
	agent.OperationsMutex.Unlock()

	// Send notification via web interface
	message := fmt.Sprintf("Tool call detected: %s", functionName)
	agent.NotificationChanMutex.Lock()
	if agent.CurrentNotificationChan != nil {
		agent.CurrentNotificationChan <- ToolCallNotification{
			OperationID:  operationID,
			FunctionName: functionName,
			Arguments:    arguments,
			Message:      message,
		}
	}
	agent.NotificationChanMutex.Unlock()

	agent.Log.Info("⏳ Waiting for validation of operation %s", operationID)

	// Wait for user response
	response := <-responseChan

	agent.Log.Info("✅ Operation %s resolved with response: %v", operationID, response)

	return response
}

// executeFunction is a placeholder that should be overridden by the user
func (agent *ServerAgent) executeFunction(functionName string, arguments string) (string, error) {
	return fmt.Sprintf(`{"error": "executeFunction not implemented for %s"}`, functionName),
		fmt.Errorf("executeFunction not implemented")
}

// SetExecuteFunction allows the user to set a custom execute function
func (agent *ServerAgent) SetExecuteFunction(fn func(string, string) (string, error)) {
	agent.ExecuteFn = fn
}

// cliConfirmationPrompt is the default CLI confirmation prompt (auto-confirms)
func (agent *ServerAgent) cliConfirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
	agent.Log.Info("🟢 Auto-confirming tool call in CLI mode: %s", functionName)
	return tools.Confirmed
}

// SetConfirmationPromptFunction allows the user to set a custom confirmation prompt function
func (agent *ServerAgent) SetConfirmationPromptFunction(fn func(string, string) tools.ConfirmationResponse) {
	agent.ConfirmationPromptFn = fn
}
