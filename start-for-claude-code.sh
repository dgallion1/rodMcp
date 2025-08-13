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

# Check and update RodMCP
check_and_update_rodmcp() {
    if [[ ! -f "$RODMCP_BIN" ]]; then
        print_error "RodMCP binary not found at $RODMCP_BIN"
        print_status "Building RodMCP..."
        cd "$SCRIPT_DIR"
        if make install-local; then
            print_success "RodMCP built and installed successfully"
        else
            print_error "Failed to build RodMCP"
            exit 1
        fi
        return
    fi
    
    # Check if we're in a git repository and can update
    if [[ -d "$SCRIPT_DIR/.git" ]]; then
        print_status "Checking for RodMCP updates..."
        cd "$SCRIPT_DIR"
        
        # Show current version (skip if server is running to avoid conflicts)
        if [[ -x "$RODMCP_BIN" ]] && ! pgrep -f "rodmcp.*http" >/dev/null 2>&1; then
            local current_version=$(timeout 3 "$RODMCP_BIN" version 2>/dev/null | head -1 | cut -d' ' -f2 2>/dev/null || echo "unknown")
            print_status "Current version: $current_version"
        elif pgrep -f "rodmcp.*http" >/dev/null 2>&1; then
            print_status "RodMCP server already running, will restart after update check"
        fi
        
        # Fetch latest changes
        if git fetch origin master --quiet 2>/dev/null; then
            local local_commit=$(git rev-parse HEAD)
            local remote_commit=$(git rev-parse origin/master 2>/dev/null || echo "$local_commit")
            
            if [[ "$local_commit" != "$remote_commit" ]]; then
                local commits_behind=$(git rev-list --count HEAD..origin/master 2>/dev/null || echo "unknown")
                print_status "Updates available! ($commits_behind commits behind)"
                print_status "Updating RodMCP..."
                
                # Stop any running rodmcp processes before update
                pkill -f "rodmcp.*http" 2>/dev/null || true
                sleep 1
                
                # Handle git state properly
                if git status --porcelain | grep -q .; then
                    print_status "Local changes detected, stashing before update..."
                    git stash push -m "Auto-stash before startup script update" --quiet
                    local stashed=true
                else
                    local stashed=false
                fi
                
                if git pull origin master --quiet && make install-local; then
                    # Skip version check after update to avoid conflicts with any running processes
                    print_success "RodMCP updated successfully"
                    
                    # Restore stash if we created one
                    if [[ "$stashed" == "true" ]]; then
                        print_status "Restoring local changes..."
                        git stash pop --quiet 2>/dev/null || print_warning "Could not restore some local changes"
                    fi
                else
                    print_error "Failed to update RodMCP, continuing with current version"
                    
                    # Restore stash even on failure
                    if [[ "$stashed" == "true" ]]; then
                        print_status "Restoring local changes..."
                        git stash pop --quiet 2>/dev/null || print_warning "Could not restore some local changes"
                    fi
                fi
            else
                print_status "RodMCP is up to date"
            fi
        else
            print_status "Cannot check for updates (offline or no remote)"
        fi
    else
        print_status "Not a git repository, skipping update check"
    fi
}

check_and_update_rodmcp

# Stop any existing RodMCP processes
stop_existing_rodmcp() {
    print_status "Stopping any existing RodMCP processes..."
    
    # Find RodMCP processes (exclude this script)
    local pids=$(pgrep -f "rodmcp.*http" 2>/dev/null | grep -v $$ || true)
    
    if [[ -n "$pids" ]]; then
        print_status "Found existing RodMCP processes: $pids"
        for pid in $pids; do
            if kill -0 $pid 2>/dev/null; then
                local cmdline=$(ps -p $pid -o args= 2>/dev/null || true)
                print_status "Stopping process $pid: $cmdline"
                kill -TERM $pid 2>/dev/null || true
                
                # Wait up to 5 seconds for graceful shutdown
                local count=0
                while kill -0 $pid 2>/dev/null && [ $count -lt 50 ]; do
                    sleep 0.1
                    count=$((count + 1))
                done
                
                # Force kill if still running
                if kill -0 $pid 2>/dev/null; then
                    print_status "Force killing process $pid..."
                    kill -KILL $pid 2>/dev/null || true
                fi
            fi
        done
        print_success "Existing RodMCP processes stopped"
    else
        print_status "No existing RodMCP processes found"
    fi
    
    # Clean up any stale lock files or temp files
    rm -f /tmp/rodmcp-*.* 2>/dev/null || true
}

# Check if port is already in use by non-RodMCP service
if command -v lsof >/dev/null 2>&1; then
    if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        # Check if it's a RodMCP process
        port_pid=$(lsof -ti :$PORT)
        if ps -p $port_pid -o args= 2>/dev/null | grep -q "rodmcp"; then
            print_status "RodMCP already using port $PORT, will restart it"
        else
            print_error "Another service is using port $PORT"
            print_status "Kill it or use: RODMCP_PORT=8091 $0"
            exit 1
        fi
    fi
fi

# Always stop existing processes before starting
stop_existing_rodmcp

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

# Update Claude Desktop configuration only (Claude Code uses claude mcp add)
update_claude_desktop_config() {
    # Update Claude Desktop (HTTP mode)  
    local desktop_config_dir="$HOME/.config/claude"
    local desktop_config_file="$desktop_config_dir/mcp_servers.json"
    
    if [[ ! -d "$desktop_config_dir" ]]; then
        mkdir -p "$desktop_config_dir"
        print_status "Created Claude Desktop config directory: $desktop_config_dir"
    fi
    
    local desktop_server_config="{
  \"mcpServers\": {
    \"rodmcp\": {
      \"url\": \"http://localhost:$PORT\"
    }
  }
}"
    
    if [[ -f "$desktop_config_file" ]]; then
        cp "$desktop_config_file" "$desktop_config_file.backup"
        print_status "Backed up Claude Desktop config to: $desktop_config_file.backup"
    fi
    
    echo "$desktop_server_config" > "$desktop_config_file"
    print_success "Updated Claude Desktop configuration (HTTP): $desktop_config_file"
}

# Check if server started successfully
if curl -s "http://localhost:$PORT/health" >/dev/null 2>&1; then
    print_success "RodMCP HTTP server started successfully!"
    print_success "PID: $SERVER_PID"
    print_success "Health check: http://localhost:$PORT/health"
    
    # Update Claude Desktop configuration (Claude Code uses claude mcp add)
    update_claude_desktop_config
    
    print_success "Ready for Claude Code integration!"
    echo
    # Configure Claude Code MCP project config
    print_status "Configuring Claude Code MCP project config..."
    project_mcp_config="$SCRIPT_DIR/.mcp.json"
    
    # Create project MCP config with proper arguments
    cat > "$project_mcp_config" << 'EOF'
{
  "mcpServers": {
    "rodmcp": {
      "type": "stdio",
      "command": "/home/darrell/.local/bin/rodmcp",
      "args": ["--headless", "--log-level=info"],
      "env": {}
    }
  }
}
EOF
    
    if [[ -f "$project_mcp_config" ]]; then
        print_success "Created Claude Code project MCP config: $project_mcp_config"
    else
        print_warning "Could not create Claude Code project MCP config"
    fi
    
    # Configure Claude Code MCP (user scope for global access)
    print_status "Configuring Claude Code MCP (global user scope)..."
    if command -v claude >/dev/null 2>&1; then
        if claude mcp add rodmcp "$SCRIPT_DIR/bin/rodmcp" -s user -- --headless --log-level=info 2>/dev/null; then
            print_success "Added rodmcp to Claude Code user configuration"
        elif claude mcp add rodmcp /home/darrell/.local/bin/rodmcp -s user -- --headless --log-level=info 2>/dev/null; then
            print_success "Added rodmcp to Claude Code user configuration (local bin)"
        else
            print_warning "Could not add rodmcp to Claude Code - may already exist or claude command not available"
        fi
    else
        print_warning "Claude Code CLI not available - configure manually with: claude mcp add rodmcp /home/darrell/.local/bin/rodmcp -s user -- --headless --log-level=info"
    fi
    
    print_status "Next steps:"
    echo "1. For Claude Desktop: Close and reopen Claude Desktop app"
    echo "2. For Claude Code: Should now be configured globally"
    echo "3. Ask Claude: 'What tools do you have available?'"
    echo "4. Start automating: 'Create a simple HTML page and take a screenshot'"
    echo
    print_status "To stop server: kill $SERVER_PID"
    print_status "Server running in background - script will now exit"
    
    # Detach from terminal and exit script
    disown $SERVER_PID
    echo "✅ RodMCP server running in background (PID: $SERVER_PID)"
else
    print_error "Failed to start RodMCP server"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi