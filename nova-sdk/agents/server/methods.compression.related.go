package server

// CompressChatAgentContextIfOverLimit compresses the chat agent context if it exceeds the size limit.
func (agent *ServerAgent) CompressChatAgentContextIfOverLimit() (int, error) {
	return agent.BaseServerAgent.CompressChatAgentContextIfOverLimit(agent.chatAgent)
}

// CompressChatAgentContext compresses the chat agent context unconditionally.
func (agent *ServerAgent) CompressChatAgentContext() (int, error) {
	return agent.BaseServerAgent.CompressChatAgentContext(agent.chatAgent)
}
