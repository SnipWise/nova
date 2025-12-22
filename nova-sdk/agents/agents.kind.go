package agents

// AgentKind represents the type of agent
type Kind string

const (
	Basic  Kind = "Basic"
	Chat  Kind = "Chat"
	ChatServer Kind = "ChatServer"
	Remote Kind = "Remote"
	Tools  Kind = "Tools"
	Orchestrator Kind = "Orchestrator"
	Rag    Kind = "Rag"
	Compressor Kind = "Compressor"
	Structured Kind = "Structured"
	Macro Kind = "Macro"
	Composite Kind = "Composite"
)

