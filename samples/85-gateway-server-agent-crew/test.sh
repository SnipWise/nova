#!/bin/bash
# ============================================================
# Manual test script for 85-gateway-server-agent-crew
# Multi-agent crew with orchestrator, compressor, tools
# in auto-execute mode on port 8080
#
# Prerequisites:
#   - LLM engine running on localhost:12434
#   - Gateway crew server running: go run main.go
# ============================================================

BASE_URL="http://localhost:8080"
CONTENT_TYPE="Content-Type: application/json"

echo "============================================"
echo " Gateway Crew Server Agent - Manual Tests"
echo " Target: ${BASE_URL}"
echo "============================================"
echo ""

# ----------------------------------------------------------
# 1. Health check
# ----------------------------------------------------------
echo "--- 1. Health Check ---"
curl -s "${BASE_URL}/health" | jq .
echo ""

# ----------------------------------------------------------
# 2. List models (should show all crew agents)
# ----------------------------------------------------------
echo "--- 2. List Models (crew agents) ---"
curl -s "${BASE_URL}/v1/models" | jq .
echo ""

# ----------------------------------------------------------
# 3. Non-streaming completion (generic agent)
# ----------------------------------------------------------
echo "--- 3. Non-Streaming - Generic Topic ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What is the capital of France?"}
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 4. Non-streaming - coding topic (should route to coder)
# ----------------------------------------------------------
echo "--- 4. Non-Streaming - Coding Topic (coder agent) ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Write a Go function that reverses a string."}
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 5. Non-streaming - philosophy topic (should route to thinker)
# ----------------------------------------------------------
echo "--- 5. Non-Streaming - Philosophy Topic (thinker agent) ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What is the meaning of life according to Stoic philosophy?"}
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 6. Streaming completion (SSE)
# ----------------------------------------------------------
echo "--- 6. Streaming Completion (SSE) ---"
curl -sN "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Explain recursion in one paragraph."}
    ],
    "stream": true
  }'
echo ""
echo ""

# ----------------------------------------------------------
# 7. Auto-execute tool call (calculate_sum)
#    In auto-execute mode, the server runs the tool
#    and returns the final answer directly.
# ----------------------------------------------------------
echo "--- 7. Auto-Execute Tool Call (calculate_sum) ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What is 42 + 58?"}
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 8. Streaming auto-execute tool call
# ----------------------------------------------------------
echo "--- 8. Streaming Auto-Execute Tool Call ---"
curl -sN "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Calculate the sum of 100 and 200."}
    ],
    "stream": true
  }'
echo ""
echo ""

# ----------------------------------------------------------
# 9. Multi-turn conversation with crew
# ----------------------------------------------------------
echo "--- 9. Multi-Turn Conversation ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "I am learning Go programming."},
      {"role": "assistant", "content": "Go is an excellent choice! It is a statically typed, compiled language with great concurrency support."},
      {"role": "user", "content": "How do I create a goroutine?"}
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 10. Request a specific agent by model name
# ----------------------------------------------------------
echo "--- 10. Request Specific Agent (coder) ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "coder",
    "messages": [
      {"role": "user", "content": "Write a simple HTTP handler in Go."}
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 11. Request with system message
# ----------------------------------------------------------
echo "--- 11. Request with System Message ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "system", "content": "You must reply in French only."},
      {"role": "user", "content": "What is 2 + 2?"}
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 12. Error handling - missing messages
# ----------------------------------------------------------
echo "--- 12. Error Handling - Missing Messages ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew"
  }' | jq .
echo ""

# ----------------------------------------------------------
# 13. Error handling - invalid JSON
# ----------------------------------------------------------
echo "--- 13. Error Handling - Invalid JSON ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d 'this is not json' | jq .
echo ""

echo "============================================"
echo " All tests completed."
echo "============================================"
