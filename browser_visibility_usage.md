# Browser Visibility Control

The rodMcp server starts in headless mode by default (as configured in `~/.config/claude/mcp_servers.json`), but you can control browser visibility at runtime using the `set_browser_visibility` MCP tool.

## Using the Browser Visibility Tool

### Through Claude (MCP Integration)

When Claude has rodMcp configured as an MCP server, you can request:

```
"Please use the set_browser_visibility tool to make the browser visible"
```

Or more specifically:

```
"Call set_browser_visibility with visible: true and reason: 'Need to see browser for debugging'"
```

### Tool Parameters

- `visible` (boolean, required): 
  - `true` = Show browser window
  - `false` = Run in headless mode
  
- `reason` (string, optional): Explanation for the visibility change (used for logging)

### Example Tool Call Structure

```json
{
  "tool": "set_browser_visibility",
  "arguments": {
    "visible": true,
    "reason": "User requested to see browser window"
  }
}
```

## How It Works

1. The browser manager receives the visibility change request
2. It saves the current state of all open pages
3. Stops the current browser instance
4. Restarts with the new visibility setting
5. Restores all previously open pages

## Important Notes

- The browser must be running (started by rodMcp server) before you can change visibility
- Switching visibility will briefly close and reopen all browser pages
- Pages are automatically restored to their previous URLs after the switch
- The visibility setting persists until changed again or the server restarts

## Troubleshooting

If the browser doesn't appear when set to visible:

1. **Check DISPLAY variable**: Ensure `DISPLAY` is set (usually `:0` or `:1`)
   ```bash
   echo $DISPLAY
   ```

2. **Verify X server**: On Linux/WSL, ensure X server is running
   - For WSL: Use VcXsrv, X410, or WSLg
   - For Linux: X11 or Wayland should be running

3. **Check logs**: Review the rodMcp logs for any errors
   ```bash
   tail -f ~/.config/claude/logs/rodmcp.log
   ```

4. **Test directly**: Run the interactive demo to verify browser can show:
   ```bash
   ./bin/interactive
   ```

## Default Configuration

The MCP server configuration at `~/.config/claude/mcp_servers.json` controls the initial state:

```json
{
  "rodmcp": {
    "command": "/home/darrell/.local/bin/rodmcp",
    "args": [
      "--headless=true",  // Initial state: headless
      "--log-level=info"
    ]
  }
}
```

To start with browser visible by default, remove the `--headless=true` line. However, using the runtime visibility control is more flexible as it doesn't require restarting the MCP server.