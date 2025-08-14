# RodMCP API Reference

Complete reference documentation for all 26 RodMCP tools, organized by category with detailed parameters, examples, and usage patterns.

## 🌐 Browser Automation Tools

### create_page
**Purpose:** Generate complete HTML pages with embedded CSS and JavaScript.

**Parameters:**
- `filename` (required): Output HTML filename
- `title` (required): Page title
- `html` (required): Body HTML content  
- `css` (optional): CSS styles
- `javascript` (optional): JavaScript code

**Returns:** Success message with absolute file path

**Example:**
```json
{
  "filename": "portfolio",
  "title": "My Portfolio",
  "html": "<header><h1>John Doe</h1></header><main><p>Welcome to my portfolio</p></main>",
  "css": "body { font-family: Arial; margin: 0; } header { background: #333; color: white; padding: 20px; }",
  "javascript": "console.log('Portfolio loaded');"
}
```

### navigate_page
**Purpose:** Navigate browser to URLs or local files.

**Parameters:**
- `url` (required): URL or file path to navigate to

**Returns:** Navigation confirmation with page information

**Example:**
```json
{
  "url": "file:///home/user/portfolio.html"
}
```

### take_screenshot
**Purpose:** Capture visual snapshots of web pages.

**Parameters:**
- `page_id` (optional): Specific page ID to screenshot
- `filename` (optional): Save screenshot to file

**Returns:** Base64 encoded image or file save confirmation

**Example:**
```json
{
  "filename": "portfolio-screenshot.png"
}
```

### take_element_screenshot 🔥 NEW
**Purpose:** Capture screenshots of specific elements with smart positioning and visibility handling.

**Parameters:**
- `selector` (required): CSS selector for the element to screenshot
- `page_id` (optional): Page ID to screenshot from
- `filename` (optional): Filename to save screenshot
- `padding` (optional): Padding around element in pixels (default: 10)
- `scroll_into_view` (optional): Scroll element into view (default: true)
- `wait_for_element` (optional): Wait for element visibility (default: true)
- `timeout` (optional): Maximum wait time in seconds (default: 10)

**Returns:** Base64 encoded image with element details

**Examples:**
```json
{
  "selector": "#submit-button",
  "filename": "button-state.png",
  "padding": 15
}
```

```json
{
  "selector": ".error-message",
  "wait_for_element": true,
  "timeout": 5,
  "padding": 20
}
```

### execute_script
**Purpose:** Execute JavaScript code in browser pages.

**Parameters:**
- `page_id` (optional): Page ID to execute script in
- `script` (required): JavaScript code to execute

**Returns:** Script execution result

**Example:**
```json
{
  "script": "document.querySelectorAll('button').forEach(btn => btn.click());"
}
```

### live_preview
**Purpose:** Start local HTTP server for live development.

**Parameters:**
- `directory` (optional): Directory to serve (default: current)
- `port` (optional): Server port (default: 8080)

**Returns:** Server startup confirmation with URL

**Example:**
```json
{
  "port": 3000,
  "directory": "./dist"
}
```

### set_browser_visibility
**Purpose:** Control browser visibility at runtime.

**Parameters:**
- `visible` (required): Show browser window (true/false)
- `reason` (optional): Reason for visibility change

**Returns:** Visibility change confirmation

**Example:**
```json
{
  "visible": true,
  "reason": "Demo mode for presentation"
}
```

## 🎯 Browser UI Control Tools

### click_element
**Purpose:** Click buttons, links, and interactive elements.

**Parameters:**
- `selector` (required): CSS selector for the element to click
- `page_id` (optional): Page ID to click on
- `timeout` (optional): Timeout in seconds (default: 10)

**Returns:** Click confirmation with element details

**Example:**
```json
{
  "selector": "button.submit-btn",
  "timeout": 15
}
```

### type_text
**Purpose:** Type text into input fields and textareas.

**Parameters:**
- `selector` (required): CSS selector for the input element
- `text` (required): Text to type
- `page_id` (optional): Page ID
- `clear` (optional): Clear field before typing (default: true)

**Returns:** Text input confirmation

**Example:**
```json
{
  "selector": "#email",
  "text": "user@example.com",
  "clear": true
}
```

### wait
**Purpose:** Pause execution for specified time.

**Parameters:**
- `seconds` (required): Number of seconds to wait (0.1-60)

**Returns:** Wait completion confirmation

**Example:**
```json
{
  "seconds": 2.5
}
```

### wait_for_element
**Purpose:** Wait for an element to appear in the DOM.

**Parameters:**
- `selector` (required): CSS selector for element to wait for
- `page_id` (optional): Page ID
- `timeout` (optional): Maximum wait time in seconds (default: 10)

**Returns:** Element appearance confirmation

**Example:**
```json
{
  "selector": ".success-message",
  "timeout": 30
}
```

### get_element_text
**Purpose:** Extract text content from browser elements.

**Parameters:**
- `selector` (required): CSS selector for the element
- `page_id` (optional): Page ID

**Returns:** Element text content

**Example:**
```json
{
  "selector": "h1.page-title"
}
```

### get_element_attribute
**Purpose:** Get attribute values from browser elements.

**Parameters:**
- `selector` (required): CSS selector for the element
- `attribute` (required): Attribute name (e.g., href, src, class)
- `page_id` (optional): Page ID

**Returns:** Attribute value

**Example:**
```json
{
  "selector": "a.download-link",
  "attribute": "href"
}
```

### scroll
**Purpose:** Scroll the page by pixels or to specific elements.

**Parameters:**
- `selector` (optional): CSS selector for element to scroll to
- `x` (optional): Horizontal pixels to scroll (default: 0)
- `y` (optional): Vertical pixels to scroll
- `page_id` (optional): Page ID

**Returns:** Scroll completion confirmation

**Example:**
```json
{
  "selector": "#footer"
}
```
or
```json
{
  "x": 0,
  "y": 500
}
```

### hover_element
**Purpose:** Hover over elements to trigger hover effects.

**Parameters:**
- `selector` (required): CSS selector for element to hover over
- `page_id` (optional): Page ID

**Returns:** Hover completion confirmation

**Example:**
```json
{
  "selector": ".dropdown-menu"
}
```

### keyboard_shortcuts
**Purpose:** Send keyboard combinations and special keys to web pages.

**Parameters:**
- `keys` (required): Key combination to send (e.g., "Ctrl+C", "F5", "Tab", "Enter")
- `selector` (optional): CSS selector to target specific element
- `page_id` (optional): Page ID to send keys to
- `timeout` (optional): Timeout in seconds (default: 10)

**Returns:** Key combination execution confirmation

**Examples:**
```json
{
  "keys": "Ctrl+C"
}
```

```json
{
  "keys": "Tab",
  "selector": "#input-field"
}
```

```json
{
  "keys": "F5"
}
```

**Supported Key Combinations:**
- **Shortcuts:** Ctrl+C, Ctrl+V, Ctrl+A, Ctrl+Z, Ctrl+Y, Ctrl+S, Ctrl+F, Ctrl+R
- **Navigation:** Tab, Shift+Tab, Enter, Escape, Backspace, Delete
- **Arrow Keys:** ArrowUp, ArrowDown, ArrowLeft, ArrowRight
- **Function Keys:** F1-F12 (F5, F11, F12 commonly used)
- **Special Keys:** Space, Home, End, PageUp, PageDown

### switch_tab
**Purpose:** Switch between browser tabs for multi-tab workflow automation.

**Parameters:**
- `action` (optional): Tab action - 'create', 'switch', 'close', 'list', 'close_all' (default: 'switch')
- `target` (optional): Target for action - page_id for switch/close, URL for create, or 'current' for current tab
- `url` (optional): URL to load when creating a new tab
- `switch_to` (optional): Switch method - 'next', 'previous', 'first', 'last', or page_id
- `timeout` (optional): Timeout in seconds for tab operations (default: 10)

**Returns:** Tab operation result with page information

**Examples:**

**Create New Tab:**
```json
{
  "action": "create",
  "url": "https://example.com"
}
```

**Switch to Next Tab:**
```json
{
  "action": "switch",
  "switch_to": "next"
}
```

**Switch to Specific Tab:**
```json
{
  "action": "switch",
  "target": "page_12345"
}
```

**List All Tabs:**
```json
{
  "action": "list"
}
```

**Close Current Tab:**
```json
{
  "action": "close",
  "target": "current"
}
```

**Close All Tabs (except current):**
```json
{
  "action": "close_all"
}
```

## 📁 File System Tools

**🔒 Security:** All file system tools implement path-based access control. By default, access is restricted to the current working directory only. See [Security](#security) section below.

### read_file
**Purpose:** Read contents of files with security validation.

**Parameters:**
- `path` (required): Path to file to read (must be within allowed paths)

**Returns:** File contents with metadata

**Security:** Path must be within allowed directories (working directory by default)

**Example:**
```json
{
  "path": "./config.json"
}
```

### write_file
**Purpose:** Write content to files, creating or overwriting as needed.

**Parameters:**
- `path` (required): Path to file to write (must be within allowed paths)
- `content` (required): Content to write (subject to size limits)
- `create_dirs` (optional): Create parent directories (default: false)

**Returns:** Write confirmation with file details

**Security:** 
- Path must be within allowed directories
- Content size limited to 10MB by default
- Parent directory must also be within allowed paths if `create_dirs` is true

**Example:**
```json
{
  "path": "./dist/index.html",
  "content": "<!DOCTYPE html><html>...</html>",
  "create_dirs": true
}
```

### list_directory
**Purpose:** List directory contents with details and security validation.

**Parameters:**
- `path` (optional): Directory path (default: current directory, must be within allowed paths)
- `show_hidden` (optional): Include hidden files (default: false)

**Returns:** Directory listing with file details

**Security:** Directory path must be within allowed directories

**Example:**
```json
{
  "path": "./src",
  "show_hidden": false
}
```

## 🕷️ Screen Scraping Tools

### screen_scrape
**Purpose:** Extract structured data from web pages using CSS selectors with advanced scraping capabilities.

**Parameters:**
- `selectors` (required): CSS selectors mapping field names to elements
- `url` (optional): URL to scrape (if page_id not provided)
- `page_id` (optional): Existing page ID to scrape from current browser session
- `extract_type` (optional): 'single' or 'multiple' extraction mode (default: 'single')
- `container_selector` (optional): Container selector for multiple items
- `wait_for` (optional): CSS selector to wait for before scraping
- `wait_timeout` (optional): Maximum seconds to wait (default: 10)
- `scroll_to_load` (optional): Auto-scroll to trigger lazy loading (default: false)
- `custom_script` (optional): Custom JavaScript to execute before scraping
- `include_metadata` (optional): Include page metadata (default: true)

**Returns:** Extracted data with metadata

**Example:**
```json
{
  "selectors": {
    "title": "h1",
    "price": ".price-value",
    "description": ".product-description"
  },
  "extract_type": "multiple",
  "container_selector": ".product-card"
}
```

### extract_table 🔥 NEW
**Purpose:** Extract structured data from HTML tables with support for multiple output formats, column filtering, and header management.

**Parameters:**
- `selector` (required): CSS selector for the table element
- `page_id` (optional): Page ID to extract from
- `include_headers` (optional): Include table headers (default: true)
- `output_format` (optional): 'objects', 'array', or 'csv' (default: 'objects')
- `skip_empty_rows` (optional): Skip completely empty rows (default: true)
- `max_rows` (optional): Maximum rows to extract
- `column_filter` (optional): Array of column indices or header names to include
- `header_row` (optional): Row index to use as headers (default: 0)

**Returns:** Structured table data in specified format with metadata

**Examples:**
```json
{
  "selector": "#products-table",
  "output_format": "objects",
  "column_filter": ["Product", "Price", "Stock"]
}
```

```json
{
  "selector": ".data-grid tbody",
  "output_format": "csv",
  "max_rows": 100,
  "skip_empty_rows": true
}
```

## 🌍 Network Tools

### http_request
**Purpose:** Make HTTP requests to URLs.

**Parameters:**
- `url` (required): URL to request
- `method` (optional): HTTP method (default: GET)
- `headers` (optional): HTTP headers as key-value pairs
- `body` (optional): Request body for POST/PUT
- `json` (optional): JSON data to send
- `timeout` (optional): Request timeout in seconds (default: 30)

**Returns:** HTTP response with status, headers, and body

**Example:**
```json
{
  "url": "https://api.example.com/users",
  "method": "POST",
  "json": {
    "name": "John Doe",
    "email": "john@example.com"
  },
  "headers": {
    "Authorization": "Bearer token123"
  }
}
```

## ❓ Help & Discovery Tools

### help
**Purpose:** Get interactive help, usage examples, and workflow suggestions.

**Parameters:**
- `topic` (optional): Help topic - 'overview', 'workflows', 'examples', or specific tool name
- `category` (optional): Tool category - 'browser_automation', 'ui_control', 'file_system', 'network'

**Returns:** Contextual help content with examples and suggestions

**Example:**
```json
{
  "topic": "create_page"
}
```
or
```json
{
  "category": "ui_control"
}
```

## Common Response Format

All tools return responses in this format:

```json
{
  "content": [
    {
      "type": "text",
      "text": "Human-readable result description",
      "data": {
        // Tool-specific structured data
      }
    }
  ],
  "isError": false  // Only present if error occurred
}
```

## Error Handling

When errors occur, tools return:

```json
{
  "content": [
    {
      "type": "text", 
      "text": "Error description"
    }
  ],
  "isError": true
}
```

## Tool Categories

### Browser Automation (7 tools)
Focus on page creation, navigation, and core browser operations.

### UI Control (10 tools)  
Handle user interactions, element manipulation, and page state.

### File System (3 tools)
Manage local files and directories.

### Network (1 tool)
Handle HTTP requests and API communication.

### Help & Discovery (1 tool)
Provide guidance and tool discovery.

## Workflow Patterns

### Basic Web Development
1. `create_page` - Build HTML with CSS/JavaScript
2. `live_preview` - Start local server
3. `navigate_page` - Open in browser
4. `take_screenshot` - Document result

### UI Testing
1. `navigate_page` - Load target page
2. `click_element` / `type_text` - Simulate interactions
3. `wait_for_element` - Handle dynamic content
4. `get_element_text` - Verify results
5. `take_screenshot` - Document test

### API Integration
1. `http_request` - Test API endpoints
2. `create_page` - Build test interface
3. `execute_script` - Make client-side API calls
4. `get_element_text` - Verify data display

## Best Practices

### Error Prevention
- Always use `wait_for_element` before interacting with dynamic content
- Include timeouts for operations that might hang
- Use specific CSS selectors to avoid ambiguity

### Performance
- Use `set_browser_visibility` to switch to headless mode for faster execution
- Batch related operations together
- Use appropriate timeouts based on expected operation duration

### Testing
- Take screenshots before and after changes
- Verify element text/attributes to confirm operations
- Use the help tool to discover optimal usage patterns

## Integration Notes

RodMCP implements the Model Context Protocol (MCP) specification and communicates via JSON-RPC 2.0. All tools are automatically discovered by MCP clients like Claude Desktop and Claude CLI.

## 🔒 Security

### File Access Control

All file system tools (`read_file`, `write_file`, `list_directory`) implement comprehensive security controls:

#### Default Security Configuration
- **Access Restricted**: Current working directory only
- **Path Validation**: Prevents directory traversal attacks  
- **File Size Limits**: 10MB maximum for write operations
- **Symlink Resolution**: Follows symlinks to prevent bypasses

#### Security Error Messages
File access violations return descriptive error messages:
```
file access denied: access denied: path /etc/passwd is not in allowed paths
directory access denied: access denied: path /home/other is in deny list
file size validation failed: file size 52428800 bytes exceeds maximum allowed size 10485760 bytes
```

#### Custom Security Configuration
For advanced use cases, RodMCP supports custom security configurations:

```go
// Example: Allow specific project directories
config := &FileAccessConfig{
    AllowedPaths: []string{
        "/home/user/projects",
        "/tmp/workspace"
    },
    DenyPaths: []string{
        "/home/user/projects/.env",  // Block secrets
        "/home/user/projects/keys"   // Block key files
    },
    MaxFileSize: 50 * 1024 * 1024,  // 50MB limit
}
```

#### Security Features
- **Absolute Path Resolution**: All paths converted to absolute before validation
- **Deny List Priority**: Explicitly blocked paths override allow lists
- **Working Directory Mode**: Restrict all access to working directory subtree
- **Temp Directory Control**: Optional access to system temp directory
- **Comprehensive Logging**: All access attempts logged for audit

See [FILE_ACCESS_SECURITY.md](FILE_ACCESS_SECURITY.md) for complete security documentation.

For development and testing, the comprehensive test suite validates all 26 tools across realistic usage scenarios.