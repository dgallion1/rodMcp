# Changelog

All notable changes to RodMCP will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-08-06

### Added
- **Initial Release** - Complete MCP web development controller implementation
- **MCP Protocol Support** - Full 2025-06-18 specification with JSON-RPC 2.0
- **5 Web Development Tools**:
  - `create_page` - Generate HTML pages with CSS and JavaScript
  - `navigate_page` - Browser navigation to URLs or local files  
  - `take_screenshot` - Capture page screenshots
  - `execute_script` - Run JavaScript code in browser
  - `live_preview` - Local HTTP server for live development
- **Rod Browser Integration** - Chrome/Chromium automation via Rod library
- **Dual Installation Methods**:
  - System-wide installation with `install.sh` (requires sudo)
  - Local user installation with `install-local.sh` (no sudo required)
- **Browser Visibility Controls**:
  - Visible browser mode for watching Claude work
  - Headless mode for faster execution
  - Easy switching between modes
- **Comprehensive Logging**:
  - Structured JSON logging with Zap
  - File rotation with Lumberjack
  - Component-specific log filtering
  - Configurable log levels
- **Configuration Management**:
  - Automated Claude Desktop configuration
  - Automated Claude Code configuration  
  - Multiple browser window sizes and speeds
  - Environment-specific settings
- **Documentation**:
  - Installation guides for both methods
  - Browser visibility configuration guide
  - MCP usage examples and workflows
  - Troubleshooting and FAQ sections
- **Examples and Demos**:
  - Interactive browser demo
  - Visual capabilities demonstration
  - LLM interaction examples
  - Test validation suite

### Technical Details
- **Language**: Go 1.24
- **Architecture**: Modular design with internal packages
- **Dependencies**: Rod, Zap, Lumberjack, Gorilla WebSocket
- **Protocol**: MCP 2025-06-18 over stdio transport
- **Browser**: Chrome/Chromium via Chrome DevTools Protocol
- **Logging**: Structured JSON with rotation
- **Installation**: User choice of system-wide or local

### Supported Platforms
- Linux (tested on Ubuntu/WSL2)
- macOS (compatibility via Rod)
- Windows (compatibility via Rod)

### Security Features
- No network server (stdio transport only)
- User-permission-based execution
- Configurable log locations
- Safe browser sandboxing via Chrome

---

## Release Notes

This is the initial release of RodMCP, providing a complete MCP server implementation for web development automation. The project was built to extend Claude's capabilities with browser automation, enabling programmatic web development, testing, and interaction.

### Key Features Delivered
1. **Seamless Claude Integration** - Works out-of-the-box with Claude Desktop and Claude Code
2. **Visual Development** - Watch Claude work in real-time with visible browser mode  
3. **Professional Toolset** - 5 comprehensive web development tools
4. **Flexible Installation** - Choose between system-wide or local user installation
5. **Production Ready** - Comprehensive logging, error handling, and documentation

### Future Roadmap
- Additional web development tools (form testing, performance analysis)
- Enhanced browser configuration options
- Multi-browser support (Firefox, Safari)
- CI/CD integration examples
- Plugin architecture for custom tools