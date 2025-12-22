#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}=== Remote Agent Demo ===${NC}"
echo ""

# Check if server is running
echo -e "${YELLOW}Checking if server is running...${NC}"
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Server is running${NC}"
else
    echo -e "${RED}✗ Server is not running${NC}"
    echo ""
    echo -e "${YELLOW}Please start the server first:${NC}"
    echo -e "  cd ../49-server-agent-with-tools"
    echo -e "  go run main.go"
    echo ""
    exit 1
fi

echo ""
echo -e "${CYAN}Starting remote agent client...${NC}"
echo ""

go run main.go
