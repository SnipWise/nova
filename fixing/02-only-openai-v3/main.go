package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {

	baseURL := os.Getenv("ENGINE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:12434/engines/v1/"
	}
	model := os.Getenv("CHAT_MODEL_ID")
	if model == "" {
		model = "ai/qwen2.5:0.5B-F16"
	}

	systemInstruction := os.Getenv("SYSTEM_INSTRUCTION")
	if systemInstruction == "" {
		systemInstruction = "You are a helpful assistant."
	}
	userPrompt := os.Getenv("USER_PROMPT")
	if userPrompt == "" {
		userPrompt = "Explain the theory of relativity in simple terms."
	}

	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(""),
	)

	ctx := context.Background()

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemInstruction),
		openai.UserMessage(userPrompt),
	}

	// IMPORTANT: Adjust temperature and top_p for desired creativity and coherence
	temperature, _ := strconv.ParseFloat(os.Getenv("TEMPERATURE"), 64)
	topP , _ := strconv.ParseFloat(os.Getenv("TOP_P"), 64)

	param := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       model,
		Temperature: openai.Opt(temperature),
		TopP: 	 openai.Opt(topP),
	}
	// NOTE:: Starting a streaming chat completion
	stream := client.Chat.Completions.NewStreaming(ctx, param)

	for stream.Next() {
		chunk := stream.Current()
		// Stream each chunk as it arrives
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}

	if err := stream.Err(); err != nil {
		log.Fatalln("😡:", err)
	}
}
