#!/bin/bash

# Setup script to configure Claude for headless browser mode (faster, no GUI)

echo "🚫 Configuring Claude for headless browser mode"
echo "==============================================="

# Create headless configurations
cat > /tmp/claude-desktop-headless.json << 'EOF'
{
  "mcpServers": {
    "rodmcp": {
      "command": "/usr/local/bin/rodmcp",
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

cat > /tmp/claude-code-headless.json << 'EOF'
{
  "rodmcp": {
    "command": "/usr/local/bin/rodmcp",
    "args": [
      "--headless=true",
      "--log-level=info",
      "--window-width=1920",
      "--window-height=1080"
    ]
  }
}
EOF

# Detect which Claude client to configure
echo "Which Claude client are you using?"
echo "1) Claude Desktop App"
echo "2) Claude Code (CLI)"
echo "3) Both"
read -p "Choose (1/2/3): " choice

case $choice in
    1)
        echo "📱 Configuring Claude Desktop for headless mode..."
        DESKTOP_CONFIG="$HOME/.config/claude-desktop/config.json"
        mkdir -p "$(dirname "$DESKTOP_CONFIG")"
        cp /tmp/claude-desktop-headless.json "$DESKTOP_CONFIG"
        echo "✅ Claude Desktop configured for headless mode"
        ;;
    2) 
        echo "💻 Configuring Claude Code for headless mode..."
        CODE_CONFIG="$HOME/.config/claude/mcp_servers.json"
        mkdir -p "$(dirname "$CODE_CONFIG")"
        cp /tmp/claude-code-headless.json "$CODE_CONFIG"
        echo "✅ Claude Code configured for headless mode"
        ;;
    3)
        echo "📱💻 Configuring both for headless mode..."
        
        DESKTOP_CONFIG="$HOME/.config/claude-desktop/config.json"
        mkdir -p "$(dirname "$DESKTOP_CONFIG")"
        cp /tmp/claude-desktop-headless.json "$DESKTOP_CONFIG"
        echo "✅ Claude Desktop configured"
        
        CODE_CONFIG="$HOME/.config/claude/mcp_servers.json"
        mkdir -p "$(dirname "$CODE_CONFIG")"
        cp /tmp/claude-code-headless.json "$CODE_CONFIG"
        echo "✅ Claude Code configured"
        ;;
    *)
        echo "❌ Invalid choice. Please run the script again."
        exit 1
        ;;
esac

# Cleanup
rm /tmp/claude-desktop-headless.json /tmp/claude-code-headless.json

echo ""
echo "🚫 Headless Browser Configuration:"
echo "   • Browser window: HIDDEN"
echo "   • Performance: FASTER"
echo "   • Resource usage: LOWER"
echo "   • Screenshots: STILL AVAILABLE"
echo ""

echo "🧪 Test the Configuration:"
echo "   After restarting Claude, ask:"
echo "   'Create an HTML page and take a screenshot of it'"
echo ""

echo "🔧 To Switch Back to Visible Mode:"
echo "   Run: ./configs/setup-visible-browser.sh"
echo ""

echo "🎉 Headless configuration complete!"