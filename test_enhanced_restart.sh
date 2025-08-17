#!/bin/bash

echo "Testing Enhanced Browser Restart Functionality"
echo "=============================================="

# Test compilation first
echo "1. Testing compilation..."
if ! go build -o /tmp/test_enhanced ./cmd/server; then
    echo "FAIL: Code does not compile"
    exit 1
fi
echo "✓ Code compiles successfully"

# Test basic functionality
echo ""
echo "2. Testing basic restart functionality..."
timeout 30s go run ./cmd/server --headless --log-level warn 2>&1 <<'EOF' | head -20
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "create_page", "arguments": {"filename": "test-page.html", "title": "Test Page", "html": "<h1>Test Page</h1><p>This is a test page.</p>"}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "browser_health", "arguments": {}}}
EOF

echo ""
echo "3. Testing context error simulation..."
# This will test the enhanced error handling
timeout 20s go run ./cmd/server --headless --log-level debug 2>&1 <<'EOF' | grep -E "(context|restart|backoff)" | head -10
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_page", "arguments": {"filename": "test-page.html", "title": "Test Page", "html": "<h1>Test Page</h1><p>This is a test page.</p>"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "take_screenshot", "arguments": {"page_id": "test"}}}
EOF

echo ""
echo "Enhanced restart functionality test completed!"
echo "=============================================="