# Local User Installation Guide

Install RodMCP for your user account without requiring administrator privileges.

## Benefits

✅ **No sudo/admin required** - Install in your home directory  
✅ **Isolated installation** - Doesn't affect system files  
✅ **Easy uninstall** - Simple removal script included  
✅ **PATH-friendly** - Uses standard `~/.local/bin` directory  
✅ **Per-user configuration** - Each user has their own setup  

## Installation

### 1. Run the Local Installer

```bash
./install-local.sh
```

This will:
- Build RodMCP from source
- Install to `~/.local/bin/rodmcp`
- Create Claude configuration files
- Generate an uninstall script

### 2. Update Your PATH (if needed)

If `~/.local/bin` is not in your PATH, add it:

**For bash:**
```bash
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
source ~/.bashrc
```

**For zsh:**
```bash
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.zshrc
source ~/.zshrc
```

### 3. Browser Visibility Control

**No configuration needed!** Claude can dynamically control browser visibility using the `set_browser_visibility` tool.

Just ask Claude:
- "Show me the browser while you work"
- "Switch to headless mode for faster execution"

**Optional manual configuration:**
```bash
./configs/setup-visible-browser-local.sh    # Show browser
./configs/setup-headless-browser-local.sh   # Hide browser
```

## File Locations

After local installation, files are located at:

| Component | Location |
|-----------|----------|
| RodMCP binary | `~/.local/bin/rodmcp` |
| Claude Desktop config | `~/.config/claude-desktop/config.json` |
| Claude Code config | `~/.config/claude/mcp_servers.json` |
| Uninstall script | `~/.local/bin/uninstall-rodmcp` |
| Log files | `/tmp/rodmcp-logs/` (default) |

## Testing

After installation, test RodMCP:

```bash
# Check if installed
rodmcp --help

# Or use full path if PATH not updated
~/.local/bin/rodmcp --help
```

## Using with Claude

1. **Restart Claude Desktop or Claude Code**
2. **Ask Claude:** "What web development tools do you have?"
3. **Test with:** "Create a simple HTML page and take a screenshot"

## Customization

### Change Log Directory

Edit your Claude config to specify a custom log directory:

```json
{
  "rodmcp": {
    "command": "~/.local/bin/rodmcp",
    "args": [
      "--log-dir=~/rodmcp-logs"
    ]
  }
}
```

### Adjust Browser Window Size

For visible browser mode:

```json
{
  "args": [
    "--headless=false",
    "--window-width=1600",
    "--window-height=900"
  ]
}
```

### Enable Debug Mode

For troubleshooting:

```json
{
  "args": [
    "--debug=true",
    "--log-level=debug"
  ]
}
```

## Uninstallation

To completely remove RodMCP:

```bash
uninstall-rodmcp
```

This will:
- Remove the RodMCP binary
- Backup configuration files (with `.uninstalled` suffix)
- Remove the uninstall script itself

To restore after uninstalling:
```bash
mv ~/.config/claude-desktop/config.json.uninstalled ~/.config/claude-desktop/config.json
mv ~/.config/claude/mcp_servers.json.uninstalled ~/.config/claude/mcp_servers.json
```

## Troubleshooting

### Command Not Found

If `rodmcp` is not found after installation:

```bash
# Use full path
~/.local/bin/rodmcp --help

# Or add to PATH for current session
export PATH="$PATH:$HOME/.local/bin"
```

### Permission Denied

```bash
# Make sure the binary is executable
chmod +x ~/.local/bin/rodmcp
```

### Browser Not Showing (Linux)

For GUI display on Linux:

```bash
# Set display variable
export DISPLAY=:0

# Or install virtual display
sudo apt-get install xvfb
```

### Configuration Not Working

Check config file locations:

```bash
# Claude Desktop
cat ~/.config/claude-desktop/config.json

# Claude Code
cat ~/.config/claude/mcp_servers.json
```

### Chrome Download Issues

Rod automatically downloads Chrome on first run. If it fails:

1. Check internet connection
2. Check disk space (~100MB needed)
3. Try manual Chrome installation:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install chromium-browser
   
   # macOS
   brew install --cask chromium
   ```

## Advantages Over System Installation

| Feature | Local Install | System Install |
|---------|--------------|----------------|
| Requires sudo | ❌ No | ✅ Yes |
| Install location | `~/.local/bin` | `/usr/local/bin` |
| Config location | `~/.config/` | `~/.config/` |
| Per-user setup | ✅ Yes | ❌ No |
| Easy removal | ✅ Script | ⚠️ Manual |
| PATH setup | May need update | Usually works |

## Security Notes

- RodMCP runs with your user permissions only
- No system-wide changes are made
- Browser runs in your user context
- Logs are written to user-accessible directories

---

💡 **Tip**: Local installation is recommended for personal use and development. Use system-wide installation for shared servers or production environments.