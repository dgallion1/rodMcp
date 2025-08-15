#!/bin/bash

# Connection stability test for RodMCP
echo "Testing RodMCP connection stability fixes..."
echo "======================================================"

# Start the server with headless mode and debug logging
echo "Starting RodMCP server with debug logging..."
timeout 60s ./rodmcp-fixed --headless --log-level debug --log-dir ./test-logs > server_output.log 2>&1 &
SERVER_PID=$!

# Give server time to start
sleep 2

echo "Server PID: $SERVER_PID"

# Test 1: Initialize connection
echo "Test 1: Testing MCP initialization..."
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}' | timeout 5s ./rodmcp-fixed --headless &

sleep 2

# Test 2: List tools
echo "Test 2: Testing tools/list..."
{
    echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}'
    echo '{"jsonrpc": "2.0", "method": "notifications/initialized"}'
    echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}'
    sleep 1
} | timeout 10s ./rodmcp-fixed --headless --log-level debug 2>&1 | tee test_output.log &

sleep 5

# Test 3: Multiple tool calls to test connection stability
echo "Test 3: Testing multiple tool calls (connection stability)..."
{
    echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}'
    echo '{"jsonrpc": "2.0", "method": "notifications/initialized"}'
    echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}'
    sleep 0.5
    echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "overview"}}}'
    sleep 0.5
    echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "wait", "arguments": {"seconds": 0.1}}}'
    sleep 0.5
    echo '{"jsonrpc": "2.0", "id": 5, "method": "tools/list"}'
    sleep 2
} | timeout 15s ./rodmcp-fixed --headless --log-level info 2>&1 | tee multiple_tools_test.log

echo "======================================================"
echo "Connection stability test completed."
echo "Check the following files for results:"
echo "  - server_output.log (server startup and operation)"
echo "  - test_output.log (basic tools/list test)"
echo "  - multiple_tools_test.log (multiple tool calls test)"
echo "  - ./test-logs/rodmcp.log (detailed server logs)"

# Kill any remaining processes
pkill -f rodmcp-fixed || true

echo ""
echo "Summary:"
echo "- If tests show JSON responses without connection errors, the fix is working"
echo "- Look for 'Input stream closed gracefully' followed by continued operation"
echo "- No 'Not connected' errors should appear after initialization"