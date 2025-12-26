package orchestrator

import (
	"context"
	"errors"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/base"
	"github.com/snipwise/nova/nova-sdk/agents/structured"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// Agent represents an orchestrator agent that identifies topics/intents from user input
// It's a specialized structured agent that uses agents.Intent as its output type
type Agent struct {
	ctx                 context.Context
	config              agents.Config
	modelConfig         models.Config
	internalStructAgent *structured.Agent[agents.Intent]
	log                 logger.Logger
}

// NewAgent creates a new orchestrator agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	// Create internal structured agent with agents.Intent type
	structAgent, err := structured.NewAgent[agents.Intent](ctx, agentConfig, modelConfig)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		ctx:                 ctx,
		config:              agentConfig,
		modelConfig:         modelConfig,
		internalStructAgent: structAgent,
		log:                 log,
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

	// Call internal structured agent to generate Intent
	intent, finishReason, err = agent.internalStructAgent.GenerateStructuredData(userMessages)
	if err != nil {
		return nil, finishReason, err
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

// === Telemetry Methods ===

// GetLastRequestJSON returns the last request sent to the LLM as JSON
func (agent *Agent) GetLastRequestJSON() (string, error) {
	return agent.internalStructAgent.GetLastRequestJSON()
}

// GetLastRequestContextLength returns the context length of the last request
func (agent *Agent) GetLastRequestContextLength() int {
	return agent.internalStructAgent.GetLastRequestContextLength()
}

// GetLastRequestMetadata returns metadata about the last request
func (agent *Agent) GetLastRequestMetadata() base.RequestMetadata {
	return agent.internalStructAgent.GetLastRequestMetadata()
}

// GetLastResponseJSON returns the last response received from the LLM as JSON
func (agent *Agent) GetLastResponseJSON() (string, error) {
	return agent.internalStructAgent.GetLastResponseJSON()
}

// GetLastResponseMetadata returns metadata about the last response
func (agent *Agent) GetLastResponseMetadata() base.ResponseMetadata {
	return agent.internalStructAgent.GetLastResponseMetadata()
}

// GetConversationHistoryJSON returns the entire conversation history as JSON
func (agent *Agent) GetConversationHistoryJSON() (string, error) {
	return agent.internalStructAgent.GetConversationHistoryJSON()
}

// GetTotalTokensUsed returns the total number of tokens used since the agent was created
func (agent *Agent) GetTotalTokensUsed() int {
	return agent.internalStructAgent.GetTotalTokensUsed()
}

// ResetTelemetry resets all telemetry counters and stored data
func (agent *Agent) ResetTelemetry() {
	agent.internalStructAgent.ResetTelemetry()
}

// SetTelemetryCallback sets a callback for receiving telemetry events in real-time
func (agent *Agent) SetTelemetryCallback(callback base.TelemetryCallback) {
	agent.internalStructAgent.SetTelemetryCallback(callback)
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
