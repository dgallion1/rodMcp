#!/bin/bash

# Test shutdown behavior to ensure no panic occurs
echo "Testing RodMCP Shutdown Behavior (Panic Fix)"
echo "============================================"

# Kill any existing processes
pkill -f rodmcp-panic-fixed || true
sleep 2

echo "Starting RodMCP server and testing graceful shutdown..."

# Start server in background and capture output
timeout 10s ./rodmcp-panic-fixed --headless --log-level info 2>&1 | tee shutdown_test_output.log &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Server started with PID: $SERVER_PID"

# Send SIGTERM for graceful shutdown
echo "Sending SIGTERM for graceful shutdown..."
kill -TERM $SERVER_PID

# Wait for shutdown
wait $SERVER_PID 2>/dev/null
EXIT_CODE=$?

echo ""
echo "Server shutdown complete with exit code: $EXIT_CODE"
echo ""

# Check for panic in logs
if grep -i "panic\|FATAL" shutdown_test_output.log; then
    echo "❌ PANIC DETECTED - Fix not working"
    echo ""
    echo "Panic details:"
    grep -A 10 -i "panic\|FATAL" shutdown_test_output.log
else
    echo "✅ NO PANIC DETECTED - Fix working correctly"
fi

echo ""
echo "Full shutdown log:"
tail -20 shutdown_test_output.log | grep -E "(stopping|stopped|shutdown|browser|panic|error|FATAL)"

# Cleanup
pkill -f rodmcp-panic-fixed || true