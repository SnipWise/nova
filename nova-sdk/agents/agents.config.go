package agents

// Config represents the core configuration parameters for creating an agent
type Config struct {
	// Name is the identifier for the agent
	Name string

	// Description provides details about the agent's purpose
	Description string

	// SystemInstructions defines the agent's behavior and role
	SystemInstructions string

	// EngineURL is the base URL for the model inference engine
	EngineURL string

	// APIKey
	APIKey string

	KeepConversationHistory bool
}
