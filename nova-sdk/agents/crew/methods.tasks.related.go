package crew

import (
	"github.com/snipwise/nova/nova-sdk/agents/tasks"
)

// SetTasksAgent sets the tasks agent
func (agent *CrewAgent) SetTasksAgent(tasksAgent *tasks.Agent) {
	agent.tasksAgent = tasksAgent
}

// GetTasksAgent returns the tasks agent
func (agent *CrewAgent) GetTasksAgent() *tasks.Agent {
	return agent.tasksAgent
}
