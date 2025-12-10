package tools

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

// Tool represents a function tool with a fluent builder API
type Tool struct {
	name        string
	description string
	parameters  map[string]Parameter
	required    []string
}

// Parameter represents a function parameter
type Parameter struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// NewTool creates a new Tool with the given name
func NewTool(name string) *Tool {
	return &Tool{
		name:       name,
		parameters: make(map[string]Parameter),
		required:   []string{},
	}
}

// SetDescription sets the description of the tool
func (t *Tool) SetDescription(description string) *Tool {
	t.description = description
	return t
}

// AddParameter adds a parameter to the tool
// paramType should be one of: "string", "number", "boolean", "object", "array"
// isRequired indicates whether the parameter is required
func (t *Tool) AddParameter(name, paramType, description string, isRequired bool) *Tool {
	t.parameters[name] = Parameter{
		Type:        paramType,
		Description: description,
	}
	if isRequired {
		t.required = append(t.required, name)
	}
	return t
}

// ToOpenAI converts the Tool to an OpenAI ChatCompletionToolUnionParam
func (t *Tool) ToOpenAI() openai.ChatCompletionToolUnionParam {
	properties := make(map[string]any)
	for name, param := range t.parameters {
		properties[name] = map[string]string{
			"type":        param.Type,
			"description": param.Description,
		}
	}

	functionParams := shared.FunctionParameters{
		"type":       "object",
		"properties": properties,
	}

	// Only add required field if there are required parameters
	if len(t.required) > 0 {
		functionParams["required"] = t.required
	}

	return openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        t.name,
		Description: openai.String(t.description),
		Parameters:  functionParams,
	})
}

// GetName returns the name of the tool
func (t *Tool) GetName() string {
	return t.name
}

// GetDescription returns the description of the tool
func (t *Tool) GetDescription() string {
	return t.description
}

// GetParameters returns the parameters of the tool
func (t *Tool) GetParameters() map[string]Parameter {
	return t.parameters
}

// GetRequired returns the list of required parameter names
func (t *Tool) GetRequired() []string {
	return t.required
}

// ToOpenAITools converts a slice of Tool pointers to a slice of OpenAI ChatCompletionToolUnionParam
func ToOpenAITools(tools []*Tool) []openai.ChatCompletionToolUnionParam {
	result := make([]openai.ChatCompletionToolUnionParam, len(tools))
	for i, tool := range tools {
		result[i] = tool.ToOpenAI()
	}
	return result
}
