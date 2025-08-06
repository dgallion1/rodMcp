package main

import (
	"encoding/json"
	"fmt"
	"log"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/mcp"
	"rodmcp/internal/webtools"
	"time"
)

// Example demonstrates how to use the MCP server programmatically for testing
func main() {
	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    "info",
		LogDir:      "test_logs",
		MaxSize:     10,
		MaxBackups:  3,
		MaxAge:      1,
		Compress:    false,
		Development: true,
	}

	logr, err := logger.New(logConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logr.Sync()

	fmt.Println("🚀 Testing RodMCP Web Development Tools")

	// Initialize browser manager
	browserConfig := browser.Config{
		Headless:     true,
		Debug:        false,
		SlowMotion:   100 * time.Millisecond,
		WindowWidth:  1280,
		WindowHeight: 720,
	}

	browserMgr := browser.NewManager(logr, browserConfig)
	if err := browserMgr.Start(browserConfig); err != nil {
		log.Fatalf("Failed to start browser: %v", err)
	}
	defer browserMgr.Stop()

	// Initialize MCP server (for tool registration)
	mcpServer := mcp.NewServer(logr)

	// Register tools
	createTool := webtools.NewCreatePageTool(logr)
	navigateTool := webtools.NewNavigatePageTool(logr, browserMgr)
	screenshotTool := webtools.NewScreenshotTool(logr, browserMgr)
	scriptTool := webtools.NewExecuteScriptTool(logr, browserMgr)
	previewTool := webtools.NewLivePreviewTool(logr)

	mcpServer.RegisterTool(createTool)
	mcpServer.RegisterTool(navigateTool)
	mcpServer.RegisterTool(screenshotTool)
	mcpServer.RegisterTool(scriptTool)
	mcpServer.RegisterTool(previewTool)

	fmt.Println("📝 Testing page creation...")

	// Test 1: Create a sample HTML page
	createArgs := map[string]interface{}{
		"filename": "test_page.html",
		"title":    "Test Page",
		"html": `
			<div id="content">
				<h1>Hello, RodMCP!</h1>
				<p>This is a test page created by the MCP web development tools.</p>
				<button id="clickme">Click Me!</button>
				<div id="result"></div>
			</div>
		`,
		"css": `
			body {
				font-family: Arial, sans-serif;
				max-width: 800px;
				margin: 0 auto;
				padding: 20px;
				background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
				color: white;
			}
			#content {
				background: rgba(255, 255, 255, 0.1);
				padding: 30px;
				border-radius: 10px;
				backdrop-filter: blur(10px);
			}
			button {
				background: #4CAF50;
				color: white;
				padding: 10px 20px;
				border: none;
				border-radius: 5px;
				cursor: pointer;
				font-size: 16px;
			}
			button:hover {
				background: #45a049;
			}
			#result {
				margin-top: 20px;
				padding: 10px;
				background: rgba(76, 175, 80, 0.2);
				border-radius: 5px;
				min-height: 20px;
			}
		`,
		"javascript": `
			document.getElementById('clickme').addEventListener('click', function() {
				const result = document.getElementById('result');
				result.innerHTML = '<p>✅ Button clicked! JavaScript is working properly.</p>';
				console.log('Button click event triggered');
			});
			
			// Add some dynamic content on load
			document.addEventListener('DOMContentLoaded', function() {
				console.log('Page loaded successfully');
				setTimeout(() => {
					const result = document.getElementById('result');
					result.innerHTML = '<p>⏰ Page has been loaded for 2 seconds</p>';
				}, 2000);
			});
		`,
	}

	result, err := createTool.Execute(createArgs)
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}
	fmt.Printf("✅ Create tool result: %s\n", result.Content[0].Text)

	fmt.Println("🌐 Testing page navigation...")

	// Test 2: Navigate to the created page
	navigateArgs := map[string]interface{}{
		"url": "test_page.html",
	}

	result, err = navigateTool.Execute(navigateArgs)
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}
	fmt.Printf("✅ Navigate tool result: %s\n", result.Content[0].Text)

	// Give the page time to load
	time.Sleep(1 * time.Second)

	fmt.Println("🖼️ Testing screenshot capture...")

	// Test 3: Take a screenshot
	screenshotArgs := map[string]interface{}{
		"filename": "test_screenshot.png",
	}

	result, err = screenshotTool.Execute(screenshotArgs)
	if err != nil {
		log.Fatalf("Failed to take screenshot: %v", err)
	}
	fmt.Printf("✅ Screenshot tool result: %s\n", result.Content[0].Text)

	fmt.Println("⚡ Testing JavaScript execution...")

	// Test 4: Execute JavaScript
	scriptArgs := map[string]interface{}{
		"script": `
			// Test DOM manipulation
			document.getElementById('clickme').click();
			
			// Return some data
			({
				title: document.title,
				buttonText: document.getElementById('clickme').textContent,
				resultContent: document.getElementById('result').innerHTML,
				timestamp: new Date().toISOString()
			})
		`,
	}

	result, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		log.Fatalf("Failed to execute script: %v", err)
	}
	fmt.Printf("✅ Script tool result: %s\n", result.Content[0].Text)

	// Pretty print the script result data
	if result.Content[0].Data != nil {
		dataJSON, _ := json.MarshalIndent(result.Content[0].Data, "", "  ")
		fmt.Printf("📊 Script execution data:\n%s\n", string(dataJSON))
	}

	fmt.Println("🔧 Testing live preview server...")

	// Test 5: Start live preview server
	previewArgs := map[string]interface{}{
		"directory": ".",
		"port":      8080,
	}

	result, err = previewTool.Execute(previewArgs)
	if err != nil {
		log.Fatalf("Failed to start preview server: %v", err)
	}
	fmt.Printf("✅ Preview server result: %s\n", result.Content[0].Text)

	fmt.Println("🧪 Testing navigation to preview server...")

	// Test 6: Navigate to the preview server
	navigatePreviewArgs := map[string]interface{}{
		"url": "http://localhost:8080/test_page.html",
	}

	result, err = navigateTool.Execute(navigatePreviewArgs)
	if err != nil {
		log.Fatalf("Failed to navigate to preview: %v", err)
	}
	fmt.Printf("✅ Preview navigation result: %s\n", result.Content[0].Text)

	// Take final screenshot
	time.Sleep(1 * time.Second)
	finalScreenshotArgs := map[string]interface{}{
		"filename": "preview_screenshot.png",
	}

	result, err = screenshotTool.Execute(finalScreenshotArgs)
	if err != nil {
		log.Fatalf("Failed to take final screenshot: %v", err)
	}
	fmt.Printf("✅ Final screenshot: %s\n", result.Content[0].Text)

	fmt.Println("\n🎉 All tests completed successfully!")
	fmt.Println("📁 Generated files:")
	fmt.Println("   • test_page.html - Sample HTML page")
	fmt.Println("   • test_screenshot.png - Initial screenshot")
	fmt.Println("   • preview_screenshot.png - Preview server screenshot")
	fmt.Println("   • test_logs/ - Log files")

	fmt.Println("\n✨ RodMCP is ready for use!")
}