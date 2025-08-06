# RodMCP Installation Guide

This guide shows how to install and configure RodMCP as an MCP (Model Context Protocol) server for use with Claude and other MCP clients.

## Installation Methods

### Local User Installation (No sudo required)

```bash
# Install for current user only
./install-local.sh

# Browser visibility is now dynamic - no configuration needed!
# Just ask Claude: "Show me the browser" or "Switch to headless mode"
```

This installs to `~/.local/bin` and doesn't require root access.

### System-Wide Installation (Requires sudo)

### 1. Build RodMCP

```bash
# Clone or navigate to the RodMCP directory
cd /path/to/rodMcp

# Build the server
go build -o bin/rodmcp cmd/server/main.go

# Make it executable and accessible
sudo cp bin/rodmcp /usr/local/bin/rodmcp
# OR add to your PATH
export PATH=$PATH:$(pwd)/bin
```

### 2. Test the Installation

```bash
# Test that RodMCP works
rodmcp --help
```

## MCP Client Configuration

### For Claude Desktop App

Create or edit `~/.config/claude-desktop/config.json`:

```json
{
  "mcpServers": {
    "rodmcp": {
      "command": "/usr/local/bin/rodmcp",
      "args": [
        "--headless=true",
        "--log-level=info",
        "--log-dir=/tmp/rodmcp-logs"
      ]
    }
  }
}
```

### For Claude Code (CLI)

Create or edit `~/.config/claude/mcp_servers.json`:

```json
{
  "rodmcp": {
    "command": "/usr/local/bin/rodmcp",
    "args": [
      "--headless=true", 
      "--log-level=info"
    ],
    "env": {
      "DISPLAY": ":0"
    }
  }
}
```

### For Other MCP Clients

RodMCP uses stdio transport, so any MCP client can connect:

```bash
# Direct connection via stdio
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | rodmcp
```

## Configuration Options

### Command Line Arguments

```bash
rodmcp [OPTIONS]
```

**Available Options:**
- `--headless` - Run browser in headless mode (default: true)
- `--debug` - Enable browser debug mode (default: false)
- `--log-level` - Set logging level: debug, info, warn, error (default: info)
- `--log-dir` - Directory for log files (default: logs)
- `--slow-motion` - Delay between browser actions (default: 0)
- `--window-width` - Browser window width (default: 1920)
- `--window-height` - Browser window height (default: 1080)

### Example Configurations

**Simple Default (Recommended):**
```json
{
  "command": "rodmcp",
  "args": ["--log-level=info"]
}
```
*Claude can dynamically control browser visibility as needed using the `set_browser_visibility` tool.*

**Development Mode (Always Visible):**
```json
{
  "command": "rodmcp",
  "args": [
    "--headless=false",
    "--debug=true", 
    "--log-level=debug",
    "--slow-motion=500ms"
  ]
}
```

**Production Mode (Always Headless):**
```json
{
  "command": "rodmcp",
  "args": [
    "--headless=true",
    "--log-level=warn",
    "--window-width=1920",
    "--window-height=1080"
  ]
}
```

## Available Tools

Once installed, Claude will have access to these web development tools:

### 🛠️ Tool Reference

1. **`create_page`** - Generate HTML pages with CSS and JavaScript
   ```
   Parameters: filename, title, html, css?, javascript?
   ```

2. **`navigate_page`** - Open URLs or local files in browser
   ```
   Parameters: url
   ```

3. **`take_screenshot`** - Capture page screenshots
   ```
   Parameters: page_id?, filename?
   ```

4. **`execute_script`** - Run JavaScript in browser
   ```
   Parameters: page_id?, script
   ```

5. **`set_browser_visibility`** - Control browser visibility at runtime
   ```
   Parameters: visible (boolean), reason?
   ```

6. **`live_preview`** - Start local development server
   ```
   Parameters: directory?, port?
   ```

## Verification

### Test MCP Connection

1. **Start Claude with MCP support**
2. **Ask Claude:** "What web development tools do you have available?"
3. **Claude should respond** with the 6 RodMCP tools listed above

### Test Functionality

Ask Claude to:
```
Create a simple HTML page and take a screenshot of it
```

Claude should be able to:
- Create an HTML file
- Open it in a browser
- Take a screenshot
- Show you the results

## Troubleshooting

### Common Issues

**1. Chrome/Chromium Not Found**
- Rod automatically downloads Chrome on first run
- Check internet connection if download fails

**2. Permission Denied**
```bash
# Make sure rodmcp is executable
chmod +x /usr/local/bin/rodmcp
```

**3. Display Issues (Linux)**
```bash
# For headless mode, you might need virtual display
export DISPLAY=:0
# OR install xvfb for virtual display
sudo apt-get install xvfb
```

**4. MCP Client Can't Find RodMCP**
```bash
# Verify path is correct
which rodmcp
# Update config.json with full path
```

### Debug Mode

Run with verbose logging:
```bash
rodmcp --debug=true --log-level=debug --headless=false
```

Check logs in the specified log directory for detailed information.

### Port Conflicts

If live preview port 8080 is busy:
```
Ask Claude: "Start live preview server on port 3000"
```

## System Requirements

- **Go 1.24+** (for building from source)
- **Chrome/Chromium** (auto-downloaded by Rod)
- **Linux/macOS/Windows** support
- **~100MB disk space** for Chrome download
- **512MB+ RAM** recommended

## Security Notes

- RodMCP runs a local browser instance
- Only operates on local files and specified URLs
- No network server - uses stdio transport only
- Logs may contain sensitive page content
- Consider log rotation and cleanup for production use

## Advanced Configuration

### Custom Browser Path

Set environment variable:
```bash
export ROD_BROWSER_PATH=/path/to/your/chrome
```

### Custom Chrome Flags

RodMCP uses Rod defaults, but you can modify `internal/browser/manager.go` to add custom Chrome flags.

### Log Rotation

Logs automatically rotate when they exceed size limits. Configure in your MCP client:

```json
{
  "args": [
    "--log-dir=/var/log/rodmcp",
    "--log-level=info"
  ]
}
```

---

🎉 **You're all set!** Claude can now use RodMCP for web development tasks in any project.