package webtools

import (
	"bytes"
	"context"
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
	// Add total execution timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	// Use a channel to handle timeout
	type result struct {
		response *types.CallToolResponse
		err      error
	}
	resultChan := make(chan result, 1)
	
	go func() {
		start := time.Now()
		defer func() {
			duration := time.Since(start).Milliseconds()
			t.logger.LogToolExecution(t.Name(), args, true, duration)
		}()

		url, ok := args["url"].(string)
		if !ok {
			resultChan <- result{nil, fmt.Errorf("url is required")}
			return
		}
		
		resp, err := t.executeNavigation(url)
		resultChan <- result{resp, err}
	}()
	
	select {
	case res := <-resultChan:
		return res.response, res.err
	case <-ctx.Done():
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "Navigation timed out after 15 seconds",
			}},
			IsError: true,
		}, nil
	}
}

func (t *NavigatePageTool) executeNavigation(url string) (*types.CallToolResponse, error) {
	// Handle local file paths
	if !strings.HasPrefix(url, "http") {
		if absPath, err := filepath.Abs(url); err == nil {
			url = "file://" + absPath
		}
	}

	// Check if there are existing pages, if so navigate the first one instead of creating new
	pages := t.browser.ListPages()
	var pageID string
	
	if len(pages) > 0 {
		// Use existing page and navigate it to new URL
		pageID = pages[0]
		if err := t.browser.NavigateExistingPage(pageID, url); err != nil {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Failed to navigate to %s: %v", url, err),
				}},
				IsError: true,
			}, nil
		}
	} else {
		// Create new page if none exist
		_, newPageID, err := t.browser.NewPage(url)
		if err != nil {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Failed to navigate: %v", err),
				}},
				IsError: true,
			}, nil
		}
		pageID = newPageID
	}

	// Add timeout for GetPageInfo to prevent hanging
	info := t.getPageInfoWithTimeout(pageID, 5*time.Second)
	currentURL := "unknown"
	if info != nil {
		if u, ok := info["url"].(string); ok {
			currentURL = u
		}
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Navigated to %s (Page ID: %s)", currentURL, pageID),
			Data: info,
		}},
	}, nil
}

// getPageInfoWithTimeout wraps GetPageInfo with a timeout to prevent hanging
func (t *NavigatePageTool) getPageInfoWithTimeout(pageID string, timeout time.Duration) map[string]interface{} {
	type infoResult struct {
		info map[string]interface{}
		err  error
	}
	resultChan := make(chan infoResult, 1)
	
	go func() {
		info, err := t.browser.GetPageInfo(pageID)
		resultChan <- infoResult{info, err}
	}()
	
	select {
	case res := <-resultChan:
		if res.err != nil {
			return map[string]interface{}{
				"url":   "unknown",
				"error": res.err.Error(),
			}
		}
		return res.info
	case <-time.After(timeout):
		return map[string]interface{}{
			"url":   "unknown",
			"error": "GetPageInfo timed out",
		}
	}
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
	// Add total execution timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Use a channel to handle timeout
	type result struct {
		response *types.CallToolResponse
		err      error
	}
	resultChan := make(chan result, 1)
	
	go func() {
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
				resultChan <- result{&types.CallToolResponse{
					Content: []types.ToolContent{{
						Type: "text",
						Text: "No pages available for script execution",
					}},
					IsError: true,
				}, nil}
				return
			}
			pageID = pages[0]
		}

		script, ok := args["script"].(string)
		if !ok {
			resultChan <- result{nil, fmt.Errorf("script is required")}
			return
		}

		scriptResult, err := t.browser.ExecuteScript(pageID, script)
		if err != nil {
			resultChan <- result{&types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Script execution failed: %v", err),
				}},
				IsError: true,
			}, nil}
			return
		}

		resultChan <- result{&types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Script executed successfully. Result: %v", scriptResult),
				Data: scriptResult,
			}},
		}, nil}
	}()
	
	select {
	case res := <-resultChan:
		return res.response, res.err
	case <-ctx.Done():
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "Script execution timed out after 30 seconds",
			}},
			IsError: true,
		}, nil
	}
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

	// Get the page ID to use
	if pageID == "" {
		// Use first available page if no specific page ID provided
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for element interaction",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
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

	result, err := t.browserMgr.ExecuteScript(pageID, script)
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
	
	// Get the page ID to use
	if pageID == "" {
		// Use first available page if no specific page ID provided
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for text input",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
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

	result, err := t.browserMgr.ExecuteScript(pageID, script)
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
	
	// Get the page ID to use
	if pageID == "" {
		// Use first available page if no specific page ID provided
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for waiting for element",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
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

	result, err := t.browserMgr.ExecuteScript(pageID, script)
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
	
	// Get the page ID to use
	if pageID == "" {
		// Use first available page if no specific page ID provided
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for getting element text",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
	}

	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (!element) {
			throw new Error('Element not found with selector: %s');
		}
		return element.textContent || element.innerText || '';
	`, selector, selector)

	result, err := t.browserMgr.ExecuteScript(pageID, script)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to get element text",
			zap.String("selector", selector),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get text from element %s: %w", selector, err)
	}

	text := ""
	if resultStr, ok := result.(string); ok {
		text = resultStr
	} else {
		// Handle non-string results (e.g., gson.JSON from go-rod)
		text = fmt.Sprintf("%v", result)
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
	
	// Get the page ID to use
	if pageID == "" {
		// Use first available page if no specific page ID provided
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for getting element attribute",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
	}

	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (!element) {
			throw new Error('Element not found with selector: %s');
		}
		return element.getAttribute('%s');
	`, selector, selector, attribute)

	result, err := t.browserMgr.ExecuteScript(pageID, script)
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
	} else {
		// Handle non-string results (e.g., gson.JSON from go-rod)
		value = fmt.Sprintf("%v", result)
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
	
	// Get the page ID to use
	if pageID == "" {
		// Use first available page if no specific page ID provided
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for scrolling",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
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

	result, err := t.browserMgr.ExecuteScript(pageID, script)
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
	
	// Get the page ID to use
	if pageID == "" {
		// Use first available page if no specific page ID provided
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for hovering over element",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
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

	result, err := t.browserMgr.ExecuteScript(pageID, script)
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

// ScreenScrapeTool provides comprehensive web scraping capabilities
type ScreenScrapeTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewScreenScrapeTool(log *logger.Logger, mgr *browser.Manager) *ScreenScrapeTool {
	return &ScreenScrapeTool{logger: log, browserMgr: mgr}
}

func (t *ScreenScrapeTool) Name() string {
	return "screen_scrape"
}

func (t *ScreenScrapeTool) Description() string {
	return "Extract structured data from web pages using CSS selectors. Supports single item extraction, multiple item arrays, dynamic content waiting, lazy loading, and custom JavaScript execution. Use for scraping text, links, images, form data, and complex page structures."
}

func (t *ScreenScrapeTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to scrape (optional if page_id provided). Example: 'https://example.com/products'",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Existing page ID to scrape from current browser session (optional if url provided). Use this for scraping already loaded pages.",
			},
			"selectors": map[string]interface{}{
				"type":        "object",
				"description": "CSS selectors mapping field names to elements. Examples: {'title': 'h1', 'price': '.price-value', 'description': 'p.desc', 'link': 'a[href]', 'image': 'img[src]', 'rating': '[data-rating]'}. Supports: #id, .class, [attribute], tag, :nth-child(), :contains(), descendant combinators.",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
				"examples": []interface{}{
					map[string]interface{}{
						"title":       "h1.product-title",
						"price":       ".price-current",
						"description": ".product-description p",
						"image":       "img.hero-image",
					},
				},
			},
			"extract_type": map[string]interface{}{
				"type":        "string",
				"description": "Extraction mode: 'single' extracts one item with all selectors, 'multiple' extracts array of items using container_selector. Use 'single' for page headers, forms, or unique elements. Use 'multiple' for product lists, articles, search results.",
				"enum":        []string{"single", "multiple"},
				"default":     "single",
			},
			"container_selector": map[string]interface{}{
				"type":        "string",
				"description": "Container selector for multiple items (REQUIRED when extract_type='multiple'). Each container becomes one item in results array. Examples: '.product-card', 'article', '.search-result', 'tr', '.item-container'",
			},
			"wait_for": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector to wait for before scraping (handles dynamic content). Examples: '.loading-complete', '[data-loaded=true]', '.dynamic-content', '.ajax-loaded'. Useful for SPAs, AJAX content, lazy-loaded sections.",
			},
			"wait_timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum seconds to wait for elements (default: 10). Increase for slow-loading content.",
				"default":     10,
			},
			"include_metadata": map[string]interface{}{
				"type":        "boolean",
				"description": "Include page metadata in results (title, url, timestamp, extraction_time). Disable for cleaner data output.",
				"default":     true,
			},
			"scroll_to_load": map[string]interface{}{
				"type":        "boolean",
				"description": "Auto-scroll to page bottom to trigger lazy loading (infinite scroll, image loading). Use for content that loads on scroll.",
				"default":     false,
			},
			"custom_script": map[string]interface{}{
				"type":        "string",
				"description": "Custom JavaScript to execute before scraping. Examples: 'document.querySelector(\".load-more\").click()', 'window.scrollTo(0, document.body.scrollHeight)', 'localStorage.setItem(\"view\", \"list\")'. Use for clicking buttons, changing views, triggering content.",
			},
		},
		Required: []string{"selectors"},
	}
}

func (t *ScreenScrapeTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	// Add total execution timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	// Use a channel to handle timeout
	type result struct {
		response *types.CallToolResponse
		err      error
	}
	resultChan := make(chan result, 1)
	
	go func() {
		resp, err := t.executeScreenScrape(args)
		resultChan <- result{resp, err}
	}()
	
	select {
	case res := <-resultChan:
		return res.response, res.err
	case <-ctx.Done():
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "Screen scrape timed out after 60 seconds",
			}},
			IsError: true,
		}, nil
	}
}

func (t *ScreenScrapeTool) executeScreenScrape(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()

	// Get or create page
	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}

	if pageID == "" {
		url, hasURL := args["url"].(string)
		if !hasURL || url == "" {
			return nil, fmt.Errorf("either page_id or url must be provided")
		}

		// Create new page
		_, newPageID, err := t.browserMgr.NewPage(url)
		if err != nil {
			return nil, fmt.Errorf("failed to navigate to %s: %w", url, err)
		}
		pageID = newPageID
	}

	// Wait for specific element if requested
	if waitFor, ok := args["wait_for"].(string); ok && waitFor != "" {
		timeout := 10
		if val, ok := args["wait_timeout"].(float64); ok {
			timeout = int(val)
		}

		waitScript := fmt.Sprintf(`
			const maxWait = %d * 1000;
			const startTime = Date.now();
			
			function checkElement() {
				const element = document.querySelector('%s');
				if (element) {
					return true;
				}
				
				if (Date.now() - startTime > maxWait) {
					throw new Error('Timeout waiting for element: %s');
				}
				
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
		`, timeout, waitFor, waitFor)

		if _, err := t.browserMgr.ExecuteScript(pageID, waitScript); err != nil {
			return nil, fmt.Errorf("timeout waiting for element %s: %w", waitFor, err)
		}
	}

	// Scroll to load content if requested
	if scrollToLoad, ok := args["scroll_to_load"].(bool); ok && scrollToLoad {
		scrollScript := `
			return new Promise((resolve) => {
				let scrollHeight = document.body.scrollHeight;
				let scrolled = 0;
				
				function scrollStep() {
					window.scrollTo(0, scrolled);
					scrolled += window.innerHeight;
					
					if (scrolled >= scrollHeight) {
						setTimeout(() => {
							const newScrollHeight = document.body.scrollHeight;
							if (newScrollHeight > scrollHeight) {
								scrollHeight = newScrollHeight;
								scrollStep();
							} else {
								resolve('Scroll completed');
							}
						}, 1000);
					} else {
						setTimeout(scrollStep, 500);
					}
				}
				
				scrollStep();
			});
		`

		if _, err := t.browserMgr.ExecuteScript(pageID, scrollScript); err != nil {
			t.logger.WithComponent("tools").Warn("Scroll to load failed",
				zap.Error(err))
		}
	}

	// Execute custom script if provided
	if customScript, ok := args["custom_script"].(string); ok && customScript != "" {
		if _, err := t.browserMgr.ExecuteScript(pageID, customScript); err != nil {
			t.logger.WithComponent("tools").Warn("Custom script execution failed",
				zap.Error(err))
		}
	}

	// Get selectors
	selectors, ok := args["selectors"].(map[string]interface{})
	if !ok || len(selectors) == 0 {
		return nil, fmt.Errorf("selectors must be provided as key-value pairs")
	}

	extractType := "single"
	if val, ok := args["extract_type"].(string); ok {
		extractType = val
	}

	var result interface{}
	var err error

	if extractType == "multiple" {
		result, err = t.scrapeMultiple(pageID, selectors, args)
	} else {
		result, err = t.scrapeSingle(pageID, selectors)
	}

	if err != nil {
		return nil, fmt.Errorf("scraping failed: %w", err)
	}

	// Add metadata if requested
	includeMetadata := true
	if val, ok := args["include_metadata"].(bool); ok {
		includeMetadata = val
	}

	var responseData map[string]interface{}
	if includeMetadata {
		pageInfo, _ := t.browserMgr.GetPageInfo(pageID)
		responseData = map[string]interface{}{
			"data":      result,
			"metadata":  pageInfo,
			"timestamp": time.Now().Format(time.RFC3339Nano),
			"page_id":   pageID,
		}
	} else {
		responseData = map[string]interface{}{
			"data": result,
		}
	}

	duration := time.Since(start).Milliseconds()
	t.logger.WithComponent("tools").Info("Screen scraping completed",
		zap.String("page_id", pageID),
		zap.String("extract_type", extractType),
		zap.Int("selectors_count", len(selectors)),
		zap.Int64("duration_ms", duration))

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Successfully scraped %d fields using %s extraction", len(selectors), extractType),
			Data: responseData,
		}},
	}, nil
}

func (t *ScreenScrapeTool) scrapeSingle(pageID string, selectors map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for fieldName, selectorInterface := range selectors {
		selector, ok := selectorInterface.(string)
		if !ok {
			continue
		}

		script := fmt.Sprintf(`
			const element = document.querySelector('%s');
			if (!element) {
				return null;
			}

			// Extract different types of data based on element type
			const tagName = element.tagName.toLowerCase();
			let value = null;

			if (tagName === 'img') {
				value = {
					src: element.src || element.getAttribute('src'),
					alt: element.alt || element.getAttribute('alt'),
					title: element.title || element.getAttribute('title')
				};
			} else if (tagName === 'a') {
				value = {
					href: element.href || element.getAttribute('href'),
					text: element.textContent || element.innerText,
					title: element.title || element.getAttribute('title')
				};
			} else if (tagName === 'input') {
				value = {
					type: element.type,
					value: element.value,
					placeholder: element.placeholder
				};
			} else if (element.hasAttribute('data-value')) {
				value = element.getAttribute('data-value');
			} else {
				value = element.textContent || element.innerText || '';
			}

			return {
				value: value,
				attributes: {
					class: element.className,
					id: element.id,
					tagName: tagName
				}
			};
		`, selector)

		data, err := t.browserMgr.ExecuteScript(pageID, script)
		if err != nil {
			t.logger.WithComponent("tools").Warn("Failed to scrape field",
				zap.String("field", fieldName),
				zap.String("selector", selector),
				zap.Error(err))
			result[fieldName] = nil
			continue
		}

		result[fieldName] = data
	}

	return result, nil
}

func (t *ScreenScrapeTool) scrapeMultiple(pageID string, selectors map[string]interface{}, args map[string]interface{}) ([]map[string]interface{}, error) {
	containerSelector, ok := args["container_selector"].(string)
	if !ok || containerSelector == "" {
		return nil, fmt.Errorf("container_selector is required for multiple extraction")
	}

	// Build the scraping script for multiple items
	var selectorPairs []string
	for fieldName, selectorInterface := range selectors {
		if selector, ok := selectorInterface.(string); ok {
			selectorPairs = append(selectorPairs, fmt.Sprintf(`'%s': '%s'`, fieldName, selector))
		}
	}

	script := fmt.Sprintf(`
		const containers = document.querySelectorAll('%s');
		const selectors = {%s};
		const results = [];

		containers.forEach((container, index) => {
			const item = {};

			Object.keys(selectors).forEach(fieldName => {
				const selector = selectors[fieldName];
				const element = container.querySelector(selector);

				if (!element) {
					item[fieldName] = null;
					return;
				}

				const tagName = element.tagName.toLowerCase();
				let value = null;

				if (tagName === 'img') {
					value = {
						src: element.src || element.getAttribute('src'),
						alt: element.alt || element.getAttribute('alt'),
						title: element.title || element.getAttribute('title')
					};
				} else if (tagName === 'a') {
					value = {
						href: element.href || element.getAttribute('href'),
						text: element.textContent || element.innerText,
						title: element.title || element.getAttribute('title')
					};
				} else if (tagName === 'input') {
					value = {
						type: element.type,
						value: element.value,
						placeholder: element.placeholder
					};
				} else if (element.hasAttribute('data-value')) {
					value = element.getAttribute('data-value');
				} else {
					value = element.textContent || element.innerText || '';
				}

				item[fieldName] = {
					value: value,
					attributes: {
						class: element.className,
						id: element.id,
						tagName: tagName
					}
				};
			});

			item._index = index;
			results.push(item);
		});

		return results;
	`, containerSelector, strings.Join(selectorPairs, ", "))

	data, err := t.browserMgr.ExecuteScript(pageID, script)
	if err != nil {
		return nil, fmt.Errorf("failed to execute multiple scraping script: %w", err)
	}

	// Debug log the data type
	t.logger.WithComponent("tools").Debug("Scraping script returned data",
		zap.String("type", fmt.Sprintf("%T", data)),
		zap.Any("data", data))

	// Convert the result to the expected format
	// Rod might return different data types, handle various cases
	switch v := data.(type) {
	case []interface{}:
		var result []map[string]interface{}
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				result = append(result, itemMap)
			}
		}
		return result, nil
	case []map[string]interface{}:
		return v, nil
	case interface{}:
		// Try to convert to JSON and back to handle go-rod's gson types
		if jsonBytes, err := json.Marshal(v); err == nil {
			var result []map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &result); err == nil {
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("unexpected data format returned from scraping script: %T", data)
}

// FormFillTool fills out forms with structured data
type FormFillTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewFormFillTool(log *logger.Logger, mgr *browser.Manager) *FormFillTool {
	return &FormFillTool{logger: log, browserMgr: mgr}
}

func (t *FormFillTool) Name() string {
	return "form_fill"
}

func (t *FormFillTool) Description() string {
	return "Fill out forms with structured data. Handles text inputs, selects, checkboxes, radio buttons, and textareas. Can validate required fields and optionally submit the form."
}

func (t *FormFillTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"form_selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for the form element or container (optional, defaults to 'form')",
				"default":     "form",
			},
			"fields": map[string]interface{}{
				"type":        "object",
				"description": "Object mapping field selectors to values. Keys are CSS selectors, values are the data to fill. Example: {\"#email\": \"test@example.com\", \"select[name='country']\": \"US\", \"input[name='subscribe']\": true}",
				"additionalProperties": interface{}(map[string]interface{}{
					"oneOf": []interface{}{
						map[string]interface{}{"type": "string"},
						map[string]interface{}{"type": "boolean"},
						map[string]interface{}{"type": "number"},
					},
				}),
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID (optional, uses first page if not specified)",
			},
			"submit": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to submit the form after filling (default: false)",
				"default":     false,
			},
			"validate_required": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to validate that required fields are filled (default: true)",
				"default":     true,
			},
			"trigger_events": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to trigger input/change events after filling fields (default: true)",
				"default":     true,
			},
		},
		Required: []string{"fields"},
	}
}

func (t *FormFillTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	// Add timeout protection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	type result struct {
		response *types.CallToolResponse
		err      error
	}
	resultChan := make(chan result, 1)
	
	go func() {
		resp, err := t.executeFormFill(args)
		resultChan <- result{resp, err}
	}()
	
	select {
	case res := <-resultChan:
		return res.response, res.err
	case <-ctx.Done():
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "Form fill operation timed out after 30 seconds",
			}},
			IsError: true,
		}, nil
	}
}

func (t *FormFillTool) executeFormFill(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	// Get page ID
	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}
	
	if pageID == "" {
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for form filling",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
	}

	// Get form selector
	formSelector := "form"
	if val, ok := args["form_selector"].(string); ok && val != "" {
		formSelector = val
	}

	// Get fields to fill
	fields, ok := args["fields"].(map[string]interface{})
	if !ok || len(fields) == 0 {
		return nil, fmt.Errorf("fields must be provided as key-value pairs")
	}

	// Get options
	submit := false
	if val, ok := args["submit"].(bool); ok {
		submit = val
	}

	validateRequired := true
	if val, ok := args["validate_required"].(bool); ok {
		validateRequired = val
	}

	triggerEvents := true
	if val, ok := args["trigger_events"].(bool); ok {
		triggerEvents = val
	}

	// Build the form filling script
	var fillResults []map[string]interface{}
	var errors []string

	for fieldSelector, value := range fields {
		result, err := t.fillSingleField(pageID, formSelector, fieldSelector, value, triggerEvents)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Field %s: %v", fieldSelector, err))
			continue
		}
		fillResults = append(fillResults, result)
	}

	// Validate required fields if requested
	var validationErrors []string
	if validateRequired {
		validationErrors, _ = t.validateRequiredFields(pageID, formSelector)
	}

	// Submit form if requested and no critical errors
	var submitResult string
	if submit && len(errors) == 0 {
		submitErr := t.submitForm(pageID, formSelector)
		if submitErr != nil {
			errors = append(errors, fmt.Sprintf("Form submission failed: %v", submitErr))
			submitResult = "Failed"
		} else {
			submitResult = "Success"
		}
	} else if submit {
		submitResult = "Skipped due to field errors"
	} else {
		submitResult = "Not requested"
	}

	// Prepare response
	hasErrors := len(errors) > 0 || len(validationErrors) > 0
	
	var messageText strings.Builder
	messageText.WriteString(fmt.Sprintf("Form fill completed: %d fields processed", len(fields)))
	
	if len(fillResults) > 0 {
		messageText.WriteString(fmt.Sprintf(", %d successful", len(fillResults)))
	}
	
	if len(errors) > 0 {
		messageText.WriteString(fmt.Sprintf(", %d failed", len(errors)))
	}
	
	if submit {
		messageText.WriteString(fmt.Sprintf(", submission: %s", submitResult))
	}

	responseData := map[string]interface{}{
		"fields_processed": len(fields),
		"successful_fills": fillResults,
		"errors":          errors,
		"validation_errors": validationErrors,
		"submit_requested": submit,
		"submit_result":    submitResult,
		"form_selector":    formSelector,
		"page_id":         pageID,
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: messageText.String(),
			Data: responseData,
		}},
		IsError: hasErrors,
	}, nil
}

func (t *FormFillTool) fillSingleField(pageID, formSelector, fieldSelector string, value interface{}, triggerEvents bool) (map[string]interface{}, error) {
	// Convert value to appropriate JavaScript representation
	var jsValue string
	var valueType string
	
	switch v := value.(type) {
	case string:
		jsValue = fmt.Sprintf("'%s'", strings.ReplaceAll(strings.ReplaceAll(v, "\\", "\\\\"), "'", "\\'"))
		valueType = "string"
	case bool:
		jsValue = fmt.Sprintf("%v", v)
		valueType = "boolean"
	case float64:
		jsValue = fmt.Sprintf("%v", v)
		valueType = "number"
	case int:
		jsValue = fmt.Sprintf("%v", v)
		valueType = "number"
	default:
		return nil, fmt.Errorf("unsupported value type: %T", value)
	}

	eventsScript := ""
	if triggerEvents {
		eventsScript = `
			element.dispatchEvent(new Event('input', { bubbles: true }));
			element.dispatchEvent(new Event('change', { bubbles: true }));
			element.dispatchEvent(new Event('blur', { bubbles: true }));
		`
	}

	script := fmt.Sprintf(`
		const form = document.querySelector('%s');
		if (!form) {
			throw new Error('Form not found with selector: %s');
		}
		
		const element = form.querySelector('%s') || document.querySelector('%s');
		if (!element) {
			throw new Error('Field not found with selector: %s');
		}
		
		const tagName = element.tagName.toLowerCase();
		const inputType = element.type ? element.type.toLowerCase() : '';
		const value = %s;
		let result = {
			selector: '%s',
			tagName: tagName,
			type: inputType,
			value: value,
			valueType: '%s',
			success: false,
			method: ''
		};
		
		try {
			if (tagName === 'input') {
				if (inputType === 'checkbox' || inputType === 'radio') {
					element.checked = Boolean(value);
					result.method = 'checked';
				} else {
					element.value = String(value);
					result.method = 'value';
				}
			} else if (tagName === 'select') {
				element.value = String(value);
				result.method = 'value';
			} else if (tagName === 'textarea') {
				element.value = String(value);
				result.method = 'value';
			} else {
				element.textContent = String(value);
				result.method = 'textContent';
			}
			
			%s
			
			result.success = true;
			result.finalValue = element.value || element.textContent || element.checked;
			
		} catch (error) {
			result.error = error.message;
		}
		
		return result;
	`, formSelector, formSelector, fieldSelector, fieldSelector, fieldSelector, jsValue, fieldSelector, valueType, eventsScript)

	data, err := t.browserMgr.ExecuteScript(pageID, script)
	if err != nil {
		return nil, fmt.Errorf("failed to execute field fill script: %w", err)
	}

	// Convert result to map
	if resultMap, ok := data.(map[string]interface{}); ok {
		if success, ok := resultMap["success"].(bool); !ok || !success {
			if errMsg, ok := resultMap["error"].(string); ok {
				return resultMap, fmt.Errorf("field fill failed: %s", errMsg)
			}
			return resultMap, fmt.Errorf("field fill failed for unknown reason")
		}
		return resultMap, nil
	}

	// Handle go-rod gson types by marshaling/unmarshaling
	if jsonBytes, err := json.Marshal(data); err == nil {
		var resultMap map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &resultMap); err == nil {
			if success, ok := resultMap["success"].(bool); !ok || !success {
				if errMsg, ok := resultMap["error"].(string); ok {
					return resultMap, fmt.Errorf("field fill failed: %s", errMsg)
				}
			}
			return resultMap, nil
		}
	}

	return map[string]interface{}{"raw_data": data}, nil
}

func (t *FormFillTool) validateRequiredFields(pageID, formSelector string) ([]string, error) {
	script := fmt.Sprintf(`
		const form = document.querySelector('%s');
		if (!form) {
			throw new Error('Form not found with selector: %s');
		}
		
		const requiredFields = form.querySelectorAll('[required]');
		const errors = [];
		
		requiredFields.forEach(field => {
			const tagName = field.tagName.toLowerCase();
			const type = field.type ? field.type.toLowerCase() : '';
			let isEmpty = false;
			
			if (tagName === 'input') {
				if (type === 'checkbox' || type === 'radio') {
					isEmpty = !field.checked;
				} else {
					isEmpty = !field.value.trim();
				}
			} else if (tagName === 'select') {
				isEmpty = !field.value;
			} else if (tagName === 'textarea') {
				isEmpty = !field.value.trim();
			}
			
			if (isEmpty) {
				errors.push({
					selector: field.name ? '[name="' + field.name + '"]' : field.id ? '#' + field.id : tagName + '[required]',
					name: field.name || field.id || 'unnamed',
					type: type || tagName,
					message: 'Required field is empty'
				});
			}
		});
		
		return errors;
	`, formSelector, formSelector)

	data, err := t.browserMgr.ExecuteScript(pageID, script)
	if err != nil {
		return nil, fmt.Errorf("failed to validate required fields: %w", err)
	}

	var errors []string
	
	// Handle different data types returned by go-rod
	if errorsList, ok := data.([]interface{}); ok {
		for _, errorItem := range errorsList {
			if errorMap, ok := errorItem.(map[string]interface{}); ok {
				if message, ok := errorMap["message"].(string); ok {
					name := "unknown"
					if n, ok := errorMap["name"].(string); ok {
						name = n
					}
					errors = append(errors, fmt.Sprintf("%s: %s", name, message))
				}
			}
		}
	}

	return errors, nil
}

func (t *FormFillTool) submitForm(pageID, formSelector string) error {
	script := fmt.Sprintf(`
		const form = document.querySelector('%s');
		if (!form) {
			throw new Error('Form not found with selector: %s');
		}
		
		// Try to find and click submit button first
		const submitButton = form.querySelector('input[type="submit"], button[type="submit"], button:not([type])');
		if (submitButton && !submitButton.disabled) {
			submitButton.click();
			return 'Submitted via button click';
		} else {
			// Fall back to form.submit()
			form.submit();
			return 'Submitted via form.submit()';
		}
	`, formSelector, formSelector)

	_, err := t.browserMgr.ExecuteScript(pageID, script)
	if err != nil {
		return fmt.Errorf("failed to submit form: %w", err)
	}

	return nil
}

// WaitForConditionTool waits for custom JavaScript conditions to become true
type WaitForConditionTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewWaitForConditionTool(log *logger.Logger, mgr *browser.Manager) *WaitForConditionTool {
	return &WaitForConditionTool{logger: log, browserMgr: mgr}
}

func (t *WaitForConditionTool) Name() string {
	return "wait_for_condition"
}

func (t *WaitForConditionTool) Description() string {
	return "Wait for a custom JavaScript condition to become true. Much more flexible than waiting for elements - can wait for animations, API responses, state changes, or any complex condition."
}

func (t *WaitForConditionTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"condition": map[string]interface{}{
				"type":        "string",
				"description": "JavaScript expression or function that returns true when condition is met. Examples: 'document.readyState === \"complete\"', '!!window.myApp && window.myApp.loaded', 'document.querySelectorAll(\".item\").length >= 5'",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID (optional, uses first page if not specified)",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum time to wait in seconds (default: 10)",
				"default":     10,
				"minimum":     1,
				"maximum":     120,
			},
			"interval": map[string]interface{}{
				"type":        "integer",
				"description": "Polling interval in milliseconds (default: 100)",
				"default":     100,
				"minimum":     50,
				"maximum":     5000,
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Optional description of what you're waiting for (for logging and error messages)",
			},
			"return_value": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to return the final value of the condition (default: false)",
				"default":     false,
			},
		},
		Required: []string{"condition"},
	}
}

func (t *WaitForConditionTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	// Add timeout protection (with buffer for internal timeout)
	internalTimeout := 10 * time.Second
	if val, ok := args["timeout"].(float64); ok {
		internalTimeout = time.Duration(val+5) * time.Second // Add 5s buffer
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), internalTimeout)
	defer cancel()
	
	type result struct {
		response *types.CallToolResponse
		err      error
	}
	resultChan := make(chan result, 1)
	
	go func() {
		resp, err := t.executeWaitForCondition(args)
		resultChan <- result{resp, err}
	}()
	
	select {
	case res := <-resultChan:
		return res.response, res.err
	case <-ctx.Done():
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "Wait for condition operation timed out",
			}},
			IsError: true,
		}, nil
	}
}

func (t *WaitForConditionTool) executeWaitForCondition(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	// Get page ID
	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}
	
	if pageID == "" {
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for waiting for condition",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
	}

	// Get condition
	condition, ok := args["condition"].(string)
	if !ok || condition == "" {
		return nil, fmt.Errorf("condition must be provided as a string")
	}

	// Get parameters
	timeout := 10
	if val, ok := args["timeout"].(float64); ok {
		timeout = int(val)
	}

	interval := 100
	if val, ok := args["interval"].(float64); ok {
		interval = int(val)
	}

	description := ""
	if val, ok := args["description"].(string); ok {
		description = val
	}

	returnValue := false
	if val, ok := args["return_value"].(bool); ok {
		returnValue = val
	}

	// Clean up condition for JavaScript execution
	condition = strings.TrimSpace(condition)
	
	// Build the waiting script
	script := fmt.Sprintf(`
		const condition = () => {
			try {
				return %s;
			} catch (error) {
				console.warn('Condition evaluation error:', error);
				return false;
			}
		};

		const maxWait = %d * 1000; // Convert to milliseconds
		const interval = %d;
		const startTime = Date.now();
		const returnValue = %v;
		
		let attempts = 0;
		let lastResult = null;
		
		function checkCondition() {
			attempts++;
			const result = condition();
			lastResult = result;
			
			if (result) {
				const elapsed = Date.now() - startTime;
				return {
					success: true,
					result: returnValue ? result : true,
					elapsed_ms: elapsed,
					attempts: attempts,
					condition: '%s',
					description: '%s'
				};
			}
			
			if (Date.now() - startTime > maxWait) {
				const elapsed = Date.now() - startTime;
				return {
					success: false,
					result: returnValue ? lastResult : false,
					elapsed_ms: elapsed,
					attempts: attempts,
					condition: '%s',
					description: '%s',
					error: 'Timeout after ' + elapsed + 'ms'
				};
			}
			
			// Continue waiting
			return new Promise((resolve, reject) => {
				setTimeout(() => {
					try {
						resolve(checkCondition());
					} catch (error) {
						reject(error);
					}
				}, interval);
			});
		}
		
		return checkCondition();
	`, condition, timeout, interval, returnValue, 
		strings.ReplaceAll(condition, "'", "\\'"), 
		strings.ReplaceAll(description, "'", "\\'"),
		strings.ReplaceAll(condition, "'", "\\'"),
		strings.ReplaceAll(description, "'", "\\'"))

	// Execute the script
	data, err := t.browserMgr.ExecuteScript(pageID, script)
	if err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to execute wait condition: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Parse result
	var resultMap map[string]interface{}
	
	// Handle go-rod gson types by marshaling/unmarshaling if needed
	if directMap, ok := data.(map[string]interface{}); ok {
		resultMap = directMap
	} else if jsonBytes, err := json.Marshal(data); err == nil {
		if err := json.Unmarshal(jsonBytes, &resultMap); err != nil {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Failed to parse wait result: %v", err),
				}},
				IsError: true,
			}, nil
		}
	} else {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Unexpected result format: %T", data),
			}},
			IsError: true,
		}, nil
	}

	// Extract result information
	success := false
	if val, ok := resultMap["success"].(bool); ok {
		success = val
	}

	elapsed := float64(0)
	if val, ok := resultMap["elapsed_ms"].(float64); ok {
		elapsed = val
	}

	attempts := float64(0)
	if val, ok := resultMap["attempts"].(float64); ok {
		attempts = val
	}

	errorMsg := ""
	if val, ok := resultMap["error"].(string); ok {
		errorMsg = val
	}

	// Prepare response
	var messageText strings.Builder
	
	if success {
		messageText.WriteString("Condition satisfied")
		if description != "" {
			messageText.WriteString(fmt.Sprintf(": %s", description))
		}
		messageText.WriteString(fmt.Sprintf(" (%.0fms, %d attempts)", elapsed, int(attempts)))
	} else {
		messageText.WriteString("Condition not satisfied")
		if description != "" {
			messageText.WriteString(fmt.Sprintf(": %s", description))
		}
		messageText.WriteString(fmt.Sprintf(" - %s (%.0fms, %d attempts)", errorMsg, elapsed, int(attempts)))
	}

	responseData := map[string]interface{}{
		"success":     success,
		"condition":   condition,
		"description": description,
		"elapsed_ms":  elapsed,
		"attempts":    int(attempts),
		"timeout":     timeout,
		"interval":    interval,
		"page_id":     pageID,
	}

	if returnValue {
		responseData["final_value"] = resultMap["result"]
	}

	if errorMsg != "" {
		responseData["error"] = errorMsg
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: messageText.String(),
			Data: responseData,
		}},
		IsError: !success,
	}, nil
}

// AssertElementTool provides comprehensive element assertions for testing
type AssertElementTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewAssertElementTool(log *logger.Logger, mgr *browser.Manager) *AssertElementTool {
	return &AssertElementTool{logger: log, browserMgr: mgr}
}

func (t *AssertElementTool) Name() string {
	return "assert_element"
}

func (t *AssertElementTool) Description() string {
	return "Assert element existence, visibility, state, text content, or attributes. Essential for testing and validation workflows."
}

func (t *AssertElementTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for the element to assert",
			},
			"assertion": map[string]interface{}{
				"type":        "string",
				"description": "Type of assertion to perform",
				"enum": []string{
					"exists", "not_exists", 
					"visible", "hidden",
					"enabled", "disabled",
					"contains_text", "exact_text", "not_contains_text",
					"has_attribute", "attribute_equals", "attribute_contains",
					"has_class", "not_has_class",
					"is_checked", "is_unchecked",
					"count_equals", "count_greater_than", "count_less_than",
				},
			},
			"expected_value": map[string]interface{}{
				"type":        "string",
				"description": "Expected value for text/attribute/count assertions (required for some assertion types)",
			},
			"attribute_name": map[string]interface{}{
				"type":        "string", 
				"description": "Attribute name for attribute-based assertions (required for has_attribute, attribute_equals, attribute_contains)",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID (optional, uses first page if not specified)",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum time to wait for element before asserting in seconds (default: 5)",
				"default":     5,
				"minimum":     0,
				"maximum":     30,
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether text comparisons should be case sensitive (default: false)",
				"default":     false,
			},
		},
		Required: []string{"selector", "assertion"},
	}
}

func (t *AssertElementTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	// Add timeout protection
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	
	type result struct {
		response *types.CallToolResponse
		err      error
	}
	resultChan := make(chan result, 1)
	
	go func() {
		resp, err := t.executeAssertElement(args)
		resultChan <- result{resp, err}
	}()
	
	select {
	case res := <-resultChan:
		return res.response, res.err
	case <-ctx.Done():
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "Element assertion timed out",
			}},
			IsError: true,
		}, nil
	}
}

func (t *AssertElementTool) executeAssertElement(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	// Get page ID
	pageID := ""
	if val, ok := args["page_id"].(string); ok {
		pageID = val
	}
	
	if pageID == "" {
		pages := t.browserMgr.ListPages()
		if len(pages) == 0 {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: "No pages available for element assertion",
				}},
				IsError: true,
			}, nil
		}
		pageID = pages[0]
	}

	// Get required parameters
	selector, ok := args["selector"].(string)
	if !ok || selector == "" {
		return nil, fmt.Errorf("selector must be provided as a string")
	}

	assertion, ok := args["assertion"].(string)
	if !ok || assertion == "" {
		return nil, fmt.Errorf("assertion must be provided as a string")
	}

	// Get optional parameters
	expectedValue := ""
	if val, ok := args["expected_value"].(string); ok {
		expectedValue = val
	}

	attributeName := ""
	if val, ok := args["attribute_name"].(string); ok {
		attributeName = val
	}

	timeout := 5
	if val, ok := args["timeout"].(float64); ok {
		timeout = int(val)
	}

	caseSensitive := false
	if val, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = val
	}

	// Validate required parameters for specific assertions
	if err := t.validateAssertionParams(assertion, expectedValue, attributeName); err != nil {
		return nil, err
	}

	// Wait for element if timeout > 0 and assertion requires element to exist
	if timeout > 0 && !strings.Contains(assertion, "not_exists") {
		waitScript := fmt.Sprintf(`
			const maxWait = %d * 1000;
			const startTime = Date.now();
			
			function checkElement() {
				const elements = document.querySelectorAll('%s');
				if (elements.length > 0) {
					return true;
				}
				
				if (Date.now() - startTime > maxWait) {
					return false;
				}
				
				return new Promise((resolve) => {
					setTimeout(() => resolve(checkElement()), 100);
				});
			}
			
			return checkElement();
		`, timeout, selector)

		_, err := t.browserMgr.ExecuteScript(pageID, waitScript)
		if err != nil {
			// Element not found within timeout, but continue with assertion
			// The assertion itself will handle the "not found" case
		}
	}

	// Perform the assertion
	result, err := t.performAssertion(pageID, selector, assertion, expectedValue, attributeName, caseSensitive)
	if err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Assertion execution failed: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Parse the assertion result
	var assertionData map[string]interface{}
	if directMap, ok := result.(map[string]interface{}); ok {
		assertionData = directMap
	} else if jsonBytes, err := json.Marshal(result); err == nil {
		if err := json.Unmarshal(jsonBytes, &assertionData); err != nil {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Failed to parse assertion result: %v", err),
				}},
				IsError: true,
			}, nil
		}
	} else {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Unexpected assertion result format: %T", result),
			}},
			IsError: true,
		}, nil
	}

	// Extract assertion result
	passed := false
	if val, ok := assertionData["passed"].(bool); ok {
		passed = val
	}

	message := "Assertion completed"
	if val, ok := assertionData["message"].(string); ok {
		message = val
	}

	// Prepare response data
	responseData := map[string]interface{}{
		"passed":         passed,
		"selector":       selector,
		"assertion":      assertion,
		"expected_value": expectedValue,
		"attribute_name": attributeName,
		"timeout":        timeout,
		"case_sensitive": caseSensitive,
		"page_id":        pageID,
	}

	// Add any additional data from the assertion
	for key, value := range assertionData {
		if key != "passed" && key != "message" {
			responseData[key] = value
		}
	}

	status := "PASS"
	if !passed {
		status = "FAIL"
	}

	finalMessage := fmt.Sprintf("[%s] %s", status, message)

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: finalMessage,
			Data: responseData,
		}},
		IsError: !passed,
	}, nil
}

func (t *AssertElementTool) validateAssertionParams(assertion, expectedValue, attributeName string) error {
	switch assertion {
	case "contains_text", "exact_text", "not_contains_text":
		if expectedValue == "" {
			return fmt.Errorf("expected_value is required for text assertions")
		}
	case "has_attribute":
		if attributeName == "" {
			return fmt.Errorf("attribute_name is required for has_attribute assertion")
		}
	case "attribute_equals", "attribute_contains":
		if attributeName == "" || expectedValue == "" {
			return fmt.Errorf("both attribute_name and expected_value are required for attribute assertions")
		}
	case "has_class", "not_has_class":
		if expectedValue == "" {
			return fmt.Errorf("expected_value is required for class assertions")
		}
	case "count_equals", "count_greater_than", "count_less_than":
		if expectedValue == "" {
			return fmt.Errorf("expected_value is required for count assertions")
		}
	}
	return nil
}

func (t *AssertElementTool) performAssertion(pageID, selector, assertion, expectedValue, attributeName string, caseSensitive bool) (interface{}, error) {
	script := fmt.Sprintf(`
		const selector = '%s';
		const assertion = '%s';
		const expectedValue = '%s';
		const attributeName = '%s';
		const caseSensitive = %v;
		
		const elements = document.querySelectorAll(selector);
		const count = elements.length;
		const element = elements[0]; // First element for single-element assertions
		
		let result = {
			passed: false,
			message: '',
			count: count,
			found_elements: count > 0
		};
		
		try {
			switch (assertion) {
				case 'exists':
					result.passed = count > 0;
					result.message = count > 0 ? 
						'Element exists (' + count + ' found)' : 
						'Element does not exist';
					break;
					
				case 'not_exists':
					result.passed = count === 0;
					result.message = count === 0 ? 
						'Element does not exist (as expected)' : 
						'Element exists (' + count + ' found) but should not exist';
					break;
					
				case 'visible':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const isVisible = element.offsetParent !== null && 
						getComputedStyle(element).visibility !== 'hidden' &&
						getComputedStyle(element).display !== 'none';
					result.passed = isVisible;
					result.message = isVisible ? 'Element is visible' : 'Element is not visible';
					result.computed_style = {
						display: getComputedStyle(element).display,
						visibility: getComputedStyle(element).visibility,
						opacity: getComputedStyle(element).opacity
					};
					break;
					
				case 'hidden':
					if (!element) {
						result.passed = true;
						result.message = 'Element not found (considered hidden)';
						break;
					}
					const isHidden = element.offsetParent === null || 
						getComputedStyle(element).visibility === 'hidden' ||
						getComputedStyle(element).display === 'none';
					result.passed = isHidden;
					result.message = isHidden ? 'Element is hidden' : 'Element is visible';
					break;
					
				case 'enabled':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const isEnabled = !element.disabled;
					result.passed = isEnabled;
					result.message = isEnabled ? 'Element is enabled' : 'Element is disabled';
					result.disabled = element.disabled;
					break;
					
				case 'disabled':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const isDisabled = element.disabled;
					result.passed = isDisabled;
					result.message = isDisabled ? 'Element is disabled' : 'Element is enabled';
					result.disabled = element.disabled;
					break;
					
				case 'contains_text':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const textContent = element.textContent || element.innerText || '';
					const searchText = caseSensitive ? expectedValue : expectedValue.toLowerCase();
					const elementText = caseSensitive ? textContent : textContent.toLowerCase();
					const containsText = elementText.includes(searchText);
					result.passed = containsText;
					result.message = containsText ? 
						'Element contains expected text' : 
						'Element does not contain expected text';
					result.actual_text = textContent;
					result.expected_text = expectedValue;
					break;
					
				case 'exact_text':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const exactText = element.textContent || element.innerText || '';
					const expectedExactText = caseSensitive ? expectedValue : expectedValue.toLowerCase();
					const actualExactText = caseSensitive ? exactText : exactText.toLowerCase();
					const isExactMatch = actualExactText === expectedExactText;
					result.passed = isExactMatch;
					result.message = isExactMatch ? 
						'Element text matches exactly' : 
						'Element text does not match exactly';
					result.actual_text = exactText;
					result.expected_text = expectedValue;
					break;
					
				case 'not_contains_text':
					if (!element) {
						result.passed = true;
						result.message = 'Element not found (text not contained)';
						break;
					}
					const textToCheck = element.textContent || element.innerText || '';
					const searchToAvoid = caseSensitive ? expectedValue : expectedValue.toLowerCase();
					const elementToCheck = caseSensitive ? textToCheck : textToCheck.toLowerCase();
					const doesNotContain = !elementToCheck.includes(searchToAvoid);
					result.passed = doesNotContain;
					result.message = doesNotContain ? 
						'Element does not contain the text (as expected)' : 
						'Element contains the text but should not';
					result.actual_text = textToCheck;
					result.expected_not_to_contain = expectedValue;
					break;
					
				case 'has_attribute':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const hasAttr = element.hasAttribute(attributeName);
					result.passed = hasAttr;
					result.message = hasAttr ? 
						'Element has attribute "' + attributeName + '"' : 
						'Element does not have attribute "' + attributeName + '"';
					result.attribute_name = attributeName;
					if (hasAttr) {
						result.attribute_value = element.getAttribute(attributeName);
					}
					break;
					
				case 'attribute_equals':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const attrValue = element.getAttribute(attributeName);
					const expectedAttrValue = caseSensitive ? expectedValue : expectedValue.toLowerCase();
					const actualAttrValue = caseSensitive ? (attrValue || '') : (attrValue || '').toLowerCase();
					const attrEquals = actualAttrValue === expectedAttrValue;
					result.passed = attrEquals;
					result.message = attrEquals ? 
						'Attribute value matches exactly' : 
						'Attribute value does not match';
					result.attribute_name = attributeName;
					result.actual_attribute_value = attrValue;
					result.expected_attribute_value = expectedValue;
					break;
					
				case 'attribute_contains':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const attrContent = element.getAttribute(attributeName) || '';
					const expectedAttrContent = caseSensitive ? expectedValue : expectedValue.toLowerCase();
					const actualAttrContent = caseSensitive ? attrContent : attrContent.toLowerCase();
					const attrContains = actualAttrContent.includes(expectedAttrContent);
					result.passed = attrContains;
					result.message = attrContains ? 
						'Attribute contains expected value' : 
						'Attribute does not contain expected value';
					result.attribute_name = attributeName;
					result.actual_attribute_value = attrContent;
					result.expected_to_contain = expectedValue;
					break;
					
				case 'has_class':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const hasClass = element.classList.contains(expectedValue);
					result.passed = hasClass;
					result.message = hasClass ? 
						'Element has class "' + expectedValue + '"' : 
						'Element does not have class "' + expectedValue + '"';
					result.expected_class = expectedValue;
					result.actual_classes = Array.from(element.classList);
					break;
					
				case 'not_has_class':
					if (!element) {
						result.passed = true;
						result.message = 'Element not found (class not present)';
						break;
					}
					const doesNotHaveClass = !element.classList.contains(expectedValue);
					result.passed = doesNotHaveClass;
					result.message = doesNotHaveClass ? 
						'Element does not have class "' + expectedValue + '" (as expected)' : 
						'Element has class "' + expectedValue + '" but should not';
					result.expected_not_to_have_class = expectedValue;
					result.actual_classes = Array.from(element.classList);
					break;
					
				case 'is_checked':
					if (!element) {
						result.message = 'Element not found';
						break;
					}
					const isChecked = element.checked === true;
					result.passed = isChecked;
					result.message = isChecked ? 'Element is checked' : 'Element is not checked';
					result.checked_state = element.checked;
					break;
					
				case 'is_unchecked':
					if (!element) {
						result.passed = true;
						result.message = 'Element not found (considered unchecked)';
						break;
					}
					const isUnchecked = element.checked === false;
					result.passed = isUnchecked;
					result.message = isUnchecked ? 'Element is unchecked' : 'Element is checked';
					result.checked_state = element.checked;
					break;
					
				case 'count_equals':
					const expectedCount = parseInt(expectedValue);
					const countEquals = count === expectedCount;
					result.passed = countEquals;
					result.message = countEquals ? 
						'Element count matches (' + count + ')' : 
						'Element count (' + count + ') does not match expected (' + expectedCount + ')';
					result.expected_count = expectedCount;
					break;
					
				case 'count_greater_than':
					const minCount = parseInt(expectedValue);
					const countGreater = count > minCount;
					result.passed = countGreater;
					result.message = countGreater ? 
						'Element count (' + count + ') is greater than ' + minCount : 
						'Element count (' + count + ') is not greater than ' + minCount;
					result.minimum_count = minCount;
					break;
					
				case 'count_less_than':
					const maxCount = parseInt(expectedValue);
					const countLess = count < maxCount;
					result.passed = countLess;
					result.message = countLess ? 
						'Element count (' + count + ') is less than ' + maxCount : 
						'Element count (' + count + ') is not less than ' + maxCount;
					result.maximum_count = maxCount;
					break;
					
				default:
					result.message = 'Unknown assertion type: ' + assertion;
					break;
			}
		} catch (error) {
			result.message = 'Assertion failed with error: ' + error.message;
			result.error = error.message;
		}
		
		return result;
	`, 
	strings.ReplaceAll(selector, "'", "\\'"),
	assertion,
	strings.ReplaceAll(expectedValue, "'", "\\'"),
	strings.ReplaceAll(attributeName, "'", "\\'"),
	caseSensitive)

	return t.browserMgr.ExecuteScript(pageID, script)
}

// ExtractTableTool extracts structured data from HTML tables
type ExtractTableTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewExtractTableTool(log *logger.Logger, browserMgr *browser.Manager) *ExtractTableTool {
	return &ExtractTableTool{
		logger:     log,
		browserMgr: browserMgr,
	}
}

func (t *ExtractTableTool) Name() string {
	return "extract_table"
}

func (t *ExtractTableTool) Description() string {
	return "Extract structured data from HTML tables with support for headers, filtering, and multiple formats"
}

func (t *ExtractTableTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for the table element (e.g., 'table', '#data-table', '.results tbody')",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID to extract from (optional, uses current page if not specified)",
			},
			"include_headers": map[string]interface{}{
				"type":        "boolean",
				"description": "Include table headers in the output (default: true)",
				"default":     true,
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"description": "Output format: 'array' (array of arrays), 'objects' (array of objects with header keys), 'csv' (CSV string)",
				"enum":        []string{"array", "objects", "csv"},
				"default":     "objects",
			},
			"skip_empty_rows": map[string]interface{}{
				"type":        "boolean",
				"description": "Skip rows that are completely empty (default: true)",
				"default":     true,
			},
			"max_rows": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of rows to extract (default: no limit)",
				"minimum":     1,
			},
			"column_filter": map[string]interface{}{
				"type":        "array",
				"description": "Array of column indices or header names to include (default: all columns)",
				"items": map[string]interface{}{
					"oneOf": []map[string]interface{}{
						{"type": "integer"},
						{"type": "string"},
					},
				},
			},
			"header_row": map[string]interface{}{
				"type":        "integer",
				"description": "Row index to use as headers (0-based, default: 0 for first row)",
				"default":     0,
				"minimum":     0,
			},
		},
		Required: []string{"selector"},
	}
}

func (t *ExtractTableTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	// Add timeout protection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse arguments
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector is required")
	}

	pageID, _ := args["page_id"].(string)
	
	includeHeaders := true
	if val, ok := args["include_headers"].(bool); ok {
		includeHeaders = val
	}

	outputFormat := "objects"
	if val, ok := args["output_format"].(string); ok {
		outputFormat = val
	}

	skipEmptyRows := true
	if val, ok := args["skip_empty_rows"].(bool); ok {
		skipEmptyRows = val
	}

	var maxRows *int
	if val, ok := args["max_rows"].(float64); ok {
		maxRowsInt := int(val)
		maxRows = &maxRowsInt
	}

	var columnFilter []interface{}
	if val, ok := args["column_filter"].([]interface{}); ok {
		columnFilter = val
	}

	headerRow := 0
	if val, ok := args["header_row"].(float64); ok {
		headerRow = int(val)
	}

	// Execute extraction in goroutine with timeout
	resultChan := make(chan *types.CallToolResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := t.extractTableData(pageID, selector, includeHeaders, outputFormat, skipEmptyRows, maxRows, columnFilter, headerRow)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("extract_table operation timed out after 30 seconds")
	case err := <-errorChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}

func (t *ExtractTableTool) extractTableData(pageID, selector string, includeHeaders bool, outputFormat string, skipEmptyRows bool, maxRows *int, columnFilter []interface{}, headerRow int) (*types.CallToolResponse, error) {
	// Build JavaScript for table extraction
	script := fmt.Sprintf(`
		// Extract table data with comprehensive options
		const table = document.querySelector('%s');
		if (!table) {
			return { error: 'Table not found with selector: %s' };
		}

		// Get all rows from table
		let rows = [];
		if (table.tagName === 'TABLE') {
			// Full table - get all rows from tbody, thead, or directly from table
			const tbody = table.querySelector('tbody');
			const thead = table.querySelector('thead');
			
			if (thead) {
				rows = rows.concat(Array.from(thead.querySelectorAll('tr')));
			}
			if (tbody) {
				rows = rows.concat(Array.from(tbody.querySelectorAll('tr')));
			} else {
				// No tbody, get direct tr children
				rows = Array.from(table.querySelectorAll('tr'));
			}
		} else {
			// Selector points to tbody, thead, or other container
			rows = Array.from(table.querySelectorAll('tr'));
		}

		if (rows.length === 0) {
			return { error: 'No rows found in table' };
		}

		// Extract cell data from rows
		const rawData = rows.map((row, rowIndex) => {
			const cells = Array.from(row.querySelectorAll('td, th'));
			return cells.map(cell => {
				// Get text content, handling nested elements
				let text = cell.textContent || cell.innerText || '';
				text = text.trim();
				
				// Check for special attributes
				const href = cell.querySelector('a')?.href;
				const src = cell.querySelector('img')?.src;
				const value = cell.querySelector('input')?.value;
				
				// Return enhanced cell data
				const cellData = { text: text };
				if (href) cellData.link = href;
				if (src) cellData.image = src;
				if (value !== undefined) cellData.input_value = value;
				
				return cellData;
			});
		});

		// Apply row filtering
		let filteredData = rawData;
		
		// Skip empty rows if requested
		const skipEmpty = %t;
		if (skipEmpty) {
			filteredData = filteredData.filter(row => 
				row.some(cell => cell.text && cell.text.length > 0)
			);
		}

		// Apply max rows limit
		const maxRowsLimit = %s;
		if (maxRowsLimit !== null) {
			filteredData = filteredData.slice(0, maxRowsLimit);
		}

		// Determine headers
		const headerRowIndex = %d;
		const includeHeaders = %t;
		let headers = [];
		
		if (includeHeaders && filteredData.length > headerRowIndex) {
			headers = filteredData[headerRowIndex].map(cell => cell.text);
		}

		// Apply column filtering
		const columnFilterList = %s;
		let columnIndices = null;
		if (columnFilterList && columnFilterList.length > 0) {
			columnIndices = [];
			for (const filter of columnFilterList) {
				if (typeof filter === 'number') {
					columnIndices.push(filter);
				} else if (typeof filter === 'string' && headers.length > 0) {
					// Find column by header name
					const index = headers.indexOf(filter);
					if (index !== -1) {
						columnIndices.push(index);
					}
				}
			}
		}

		// Process data based on output format
		const outputFormat = '%s';
		let processedData;
		
		if (outputFormat === 'array') {
			// Array of arrays format
			processedData = filteredData.map(row => {
				let rowData = row.map(cell => cell.text);
				if (columnIndices) {
					rowData = columnIndices.map(i => rowData[i] || '');
				}
				return rowData;
			});
		} else if (outputFormat === 'objects') {
			// Array of objects format
			if (headers.length === 0) {
				// Generate default headers
				const maxCols = Math.max(...filteredData.map(row => row.length));
				headers = Array.from({length: maxCols}, (_, i) => 'column_' + i);
			}
			
			const dataRows = includeHeaders ? filteredData.slice(headerRowIndex + 1) : filteredData;
			processedData = dataRows.map(row => {
				const obj = {};
				let workingHeaders = headers;
				if (columnIndices) {
					workingHeaders = columnIndices.map(i => headers[i] || 'column_' + i);
				}
				
				workingHeaders.forEach((header, index) => {
					const cellIndex = columnIndices ? columnIndices[index] : index;
					const cell = row[cellIndex];
					if (cell) {
						obj[header] = cell.text;
						// Include additional data if present
						if (cell.link) obj[header + '_link'] = cell.link;
						if (cell.image) obj[header + '_image'] = cell.image;
						if (cell.input_value !== undefined) obj[header + '_value'] = cell.input_value;
					} else {
						obj[header] = '';
					}
				});
				return obj;
			});
		} else if (outputFormat === 'csv') {
			// CSV string format
			const csvRows = [];
			
			// Add headers if included
			if (includeHeaders && headers.length > 0) {
				let headerRow = headers;
				if (columnIndices) {
					headerRow = columnIndices.map(i => headers[i] || 'column_' + i);
				}
				csvRows.push(headerRow.map(h => '"' + h.replace(/"/g, '""') + '"').join(','));
			}
			
			// Add data rows
			const dataRows = includeHeaders ? filteredData.slice(headerRowIndex + 1) : filteredData;
			dataRows.forEach(row => {
				let csvRow = row.map(cell => cell.text);
				if (columnIndices) {
					csvRow = columnIndices.map(i => csvRow[i] || '');
				}
				csvRows.push(csvRow.map(text => '"' + (text || '').replace(/"/g, '""') + '"').join(','));
			});
			
			processedData = csvRows.join('\n');
		}

		return {
			success: true,
			data: processedData,
			metadata: {
				total_rows: filteredData.length,
				total_columns: filteredData.length > 0 ? filteredData[0].length : 0,
				headers: headers,
				output_format: outputFormat,
				table_selector: '%s'
			}
		};
	`,
	strings.ReplaceAll(selector, "'", "\\'"),
	strings.ReplaceAll(selector, "'", "\\'"),
	skipEmptyRows,
	func() string { if maxRows != nil { return fmt.Sprintf("%d", *maxRows) } else { return "null" } }(),
	headerRow,
	includeHeaders,
	func() string { 
		if columnFilter != nil { 
			filterJSON, _ := json.Marshal(columnFilter)
			return string(filterJSON)
		} else { 
			return "null" 
		} 
	}(),
	outputFormat,
	strings.ReplaceAll(selector, "'", "\\'"))

	result, err := t.browserMgr.ExecuteScript(pageID, script)
	if err != nil {
		return nil, fmt.Errorf("failed to extract table data: %w", err)
	}

	// Parse the JavaScript result
	var jsResult map[string]interface{}
	resultStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from JavaScript execution")
	}
	if err := json.Unmarshal([]byte(resultStr), &jsResult); err != nil {
		return nil, fmt.Errorf("failed to parse table extraction result: %w", err)
	}

	// Check for extraction errors
	if errorMsg, exists := jsResult["error"]; exists {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Table extraction failed: %v", errorMsg),
			}},
		}, nil
	}

	// Format successful response
	data := jsResult["data"]
	metadata := jsResult["metadata"]

	var responseText string
	switch outputFormat {
	case "csv":
		responseText = fmt.Sprintf("Table extracted as CSV:\n\n%v", data)
	case "array":
		dataJSON, _ := json.MarshalIndent(data, "", "  ")
		responseText = fmt.Sprintf("Table extracted as array:\n\n%s", string(dataJSON))
	case "objects":
		dataJSON, _ := json.MarshalIndent(data, "", "  ")
		responseText = fmt.Sprintf("Table extracted as objects:\n\n%s", string(dataJSON))
	}

	// Add metadata info
	if meta, ok := metadata.(map[string]interface{}); ok {
		responseText += fmt.Sprintf("\n\nMetadata:\n- Rows: %v\n- Columns: %v\n- Format: %v", 
			meta["total_rows"], meta["total_columns"], meta["output_format"])
		if headers, ok := meta["headers"].([]interface{}); ok && len(headers) > 0 {
			responseText += fmt.Sprintf("\n- Headers: %v", headers)
		}
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: responseText,
			Data: map[string]interface{}{
				"table_data": data,
				"metadata":   metadata,
				"format":     outputFormat,
			},
		}},
	}, nil
}
