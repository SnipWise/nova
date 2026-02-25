package server

import (
	"github.com/snipwise/nova/nova-sdk/agents/tasks"
)

// SetTasksAgent sets the tasks agent
func (agent *ServerAgent) SetTasksAgent(tasksAgent *tasks.Agent) {
	agent.tasksAgentConfig = tasksAgent
}

// GetTasksAgent returns the tasks agent
func (agent *ServerAgent) GetTasksAgent() *tasks.Agent {
	return agent.tasksAgentConfig
}
