#!/bin/bash
SERVICE_URL=${SERVICE_URL:-http://localhost:8080/operation/validate}

# Check if operation_id is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <operation_id>"
    echo "Example: $0 op_0x140003dcbe0"
    exit 1
fi

OPERATION_ID=$1

read -r -d '' DATA <<- EOM
{
  "operation_id": "${OPERATION_ID}"
}
EOM

echo "Validating operation: ${OPERATION_ID}"
echo ""

curl -s ${SERVICE_URL} \
  -H "Content-Type: application/json" \
  -d "${DATA}" | sed 's/^data: //' | jq .

echo ""
