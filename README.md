# RodMCP - MCP Web Development Controller

A Go-based Model Context Protocol (MCP) server that provides web development tools using the Rod browser automation library. Built for Claude and other MCP clients to enable programmatic web development, testing, and automation.

## 🌟 Highlights

- 🤖 **Works with Claude** - Full MCP protocol support for seamless integration
- 🔄 **Robust Connection Management** - Automatic reconnection and health monitoring prevents timeout errors
- 🎬 **Visible Browser Mode** - Watch Claude work in real-time or run headless (browser visibility fixed!)
- 🛠️ **19 Comprehensive Tools** - Complete browser control + screen scraping + file system + HTTP requests + interactive help
- 🏠 **Easy Install** - No sudo required with local user installation
- 🚀 **Auto Go Install** - Makefile can install Go locally if not present
- ⚡ **Go 1.24.5+ Performance** - Fast, reliable browser automation

## 🛠️ Available Tools

Once installed, Claude gains access to these 19 comprehensive web development tools:

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

### 🕷️ Screen Scraping Tools

### 📊 `screen_scrape`
Extract structured data from web pages using CSS selectors with advanced scraping capabilities
- **Purpose**: Automated data extraction from websites and web applications
- **Features**: Single & multiple item extraction, dynamic content waiting, lazy loading support, custom JavaScript execution
- **CSS Selectors**: Supports #id, .class, [attribute], :nth-child(), descendant combinators, and complex selectors
- **Advanced**: Wait for dynamic content, trigger lazy loading, execute custom scripts before scraping
- **Use Cases**: Product catalogs, news articles, search results, form data, image galleries, API alternatives
- **Examples**: 
  - Single item: "Extract the main article title, content, and author from this blog post"
  - Multiple items: "Scrape all product cards with name, price, and image from this shopping page"
  - Dynamic content: "Wait for the AJAX content to load, then extract the data"

### ❓ Help & Discovery Tools

### 💡 `help`
Get interactive help, usage examples, and workflow suggestions for rodmcp tools
- **Purpose**: Discover tool capabilities and learn effective usage patterns
- **Example**: "Get help for create_page" or "Show me common workflows"

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

## 🤖 Claude Code Integration

**RodMCP works perfectly with Claude Code!** Get instant access to 19 web development tools:

### Quick Setup
```bash
# 1. Install RodMCP
git clone https://github.com/dgallion1/rodMcp.git && cd rodMcp && make install-local

# 2. Start HTTP server for Claude Code
./start-for-claude-code.sh

# 3. Configure Claude Code (add to mcp-servers.json)
{"mcpServers": {"rodmcp": {"url": "http://localhost:8090"}}}

# 4. Restart Claude Code and ask: "What tools do you have available?"
```

**🎯 What You Get:**
- 🌐 Complete HTML/CSS/JS page creation with live preview
- 🖱️ Full browser automation (click, type, navigate, screenshot)
- 🕷️ Advanced web scraping with CSS selectors
- 📁 File system access for project management
- 🌍 HTTP client for API testing

See **[CLAUDE_CODE_INTEGRATION.md](CLAUDE_CODE_INTEGRATION.md)** for detailed setup, auto-start configuration, and troubleshooting.

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
make install-local  # Automatically stops existing processes
```

**🌍 System-Wide Installation:**
```bash
make install  # Automatically stops existing processes (requires sudo)
```

**🛠️ Development Build:**
```bash
make build
```

### 4. Configure Claude Code (Choose Connection Mode)

RodMCP supports two connection modes: **stdio** (recommended) and **HTTP**. Choose based on your needs:

#### 📡 **Stdio Mode (Recommended)**
**⚠️ Use the direct `rodmcp` binary, NOT `rodmcp-manager`:**

```bash
# Add RodMCP to Claude Code (use the DIRECT binary)
claude mcp add-json rodmcp '{"type": "stdio", "command": "'"$HOME"'/.local/bin/rodmcp", "args": ["--headless", "--log-level=info"], "env": {}}'

# Verify it's working
claude mcp list  # Should show ✓ Connected
```

**🚫 Common Mistake**: Never use `rodmcp-manager` as it causes connection conflicts by running multiple instances simultaneously.

#### 🌐 **HTTP Mode (Alternative)**
For environments where stdio doesn't work or for debugging purposes:

**1. Start RodMCP in HTTP mode:**
```bash
# Run in background (recommended)
rodmcp --http --port=8090 --headless --log-level=info &

# Or run in foreground for debugging
rodmcp --http --port=8090 --headless --log-level=debug
```

**2. Configure Claude Code for HTTP:**
```bash
# Add RodMCP HTTP server to Claude Code
claude mcp add-json rodmcp-http '{"type": "http", "url": "http://localhost:8090", "env": {}}'

# Verify connection
claude mcp list  # Should show ✓ Connected
```

**HTTP Mode Features:**
- Same 19 tools as stdio mode
- RESTful API at `http://localhost:8090`
- Useful for debugging with browser dev tools
- Can be accessed by multiple clients simultaneously
- Optional authentication via environment variables

**Stop HTTP server:**
```bash
# Stop background process
pkill -f "rodmcp.*http"

# Or use built-in process management
make stop-processes
```

### 5. Test with Claude
Ask Claude: *"What web development tools do you have available?"*

Claude should respond with the 19 RodMCP tools listed above.

You can also ask: *"Help me get started with rodmcp"* and Claude will use the interactive help system to guide you.

## 🏆 Why Choose RodMCP Over Playwright?

While Playwright is excellent for traditional automation, RodMCP is **specifically designed for AI integration** with unique advantages:

### 🎯 **Built for AI from Day One**
- **Native MCP Integration** - Purpose-built for Claude and AI agents, not adapted later
- **Zero Configuration** - Works instantly with Claude Desktop/CLI without complex setup
- **AI-Optimized API** - Tools designed for natural language commands, not programmatic scripts

### 🚀 **Superior Performance & Resource Efficiency**
- **Lightweight Go Runtime** - Single binary, minimal memory footprint vs Node.js + dependencies
- **Faster Startup** - Instant browser launching vs Playwright's heavier initialization
- **Lower Resource Usage** - Rod's native Chrome integration vs Playwright's multiple browser engines

### 🛠️ **Complete Development Toolkit**
- **19 Comprehensive Tools** - Browser automation + screen scraping + file system + HTTP + help system
- **Advanced Screen Scraping** - Single & multiple item extraction with smart element detection
- **Integrated Web Server** - Built-in `live_preview` for instant development servers
- **Page Creation Tools** - Generate complete HTML/CSS/JS pages directly
- **Interactive Help System** - AI-guided tool discovery and workflow suggestions

### 🎬 **Dynamic Visibility Control**
- **Runtime Browser Switching** - Switch visible/headless during operation without restart
- **Watch Claude Work** - See automation in real-time for learning and debugging
- **Adaptive Automation** - Visible for demos, headless for speed - all in one session

### 🏠 **Installation & Deployment**
- **No sudo Required** - Local user installation option
- **Single Binary** - No Node.js, npm, or package management complexity
- **Auto Go Installation** - Can install Go locally if not present
- **Zero Dependencies** - Self-contained executable

### 💬 **Natural AI Interaction**
```
❌ Playwright MCP: Complex setup, requires Node.js ecosystem
✅ RodMCP: "Help me create a coffee shop landing page with contact form"

❌ Playwright: Separate tools for different tasks
✅ RodMCP: "Create the page, start preview server, and take screenshot"

❌ Playwright: Limited real-time visibility options  
✅ RodMCP: "Show me the browser while you work, then switch to headless"
```

### 🎓 **Learning & Development**
- **Built-in Guidance** - Interactive help system with workflow suggestions
- **Real-time Observation** - Watch and learn from Claude's automation techniques
- **Complete Examples** - API reference with concrete usage patterns
- **Educational Focus** - Perfect for learning web development with AI assistance

### 🔒 **Enterprise Ready**
- **MIT License** - Clean licensing across all dependencies
- **Go Security** - Memory-safe language with excellent security track record
- **No External APIs** - Runs completely locally with stdio transport
- **Comprehensive Testing** - 25+ automated tests with 100% success rate

### ⚡ **Technical Advantages**
| Feature | RodMCP | Playwright MCP |
|---------|--------|----------------|
| **Language** | Go (fast, compiled) | Node.js (interpreted) |
| **Memory Usage** | Low | High |
| **Startup Time** | Instant | Slow |
| **AI Integration** | Native MCP design | Retrofitted |
| **Installation** | Single binary | npm + dependencies |
| **Dependencies** | Self-contained | Requires Node.js ecosystem |
| **Browser Support** | Chrome/Chromium (optimized) | Multi-browser (heavier) |
| **Real-time Visibility** | Dynamic switching | Limited |
| **Development Tools** | Integrated | Requires external tools |

### 🎯 **Perfect For**
- **AI-First Development** - Building with Claude as your development partner
- **Rapid Prototyping** - From idea to working webpage in minutes
- **Learning Web Development** - Watch AI techniques in real-time
- **Local Development** - Complete toolkit without cloud dependencies
- **Educational Use** - Teaching automation and web development
- **Lightweight Automation** - When you need speed and efficiency

### 🔄 **Migration from Playwright**
RodMCP provides equivalent functionality with AI-optimized design:
- **Page Navigation** ✅ More intuitive with file:// support
- **Element Interaction** ✅ Simplified CSS selector approach  
- **Screenshots** ✅ Built-in with flexible output options
- **JavaScript Execution** ✅ Direct execution with result handling
- **Form Testing** ✅ Type text + click elements seamlessly
- **Data Extraction** ✅ Advanced screen scraping with single & multiple item support
- **API Testing** ✅ Built-in HTTP request tools

**The Bottom Line:** If you're working with AI agents like Claude, RodMCP delivers a **native, optimized experience** that Playwright MCP can't match. Built for AI from the ground up with advanced screen scraping, not adapted afterward.

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

RodMCP includes a comprehensive test suite that validates all 19 MCP tools across 5 categories:

### Run Comprehensive Test Suite
```bash
go run comprehensive_suite.go
```

**Test Coverage:**
- **📁 File System Tools** (4 tests): Write files with directory creation, read files, list directories, write JSON files
- **🌐 Browser Automation** (4 tests): Create HTML pages, navigate to pages, take screenshots, execute JavaScript
- **🖱️ UI Control Tools** (10 tests): Click elements, type text, wait operations, get element text/attributes, scroll, hover, form interactions
- **🕷️ Screen Scraping Tools** (2 tests): Single item extraction, multiple item extraction with containers
- **🌍 Network Tools** (3 tests): HTTP GET, POST with JSON, custom headers
- **⚡ JavaScript Execution** (4 tests): Complex object returns, DOM manipulation, async operations, error handling

**Features:**
- ✅ **100% Success Rate** - All 27 tests pass
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

### 🕷️ Data Extraction & Analysis
```
"Navigate to this e-commerce site and extract all product names,
prices, and ratings. Then analyze the pricing trends."
```

### 📊 Content Monitoring
```
"Scrape the latest articles from these news websites and 
create a summary report with headlines and links."
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
                                    ├── Screen Scraping Tools (1)
                                    ├── File System Tools (3)  
                                    ├── Network Tools (1)
                                    ├── Help & Discovery Tools (1)
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

### screen_scrape
Extract structured data from web pages using CSS selectors with advanced scraping capabilities.

**Parameters:**
- `url` (optional): URL to scrape (optional if page_id provided). Example: 'https://example.com/products'
- `page_id` (optional): Existing page ID to scrape from current browser session (optional if url provided). Use for already loaded pages
- `selectors` (required): CSS selectors mapping field names to elements. Examples: {'title': 'h1', 'price': '.price-value', 'description': 'p.desc', 'link': 'a[href]', 'image': 'img[src]', 'rating': '[data-rating]'}
- `extract_type` (optional): Extraction mode - 'single' for one item, 'multiple' for array using container_selector (default: single)
- `container_selector` (required for multiple): Container selector for multiple items. Examples: '.product-card', 'article', '.search-result'
- `wait_for` (optional): CSS selector to wait for before scraping (handles dynamic content). Examples: '.loading-complete', '[data-loaded=true]'
- `wait_timeout` (optional): Maximum seconds to wait for elements (default: 10)
- `include_metadata` (optional): Include page metadata (title, url, timestamp, extraction_time) (default: true)
- `scroll_to_load` (optional): Auto-scroll to trigger lazy loading (infinite scroll, image loading) (default: false)
- `custom_script` (optional): Custom JavaScript to execute before scraping. Examples: button clicks, view changes, content triggers

**CSS Selector Support:** #id, .class, [attribute], tag, :nth-child(), :contains(), descendant combinators, complex selectors

**Returns:** Structured data with extracted content, element attributes, and optional metadata.

**Examples:**
```json
// Single item extraction - Product page
{
  "url": "https://example.com/product/123",
  "selectors": {
    "title": "h1.product-title",
    "price": ".price-current",
    "description": ".product-description p",
    "image": "img.hero-image",
    "rating": "[data-rating]"
  }
}

// Multiple items extraction - Product catalog
{
  "url": "https://store.com/products",
  "extract_type": "multiple",
  "container_selector": ".product-card",
  "selectors": {
    "name": ".product-name",
    "price": ".price",
    "link": "a[href]",
    "image": "img[src]"
  },
  "scroll_to_load": true
}

// Dynamic content with wait and custom script
{
  "url": "https://spa-app.com/data",
  "selectors": {
    "content": ".ajax-loaded-content",
    "status": ".loading-status"
  },
  "wait_for": ".loading-complete",
  "wait_timeout": 15,
  "custom_script": "document.querySelector('.load-more').click();"
}
```

### help
Get interactive help, usage examples, and workflow suggestions for rodmcp tools.

**Parameters:**
- `topic` (optional): Help topic - 'overview', 'workflows', 'examples', or specific tool name (e.g., 'create_page')
- `category` (optional): Tool category - 'browser_automation', 'ui_control', 'file_system', 'network'

**Returns:** Contextual help content with usage examples, common workflows, and tool relationships.

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
make install         # Install system-wide (auto-stops processes)
make install-local   # Install to user bin (auto-stops processes, no sudo)
make stop-processes  # Stop all running rodmcp processes
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

## 🔄 Connection Management

RodMCP includes robust connection management to prevent timeout errors during idle periods:

### 🚀 **Automatic Health Monitoring**
- **Connection Health Checks** - Monitors MCP client connection every 30 seconds
- **Browser Health Monitoring** - Verifies browser connectivity and restarts if needed
- **Automatic Recovery** - Seamlessly restarts browser with page restoration
- **Heartbeat System** - Sends periodic pings to maintain connection during idle periods

### 🛡️ **No More "Not Connected" Errors**
The enhanced connection management prevents the common issue where rodmcp would become unresponsive after periods of inactivity. Now it:
- **Maintains Active Connections** - Non-blocking input processing with timeout handling
- **Proactive Health Checks** - Detects and resolves connection issues before they cause failures
- **Graceful Recovery** - Automatic reconnection and browser restart when problems are detected
- **Activity Tracking** - Monitors client activity and adapts monitoring frequency accordingly

## 🔧 Troubleshooting

### ⚡ Connection Issues: "Not Connected" Error

If you experience "Not connected" errors after periods of inactivity, this is likely due to conflicting processes or old configurations.

#### **Root Cause: rodmcp-manager Conflicts**
The most common issue is using the problematic `rodmcp-manager` script instead of the direct `rodmcp` binary. The manager script tries to run both HTTP and stdio modes simultaneously, causing conflicts.

#### **✅ Solution: Use Direct Binary**
1. **Check your current configuration:**
   ```bash
   claude mcp list
   ```

2. **If you see `rodmcp-manager` in the command, fix it:**
   ```bash
   # Remove the problematic configuration
   claude mcp remove rodmcp
   
   # Add the correct direct binary configuration
   claude mcp add-json rodmcp '{"type": "stdio", "command": "/home/darrell/.local/bin/rodmcp", "args": ["--headless", "--log-level=info"], "env": {}}'
   ```

3. **Clean up any conflicting processes:**
   ```bash
   # Use the built-in process stopping (recommended)
   make stop-processes
   
   # Or manual cleanup
   pkill -f "rodmcp.*http" || echo "No conflicting processes found"
   rm -f /tmp/rodmcp-http-manager.*
   ```

4. **Test the connection:**
   ```bash
   claude mcp list  # Should show ✓ Connected
   ```

#### **🔄 Other Connection Issues**
- **Multiple Instances**: Only run one rodmcp instance at a time
- **Port Conflicts**: Check for conflicting services on ports 8090 or browser debug ports  
- **Outdated Version**: Update to the latest version with `make install-local`
- **Browser Issues**: If browser fails to start, check that Chrome/Chromium is installed

### 🐛 General Troubleshooting Steps

1. **Update to Latest Version:**
   ```bash
   cd /path/to/rodmcp
   git pull origin master
   make install-local  # Automatically stops existing processes
   ```

2. **Check Service Health:**
   ```bash
   claude mcp list  # Verify connection status
   rodmcp --help     # Verify installation
   ```

3. **View Logs for Debugging:**
   ```bash
   # Run with debug logging
   rodmcp --headless --log-level=debug
   ```

4. **Reset Configuration (if needed):**
   ```bash
   # Remove and re-add MCP server
   claude mcp remove rodmcp
   claude mcp add-json rodmcp '{"type": "stdio", "command": "/home/darrell/.local/bin/rodmcp", "args": ["--headless", "--log-level=info"], "env": {}}'
   ```

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
- **[API Reference](API_REFERENCE.md)** - Complete technical documentation for all 19 tools

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