# RodMCP API Reference

Complete reference documentation for all 23 RodMCP tools, organized by category with detailed parameters, examples, and usage patterns.

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

## 📁 File System Tools

### read_file
**Purpose:** Read contents of files.

**Parameters:**
- `path` (required): Path to file to read

**Returns:** File contents with metadata

**Example:**
```json
{
  "path": "./config.json"
}
```

### write_file
**Purpose:** Write content to files, creating or overwriting as needed.

**Parameters:**
- `path` (required): Path to file to write
- `content` (required): Content to write
- `create_dirs` (optional): Create parent directories (default: false)

**Returns:** Write confirmation with file details

**Example:**
```json
{
  "path": "./dist/index.html",
  "content": "<!DOCTYPE html><html>...</html>",
  "create_dirs": true
}
```

### list_directory
**Purpose:** List directory contents with details.

**Parameters:**
- `path` (optional): Directory path (default: current directory)
- `show_hidden` (optional): Include hidden files (default: false)

**Returns:** Directory listing with file details

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

### Browser Automation (6 tools)
Focus on page creation, navigation, and core browser operations.

### UI Control (8 tools)  
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

For development and testing, the comprehensive test suite validates all 23 tools across realistic usage scenarios.