#!/bin/bash

echo "Testing MCP protocol compliance..."
echo "=================================="

# Test basic JSON-RPC initialization sequence
test_mcp_initialization() {
    echo "Testing MCP initialization sequence..."
    
    local output=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}' | timeout 15s ./rodmcp-enhanced --headless --log-level error 2>&1)
    
    if echo "$output" | grep -q '"jsonrpc":"2.0"' && echo "$output" | grep -q '"result"' && echo "$output" | grep -q '"protocolVersion"'; then
        echo "✅ MCP initialization successful"
        return 0
    else
        echo "❌ MCP initialization failed"
        echo "Output: $output"
        return 1
    fi
}

# Test tools list functionality
test_tools_list() {
    echo "Testing tools/list functionality..."
    
    local output=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}' | timeout 15s ./rodmcp-enhanced --headless --log-level error 2>&1)
    
    if echo "$output" | grep -q '"tools"' && echo "$output" | grep -q '"name"' && echo "$output" | grep -q '"description"'; then
        local tool_count=$(echo "$output" | grep -o '"name"' | wc -l)
        echo "✅ Tools list successful - found $tool_count tools"
        return 0
    else
        echo "❌ Tools list failed"
        return 1
    fi
}

# Test basic tool execution
test_tool_execution() {
    echo "Testing basic tool execution..."
    
    local output=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "overview"}}}' | timeout 15s ./rodmcp-enhanced --headless --log-level error 2>&1)
    
    if echo "$output" | grep -q '"result"' && echo "$output" | grep -q '"content"'; then
        echo "✅ Tool execution successful"
        return 0
    else
        echo "❌ Tool execution failed"
        echo "Output sample: $(echo "$output" | head -c 200)..."
        return 1
    fi
}

# Test invalid tool execution
test_invalid_tool() {
    echo "Testing invalid tool handling..."
    
    local output=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "nonexistent_tool", "arguments": {}}}' | timeout 15s ./rodmcp-enhanced --headless --log-level error 2>&1)
    
    if echo "$output" | grep -q '"error"' && echo "$output" | grep -q '"code":-32601'; then
        echo "✅ Invalid tool properly rejected with correct error code"
        return 0
    else
        echo "❌ Invalid tool not properly handled"
        return 1
    fi
}

# Build server
echo "Building MCP server..."
if ! go build -o rodmcp-enhanced ./cmd/server 2>/dev/null; then
    echo "❌ Failed to build MCP server"
    exit 1
fi

echo ""

# Run tests
test_mcp_initialization
init_result=$?

echo ""

test_tools_list
list_result=$?

echo ""

test_tool_execution  
exec_result=$?

echo ""

test_invalid_tool
invalid_result=$?

echo ""
echo "Summary:"
echo "========"

if [ $init_result -eq 0 ]; then
    echo "✅ MCP initialization: PASS"
else
    echo "❌ MCP initialization: FAIL"
fi

if [ $list_result -eq 0 ]; then
    echo "✅ Tools list: PASS"
else
    echo "❌ Tools list: FAIL"
fi

if [ $exec_result -eq 0 ]; then
    echo "✅ Tool execution: PASS"
else
    echo "❌ Tool execution: FAIL"
fi

if [ $invalid_result -eq 0 ]; then
    echo "✅ Error handling: PASS"
else
    echo "❌ Error handling: FAIL"
fi

# Overall result
if [ $init_result -eq 0 ] && [ $list_result -eq 0 ] && [ $exec_result -eq 0 ] && [ $invalid_result -eq 0 ]; then
    echo ""
    echo "🎉 All MCP protocol tests PASSED!"
    exit 0
else
    echo ""
    echo "💥 Some MCP protocol tests FAILED!"
    exit 1
fi