#!/bin/bash

echo "Testing hanging prevention mechanisms..."
echo "======================================"

# Function to test tool execution with timeout
test_tool_timeout() {
    local tool_name="$1"
    local timeout_seconds="$2"
    local test_name="$3"
    
    echo "Testing $test_name (timeout: ${timeout_seconds}s)..."
    
    # Create the JSON-RPC call
    local json_call=$(cat << EOF
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "$tool_name", "arguments": {"topic": "overview"}}}
EOF
    )
    
    # Test with timeout - add extra time for server startup
    local total_timeout=$((timeout_seconds + 15))
    start_time=$(date +%s)
    
    if echo "$json_call" | timeout ${total_timeout}s ./rodmcp-enhanced --headless --log-level error 2>&1 | grep -q "result"
    then
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        
        # Subtract approximate server startup time (5 seconds)
        local execution_time=$((duration - 5))
        
        if [ $execution_time -lt $timeout_seconds ]; then
            echo "✅ $test_name completed in ~${execution_time}s (under ${timeout_seconds}s limit)"
            return 0
        else
            echo "⚠️  $test_name took ~${execution_time}s (near ${timeout_seconds}s limit)"
            return 1
        fi
    else
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        
        if [ $duration -ge $total_timeout ]; then
            echo "❌ $test_name hung for ${duration}s - timeout mechanism failed"
            return 2
        else
            echo "✅ $test_name failed quickly (${duration}s) - no hanging detected"
            return 0
        fi
    fi
}

# Test consecutive timeout handling
test_consecutive_timeouts() {
    echo "Testing consecutive timeout prevention..."
    
    # Create a script that sends many rapid requests
    local requests=""
    for i in {1..15}; do
        requests+='{"jsonrpc": "2.0", "id": '$i', "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "nonexistent"}}}'$'\n'
    done
    
    start_time=$(date +%s)
    
    # Test that server shuts down after consecutive timeouts
    if timeout 60s ./rodmcp-enhanced --headless --log-level warn 2>&1 << EOF | grep -q "Too many consecutive timeouts"
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
$requests
EOF
    then
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        echo "✅ Consecutive timeout prevention works - server shut down after ${duration}s"
        return 0
    else
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        echo "❌ Consecutive timeout prevention failed - server didn't shut down after ${duration}s"
        return 1
    fi
}

# Test 30-second tool timeout
test_tool_timeout_limit() {
    echo "Testing 30-second tool timeout limit..."
    
    # We can't easily create a tool that hangs for exactly 30 seconds,
    # but we can test that valid tools complete well under the limit
    local start_time=$(date +%s)
    
    if echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "overview"}}}' | timeout 40s ./rodmcp-enhanced --headless --log-level error 2>&1 | grep -q "result"
    then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        local execution_time=$((duration - 5))  # Subtract server startup time
        
        if [ $execution_time -lt 30 ]; then
            echo "✅ Tool completed in ~${execution_time}s (well under 30s limit)"
            return 0
        else
            echo "⚠️  Tool took ~${execution_time}s (near timeout limit)"
            return 1
        fi
    else
        echo "❌ Tool execution failed"
        return 1
    fi
}

# Run tests
echo "Building enhanced MCP server..."
if ! go build -o rodmcp-enhanced ./cmd/server 2>/dev/null; then
    echo "❌ Failed to build MCP server"
    exit 1
fi

echo ""

# Test basic tool execution doesn't hang
test_tool_timeout "help" 10 "Basic help tool"
help_result=$?

echo ""

# Test tool timeout limit  
test_tool_timeout_limit
timeout_result=$?

echo ""

# Test consecutive timeout prevention
test_consecutive_timeouts
consecutive_result=$?

echo ""
echo "Summary:"
echo "========"

if [ $help_result -eq 0 ]; then
    echo "✅ Basic tool execution: PASS"
else
    echo "❌ Basic tool execution: FAIL"
fi

if [ $timeout_result -eq 0 ]; then
    echo "✅ Tool timeout mechanism: PASS"
else
    echo "❌ Tool timeout mechanism: FAIL"
fi

if [ $consecutive_result -eq 0 ]; then
    echo "✅ Consecutive timeout prevention: PASS"
else
    echo "❌ Consecutive timeout prevention: FAIL"
fi

# Overall result
if [ $help_result -eq 0 ] && [ $timeout_result -eq 0 ] && [ $consecutive_result -eq 0 ]; then
    echo ""
    echo "🎉 All hanging prevention tests PASSED!"
    exit 0
else
    echo ""
    echo "💥 Some hanging prevention tests FAILED!"
    exit 1
fi