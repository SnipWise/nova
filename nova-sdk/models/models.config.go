package models

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

// Config represents model configuration parameters
type Config struct {
	Name              string
	Temperature       *float64
	TopP              *float64
	TopK              *int64
	MinP              *float64
	MaxTokens         *int64
	FrequencyPenalty  *float64
	PresencePenalty   *float64
	RepeatPenalty     *float64
	Seed              *int64
	Stop              []string
	N                 *int64
	ToolChoice        *openai.ChatCompletionToolChoiceOptionUnionParam
	ParallelToolCalls *bool
	Tools             []openai.ChatCompletionToolUnionParam
	ReasoningEffort   *string
}

// Helper functions to create pointers for optional parameters
func Float(v float64) *float64 {
	return &v
}

func Int(v int64) *int64 {
	return &v
}

// Float64 returns a pointer to the given float64 value
func Float64(v float64) *float64 {
	return &v
}

// Int64 returns a pointer to the given int64 value
func Int64(v int64) *int64 {
	return &v
}

// Bool returns a pointer to the given bool value
func Bool(v bool) *bool {
	return &v
}

// String returns a pointer to the given string value
func String(v string) *string {
	return &v
}


// NewConfig creates a Config with just the model name
func NewConfig(name string) Config {
	return Config{Name: name}
}

// WithTemperature sets the temperature parameter
func (mc Config) WithTemperature(temp float64) Config {
	mc.Temperature = Float(temp)
	return mc
}

// WithTopP sets the top_p parameter
func (mc Config) WithTopP(topP float64) Config {
	mc.TopP = Float(topP)
	return mc
}

// WithMaxTokens sets the max_tokens parameter
func (mc Config) WithMaxTokens(maxTokens int64) Config {
	mc.MaxTokens = Int(maxTokens)
	return mc
}

// WithFrequencyPenalty sets the frequency_penalty parameter
func (mc Config) WithFrequencyPenalty(penalty float64) Config {
	mc.FrequencyPenalty = Float(penalty)
	return mc
}

// WithPresencePenalty sets the presence_penalty parameter
func (mc Config) WithPresencePenalty(penalty float64) Config {
	mc.PresencePenalty = Float(penalty)
	return mc
}

// WithSeed sets the seed parameter for deterministic sampling
func (mc Config) WithSeed(seed int64) Config {
	mc.Seed = Int(seed)
	return mc
}

// WithTopK sets the top_k parameter
func (mc Config) WithTopK(topK int64) Config {
	mc.TopK = Int(topK)
	return mc
}

// WithMinP sets the min_p parameter (minimum probability threshold)
func (mc Config) WithMinP(minP float64) Config {
	mc.MinP = Float(minP)
	return mc
}

// WithRepeatPenalty sets the repeat_penalty parameter
func (mc Config) WithRepeatPenalty(penalty float64) Config {
	mc.RepeatPenalty = Float(penalty)
	return mc
}

// WithStop sets the stop sequences
func (mc Config) WithStop(stop ...string) Config {
	mc.Stop = stop
	return mc
}

// WithN sets the number of completions to generate
func (mc Config) WithN(n int64) Config {
	mc.N = Int(n)
	return mc
}

// WithToolChoice sets the tool choice parameter
func (mc Config) WithToolChoice(toolChoice openai.ChatCompletionToolChoiceOptionUnionParam) Config {
	mc.ToolChoice = &toolChoice
	return mc
}

// WithToolChoiceAuto sets tool choice to "auto" - the model decides whether to use tools
func (mc Config) WithToolChoiceAuto() Config {
	mc.ToolChoice = &openai.ChatCompletionToolChoiceOptionUnionParam{
		OfAuto: openai.String("auto"),
	}
	return mc
}

// WithToolChoiceFunction forces the model to use a specific function/tool
func (mc Config) WithToolChoiceFunction(functionName string) Config {
	toolChoice := openai.ToolChoiceOptionFunctionToolChoice(openai.ChatCompletionNamedToolChoiceFunctionParam{
		Name: functionName,
	})
	mc.ToolChoice = &toolChoice
	return mc
}

// WithParallelToolCalls sets whether to allow parallel tool calls
func (mc Config) WithParallelToolCalls(parallel bool) Config {
	mc.ParallelToolCalls = &parallel
	return mc
}

// WithTools sets the available tools for the model
func (mc Config) WithTools(tools []openai.ChatCompletionToolUnionParam) Config {
	mc.Tools = tools
	return mc
}

// WithReasoningEffort sets the reasoning effort parameter
// Supported values are: none, minimal, low, medium, high, and xhigh
func (mc Config) WithReasoningEffort(effort string) Config {
	mc.ReasoningEffort = String(effort)
	return mc
}

// Reasoning effort constants
/*
The reasoning_effort parameter allows you to control the depth of reasoning the model applies when generating responses. Higher levels of reasoning effort can lead to more thoughtful and accurate answers, especially for complex queries.
From the OpenAI documentation:
Constrains effort on reasoning for
[reasoning models](https://platform.openai.com/docs/guides/reasoning). Currently
supported values are `none`, `minimal`, `low`, `medium`, `high`, and `xhigh`.
Reducing reasoning effort can result in faster responses and fewer tokens used
on reasoning in a response.
*/
const (
	ReasoningEffortNone    = "none"
	ReasoningEffortMinimal = "minimal"
	ReasoningEffortLow     = "low"
	ReasoningEffortMedium  = "medium"
	ReasoningEffortHigh    = "high"
	ReasoningEffortXHigh   = "xhigh"
)

// WithReasoningEffortNone sets reasoning effort to none
func (mc Config) WithReasoningEffortNone() Config {
	mc.ReasoningEffort = String(ReasoningEffortNone)
	return mc
}

// WithReasoningEffortMinimal sets reasoning effort to minimal
func (mc Config) WithReasoningEffortMinimal() Config {
	mc.ReasoningEffort = String(ReasoningEffortMinimal)
	return mc
}

// WithReasoningEffortLow sets reasoning effort to low
func (mc Config) WithReasoningEffortLow() Config {
	mc.ReasoningEffort = String(ReasoningEffortLow)
	return mc
}

// WithReasoningEffortMedium sets reasoning effort to medium
func (mc Config) WithReasoningEffortMedium() Config {
	mc.ReasoningEffort = String(ReasoningEffortMedium)
	return mc
}

// WithReasoningEffortHigh sets reasoning effort to high
func (mc Config) WithReasoningEffortHigh() Config {
	mc.ReasoningEffort = String(ReasoningEffortHigh)
	return mc
}

// WithReasoningEffortXHigh sets reasoning effort to xhigh
func (mc Config) WithReasoningEffortXHigh() Config {
	mc.ReasoningEffort = String(ReasoningEffortXHigh)
	return mc
}

// =====

func ConvertToOpenAIModelConfig(modelConfig Config) openai.ChatCompletionNewParams {
	openaiModelConfig := openai.ChatCompletionNewParams{
		Model: modelConfig.Name,
	}

	// Set optional parameters if provided
	if modelConfig.Temperature != nil {
		openaiModelConfig.Temperature = openai.Float(*modelConfig.Temperature)
	}
	if modelConfig.TopP != nil {
		openaiModelConfig.TopP = openai.Float(*modelConfig.TopP)
	}
	if modelConfig.MaxTokens != nil {
		openaiModelConfig.MaxTokens = openai.Int(*modelConfig.MaxTokens)
	}
	if modelConfig.FrequencyPenalty != nil {
		openaiModelConfig.FrequencyPenalty = openai.Float(*modelConfig.FrequencyPenalty)
	}
	if modelConfig.PresencePenalty != nil {
		openaiModelConfig.PresencePenalty = openai.Float(*modelConfig.PresencePenalty)
	}
	if modelConfig.Seed != nil {
		openaiModelConfig.Seed = openai.Int(*modelConfig.Seed)
	}
	if modelConfig.N != nil {
		openaiModelConfig.N = openai.Int(*modelConfig.N)
	}
	if modelConfig.ToolChoice != nil {
		openaiModelConfig.ToolChoice = *modelConfig.ToolChoice
	}
	if modelConfig.ParallelToolCalls != nil {
		openaiModelConfig.ParallelToolCalls = openai.Bool(*modelConfig.ParallelToolCalls)
	}
	if modelConfig.Tools != nil {
		openaiModelConfig.Tools = modelConfig.Tools
	}
	if modelConfig.ReasoningEffort != nil {
		openaiModelConfig.ReasoningEffort = shared.ReasoningEffort(*modelConfig.ReasoningEffort)
	}

	return openaiModelConfig
}
