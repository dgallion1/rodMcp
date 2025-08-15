#!/bin/bash

# Quick test suite for RodMCP improvements
# Tests basic functionality with existing tools

set -e

echo "=================================================="
echo "RodMCP Quick Test Suite"
echo "=================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -e "${YELLOW}Running test: $test_name${NC}"
    if eval "$test_command"; then
        echo -e "${GREEN}✅ PASSED: $test_name${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED: $test_name${NC}"
        ((TESTS_FAILED++))
    fi
    echo ""
}

# Build the binary first
echo "Building RodMCP..."
go build -o rodmcp-test ./cmd/server
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful!${NC}"
echo ""

# Test 1: Basic initialization and tool listing
run_test "Initialization and Tool List" '
echo "{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"initialize\", \"params\": {\"protocolVersion\": \"2024-11-05\", \"capabilities\": {}}}
{\"jsonrpc\": \"2.0\", \"method\": \"notifications/initialized\"}
{\"jsonrpc\": \"2.0\", \"id\": 2, \"method\": \"tools/list\"}" | timeout 5s ./rodmcp-test --headless 2>&1 | grep -q "navigate_page"
'

# Test 2: Help tool
run_test "Help Tool" '
cat << EOF | timeout 5s ./rodmcp-test --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "overview"}}}
EOF
'

# Test 3: Navigate to a simple page
run_test "Navigation" '
cat << EOF | timeout 10s ./rodmcp-test --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://example.com"}}}
EOF
'

# Test 4: Create HTML page
run_test "Create Page" '
cat << EOF | timeout 10s ./rodmcp-test --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_page", "arguments": {"filename": "test.html", "title": "Test Page", "html": "<h1>Hello World</h1>"}}}
EOF
'

# Test 5: Execute simple script
run_test "Execute Script" '
cat << EOF | timeout 10s ./rodmcp-test --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "execute_script", "arguments": {"script": "1 + 1"}}}
EOF
'

# Clean up
rm -f rodmcp-test test.html

# Print summary
echo ""
echo "=================================================="
echo "Test Results Summary"
echo "=================================================="
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}⚠️ Some tests failed. Please review the output above.${NC}"
    exit 1
fi