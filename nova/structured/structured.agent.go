package structured

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova/agents"
	"github.com/snipwise/nova/nova/messages"
	"github.com/snipwise/nova/nova/models"
	"github.com/snipwise/nova/nova/toolbox/logger"
)

// StructuredResult represents the result of structured data generation
type StructuredResult[Output any] struct {
	Data         *Output
	FinishReason string
}

// Agent represents a simplified structured data agent that hides OpenAI SDK details
type Agent[Output any] struct {
	ctx           context.Context
	config        agents.AgentConfig
	modelConfig   models.Config
	messages      []messages.Message
	internalAgent *BaseAgent[Output]
	log           logger.Logger
}

// NewAgent creates a new simplified structured data agent
func NewAgent[Output any](
	ctx context.Context,
	agentConfig agents.AgentConfig,
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
		messages:      []messages.Message{},
		internalAgent: internalAgent,
		log:           log,
	}

	// Add system instruction as first message
	agent.messages = append(agent.messages, messages.Message{
		Role:    "system",
		Content: agentConfig.SystemInstructions,
	})

	return agent, nil
}

// Kind returns the agent type
func (agent *Agent[Output]) Kind() agents.Kind {
	return agents.Structured
}


// Generate sends messages and returns structured data
func (agent *Agent[Output]) GenerateStructuredData(userMessages []messages.Message) (response *Output, finishReason string, err error) {
	if len(userMessages) == 0 {
		return nil, "", errors.New("no messages provided")
	}

	// Add user messages to history
	agent.messages = append(agent.messages, userMessages...)

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
	agent.messages = append(agent.messages, messages.Message{
		Role:    "assistant",
		Content: string(jsonData),
	})

	return response, finishReason, nil
}
