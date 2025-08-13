#!/bin/bash

# RodMCP Local User Installation Script
# Installs RodMCP for the current user without requiring sudo

set -e

echo "🏠 RodMCP Local User Installation"
echo "================================="

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

# Determine installation directory
LOCAL_BIN="$HOME/.local/bin"
if [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
    echo "📁 Creating local bin directory: $LOCAL_BIN"
    mkdir -p "$LOCAL_BIN"
    
    echo "⚠️  WARNING: $LOCAL_BIN is not in your PATH"
    echo ""
    echo "Add this line to your shell configuration file:"
    echo "  For bash (~/.bashrc):"
    echo "    export PATH=\"\$PATH:$LOCAL_BIN\""
    echo "  For zsh (~/.zshrc):"
    echo "    export PATH=\"\$PATH:$LOCAL_BIN\""
    echo ""
    read -p "Press Enter to continue installation..."
fi

# Build RodMCP
echo "🔨 Building RodMCP..."
if ! go build -o bin/rodmcp cmd/server/main.go; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Install to local user directory
INSTALL_PATH="$LOCAL_BIN/rodmcp"

# Stop any running rodmcp processes before installation
echo "🛑 Checking for running RodMCP processes..."
if pgrep -f "rodmcp" > /dev/null; then
    echo "⚠️  Found running RodMCP processes. Stopping them..."
    pkill -f "rodmcp" || true
    sleep 2
    
    # Check if any processes are still running and force kill if necessary
    if pgrep -f "rodmcp" > /dev/null; then
        echo "⚠️  Force stopping remaining processes..."
        pkill -9 -f "rodmcp" || true
        sleep 1
    fi
    echo "✅ RodMCP processes stopped"
else
    echo "✅ No running RodMCP processes found"
fi

echo "📦 Installing RodMCP to $INSTALL_PATH..."
cp bin/rodmcp "$INSTALL_PATH"
chmod +x "$INSTALL_PATH"

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
    echo "   Creating backup: ${CLAUDE_DESKTOP_CONFIG}.backup"
    cp "$CLAUDE_DESKTOP_CONFIG" "${CLAUDE_DESKTOP_CONFIG}.backup"
fi

cat << EOF > "$CLAUDE_DESKTOP_CONFIG"
{
  "mcpServers": {
    "rodmcp": {
      "command": "$INSTALL_PATH",
      "args": [
        "--headless=true",
        "--log-level=info"
      ]
    }
  }
}
EOF
echo "✅ Claude Desktop config created at $CLAUDE_DESKTOP_CONFIG"

# Generate Claude Code configuration
CLAUDE_CODE_CONFIG="$HOME/.config/claude/mcp_servers.json"
echo "⚙️  Generating Claude Code configuration..."

if [ -f "$CLAUDE_CODE_CONFIG" ]; then
    echo "⚠️  Existing Claude Code config found at $CLAUDE_CODE_CONFIG"
    echo "   Creating backup: ${CLAUDE_CODE_CONFIG}.backup"
    cp "$CLAUDE_CODE_CONFIG" "${CLAUDE_CODE_CONFIG}.backup"
fi

cat << EOF > "$CLAUDE_CODE_CONFIG"
{
  "rodmcp": {
    "command": "$INSTALL_PATH",
    "args": [
      "--headless=true",
      "--log-level=info"
    ]
  }
}
EOF
echo "✅ Claude Code config created at $CLAUDE_CODE_CONFIG"

# Create shell alias (optional)
echo ""
echo "📝 Optional: Add this alias to your shell configuration for easy access:"
echo "   alias rodmcp='$INSTALL_PATH'"

# Create uninstall script
UNINSTALL_SCRIPT="$HOME/.local/bin/uninstall-rodmcp"
cat << 'EOF' > "$UNINSTALL_SCRIPT"
#!/bin/bash
echo "🗑️  Uninstalling RodMCP..."

# Remove binary
rm -f "$HOME/.local/bin/rodmcp"

# Backup and remove configs
if [ -f "$HOME/.config/claude-desktop/config.json" ]; then
    mv "$HOME/.config/claude-desktop/config.json" "$HOME/.config/claude-desktop/config.json.uninstalled"
    echo "✅ Claude Desktop config backed up"
fi

if [ -f "$HOME/.config/claude/mcp_servers.json" ]; then
    mv "$HOME/.config/claude/mcp_servers.json" "$HOME/.config/claude/mcp_servers.json.uninstalled"
    echo "✅ Claude Code config backed up"
fi

# Remove self
rm -f "$HOME/.local/bin/uninstall-rodmcp"

echo "✅ RodMCP uninstalled"
echo "   Backup configs saved with .uninstalled extension"
EOF

chmod +x "$UNINSTALL_SCRIPT"
echo "✅ Uninstall script created at $UNINSTALL_SCRIPT"

echo ""
echo "🎉 Local User Installation Complete!"
echo "===================================="
echo ""
echo "📋 Installation Summary:"
echo "   • RodMCP binary: $INSTALL_PATH"
echo "   • Claude Desktop config: $CLAUDE_DESKTOP_CONFIG"
echo "   • Claude Code config: $CLAUDE_CODE_CONFIG"
echo "   • Uninstall script: $UNINSTALL_SCRIPT"
echo ""

# Check if PATH needs to be updated
if [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
    echo "⚠️  IMPORTANT: Add $LOCAL_BIN to your PATH:"
    echo ""
    echo "   Run this command (for current session):"
    echo "     export PATH=\"\$PATH:$LOCAL_BIN\""
    echo ""
    echo "   Then add it permanently to your shell config:"
    echo "     echo 'export PATH=\"\$PATH:$LOCAL_BIN\"' >> ~/.bashrc"
    echo "     source ~/.bashrc"
    echo ""
fi

echo "🔧 Available Tools (ask Claude to use these):"
echo "   • create_page - Generate HTML pages"
echo "   • navigate_page - Open URLs in browser"
echo "   • take_screenshot - Capture page screenshots"
echo "   • execute_script - Run JavaScript code"
echo "   • set_browser_visibility - Control browser visibility dynamically"
echo "   • live_preview - Start development server"
echo ""
echo "✅ Next Steps:"
echo "   1. Restart Claude Desktop/Code (if running)"
echo "   2. Ask Claude: 'What web development tools do you have?'"
echo "   3. Test with: 'Create a simple HTML page and screenshot it'"
echo ""
echo "👁️  Browser Visibility:"
echo "   • Claude can dynamically control visibility using set_browser_visibility tool"
echo "   • Ask: 'Show me the browser while you work' or 'Switch to headless mode'"
echo "   • Optional manual config: ./configs/setup-visible-browser-local.sh"
echo ""
echo "🗑️  To uninstall later:"
echo "   Run: uninstall-rodmcp"
echo ""

# Check for Chrome/Chromium
if ! command -v google-chrome &> /dev/null && ! command -v chromium &> /dev/null && ! command -v chromium-browser &> /dev/null; then
    echo "ℹ️  Note: Chrome/Chromium not found - Rod will download it automatically on first use"
fi

echo "🚀 RodMCP is ready to use!"