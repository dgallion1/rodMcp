#!/bin/bash

echo "Browser Connection Corruption Fix Test"
echo "======================================"
echo ""

# Start the enhanced server and test the specific tools that were causing timeouts
{
    echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}'
    echo '{"jsonrpc": "2.0", "method": "notifications/initialized"}'  
    sleep 0.2
    
    echo "Testing create_page (was failing with parameter validation)..."
    echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_page", "arguments": {"filename": "test.html", "title": "Test Page", "html": "<h1>Test</h1>"}}}'
    sleep 0.3
    
    echo "Testing navigate_page (was causing timeouts)..."
    echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://example.com"}}}'
    sleep 1.0
    
    echo "Testing take_screenshot..."
    echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "take_screenshot", "arguments": {"page_id": "test"}}}'
    sleep 0.5
    
} | timeout 20s /home/darrell/work/git/rodMcp/rodmcp-enhanced --headless --log-level info 2>&1 | \
    grep -E '(jsonrpc|Browser process started|Health monitoring|error|result|isError)'

echo ""
echo "Key improvements implemented:"
echo "✅ Browser PID tracking and process lifecycle management"  
echo "✅ Health monitoring with automatic corruption detection"
echo "✅ Graceful browser restart on connection failures"
echo "✅ Proper error responses instead of timeouts"
echo "✅ Panic recovery in all browser operations"