#!/bin/bash

# Setup script to configure Claude for headless browser mode (local user installation)

echo "🚫 Configuring Claude for headless browser (Local User)"
echo "======================================================="

# Get the local installation path
LOCAL_BIN="$HOME/.local/bin"
RODMCP_PATH="$LOCAL_BIN/rodmcp"

if [ ! -f "$RODMCP_PATH" ]; then
    echo "❌ Error: RodMCP not found at $RODMCP_PATH"
    echo "   Please run ./install-local.sh first"
    exit 1
fi

# Create headless browser configurations
echo "Which Claude client are you using?"
echo "1) Claude Desktop App"
echo "2) Claude Code (CLI)"
echo "3) Both"
read -p "Choose (1/2/3): " choice

case $choice in
    1)
        echo "📱 Configuring Claude Desktop for headless browser..."
        DESKTOP_CONFIG="$HOME/.config/claude-desktop/config.json"
        mkdir -p "$(dirname "$DESKTOP_CONFIG")"
        
        cat << EOF > "$DESKTOP_CONFIG"
{
  "mcpServers": {
    "rodmcp": {
      "command": "$RODMCP_PATH",
      "args": [
        "--headless=true",
        "--log-level=info",
        "--window-width=1920",
        "--window-height=1080"
      ]
    }
  }
}
EOF
        echo "✅ Claude Desktop configured for headless browser"
        echo "⚠️  Please restart Claude Desktop for changes to take effect"
        ;;
        
    2) 
        echo "💻 Configuring Claude Code for headless browser..."
        CODE_CONFIG="$HOME/.config/claude/mcp_servers.json"
        mkdir -p "$(dirname "$CODE_CONFIG")"
        
        cat << EOF > "$CODE_CONFIG"
{
  "rodmcp": {
    "command": "$RODMCP_PATH",
    "args": [
      "--headless=true",
      "--log-level=info",
      "--window-width=1920",
      "--window-height=1080"
    ]
  }
}
EOF
        echo "✅ Claude Code configured for headless browser"
        echo "⚠️  Please restart Claude Code for changes to take effect"
        ;;
        
    3)
        echo "📱💻 Configuring both Claude Desktop and Claude Code..."
        
        DESKTOP_CONFIG="$HOME/.config/claude-desktop/config.json"
        mkdir -p "$(dirname "$DESKTOP_CONFIG")"
        cat << EOF > "$DESKTOP_CONFIG"
{
  "mcpServers": {
    "rodmcp": {
      "command": "$RODMCP_PATH",
      "args": [
        "--headless=true",
        "--log-level=info",
        "--window-width=1920",
        "--window-height=1080"
      ]
    }
  }
}
EOF
        echo "✅ Claude Desktop configured"
        
        CODE_CONFIG="$HOME/.config/claude/mcp_servers.json"
        mkdir -p "$(dirname "$CODE_CONFIG")"
        cat << EOF > "$CODE_CONFIG"
{
  "rodmcp": {
    "command": "$RODMCP_PATH",
    "args": [
      "--headless=true",
      "--log-level=info",
      "--window-width=1920",
      "--window-height=1080"
    ]
  }
}
EOF
        echo "✅ Claude Code configured"
        echo "⚠️  Please restart both Claude Desktop and Claude Code"
        ;;
        
    *)
        echo "❌ Invalid choice. Please run the script again."
        exit 1
        ;;
esac

echo ""
echo "🚫 Headless Browser Configuration:"
echo "   • Browser window: HIDDEN"
echo "   • Performance: FASTER"
echo "   • Resource usage: LOWER"
echo "   • Screenshots: STILL AVAILABLE"
echo "   • Window size: 1920x1080"
echo "   • RodMCP path: $RODMCP_PATH"
echo ""

echo "🧪 Test the Configuration:"
echo "   After restarting Claude, ask:"
echo "   'Create an HTML page and take a screenshot of it'"
echo ""

echo "🔧 To Switch Back to Visible Mode:"
echo "   Run: ./configs/setup-visible-browser-local.sh"
echo ""

echo "🎉 Configuration complete!"