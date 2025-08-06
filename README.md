# RodMCP - MCP Web Development Controller

A Go-based Model Context Protocol (MCP) server that provides web development tools using the Rod browser automation library. Built for Claude and other MCP clients to enable programmatic web development, testing, and automation.

## 🌟 Highlights

- 🤖 **Works with Claude** - Full MCP protocol support for seamless integration
- 🎬 **Visible Browser Mode** - Watch Claude work in real-time or run headless (browser visibility fixed!)
- 🛠️ **18 Comprehensive Tools** - Complete browser control + file system + HTTP requests
- 🏠 **Easy Install** - No sudo required with local user installation
- 🚀 **Auto Go Install** - Makefile can install Go locally if not present
- ⚡ **Go 1.24.5+ Performance** - Fast, reliable browser automation

## 🛠️ Available Tools

Once installed, Claude gains access to these 18 comprehensive web development tools:

### 🌐 Browser Automation Tools

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

### 👁️ `set_browser_visibility`
Control browser visibility at runtime - switch between visible and headless modes
- **Purpose**: Adaptive automation - visible for demos/debugging, headless for speed
- **Example**: "Show me the browser while you work" or "Switch to headless for faster execution"

### 🚀 `live_preview`
Start local development server with auto-reload
- **Purpose**: Live development and multi-page testing
- **Example**: "Create a website and start preview server"

### 🎯 Browser UI Control Tools

### 🖱️ `click_element`
Click on specific browser elements using CSS selectors
- **Purpose**: Interact with buttons, links, and clickable elements
- **Example**: "Click the submit button"

### ⌨️ `type_text`
Type text into input fields and textareas
- **Purpose**: Fill forms and input fields
- **Example**: "Type my email address in the login field"

### ⏱️ `wait`
Pause execution for a specified number of seconds
- **Purpose**: Wait for animations, loading, or timed events
- **Example**: "Wait 3 seconds for the animation to complete"

### 🔍 `wait_for_element`
Wait for an element to appear in the DOM
- **Purpose**: Handle dynamic content and loading states
- **Example**: "Wait for the success message to appear"

### 📖 `get_element_text`
Extract text content from browser elements
- **Purpose**: Read page content, error messages, or form values
- **Example**: "Get the text from the error message"

### 🏷️ `get_element_attribute`
Get attribute values from browser elements
- **Purpose**: Read href, src, class, or any element attributes
- **Example**: "Get the href attribute from the first link"

### 📜 `scroll`
Scroll the page by pixels or to specific elements
- **Purpose**: Navigate long pages or bring elements into view
- **Example**: "Scroll to the footer section"

### 🎯 `hover_element`
Hover over elements to trigger hover effects
- **Purpose**: Reveal dropdown menus or hover-triggered content
- **Example**: "Hover over the navigation menu"

### 📁 File System Tools

### 📖 `read_file`
Read the contents of any file
- **Purpose**: Load existing files for editing or analysis
- **Example**: "Read index.html and show me the current structure"

### ✏️ `write_file`
Write content to files, creating or overwriting as needed
- **Purpose**: Save HTML, CSS, JS, or any text files
- **Example**: "Save this modified CSS to styles.css"

### 📋 `list_directory`
List directory contents with file details
- **Purpose**: Navigate project structure and find files
- **Example**: "Show me all files in the src/ directory"

### 🌐 Network Tools

### 📡 `http_request`
Make HTTP requests (GET, POST, PUT, DELETE, etc.)
- **Purpose**: Test APIs, webhooks, and web services
- **Example**: "Test the /api/users endpoint with a POST request"

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

### 2. Check Go Installation
```bash
make check-go
```

If Go is not installed, install it locally (no sudo required):
```bash
make install-go
```

### 3. Build & Install (Choose One)

**🏠 Local Installation (Recommended - No sudo required):**
```bash
make install-local
```

**🌍 System-Wide Installation:**
```bash
make install
```

**🛠️ Development Build:**
```bash
make build
```

That's it! No additional configuration needed. Claude can dynamically control browser visibility.

### 4. Test with Claude
Ask Claude: *"What web development tools do you have available?"*

Claude should respond with the 18 RodMCP tools listed above.

## 💡 Example Use Cases

### 🎨 Creative Web Development
```
"Read my existing index.html file, create an enhanced version with 
a dark theme and animations, then show me the result in the browser."
```

### 🧪 Automated Testing  
```
"Navigate to localhost:3000, test all API endpoints with HTTP requests,
then take screenshots of any issues you find."
```

## 🧪 Testing

RodMCP includes a comprehensive test suite that validates all 18 MCP tools across 5 categories:

### Run Comprehensive Test Suite
```bash
go run comprehensive_suite.go
```

**Test Coverage:**
- **📁 File System Tools** (4 tests): Write files with directory creation, read files, list directories, write JSON files
- **🌐 Browser Automation** (4 tests): Create HTML pages, navigate to pages, take screenshots, execute JavaScript
- **🖱️ UI Control Tools** (10 tests): Click elements, type text, wait operations, get element text/attributes, scroll, hover, form interactions
- **🌍 Network Tools** (3 tests): HTTP GET, POST with JSON, custom headers
- **⚡ JavaScript Execution** (4 tests): Complex object returns, DOM manipulation, async operations, error handling

**Features:**
- ✅ **100% Success Rate** - All 25 tests pass
- ⏱️ **Performance Metrics** - Detailed timing for each operation
- 📊 **Category Summaries** - Success rates per tool category
- 🎯 **Comprehensive Coverage** - Tests every MCP tool with real scenarios
- 🔧 **Automated Validation** - Verifies element text extraction, attribute reading, HTTP responses, and JavaScript execution

The test suite creates an interactive HTML test page, navigates to it, performs UI interactions, validates responses, and confirms all functionality works as expected.

### 📝 Full-Stack Development
```
"List all files in my project, read the API docs, create a new feature 
page, and test it with HTTP requests to the backend."
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
                                    ├── Browser Automation Tools (6)
                                    ├── Browser UI Control Tools (8)
                                    ├── File System Tools (3)  
                                    ├── Network Tools (1)
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

### read_file
Reads the contents of a file.

**Parameters:**
- `path` (required): Path to the file to read

**Returns:** File contents as text with metadata including file size and encoding.

### write_file
Writes content to a file, creating or overwriting as needed.

**Parameters:**
- `path` (required): Path to the file to write
- `content` (required): Content to write to the file
- `create_dirs` (optional): Create parent directories if they don't exist (default: false)

**Returns:** Success message with file size and path information.

### list_directory
Lists the contents of a directory.

**Parameters:**
- `path` (optional): Path to the directory to list (default: current directory)
- `show_hidden` (optional): Include hidden files starting with '.' (default: false)

**Returns:** Formatted directory listing with file types, sizes, and modification dates.

### http_request
Makes HTTP requests to URLs.

**Parameters:**
- `url` (required): URL to request
- `method` (optional): HTTP method - GET, POST, PUT, DELETE, etc. (default: GET)
- `headers` (optional): HTTP headers as key-value pairs
- `body` (optional): Request body for POST/PUT requests
- `json` (optional): JSON data to send (automatically sets Content-Type)
- `timeout` (optional): Request timeout in seconds (default: 30)

**Returns:** HTTP response with status, headers, and body content.

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
├── examples/           # Usage examples
├── configs/            # Configuration scripts
├── bin/                # Built binaries and scripts
└── Makefile           # Build and development automation
```

### Build Commands

Use the Makefile for common development tasks:

```bash
make check-go        # Check Go installation status
make install-go      # Install Go locally (no sudo)
make build           # Build the binary
make install         # Install system-wide
make install-local   # Install to user bin (no sudo)
make clean          # Clean build artifacts
make test           # Run tests
make test-comprehensive # Run comprehensive test suite
make demo           # Run demo
make config-visible # Configure visible browser
make config-headless # Configure headless browser
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
- Configure with: `make config-visible`

### ⚡ Headless Mode  
- **Faster execution, no GUI**
- Better for automation and production
- Configure with: `make config-headless`

### Switch Anytime
You can easily switch between modes - just run `make config-visible` or `make config-headless` and restart Claude.

### ⚠️ Important: Browser Visibility Fix
If the browser window doesn't appear in visible mode, ensure you've built with the latest version that includes the browser visibility fix:

```bash
make clean && make build && make install-local
```

This fix resolves Chrome's `--no-startup-window` flag that was preventing the browser window from appearing.

## 📚 Documentation

- **[Installation Guide](INSTALLATION.md)** - Complete setup and configuration
- **[Local Install Guide](LOCAL_INSTALL.md)** - Install without sudo (recommended)
- **[Browser Visibility](BROWSER_VISIBILITY.md)** - Control browser display modes
- **[MCP Usage Examples](MCP_USAGE.md)** - How to use with Claude effectively

## 📋 System Requirements

- **Go 1.24.5+** (can be installed locally with `make install-go`)
- **Chrome/Chromium** (automatically downloaded by Rod on first use)
- **Platform**: Linux, macOS, or Windows
- **RAM**: 512MB+ recommended
- **Disk**: ~100MB for Chrome download + ~150MB for Go (if installed locally)
- **Network**: Required for Go and Chrome downloads

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