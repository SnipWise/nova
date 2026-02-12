#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Gateway Passthrough First - Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

GATEWAY_URL="http://localhost:8080/v1/chat/completions"

# Test 1: Requête AVEC tools - devrait passer par passthrough agent
echo -e "${YELLOW}📝 Test 1: Requête avec tools (passthrough first)${NC}"
echo -e "${YELLOW}Attente: Passthrough agent détecte tool_calls et répond${NC}"
echo ""

curl -s $GATEWAY_URL \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What is the weather in Paris?"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather",
          "description": "Get the current weather in a location",
          "parameters": {
            "type": "object",
            "properties": {
              "location": {
                "type": "string",
                "description": "The city and country, e.g. Paris, France"
              }
            },
            "required": ["location"]
          }
        }
      }
    ]
  }' | jq -r '.choices[0] | {finish_reason, has_tool_calls: (.message.tool_calls != null)}'

echo ""
echo -e "${GREEN}✓ Test 1 terminé${NC}"
echo ""
echo "---"
echo ""

# Test 2: Requête AVEC tools mais question simple - devrait rediriger vers agent sélectionné
echo -e "${YELLOW}📝 Test 2: Requête avec tools mais question simple (redirection)${NC}"
echo -e "${YELLOW}Attente: Passthrough ne détecte pas de tool_calls, redirige vers generic agent${NC}"
echo ""

curl -s $GATEWAY_URL \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What is 2+2?"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather",
          "description": "Get the current weather in a location",
          "parameters": {
            "type": "object",
            "properties": {
              "location": {
                "type": "string",
                "description": "The city and country"
              }
            },
            "required": ["location"]
          }
        }
      }
    ]
  }' | jq -r '.choices[0] | {finish_reason, content: .message.content[:100], has_tool_calls: (.message.tool_calls != null)}'

echo ""
echo -e "${GREEN}✓ Test 2 terminé${NC}"
echo ""
echo "---"
echo ""

# Test 3: Requête SANS tools - comportement normal
echo -e "${YELLOW}📝 Test 3: Requête sans tools (comportement normal)${NC}"
echo -e "${YELLOW}Attente: Passe directement à l'agent sélectionné par orchestrator${NC}"
echo ""

curl -s $GATEWAY_URL \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Write a simple hello world function in Go"}
    ]
  }' | jq -r '.choices[0] | {finish_reason, content: .message.content[:150]}'

echo ""
echo -e "${GREEN}✓ Test 3 terminé${NC}"
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}  Tous les tests terminés !${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}💡 Consultez les logs du serveur pour voir le traçage complet${NC}"
