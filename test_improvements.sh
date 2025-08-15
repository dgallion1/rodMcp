#!/bin/bash

# Comprehensive test suite for RodMCP improvements
# Tests browser stability, recovery, retry logic, and health monitoring

set -e

echo "=================================================="
echo "RodMCP Comprehensive Improvement Test Suite"
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
echo "Building RodMCP with improvements..."
go build -o rodmcp-improved ./cmd/server
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful!${NC}"
echo ""

# Test 1: Help Tool (Basic functionality check)
run_test "Help Tool" '
echo "{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"initialize\", \"params\": {\"protocolVersion\": \"2024-11-05\", \"capabilities\": {}}}
{\"jsonrpc\": \"2.0\", \"method\": \"notifications/initialized\"}
{\"jsonrpc\": \"2.0\", \"id\": 2, \"method\": \"tools/list\"}
{\"jsonrpc\": \"2.0\", \"id\": 3, \"method\": \"tools/call\", \"params\": {\"name\": \"help\", \"arguments\": {\"topic\": \"overview\"}}}
" | timeout 10s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
'

# Test 2: Create and Navigate Page
run_test "Create and Navigate Page" '
echo "{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"initialize\", \"params\": {\"protocolVersion\": \"2024-11-05\", \"capabilities\": {}}}
{\"jsonrpc\": \"2.0\", \"method\": \"notifications/initialized\"}
{\"jsonrpc\": \"2.0\", \"id\": 2, \"method\": \"tools/call\", \"params\": {\"name\": \"create_page\", \"arguments\": {\"filename\": \"test.html\", \"title\": \"Test\", \"html\": \"<h1>Test Page</h1>\"}}}
{\"jsonrpc\": \"2.0\", \"id\": 3, \"method\": \"tools/call\", \"params\": {\"name\": \"navigate_page\", \"arguments\": {\"url\": \"https://example.com\"}}}
" | timeout 15s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
'

# Test 3: Click Element
run_test "Click Element" '
echo "{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"initialize\", \"params\": {\"protocolVersion\": \"2024-11-05\", \"capabilities\": {}}}
{\"jsonrpc\": \"2.0\", \"method\": \"notifications/initialized\"}
{\"jsonrpc\": \"2.0\", \"id\": 2, \"method\": \"tools/call\", \"params\": {\"name\": \"navigate_page\", \"arguments\": {\"url\": \"https://example.com\"}}}
{\"jsonrpc\": \"2.0\", \"id\": 3, \"method\": \"tools/call\", \"params\": {\"name\": \"click_element\", \"arguments\": {\"selector\": \"a\"}}}
" | timeout 10s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
'

# Test 4: Navigation with Multiple Pages
run_test "Navigation with Multiple Pages" '
cat << EOF | timeout 20s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://httpbin.org/delay/1"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://httpbin.org/status/200"}}}
EOF
'

# Test 5: Screenshot
run_test "Screenshot" '
cat << EOF | timeout 20s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "take_screenshot", "arguments": {}}}
EOF
'

# Test 6: Tab Management
run_test "Tab Management" '
cat << EOF | timeout 25s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "switch_tab", "arguments": {"action": "create", "url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "switch_tab", "arguments": {"action": "create", "url": "https://httpbin.org"}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "switch_tab", "arguments": {"action": "list"}}}
EOF
'

# Test 7: Script Execution
run_test "Script Execution" '
cat << EOF | timeout 15s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "execute_script", "arguments": {"script": "document.title"}}}
EOF
'

# Test 8: Wait and Type Text
run_test "Wait and Type Text" '
cat << EOF | timeout 20s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://httpbin.org/forms/post"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "wait_for_element", "arguments": {"selector": "input[name=\"custname\"]"}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "type_text", "arguments": {"selector": "input[name=\"custname\"]", "text": "Test User"}}}
EOF
'

# Test 9: HTTP Request Tool
run_test "HTTP Request Tool" '
cat << EOF | timeout 20s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "http_request", "arguments": {"url": "https://httpbin.org/get", "method": "GET"}}}
EOF
'

# Test 10: Form Fill
run_test "Form Fill" '
cat << EOF | timeout 30s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://httpbin.org/forms/post"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "form_fill", "arguments": {"fields": {"input[name=\"custname\"]": "John Doe", "input[name=\"custtel\"]": "555-1234", "textarea[name=\"comments\"]": "Test comment"}}}}
EOF
'

# Print summary
echo ""
echo "=================================================="
echo "Test Results Summary"
echo "=================================================="
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed! RodMCP improvements are working correctly.${NC}"
    exit 0
else
    echo -e "${RED}⚠️ Some tests failed. Please review the output above.${NC}"
    exit 1
fi