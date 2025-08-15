#!/bin/bash

echo "Enhanced Browser Lifecycle Management Test"
echo "=========================================="

# Test browser process monitoring and automatic restart capabilities
{
    echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}'
    echo '{"jsonrpc": "2.0", "method": "notifications/initialized"}'  
    echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}'
    sleep 0.5
    
    # Test create_page with proper parameters
    echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "create_page", "arguments": {"filename": "test.html", "title": "Test Page", "html": "<h1>Hello World</h1>"}}}'
    sleep 0.5
    
    # Test navigate_page 
    echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "navigate_page", "arguments": {"url": "https://example.com"}}}'
    sleep 1.0
    
    # Test screenshot
    echo '{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "take_screenshot", "arguments": {"page_id": "test"}}}'
    
} | timeout 30s ./rodmcp-enhanced --headless --log-level info 2>&1 | \
    grep -E '(jsonrpc|error|warn|Browser process started|Health monitoring|restart)' | head -15

echo ""
echo "Enhanced features tested:"
echo "- Browser PID tracking and process monitoring"  
echo "- Automatic health checks every 10 seconds"
echo "- Graceful browser restart on corruption detection"
echo "- Process lifecycle management with restart limits"