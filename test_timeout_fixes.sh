#!/bin/bash

echo "Testing timeout fixes validation..."
echo "=================================="

# Test that server responds quickly for basic operations
test_server_responsiveness() {
    echo "Testing server responsiveness..."
    
    local start_time=$(date +%s.%N)
    
    # Test basic JSON-RPC functionality without browser tools
    local output=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "overview"}}}' | timeout 15s ./rodmcp-enhanced --headless --log-level error 2>&1)
    
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    local duration_int=$(printf "%.0f" "$duration")
    
    if echo "$output" | grep -q '"result"' && echo "$output" | grep -q '"content"'; then
        echo "✅ Server responds in ${duration_int}s - no hanging detected"
        return 0
    else
        echo "❌ Server failed to respond within 15s"
        return 1
    fi
}

# Test error handling for invalid tool
test_error_handling() {
    echo "Testing error handling..."
    
    local output=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "nonexistent_tool", "arguments": {}}}' | timeout 10s ./rodmcp-enhanced --headless --log-level error 2>&1)
    
    if echo "$output" | grep -q '"error"' && echo "$output" | grep -q '"code":-32601'; then
        echo "✅ Error handling works correctly"
        return 0
    else
        echo "❌ Error handling failed"
        return 1
    fi
}

# Test that server shuts down gracefully (no infinite loops)
test_graceful_shutdown() {
    echo "Testing graceful shutdown..."
    
    # Start server in background
    echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}' | timeout 10s ./rodmcp-enhanced --headless --log-level error &
    local server_pid=$!
    
    # Wait a moment for server to start
    sleep 2
    
    # Check if server is still running
    if kill -0 $server_pid 2>/dev/null; then
        # Send SIGTERM and see if it shuts down gracefully
        kill -TERM $server_pid 2>/dev/null
        
        # Wait up to 5 seconds for graceful shutdown
        local countdown=5
        while [ $countdown -gt 0 ] && kill -0 $server_pid 2>/dev/null; do
            sleep 1
            countdown=$((countdown - 1))
        done
        
        if kill -0 $server_pid 2>/dev/null; then
            echo "❌ Server did not shut down gracefully"
            kill -KILL $server_pid 2>/dev/null
            return 1
        else
            echo "✅ Server shut down gracefully"
            return 0
        fi
    else
        echo "❌ Server failed to start"
        return 1
    fi
}

# Validate timeout configuration in source code
test_timeout_configuration() {
    echo "Testing timeout configuration..."
    
    # Check if ReadTimeout was properly configured
    if grep -q "ReadTimeout.*5.*time.Minute" internal/mcp/server.go; then
        echo "✅ ReadTimeout properly configured to 5 minutes"
    else
        echo "❌ ReadTimeout not properly configured"
        return 1
    fi
    
    # Check if tool timeout was implemented
    if grep -q "context.WithTimeout.*30.*time.Second" internal/mcp/server.go; then
        echo "✅ Tool execution timeout properly configured to 30 seconds"
    else
        echo "❌ Tool execution timeout not properly configured"
        return 1
    fi
    
    # Check if consecutive timeout protection was added
    if grep -q "consecutiveTimeouts" internal/mcp/server.go; then
        echo "✅ Consecutive timeout protection implemented"
    else
        echo "❌ Consecutive timeout protection not implemented"
        return 1
    fi
    
    return 0
}

# Build server
echo "Building MCP server..."
if ! go build -o rodmcp-enhanced ./cmd/server 2>/dev/null; then
    echo "❌ Failed to build MCP server"
    exit 1
fi

echo ""

# Run tests
test_timeout_configuration
config_result=$?

echo ""

test_server_responsiveness
response_result=$?

echo ""

test_error_handling
error_result=$?

echo ""

test_graceful_shutdown
shutdown_result=$?

echo ""
echo "Summary:"
echo "========"

if [ $config_result -eq 0 ]; then
    echo "✅ Timeout configuration: PASS"
else
    echo "❌ Timeout configuration: FAIL"
fi

if [ $response_result -eq 0 ]; then
    echo "✅ Server responsiveness: PASS"
else
    echo "❌ Server responsiveness: FAIL"
fi

if [ $error_result -eq 0 ]; then
    echo "✅ Error handling: PASS"
else
    echo "❌ Error handling: FAIL"
fi

if [ $shutdown_result -eq 0 ]; then
    echo "✅ Graceful shutdown: PASS"
else
    echo "❌ Graceful shutdown: FAIL"
fi

# Overall result
if [ $config_result -eq 0 ] && [ $response_result -eq 0 ] && [ $error_result -eq 0 ] && [ $shutdown_result -eq 0 ]; then
    echo ""
    echo "🎉 All timeout fix validation tests PASSED!"
    echo "The API hanging issues have been successfully resolved!"
    exit 0
else
    echo ""
    echo "💥 Some timeout fix validation tests FAILED!"
    exit 1
fi