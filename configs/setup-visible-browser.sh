#!/bin/bash

# Setup script to configure Claude to show the browser window

echo "🖥️  Configuring Claude to show browser windows"
echo "=============================================="

# Detect which Claude client is being configured
echo "Which Claude client are you using?"
echo "1) Claude Desktop App"
echo "2) Claude Code (CLI)"
echo "3) Both"
read -p "Choose (1/2/3): " choice

case $choice in
    1)
        echo "📱 Configuring Claude Desktop..."
        DESKTOP_CONFIG="$HOME/.config/claude-desktop/config.json"
        mkdir -p "$(dirname "$DESKTOP_CONFIG")"
        cp configs/claude-desktop-visible.json "$DESKTOP_CONFIG"
        echo "✅ Claude Desktop configured at $DESKTOP_CONFIG"
        echo "⚠️  Please restart Claude Desktop for changes to take effect"
        ;;
    2) 
        echo "💻 Configuring Claude Code..."
        CODE_CONFIG="$HOME/.config/claude/mcp_servers.json"
        mkdir -p "$(dirname "$CODE_CONFIG")"
        cp configs/claude-code-visible.json "$CODE_CONFIG"
        echo "✅ Claude Code configured at $CODE_CONFIG"
        echo "⚠️  Please restart Claude Code for changes to take effect"
        ;;
    3)
        echo "📱💻 Configuring both Claude Desktop and Claude Code..."
        
        DESKTOP_CONFIG="$HOME/.config/claude-desktop/config.json"
        mkdir -p "$(dirname "$DESKTOP_CONFIG")"
        cp configs/claude-desktop-visible.json "$DESKTOP_CONFIG"
        echo "✅ Claude Desktop configured"
        
        CODE_CONFIG="$HOME/.config/claude/mcp_servers.json"
        mkdir -p "$(dirname "$CODE_CONFIG")"
        cp configs/claude-code-visible.json "$CODE_CONFIG"
        echo "✅ Claude Code configured"
        
        echo "⚠️  Please restart both Claude Desktop and Claude Code"
        ;;
    *)
        echo "❌ Invalid choice. Please run the script again."
        exit 1
        ;;
esac

echo ""
echo "🎬 Browser Visibility Configuration:"
echo "   • Headless mode: DISABLED"
echo "   • Browser window: VISIBLE"
echo "   • Window size: 1200x800"
echo "   • Slow motion: 500ms (for visibility)"
echo "   • Debug mode: OFF"
echo ""

# Check for display environment (Linux)
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    if [ -z "$DISPLAY" ]; then
        echo "⚠️  Warning: DISPLAY environment variable not set"
        echo "   For GUI display on Linux, you may need:"
        echo "   export DISPLAY=:0"
        echo "   Or install a virtual display (xvfb)"
    else
        echo "✅ Display environment: $DISPLAY"
    fi
fi

echo ""
echo "🧪 Test the Configuration:"
echo "   After restarting Claude, ask:"
echo "   'Create a simple HTML page and show me the browser while you work'"
echo ""

echo "🔧 To Switch Back to Headless Mode:"
echo "   Run: ./configs/setup-headless-browser.sh"
echo ""

echo "🎉 Configuration complete!"