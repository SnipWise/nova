package crew

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
)

// SetToolsAgent sets the tools agent
func (agent *CrewAgent) SetToolsAgent(toolsAgent *tools.Agent) {
	agent.toolsAgent = toolsAgent
}

// GetToolsAgent returns the tools agent
func (agent *CrewAgent) GetToolsAgent() *tools.Agent {
	return agent.toolsAgent
}

// executeFunction is a placeholder that should be overridden by the user
func (agent *CrewAgent) executeFunction(functionName string, arguments string) (string, error) {
	return fmt.Sprintf(`{"error": "executeFunction not implemented for %s"}`, functionName),
		fmt.Errorf("executeFunction not implemented")
}

// SetExecuteFunction allows the user to set a custom execute function
func (agent *CrewAgent) SetExecuteFunction(fn func(string, string) (string, error)) {
	agent.executeFn = fn
}

// confirmationPrompt is a placeholder that should be overridden by the user
func (agent *CrewAgent) confirmationPrompt(functionName string, arguments string) tools.ConfirmationResponse {
	return 0
}

// SetConfirmationPromptFunction allows the user to set a custom confirmation prompt function
func (agent *CrewAgent) SetConfirmationPromptFunction(fn func(string, string) tools.ConfirmationResponse) {
	agent.confirmationPromptFn = fn
}

