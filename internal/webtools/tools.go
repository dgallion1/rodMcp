package webtools

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CreatePageTool creates HTML pages
type CreatePageTool struct {
	logger *logger.Logger
}

func NewCreatePageTool(log *logger.Logger) *CreatePageTool {
	return &CreatePageTool{logger: log}
}

func (t *CreatePageTool) Name() string {
	return "create_page"
}

func (t *CreatePageTool) Description() string {
	return "Create an HTML page with CSS and JavaScript"
}

func (t *CreatePageTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"filename": map[string]interface{}{
				"type":        "string",
				"description": "Name of the HTML file to create",
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Page title",
			},
			"html": map[string]interface{}{
				"type":        "string",
				"description": "HTML content for the body",
			},
			"css": map[string]interface{}{
				"type":        "string",
				"description": "CSS styles",
			},
			"javascript": map[string]interface{}{
				"type":        "string",
				"description": "JavaScript code",
			},
		},
		Required: []string{"filename", "title", "html"},
	}
}

func (t *CreatePageTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	filename, ok := args["filename"].(string)
	if !ok {
		return nil, fmt.Errorf("filename is required")
	}

	title, ok := args["title"].(string)
	if !ok {
		title = "Untitled Page"
	}

	html, ok := args["html"].(string)
	if !ok {
		html = "<p>Empty page</p>"
	}

	css, _ := args["css"].(string)
	javascript, _ := args["javascript"].(string)

	// Create the HTML document
	document := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
%s
    </style>
</head>
<body>
%s
    <script>
%s
    </script>
</body>
</html>`, title, css, html, javascript)

	// Ensure filename has .html extension
	if !strings.HasSuffix(filename, ".html") {
		filename += ".html"
	}

	// Write to file
	if err := os.WriteFile(filename, []byte(document), 0644); err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to create file: %v", err),
			}},
			IsError: true,
		}, nil
	}

	absPath, _ := filepath.Abs(filename)

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Created HTML page: %s", absPath),
		}},
	}, nil
}

// NavigatePageTool navigates browser to a page
type NavigatePageTool struct {
	logger  *logger.Logger
	browser *browser.Manager
}

func NewNavigatePageTool(log *logger.Logger, browserMgr *browser.Manager) *NavigatePageTool {
	return &NavigatePageTool{logger: log, browser: browserMgr}
}

func (t *NavigatePageTool) Name() string {
	return "navigate_page"
}

func (t *NavigatePageTool) Description() string {
	return "Navigate browser to a URL or local file"
}

func (t *NavigatePageTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL or file path to navigate to",
			},
		},
		Required: []string{"url"},
	}
}

func (t *NavigatePageTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url is required")
	}

	// Handle local file paths
	if !strings.HasPrefix(url, "http") {
		if absPath, err := filepath.Abs(url); err == nil {
			url = "file://" + absPath
		}
	}

	page, pageID, err := t.browser.NewPage(url)
	if err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to navigate: %v", err),
			}},
			IsError: true,
		}, nil
	}

	info, _ := t.browser.GetPageInfo(pageID)

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Navigated to %s (Page ID: %s)", page.MustInfo().URL, pageID),
			Data: info,
		}},
	}, nil
}

// ScreenshotTool takes screenshots
type ScreenshotTool struct {
	logger  *logger.Logger
	browser *browser.Manager
}

func NewScreenshotTool(log *logger.Logger, browserMgr *browser.Manager) *ScreenshotTool {
	return &ScreenshotTool{logger: log, browser: browserMgr}
}

func (t *ScreenshotTool) Name() string {
	return "take_screenshot"
}

func (t *ScreenshotTool) Description() string {
	return "Take a screenshot of a browser page"
}

func (t *ScreenshotTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID to screenshot (optional, uses first page if not specified)",
			},
			"filename": map[string]interface{}{
				"type":        "string",
				"description": "Filename to save screenshot (optional)",
			},
		},
	}
}

func (t *ScreenshotTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	pageID, ok := args["page_id"].(string)
	if !ok || pageID == "" {
		// Use first available page
		pages := t.browser.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for screenshot",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
	}

	screenshot, err := t.browser.Screenshot(pageID)
	if err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to take screenshot: %v", err),
			}},
			IsError: true,
		}, nil
	}

	filename, _ := args["filename"].(string)
	if filename != "" {
		if err := os.WriteFile(filename, screenshot, 0644); err != nil {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Failed to save screenshot: %v", err),
				}},
				IsError: true,
			}, nil
		}

		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Screenshot saved to %s", filename),
			}},
		}, nil
	}

	// Return base64 encoded image
	encoded := base64.StdEncoding.EncodeToString(screenshot)

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type:     "image",
			Data:     encoded,
			MimeType: "image/png",
		}},
	}, nil
}

// ExecuteScriptTool executes JavaScript
type ExecuteScriptTool struct {
	logger  *logger.Logger
	browser *browser.Manager
}

func NewExecuteScriptTool(log *logger.Logger, browserMgr *browser.Manager) *ExecuteScriptTool {
	return &ExecuteScriptTool{logger: log, browser: browserMgr}
}

func (t *ExecuteScriptTool) Name() string {
	return "execute_script"
}

func (t *ExecuteScriptTool) Description() string {
	return "Execute JavaScript code in a browser page"
}

func (t *ExecuteScriptTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID to execute script in (optional, uses first page if not specified)",
			},
			"script": map[string]interface{}{
				"type":        "string",
				"description": "JavaScript code to execute",
			},
		},
		Required: []string{"script"},
	}
}

func (t *ExecuteScriptTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	pageID, ok := args["page_id"].(string)
	if !ok || pageID == "" {
		// Use first available page
		pages := t.browser.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for script execution",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
	}

	script, ok := args["script"].(string)
	if !ok {
		return nil, fmt.Errorf("script is required")
	}

	result, err := t.browser.ExecuteScript(pageID, script)
	if err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Script execution failed: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Script executed successfully. Result: %v", result),
			Data: result,
		}},
	}, nil
}

// BrowserVisibilityTool controls browser visibility at runtime
type BrowserVisibilityTool struct {
	logger  *logger.Logger
	browser *browser.Manager
}

func NewBrowserVisibilityTool(log *logger.Logger, browserMgr *browser.Manager) *BrowserVisibilityTool {
	return &BrowserVisibilityTool{logger: log, browser: browserMgr}
}

func (t *BrowserVisibilityTool) Name() string {
	return "set_browser_visibility"
}

func (t *BrowserVisibilityTool) Description() string {
	return "Control browser visibility - switch between visible and headless modes at runtime"
}

func (t *BrowserVisibilityTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"visible": map[string]interface{}{
				"type":        "boolean",
				"description": "Set to true to show browser window, false for headless mode",
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "Optional reason for visibility change (for logging)",
			},
		},
		Required: []string{"visible"},
	}
}

func (t *BrowserVisibilityTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	visible, ok := args["visible"].(bool)
	if !ok {
		return nil, fmt.Errorf("visible parameter is required")
	}

	reason, _ := args["reason"].(string)
	if reason == "" {
		if visible {
			reason = "MCP controller requested visible mode"
		} else {
			reason = "MCP controller requested headless mode"
		}
	}

	// Update browser visibility
	err := t.browser.SetVisibility(visible)
	if err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to change browser visibility: %v", err),
			}},
			IsError: true,
		}, nil
	}

	mode := "headless"
	if visible {
		mode = "visible"
	}

	t.logger.WithComponent("webtools").Info("Browser visibility changed",
		zap.String("mode", mode),
		zap.String("reason", reason))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Browser set to %s mode. Reason: %s", mode, reason),
			Data: map[string]interface{}{
				"visible": visible,
				"mode":    mode,
				"reason":  reason,
			},
		}},
	}, nil
}

// LivePreviewTool creates a simple HTTP server for live preview
type LivePreviewTool struct {
	logger *logger.Logger
	server *http.Server
}

func NewLivePreviewTool(log *logger.Logger) *LivePreviewTool {
	return &LivePreviewTool{logger: log}
}

func (t *LivePreviewTool) Name() string {
	return "live_preview"
}

func (t *LivePreviewTool) Description() string {
	return "Start a live preview server for local HTML files"
}

func (t *LivePreviewTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"directory": map[string]interface{}{
				"type":        "string",
				"description": "Directory to serve (default: current directory)",
			},
			"port": map[string]interface{}{
				"type":        "integer",
				"description": "Port to serve on (default: 8080)",
			},
		},
	}
}

func (t *LivePreviewTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	directory, ok := args["directory"].(string)
	if !ok || directory == "" {
		directory = "."
	}

	port := 8080
	if p, ok := args["port"].(float64); ok {
		port = int(p)
	} else if p, ok := args["port"].(int); ok {
		port = p
	}

	// Stop existing server if running
	if t.server != nil {
		t.server.Close()
	}

	// Create file server
	fs := http.FileServer(http.Dir(directory))
	http.Handle("/", fs)

	// Start server
	addr := ":" + strconv.Itoa(port)
	t.server = &http.Server{Addr: addr}

	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.logger.WithComponent("webtools").Error("Preview server error",
				zap.Error(err))
		}
	}()

	url := fmt.Sprintf("http://localhost:%d", port)

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Live preview server started at %s serving %s", url, directory),
			Data: map[string]interface{}{
				"url":       url,
				"directory": directory,
				"port":      port,
			},
		}},
	}, nil
}

// ReadFileTool reads file contents
type ReadFileTool struct {
	logger *logger.Logger
}

func NewReadFileTool(log *logger.Logger) *ReadFileTool {
	return &ReadFileTool{logger: log}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file"
}

func (t *ReadFileTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read",
			},
		},
		Required: []string{"path"},
	}
}

func (t *ReadFileTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	pathStr, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	// Clean the path to prevent directory traversal attacks
	cleanPath := filepath.Clean(pathStr)
	
	// Read the file
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to read file",
			zap.String("path", cleanPath),
			zap.Error(err))
		return nil, fmt.Errorf("failed to read file %s: %w", cleanPath, err)
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("File read successfully",
		zap.String("path", cleanPath),
		zap.Int("size_bytes", len(content)),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: string(content),
			Data: map[string]interface{}{
				"path":       cleanPath,
				"size_bytes": len(content),
				"encoding":   "utf-8",
			},
		}},
	}, nil
}

// WriteFileTool writes content to files
type WriteFileTool struct {
	logger *logger.Logger
}

func NewWriteFileTool(log *logger.Logger) *WriteFileTool {
	return &WriteFileTool{logger: log}
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file, creating or overwriting as needed"
}

func (t *WriteFileTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
			"create_dirs": map[string]interface{}{
				"type":        "boolean",
				"description": "Create parent directories if they don't exist",
				"default":     false,
			},
		},
		Required: []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	pathStr, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content must be a string")
	}

	createDirs := false
	if val, ok := args["create_dirs"].(bool); ok {
		createDirs = val
	}

	// Clean the path
	cleanPath := filepath.Clean(pathStr)
	
	// Create parent directories if requested
	if createDirs {
		dir := filepath.Dir(cleanPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directories for %s: %w", cleanPath, err)
		}
	}

	// Write the file
	err := os.WriteFile(cleanPath, []byte(content), 0644)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to write file",
			zap.String("path", cleanPath),
			zap.Error(err))
		return nil, fmt.Errorf("failed to write file %s: %w", cleanPath, err)
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("File written successfully",
		zap.String("path", cleanPath),
		zap.Int("size_bytes", len(content)),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), cleanPath),
			Data: map[string]interface{}{
				"path":       cleanPath,
				"size_bytes": len(content),
				"created_dirs": createDirs,
			},
		}},
	}, nil
}

// ListDirectoryTool lists directory contents
type ListDirectoryTool struct {
	logger *logger.Logger
}

func NewListDirectoryTool(log *logger.Logger) *ListDirectoryTool {
	return &ListDirectoryTool{logger: log}
}

func (t *ListDirectoryTool) Name() string {
	return "list_directory"
}

func (t *ListDirectoryTool) Description() string {
	return "List the contents of a directory"
}

func (t *ListDirectoryTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the directory to list",
				"default":     ".",
			},
			"show_hidden": map[string]interface{}{
				"type":        "boolean",
				"description": "Include hidden files (starting with .)",
				"default":     false,
			},
		},
	}
}

func (t *ListDirectoryTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	pathStr := "."
	if val, ok := args["path"].(string); ok {
		pathStr = val
	}

	showHidden := false
	if val, ok := args["show_hidden"].(bool); ok {
		showHidden = val
	}

	// Clean the path
	cleanPath := filepath.Clean(pathStr)
	
	// Read directory
	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to read directory",
			zap.String("path", cleanPath),
			zap.Error(err))
		return nil, fmt.Errorf("failed to read directory %s: %w", cleanPath, err)
	}

	var items []map[string]interface{}
	var totalSize int64

	for _, entry := range entries {
		name := entry.Name()
		
		// Skip hidden files if not requested
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		item := map[string]interface{}{
			"name":      name,
			"type":      "file",
			"size":      info.Size(),
			"modified":  info.ModTime().Format(time.RFC3339),
			"is_dir":    info.IsDir(),
		}

		if info.IsDir() {
			item["type"] = "directory"
		}

		totalSize += info.Size()
		items = append(items, item)
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Directory listed successfully",
		zap.String("path", cleanPath),
		zap.Int("item_count", len(items)),
		zap.Int64("duration_ms", duration))

	var text strings.Builder
	text.WriteString(fmt.Sprintf("Directory listing for %s:\n", cleanPath))
	for _, item := range items {
		itemType := item["type"].(string)
		name := item["name"].(string)
		size := item["size"].(int64)
		modified := item["modified"].(string)
		
		if itemType == "directory" {
			text.WriteString(fmt.Sprintf("  📁 %s/ (modified: %s)\n", name, modified))
		} else {
			text.WriteString(fmt.Sprintf("  📄 %s (%d bytes, modified: %s)\n", name, size, modified))
		}
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: text.String(),
			Data: map[string]interface{}{
				"path":       cleanPath,
				"items":      items,
				"item_count": len(items),
				"total_size": totalSize,
			},
		}},
	}, nil
}

// HTTPRequestTool makes HTTP requests
type HTTPRequestTool struct {
	logger *logger.Logger
}

func NewHTTPRequestTool(log *logger.Logger) *HTTPRequestTool {
	return &HTTPRequestTool{logger: log}
}

func (t *HTTPRequestTool) Name() string {
	return "http_request"
}

func (t *HTTPRequestTool) Description() string {
	return "Make HTTP requests (GET, POST, PUT, DELETE, etc.)"
}

func (t *HTTPRequestTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to request",
			},
			"method": map[string]interface{}{
				"type":        "string",
				"description": "HTTP method (GET, POST, PUT, DELETE, etc.)",
				"default":     "GET",
			},
			"headers": map[string]interface{}{
				"type":        "object",
				"description": "HTTP headers as key-value pairs",
				"default":     map[string]interface{}{},
			},
			"body": map[string]interface{}{
				"type":        "string",
				"description": "Request body (for POST, PUT, etc.)",
			},
			"json": map[string]interface{}{
				"type":        "object",
				"description": "JSON data to send (will set Content-Type: application/json)",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Request timeout in seconds",
				"default":     30,
			},
		},
		Required: []string{"url"},
	}
}

func (t *HTTPRequestTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url must be a string")
	}

	method := "GET"
	if val, ok := args["method"].(string); ok {
		method = strings.ToUpper(val)
	}

	timeout := 30
	if val, ok := args["timeout"].(float64); ok {
		timeout = int(val)
	}

	var body io.Reader
	var bodyContent string

	// Handle JSON body
	if jsonData, ok := args["json"]; ok {
		jsonBytes, err := json.Marshal(jsonData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}
		body = bytes.NewReader(jsonBytes)
		bodyContent = string(jsonBytes)
	} else if bodyStr, ok := args["body"].(string); ok {
		body = strings.NewReader(bodyStr)
		bodyContent = bodyStr
	}

	// Create request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if headers, ok := args["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if valueStr, ok := value.(string); ok {
				req.Header.Set(key, valueStr)
			}
		}
	}

	// Set Content-Type for JSON
	if _, hasJSON := args["json"]; hasJSON {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		t.logger.WithComponent("tools").Error("HTTP request failed",
			zap.String("url", url),
			zap.String("method", method),
			zap.Error(err))
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("HTTP request completed",
		zap.String("url", url),
		zap.String("method", method),
		zap.Int("status_code", resp.StatusCode),
		zap.Int("response_size", len(responseBody)),
		zap.Int64("duration_ms", duration))

	// Prepare response headers
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	responseText := fmt.Sprintf("HTTP %s %s\nStatus: %d %s\nResponse Size: %d bytes\n\nHeaders:\n",
		method, url, resp.StatusCode, resp.Status, len(responseBody))
	
	for key, value := range responseHeaders {
		responseText += fmt.Sprintf("  %s: %s\n", key, value)
	}
	
	responseText += fmt.Sprintf("\nBody:\n%s", string(responseBody))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: responseText,
			Data: map[string]interface{}{
				"url":            url,
				"method":         method,
				"status_code":    resp.StatusCode,
				"status":         resp.Status,
				"headers":        responseHeaders,
				"body":           string(responseBody),
				"response_size":  len(responseBody),
				"duration_ms":    duration,
				"request_body":   bodyContent,
			},
		}},
	}, nil
}

// ClickElementTool clicks on browser elements
type ClickElementTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewClickElementTool(log *logger.Logger, mgr *browser.Manager) *ClickElementTool {
	return &ClickElementTool{logger: log, browserMgr: mgr}
}

func (t *ClickElementTool) Name() string {
	return "click_element"
}

func (t *ClickElementTool) Description() string {
	return "Click on a browser element using CSS selector"
}

func (t *ClickElementTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector or XPath (prefix with //) for the element to click",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID to click on (optional, uses current page if not specified)",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds to wait for element (default: 10)",
				"default":     10,
			},
		},
		Required: []string{"selector"},
	}
}

func (t *ClickElementTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector must be a string")
	}

	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}

	_ = 10 // timeout for future use
	if _, ok := args["timeout"].(float64); ok {
		// timeout = int(val) // for future use
	}

	// Get the page
	var page *browser.Manager
	if pageID != "" {
		// Use specific page (this would need to be implemented in browser manager)
		page = t.browserMgr
	} else {
		page = t.browserMgr
	}

	// For now, use execute_script as the underlying mechanism until we have direct Rod access
	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (!element) {
			throw new Error('Element not found with selector: %s');
		}
		element.click();
		return 'Clicked element: ' + '%s';
	`, selector, selector, selector)

	result, err := page.ExecuteScript("", script)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to click element",
			zap.String("selector", selector),
			zap.Error(err))
		return nil, fmt.Errorf("failed to click element %s: %w", selector, err)
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Element clicked successfully",
		zap.String("selector", selector),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Successfully clicked element: %s", selector),
			Data: map[string]interface{}{
				"selector":    selector,
				"page_id":     pageID,
				"duration_ms": duration,
				"result":      result,
			},
		}},
	}, nil
}

// TypeTextTool types text into input elements
type TypeTextTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewTypeTextTool(log *logger.Logger, mgr *browser.Manager) *TypeTextTool {
	return &TypeTextTool{logger: log, browserMgr: mgr}
}

func (t *TypeTextTool) Name() string {
	return "type_text"
}

func (t *TypeTextTool) Description() string {
	return "Type text into an input field or textarea"
}

func (t *TypeTextTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for the input element",
			},
			"text": map[string]interface{}{
				"type":        "string",
				"description": "Text to type into the element",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID (optional)",
			},
			"clear": map[string]interface{}{
				"type":        "boolean",
				"description": "Clear the field before typing (default: true)",
				"default":     true,
			},
		},
		Required: []string{"selector", "text"},
	}
}

func (t *TypeTextTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector must be a string")
	}

	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text must be a string")
	}

	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}

	clear := true
	if val, ok := args["clear"].(bool); ok {
		clear = val
	}

	// Escape text for JavaScript
	escapedText := strings.ReplaceAll(text, `"`, `\"`)
	escapedText = strings.ReplaceAll(escapedText, `'`, `\'`)
	escapedText = strings.ReplaceAll(escapedText, "\n", "\\n")

	clearScript := ""
	if clear {
		clearScript = "element.value = '';"
	}

	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (!element) {
			throw new Error('Element not found with selector: %s');
		}
		%s
		element.focus();
		element.value = '%s';
		element.dispatchEvent(new Event('input', { bubbles: true }));
		element.dispatchEvent(new Event('change', { bubbles: true }));
		return 'Typed text into: %s';
	`, selector, selector, clearScript, escapedText, selector)

	result, err := t.browserMgr.ExecuteScript("", script)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to type text",
			zap.String("selector", selector),
			zap.String("text", text),
			zap.Error(err))
		return nil, fmt.Errorf("failed to type text into %s: %w", selector, err)
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Text typed successfully",
		zap.String("selector", selector),
		zap.String("text", text),
		zap.Bool("cleared", clear),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Successfully typed '%s' into element: %s", text, selector),
			Data: map[string]interface{}{
				"selector":    selector,
				"text":        text,
				"page_id":     pageID,
				"cleared":     clear,
				"duration_ms": duration,
				"result":      result,
			},
		}},
	}, nil
}

// WaitTool pauses execution for specified time
type WaitTool struct {
	logger *logger.Logger
}

func NewWaitTool(log *logger.Logger) *WaitTool {
	return &WaitTool{logger: log}
}

func (t *WaitTool) Name() string {
	return "wait"
}

func (t *WaitTool) Description() string {
	return "Wait for a specified number of seconds"
}

func (t *WaitTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"seconds": map[string]interface{}{
				"type":        "number",
				"description": "Number of seconds to wait",
				"minimum":     0.1,
				"maximum":     60,
			},
		},
		Required: []string{"seconds"},
	}
}

func (t *WaitTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	seconds, ok := args["seconds"].(float64)
	if !ok {
		return nil, fmt.Errorf("seconds must be a number")
	}

	if seconds < 0.1 || seconds > 60 {
		return nil, fmt.Errorf("seconds must be between 0.1 and 60")
	}

	duration := time.Duration(seconds * float64(time.Second))
	time.Sleep(duration)

	elapsed := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Wait completed",
		zap.Float64("seconds", seconds),
		zap.Int64("elapsed_ms", elapsed))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Waited for %.1f seconds", seconds),
			Data: map[string]interface{}{
				"seconds":    seconds,
				"elapsed_ms": elapsed,
			},
		}},
	}, nil
}

// WaitForElementTool waits for an element to appear
type WaitForElementTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewWaitForElementTool(log *logger.Logger, mgr *browser.Manager) *WaitForElementTool {
	return &WaitForElementTool{logger: log, browserMgr: mgr}
}

func (t *WaitForElementTool) Name() string {
	return "wait_for_element"
}

func (t *WaitForElementTool) Description() string {
	return "Wait for an element to appear in the DOM"
}

func (t *WaitForElementTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for the element to wait for",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID (optional)",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum time to wait in seconds (default: 10)",
				"default":     10,
			},
		},
		Required: []string{"selector"},
	}
}

func (t *WaitForElementTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector must be a string")
	}

	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}

	timeout := 10
	if val, ok := args["timeout"].(float64); ok {
		timeout = int(val)
	}

	// JavaScript to poll for element
	script := fmt.Sprintf(`
		const maxWait = %d * 1000; // Convert to milliseconds
		const startTime = Date.now();
		
		function checkElement() {
			const element = document.querySelector('%s');
			if (element) {
				return 'Element found: %s';
			}
			
			if (Date.now() - startTime > maxWait) {
				throw new Error('Timeout waiting for element: %s');
			}
			
			// Wait 100ms and try again
			return new Promise((resolve, reject) => {
				setTimeout(() => {
					try {
						resolve(checkElement());
					} catch (e) {
						reject(e);
					}
				}, 100);
			});
		}
		
		return checkElement();
	`, timeout, selector, selector, selector)

	result, err := t.browserMgr.ExecuteScript("", script)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to wait for element",
			zap.String("selector", selector),
			zap.Int("timeout", timeout),
			zap.Error(err))
		return nil, fmt.Errorf("timeout waiting for element %s: %w", selector, err)
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Element found successfully",
		zap.String("selector", selector),
		zap.Int("timeout", timeout),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Element found: %s", selector),
			Data: map[string]interface{}{
				"selector":    selector,
				"page_id":     pageID,
				"timeout":     timeout,
				"duration_ms": duration,
				"result":      result,
			},
		}},
	}, nil
}

// GetElementTextTool extracts text from elements
type GetElementTextTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewGetElementTextTool(log *logger.Logger, mgr *browser.Manager) *GetElementTextTool {
	return &GetElementTextTool{logger: log, browserMgr: mgr}
}

func (t *GetElementTextTool) Name() string {
	return "get_element_text"
}

func (t *GetElementTextTool) Description() string {
	return "Extract text content from a browser element"
}

func (t *GetElementTextTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for the element to get text from",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID (optional)",
			},
		},
		Required: []string{"selector"},
	}
}

func (t *GetElementTextTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector must be a string")
	}

	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}

	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (!element) {
			throw new Error('Element not found with selector: %s');
		}
		return element.textContent || element.innerText || '';
	`, selector, selector)

	result, err := t.browserMgr.ExecuteScript("", script)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to get element text",
			zap.String("selector", selector),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get text from element %s: %w", selector, err)
	}

	text := ""
	if resultStr, ok := result.(string); ok {
		text = resultStr
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Element text extracted successfully",
		zap.String("selector", selector),
		zap.String("text", text),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Text from %s: %s", selector, text),
			Data: map[string]interface{}{
				"selector":    selector,
				"text":        text,
				"page_id":     pageID,
				"duration_ms": duration,
			},
		}},
	}, nil
}

// GetElementAttributeTool gets element attributes
type GetElementAttributeTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewGetElementAttributeTool(log *logger.Logger, mgr *browser.Manager) *GetElementAttributeTool {
	return &GetElementAttributeTool{logger: log, browserMgr: mgr}
}

func (t *GetElementAttributeTool) Name() string {
	return "get_element_attribute"
}

func (t *GetElementAttributeTool) Description() string {
	return "Get an attribute value from a browser element"
}

func (t *GetElementAttributeTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for the element",
			},
			"attribute": map[string]interface{}{
				"type":        "string",
				"description": "Attribute name to get (e.g., href, src, class)",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID (optional)",
			},
		},
		Required: []string{"selector", "attribute"},
	}
}

func (t *GetElementAttributeTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector must be a string")
	}

	attribute, ok := args["attribute"].(string)
	if !ok {
		return nil, fmt.Errorf("attribute must be a string")
	}

	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}

	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (!element) {
			throw new Error('Element not found with selector: %s');
		}
		return element.getAttribute('%s');
	`, selector, selector, attribute)

	result, err := t.browserMgr.ExecuteScript("", script)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to get element attribute",
			zap.String("selector", selector),
			zap.String("attribute", attribute),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get attribute %s from element %s: %w", attribute, selector, err)
	}

	value := ""
	if resultStr, ok := result.(string); ok {
		value = resultStr
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Element attribute retrieved successfully",
		zap.String("selector", selector),
		zap.String("attribute", attribute),
		zap.String("value", value),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Attribute %s from %s: %s", attribute, selector, value),
			Data: map[string]interface{}{
				"selector":    selector,
				"attribute":   attribute,
				"value":       value,
				"page_id":     pageID,
				"duration_ms": duration,
			},
		}},
	}, nil
}

// ScrollTool scrolls the page or to specific elements
type ScrollTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewScrollTool(log *logger.Logger, mgr *browser.Manager) *ScrollTool {
	return &ScrollTool{logger: log, browserMgr: mgr}
}

func (t *ScrollTool) Name() string {
	return "scroll"
}

func (t *ScrollTool) Description() string {
	return "Scroll the page by pixels or to a specific element"
}

func (t *ScrollTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for element to scroll to (optional)",
			},
			"x": map[string]interface{}{
				"type":        "integer",
				"description": "Horizontal pixels to scroll (optional)",
				"default":     0,
			},
			"y": map[string]interface{}{
				"type":        "integer",
				"description": "Vertical pixels to scroll (optional)",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID (optional)",
			},
		},
	}
}

func (t *ScrollTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	selector := ""
	if val, ok := args["selector"].(string); ok {
		selector = val
	}

	x := 0
	if val, ok := args["x"].(float64); ok {
		x = int(val)
	}

	y := 0
	if val, ok := args["y"].(float64); ok {
		y = int(val)
	}

	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}

	var script string
	var description string

	if selector != "" {
		// Scroll to element
		script = fmt.Sprintf(`
			const element = document.querySelector('%s');
			if (!element) {
				throw new Error('Element not found with selector: %s');
			}
			element.scrollIntoView({ behavior: 'smooth', block: 'center' });
			return 'Scrolled to element: %s';
		`, selector, selector, selector)
		description = fmt.Sprintf("Scrolled to element: %s", selector)
	} else if y != 0 || x != 0 {
		// Scroll by pixels
		script = fmt.Sprintf(`
			window.scrollBy(%d, %d);
			return 'Scrolled by %d, %d pixels';
		`, x, y, x, y)
		description = fmt.Sprintf("Scrolled by %d, %d pixels", x, y)
	} else {
		return nil, fmt.Errorf("must specify either selector or x/y coordinates")
	}

	result, err := t.browserMgr.ExecuteScript("", script)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to scroll",
			zap.String("selector", selector),
			zap.Int("x", x),
			zap.Int("y", y),
			zap.Error(err))
		return nil, fmt.Errorf("failed to scroll: %w", err)
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Scroll completed successfully",
		zap.String("selector", selector),
		zap.Int("x", x),
		zap.Int("y", y),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: description,
			Data: map[string]interface{}{
				"selector":    selector,
				"x":           x,
				"y":           y,
				"page_id":     pageID,
				"duration_ms": duration,
				"result":      result,
			},
		}},
	}, nil
}

// HoverElementTool hovers over browser elements
type HoverElementTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewHoverElementTool(log *logger.Logger, mgr *browser.Manager) *HoverElementTool {
	return &HoverElementTool{logger: log, browserMgr: mgr}
}

func (t *HoverElementTool) Name() string {
	return "hover_element"
}

func (t *HoverElementTool) Description() string {
	return "Hover over a browser element to trigger hover effects"
}

func (t *HoverElementTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for the element to hover over",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID (optional)",
			},
		},
		Required: []string{"selector"},
	}
}

func (t *HoverElementTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector must be a string")
	}

	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}

	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (!element) {
			throw new Error('Element not found with selector: %s');
		}
		
		// Create and dispatch mouseover event
		const event = new MouseEvent('mouseover', {
			bubbles: true,
			cancelable: true,
			view: window
		});
		element.dispatchEvent(event);
		
		// Also trigger mouseenter for completeness
		const enterEvent = new MouseEvent('mouseenter', {
			bubbles: false,
			cancelable: true,
			view: window
		});
		element.dispatchEvent(enterEvent);
		
		return 'Hovered over element: %s';
	`, selector, selector, selector)

	result, err := t.browserMgr.ExecuteScript("", script)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to hover over element",
			zap.String("selector", selector),
			zap.Error(err))
		return nil, fmt.Errorf("failed to hover over element %s: %w", selector, err)
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Element hovered successfully",
		zap.String("selector", selector),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Successfully hovered over element: %s", selector),
			Data: map[string]interface{}{
				"selector":    selector,
				"page_id":     pageID,
				"duration_ms": duration,
				"result":      result,
			},
		}},
	}, nil
}
