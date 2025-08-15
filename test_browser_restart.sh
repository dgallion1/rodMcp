#!/bin/bash

echo "Testing browser restart functionality..."
echo "========================================="

# Kill any existing rodmcp processes
pkill -f rodmcp-fixed 2>/dev/null
sleep 1

# Start the server in background
echo "Starting rodmcp server..."
./rodmcp-fixed --headless --log-level debug 2>&1 | tee rodmcp-test.log &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Send initialization and test commands
echo "Sending MCP commands..."
cat << 'EOF' | timeout 30s nc localhost 8080
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"page_id": "test", "url": "https://google.com"}}}
{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "take_screenshot", "arguments": {"page_id": "test"}}}
EOF

echo ""
echo "Test completed. Checking logs for restart behavior..."
sleep 2

# Check if restart occurred
if grep -q "Browser restarted successfully" rodmcp-test.log; then
    echo "✓ Browser restart detected in logs"
else
    echo "✗ No browser restart detected"
fi

# Check for panic recovery
if grep -q "Browser health check recovered from panic" rodmcp-test.log; then
    echo "✓ Panic recovery working"
else
    echo "✗ No panic recovery detected"
fi

# Clean up
echo "Cleaning up..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "Test complete!"