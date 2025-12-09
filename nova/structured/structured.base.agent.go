package structured

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/snipwise/nova/nova/agents"
	"github.com/snipwise/nova/nova/toolbox/logger"
)

type BaseAgent[Output any] struct {
	ctx    context.Context
	config agents.AgentConfig

	chatCompletionParams openai.ChatCompletionNewParams
	openaiClient         openai.Client
	log                  logger.Logger
}

type AgentOption[Output any] func(*BaseAgent[Output])

func NewBaseAgent[Output any](
	ctx context.Context,
	agentConfig agents.AgentConfig,
	modelConfig openai.ChatCompletionNewParams,
	options ...AgentOption[Output],
) (structuredAgent *BaseAgent[Output], err error) {

	// export SNIP_LOG_LEVEL=debug  # Shows all logs
	// export SNIP_LOG_LEVEL=info   # Shows info, warn, error
	// export SNIP_LOG_LEVEL=warn   # Shows warn, error only
	// export SNIP_LOG_LEVEL=error  # Shows errors only
	// export SNIP_LOG_LEVEL=none   # No logging (default)

	// Create logger from environment variable
	log := logger.GetLoggerFromEnv()

	client := openai.NewClient(
		option.WithBaseURL(agentConfig.EngineURL),
		option.WithAPIKey("I💙DockerModelRunner"),
	)

	_, err = client.Models.Get(ctx, modelConfig.Model)
	if err != nil {
		log.Error("Model not available:", err)
		return nil, errors.New("model not available on the specified engine URL")
	}
	log.Info("Model %s is available on %s", modelConfig.Model, agentConfig.EngineURL)

	something := reflect.TypeOf((*Output)(nil)).Elem()
	schema := StructToJSONSchema(something)

	// schema to json string
	// jsonSchemaBytes, err := json.MarshalIndent(schema, "", "  ")
	// if err != nil {
	// 	log.Error("Error marshaling schema to JSON:", err)
	// 	return nil, err
	// }
	// log.Info(string(jsonSchemaBytes))

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        something.Name(),
		Description: openai.String("Notable information about a " + strings.ToLower(something.Name())),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}

	modelConfig.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			JSONSchema: schemaParam,
		},
	}

	structuredAgent = &BaseAgent[Output]{
		ctx:                  ctx,
		config:               agentConfig,
		chatCompletionParams: modelConfig,
		openaiClient:         client,
		log:                  log,
	}

	structuredAgent.chatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{}

	structuredAgent.chatCompletionParams.Messages = append(structuredAgent.chatCompletionParams.Messages, openai.SystemMessage(agentConfig.SystemInstructions))

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
	agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, messages...)
	completion, err := agent.openaiClient.Chat.Completions.New(agent.ctx, agent.chatCompletionParams)

	if err != nil {
		return nil, "", err
	}

	if len(completion.Choices) > 0 {
		// Append the full response as an assistant message to the agent's messages
		agent.chatCompletionParams.Messages = append(agent.chatCompletionParams.Messages, openai.AssistantMessage(completion.Choices[0].Message.Content))

		responseStr := completion.Choices[0].Message.Content

		var structuredResponse Output
		err = json.Unmarshal([]byte(responseStr), &structuredResponse)
		if err != nil {
			agent.log.Error("Error unmarshaling structured response:", err)
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
