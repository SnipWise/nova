package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	engineURL := os.Getenv("ENGINE_URL")
	ctx := context.Background()

	systemInstructionsPath := os.Getenv("DATA_PATH") + "/system.instructions.md"
	systemInstructionsBytes, err := os.ReadFile(systemInstructionsPath)
	if err != nil {
		log.Fatalf("Failed to read system instructions file: %v", err)
	}
	systemInstructions := string(systemInstructionsBytes)
	//messages = append(messages, ai.NewSystemTextMessage(systemInstructions))
	nbSizeOfSystemInstructions := len([]rune(systemInstructions))
	fmt.Printf("🧠 Loaded system instructions (%d characters) from %s\n", nbSizeOfSystemInstructions, systemInstructionsPath)

	// Create a simple agent without exposing OpenAI SDK types
	agent, err := chat.NewAgent(
		ctx,
		agents.Config{
			EngineURL:               engineURL,
			SystemInstructions:      systemInstructions,
			KeepConversationHistory: true,
		},
		models.Config{
			Name:        os.Getenv("CHAT_MODEL_ID"),
			Temperature: models.Float64(0.7),
		},
	)
	if err != nil {
		panic(err)
	}

	questionList := []string{
		"Hello who are you? Tell me something about yourself.",
		"Tell me the story of the hawaiian pizza.",
		"How to cook a Diavola pizza?",
		"What wine would you pair with a Margherita pizza?",
		"How do you calculate the hydration ratio for your dough?",
		"Which part of the dough making process is the most important?",
		"I need a recipe for a dessert pizza. Can you help me with that?",
		"What kind of cheese would you recommend for a four-cheese pizza?",
		"Can you give me tips to make a perfect pizza at home?",
		"What are the common mistakes when making pizza dough?",
		"How do you achieve the perfect crust texture?",
		"What toppings go best with a spicy pizza?",
		"Can you suggest a wine pairing for a Diavola pizza?",
		"What is the history behind the Margherita pizza?",
		"How long should I ferment my pizza dough for optimal flavor?",
		"What type of flour do you recommend for pizza dough?",
		"How do you store leftover pizza dough?",
		"What are some unique pizza recipes you recommend?",
		"Can you explain the difference between Neapolitan and New York-style pizza?",
		"What are your thoughts on gluten-free pizza dough?",
		"How do you make a vegan pizza that tastes great?",
		"What are the best techniques for stretching pizza dough?",
		"How do you prevent soggy pizza crusts?",
		"What are some creative pizza topping combinations?",
		"Can you share a secret ingredient that enhances pizza flavor?",
		"What are the health benefits of homemade pizza?",
		"How do you make a stuffed crust pizza?",
		"What are some tips for baking pizza in a home oven?",
		"How do you make a pizza sauce from scratch?",
		"What are the best cheeses to use on pizza?",
		"How do you make a pizza that appeals to kids?",
		"What are some popular pizza styles around the world?",
		"How do you make a pizza with a crispy thin crust?",
		"What are some tips for making pizza dough ahead of time?",
		"How do you make a pizza that is both spicy and sweet?",
		"What are some traditional Italian pizza recipes?",
		"How do you make a pizza that is low in carbs?",
		"What are some tips for grilling pizza outdoors?",
		"How do you make a breakfast pizza?",
		"What are some gourmet pizza recipes you recommend?",
	}

	for idx, question := range questionList {

		_, err := agent.GenerateStreamCompletion(
			[]messages.Message{
				{
					Role:    roles.User,
					Content: question,
				},
			},
			func(chunk string, finishReason string) error {

				// Use markdown chunk parser for colorized streaming output
				if chunk != "" {
					//display.MarkdownChunk(markdownParser, chunk)
					fmt.Print(chunk)
				}
				if finishReason == "stop" {
					//markdownParser.Flush()
					fmt.Println()
				}
				return nil
			},
		)

		if err != nil {
			log.Fatal(err)
		}

		totalSize := 0
		for _, msg := range agent.GetMessages() {
			totalSize += len([]rune(msg.Content))
		}

		fmt.Printf("\n\n🧠 Q:%d - Total conversation size: %d characters (including system instructions)\n", idx+1, totalSize)
		// tokens estimation
		// approx 4 characters per token for English text
		approxTokens := int(float64(totalSize) / 3.42)
		//fmt.Printf("🧮 (~ %d tokens) %d\n", approxTokens, fullResponse.Usage.OutputTokens)
		fmt.Printf("🧮 (~ %d tokens)\n", approxTokens)

		fmt.Println() // New line after the response
		// Append user message to history

	}

	// for {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	fmt.Printf("🤖🧠 [%s](%s) ask me something - /bye to exit> ", agentName, modelId)
	// 	userMessage, _ := reader.ReadString('\n')

	// 	if strings.HasPrefix(userMessage, "/bye") {
	// 		fmt.Println("👋 Bye!")
	// 		break
	// 	}

	// 	if strings.HasPrefix(userMessage, "/history") {
	// 		fmt.Println("📝 Conversation history:")
	// 		for i, msg := range messages {
	// 			// Convert []*ai.Part to string for display
	// 			var parts []string
	// 			for _, part := range msg.Content {
	// 				parts = append(parts, part.Text)
	// 			}
	// 			fmt.Printf("  [%d] %s: %s\n", i, msg.Role, strings.Join(parts, " "))
	// 		}
	// 		continue
	// 	}

	// 	fullResponse, err := genkit.Generate(ctx, g,
	// 		ai.WithModelName(modelId),
	// 		ai.WithSystem(systemInstructions),
	// 		// WithMessages sets the messages. These messages will be sandwiched between the system and user prompts.
	// 		ai.WithMessages(
	// 			messages...,
	// 		),
	// 		ai.WithPrompt(userMessage),
	// 		ai.WithConfig(map[string]any{"temperature": 0.7}),

	// 		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
	// 			// Do something with the chunk...
	// 			fmt.Print(chunk.Text())
	// 			return nil
	// 		}),
	// 	)

	// 	totalSize := 0
	// 	for _, msg := range messages {
	// 		for _, part := range msg.Content {
	// 			totalSize += len([]rune(part.Text))
	// 		}
	// 	}
	// 	totalSize += len([]rune(userMessage))
	// 	totalSize += len([]rune(fullResponse.Text()))
	// 	totalSize += nbSizeOfSystemInstructions
	// 	fmt.Printf("\n🧠 Total conversation size: %d characters (including system instructions)\n", totalSize)

	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	fmt.Println() // New line after the response
	// 	// Append user message to history
	// 	messages = append(messages, ai.NewUserTextMessage(strings.TrimSpace(userMessage)))
	// 	// Append assistant response to history
	// 	messages = append(messages, ai.NewModelTextMessage(strings.TrimSpace(fullResponse.Text())))

	// }

}
