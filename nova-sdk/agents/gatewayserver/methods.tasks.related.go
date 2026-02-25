package gatewayserver

import (
	"github.com/snipwise/nova/nova-sdk/agents/tasks"
)

// SetTasksAgent sets the tasks agent
func (agent *GatewayServerAgent) SetTasksAgent(tasksAgent *tasks.Agent) {
	agent.tasksAgent = tasksAgent
}

// GetTasksAgent returns the tasks agent
func (agent *GatewayServerAgent) GetTasksAgent() *tasks.Agent {
	return agent.tasksAgent
}
