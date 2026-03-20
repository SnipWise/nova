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
