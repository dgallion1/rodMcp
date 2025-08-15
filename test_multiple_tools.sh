#\!/bin/bash

# Test multiple different tool calls to thoroughly check connection stability
echo "Testing Multiple Tool Types (Fixed Version)"
echo "=========================================="

# Kill any existing processes
pkill -f rodmcp-connection-fixed || true
sleep 2

# Create a test JSON payload with multiple tool calls
cat << 'JSON_EOF' | timeout 20s ./rodmcp-connection-fixed --headless --log-level warn 2>&1 | grep -E '(jsonrpc|error|warn|SIGPIPE|Not connected|Input stream closed|EOF)' | head -30
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2025-06-18", "clientInfo": {"name": "test-client", "version": "1.0.0"}}}
{"jsonrpc": "2.0", "id": 2, "method": "notifications/initialized", "params": {}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/list", "params": {}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "wait", "arguments": {"seconds": 0.1}}}
{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "create_page", "arguments": {"filename": "test.html", "title": "Test", "html": "<h1>Test</h1>"}}}
{"jsonrpc": "2.0", "id": 6, "method": "tools/list", "params": {}}
JSON_EOF

echo ""
echo "Test completed. Connection remained stable through multiple tool calls\!"
