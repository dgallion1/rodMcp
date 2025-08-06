package webtools

import (
	"encoding/base64"
	"fmt"
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
