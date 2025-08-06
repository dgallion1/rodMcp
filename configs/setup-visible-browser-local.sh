#!/bin/bash

# Setup script to configure Claude for visible browser mode (local user installation)

echo "🖥️  Configuring Claude for visible browser (Local User)"
echo "======================================================"

# Get the local installation path
LOCAL_BIN="$HOME/.local/bin"
RODMCP_PATH="$LOCAL_BIN/rodmcp"

if [ ! -f "$RODMCP_PATH" ]; then
    echo "❌ Error: RodMCP not found at $RODMCP_PATH"
    echo "   Please run ./install-local.sh first"
    exit 1
fi

# Create visible browser configurations
echo "Which Claude client are you using?"
echo "1) Claude Desktop App"
echo "2) Claude Code (CLI)"
echo "3) Both"
read -p "Choose (1/2/3): " choice

case $choice in
    1)
        echo "📱 Configuring Claude Desktop for visible browser..."
        DESKTOP_CONFIG="$HOME/.config/claude-desktop/config.json"
        mkdir -p "$(dirname "$DESKTOP_CONFIG")"
        
        cat << EOF > "$DESKTOP_CONFIG"
{
  "mcpServers": {
    "rodmcp": {
      "command": "$RODMCP_PATH",
      "args": [
        "--headless=false",
        "--debug=false",
        "--log-level=info",
        "--slow-motion=500ms",
        "--window-width=1200",
        "--window-height=800"
      ]
    }
  }
}
EOF
        echo "✅ Claude Desktop configured for visible browser"
        echo "⚠️  Please restart Claude Desktop for changes to take effect"
        ;;
        
    2) 
        echo "💻 Configuring Claude Code for visible browser..."
        CODE_CONFIG="$HOME/.config/claude/mcp_servers.json"
        mkdir -p "$(dirname "$CODE_CONFIG")"
        
        cat << EOF > "$CODE_CONFIG"
{
  "rodmcp": {
    "command": "$RODMCP_PATH",
    "args": [
      "--headless=false",
      "--debug=false",
      "--log-level=info",
      "--slow-motion=500ms",
      "--window-width=1200",
      "--window-height=800"
    ],
    "env": {
      "DISPLAY": ":0"
    }
  }
}
EOF
        echo "✅ Claude Code configured for visible browser"
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
        "--headless=false",
        "--debug=false",
        "--log-level=info",
        "--slow-motion=500ms",
        "--window-width=1200",
        "--window-height=800"
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
      "--headless=false",
      "--debug=false",
      "--log-level=info",
      "--slow-motion=500ms",
      "--window-width=1200",
      "--window-height=800"
    ],
    "env": {
      "DISPLAY": ":0"
    }
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
echo "🎬 Browser Visibility Configuration:"
echo "   • Headless mode: DISABLED"
echo "   • Browser window: VISIBLE"
echo "   • Window size: 1200x800"
echo "   • Slow motion: 500ms (for visibility)"
echo "   • RodMCP path: $RODMCP_PATH"
echo ""

# Check for display environment (Linux)
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    if [ -z "$DISPLAY" ]; then
        echo "⚠️  Warning: DISPLAY environment variable not set"
        echo "   For GUI display on Linux, you may need:"
        echo "   export DISPLAY=:0"
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
echo "   Run: ./configs/setup-headless-browser-local.sh"
echo ""

echo "🎉 Configuration complete!"