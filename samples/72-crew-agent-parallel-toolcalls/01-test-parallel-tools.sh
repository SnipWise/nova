#!/bin/bash
# Test parallel tool calls:
# The LLM should detect multiple tools at once (add + multiply + hello)
# and execute them all in a single pass

curl -X POST http://localhost:3500/completion \
  -H "Content-Type: application/json" \
  -d '{"data":{"message":"Calculate 40+2, multiply 6*7, and say hello to Alice"}}' \
  --no-buffer

echo ""
