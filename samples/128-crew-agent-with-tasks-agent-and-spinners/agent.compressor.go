package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/ui/spinner"
)

// createCompressorAgent creates the Compressor Agent for context packing.
// When conversation history exceeds the context size limit, the compressor
// summarises older messages so the chat agent stays within its token budget.
func createCompressorAgent(ctx context.Context, cfg *AppConfig) (*compressor.Agent, error) {
	ac, err := cfg.getAgentConfig("compressor")
	if err != nil {
		return nil, err
	}

	compressorSpinner := spinner.NewWithColor("").
		SetFrameColor(spinner.ColorCyan).
		SetFrames(spinner.FramesDots).
		SetSuffix("Compressing context...").
		SetSuffixColor(spinner.ColorBold + spinner.ColorBrightCyan)

	return compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "compressor-agent",
			EngineURL:          cfg.EngineURL,
			SystemInstructions: ac.Instructions,
		},
		models.Config{
			Name:        ac.Model,
			Temperature: models.Float64(ac.Temperature),
		},
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
		compressor.BeforeCompletion(func(a *compressor.Agent) {
			compressorSpinner.Start()
		}),
		compressor.AfterCompletion(func(a *compressor.Agent) {
			compressorSpinner.Success("Context compressed!")
		}),
	)
}
