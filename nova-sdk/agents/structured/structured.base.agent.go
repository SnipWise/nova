package structured

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/base"
)

// BaseAgent wraps the shared base.Agent for structured output functionality
type BaseAgent[Output any] struct {
	*base.Agent
}

type AgentOption[Output any] func(*BaseAgent[Output])

func NewBaseAgent[Output any](
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig openai.ChatCompletionNewParams,
	options ...AgentOption[Output],
) (structuredAgent *BaseAgent[Output], err error) {

	// Prepare the response format schema for structured output
	something := reflect.TypeOf((*Output)(nil)).Elem()
	schema := StructToJSONSchema(something)

	// Get schema name - handle slices/arrays
	schemaName := something.Name()
	if schemaName == "" {
		// For slices/arrays, use the element type name
		if something.Kind() == reflect.Slice || something.Kind() == reflect.Array {
			elemType := something.Elem()
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

	modelConfig.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			JSONSchema: schemaParam,
		},
	}

	// Create the shared base agent
	baseAgent, err := base.NewAgent(ctx, agentConfig, modelConfig)
	if err != nil {
		return nil, err
	}

	structuredAgent = &BaseAgent[Output]{
		Agent: baseAgent,
	}

	// Apply structured-specific options
	for _, option := range options {
		option(structuredAgent)
	}

	return structuredAgent, nil
}

func (agent *BaseAgent[Output]) Kind() (kind agents.Kind) {
	return agents.Structured
}

func (agent *BaseAgent[Output]) GenerateStructuredData(messages []openai.ChatCompletionMessageParamUnion) (response *Output, finishReason string, err error) {
	// Preserve existing system messages from agent.Params
	// Combine existing system messages with new messages
	agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, messages...)
	completion, err := agent.OpenaiClient.Chat.Completions.New(agent.Ctx, agent.ChatCompletionParams)

	if err != nil {
		return nil, "", err
	}

	if len(completion.Choices) > 0 {
		// Append the full response as an assistant message to the agent's messages
		agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, openai.AssistantMessage(completion.Choices[0].Message.Content))

		responseStr := completion.Choices[0].Message.Content

		var structuredResponse Output
		err = json.Unmarshal([]byte(responseStr), &structuredResponse)
		if err != nil {
			agent.Log.Error("Error unmarshaling structured response:", err)
			return nil, "", err
		}

		response = &structuredResponse

		finishReason = completion.Choices[0].FinishReason

		return response, finishReason, nil
	} else {
		return nil, "", errors.New("no choices returned from completion")
	}
}

func StructToJSONSchema(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle slices/arrays - wrap the element type in an array schema
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		return map[string]any{
			"type":  "array",
			"items": StructToJSONSchema(t.Elem()),
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}

	properties := schema["properties"].(map[string]any)
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Extract field name (before the comma)
		fieldName := strings.Split(jsonTag, ",")[0]

		// Determine the type
		fieldSchema := getFieldSchema(field.Type)
		properties[fieldName] = fieldSchema

		// Add to required fields if no omitempty tag
		if !strings.Contains(jsonTag, "omitempty") {
			required = append(required, fieldName)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func getFieldSchema(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice, reflect.Array:
		return map[string]any{
			"type":  "array",
			"items": getFieldSchema(t.Elem()),
		}
	case reflect.Struct:
		// For nested structures
		return map[string]any{"type": "object"}
	case reflect.Ptr:
		return getFieldSchema(t.Elem())
	default:
		return map[string]any{"type": "string"}
	}
}
