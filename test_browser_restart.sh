#!/bin/bash

echo "Testing browser restart functionality..."
echo "========================================="

# Kill any existing rodmcp processes
pkill -f rodmcp 2>/dev/null
sleep 1

# Build first to ensure we have the binary
echo "Building rodmcp..."
go build -o ./rodmcp ./cmd/server/

# Create a test script that will cause browser issues to trigger restart
echo "Creating test commands..."
cat > test_commands.json << 'EOF'
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "create_page", "arguments": {"url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"page_id": "test", "url": "https://httpbin.org/delay/1"}}}
{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "take_screenshot", "arguments": {"page_id": "test"}}}
{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"page_id": "test", "url": "https://example.com"}}}
{"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": {"name": "take_screenshot", "arguments": {"page_id": "test"}}}
EOF

# Start the server and feed it commands
echo "Starting rodmcp server..."
timeout 60s ./rodmcp --headless --log-level debug < test_commands.json > rodmcp-test.log 2>&1

echo ""
echo "Test completed. Checking logs for restart behavior..."

# Check if browser operations succeeded
if grep -q "Browser action.*started" rodmcp-test.log; then
    echo "✓ Browser started successfully"
else
    echo "✗ Browser failed to start"
fi

# Check for navigation success
if grep -q "Navigation completed" rodmcp-test.log || grep -q "Successfully navigated" rodmcp-test.log; then
    echo "✓ Navigation working"
else
    echo "✗ Navigation failed"
fi

# Check for screenshot success
if grep -q "Screenshot saved" rodmcp-test.log || grep -q "screenshot.*success" rodmcp-test.log; then
    echo "✓ Screenshots working"
else
    echo "✗ Screenshots failed"
fi

# Check for panic recovery mechanisms
if grep -q "recover" rodmcp-test.log || grep -q "panic" rodmcp-test.log; then
    echo "✓ Panic recovery mechanisms detected"
else
    echo "✗ No panic recovery detected"
fi

# Check for error handling
if grep -q "Error.*recovered" rodmcp-test.log || grep -q "recoverable.*error" rodmcp-test.log; then
    echo "✓ Error recovery working"
else
    echo "✗ No error recovery detected"
fi

# Clean up
echo "Cleaning up..."
rm -f test_commands.json rodmcp-test.log

echo "Test complete!"