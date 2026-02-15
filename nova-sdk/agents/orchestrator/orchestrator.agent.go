package orchestrator

import (
	"context"
	"errors"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// OrchestratorAgentOption is a functional option for configuring an Agent during creation
type OrchestratorAgentOption func(*Agent)

// BeforeCompletion sets a hook that is called before each intent identification
func BeforeCompletion(fn func(*Agent)) OrchestratorAgentOption {
	return func(a *Agent) {
		a.beforeCompletion = fn
	}
}

// AfterCompletion sets a hook that is called after each intent identification
func AfterCompletion(fn func(*Agent)) OrchestratorAgentOption {
	return func(a *Agent) {
		a.afterCompletion = fn
	}
}

// WithRoutingConfig sets the agent routing configuration
func WithRoutingConfig(config AgentRoutingConfig) OrchestratorAgentOption {
	return func(a *Agent) {
		a.agentRoutingConfig = &config
	}
}

// AgentRoutingConfig defines the routing configuration for the orchestrator
// It maps topics to specific agents and provides a default fallback
type AgentRoutingConfig struct {
	Routing []struct {
		Topics []string `json:"topics"`
		Agent  string   `json:"agent"`
	} `json:"routing"`
	DefaultAgent string `json:"default_agent"`
}

// Agent represents an orchestrator agent that identifies topics/intents from user input
// It's a specialized structured agent that uses agents.Intent as its output type
type Agent struct {
	config              agents.Config
	modelConfig         models.Config
	internalStructAgent *structured.Agent[agents.Intent]
	log                 logger.Logger
	agentRoutingConfig  *AgentRoutingConfig

	// Lifecycle hooks
	beforeCompletion func(*Agent)
	afterCompletion  func(*Agent)
}

// NewAgent creates a new orchestrator agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
	opts ...OrchestratorAgentOption,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	// Create internal structured agent with agents.Intent type
	structAgent, err := structured.NewAgent[agents.Intent](ctx, agentConfig, modelConfig)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		config:              agentConfig,
		modelConfig:         modelConfig,
		internalStructAgent: structAgent,
		log:                 log,
	}

	// Apply optional configurations
	for _, opt := range opts {
		opt(agent)
	}

	return agent, nil
}

// Kind returns the agent type
func (agent *Agent) Kind() agents.Kind {
	return agents.Orchestrator
}

func (agent *Agent) GetName() string {
	return agent.config.Name
}

func (agent *Agent) GetModelID() string {
	return agent.modelConfig.Name
}

// GetMessages returns all conversation messages
func (agent *Agent) GetMessages() []messages.Message {
	return agent.internalStructAgent.GetMessages()
}

func (agent *Agent) AddMessage(role roles.Role, content string) {
	agent.internalStructAgent.AddMessage(role, content)
}

// AddMessages adds multiple messages to the conversation history
func (agent *Agent) AddMessages(msgs []messages.Message) {
	agent.internalStructAgent.AddMessages(msgs)
}

// ResetMessages clears all messages except the system instruction
func (agent *Agent) ResetMessages() {
	agent.internalStructAgent.ResetMessages()
}

// IdentifyIntent sends messages and returns the identified intent
func (agent *Agent) IdentifyIntent(userMessages []messages.Message) (intent *agents.Intent, finishReason string, err error) {
	if len(userMessages) == 0 {
		return nil, "", errors.New("no messages provided")
	}

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Call internal structured agent to generate Intent
	intent, finishReason, err = agent.internalStructAgent.GenerateStructuredData(userMessages)
	if err != nil {
		return nil, finishReason, err
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return intent, finishReason, nil
}

// IdentifyTopicFromText is a convenience method that takes a text string and returns the topic
func (agent *Agent) IdentifyTopicFromText(text string) (string, error) {
	userMessages := []messages.Message{
		{
			Role:    roles.User,
			Content: text,
		},
	}

	intent, _, err := agent.IdentifyIntent(userMessages)
	if err != nil {
		return "", err
	}

	return intent.TopicDiscussion, nil
}

// === Config Getters and Setters ===

// GetConfig returns the agent configuration
func (agent *Agent) GetConfig() agents.Config {
	return agent.config
}

// SetConfig updates the agent configuration
func (agent *Agent) SetConfig(config agents.Config) {
	agent.config = config
	agent.internalStructAgent.SetConfig(config)
}

// GetModelConfig returns the model configuration
func (agent *Agent) GetModelConfig() models.Config {
	return agent.modelConfig
}

// SetModelConfig updates the model configuration
// Note: This updates the stored config but doesn't regenerate the internal OpenAI params
// For most parameters to take effect, create a new agent with the new config
func (agent *Agent) SetModelConfig(config models.Config) {
	agent.modelConfig = config
	agent.internalStructAgent.SetModelConfig(config)
}

// GetRoutingConfig returns the agent routing configuration
// Returns nil if no routing config is set
func (agent *Agent) GetRoutingConfig() *AgentRoutingConfig {
	return agent.agentRoutingConfig
}

// SetRoutingConfig updates the agent routing configuration
func (agent *Agent) SetRoutingConfig(config *AgentRoutingConfig) {
	agent.agentRoutingConfig = config
}

func (agent *Agent) GetLastRequestRawJSON() string {
	return agent.internalStructAgent.GetLastRequestRawJSON()
}
func (agent *Agent) GetLastResponseRawJSON() string {
	return agent.internalStructAgent.GetLastResponseRawJSON()
}

func (agent *Agent) GetLastRequestJSON() (string, error) {
	return agent.internalStructAgent.GetLastRequestJSON()
}

func (agent *Agent) GetLastResponseJSON() (string, error) {
	return agent.internalStructAgent.GetLastResponseJSON()
}

// GetContext returns the agent's context
func (agent *Agent) GetContext() context.Context {
	return agent.internalStructAgent.GetContext()
}

// SetContext updates the agent's context
func (agent *Agent) SetContext(ctx context.Context) {
	agent.internalStructAgent.SetContext(ctx)
}

// GetAgentForTopic returns the agent ID for a given topic based on routing configuration
// Returns empty string if no routing config is set or no match is found
func (agent *Agent) GetAgentForTopic(topic string) string {
	// Return empty string if no routing config is set
	if agent.agentRoutingConfig == nil {
		return ""
	}

	topicLower := strings.ToLower(topic)

	// Search through routing rules
	for _, rule := range agent.agentRoutingConfig.Routing {
		for _, configTopic := range rule.Topics {
			if strings.ToLower(configTopic) == topicLower {
				return rule.Agent
			}
		}
	}

	// Return default agent if no match found
	return agent.agentRoutingConfig.DefaultAgent
}
