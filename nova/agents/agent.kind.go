package agents

// AgentKind represents the type of agent
type Kind string

const (
	Basic  Kind = "Basic"
	Chat  Kind = "Chat"
	ChatServer Kind = "ChatServer"
	Remote Kind = "Remote"
	Tool  Kind = "Tools"
	Intent Kind = "Intent"
	Rag    Kind = "Rag"
	Compressor Kind = "Compressor"
	Structured Kind = "Structured"
	Macro Kind = "Macro"
)

