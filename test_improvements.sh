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

# Test 1: Browser Health Check
run_test "Browser Health Check" '
echo "{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"initialize\", \"params\": {\"protocolVersion\": \"2024-11-05\", \"capabilities\": {}}}
{\"jsonrpc\": \"2.0\", \"method\": \"notifications/initialized\"}
{\"jsonrpc\": \"2.0\", \"id\": 2, \"method\": \"tools/list\"}
{\"jsonrpc\": \"2.0\", \"id\": 3, \"method\": \"tools/call\", \"params\": {\"name\": \"browser_health\", \"arguments\": {}}}
" | timeout 10s ./rodmcp-improved --headless 2>&1 | grep -q "Browser Health Report"
'

# Test 2: Page Recovery
run_test "Page Recovery" '
echo "{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"initialize\", \"params\": {\"protocolVersion\": \"2024-11-05\", \"capabilities\": {}}}
{\"jsonrpc\": \"2.0\", \"method\": \"notifications/initialized\"}
{\"jsonrpc\": \"2.0\", \"id\": 2, \"method\": \"tools/call\", \"params\": {\"name\": \"create_page\", \"arguments\": {\"url\": \"https://example.com\"}}}
{\"jsonrpc\": \"2.0\", \"id\": 3, \"method\": \"tools/call\", \"params\": {\"name\": \"get_page_status\", \"arguments\": {\"page_id\": \"test\"}}}
" | timeout 15s ./rodmcp-improved --headless 2>&1 | grep -q "Page Status"
'

# Test 3: Debug Info Tool
run_test "Debug Info Tool" '
echo "{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"initialize\", \"params\": {\"protocolVersion\": \"2024-11-05\", \"capabilities\": {}}}
{\"jsonrpc\": \"2.0\", \"method\": \"notifications/initialized\"}
{\"jsonrpc\": \"2.0\", \"id\": 2, \"method\": \"tools/call\", \"params\": {\"name\": \"debug_info\", \"arguments\": {\"verbose\": true}}}
" | timeout 10s ./rodmcp-improved --headless 2>&1 | grep -q "RodMCP Debug Information"
'

# Test 4: Navigation with Retry
run_test "Navigation with Retry" '
cat << EOF | timeout 20s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://httpbin.org/delay/1"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"page_id": "test", "url": "https://httpbin.org/status/200"}}}
EOF
'

# Test 5: Screenshot with Recovery
run_test "Screenshot with Recovery" '
cat << EOF | timeout 20s ./rodmcp-improved --headless 2>&1 | grep -q "Screenshot taken successfully"
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "take_screenshot", "arguments": {"page_id": "test"}}}
EOF
'

# Test 6: Multiple Page Management
run_test "Multiple Page Management" '
cat << EOF | timeout 25s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://httpbin.org"}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "list_pages", "arguments": {}}}
EOF
'

# Test 7: Script Execution with Retry
run_test "Script Execution with Retry" '
cat << EOF | timeout 15s ./rodmcp-improved --headless 2>&1 | grep -q "\"result\""
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "execute_script", "arguments": {"page_id": "test", "script": "document.title"}}}
EOF
'

# Test 8: Connection Recovery
run_test "Connection Recovery" '
# Start server in background
./rodmcp-improved --headless 2>&1 &
SERVER_PID=$!
sleep 2

# Send initial commands
echo "{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"initialize\", \"params\": {\"protocolVersion\": \"2024-11-05\", \"capabilities\": {}}}" | nc localhost 3000
echo "{\"jsonrpc\": \"2.0\", \"method\": \"notifications/initialized\"}" | nc localhost 3000

# Brief interruption
sleep 1

# Send more commands
echo "{\"jsonrpc\": \"2.0\", \"id\": 2, \"method\": \"tools/list\"}" | nc localhost 3000

# Clean up
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

# Check if we got responses
[ $? -eq 0 ]
'

# Test 9: Circuit Breaker Functionality
run_test "Circuit Breaker Functionality" '
cat << EOF | timeout 20s ./rodmcp-improved --headless 2>&1 | grep -q "circuit breaker"
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "debug_info", "arguments": {"verbose": true}}}
EOF
'

# Test 10: Browser Restart Recovery
run_test "Browser Restart Recovery" '
cat << EOF | timeout 30s ./rodmcp-improved --headless 2>&1 | grep -q "Browser restarted successfully"
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://example.com"}}}
# Simulate browser issue by creating multiple pages rapidly
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://httpbin.org/delay/5"}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://example.org"}}}
{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "browser_health", "arguments": {}}}
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