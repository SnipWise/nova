#!/bin/bash

# Test Models Endpoint with CORS
# Usage: ./test-models.sh

PROXY_URL=${PROXY_URL:-http://localhost:8081}
ENDPOINT="${PROXY_URL}/models"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Testing Models Endpoint${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Endpoint: ${ENDPOINT}"
echo ""

response=$(curl -s "${ENDPOINT}")

echo -e "${BLUE}📨 Models Information:${NC}"
echo "$response" | jq . 2>/dev/null || echo "$response"
echo ""

# Check if models are present
if echo "$response" | jq -e '.chat_model' >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Models retrieved successfully${NC}"
else
    echo -e "${RED}✗ Failed to retrieve models${NC}"
    exit 1
fi
