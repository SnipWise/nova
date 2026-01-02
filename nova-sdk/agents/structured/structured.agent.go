package structured

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// StructuredResult represents the result of structured data generation
type StructuredResult[Output any] struct {
	Data         *Output
	FinishReason string
}

// Agent represents a simplified structured data agent that hides OpenAI SDK details
type Agent[Output any] struct {
	config        agents.Config
	modelConfig   models.Config
	internalAgent *BaseAgent[Output]
	log           logger.Logger
}

// NewAgent creates a new simplified structured data agent
func NewAgent[Output any](
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
) (*Agent[Output], error) {
	log := logger.GetLoggerFromEnv()

	// Create internal OpenAI-based agent with converted parameters
	openaiModelConfig := models.ConvertToOpenAIModelConfig(modelConfig)

	// Generate JSON Schema from Output type
	outputType := reflect.TypeOf((*Output)(nil)).Elem()
	schema := StructToJSONSchema(outputType)

	// Get schema name - handle slices/arrays
	schemaName := outputType.Name()
	if schemaName == "" {
		// For slices/arrays, use the element type name
		if outputType.Kind() == reflect.Slice || outputType.Kind() == reflect.Array {
			elemType := outputType.Elem()
			schemaName = elemType.Name() + "Array"
		} else {
			schemaName = "Response"
		}
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        schemaName,
		Description: openai.String("Notable information about " + strings.ToLower(schemaName)),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}

	openaiModelConfig.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			JSONSchema: schemaParam,
		},
	}

	internalAgent, err := NewBaseAgent[Output](ctx, agentConfig, openaiModelConfig)
	if err != nil {
		return nil, err
	}

	agent := &Agent[Output]{
		config:        agentConfig,
		modelConfig:   modelConfig,
		internalAgent: internalAgent,
		log:           log,
	}

	// System message is already added by the BaseAgent constructor
	// No need to add it again here

	return agent, nil
}

// Kind returns the agent type
func (agent *Agent[Output]) Kind() agents.Kind {
	return agents.Structured
}

func (agent *Agent[Output]) GetName() string {
	return agent.config.Name
}

func (agent *Agent[Output]) GetModelID() string {
	return agent.modelConfig.Name
}

// GetMessages returns all conversation messages
func (agent *Agent[Output]) GetMessages() []messages.Message {
	openaiMessages := agent.internalAgent.GetMessages()
	agentMessages := messages.ConvertFromOpenAIMessages(openaiMessages)
	return agentMessages
}

func (agent *Agent[Output]) ExportMessagesToJSON() (string, error) {
	messagesList := agent.GetMessages()
	jsonData, err := json.MarshalIndent(messagesList, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (agent *Agent[Output]) AddMessage(role roles.Role, content string) {

	agent.internalAgent.AddMessage(
		messages.ConvertToOpenAIMessage(messages.Message{
			Role:    role,
			Content: content,
		}),
	)
}

// AddMessages adds multiple messages to the conversation history
func (agent *Agent[Output]) AddMessages(msgs []messages.Message) {
	openaiMessages := messages.ConvertToOpenAIMessages(msgs)
	agent.internalAgent.AddMessages(openaiMessages)
}

// ResetMessages clears all messages except the system instruction
func (agent *Agent[Output]) ResetMessages() {
	agent.internalAgent.ResetMessages()
}

// Generate sends messages and returns structured data
func (agent *Agent[Output]) GenerateStructuredData(userMessages []messages.Message) (response *Output, finishReason string, err error) {
	if len(userMessages) == 0 {
		return nil, "", errors.New("no messages provided")
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent - it handles the conversation history based on KeepConversationHistory
	response, finishReason, err = agent.internalAgent.GenerateStructuredData(openaiMessages)
	if err != nil {
		return nil, finishReason, err
	}

	return response, finishReason, nil
}

// === Config Getters and Setters ===

// GetConfig returns the agent configuration
func (agent *Agent[Output]) GetConfig() agents.Config {
	return agent.config
}

// SetConfig updates the agent configuration
func (agent *Agent[Output]) SetConfig(config agents.Config) {
	agent.config = config
	agent.internalAgent.Config = config
}

// GetModelConfig returns the model configuration
func (agent *Agent[Output]) GetModelConfig() models.Config {
	return agent.modelConfig
}

// SetModelConfig updates the model configuration
// Note: This updates the stored config but doesn't regenerate the internal OpenAI params
// For most parameters to take effect, create a new agent with the new config
func (agent *Agent[Output]) SetModelConfig(config models.Config) {
	agent.modelConfig = config
	// Update the internal OpenAI params with the new config
	agent.internalAgent.ChatCompletionParams = models.ConvertToOpenAIModelConfig(config)
}

func (agent *Agent[Output]) GetLastRequestRawJSON() string {
	return agent.internalAgent.GetLastRequestRawJSON()
}
func (agent *Agent[Output]) GetLastResponseRawJSON() string {
	return agent.internalAgent.GetLastResponseRawJSON()
}

func (agent *Agent[Output]) GetLastRequestJSON() (string, error) {
	return agent.internalAgent.GetLastRequestSON()
}

func (agent *Agent[Output]) GetLastResponseJSON() (string, error) {
	return agent.internalAgent.GetLastResponseJSON()
}

// GetContext returns the agent's context
func (agent *Agent[Output]) GetContext() context.Context {
	return agent.internalAgent.GetContext()
}

// SetContext updates the agent's context
func (agent *Agent[Output]) SetContext(ctx context.Context) {
	agent.internalAgent.SetContext(ctx)
}
