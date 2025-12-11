package tools

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

type ConfirmationResponse int

const (
	Confirmed ConfirmationResponse = iota
	Denied
	Quit
)

// Tool represents a function tool with a fluent builder API
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]Parameter
	Required    []string
	//Function   func(string) (string, error)
}

// Parameter represents a function parameter
type Parameter struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// NewTool creates a new Tool with the given name
func NewTool(name string) *Tool {
	return &Tool{
		Name:       name,
		Parameters: make(map[string]Parameter),
		Required:   []string{},
	}
}

// SetDescription sets the description of the tool
func (t *Tool) SetDescription(description string) *Tool {
	t.Description = description
	return t
}

// AddParameter adds a parameter to the tool
// paramType should be one of: "string", "number", "boolean", "object", "array"
// isRequired indicates whether the parameter is required
func (t *Tool) AddParameter(name, paramType, description string, isRequired bool) *Tool {
	t.Parameters[name] = Parameter{
		Type:        paramType,
		Description: description,
	}
	if isRequired {
		t.Required = append(t.Required, name)
	}
	return t
}

// ToOpenAI converts the Tool to an OpenAI ChatCompletionToolUnionParam
func (t *Tool) ToOpenAI() openai.ChatCompletionToolUnionParam {
	properties := make(map[string]any)
	for name, param := range t.Parameters {
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
	if len(t.Required) > 0 {
		functionParams["required"] = t.Required
	}

	return openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        t.Name,
		Description: openai.String(t.Description),
		Parameters:  functionParams,
	})
}

// GetName returns the name of the tool
func (t *Tool) GetName() string {
	return t.Name
}

// GetDescription returns the description of the tool
func (t *Tool) GetDescription() string {
	return t.Description
}

// GetParameters returns the parameters of the tool
func (t *Tool) GetParameters() map[string]Parameter {
	return t.Parameters
}

// GetRequired returns the list of required parameter names
func (t *Tool) GetRequired() []string {
	return t.Required
}

// ToOpenAITools converts a slice of Tool pointers to a slice of OpenAI ChatCompletionToolUnionParam
func ToOpenAITools(tools []*Tool) []openai.ChatCompletionToolUnionParam {
	result := make([]openai.ChatCompletionToolUnionParam, len(tools))
	for i, tool := range tools {
		result[i] = tool.ToOpenAI()
	}
	return result
}
