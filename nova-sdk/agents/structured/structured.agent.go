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
	ctx           context.Context
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
		ctx:           ctx,
		config:        agentConfig,
		modelConfig:   modelConfig,
		internalAgent: internalAgent,
		log:           log,
	}

	// Add system instruction as first message
	agent.internalAgent.AddMessage(
		openai.SystemMessage(agentConfig.SystemInstructions),
	)

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

func (agent *Agent[Output]) AddMessage(role roles.Role, content string) {

	agent.internalAgent.AddMessage(
		messages.ConvertToOpenAIMessage(messages.Message{
			Role:    role,
			Content: content,
		}),
	)
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

	// Call internal agent
	response, finishReason, err = agent.internalAgent.GenerateStructuredData(openaiMessages)
	if err != nil {
		return nil, finishReason, err
	}

	// Add assistant response to history (as JSON string)
	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, finishReason, err
	}

	agent.internalAgent.AddMessage(
		openai.AssistantMessage(string(jsonData)),
	)

	return response, finishReason, nil
}
