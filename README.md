# RodMCP - MCP Web Development Controller

A Go-based Model Context Protocol (MCP) server that provides web development tools using the Rod browser automation library.

## Features

- **MCP Protocol 2025-06-18 Support**: Full JSON-RPC 2.0 implementation
- **Browser Automation**: Chrome/Chromium control via Rod library
- **Web Development Tools**:
  - `create_page`: Generate HTML pages with CSS and JavaScript
  - `navigate_page`: Browser navigation to URLs or local files
  - `take_screenshot`: Capture page screenshots
  - `execute_script`: Run JavaScript code in browser
  - `live_preview`: Local HTTP server for live development
- **Extensive Logging**: Structured JSON logging with file rotation
- **Go 1.24**: Built with the latest Go version

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd rodmcp

# Install dependencies
go mod download

# Build the server
go build -o bin/rodmcp cmd/server/main.go

# Build the test example
go build -o bin/test-example examples/test_example.go
```

## Usage

### As MCP Server

```bash
# Start the MCP server (stdio transport)
./bin/rodmcp

# With custom options
./bin/rodmcp -headless=false -debug=true -log-level=debug
```

### Command Line Options

- `-headless`: Run browser in headless mode (default: true)
- `-debug`: Enable browser debug mode (default: false)
- `-log-level`: Logging level (debug, info, warn, error)
- `-log-dir`: Log directory (default: logs)
- `-slow-motion`: Delay between browser actions
- `-window-width`: Browser window width (default: 1920)
- `-window-height`: Browser window height (default: 1080)

### Test Example

```bash
# Run the comprehensive test
./bin/test-example
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

## Requirements

- Go 1.24+
- Chrome/Chromium (automatically downloaded by Rod)
- Linux/macOS/Windows

## Installation

See [INSTALLATION.md](INSTALLATION.md) for detailed installation instructions.

### Quick Install

```bash
# Option 1: Local user installation (no sudo required)
./install-local.sh
./configs/setup-visible-browser-local.sh  # To see browser

# Option 2: System-wide installation (requires sudo)
./install.sh
./configs/setup-visible-browser.sh  # To see browser
```

## Documentation

- [Installation Guide](INSTALLATION.md) - Complete setup and configuration
- [Local Install Guide](LOCAL_INSTALL.md) - Install without sudo (recommended)
- [Browser Visibility](BROWSER_VISIBILITY.md) - Control browser display
- [MCP Usage Examples](MCP_USAGE.md) - How to use with Claude

## Demo

Run the live browser demo to see RodMCP in action:

```bash
./bin/demo
```

## License

MIT License