#!/bin/bash

echo "Testing realistic API hanging fixes..."
echo "====================================="

# Test 1: Verify server doesn't hang on normal MCP workflow
test_normal_workflow() {
    echo "Test 1: Normal MCP workflow (initialize -> list tools -> execute tool)"
    
    # This is the exact sequence Claude Code would use
    local mcp_sequence='{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "overview"}}}'
    
    echo "Executing normal MCP sequence..."
    local result=$(echo "$mcp_sequence" | ./rodmcp-enhanced --headless --log-level warn 2>&1)
    
    # Check for successful responses
    local init_success=$(echo "$result" | grep -c '"id":1.*"result"')
    local tools_success=$(echo "$result" | grep -c '"id":2.*"result".*"tools"')
    local exec_success=$(echo "$result" | grep -c '"id":3.*"result".*"content"')
    
    if [ "$init_success" -eq 1 ] && [ "$tools_success" -eq 1 ] && [ "$exec_success" -eq 1 ]; then
        echo "✅ Normal workflow completed successfully"
        return 0
    else
        echo "❌ Normal workflow failed"
        echo "Init responses: $init_success, Tools responses: $tools_success, Exec responses: $exec_success"
        return 1
    fi
}

# Test 2: Verify server handles multiple rapid requests without hanging
test_rapid_requests() {
    echo "Test 2: Multiple rapid tool requests"
    
    # Send multiple help requests rapidly
    local rapid_sequence='{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "overview"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "workflows"}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "examples"}}}'
    
    echo "Sending multiple rapid requests..."
    local result=$(echo "$rapid_sequence" | ./rodmcp-enhanced --headless --log-level warn 2>&1)
    
    # Count successful responses
    local response_count=$(echo "$result" | grep -c '"result"')
    
    if [ "$response_count" -ge 4 ]; then
        echo "✅ Multiple rapid requests handled successfully ($response_count responses)"
        return 0
    else
        echo "❌ Rapid requests failed (only $response_count responses)"
        return 1
    fi
}

# Test 3: Verify server properly rejects invalid tools without hanging
test_invalid_tool_handling() {
    echo "Test 3: Invalid tool handling"
    
    local invalid_sequence='{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "nonexistent_tool", "arguments": {}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "another_fake_tool", "arguments": {}}}'
    
    echo "Testing invalid tool requests..."
    local result=$(echo "$invalid_sequence" | ./rodmcp-enhanced --headless --log-level warn 2>&1)
    
    # Check for proper error responses
    local error_count=$(echo "$result" | grep -c '"error"')
    local tool_not_found=$(echo "$result" | grep -c '"code":-32601')
    
    if [ "$error_count" -ge 2 ] && [ "$tool_not_found" -ge 2 ]; then
        echo "✅ Invalid tools properly rejected with correct errors"
        return 0
    else
        echo "❌ Invalid tool handling failed (errors: $error_count, not found: $tool_not_found)"
        return 1
    fi
}

# Test 4: Test that server handles file operations without hanging
test_file_operations() {
    echo "Test 4: File operations"
    
    # Create a test file
    echo "test content" > /tmp/test_file.txt
    
    local file_sequence='{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "read_file", "arguments": {"path": "/tmp/test_file.txt"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "list_directory", "arguments": {"path": "/tmp"}}}'
    
    echo "Testing file operations..."
    local result=$(echo "$file_sequence" | ./rodmcp-enhanced --headless --log-level warn 2>&1)
    
    # Check for successful file operations
    local read_success=$(echo "$result" | grep -c '"id":2.*"result"')
    local list_success=$(echo "$result" | grep -c '"id":3.*"result"')
    
    # Cleanup
    rm -f /tmp/test_file.txt
    
    if [ "$read_success" -eq 1 ] && [ "$list_success" -eq 1 ]; then
        echo "✅ File operations completed successfully"
        return 0
    else
        echo "❌ File operations failed"
        return 1
    fi
}

# Test 5: Verify server handles malformed JSON without hanging
test_malformed_json() {
    echo "Test 5: Malformed JSON handling"
    
    local malformed_sequence='{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}}}
{"jsonrpc": "2.0", "method": "notifications/initialized"}
{invalid json}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "help", "arguments": {"topic": "overview"}}}'
    
    echo "Testing malformed JSON handling..."
    local result=$(echo "$malformed_sequence" | ./rodmcp-enhanced --headless --log-level warn 2>&1)
    
    # Should have init response, parse error, and help response
    local init_success=$(echo "$result" | grep -c '"id":1.*"result"')
    local parse_error=$(echo "$result" | grep -c '"error".*-32700')
    local help_success=$(echo "$result" | grep -c '"id":2.*"result"')
    
    if [ "$init_success" -eq 1 ] && [ "$parse_error" -ge 1 ] && [ "$help_success" -eq 1 ]; then
        echo "✅ Malformed JSON handled correctly (continues processing after error)"
        return 0
    else
        echo "❌ Malformed JSON handling failed"
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

# Run realistic tests
test_normal_workflow
test1_result=$?

echo ""

test_rapid_requests  
test2_result=$?

echo ""

test_invalid_tool_handling
test3_result=$?

echo ""

test_file_operations
test4_result=$?

echo ""

test_malformed_json
test5_result=$?

echo ""
echo "========================================="
echo "REALISTIC API HANGING FIXES TEST RESULTS"
echo "========================================="

if [ $test1_result -eq 0 ]; then
    echo "✅ Normal MCP workflow: PASS"
else
    echo "❌ Normal MCP workflow: FAIL"
fi

if [ $test2_result -eq 0 ]; then
    echo "✅ Rapid requests: PASS"
else
    echo "❌ Rapid requests: FAIL"
fi

if [ $test3_result -eq 0 ]; then
    echo "✅ Invalid tool handling: PASS"
else
    echo "❌ Invalid tool handling: FAIL"
fi

if [ $test4_result -eq 0 ]; then
    echo "✅ File operations: PASS"
else
    echo "❌ File operations: FAIL"
fi

if [ $test5_result -eq 0 ]; then
    echo "✅ Malformed JSON handling: PASS"
else
    echo "❌ Malformed JSON handling: FAIL"
fi

echo ""

# Overall result
if [ $test1_result -eq 0 ] && [ $test2_result -eq 0 ] && [ $test3_result -eq 0 ] && [ $test4_result -eq 0 ] && [ $test5_result -eq 0 ]; then
    echo "🎉 ALL REALISTIC TESTS PASSED!"
    echo "✅ API hanging issues have been successfully resolved"
    echo "✅ Server handles normal workflows without hanging"
    echo "✅ Server handles error conditions gracefully"
    echo "✅ Server processes multiple requests reliably"
    exit 0
else
    echo "💥 SOME REALISTIC TESTS FAILED!"
    echo "API hanging fixes may need additional work"
    exit 1
fi