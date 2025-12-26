package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()
	engineURL := "http://localhost:12434/engines/llama.cpp/v1"

	// === ORCHESTRATOR SYSTEM INSTRUCTIONS ===
	// Define the cooking topics the orchestrator should recognize
	systemInstructions := `
You are an expert at identifying cooking and food-related topics.
Given a user's input, identify the main culinary topic in only one word.

The possible cooking topics are:
- Baking (bread, cakes, pastries, cookies)
- Grilling (BBQ, grilled meats, vegetables)
- Desserts (sweets, ice cream, puddings)
- Vegetables (salads, vegetable dishes, veggie prep)
- Seafood (fish, shellfish, sushi)
- Pasta (Italian pasta dishes, noodles)
- Meat (steaks, roasts, meat preparation)
- Beverages (drinks, smoothies, cocktails)
- Asian (Asian cuisine, stir-fry, rice dishes)
- Mexican (tacos, burritos, salsa)
- Breakfast (eggs, pancakes, morning meals)
- Soup (broths, stews, chowders)
- Sauce (condiments, dressings, gravies)
- Technique (cooking methods, knife skills, tips)

Respond in JSON format with the field 'topic_discussion'.
Example: {"topic_discussion": "Baking"}

If the topic is not food-related, respond with: {"topic_discussion": "General"}
	`

	// === CREATE ORCHESTRATOR AGENT ===
	orchestratorAgent, err := orchestrator.NewAgent(
		ctx,
		agents.Config{
			Name:               "cooking-orchestrator",
			EngineURL:          engineURL,
			SystemInstructions: systemInstructions,
		},
		models.Config{
			Name:        "hf.co/menlo/lucy-gguf:q4_k_m",
			Temperature: models.Float64(0.0), // Low temperature for consistent classification
		},
	)
	if err != nil {
		panic(err)
	}

	// === TEST COOKING TOPIC DETECTION ===
	testQueries := []string{
		"How do I make chocolate chip cookies?",
		"What's the best way to grill a steak?",
		"Can you give me a recipe for tomato soup?",
		"How do I make homemade pasta?",
		"What's a good marinade for salmon?",
		"How do I make scrambled eggs fluffy?",
		"What spices go well with chicken curry?",
		"How do I prepare a Caesar salad?",
		"What's the best chocolate cake recipe?",
		"How do I cook rice perfectly?",
		"Tell me about knife sharpening techniques",
		"What's the capital of France?", // Non-cooking topic
	}

	fmt.Println("🍳 Cooking Topic Detection Orchestrator")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	for i, query := range testQueries {
		fmt.Printf("[%2d] Query: %s\n", i+1, query)

		// Detect topic using orchestrator
		topic, err := orchestratorAgent.IdentifyTopicFromText(query)
		if err != nil {
			fmt.Printf("     ❌ Error: %v\n\n", err)
			continue
		}

		// Display detected topic with emoji
		emoji := getTopicEmoji(topic)
		fmt.Printf("     ✅ Detected Topic: %s %s\n\n", emoji, topic)
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// === DEMONSTRATE ROUTING BASED ON TOPIC ===
	fmt.Println("📊 Topic Routing Examples")
	fmt.Println(strings.Repeat("-", 70))

	routingExamples := map[string]string{
		"Baking":     "Route to → Pastry Chef Agent",
		"Grilling":   "Route to → BBQ Master Agent",
		"Seafood":    "Route to → Seafood Specialist Agent",
		"Pasta":      "Route to → Italian Chef Agent",
		"Asian":      "Route to → Asian Cuisine Expert Agent",
		"Technique":  "Route to → Cooking Instructor Agent",
		"General":    "Route to → General Assistant Agent",
	}

	for topic, route := range routingExamples {
		emoji := getTopicEmoji(topic)
		fmt.Printf("%s %-12s → %s\n", emoji, topic, route)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
}

// getTopicEmoji returns an appropriate emoji for each cooking topic
func getTopicEmoji(topic string) string {
	emojiMap := map[string]string{
		"Baking":     "🥐",
		"Grilling":   "🔥",
		"Desserts":   "🍰",
		"Vegetables": "🥗",
		"Seafood":    "🐟",
		"Pasta":      "🍝",
		"Meat":       "🥩",
		"Beverages":  "🍹",
		"Asian":      "🍜",
		"Mexican":    "🌮",
		"Breakfast":  "🍳",
		"Soup":       "🍲",
		"Sauce":      "🥫",
		"Technique":  "🔪",
		"General":    "💬",
	}

	if emoji, exists := emojiMap[topic]; exists {
		return emoji
	}
	return "🍴" // Default food emoji
}
