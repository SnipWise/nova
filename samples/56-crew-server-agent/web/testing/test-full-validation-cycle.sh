#!/bin/bash

# Test Full Validation Cycle
# Tests the complete flow from message → tool detection → validation → completion

PROXY_URL=${PROXY_URL:-http://localhost:8081}

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🧪 Testing Full Validation Cycle${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Message that triggers tool
USER_CONTENT="Say hello to TestUser"

read -r -d '' DATA <<- EOM
{
  "data": {
    "message":"${USER_CONTENT}"
  }
}
EOM

DATA=$(echo ${DATA} | tr -d '\n')

echo -e "${YELLOW}📨 Sending message: ${USER_CONTENT}${NC}"
echo ""

# Counter for events
event_count=0
tool_notification_received=false
operation_id=""
validation_sent=false
response_after_validation_received=false
finish_received=false

# Start streaming
curl --no-buffer --silent ${PROXY_URL}/completion \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d "${DATA}" \
  | while IFS= read -r line; do

    # Skip empty lines
    if [[ -z "$line" ]]; then
        continue
    fi

    if [[ $line == data:* ]]; then
        ((event_count++))
        json_data="${line#data: }"

        echo -e "${BLUE}[Event #${event_count}]${NC}"
        echo "$json_data" | jq . 2>/dev/null || echo "$json_data"
        echo ""

        # Check for tool_call notification
        if echo "$json_data" | grep -q '"kind":"tool_call"'; then
            if [[ "$tool_notification_received" == false ]]; then
                tool_notification_received=true
                echo -e "${GREEN}✓ Tool call notification received!${NC}"

                # Extract operation_id
                operation_id=$(echo "$json_data" | jq -r '.operation_id' 2>/dev/null)
                echo -e "${YELLOW}Operation ID: ${operation_id}${NC}"
                echo ""

                # Wait a bit for realism
                sleep 1

                # Send validation
                if [[ ! -z "$operation_id" ]] && [[ "$operation_id" != "null" ]]; then
                    echo -e "${CYAN}📤 Sending validation...${NC}"

                    validation_response=$(curl -s -X POST ${PROXY_URL}/operation/validate \
                        -H 'Content-Type: application/json' \
                        -d "{\"operation_id\":\"${operation_id}\"}")

                    echo "Validation response: $validation_response"
                    echo ""
                    validation_sent=true

                    echo -e "${BLUE}⏳ Waiting for response after validation...${NC}"
                    echo ""
                fi
            fi
        fi

        # Check for response after validation
        if [[ "$validation_sent" == true ]] && [[ "$response_after_validation_received" == false ]]; then
            if echo "$json_data" | grep -q '"message"'; then
                # Check if it's not the tool notification
                if ! echo "$json_data" | grep -q '"kind":"tool_call"'; then
                    response_after_validation_received=true
                    echo -e "${GREEN}✓ Response received after validation!${NC}"

                    message=$(echo "$json_data" | jq -r '.message' 2>/dev/null)
                    echo -e "${CYAN}Message: ${message}${NC}"
                    echo ""
                fi
            fi
        fi

        # Check for finish_reason
        if echo "$json_data" | grep -q '"finish_reason"'; then
            finish_reason=$(echo "$json_data" | jq -r '.finish_reason' 2>/dev/null)

            if [[ "$finish_reason" == "stop" ]]; then
                finish_received=true
                echo -e "${GREEN}✓ Stream finished with finish_reason: stop${NC}"
                echo ""
                break
            fi
        fi
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${CYAN}📊 Test Summary${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo -e "Total SSE events: ${event_count}"
echo ""

# Check results
all_passed=true

if [[ "$tool_notification_received" == true ]]; then
    echo -e "${GREEN}✓ Tool notification received${NC}"
else
    echo -e "${RED}✗ Tool notification NOT received${NC}"
    all_passed=false
fi

if [[ "$validation_sent" == true ]]; then
    echo -e "${GREEN}✓ Validation sent successfully${NC}"
else
    echo -e "${RED}✗ Validation NOT sent${NC}"
    all_passed=false
fi

if [[ "$response_after_validation_received" == true ]]; then
    echo -e "${GREEN}✓ Response received after validation${NC}"
else
    echo -e "${RED}✗ Response NOT received after validation${NC}"
    all_passed=false
    echo -e "${YELLOW}  ⚠️  This means the stream may be stuck after validation!${NC}"
fi

if [[ "$finish_received" == true ]]; then
    echo -e "${GREEN}✓ Stream finished with finish_reason: stop${NC}"
else
    echo -e "${RED}✗ finish_reason: stop NOT received${NC}"
    all_passed=false
    echo -e "${YELLOW}  ⚠️  This causes the frontend to remain in 'loading' state!${NC}"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ "$all_passed" == true ]]; then
    echo -e "${GREEN}✅ All checks passed! The flow works correctly.${NC}"
else
    echo -e "${RED}❌ Some checks failed. See issues above.${NC}"
    echo ""
    echo -e "${YELLOW}Possible causes:${NC}"
    echo "1. Backend doesn't send finish_reason after tool execution"
    echo "2. Stream closes unexpectedly"
    echo "3. Tool execution blocks the stream"
    echo ""
    echo -e "${BLUE}Check the backend logs for more information.${NC}"
fi

echo ""
