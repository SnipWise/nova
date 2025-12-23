#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

if [ -z "$1" ]; then
    echo -e "${RED}Usage: $0 <operation_id>${NC}"
    echo "Example: $0 op_0x12345678"
    exit 1
fi

OPERATION_ID=$1

echo -e "${GREEN}Validating operation: ${OPERATION_ID}${NC}"
curl -N -X POST http://localhost:8080/operation/validate \
  -H "Content-Type: application/json" \
  -d "{\"operation_id\":\"${OPERATION_ID}\"}"
echo ""
