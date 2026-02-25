package crewserver

import (
	"github.com/snipwise/nova/nova-sdk/agents/tasks"
)

// SetTasksAgent sets the tasks agent
func (agent *CrewServerAgent) SetTasksAgent(tasksAgent *tasks.Agent) {
	agent.tasksAgentConfig = tasksAgent
}

// GetTasksAgent returns the tasks agent
func (agent *CrewServerAgent) GetTasksAgent() *tasks.Agent {
	return agent.tasksAgentConfig
}
