#\!/bin/bash

# Test connection stability with the fixed version
echo "Testing RodMCP Connection Stability (Fixed Version)"
echo "================================================="

# Kill any existing rodmcp processes
pkill -f rodmcp || true
sleep 2

# Start the fixed server with explicit logging
echo "Starting fixed RodMCP server..."
timeout 30s ./rodmcp-connection-fixed --headless --log-level info 2>&1 | tee fixed_server_output.log &

# Wait for server startup
sleep 3

echo "Testing multiple tool calls to check connection stability:"
echo ""

# Send multiple MCP requests to test connection stability
for i in {1..5}; do
    echo "=== Test $i ==="
    echo '{"jsonrpc": "2.0", "id": '$i', "method": "tools/list", "params": {}}'
    sleep 1
done | timeout 15s ./rodmcp-connection-fixed --headless --log-level warn 2>&1 | grep -E '(jsonrpc|error|warn|SIGPIPE|Not connected|Input stream closed|EOF)' | head -20

echo ""
echo "Check server logs for connection stability:"
tail -20 fixed_server_output.log | grep -E '(connection|EOF|SIGPIPE|closed|error)'

# Cleanup
pkill -f rodmcp-connection-fixed || true
