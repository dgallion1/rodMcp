#!/bin/bash

# RodMCP Installation Script
# This script installs RodMCP as an MCP server for use with Claude

set -e

echo "🚀 RodMCP Installation Script"
echo "=============================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go 1.24+ is required but not installed"
    echo "   Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
MIN_VERSION="1.24"

if ! printf '%s\n%s\n' "$MIN_VERSION" "$GO_VERSION" | sort -V -C; then
    echo "❌ Error: Go $MIN_VERSION+ required, found $GO_VERSION"
    exit 1
fi

echo "✅ Go $GO_VERSION detected"

# Build RodMCP
echo "🔨 Building RodMCP..."
if ! go build -o bin/rodmcp cmd/server/main.go; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Install to system PATH
INSTALL_PATH="/usr/local/bin/rodmcp"

echo "📦 Installing RodMCP to $INSTALL_PATH..."
if command -v sudo &> /dev/null; then
    sudo cp bin/rodmcp "$INSTALL_PATH"
    sudo chmod +x "$INSTALL_PATH"
else
    echo "⚠️  sudo not available - manual installation required"
    echo "   Please copy bin/rodmcp to a directory in your PATH"
    echo "   Example: cp bin/rodmcp ~/.local/bin/rodmcp"
    exit 1
fi

echo "✅ Installation successful"

# Test installation
echo "🧪 Testing installation..."
if "$INSTALL_PATH" --help > /dev/null 2>&1; then
    echo "✅ RodMCP is working correctly"
else
    echo "❌ Installation test failed"
    exit 1
fi

# Create configuration directories
echo "📁 Creating configuration directories..."
mkdir -p ~/.config/claude-desktop 2>/dev/null || true
mkdir -p ~/.config/claude 2>/dev/null || true

# Generate Claude Desktop configuration
CLAUDE_DESKTOP_CONFIG="$HOME/.config/claude-desktop/config.json"
echo "⚙️  Generating Claude Desktop configuration..."

if [ -f "$CLAUDE_DESKTOP_CONFIG" ]; then
    echo "⚠️  Existing Claude Desktop config found at $CLAUDE_DESKTOP_CONFIG"
    echo "   Please manually add the RodMCP server configuration:"
    echo ""
    cat << 'EOF'
{
  "mcpServers": {
    "rodmcp": {
      "command": "/usr/local/bin/rodmcp",
      "args": [
        "--headless=true",
        "--log-level=info"
      ]
    }
  }
}
EOF
else
    cat << 'EOF' > "$CLAUDE_DESKTOP_CONFIG"
{
  "mcpServers": {
    "rodmcp": {
      "command": "/usr/local/bin/rodmcp",
      "args": [
        "--headless=true",
        "--log-level=info"
      ]
    }
  }
}
EOF
    echo "✅ Claude Desktop config created at $CLAUDE_DESKTOP_CONFIG"
fi

# Generate Claude Code configuration
CLAUDE_CODE_CONFIG="$HOME/.config/claude/mcp_servers.json"
echo "⚙️  Generating Claude Code configuration..."

if [ -f "$CLAUDE_CODE_CONFIG" ]; then
    echo "⚠️  Existing Claude Code config found at $CLAUDE_CODE_CONFIG"
    echo "   Please manually add the RodMCP server configuration"
else
    cat << 'EOF' > "$CLAUDE_CODE_CONFIG"
{
  "rodmcp": {
    "command": "/usr/local/bin/rodmcp",
    "args": [
      "--headless=true",
      "--log-level=info"
    ]
  }
}
EOF
    echo "✅ Claude Code config created at $CLAUDE_CODE_CONFIG"
fi

echo ""
echo "🎉 RodMCP Installation Complete!"
echo "================================="
echo ""
echo "📋 What was installed:"
echo "   • RodMCP server: $INSTALL_PATH"
echo "   • Claude Desktop config: $CLAUDE_DESKTOP_CONFIG"
echo "   • Claude Code config: $CLAUDE_CODE_CONFIG"
echo ""
echo "🔧 Available Tools (ask Claude to use these):"
echo "   • create_page - Generate HTML pages"
echo "   • navigate_page - Open URLs in browser"
echo "   • take_screenshot - Capture page screenshots"
echo "   • execute_script - Run JavaScript code"
echo "   • set_browser_visibility - Control browser visibility dynamically"
echo "   • live_preview - Start development server"
echo ""
echo "✅ Next Steps:"
echo "   1. Restart Claude Desktop (if using)"
echo "   2. Ask Claude: 'What web development tools do you have?'"
echo "   3. Test with: 'Create a simple HTML page and screenshot it'"
echo ""
echo "👁️  Browser Visibility:"
echo "   • Claude can dynamically control visibility using set_browser_visibility tool"
echo "   • Ask: 'Show me the browser while you work' or 'Switch to headless mode'"
echo "   • Optional manual config available in ./configs/ directory"
echo ""
echo "📚 For more info, see: INSTALLATION.md and README.md"
echo ""

# Check for Chrome/Chromium
if ! command -v google-chrome &> /dev/null && ! command -v chromium &> /dev/null && ! command -v chromium-browser &> /dev/null; then
    echo "ℹ️  Note: Chrome/Chromium not found - Rod will download it automatically on first use"
fi

echo "🚀 RodMCP is ready to use!"