package crewserver

// CompressChatAgentContextIfOverLimit compresses the chat agent context if it exceeds the size limit.
func (agent *CrewServerAgent) CompressChatAgentContextIfOverLimit() (int, error) {
	return agent.BaseServerAgent.CompressChatAgentContextIfOverLimit(agent.currentChatAgent)
}

// CompressChatAgentContext compresses the chat agent context unconditionally.
func (agent *CrewServerAgent) CompressChatAgentContext() (int, error) {
	return agent.BaseServerAgent.CompressChatAgentContext(agent.currentChatAgent)
}
