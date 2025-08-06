# Browser Visibility Configuration

Control whether Claude shows the browser window when working on web development tasks.

## 🖥️ Show Browser Window (Visible Mode)

**Benefits:**
- ✅ Watch Claude work in real-time
- ✅ See page loading and interactions
- ✅ Debug visual issues as they happen
- ✅ Educational - learn how automation works
- ✅ Impressive demonstrations

**Use Cases:**
- Learning web development
- Debugging visual issues
- Demonstrating to others
- Understanding automation workflows
- Development and testing

### Quick Setup

```bash
# Run the automated setup
./configs/setup-visible-browser.sh
```

### Manual Configuration

**Claude Desktop** (`~/.config/claude-desktop/config.json`):
```json
{
  "mcpServers": {
    "rodmcp": {
      "command": "/usr/local/bin/rodmcp",
      "args": [
        "--headless=false",
        "--slow-motion=500ms",
        "--window-width=1200",
        "--window-height=800"
      ]
    }
  }
}
```

**Claude Code** (`~/.config/claude/mcp_servers.json`):
```json
{
  "rodmcp": {
    "command": "/usr/local/bin/rodmcp",
    "args": [
      "--headless=false",
      "--slow-motion=500ms",
      "--window-width=1200",
      "--window-height=800"
    ],
    "env": {
      "DISPLAY": ":0"
    }
  }
}
```

## 🚫 Hidden Browser (Headless Mode)

**Benefits:**
- ⚡ Faster execution
- 💾 Lower resource usage
- 🔒 No GUI interference
- 📊 Better for automation
- 🚀 Production-ready

**Use Cases:**
- Production workflows
- Automated testing
- Background processing
- Server environments
- Resource-constrained systems

### Quick Setup

```bash
# Run the automated setup
./configs/setup-headless-browser.sh
```

### Manual Configuration

**Claude Desktop** (`~/.config/claude-desktop/config.json`):
```json
{
  "mcpServers": {
    "rodmcp": {
      "command": "/usr/local/bin/rodmcp",
      "args": [
        "--headless=true",
        "--window-width=1920",
        "--window-height=1080"
      ]
    }
  }
}
```

## ⚙️ Configuration Options

### Window Size
```bash
--window-width=1200    # Browser width in pixels
--window-height=800    # Browser height in pixels
```

Common sizes:
- Desktop: `1920x1080`
- Laptop: `1366x768`
- Tablet: `1024x768`
- Mobile: `375x667`

### Animation Speed
```bash
--slow-motion=500ms    # Delay between actions (visible mode)
--slow-motion=0        # No delay (headless mode)
```

### Debug Mode
```bash
--debug=true          # Enable Chrome DevTools
--debug=false         # Normal mode
```

### Logging
```bash
--log-level=debug     # Verbose logging
--log-level=info      # Standard logging
--log-level=warn      # Warnings only
--log-level=error     # Errors only
```

## 🎬 Example Workflows

### Educational Demo
Ask Claude:
```
Create an interactive todo app with animations, and show me the browser 
while you work so I can see each step of the development process.
```

### Production Testing
Ask Claude:
```
Run automated tests on my website in headless mode and generate a report 
with screenshots of any issues found.
```

### Visual Debugging
Ask Claude:
```
My CSS layout is broken. Show me the browser window while you inspect 
and fix the layout issues so I can understand what went wrong.
```

### Performance Testing
Ask Claude:
```
Test my website's loading speed with the browser visible so I can see 
how the page renders and identify bottlenecks.
```

### Adaptive Automation
Ask Claude:
```
Create and test my website. Show me the browser during development so I 
can see the progress, then switch to headless mode for the final testing 
to speed things up.
```

## 🔧 Switching Between Modes

### Quick Switch Scripts

**Enable Visible Mode:**
```bash
./configs/setup-visible-browser.sh
```

**Enable Headless Mode:**
```bash
./configs/setup-headless-browser.sh
```

### Runtime Control (NEW!)
Claude can now dynamically control browser visibility using the `set_browser_visibility` tool:

```
"Show me the browser while you work on this task"
```

```
"Switch to headless mode for faster execution"
```

```
"Make the browser visible so I can see the animations, then switch back to headless"
```

**How it works:**
- Claude automatically restarts the browser with new visibility settings
- All open pages are preserved and restored after the change
- No need to manually restart the MCP server
- Works seamlessly during active development sessions

## 🐛 Troubleshooting

### Linux Display Issues
If browser doesn't appear on Linux:
```bash
export DISPLAY=:0
```

Or install virtual display:
```bash
sudo apt-get install xvfb
```

### macOS Permission Issues
Grant accessibility permissions:
1. System Preferences → Security & Privacy → Privacy
2. Select "Accessibility" 
3. Add Terminal/Claude application

### Windows Display Issues
Ensure Windows Subsystem for Linux (WSL) has GUI support:
```bash
# Install WSLg or use X11 forwarding
```

### Performance Issues
For slower systems, use:
```bash
--headless=true          # Hide browser
--slow-motion=0          # No delays
--window-width=800       # Smaller window
--window-height=600
```

## 📊 Comparison

| Feature | Visible Mode | Headless Mode |
|---------|--------------|---------------|
| Speed | Slower | Faster |
| Resources | Higher | Lower |
| Debugging | Excellent | Screenshots only |
| Learning | Great | Limited |
| Production | No | Yes |
| Automation | Good | Excellent |

## 🎯 Recommendations

**For Learning/Development:**
- Use visible mode
- Add slow motion
- Enable debug logging

**For Production/Testing:**
- Use headless mode
- Maximize window size
- Minimal logging

**For Demonstrations:**
- Use visible mode
- Moderate slow motion
- Medium window size

---

⚠️ **Remember**: Always restart Claude after changing browser visibility configuration!