package models

import (
	"context"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

)

func GetModelsList(ctx context.Context, modelRunnerEndpoint string) ([]string, error) {

	// Initialize OpenAI client
	openaiClient := openai.NewClient(
		option.WithBaseURL(modelRunnerEndpoint),
		option.WithAPIKey(""),
	)
	modelsResponse, err := openaiClient.Models.List(ctx)
	if err != nil {
		log.Printf("Error fetching models: %v", err)
		return []string{}, err
	}
	models := []string{}
	for _, model := range modelsResponse.Data {

		models = append(models, model.ID)
	}
	return models, nil
}

func IsModelAvailable(ctx context.Context, modelRunnerEndpoint, modelID string) bool {
	openaiClient := openai.NewClient(
		option.WithBaseURL(modelRunnerEndpoint),
		option.WithAPIKey(""),
	)
	_, err := openaiClient.Models.Get(ctx, modelID)
	if err != nil {
		log.Printf("Model %s not available: %v", modelID, err)
		return false
	}
	return true
}




