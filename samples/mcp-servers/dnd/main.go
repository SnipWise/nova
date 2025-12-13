package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"mcp-dnd",
		"0.0.0",
	)

	// Roll dice tool
	rollDiceTool := mcp.NewTool("roll_dice",
		mcp.WithDescription("Roll n dice with n faces each"),
		mcp.WithNumber("num_dice",
			mcp.Required(),
			mcp.Description("Number of dice to roll"),
		),
		mcp.WithNumber("num_faces",
			mcp.Required(),
			mcp.Description("Number of faces on each die"),
		),
	)
	s.AddTool(rollDiceTool, rollDiceHandler)

	// Generate character name tool
	generateCharacterNameTool := mcp.NewTool("generate_character_name",
		mcp.WithDescription("Generate a D&D character name for a specific race"),
		mcp.WithString("race",
			mcp.Required(),
			mcp.Description("The race of the character (e.g., 'elf', 'dwarf', 'human', 'halfling', 'orc', 'tiefling')"),
		),
	)
	s.AddTool(generateCharacterNameTool, generateCharacterNameHandler)

	// Start the HTTP server
	httpPort := os.Getenv("MCP_HTTP_PORT")
	if httpPort == "" {
		httpPort = "9092"
	}

	log.Println("MCP 🎲 D&D Server is running on port", httpPort)

	// Create a custom mux to handle both MCP and health endpoints
	mux := http.NewServeMux()

	// Add healthcheck endpoint
	mux.HandleFunc("/health", healthCheckHandler)

	// Add MCP endpoint
	httpServer := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath("/mcp"),
	)

	// Register MCP handler with the mux
	mux.Handle("/mcp", httpServer)

	// Start the HTTP server with custom mux
	log.Fatal(http.ListenAndServe(":"+httpPort, mux))
}

func rollDiceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	numDiceArg, exists := args["num_dice"]
	if !exists || numDiceArg == nil {
		return nil, fmt.Errorf("missing required parameter 'num_dice'")
	}

	numFacesArg, exists := args["num_faces"]
	if !exists || numFacesArg == nil {
		return nil, fmt.Errorf("missing required parameter 'num_faces'")
	}

	numDice, ok := numDiceArg.(float64)
	if !ok {
		return nil, fmt.Errorf("parameter 'num_dice' must be a number")
	}

	numFaces, ok := numFacesArg.(float64)
	if !ok {
		return nil, fmt.Errorf("parameter 'num_faces' must be a number")
	}

	result := rollDice(int(numDice), int(numFaces))

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("error marshaling result: %v", err)
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

func generateCharacterNameHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	raceArg, exists := args["race"]
	if !exists || raceArg == nil {
		return nil, fmt.Errorf("missing required parameter 'race'")
	}

	race, ok := raceArg.(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'race' must be a string")
	}

	result := generateCharacterName(race)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("error marshaling result: %v", err)
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

type DiceRollResult struct {
	Rolls []int `json:"rolls"`
	Total int   `json:"total"`
}

func rollDice(numDice, numFaces int) DiceRollResult {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rolls := make([]int, numDice)
	total := 0

	for i := 0; i < numDice; i++ {
		roll := r.Intn(numFaces) + 1
		rolls[i] = roll
		total += roll
	}

	return DiceRollResult{
		Rolls: rolls,
		Total: total,
	}
}

type CharacterNameResult struct {
	Name string `json:"name"`
	Race string `json:"race"`
}

func generateCharacterName(race string) CharacterNameResult {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	namesByRace := map[string][]string{
		"elf":      {"Aerdrie", "Ahvonna", "Aramil", "Aranea", "Berrian", "Caelynn", "Carric", "Dayereth", "Enna", "Galinndan"},
		"dwarf":    {"Adrik", "Baern", "Darrak", "Eberk", "Fargrim", "Gardain", "Harbek", "Kildrak", "Morgran", "Thorek"},
		"human":    {"Aerdrie", "Aramil", "Berris", "Cithreth", "Dayereth", "Enna", "Galinndan", "Hadarai", "Immeral", "Lamlis"},
		"halfling": {"Alton", "Ander", "Bernie", "Bobbin", "Cade", "Callus", "Corrin", "Dannad", "Garret", "Lindal"},
		"orc":      {"Gash", "Gell", "Henk", "Holg", "Imsh", "Keth", "Krusk", "Mhurren", "Ront", "Shump"},
		"tiefling": {"Akmenos", "Amnon", "Barakas", "Damakos", "Ekemon", "Iados", "Kairon", "Leucis", "Melech", "Mordai"},
	}

	raceLower := strings.ToLower(race)
	names, exists := namesByRace[raceLower]
	if !exists {
		names = namesByRace["human"] // Default to human names
	}

	selectedName := names[r.Intn(len(names))]

	return CharacterNameResult{
		Name: selectedName,
		Race: race,
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status": "healthy",
		"server": "mcp-dnd-server",
	}
	json.NewEncoder(w).Encode(response)
}
