package tasks

import (
	"context"
	"errors"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

type TasksAgentOption func(*Agent)

// BeforeCompletion sets a hook that is called before each intent identification
func BeforeCompletion(fn func(*Agent)) TasksAgentOption {
	return func(a *Agent) {
		a.beforeCompletion = fn
	}
}

// AfterCompletion sets a hook that is called after each intent identification
func AfterCompletion(fn func(*Agent)) TasksAgentOption {
	return func(a *Agent) {
		a.afterCompletion = fn
	}
}

// Agent represents an tasks agent that identifies tasks (plan) from user input
// It's a specialized structured agent that uses agents.Plan as its output type
type Agent struct {
	config              agents.Config
	modelConfig         models.Config
	internalStructAgent *structured.Agent[agents.Plan]
	log                 logger.Logger

	// Lifecycle hooks
	beforeCompletion func(*Agent)
	afterCompletion  func(*Agent)
}

// NewAgent creates a new tasks agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
	opts ...TasksAgentOption,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	// Create internal structured agent with agents.Intent type
	structAgent, err := structured.NewAgent[agents.Plan](ctx, agentConfig, modelConfig)
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
func (a *Agent) Kind() agents.Kind {
	return agents.Tasks
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

func (agent *Agent) IdentifyPlan(userMessages []messages.Message) (plan *agents.Plan, finishReason string, err error) {
	if len(userMessages) == 0 {
		return nil, "", errors.New("no messages provided")
	}

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Generate structured data (plan) from user messages
	plan, finishReason, err = agent.internalStructAgent.GenerateStructuredData(userMessages)
	if err != nil {
		return nil, finishReason, err
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return plan, finishReason, nil

}

func (agent *Agent) IdentifyPlanFromText(text string) (*agents.Plan, error) {
	userMessages := []messages.Message{
		{
			Role:    roles.User,
			Content: text,
		},
	}
	plan, _, err := agent.IdentifyPlan(userMessages)
	if err != nil {
		return nil, err
	}
	return plan, nil

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
