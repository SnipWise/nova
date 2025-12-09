package models

import (
	"github.com/openai/openai-go/v3"
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
}

// Helper functions to create pointers for optional parameters
func Float(v float64) *float64 {
	return &v
}

func Int(v int64) *int64 {
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

// WithParallelToolCalls sets whether to allow parallel tool calls
func (mc Config) WithParallelToolCalls(parallel bool) Config {
	mc.ParallelToolCalls = &parallel
	return mc
}

// WithTools sets the available tools for the model
func (mc Config) WithTools(tools ...openai.ChatCompletionToolUnionParam) Config {
	mc.Tools = tools
	return mc
}
