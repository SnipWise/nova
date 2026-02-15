package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/files"
	"github.com/snipwise/nova/nova-sdk/ui/display"
)

func main() {
	ctx := context.Background()
	engineUrl := "http://localhost:12434/engines/llama.cpp/v1"

	knowledgeBase, err := files.ReadTextFile("menu.xml")
	if err != nil {
		panic(err)
	}

	toolsAgent, err := tools.NewAgent(
		ctx,
		agents.Config{
			EngineURL: engineUrl,
			SystemInstructions: `You are Riker, a helpful AI assistant.
			IMPORTANT RULES FOR TOOL USAGE:
			- ONLY use the calculation tools when the user EXPLICITLY asks to calculate, compute, or determine prices/totals
			- Keywords that require tool usage: "calculate", "compute", "total", "how much", "price for X items"
			- Do NOT use tools when the user asks to "search", "display", "show", "tell me about", or "what is" - these are information queries only
			- If the user just wants information about menu items without calculations, do NOT call any tools
			- Always check if the user's question contains a calculation request before using tools

			memorizes each tool call result so that it can be used in subsequent tool calls
			`,
			KeepConversationHistory: true,
		},

		models.Config{
			Name: "hf.co/menlo/jan-nano-gguf:q4_k_m",
			//Name: "hf.co/menlo/lucy-gguf:q4_k_m",
			//Name: "ai/qwen3",
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(false),
		},

		tools.WithTools([]*tools.Tool{

			tools.NewTool("get_price_item").
				SetDescription("Get the price of a menu item by its name").
				AddParameter("item_name", "string", "the name of the menu item (case insensitive)", true),

			tools.NewTool("calculate_line_item_total").
				SetDescription("Get the price of a menu item and calculate the total for a given quantity. Use this when user asks to calculate X times the price of an item.").
				//AddParameter("item_name", "string", "the name of the menu item (case insensitive)", true).
				AddParameter("quantity", "number", "the number of items ordered", true).
				AddParameter("item_price", "number", "the price of the item (if already known, otherwise can be 0 or omitted)", true),

			tools.NewTool("calculate_order_total").
				SetDescription("Calculate the grand total by summing multiple line item totals").
				AddParameter("line_totals", "array", "array of line item totals to sum up", true),
		}),

		tools.WithExecuteFn(func(functionName string, arguments string) (string, error) {
			fmt.Printf("🟢 🔧 Executing tool: %s with arguments: %s\n", functionName,arguments)
			switch functionName {
			case "get_price_item":
				var args struct {
					ItemName string `json:"item_name"`
				}
				if err := json.Unmarshal([]byte(arguments), &args); err != nil {
					return `{"error": "Invalid arguments for get_price_item"}`, nil
				}

				// Search for the item in the knowledgeBase XML (case insensitive)
				searchName := strings.ToLower(args.ItemName)

				// Parse XML to find the item
				// Pattern: <name>ItemName</name> followed by <price>XX.XX</price>
				// (?is) = case insensitive + dotall (. matches newlines)
				pattern := `(?is)<item[^>]*>.*?<name>\s*([^<]+)\s*</name>.*?<price[^>]*>\s*([0-9.]+)\s*</price>.*?</item>`
				re := regexp.MustCompile(pattern)
				matches := re.FindAllStringSubmatch(knowledgeBase, -1)

				for _, match := range matches {
					if len(match) >= 3 {
						itemName := strings.TrimSpace(match[1])
						priceStr := strings.TrimSpace(match[2])

						if strings.ToLower(itemName) == searchName {
							price, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								return `{"error": "Failed to parse price"}`, nil
							}
							return fmt.Sprintf(`{"item_name": "%s", "price": %.2f}`, itemName, price), nil
						}
					}
				}

				return fmt.Sprintf(`{"error": "Item '%s' not found in menu"}`, args.ItemName), nil

			case "calculate_line_item_total":
				var args struct {
					//ItemName string `json:"item_name"`
					ItemPrice float64 `json:"item_price"`
					Quantity  float64 `json:"quantity"`
				}
				if err := json.Unmarshal([]byte(arguments), &args); err != nil {
					return `{"error": "Invalid arguments for calculate_sum"}`, nil
				}
				lineTotal := args.ItemPrice * args.Quantity
				return fmt.Sprintf(`{"line_total": %.2f}`, lineTotal), nil

			case "calculate_order_total":
				var args struct {
					LineTotals []float64 `json:"line_totals"`
				}
				if err := json.Unmarshal([]byte(arguments), &args); err != nil {
					return `{"error": "Invalid arguments for calculate_order_total"}`, nil
				}
				var orderTotal float64
				for _, lineTotal := range args.LineTotals {
					orderTotal += lineTotal
				}
				return fmt.Sprintf(`{"order_total": %.2f}`, orderTotal), nil

			default:
				return "", fmt.Errorf("unknown function: %s", functionName)
			}
		}),
	)
	if err != nil {
		panic(err)
	}

	question := `
	get the price of the Greek Salad
	get the price of the Lamb Chops

	calculate the line with quantity = 6 x the price of the Greek Salad
	calculate the line with quantity = 3 x the price of the Lamb Chops
	`
	//	calculate 6 x the price of the Greek Salad

	messagesList := []messages.Message{}

	messagesList = append(messagesList, messages.Message{
		Role:    roles.User,
		Content: question,
	})

	toolCallsResult, err := toolsAgent.DetectToolCallsLoop(
		messagesList,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tool calls result: %+v\n", toolCallsResult)



	fmt.Printf("👋 Detected %d tool calls\n", len(toolCallsResult.Results))
	for i, toolCall := range toolCallsResult.Results {
		fmt.Printf("Tool call %d: %s\n", i+1, toolCall)
	}
	//display.NewLine()
	display.Separator()

	fmt.Println(toolCallsResult.LastAssistantMessage)

}
