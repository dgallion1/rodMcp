#!/bin/bash

echo "Quick Connection Stability Test"
echo "==============================="

# Test multiple tool calls without excessive logging
{
    echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}'
    echo '{"jsonrpc": "2.0", "method": "notifications/initialized"}'  
    echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}'
    sleep 0.2
    echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "wait", "arguments": {"seconds": 0.1}}}'
    sleep 0.2
    echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "overview"}}}'
} | timeout 10s ./rodmcp-final --headless --log-level warn 2>&1 | grep -E '(jsonrpc|error|warn|SIGPIPE|Not connected)' | head -10

echo ""
echo "If you see JSON responses above without 'Not connected' errors, the fix is working!"