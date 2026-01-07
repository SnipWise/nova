#!/bin/bash

# Test Streaming Completion with CORS
# Usage: ./test-stream.sh

PROXY_URL=${PROXY_URL:-http://localhost:8081}
ENDPOINT="${PROXY_URL}/completion"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

USER_CONTENT=${1:-"Explain what is Vue.js in 2 sentences"}

read -r -d '' DATA <<- EOM
{
  "data": {
    "message":"${USER_CONTENT}"
  }
}
EOM

DATA=$(echo ${DATA} | tr -d '\n')

echo -e "${BLUE}🧪 Testing Streaming Completion${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Endpoint: ${ENDPOINT}"
echo -e "Question: ${YELLOW}${USER_CONTENT}${NC}"
echo ""
echo -e "${BLUE}📡 Streaming response:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

callback() {
  echo -ne "$1"
}

unescape_quotes() {
    local str="$1"
    str="${str//\\\"/\"}"
    echo "$str"
}

remove_quotes() {
    local str="$1"
    str="${str%\"}"
    str="${str#\"}"
    echo "$str"
}

curl --no-buffer --silent ${ENDPOINT} \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d "${DATA}" \
  | while IFS= read -r line; do
    if [[ $line == data:* ]]; then
      json_data="${line#data: }"
      content_chunk=$(echo "$json_data" | jq '.message // "null"' 2>/dev/null)
      result=$(remove_quotes "$content_chunk")
      clean_result=$(unescape_quotes "$result")
      callback "$clean_result"

      # Check for finish
      finish_reason=$(echo "$json_data" | jq -r '.finish_reason // ""' 2>/dev/null)
      if [[ "$finish_reason" == "stop" ]]; then
        echo ""
        echo ""
        echo -e "${GREEN}✓ Stream completed${NC}"
        break
      fi
    fi
  done

echo ""
