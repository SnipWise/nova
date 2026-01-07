#!/bin/bash

# Run All Tests
# Usage: ./run-all-tests.sh

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}🧪 Running All API Tests${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if proxy is running
if ! curl -s http://localhost:8081/health > /dev/null 2>&1; then
    echo -e "${RED}✗ Error: CORS Proxy is not running on port 8081${NC}"
    echo ""
    echo "Please start the proxy:"
    echo "  cd samples/56-crew-server-agent/web/proxy"
    echo "  go run main.go"
    echo ""
    exit 1
fi

tests_passed=0
tests_failed=0

run_test() {
    local test_name=$1
    local test_script=$2

    echo ""
    echo -e "${YELLOW}▶ Running: ${test_name}${NC}"
    echo ""

    if bash "$test_script"; then
        ((tests_passed++))
        echo -e "${GREEN}✓ ${test_name} passed${NC}"
    else
        ((tests_failed++))
        echo -e "${RED}✗ ${test_name} failed${NC}"
    fi

    sleep 1
}

# Run tests in order
run_test "Health Check" "./test-health.sh"
run_test "Models Info" "./test-models.sh"
run_test "Streaming" "./test-stream.sh"

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}📊 Test Summary${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ Passed: ${tests_passed}${NC}"
echo -e "${RED}✗ Failed: ${tests_failed}${NC}"
echo ""

if [ $tests_failed -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some tests failed${NC}"
    exit 1
fi
