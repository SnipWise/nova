#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo -e "${CYAN}=== Testing Server Agent API ===${NC}"
echo ""

# Health check
echo -e "${YELLOW}1. Health Check${NC}"
curl -s "${BASE_URL}/health" | jq '.'
echo ""
echo ""

# Models information
echo -e "${YELLOW}2. Models Information${NC}"
curl -s "${BASE_URL}/models" | jq '.'
echo ""
echo ""

# Simple completion
echo -e "${YELLOW}3. Simple Completion (Streaming)${NC}"
echo -e "${CYAN}Question: What is the capital of France?${NC}"
curl -N -X POST "${BASE_URL}/completion" \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "message": "What is the capital of France?"
    }
  }' 2>/dev/null
echo ""
echo ""

# Get messages
echo -e "${YELLOW}4. Get Messages${NC}"
curl -s "${BASE_URL}/memory/messages/list" | jq '.messages[] | {role: .role, content: .content[:50]}'
echo ""
echo ""

# Get tokens
echo -e "${YELLOW}5. Get Token Count${NC}"
curl -s "${BASE_URL}/memory/messages/context-size" | jq '.'
echo ""
echo ""

# Completion with tool calls (requires manual validation)
echo -e "${YELLOW}6. Completion with Tool Calls${NC}"
echo -e "${RED}Note: This requires manual validation on the server console!${NC}"
echo -e "${CYAN}Sending: Say hello to Alice${NC}"
curl -N -X POST "${BASE_URL}/completion" \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "message": "Say hello to Alice"
    }
  }' 2>/dev/null
echo ""
echo ""

# Reset memory
echo -e "${YELLOW}7. Reset Memory${NC}"
curl -N -X POST "${BASE_URL}/memory/reset" \
  -H "Content-Type: application/json" 2>/dev/null
echo ""
echo ""

# Final token count
echo -e "${YELLOW}8. Final Token Count (after reset)${NC}"
curl -s "${BASE_URL}/memory/messages/context-size" | jq '.'
echo ""

echo -e "${GREEN}=== Tests Complete ===${NC}"
