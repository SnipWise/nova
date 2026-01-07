#!/bin/bash

# Test SSE Flushing
# Vérifie que le proxy flush correctement les événements SSE

PROXY_URL=${PROXY_URL:-http://localhost:8081}
ENDPOINT="${PROXY_URL}/completion"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Testing SSE Flushing${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Endpoint: ${ENDPOINT}"
echo ""

# Send a message that will trigger tool detection
USER_CONTENT="Say hello to TestUser"

read -r -d '' DATA <<- EOM
{
  "data": {
    "message":"${USER_CONTENT}"
  }
}
EOM

DATA=$(echo ${DATA} | tr -d '\n')

echo -e "${YELLOW}Message: ${USER_CONTENT}${NC}"
echo -e "${BLUE}This should trigger tool detection...${NC}"
echo ""
echo -e "${BLUE}📡 Watching SSE stream:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

event_count=0
tool_call_found=false
start_time=$(date +%s)

curl --no-buffer --silent ${ENDPOINT} \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d "${DATA}" \
  | while IFS= read -r line; do
    current_time=$(date +%s)
    elapsed=$((current_time - start_time))

    if [[ $line == data:* ]]; then
      ((event_count++))
      json_data="${line#data: }"

      # Check for tool_call notification
      if echo "$json_data" | grep -q '"kind":"tool_call"'; then
        echo -e "${GREEN}✓ [${elapsed}s] Tool call notification received!${NC}"
        echo "$json_data" | jq . 2>/dev/null || echo "$json_data"
        tool_call_found=true

        # Extract operation_id
        op_id=$(echo "$json_data" | jq -r '.operation_id' 2>/dev/null)
        echo ""
        echo -e "${YELLOW}⚠️  Operation ID: ${op_id}${NC}"
        echo -e "${YELLOW}⚠️  Waiting for validation...${NC}"
        echo ""
        echo -e "${BLUE}To validate, run:${NC}"
        echo "  curl -X POST ${PROXY_URL}/operation/validate \\"
        echo "    -H 'Content-Type: application/json' \\"
        echo "    -d '{\"operation_id\":\"${op_id}\"}'"
        echo ""
      fi

      # Check for finish
      if echo "$json_data" | grep -q '"finish_reason":"stop"'; then
        echo -e "${GREEN}✓ [${elapsed}s] Stream completed${NC}"
        break
      fi

      # Show first few events
      if [ $event_count -le 5 ]; then
        echo -e "${BLUE}[${elapsed}s] Event #${event_count}${NC}"
      fi
    fi
  done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ "$tool_call_found" = true ]; then
    echo -e "${GREEN}✓ SUCCESS: Tool call notification was received in real-time${NC}"
    echo -e "${GREEN}  This means SSE flushing is working correctly!${NC}"
else
    echo -e "${RED}✗ FAILED: No tool call notification received${NC}"
    echo -e "${RED}  The proxy may not be flushing SSE events correctly${NC}"
fi

echo ""
