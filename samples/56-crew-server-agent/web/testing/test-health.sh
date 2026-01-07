#!/bin/bash

# Test Health Endpoint with CORS
# Usage: ./test-health.sh

PROXY_URL=${PROXY_URL:-http://localhost:8081}
ENDPOINT="${PROXY_URL}/health"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Testing Health Endpoint${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Endpoint: ${ENDPOINT}"
echo ""

# Make request and save response + headers
response=$(curl -s -i "${ENDPOINT}")

# Extract headers and body
headers=$(echo "$response" | sed -n '1,/^\r$/p')
body=$(echo "$response" | sed -n '/^\r$/,$p' | tail -n +2)

# Check CORS headers
echo -e "${BLUE}📋 CORS Headers:${NC}"
echo "$headers" | grep -i "access-control" || echo -e "${RED}✗ No CORS headers found${NC}"
echo ""

# Display response
echo -e "${BLUE}📨 Response:${NC}"
echo "$body" | jq . 2>/dev/null || echo "$body"
echo ""

# Check status
if echo "$body" | jq -e '.status == "ok"' >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
    exit 1
fi
