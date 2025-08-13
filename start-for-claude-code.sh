#!/bin/bash

# RodMCP Startup Script for Claude Code Integration
# Starts RodMCP HTTP server optimized for Claude Code usage

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RODMCP_BIN="$SCRIPT_DIR/rodmcp"
PORT="${RODMCP_PORT:-8090}"
LOG_LEVEL="${LOG_LEVEL:-info}"
HEADLESS="${HEADLESS:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[RodMCP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[RodMCP]${NC} ✅ $1"
}

print_warning() {
    echo -e "${YELLOW}[RodMCP]${NC} ⚠️  $1"
}

print_error() {
    echo -e "${RED}[RodMCP]${NC} ❌ $1"
}

# Check if binary exists
if [[ ! -f "$RODMCP_BIN" ]]; then
    print_error "RodMCP binary not found at $RODMCP_BIN"
    print_status "Run 'make build' or 'make install-local' first"
    exit 1
fi

# Check if port is already in use
if command -v lsof >/dev/null 2>&1; then
    if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        print_warning "Port $PORT is already in use"
        print_status "Checking if it's RodMCP..."
        
        if curl -s "http://localhost:$PORT/health" | grep -q "rodmcp\|tools"; then
            print_success "RodMCP is already running on port $PORT"
            print_status "Server ready for Claude Code at: http://localhost:$PORT"
            exit 0
        else
            print_error "Another service is using port $PORT"
            print_status "Kill it or use: RODMCP_PORT=8091 $0"
            exit 1
        fi
    fi
fi

print_status "Starting RodMCP HTTP server for Claude Code integration..."
print_status "Port: $PORT"
print_status "Log Level: $LOG_LEVEL"
print_status "Headless: $HEADLESS"

# Build command arguments for HTTP mode
ARGS="http --port=$PORT --log-level=$LOG_LEVEL"
if [[ "$HEADLESS" == "true" ]]; then
    ARGS="$ARGS --headless"
fi

# Detect Chrome/Chromium path
if [[ -z "$CHROME_PATH" ]]; then
    for chrome_candidate in \
        "/usr/bin/google-chrome" \
        "/usr/bin/chromium-browser" \
        "/usr/bin/chromium" \
        "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
        "/Applications/Chromium.app/Contents/MacOS/Chromium" \
        "$HOME/.nix-profile/bin/chromium-browser"; do
        
        if [[ -x "$chrome_candidate" ]]; then
            export CHROME_PATH="$chrome_candidate"
            break
        fi
    done
    
    if [[ -n "$CHROME_PATH" ]]; then
        print_status "Using Chrome: $CHROME_PATH"
    else
        print_warning "Chrome/Chromium not found - browser automation may fail"
    fi
fi

print_status "Command: $RODMCP_BIN $ARGS"

# Start the server in background
"$RODMCP_BIN" $ARGS &
SERVER_PID=$!

# Wait a moment for server to start
sleep 2

# Check if server started successfully
if curl -s "http://localhost:$PORT/health" >/dev/null 2>&1; then
    print_success "RodMCP HTTP server started successfully!"
    print_success "PID: $SERVER_PID"
    print_success "Health check: http://localhost:$PORT/health"
    print_success "Ready for Claude Code integration"
    
    echo
    print_status "Next steps:"
    echo "1. Configure Claude Code with:"
    echo "   {\"mcpServers\": {\"rodmcp\": {\"url\": \"http://localhost:$PORT\"}}}"
    echo "2. Restart Claude Code"
    echo "3. Ask Claude: 'What tools do you have available?'"
    echo
    print_status "To stop: kill $SERVER_PID"
    
    # Keep script running to show PID
    echo "Press Ctrl+C to stop the server..."
    wait $SERVER_PID
else
    print_error "Failed to start RodMCP server"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi