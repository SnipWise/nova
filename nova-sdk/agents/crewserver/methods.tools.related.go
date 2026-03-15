package crewserver

import (
	"fmt"

	"github.com/snipwise/nova/nova-sdk/agents/tools"
)

// SetToolsAgent sets the tools agent
func (agent *CrewServerAgent) SetToolsAgent(toolsAgent *tools.Agent) {
	agent.ToolsAgent = toolsAgent
}

// GetToolsAgent returns the tools agent
func (agent *CrewServerAgent) GetToolsAgent() *tools.Agent {
	return agent.ToolsAgent
}

// executeFunction is a placeholder that should be overridden by the user
func (agent *CrewServerAgent) executeFunction(functionName string, arguments string) (string, error) {
	return fmt.Sprintf(`{"error": "executeFunction not implemented for %s"}`, functionName),
		fmt.Errorf("executeFunction not implemented")
}

// SetExecuteFunction allows the user to set a custom execute function
func (agent *CrewServerAgent) SetExecuteFunction(fn func(string, string) (string, error)) {
	agent.ExecuteFn = fn
}

// SetConfirmationPromptFn sets the confirmation prompt function for tool call confirmation
func (agent *CrewServerAgent) SetConfirmationPromptFn(fn func(string, string) tools.ConfirmationResponse) {
	agent.ConfirmationPromptFn = fn
}

// GetConfirmationPromptFn returns the confirmation prompt function
func (agent *CrewServerAgent) GetConfirmationPromptFn() func(string, string) tools.ConfirmationResponse {
	return agent.ConfirmationPromptFn
}
