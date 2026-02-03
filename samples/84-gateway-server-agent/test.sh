#!/bin/bash
# ============================================================
# Manual test script for 84-gateway-server-agent
# Single agent (assistant) in passthrough mode on port 8080
#
# Prerequisites:
#   - LLM engine running on localhost:12434
#   - Gateway server running: go run main.go
# ============================================================

BASE_URL="http://localhost:8080"
CONTENT_TYPE="Content-Type: application/json"

echo "============================================"
echo " Gateway Server Agent - Manual Tests"
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
# 2. List models
# ----------------------------------------------------------
echo "--- 2. List Models ---"
curl -s "${BASE_URL}/v1/models" | jq .
echo ""

# ----------------------------------------------------------
# 3. Non-streaming completion
# ----------------------------------------------------------
echo "--- 3. Non-Streaming Completion ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "assistant",
    "messages": [
      {"role": "user", "content": "Say hello in French in one sentence."}
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 4. Streaming completion (SSE)
# ----------------------------------------------------------
echo "--- 4. Streaming Completion (SSE) ---"
curl -sN "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "assistant",
    "messages": [
      {"role": "user", "content": "Count from 1 to 5."}
    ],
    "stream": true
  }'
echo ""
echo ""

# ----------------------------------------------------------
# 5. Non-streaming with temperature
# ----------------------------------------------------------
echo "--- 5. Non-Streaming with Temperature ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "assistant",
    "messages": [
      {"role": "user", "content": "Tell me a one-line joke."}
    ],
    "stream": false,
    "temperature": 1.0
  }' | jq .
echo ""

# ----------------------------------------------------------
# 6. Multi-turn conversation
# ----------------------------------------------------------
echo "--- 6. Multi-Turn Conversation ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "assistant",
    "messages": [
      {"role": "user", "content": "My name is Philippe."},
      {"role": "assistant", "content": "Hello Philippe! Nice to meet you."},
      {"role": "user", "content": "What is my name?"}
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 7. Passthrough with tools (client-managed tool_calls)
# ----------------------------------------------------------
echo "--- 7. Passthrough Tool Call (client sends tools) ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "assistant",
    "messages": [
      {"role": "user", "content": "What is 3 + 5?"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "calculate_sum",
          "description": "Calculate the sum of two numbers",
          "parameters": {
            "type": "object",
            "properties": {
              "a": {"type": "number", "description": "The first number"},
              "b": {"type": "number", "description": "The second number"}
            },
            "required": ["a", "b"]
          }
        }
      }
    ],
    "stream": false
  }' | jq .
echo ""

# ----------------------------------------------------------
# 8. Tool result submission (simulate tool round-trip)
#    If test 7 returned tool_calls, copy the tool_call_id
#    and send a tool result message back.
# ----------------------------------------------------------
echo "--- 8. Tool Result Submission (template) ---"
echo "If test 7 returned tool_calls, use this pattern:"
echo ""
cat <<'TEMPLATE'
curl -s http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "assistant",
    "messages": [
      {"role": "user", "content": "What is 3 + 5?"},
      {"role": "assistant", "content": null, "tool_calls": [
        {
          "id": "call_REPLACE_ME",
          "type": "function",
          "function": {"name": "calculate_sum", "arguments": "{\"a\":3,\"b\":5}"}
        }
      ]},
      {"role": "tool", "tool_call_id": "call_REPLACE_ME", "content": "{\"result\": 8}"}
    ],
    "stream": false
  }' | jq .
TEMPLATE
echo ""

# ----------------------------------------------------------
# 9. Streaming with tools (passthrough)
# ----------------------------------------------------------
echo "--- 9. Streaming Tool Call (passthrough) ---"
curl -sN "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "assistant",
    "messages": [
      {"role": "user", "content": "Calculate 10 + 20"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "calculate_sum",
          "description": "Calculate the sum of two numbers",
          "parameters": {
            "type": "object",
            "properties": {
              "a": {"type": "number", "description": "The first number"},
              "b": {"type": "number", "description": "The second number"}
            },
            "required": ["a", "b"]
          }
        }
      }
    ],
    "stream": true
  }'
echo ""
echo ""

# ----------------------------------------------------------
# 10. Invalid request (missing messages)
# ----------------------------------------------------------
echo "--- 10. Error Handling - Missing Messages ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "assistant"
  }' | jq .
echo ""

echo "============================================"
echo " All tests completed."
echo "============================================"
