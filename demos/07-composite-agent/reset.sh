#!/bin/bash
SERVICE_URL=${SERVICE_URL:-http://localhost:3500/operation/reset}

# Check if operation_id is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <operation_id>"
    echo "Example: $0 op_0x14000102300"
    exit 1
fi

OPERATION_ID=$1

read -r -d '' DATA <<- EOM
{
  "operation_id": "${OPERATION_ID}"
}
EOM

echo "Resetting operation (will stop stream): ${OPERATION_ID}"
echo ""

curl -s ${SERVICE_URL} \
  -H "Content-Type: application/json" \
  -d "${DATA}" | jq .

echo ""
