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
	debugpkg "runtime/debug"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Helper function to create a consistent error response when no pages are available
func createNoPagesErrorResponse(toolName string) *types.CallToolResponse {
	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("No browser pages are currently open. To use `%s`, you first need to:\n\n"+
				"1. Create a page: use `create_page` to make a new HTML page, or\n"+
				"2. Navigate to a URL: use `navigate_page` to load an existing website\n\n"+
				"Then you can interact with elements on the page.", toolName),
		}},
		IsError: true,
	}
}

// Helper function to execute tool operations with panic recovery
func executeWithPanicRecovery(toolName string, logger *logger.Logger, operation func() (*types.CallToolResponse, error)) (*types.CallToolResponse, error) {
	var result *types.CallToolResponse
	var err error
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := debugpkg.Stack()
				logger.WithComponent("tools").Error("Tool execution panic",
					zap.String("tool", toolName),
					zap.Any("panic", r),
					zap.String("stack", string(stackTrace)))
				err = fmt.Errorf("tool execution panicked: %v", r)
			}
		}()
		result, err = operation()
	}()
	
	return result, err
}

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
				"description": "Name of the HTML file to create (with or without .html extension). Examples: 'landing-page', 'contact-form.html', 'dashboard'",
				"examples":    []string{"landing-page", "contact-form.html", "dashboard", "app"},
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Page title that appears in browser tab and search results. Examples: 'Coffee Shop Landing', 'Contact Form', 'Dashboard'",
				"examples":    []string{"Coffee Shop Landing", "Contact Form", "My Dashboard", "Product Catalog"},
			},
			"html": map[string]interface{}{
				"type":        "string",
				"description": "HTML content for the page body. Use semantic HTML5 elements like <header>, <main>, <section>, <nav>. Examples: '<h1>Welcome</h1><p>Description</p>', '<form><input type=\"email\" required></form>'",
				"examples":    []string{"<h1>Welcome</h1><p>Sample content</p>", "<header><nav><a href=\"#\">Home</a></nav></header><main><h1>Title</h1></main>"},
			},
			"css": map[string]interface{}{
				"type":        "string",
				"description": "CSS styles to embed in the page. Can include responsive design, animations, custom properties. Examples: 'body{font-family:Arial;margin:0} .hero{background:#333;color:white}'",
				"examples":    []string{"body{font-family:Arial;margin:0}", ".btn{padding:10px;background:#007bff;color:white;border:none;border-radius:4px}"},
			},
			"javascript": map[string]interface{}{
				"type":        "string",
				"description": "JavaScript code for interactivity, event handlers, and dynamic behavior. Examples: 'document.querySelector(\".btn\").onclick = () => alert(\"Clicked!\");'",
				"examples":    []string{"console.log('Page loaded');", "document.querySelector('.btn').onclick = () => alert('Hello!');"},
			},
		},
		Required: []string{"filename", "title", "html"},
	}
}

func (t *CreatePageTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	return executeWithPanicRecovery(t.Name(), t.logger, func() (*types.CallToolResponse, error) {
		start := time.Now()
		defer func() {
			duration := time.Since(start).Milliseconds()
			t.logger.LogToolExecution(t.Name(), args, true, duration)
		}()

	filename, ok := args["filename"].(string)
	if !ok {
		return nil, fmt.Errorf("filename parameter must be a string")
	}
	
	if err := ValidateFilename(filename, t.Name()); err != nil {
		return nil, err
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
	})
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
				"description": "URL or file path to navigate to. Supports HTTP/HTTPS URLs, local files (file://), and relative paths. Examples: 'https://example.com', 'localhost:3000', './index.html', 'file:///path/to/file.html'",
				"examples":    []string{"https://example.com", "localhost:3000", "./index.html", "file:///home/user/page.html", "http://localhost:8080/dashboard"},
			},
		},
		Required: []string{"url"},
	}
}

func (t *NavigatePageTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	return executeWithPanicRecovery(t.Name(), t.logger, func() (*types.CallToolResponse, error) {
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
			resultChan <- result{nil, fmt.Errorf("url parameter must be a string")}
			return
		}
		
		if err := ValidateURL(url, "navigate_page"); err != nil {
			resultChan <- result{nil, err}
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
	})
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
	logger    *logger.Logger
	browser   *browser.Manager
	validator *PathValidator
}

func NewScreenshotTool(log *logger.Logger, browserMgr *browser.Manager) *ScreenshotTool {
	return &ScreenshotTool{
		logger:    log,
		browser:   browserMgr,
		validator: NewPathValidator(DefaultFileAccessConfig()),
	}
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
	return executeWithPanicRecovery(t.Name(), t.logger, func() (*types.CallToolResponse, error) {
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
		// Validate file path for security
		cleanPath := filepath.Clean(filename)
		if err := t.validator.ValidatePath(cleanPath, "write"); err != nil {
			t.logger.WithComponent("tools").Warn("Screenshot file access denied",
				zap.String("path", cleanPath),
				zap.Error(err))
			
			// Provide helpful error message with allowed paths
			allowedPaths := t.validator.GetAllowedPaths()
			errorMsg := fmt.Sprintf("Screenshot file access denied: %v", err)
			if len(allowedPaths) > 0 {
				errorMsg += fmt.Sprintf("\n\nAllowed paths:\n")
				for _, path := range allowedPaths {
					errorMsg += fmt.Sprintf("  • %s\n", path)
				}
			}
			
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: errorMsg,
				}},
				IsError: true,
			}, nil
		}

		// Validate file size
		if err := t.validator.ValidateFileSize(int64(len(screenshot))); err != nil {
			t.logger.WithComponent("tools").Warn("Screenshot file size validation failed",
				zap.String("path", cleanPath),
				zap.Int("size", len(screenshot)),
				zap.Error(err))
			
			sizeInKB := float64(len(screenshot)) / 1024
			maxSizeInKB := float64(10*1024*1024) / 1024  // Default 10MB limit
			errorMsg := fmt.Sprintf("Screenshot file size validation failed: %v\n\nScreenshot size: %.1f KB\nMaximum allowed: %.1f KB", 
				err, sizeInKB, maxSizeInKB)
			
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: errorMsg,
				}},
				IsError: true,
			}, nil
		}

		if err := os.WriteFile(cleanPath, screenshot, 0644); err != nil {
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
				Text: fmt.Sprintf("Screenshot saved to %s", cleanPath),
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
	})
}

// TakeElementScreenshotTool captures screenshots of specific elements
type TakeElementScreenshotTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
	validator  *PathValidator
}

func NewTakeElementScreenshotTool(log *logger.Logger, browserMgr *browser.Manager) *TakeElementScreenshotTool {
	return &TakeElementScreenshotTool{
		logger:     log,
		browserMgr: browserMgr,
		validator:  NewPathValidator(DefaultFileAccessConfig()),
	}
}

func (t *TakeElementScreenshotTool) Name() string {
	return "take_element_screenshot"
}

func (t *TakeElementScreenshotTool) Description() string {
	return "Take a screenshot of a specific element on the page"
}

func (t *TakeElementScreenshotTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for the element to screenshot",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID to screenshot from (optional, uses current page if not specified)",
			},
			"filename": map[string]interface{}{
				"type":        "string",
				"description": "Filename to save screenshot (optional)",
			},
			"padding": map[string]interface{}{
				"type":        "integer",
				"description": "Padding around the element in pixels (default: 10)",
				"default":     10,
				"minimum":     0,
				"maximum":     100,
			},
			"scroll_into_view": map[string]interface{}{
				"type":        "boolean",
				"description": "Scroll element into view before screenshot (default: true)",
				"default":     true,
			},
			"wait_for_element": map[string]interface{}{
				"type":        "boolean",
				"description": "Wait for element to be visible before screenshot (default: true)",
				"default":     true,
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum time to wait for element in seconds (default: 10)",
				"default":     10,
				"minimum":     1,
				"maximum":     60,
			},
		},
		Required: []string{"selector"},
	}
}

func (t *TakeElementScreenshotTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	return executeWithPanicRecovery(t.Name(), t.logger, func() (*types.CallToolResponse, error) {
		start := time.Now()
		defer func() {
			duration := time.Since(start).Milliseconds()
			t.logger.LogToolExecution(t.Name(), args, true, duration)
		}()

	// Add timeout protection
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Parse arguments
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector is required")
	}

	pageID, _ := args["page_id"].(string)
	filename, _ := args["filename"].(string)

	padding := 10
	if val, ok := args["padding"].(float64); ok {
		padding = int(val)
	}

	scrollIntoView := true
	if val, ok := args["scroll_into_view"].(bool); ok {
		scrollIntoView = val
	}

	waitForElement := true
	if val, ok := args["wait_for_element"].(bool); ok {
		waitForElement = val
	}

	timeout := 10
	if val, ok := args["timeout"].(float64); ok {
		timeout = int(val)
	}

	// Execute screenshot in goroutine with timeout
	resultChan := make(chan *types.CallToolResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := t.captureElementScreenshot(pageID, selector, filename, padding, scrollIntoView, waitForElement, timeout)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("element screenshot operation timed out after 60 seconds")
	case err := <-errorChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
	})
}

func (t *TakeElementScreenshotTool) captureElementScreenshot(pageID, selector, filename string, padding int, scrollIntoView, waitForElement bool, timeout int) (*types.CallToolResponse, error) {
	// First, find and prepare the element
	script := fmt.Sprintf(`
		// Find the target element
		const element = document.querySelector('%s');
		if (!element) {
			return { error: 'Element not found with selector: %s' };
		}

		// Wait for element to be visible if requested
		const waitForVisible = %t;
		const timeoutMs = %d * 1000;
		
		if (waitForVisible) {
			const startTime = Date.now();
			
			// Check if element is visible
			function isVisible(el) {
				const rect = el.getBoundingClientRect();
				const style = window.getComputedStyle(el);
				return rect.width > 0 && 
					   rect.height > 0 && 
					   style.display !== 'none' && 
					   style.visibility !== 'hidden' && 
					   style.opacity !== '0';
			}
			
			// Wait for visibility with timeout
			while (!isVisible(element)) {
				if (Date.now() - startTime > timeoutMs) {
					return { error: 'Element not visible within timeout period' };
				}
				// Small delay to prevent busy waiting
				await new Promise(resolve => setTimeout(resolve, 100));
			}
		}

		// Scroll element into view if requested
		const shouldScroll = %t;
		if (shouldScroll) {
			element.scrollIntoView({ 
				behavior: 'auto', 
				block: 'center', 
				inline: 'center' 
			});
			// Wait a moment for scroll to complete
			await new Promise(resolve => setTimeout(resolve, 200));
		}

		// Get element position and dimensions
		const rect = element.getBoundingClientRect();
		const padding = %d;
		
		// Calculate screenshot bounds with padding
		const bounds = {
			x: Math.max(0, rect.left - padding),
			y: Math.max(0, rect.top - padding),
			width: rect.width + (padding * 2),
			height: rect.height + (padding * 2)
		};
		
		// Ensure bounds don't exceed viewport
		bounds.width = Math.min(bounds.width, window.innerWidth - bounds.x);
		bounds.height = Math.min(bounds.height, window.innerHeight - bounds.y);

		return {
			success: true,
			bounds: bounds,
			element_info: {
				tag_name: element.tagName,
				id: element.id,
				class_name: element.className,
				text_content: element.textContent?.slice(0, 100) // First 100 chars
			}
		};
	`,
	strings.ReplaceAll(selector, "'", "\\'"),
	strings.ReplaceAll(selector, "'", "\\'"),
	waitForElement,
	timeout,
	scrollIntoView,
	padding)

	result, err := t.browserMgr.ExecuteScript(pageID, script)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare element for screenshot: %w", err)
	}

	// Parse the JavaScript result
	var jsResult map[string]interface{}
	resultStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from JavaScript execution")
	}
	if err := json.Unmarshal([]byte(resultStr), &jsResult); err != nil {
		return nil, fmt.Errorf("failed to parse element preparation result: %w", err)
	}

	// Check for errors
	if errorMsg, exists := jsResult["error"]; exists {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Element screenshot failed: %v", errorMsg),
			}},
		}, nil
	}

	// Extract bounds information
	boundsData, ok := jsResult["bounds"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid bounds data returned from JavaScript")
	}

	// Get element info for metadata
	elementInfo, _ := jsResult["element_info"].(map[string]interface{})

	// Take the full page screenshot first
	fullScreenshot, err := t.browserMgr.Screenshot(pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to take full page screenshot: %w", err)
	}

	// For now, we'll return the full screenshot with bounds info
	// TODO: In a future enhancement, we could crop the image to just the element bounds
	
	// If filename is provided, save the screenshot
	if filename != "" {
		// Validate file path for security
		cleanPath := filepath.Clean(filename)
		if err := t.validator.ValidatePath(cleanPath, "write"); err != nil {
			t.logger.WithComponent("tools").Warn("Element screenshot file access denied",
				zap.String("path", cleanPath),
				zap.Error(err))
			
			// Provide helpful error message with allowed paths
			allowedPaths := t.validator.GetAllowedPaths()
			errorMsg := fmt.Sprintf("Element screenshot file access denied: %v", err)
			if len(allowedPaths) > 0 {
				errorMsg += fmt.Sprintf("\n\nAllowed paths:\n")
				for _, path := range allowedPaths {
					errorMsg += fmt.Sprintf("  • %s\n", path)
				}
			}
			
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: errorMsg,
				}},
				IsError: true,
			}, nil
		}

		// Validate file size
		if err := t.validator.ValidateFileSize(int64(len(fullScreenshot))); err != nil {
			t.logger.WithComponent("tools").Warn("Element screenshot file size validation failed",
				zap.String("path", cleanPath),
				zap.Int("size", len(fullScreenshot)),
				zap.Error(err))
			
			sizeInKB := float64(len(fullScreenshot)) / 1024
			maxSizeInKB := float64(10*1024*1024) / 1024  // Default 10MB limit
			errorMsg := fmt.Sprintf("Element screenshot file size validation failed: %v\n\nScreenshot size: %.1f KB\nMaximum allowed: %.1f KB", 
				err, sizeInKB, maxSizeInKB)
			
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: errorMsg,
				}},
				IsError: true,
			}, nil
		}

		if err := os.WriteFile(cleanPath, fullScreenshot, 0644); err != nil {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Failed to save element screenshot: %v", err),
				}},
				IsError: true,
			}, nil
		}

		responseText := fmt.Sprintf("Element screenshot saved to %s", cleanPath)
		if elementInfo != nil {
			responseText += fmt.Sprintf("\n\nElement details:\n- Tag: %v\n- ID: %v\n- Classes: %v",
				elementInfo["tag_name"], elementInfo["id"], elementInfo["class_name"])
			if textContent, ok := elementInfo["text_content"].(string); ok && textContent != "" {
				responseText += fmt.Sprintf("\n- Text: %s", textContent)
			}
		}
		if boundsData != nil {
			responseText += fmt.Sprintf("\n\nScreenshot bounds:\n- X: %.0f, Y: %.0f\n- Width: %.0f, Height: %.0f",
				boundsData["x"], boundsData["y"], boundsData["width"], boundsData["height"])
		}

		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: responseText,
				Data: map[string]interface{}{
					"filename": cleanPath,
					"bounds":   boundsData,
					"element":  elementInfo,
				},
			}},
		}, nil
	}

	// Return base64 encoded image with element metadata
	encoded := base64.StdEncoding.EncodeToString(fullScreenshot)
	
	responseText := "Element screenshot captured"
	if elementInfo != nil {
		responseText += fmt.Sprintf("\n\nElement: %v", elementInfo["tag_name"])
		if id, ok := elementInfo["id"].(string); ok && id != "" {
			responseText += fmt.Sprintf("#%s", id)
		}
		if className, ok := elementInfo["class_name"].(string); ok && className != "" {
			responseText += fmt.Sprintf(".%s", strings.ReplaceAll(className, " ", "."))
		}
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type:     "image",
			Data:     encoded,
			MimeType: "image/png",
		}},
	}, nil
}

// KeyboardShortcutTool sends keyboard combinations and special keys
type KeyboardShortcutTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewKeyboardShortcutTool(log *logger.Logger, browserMgr *browser.Manager) *KeyboardShortcutTool {
	return &KeyboardShortcutTool{
		logger:     log,
		browserMgr: browserMgr,
	}
}

func (t *KeyboardShortcutTool) Name() string {
	return "keyboard_shortcuts"
}

func (t *KeyboardShortcutTool) Description() string {
	return "Send keyboard combinations and special keys like Ctrl+C/V, F5, Tab, Enter, etc."
}

func (t *KeyboardShortcutTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"keys": map[string]interface{}{
				"type":        "string",
				"description": "Key combination to send (e.g., 'Ctrl+C', 'F5', 'Tab', 'Enter', 'Alt+Tab')",
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID to send keys to (optional, uses current page if not specified)",
			},
			"element_selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for element to focus before sending keys (optional)",
			},
			"repeat": map[string]interface{}{
				"type":        "integer",
				"description": "Number of times to repeat the key combination (default: 1)",
				"default":     1,
				"minimum":     1,
				"maximum":     10,
			},
			"delay": map[string]interface{}{
				"type":        "integer",
				"description": "Delay between key repeats in milliseconds (default: 100)",
				"default":     100,
				"minimum":     0,
				"maximum":     5000,
			},
		},
		Required: []string{"keys"},
	}
}

func (t *KeyboardShortcutTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	// Add timeout protection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse arguments
	keys, ok := args["keys"].(string)
	if !ok {
		return nil, fmt.Errorf("keys is required")
	}

	pageID, _ := args["page_id"].(string)
	elementSelector, _ := args["element_selector"].(string)

	repeat := 1
	if val, ok := args["repeat"].(float64); ok {
		repeat = int(val)
	}

	delay := 100
	if val, ok := args["delay"].(float64); ok {
		delay = int(val)
	}

	// Execute keyboard shortcut in goroutine with timeout
	resultChan := make(chan *types.CallToolResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := t.sendKeyboardShortcut(pageID, elementSelector, keys, repeat, delay)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("keyboard shortcut operation timed out after 30 seconds")
	case err := <-errorChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}

func (t *KeyboardShortcutTool) sendKeyboardShortcut(pageID, elementSelector, keys string, repeat, delay int) (*types.CallToolResponse, error) {
	// Parse the key combination
	keyConfig, err := t.parseKeyCombination(keys)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key combination '%s': %w", keys, err)
	}

	// Build JavaScript for sending keyboard events
	script := fmt.Sprintf(`
		// Parse key configuration
		const keyConfig = %s;
		const elementSelector = '%s';
		const repeat = %d;
		const delay = %d;

		// Focus on specific element if provided
		let targetElement = document.activeElement;
		if (elementSelector) {
			const element = document.querySelector(elementSelector);
			if (element) {
				element.focus();
				targetElement = element;
			} else {
				return { error: 'Element not found with selector: ' + elementSelector };
			}
		}

		// Helper function to create and dispatch keyboard event
		function dispatchKeyEvent(eventType, keyConfig, target) {
			const event = new KeyboardEvent(eventType, {
				key: keyConfig.key,
				code: keyConfig.code,
				keyCode: keyConfig.keyCode,
				which: keyConfig.keyCode,
				ctrlKey: keyConfig.ctrlKey,
				altKey: keyConfig.altKey,
				shiftKey: keyConfig.shiftKey,
				metaKey: keyConfig.metaKey,
				bubbles: true,
				cancelable: true
			});
			
			target.dispatchEvent(event);
			return event;
		}

		// Send the key combination
		const results = [];
		for (let i = 0; i < repeat; i++) {
			// Send keydown event
			const keydownEvent = dispatchKeyEvent('keydown', keyConfig, targetElement);
			
			// Send keypress event (for printable characters)
			if (keyConfig.isPrintable) {
				dispatchKeyEvent('keypress', keyConfig, targetElement);
			}
			
			// Send keyup event
			const keyupEvent = dispatchKeyEvent('keyup', keyConfig, targetElement);
			
			results.push({
				iteration: i + 1,
				keydown_prevented: keydownEvent.defaultPrevented,
				keyup_prevented: keyupEvent.defaultPrevented
			});

			// Add delay between repeats (except for last iteration)
			if (i < repeat - 1 && delay > 0) {
				await new Promise(resolve => setTimeout(resolve, delay));
			}
		}

		return {
			success: true,
			keys_sent: '%s',
			target_element: targetElement.tagName + (targetElement.id ? '#' + targetElement.id : '') + (targetElement.className ? '.' + targetElement.className.split(' ').join('.') : ''),
			repeat_count: repeat,
			results: results,
			key_info: keyConfig
		};
	`,
	keyConfig,
	strings.ReplaceAll(elementSelector, "'", "\\'"),
	repeat,
	delay,
	strings.ReplaceAll(keys, "'", "\\'"))

	result, err := t.browserMgr.ExecuteScript(pageID, script)
	if err != nil {
		return nil, fmt.Errorf("failed to send keyboard shortcut: %w", err)
	}

	// Parse the JavaScript result
	var jsResult map[string]interface{}
	resultStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from JavaScript execution")
	}
	if err := json.Unmarshal([]byte(resultStr), &jsResult); err != nil {
		return nil, fmt.Errorf("failed to parse keyboard shortcut result: %w", err)
	}

	// Check for errors
	if errorMsg, exists := jsResult["error"]; exists {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Keyboard shortcut failed: %v", errorMsg),
			}},
		}, nil
	}

	// Format successful response
	responseText := fmt.Sprintf("Successfully sent keyboard shortcut: %s", keys)
	if targetElement, ok := jsResult["target_element"].(string); ok {
		responseText += fmt.Sprintf("\nTarget: %s", targetElement)
	}
	if repeat > 1 {
		responseText += fmt.Sprintf("\nRepeated: %d times", repeat)
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: responseText,
			Data: jsResult,
		}},
	}, nil
}

func (t *KeyboardShortcutTool) parseKeyCombination(keys string) (string, error) {
	// Common key mappings
	keyMappings := map[string]map[string]interface{}{
		// Navigation keys
		"Tab":       {"key": "Tab", "code": "Tab", "keyCode": 9, "isPrintable": false},
		"Shift+Tab": {"key": "Tab", "code": "Tab", "keyCode": 9, "shiftKey": true, "isPrintable": false},
		"Enter":     {"key": "Enter", "code": "Enter", "keyCode": 13, "isPrintable": false},
		"Escape":    {"key": "Escape", "code": "Escape", "keyCode": 27, "isPrintable": false},
		"Backspace": {"key": "Backspace", "code": "Backspace", "keyCode": 8, "isPrintable": false},
		"Delete":    {"key": "Delete", "code": "Delete", "keyCode": 46, "isPrintable": false},

		// Arrow keys
		"ArrowUp":    {"key": "ArrowUp", "code": "ArrowUp", "keyCode": 38, "isPrintable": false},
		"ArrowDown":  {"key": "ArrowDown", "code": "ArrowDown", "keyCode": 40, "isPrintable": false},
		"ArrowLeft":  {"key": "ArrowLeft", "code": "ArrowLeft", "keyCode": 37, "isPrintable": false},
		"ArrowRight": {"key": "ArrowRight", "code": "ArrowRight", "keyCode": 39, "isPrintable": false},

		// Page navigation
		"PageUp":   {"key": "PageUp", "code": "PageUp", "keyCode": 33, "isPrintable": false},
		"PageDown": {"key": "PageDown", "code": "PageDown", "keyCode": 34, "isPrintable": false},
		"Home":     {"key": "Home", "code": "Home", "keyCode": 36, "isPrintable": false},
		"End":      {"key": "End", "code": "End", "keyCode": 35, "isPrintable": false},

		// Function keys
		"F1":  {"key": "F1", "code": "F1", "keyCode": 112, "isPrintable": false},
		"F2":  {"key": "F2", "code": "F2", "keyCode": 113, "isPrintable": false},
		"F3":  {"key": "F3", "code": "F3", "keyCode": 114, "isPrintable": false},
		"F4":  {"key": "F4", "code": "F4", "keyCode": 115, "isPrintable": false},
		"F5":  {"key": "F5", "code": "F5", "keyCode": 116, "isPrintable": false},
		"F6":  {"key": "F6", "code": "F6", "keyCode": 117, "isPrintable": false},
		"F7":  {"key": "F7", "code": "F7", "keyCode": 118, "isPrintable": false},
		"F8":  {"key": "F8", "code": "F8", "keyCode": 119, "isPrintable": false},
		"F9":  {"key": "F9", "code": "F9", "keyCode": 120, "isPrintable": false},
		"F10": {"key": "F10", "code": "F10", "keyCode": 121, "isPrintable": false},
		"F11": {"key": "F11", "code": "F11", "keyCode": 122, "isPrintable": false},
		"F12": {"key": "F12", "code": "F12", "keyCode": 123, "isPrintable": false},

		// Common shortcuts with Ctrl
		"Ctrl+A": {"key": "a", "code": "KeyA", "keyCode": 65, "ctrlKey": true, "isPrintable": false},
		"Ctrl+C": {"key": "c", "code": "KeyC", "keyCode": 67, "ctrlKey": true, "isPrintable": false},
		"Ctrl+V": {"key": "v", "code": "KeyV", "keyCode": 86, "ctrlKey": true, "isPrintable": false},
		"Ctrl+X": {"key": "x", "code": "KeyX", "keyCode": 88, "ctrlKey": true, "isPrintable": false},
		"Ctrl+Z": {"key": "z", "code": "KeyZ", "keyCode": 90, "ctrlKey": true, "isPrintable": false},
		"Ctrl+Y": {"key": "y", "code": "KeyY", "keyCode": 89, "ctrlKey": true, "isPrintable": false},
		"Ctrl+S": {"key": "s", "code": "KeyS", "keyCode": 83, "ctrlKey": true, "isPrintable": false},
		"Ctrl+O": {"key": "o", "code": "KeyO", "keyCode": 79, "ctrlKey": true, "isPrintable": false},
		"Ctrl+N": {"key": "n", "code": "KeyN", "keyCode": 78, "ctrlKey": true, "isPrintable": false},
		"Ctrl+W": {"key": "w", "code": "KeyW", "keyCode": 87, "ctrlKey": true, "isPrintable": false},
		"Ctrl+F": {"key": "f", "code": "KeyF", "keyCode": 70, "ctrlKey": true, "isPrintable": false},
		"Ctrl+R": {"key": "r", "code": "KeyR", "keyCode": 82, "ctrlKey": true, "isPrintable": false},

		// Common shortcuts with Alt
		"Alt+Tab":  {"key": "Tab", "code": "Tab", "keyCode": 9, "altKey": true, "isPrintable": false},
		"Alt+F4":   {"key": "F4", "code": "F4", "keyCode": 115, "altKey": true, "isPrintable": false},
		"Alt+Left": {"key": "ArrowLeft", "code": "ArrowLeft", "keyCode": 37, "altKey": true, "isPrintable": false},
		"Alt+Right": {"key": "ArrowRight", "code": "ArrowRight", "keyCode": 39, "altKey": true, "isPrintable": false},

		// Common shortcuts with Shift
		"Shift+F10": {"key": "F10", "code": "F10", "keyCode": 121, "shiftKey": true, "isPrintable": false},

		// Space
		"Space": {"key": " ", "code": "Space", "keyCode": 32, "isPrintable": true},
	}

	// Check if the key combination exists in our mappings
	if keyData, exists := keyMappings[keys]; exists {
		// Convert to JSON string for JavaScript
		jsonBytes, err := json.Marshal(keyData)
		if err != nil {
			return "", fmt.Errorf("failed to marshal key data: %w", err)
		}
		return string(jsonBytes), nil
	}

	return "", fmt.Errorf("unsupported key combination: %s. Supported keys include: Tab, Enter, F5, Ctrl+C, Ctrl+V, Alt+Tab, etc.", keys)
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
	return executeWithPanicRecovery(t.Name(), t.logger, func() (*types.CallToolResponse, error) {
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
				resultChan <- result{createNoPagesErrorResponse("execute_script"), nil}
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
	})
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
	return executeWithPanicRecovery(t.Name(), t.logger, func() (*types.CallToolResponse, error) {
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
	})
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
	logger    *logger.Logger
	validator *PathValidator
}

func NewReadFileTool(log *logger.Logger, validator *PathValidator) *ReadFileTool {
	if validator == nil {
		validator = NewPathValidator(DefaultFileAccessConfig())
	}
	return &ReadFileTool{
		logger:    log,
		validator: validator,
	}
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
	
	// Validate path access permissions
	if err := t.validator.ValidatePath(cleanPath, "read"); err != nil {
		t.logger.WithComponent("tools").Warn("File access denied",
			zap.String("path", cleanPath),
			zap.Error(err))
		return nil, fmt.Errorf("file access denied: %w", err)
	}
	
	// Check file size before reading
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		t.logger.WithComponent("tools").Error("Failed to get file info",
			zap.String("path", cleanPath),
			zap.Error(err))
		return nil, fmt.Errorf("failed to access file %s: %w", cleanPath, err)
	}
	
	// Use the configured max file size from the validator
	maxSize := t.validator.config.MaxFileSize
	if fileInfo.Size() > maxSize {
		return nil, fmt.Errorf("file %s is too large (%d bytes) - maximum allowed size is %d bytes", 
			cleanPath, fileInfo.Size(), maxSize)
	}
	
	// Read the file with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	type readResult struct {
		content []byte
		err     error
	}
	resultChan := make(chan readResult, 1)
	
	go func() {
		content, err := os.ReadFile(cleanPath)
		resultChan <- readResult{content, err}
	}()
	
	var content []byte
	select {
	case result := <-resultChan:
		content, err = result.content, result.err
	case <-ctx.Done():
		return nil, fmt.Errorf("file read timed out after 30 seconds: %s", cleanPath)
	}
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
	logger    *logger.Logger
	validator *PathValidator
}

func NewWriteFileTool(log *logger.Logger, validator *PathValidator) *WriteFileTool {
	if validator == nil {
		validator = NewPathValidator(DefaultFileAccessConfig())
	}
	return &WriteFileTool{
		logger:    log,
		validator: validator,
	}
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
	
	// Validate path access permissions
	if err := t.validator.ValidatePath(cleanPath, "write"); err != nil {
		t.logger.WithComponent("tools").Warn("File access denied",
			zap.String("path", cleanPath),
			zap.Error(err))
		return nil, fmt.Errorf("file access denied: %w", err)
	}
	
	// Validate file size
	if err := t.validator.ValidateFileSize(int64(len(content))); err != nil {
		t.logger.WithComponent("tools").Warn("File size validation failed",
			zap.String("path", cleanPath),
			zap.Int("size_bytes", len(content)),
			zap.Error(err))
		return nil, fmt.Errorf("file size validation failed: %w", err)
	}
	
	// Create parent directories if requested
	if createDirs {
		dir := filepath.Dir(cleanPath)
		// Also validate that the parent directory is allowed
		if err := t.validator.ValidatePath(dir, "write"); err != nil {
			t.logger.WithComponent("tools").Warn("Parent directory access denied",
				zap.String("dir", dir),
				zap.Error(err))
			return nil, fmt.Errorf("parent directory access denied: %w", err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directories for %s: %w", cleanPath, err)
		}
	}

	// Check content size before writing
	contentSize := int64(len(content))
	maxSize := t.validator.config.MaxFileSize
	if contentSize > maxSize {
		return nil, fmt.Errorf("content is too large (%d bytes) - maximum allowed size is %d bytes", 
			contentSize, maxSize)
	}
	
	// Write the file with timeout context
	writeCtx, writeCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer writeCancel()
	
	type writeResult struct {
		err error
	}
	writeResultChan := make(chan writeResult, 1)
	
	go func() {
		err := os.WriteFile(cleanPath, []byte(content), 0644)
		writeResultChan <- writeResult{err}
	}()
	
	var writeErr error
	select {
	case result := <-writeResultChan:
		writeErr = result.err
	case <-writeCtx.Done():
		return nil, fmt.Errorf("file write timed out after 30 seconds: %s", cleanPath)
	}
	if writeErr != nil {
		t.logger.WithComponent("tools").Error("Failed to write file",
			zap.String("path", cleanPath),
			zap.Error(writeErr))
		return nil, fmt.Errorf("failed to write file %s: %w", cleanPath, writeErr)
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
	logger    *logger.Logger
	validator *PathValidator
}

func NewListDirectoryTool(log *logger.Logger, validator *PathValidator) *ListDirectoryTool {
	if validator == nil {
		validator = NewPathValidator(DefaultFileAccessConfig())
	}
	return &ListDirectoryTool{
		logger:    log,
		validator: validator,
	}
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
	
	// Validate path access permissions
	if err := t.validator.ValidatePath(cleanPath, "read"); err != nil {
		t.logger.WithComponent("tools").Warn("Directory access denied",
			zap.String("path", cleanPath),
			zap.Error(err))
		return nil, fmt.Errorf("directory access denied: %w", err)
	}
	
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
				"description": "CSS selector or XPath (prefix with //) for the element to click. CSS selectors: #id (ID), .class (class), tag (element), [attr] (attribute). XPath: //tag[@attr='value'] or //text()='content'. Examples: '#submit-btn', '.nav-link', 'button[type=\"submit\"]', '//button[text()=\"Login\"]'",
				"examples":    []string{"#submit-button", ".btn-primary", "button[type='submit']", "input[value='Submit']", "//button[contains(text(), 'Login')]", ".modal .close-btn"},
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID to click on (optional, uses current active page if not specified). Get page IDs from switch_tab list action",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum seconds to wait for element to become clickable. Use 2-5s for static elements, 5-10s for dynamic content, 10-30s for heavy AJAX (default: 10)",
				"default":     10,
				"minimum":     1,
				"maximum":     60,
				"examples":    []interface{}{5, 10, 15, 30},
			},
		},
		Required: []string{"selector"},
	}
}

func (t *ClickElementTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	return executeWithPanicRecovery(t.Name(), t.logger, func() (*types.CallToolResponse, error) {
		start := time.Now()
	
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector parameter must be a string")
	}
	
	if err := ValidateSelector(selector, t.Name()); err != nil {
		return nil, err
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
			return createNoPagesErrorResponse("click_element"), nil
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
	})
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
				"description": "CSS selector for the input element (input, textarea, contenteditable). Examples: 'input[name=\"email\"]', '#password', '.search-box', 'textarea[placeholder=\"Message\"]'",
				"examples":    []string{"input[name='email']", "#username", ".search-input", "textarea[placeholder='Message']", "input[type='password']"},
			},
			"text": map[string]interface{}{
				"type":        "string",
				"description": "Text content to type into the element. Can include newlines (\\n) for textareas and special characters. Examples: 'user@example.com', 'Hello\\nWorld', '123-456-7890'",
				"examples":    []string{"user@example.com", "Hello World", "123-456-7890", "Multi-line\\ntext content", "Special chars: !@#$%"},
			},
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "Page ID to type in (optional, uses current active page if not specified). Get page IDs from switch_tab list action",
			},
			"clear": map[string]interface{}{
				"type":        "boolean",
				"description": "Clear existing content before typing. Set to false to append text (default: true)",
				"default":     true,
				"examples":    []interface{}{true, false},
			},
		},
		Required: []string{"selector", "text"},
	}
}

func (t *TypeTextTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	
	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector parameter must be a string")
	}
	
	if err := ValidateSelector(selector, t.Name()); err != nil {
		return nil, err
	}

	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text parameter must be a string")
	}
	
	if err := ValidateText(text, t.Name(), false); err != nil {
		return nil, err
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
			return createNoPagesErrorResponse("type_text"), nil
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
			return createNoPagesErrorResponse("wait_for_element"), nil
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
			return createNoPagesErrorResponse("get_element_text"), nil
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

// SwitchTabTool switches between browser tabs for multi-tab workflows
type SwitchTabTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewSwitchTabTool(log *logger.Logger, browserMgr *browser.Manager) *SwitchTabTool {
	return &SwitchTabTool{
		logger:     log,
		browserMgr: browserMgr,
	}
}

func (t *SwitchTabTool) Name() string {
	return "switch_tab"
}

func (t *SwitchTabTool) Description() string {
	return "Switch between browser tabs for multi-tab workflow automation"
}

func (t *SwitchTabTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Tab action: 'create', 'switch', 'close', 'list', 'close_all'",
				"enum":        []string{"create", "switch", "close", "list", "close_all"},
				"default":     "switch",
			},
			"target": map[string]interface{}{
				"type":        "string",
				"description": "Target for action: page_id for switch/close, URL for create, or 'current' for current tab",
			},
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to load when creating a new tab",
			},
			"switch_to": map[string]interface{}{
				"type":        "string",
				"description": "Switch method: 'next', 'previous', 'first', 'last', or page_id",
				"enum":        []string{"next", "previous", "first", "last"},
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds for tab operations (default: 10)",
				"default":     10,
				"minimum":     1,
				"maximum":     60,
			},
		},
	}
}

func (t *SwitchTabTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	action := "switch"
	if val, ok := args["action"].(string); ok {
		action = val
	}

	timeout := 10
	if val, ok := args["timeout"].(float64); ok {
		timeout = int(val)
	}

	switch action {
	case "create":
		return t.createTab(args, timeout)
	case "switch":
		return t.switchTab(args, timeout)
	case "close":
		return t.closeTab(args, timeout)
	case "list":
		return t.listTabs(timeout)
	case "close_all":
		return t.closeAllTabs(timeout)
	default:
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Unknown action: %s. Use 'create', 'switch', 'close', 'list', or 'close_all'", action),
			}},
			IsError: true,
		}, nil
	}
}

func (t *SwitchTabTool) createTab(args map[string]interface{}, timeout int) (*types.CallToolResponse, error) {
	url, hasURL := args["url"].(string)
	if !hasURL || url == "" {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "URL is required when creating a new tab",
			}},
			IsError: true,
		}, nil
	}

	// Create new page (tab)
	page, pageID, err := t.browserMgr.NewPage(url)
	if err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to create new tab: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Wait for page to load with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	loadScript := `
		new Promise((resolve) => {
			if (document.readyState === 'complete') {
				resolve({
					ready: true,
					title: document.title,
					url: window.location.href
				});
			} else {
				window.addEventListener('load', () => {
					resolve({
						ready: true,
						title: document.title,
						url: window.location.href
					});
				});
			}
		});
	`

	done := make(chan bool, 1)
	var pageInfo map[string]interface{}

	go func() {
		if data, err := t.browserMgr.ExecuteScript(pageID, loadScript); err == nil {
			if info, ok := data.(map[string]interface{}); ok {
				pageInfo = info
			}
		}
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-ctx.Done():
		t.logger.Info("Tab creation timed out, but tab was created")
	}

	title := "New Tab"
	if pageInfo != nil {
		if t, ok := pageInfo["title"].(string); ok && t != "" {
			title = t
		}
	}

	// Switch to the new tab
	if _, err = page.Activate(); err != nil {
		t.logger.Info("Failed to activate new tab, but tab was created")
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Created and switched to new tab: %s", title),
			Data: map[string]interface{}{
				"page_id": pageID,
				"url":     url,
				"title":   title,
				"action":  "create",
			},
		}},
	}, nil
}

func (t *SwitchTabTool) switchTab(args map[string]interface{}, timeout int) (*types.CallToolResponse, error) {
	// Get all pages
	pages := t.browserMgr.GetAllPages()
	if len(pages) == 0 {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "No tabs available to switch to",
			}},
			IsError: true,
		}, nil
	}

	var targetPage *browser.PageInfo
	var targetID string

	// Check if specific page_id is provided in target
	if target, ok := args["target"].(string); ok && target != "" {
		targetID = target
		for _, page := range pages {
			if page.PageID == targetID {
				targetPage = &page
				break
			}
		}
		if targetPage == nil {
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Tab with page_id '%s' not found", targetID),
				}},
				IsError: true,
			}, nil
		}
	} else if switchTo, ok := args["switch_to"].(string); ok {
		// Handle directional switching
		currentPageID := t.browserMgr.GetCurrentPageID()
		currentIndex := -1
		
		// Find current page index
		for i, page := range pages {
			if page.PageID == currentPageID {
				currentIndex = i
				break
			}
		}

		switch switchTo {
		case "next":
			nextIndex := (currentIndex + 1) % len(pages)
			targetPage = &pages[nextIndex]
		case "previous":
			prevIndex := currentIndex - 1
			if prevIndex < 0 {
				prevIndex = len(pages) - 1
			}
			targetPage = &pages[prevIndex]
		case "first":
			targetPage = &pages[0]
		case "last":
			targetPage = &pages[len(pages)-1]
		default:
			return &types.CallToolResponse{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Unknown switch_to value: %s. Use 'next', 'previous', 'first', or 'last'", switchTo),
				}},
				IsError: true,
			}, nil
		}
		targetID = targetPage.PageID
	} else {
		// Default to next tab
		currentPageID := t.browserMgr.GetCurrentPageID()
		currentIndex := -1
		
		for i, page := range pages {
			if page.PageID == currentPageID {
				currentIndex = i
				break
			}
		}
		
		nextIndex := (currentIndex + 1) % len(pages)
		targetPage = &pages[nextIndex]
		targetID = targetPage.PageID
	}

	// Switch to target tab
	if err := t.browserMgr.SwitchToPage(targetID); err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to switch to tab: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Switched to tab: %s (%s)", targetPage.Title, targetPage.URL),
			Data: map[string]interface{}{
				"page_id":      targetID,
				"url":          targetPage.URL,
				"title":        targetPage.Title,
				"action":       "switch",
				"total_tabs":   len(pages),
			},
		}},
	}, nil
}

func (t *SwitchTabTool) closeTab(args map[string]interface{}, timeout int) (*types.CallToolResponse, error) {
	pages := t.browserMgr.GetAllPages()
	if len(pages) <= 1 {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "Cannot close the last remaining tab",
			}},
			IsError: true,
		}, nil
	}

	var targetID string
	if target, ok := args["target"].(string); ok && target != "" {
		if target == "current" {
			targetID = t.browserMgr.GetCurrentPageID()
		} else {
			targetID = target
		}
	} else {
		targetID = t.browserMgr.GetCurrentPageID()
	}

	// Find the tab to close
	var targetPage *browser.PageInfo
	for _, page := range pages {
		if page.PageID == targetID {
			targetPage = &page
			break
		}
	}

	if targetPage == nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Tab with page_id '%s' not found", targetID),
			}},
			IsError: true,
		}, nil
	}

	// If closing current tab, switch to another first
	currentPageID := t.browserMgr.GetCurrentPageID()
	if targetID == currentPageID {
		// Find next tab to switch to
		var nextPageID string
		for _, page := range pages {
			if page.PageID != targetID {
				nextPageID = page.PageID
				break
			}
		}
		
		if nextPageID != "" {
			if err := t.browserMgr.SwitchToPage(nextPageID); err != nil {
				t.logger.Info("Failed to switch before closing, continuing with close")
			}
		}
	}

	// Close the tab
	if err := t.browserMgr.ClosePage(targetID); err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to close tab: %v", err),
			}},
			IsError: true,
		}, nil
	}

	remainingTabs := len(pages) - 1
	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Closed tab: %s (%d tabs remaining)", targetPage.Title, remainingTabs),
			Data: map[string]interface{}{
				"closed_page_id": targetID,
				"closed_title":   targetPage.Title,
				"closed_url":     targetPage.URL,
				"action":         "close",
				"remaining_tabs": remainingTabs,
			},
		}},
	}, nil
}

func (t *SwitchTabTool) listTabs(timeout int) (*types.CallToolResponse, error) {
	pages := t.browserMgr.GetAllPages()
	currentPageID := t.browserMgr.GetCurrentPageID()

	if len(pages) == 0 {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "No tabs are currently open",
			}},
		}, nil
	}

	var tabList []string
	tabData := make([]map[string]interface{}, 0, len(pages))

	tabList = append(tabList, fmt.Sprintf("Open tabs (%d total):", len(pages)))
	tabList = append(tabList, "")

	for i, page := range pages {
		status := ""
		if page.PageID == currentPageID {
			status = " [CURRENT]"
		}
		
		title := page.Title
		if title == "" {
			title = "Untitled"
		}

		tabList = append(tabList, fmt.Sprintf("%d. %s%s", i+1, title, status))
		tabList = append(tabList, fmt.Sprintf("   URL: %s", page.URL))
		tabList = append(tabList, fmt.Sprintf("   Page ID: %s", page.PageID))
		if i < len(pages)-1 {
			tabList = append(tabList, "")
		}

		tabData = append(tabData, map[string]interface{}{
			"index":      i + 1,
			"page_id":    page.PageID,
			"title":      title,
			"url":        page.URL,
			"is_current": page.PageID == currentPageID,
		})
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: strings.Join(tabList, "\n"),
			Data: map[string]interface{}{
				"tabs":        tabData,
				"total_tabs":  len(pages),
				"current_id":  currentPageID,
				"action":      "list",
			},
		}},
	}, nil
}

func (t *SwitchTabTool) closeAllTabs(timeout int) (*types.CallToolResponse, error) {
	pages := t.browserMgr.GetAllPages()
	if len(pages) <= 1 {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: "Only one tab open, cannot close all tabs",
			}},
			IsError: true,
		}, nil
	}

	currentPageID := t.browserMgr.GetCurrentPageID()
	var closedCount int
	var errors []string

	// Close all tabs except current
	for _, page := range pages {
		if page.PageID != currentPageID {
			if err := t.browserMgr.ClosePage(page.PageID); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to close %s: %v", page.Title, err))
			} else {
				closedCount++
			}
		}
	}

	responseText := fmt.Sprintf("Closed %d tabs, keeping current tab open", closedCount)
	if len(errors) > 0 {
		responseText += fmt.Sprintf("\n\nErrors encountered:\n%s", strings.Join(errors, "\n"))
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: responseText,
			Data: map[string]interface{}{
				"closed_count": closedCount,
				"errors":       errors,
				"action":       "close_all",
				"remaining_tabs": 1,
			},
		}},
		IsError: len(errors) > 0,
	}, nil
}
