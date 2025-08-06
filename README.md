# RodMCP - MCP Web Development Controller

A Go-based Model Context Protocol (MCP) server that provides web development tools using the Rod browser automation library. Built for Claude and other MCP clients to enable programmatic web development, testing, and automation.

## 🌟 Highlights

- 🤖 **Works with Claude** - Full MCP protocol support for seamless integration
- 🎬 **Visible Browser Mode** - Watch Claude work in real-time or run headless
- 🛠️ **5 Powerful Tools** - Create pages, navigate, screenshot, execute JS, live preview
- 🏠 **Easy Install** - No sudo required with local user installation
- ⚡ **Go 1.24 Performance** - Fast, reliable browser automation

## 🛠️ Available Tools

Once installed, Claude gains access to these powerful web development tools:

### 📝 `create_page`
Generate complete HTML pages with embedded CSS and JavaScript
- **Purpose**: Rapid prototyping and page creation
- **Example**: "Create a responsive landing page for a coffee shop"

### 🌐 `navigate_page`  
Open URLs or local files in the browser
- **Purpose**: Load pages for testing and interaction
- **Example**: "Navigate to my website and test the contact form"

### 📸 `take_screenshot`
Capture visual snapshots of web pages
- **Purpose**: Visual validation and documentation
- **Example**: "Take a screenshot of the page after applying dark mode"

### ⚡ `execute_script`
Run JavaScript code in browser pages
- **Purpose**: Dynamic interaction and testing
- **Example**: "Click all buttons and test form validation"

### 🚀 `live_preview`
Start local development server with auto-reload
- **Purpose**: Live development and multi-page testing
- **Example**: "Create a website and start preview server"

## 🎬 Demo

Watch RodMCP in action:

```bash
./bin/demo
```

This shows Claude creating an interactive webpage, opening it in a visible browser, clicking buttons, changing themes, and taking screenshots - all automated!

## 📦 Quick Start

### 1. Get RodMCP
```bash
git clone <repository-url>
cd rodmcp
```

### 2. Install (Choose One)

**🏠 Local Installation (Recommended - No sudo required):**
```bash
./install-local.sh
./configs/setup-visible-browser-local.sh  # To watch Claude work
```

**🌍 System-Wide Installation:**
```bash
./install.sh
./configs/setup-visible-browser.sh  # To watch Claude work
```

### 3. Test with Claude
Ask Claude: *"What web development tools do you have available?"*

Claude should respond with the 5 RodMCP tools listed above.

## 💡 Example Use Cases

### 🎨 Creative Web Development
```
"Create a portfolio website with a dark theme, smooth animations, 
and a contact form. Show me the browser while you work."
```

### 🧪 Automated Testing  
```
"Navigate to localhost:3000 and test all the buttons and forms. 
Take screenshots of any issues you find."
```

### 📱 Responsive Design
```
"Create a mobile-friendly dashboard and test it at different 
screen sizes. Show me how it looks on tablet and phone."
```

### 🚀 Live Development
```
"Build an interactive todo app with local storage, start a 
preview server, and demonstrate all the features working."
```

### 🎓 Learning Tool
```
"Create an interactive CSS tutorial showing flexbox examples. 
Use visible browser mode so I can see each step."
```

## Architecture

```
MCP Client ←→ JSON-RPC 2.0 ←→ MCP Server
                                    ├── Web Dev Tools
                                    ├── Browser Manager (Rod)
                                    └── Logging System
```

## Tools Documentation

### create_page
Creates HTML pages with embedded CSS and JavaScript.

**Parameters:**
- `filename` (required): Output HTML filename
- `title` (required): Page title
- `html` (required): Body HTML content
- `css` (optional): CSS styles
- `javascript` (optional): JavaScript code

### navigate_page
Navigates browser to a URL or local file.

**Parameters:**
- `url` (required): URL or file path to navigate to

### take_screenshot
Captures a screenshot of the current page.

**Parameters:**
- `page_id` (optional): Specific page ID to screenshot
- `filename` (optional): Save screenshot to file

### execute_script
Executes JavaScript code in the browser.

**Parameters:**
- `page_id` (optional): Page ID to execute script in
- `script` (required): JavaScript code to execute

### live_preview
Starts a local HTTP server for live development.

**Parameters:**
- `directory` (optional): Directory to serve (default: current)
- `port` (optional): Server port (default: 8080)

## Development

### Project Structure

```
rodmcp/
├── cmd/server/          # Main server application
├── internal/
│   ├── browser/         # Rod browser management
│   ├── logger/          # Logging system
│   ├── mcp/            # MCP protocol implementation
│   └── webtools/       # Web development tools
├── pkg/types/          # Shared type definitions
└── examples/           # Usage examples
```

### Adding New Tools

1. Implement the `Tool` interface:
```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() types.ToolSchema
    Execute(args map[string]interface{}) (*types.CallToolResponse, error)
}
```

2. Register the tool in `cmd/server/main.go`:
```go
mcpServer.RegisterTool(webtools.NewYourTool(log, browserMgr))
```

## Contributing

1. Follow Go best practices and conventions
2. Add comprehensive logging for debugging
3. Include proper error handling
4. Write tests for new functionality
5. Update documentation

## 🎬 Browser Visibility

RodMCP can run in two modes:

### 👀 Visible Browser Mode
- **See Claude work in real-time**
- Great for learning and debugging
- Configure with: `./configs/setup-visible-browser-local.sh`

### ⚡ Headless Mode  
- **Faster execution, no GUI**
- Better for automation and production
- Configure with: `./configs/setup-headless-browser-local.sh`

### Switch Anytime
You can easily switch between modes - just run the appropriate setup script and restart Claude.

## 📚 Documentation

- **[Installation Guide](INSTALLATION.md)** - Complete setup and configuration
- **[Local Install Guide](LOCAL_INSTALL.md)** - Install without sudo (recommended)
- **[Browser Visibility](BROWSER_VISIBILITY.md)** - Control browser display modes
- **[MCP Usage Examples](MCP_USAGE.md)** - How to use with Claude effectively

## 📋 System Requirements

- **Go 1.24+** (for building from source)
- **Chrome/Chromium** (automatically downloaded by Rod on first use)
- **Platform**: Linux, macOS, or Windows
- **RAM**: 512MB+ recommended
- **Disk**: ~100MB for Chrome download

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Adding new web development tools
- Reporting bugs and requesting features  
- Improving documentation
- Testing and platform support

## 📋 Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed release history and changes.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.