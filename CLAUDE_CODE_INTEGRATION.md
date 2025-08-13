# RodMCP + Claude Code Integration

This guide shows how to integrate RodMCP as an MCP server with Claude Code, giving Claude access to 19 comprehensive web development tools.

## 🚀 Quick Setup

### Step 1: Install RodMCP
```bash
git clone https://github.com/dgallion1/rodMcp.git
cd rodMcp
make install-local
```

### Step 2: Start RodMCP HTTP Server
```bash
# Start RodMCP in HTTP mode (runs in background)
rodmcp --http --port=8090 --headless --log-level=info &

# Verify it's running
curl http://localhost:8090/health
# Should return: {"status": "ok", "tools": 19}
```

### Step 3: Configure Claude Code

Add this to your Claude Code MCP configuration:

**Option A: Direct configuration**
```json
{
  "mcpServers": {
    "rodmcp-web-automation": {
      "url": "http://localhost:8090"
    }
  }
}
```

**Option B: Copy provided config**
```bash
# Copy the ready-made configuration
cp claude-code-config.json ~/.config/claude-code/mcp-servers.json
```

### Step 4: Restart Claude Code
```bash
# Restart Claude Code to load the new MCP server
pkill claude-code
claude-code
```

### Step 5: Verify Integration
Ask Claude: "What tools do you have available?"

You should see RodMCP's 19 tools including:
- 🌐 Browser automation (create_page, navigate_page, screenshot)
- 🖱️ UI control (click, type, wait, scroll)
- 🕷️ Screen scraping (advanced data extraction)
- 📁 File system (read, write, list)
- 🌍 Network requests (HTTP client)

## 🔧 Advanced Configuration

### Custom Chrome Path
```bash
# If using custom Chrome/Chromium path
rodmcp --http --port=8090 --chrome-path="/usr/bin/google-chrome" &
```

### Visible Browser Mode
```bash
# For demos or debugging - show browser window
rodmcp --http --port=8090 --log-level=debug &
```

### Different Port
```bash
# Use different port
rodmcp --http --port=8091 --headless &
```

Then update your Claude Code config:
```json
{
  "mcpServers": {
    "rodmcp-web-automation": {
      "url": "http://localhost:8091"
    }
  }
}
```

## 🔄 Auto-Start Setup

### Linux/macOS Systemd Service
Create `/etc/systemd/user/rodmcp.service`:
```ini
[Unit]
Description=RodMCP Web Automation Server
After=network.target

[Service]
Type=simple
ExecStart=/home/YOUR_USER/.local/bin/rodmcp --http --port=8090 --headless --log-level=info
Restart=always
RestartSec=5
Environment=CHROME_PATH=/usr/bin/chromium-browser

[Install]
WantedBy=default.target
```

Enable auto-start:
```bash
systemctl --user daemon-reload
systemctl --user enable rodmcp.service
systemctl --user start rodmcp.service
```

### macOS LaunchAgent
Create `~/Library/LaunchAgents/com.rodmcp.server.plist`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.rodmcp.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/YOUR_USER/.local/bin/rodmcp</string>
        <string>--http</string>
        <string>--port=8090</string>
        <string>--headless</string>
        <string>--log-level=info</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

Load the service:
```bash
launchctl load ~/Library/LaunchAgents/com.rodmcp.server.plist
```

## 🧪 Testing Integration

### Basic Test
```bash
# 1. Ensure RodMCP is running
curl http://localhost:8090/health

# 2. Test a simple tool call
curl -X POST http://localhost:8090/call \
  -H "Content-Type: application/json" \
  -d '{
    "method": "tools/call",
    "params": {
      "name": "create_page",
      "arguments": {
        "title": "Test Page",
        "html": "<h1>Hello from RodMCP!</h1>"
      }
    }
  }'
```

### Claude Code Test
Ask Claude:
- "Create a simple HTML page and take a screenshot"
- "Navigate to google.com and search for 'web automation'"  
- "Scrape the main heading from this website: https://example.com"

## 🛠️ Troubleshooting

### "Connection refused" errors
```bash
# Check if RodMCP is running
ps aux | grep rodmcp
curl http://localhost:8090/health

# Restart if needed
pkill rodmcp
rodmcp --http --port=8090 --headless &
```

### "Tools not found" in Claude Code
```bash
# Verify Claude Code config location
cat ~/.config/claude-code/mcp-servers.json

# Restart Claude Code
pkill claude-code
claude-code
```

### Port conflicts
```bash
# Check what's using port 8090
sudo lsof -i :8090

# Kill conflicting process or use different port
rodmcp --http --port=8091 --headless &
```

## 🎯 What You Get

Once integrated, Claude gains these capabilities:

### 🌐 **Web Development**
- Create complete HTML/CSS/JS pages instantly
- Live preview with auto-reload server
- Visual debugging with screenshots

### 🤖 **Browser Automation** 
- Navigate websites and interact with elements
- Fill forms, click buttons, test workflows
- Switch between visible/headless modes

### 🕷️ **Advanced Scraping**
- Extract structured data with CSS selectors
- Handle dynamic content and lazy loading
- Single item or bulk extraction

### 📁 **File System**
- Read/write project files
- Manage directory structures
- Process configuration files

### 🌍 **Network**
- Make HTTP requests with full control
- Test APIs and webhooks
- Handle authentication and headers

This transforms Claude Code into a complete web development and automation environment!