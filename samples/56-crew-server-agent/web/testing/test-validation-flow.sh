#!/bin/bash

# Test Validation Flow
# Validates the complete tool call validation workflow

PROXY_URL=${PROXY_URL:-http://localhost:8081}

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Testing Validation Flow${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Step 1: Send a message that triggers tool detection
echo -e "${YELLOW}Step 1: Sending message to trigger tool call...${NC}"
USER_CONTENT="Say hello to TestUser"

read -r -d '' DATA <<- EOM
{
  "data": {
    "message":"${USER_CONTENT}"
  }
}
EOM

DATA=$(echo ${DATA} | tr -d '\n')

# Start streaming in background and capture operation_id
operation_id=""
curl --no-buffer --silent ${PROXY_URL}/completion \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d "${DATA}" \
  | while IFS= read -r line; do
    if [[ $line == data:* ]]; then
      json_data="${line#data: }"

      # Check for tool_call notification
      if echo "$json_data" | grep -q '"kind":"tool_call"'; then
        echo -e "${GREEN}✓ Tool call notification received!${NC}"

        # Extract operation_id
        op_id=$(echo "$json_data" | jq -r '.operation_id' 2>/dev/null)

        if [ ! -z "$op_id" ] && [ "$op_id" != "null" ]; then
          echo -e "${BLUE}Operation ID: ${op_id}${NC}"
          echo ""

          # Step 2: Test validation endpoint
          echo -e "${YELLOW}Step 2: Testing validation endpoint...${NC}"

          validation_response=$(curl -s -X POST ${PROXY_URL}/operation/validate \
            -H 'Content-Type: application/json' \
            -d "{\"operation_id\":\"${op_id}\"}")

          validation_status=$?

          if [ $validation_status -eq 0 ]; then
            echo -e "${GREEN}✓ Validation request successful${NC}"
            echo "Response: $validation_response"
          else
            echo -e "${RED}✗ Validation request failed${NC}"
          fi

          echo ""
          echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
          echo -e "${GREEN}✓ Test completed${NC}"

          # Kill the stream
          break
        else
          echo -e "${RED}✗ Failed to extract operation_id${NC}"
          break
        fi
      fi

      # Check for finish
      if echo "$json_data" | grep -q '"finish_reason":"stop"'; then
        echo -e "${YELLOW}⚠️  Stream finished without tool call notification${NC}"
        break
      fi
    fi
  done &

# Wait for background process
stream_pid=$!
echo -e "${BLUE}Waiting for tool call notification...${NC}"
wait $stream_pid

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Open web interface at http://localhost:3000"
echo "2. Send message: 'Say hello to Alice'"
echo "3. Click 'Validate' button"
echo "4. Check that:"
echo "   - Card turns green with ✅"
echo "   - Message shows 'Operation validated successfully'"
echo "   - Card disappears after 3 seconds"
echo ""
echo "5. Send another message: 'Say hello to Bob'"
echo "6. Click 'Cancel' button"
echo "7. Check that:"
echo "   - Card turns red with ❌"
echo "   - Message shows 'Operation cancelled'"
echo "   - Card disappears after 3 seconds"
echo ""
