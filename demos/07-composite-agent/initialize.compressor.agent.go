package main

import (
	"context"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/env"
)

func (ca *CompositeAgent) initializeCompressorAgent(ctx context.Context, engineURL string) error {
	compressorModel := env.GetEnvOrDefault("COMPRESSOR_MODEL", "hf.co/menlo/jan-nano-gguf:q4_k_m")
	
	compressorAgent, err := compressor.NewAgent(
		ctx,
		agents.Config{
			Name:               "compressor-agent",
			EngineURL:          engineURL,
			SystemInstructions: compressor.Instructions.Expert,
		},
		models.Config{
			Name:        compressorModel,
			Temperature: models.Float64(0.0),
		},
		compressor.WithCompressionPrompt(compressor.Prompts.Minimalist),
	)
	if err != nil {
		return err
	}

	ca.compressorAgent = compressorAgent
	return nil
}
