package main

import (
	"fmt"
	"log"
	"path/filepath"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/mcp"
	"rodmcp/internal/webtools"
	"rodmcp/pkg/types"
	"strings"
	"time"
)

type TestResult struct {
	Name     string
	Success  bool
	Error    string
	Duration time.Duration
}

type TestSuite struct {
	Name    string
	Results []TestResult
	logger  *logger.Logger
}

func (ts *TestSuite) runTest(name string, testFunc func() error) {
	fmt.Printf("  Running %s...", name)
	start := time.Now()
	
	err := testFunc()
	duration := time.Since(start)
	
	result := TestResult{
		Name:     name,
		Success:  err == nil,
		Duration: duration,
	}
	
	if err != nil {
		result.Error = err.Error()
		fmt.Printf(" ❌ (%v) - %v\n", duration, err)
	} else {
		fmt.Printf(" ✅ (%v)\n", duration)
	}
	
	ts.Results = append(ts.Results, result)
}

func (ts *TestSuite) printSummary() {
	successful := 0
	total := len(ts.Results)
	
	fmt.Printf("\n📊 %s Summary:\n", ts.Name)
	for _, result := range ts.Results {
		status := "✅"
		if !result.Success {
			status = "❌"
		}
		fmt.Printf("   %s %s (%v)\n", status, result.Name, result.Duration)
		if !result.Success {
			fmt.Printf("      Error: %s\n", result.Error)
		}
		if result.Success {
			successful++
		}
	}
	fmt.Printf("   Success Rate: %d/%d (%.1f%%)\n", successful, total, float64(successful)/float64(total)*100)
}

func main() {
	fmt.Println("🚀 Comprehensive RodMCP Test Suite")
	fmt.Println("Testing all MCP tools and functionality")
	fmt.Println("=" + strings.Repeat("=", 50))

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

	// Initialize browser manager
	browserConfig := browser.Config{
		Headless:     true,
		Debug:        false,
		SlowMotion:   50 * time.Millisecond,
		WindowWidth:  1280,
		WindowHeight: 720,
	}

	browserMgr := browser.NewManager(logr, browserConfig)
	if err := browserMgr.Start(browserConfig); err != nil {
		log.Fatalf("Failed to start browser: %v", err)
	}
	defer browserMgr.Stop()

	// Initialize MCP server and register all tools
	mcpServer := mcp.NewServer(logr)
	tools := registerAllTools(logr, browserMgr, mcpServer)

	// Run comprehensive tests
	runAllTests(tools, browserMgr, logr)

	fmt.Println("\n🎉 Comprehensive test suite completed!")
}

func registerAllTools(logr *logger.Logger, browserMgr *browser.Manager, mcpServer *mcp.Server) map[string]interface{} {
	tools := make(map[string]interface{})

	// Browser automation tools
	tools["create_page"] = webtools.NewCreatePageTool(logr)
	tools["navigate_page"] = webtools.NewNavigatePageTool(logr, browserMgr)
	tools["screenshot"] = webtools.NewScreenshotTool(logr, browserMgr)
	tools["execute_script"] = webtools.NewExecuteScriptTool(logr, browserMgr)
	tools["browser_visibility"] = webtools.NewBrowserVisibilityTool(logr, browserMgr)
	tools["live_preview"] = webtools.NewLivePreviewTool(logr)

	// UI control tools
	tools["click_element"] = webtools.NewClickElementTool(logr, browserMgr)
	tools["type_text"] = webtools.NewTypeTextTool(logr, browserMgr)
	tools["wait"] = webtools.NewWaitTool(logr)
	tools["wait_for_element"] = webtools.NewWaitForElementTool(logr, browserMgr)
	tools["get_element_text"] = webtools.NewGetElementTextTool(logr, browserMgr)
	tools["get_element_attribute"] = webtools.NewGetElementAttributeTool(logr, browserMgr)
	tools["scroll"] = webtools.NewScrollTool(logr, browserMgr)
	tools["hover_element"] = webtools.NewHoverElementTool(logr, browserMgr)

	// Screen scraping tools
	tools["screen_scrape"] = webtools.NewScreenScrapeTool(logr, browserMgr)

	// File system tools with path validation
	fileValidator := webtools.NewPathValidator(webtools.DefaultFileAccessConfig())
	tools["read_file"] = webtools.NewReadFileTool(logr, fileValidator)
	tools["write_file"] = webtools.NewWriteFileTool(logr, fileValidator)
	tools["list_directory"] = webtools.NewListDirectoryTool(logr, fileValidator)

	// Network tools
	tools["http_request"] = webtools.NewHTTPRequestTool(logr)

	// Register all tools with MCP server
	for _, tool := range tools {
		if mcpTool, ok := tool.(mcp.Tool); ok {
			mcpServer.RegisterTool(mcpTool)
		}
	}

	return tools
}

func runAllTests(tools map[string]interface{}, browserMgr *browser.Manager, logr *logger.Logger) {
	// Test Suite 1: File System Operations
	fmt.Println("\n📁 File System Tools Test Suite")
	fileSystemSuite := &TestSuite{Name: "File System Tools", logger: logr}
	runFileSystemTests(fileSystemSuite, tools)
	fileSystemSuite.printSummary()

	// Test Suite 2: Basic Browser Automation
	fmt.Println("\n🌐 Browser Automation Test Suite")
	browserSuite := &TestSuite{Name: "Browser Automation", logger: logr}
	pageID := runBrowserAutomationTests(browserSuite, tools, browserMgr)
	browserSuite.printSummary()

	// Test Suite 3: UI Control Tools (requires active page)
	if pageID != "" {
		fmt.Println("\n🖱️  UI Control Tools Test Suite")
		uiSuite := &TestSuite{Name: "UI Control Tools", logger: logr}
		runUIControlTests(uiSuite, tools, pageID)
		uiSuite.printSummary()
	}

	// Test Suite 4: Network Tools
	fmt.Println("\n🌍 Network Tools Test Suite")
	networkSuite := &TestSuite{Name: "Network Tools", logger: logr}
	runNetworkTests(networkSuite, tools)
	networkSuite.printSummary()

	// Test Suite 5: Screen Scraping Tools
	if pageID != "" {
		fmt.Println("\n🕷️  Screen Scraping Test Suite")
		scrapeSuite := &TestSuite{Name: "Screen Scraping Tools", logger: logr}
		runScreenScrapingTests(scrapeSuite, tools, pageID)
		scrapeSuite.printSummary()
	}

	// Test Suite 6: Advanced JavaScript Execution
	if pageID != "" {
		fmt.Println("\n⚡ Advanced JavaScript Test Suite")
		jsSuite := &TestSuite{Name: "JavaScript Execution", logger: logr}
		runAdvancedJavaScriptTests(jsSuite, tools, pageID)
		jsSuite.printSummary()
	}
}

func runFileSystemTests(suite *TestSuite, tools map[string]interface{}) {
	testDir := "test_files"
	testFile := filepath.Join(testDir, "test.txt")
	testContent := "Hello, RodMCP!\nThis is a test file."

	// Test directory creation and file writing
	suite.runTest("Write file with directory creation", func() error {
		writeFileTool := tools["write_file"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"path":        testFile,
			"content":     testContent,
			"create_dirs": true,
		}
		
		_, err := writeFileTool.Execute(args)
		return err
	})

	// Test file reading
	suite.runTest("Read file", func() error {
		readFileTool := tools["read_file"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"path": testFile,
		}
		
		result, err := readFileTool.Execute(args)
		if err != nil {
			return err
		}
		
		if !strings.Contains(result.Content[0].Text, testContent) {
			return fmt.Errorf("file content mismatch")
		}
		
		return nil
	})

	// Test directory listing
	suite.runTest("List directory", func() error {
		listDirTool := tools["list_directory"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"path": testDir,
		}
		
		result, err := listDirTool.Execute(args)
		if err != nil {
			return err
		}
		
		if !strings.Contains(result.Content[0].Text, "test.txt") {
			return fmt.Errorf("test file not found in directory listing")
		}
		
		return nil
	})

	// Test writing JSON file
	suite.runTest("Write JSON file", func() error {
		writeFileTool := tools["write_file"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		jsonContent := `{
	"name": "RodMCP Test",
	"version": "1.0.0",
	"features": ["browser_automation", "ui_control", "file_system"]
}`
		
		args := map[string]interface{}{
			"path":        filepath.Join(testDir, "config.json"),
			"content":     jsonContent,
			"create_dirs": true,
		}
		
		_, err := writeFileTool.Execute(args)
		return err
	})
}

func runBrowserAutomationTests(suite *TestSuite, tools map[string]interface{}, browserMgr *browser.Manager) string {
	var pageID string
	testHTML := `<!DOCTYPE html>
<html>
<head>
	<title>RodMCP Test Page</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; }
		.container { max-width: 800px; margin: 0 auto; }
		button { padding: 10px 20px; margin: 5px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
		button:hover { background: #0056b3; }
		input { padding: 8px; margin: 5px; border: 1px solid #ddd; border-radius: 4px; }
		#result { margin: 20px 0; padding: 20px; background: #f8f9fa; border-radius: 4px; }
		.hidden { display: none; }
		.test-card { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 8px; }
		.lazy-content { opacity: 0; transition: opacity 0.5s; }
		.lazy-content.loaded { opacity: 1; }
		.delayed-content { display: none; }
	</style>
</head>
<body>
	<div class="container">
		<h1 id="title">RodMCP Test Page</h1>
		<p id="description">This page is used for comprehensive UI testing.</p>
		
		<div class="test-section">
			<h2>Button Tests</h2>
			<button id="simple-btn">Simple Button</button>
			<button id="counter-btn">Click Counter: <span id="counter">0</span></button>
		</div>
		
		<div class="test-section">
			<h2>Input Tests</h2>
			<input type="text" id="text-input" placeholder="Type here...">
			<input type="password" id="password-input" placeholder="Password">
			<textarea id="textarea" placeholder="Multi-line text..."></textarea>
		</div>
		
		<div class="test-section">
			<h2>Form Tests</h2>
			<form id="test-form">
				<input type="text" id="name" name="name" placeholder="Your name" required>
				<input type="email" id="email" name="email" placeholder="Your email" required>
				<button type="submit" id="submit-btn">Submit</button>
			</form>
		</div>
		
		<!-- Multiple Items Test Section -->
		<div class="test-section">
			<h2>Product Cards (Multiple Items)</h2>
			<div class="test-card" data-product-id="1">
				<h3 class="card-title">Wireless Headphones</h3>
				<img src="/images/headphones.jpg" alt="Headphones" class="card-image">
				<p class="card-description">High-quality wireless headphones</p>
				<span class="card-price" data-value="99.99">$99.99</span>
				<a href="/product/1" class="card-link">Buy Now</a>
			</div>
			<div class="test-card" data-product-id="2">
				<h3 class="card-title">Smart Watch</h3>
				<img src="/images/watch.jpg" alt="Smart Watch" class="card-image">
				<p class="card-description">Feature-rich smartwatch</p>
				<span class="card-price" data-value="199.99">$199.99</span>
				<a href="/product/2" class="card-link">Buy Now</a>
			</div>
		</div>
		
		<!-- Advanced Features Test Section -->
		<div class="test-section">
			<h2>Advanced Features</h2>
			<button id="show-hidden">Show Hidden Content</button>
			<div id="hidden-content" class="hidden">
				<p>This content was hidden!</p>
			</div>
			
			<div id="wait-trigger">
				<p>Click to load delayed content</p>
				<button id="load-delayed">Load Delayed Content</button>
				<div id="delayed-content" class="delayed-content">
					<p class="delayed-text">This content loaded after delay!</p>
				</div>
			</div>
			
			<div class="lazy-content" id="lazy-item">
				<p>This content appears after scrolling</p>
			</div>
		</div>
		
		<!-- Image and Link Tests -->
		<div class="test-section">
			<h2>Media and Links</h2>
			<img id="hero-image" src="/images/hero.jpg" alt="Hero Image" title="Main Hero Image" class="hero-img">
			<a id="main-link" href="https://example.com" title="External Link" class="external-link">Visit Example</a>
			<a id="internal-link" href="/page/about" class="internal-link">About Page</a>
		</div>
		
		<div id="result"></div>
	</div>
	
	<script>
		let counter = 0;
		
		document.getElementById('simple-btn').addEventListener('click', function() {
			document.getElementById('result').innerHTML = '<p style="color: green;">✅ Simple button clicked!</p>';
		});
		
		document.getElementById('counter-btn').addEventListener('click', function() {
			counter++;
			document.getElementById('counter').textContent = counter;
			document.getElementById('result').innerHTML = '<p style="color: blue;">🔢 Counter: ' + counter + '</p>';
		});
		
		document.getElementById('show-hidden').addEventListener('click', function() {
			const hidden = document.getElementById('hidden-content');
			hidden.classList.toggle('hidden');
			this.textContent = hidden.classList.contains('hidden') ? 'Show Hidden Content' : 'Hide Content';
		});
		
		document.getElementById('load-delayed').addEventListener('click', function() {
			setTimeout(() => {
				document.getElementById('delayed-content').style.display = 'block';
			}, 1000);
		});
		
		document.getElementById('test-form').addEventListener('submit', function(e) {
			e.preventDefault();
			const name = document.getElementById('name').value;
			const email = document.getElementById('email').value;
			document.getElementById('result').innerHTML = 
				'<p style="color: purple;">📝 Form submitted: ' + name + ' (' + email + ')</p>';
		});
		
		// Simulate lazy loading on scroll
		window.addEventListener('scroll', function() {
			const lazyItem = document.getElementById('lazy-item');
			const rect = lazyItem.getBoundingClientRect();
			if (rect.top < window.innerHeight && rect.bottom > 0) {
				lazyItem.classList.add('loaded');
			}
		});
		
		// Add some test data attributes
		document.getElementById('title').setAttribute('data-test', 'main-title');
		document.getElementById('simple-btn').setAttribute('data-testid', 'simple-button');
	</script>
</body>
</html>`

	// Test page creation
	suite.runTest("Create comprehensive test page", func() error {
		createPageTool := tools["create_page"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"filename": "comprehensive_test_page.html",
			"title":    "RodMCP Comprehensive Test Page",
			"html":     testHTML,
		}
		
		_, err := createPageTool.Execute(args)
		return err
	})

	// Test page navigation
	suite.runTest("Navigate to test page", func() error {
		navigateTool := tools["navigate_page"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"url": "comprehensive_test_page.html",
		}
		
		result, err := navigateTool.Execute(args)
		if err != nil {
			return err
		}
		
		// Extract page ID from result
		if len(result.Content) > 0 {
			resultText := result.Content[0].Text
			if strings.Contains(resultText, "page_") {
				// Look for pattern like "page_1234567890123456789"
				start := strings.Index(resultText, "page_")
				if start != -1 {
					// Extract everything from "page_" until whitespace or end
					substr := resultText[start:]
					end := strings.IndexAny(substr, " \t\n\r)")
					if end == -1 {
						pageID = substr
					} else {
						pageID = substr[:end]
					}
				}
			}
		}
		
		return nil
	})

	// Test screenshot
	suite.runTest("Take screenshot", func() error {
		screenshotTool := tools["screenshot"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"filename": "comprehensive_test_screenshot.png",
		}
		
		_, err := screenshotTool.Execute(args)
		return err
	})

	// Test basic script execution
	suite.runTest("Execute basic JavaScript", func() error {
		scriptTool := tools["execute_script"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"script": "document.title",
		}
		
		result, err := scriptTool.Execute(args)
		if err != nil {
			return err
		}
		
		if !strings.Contains(result.Content[0].Text, "RodMCP") {
			return fmt.Errorf("unexpected page title in script result")
		}
		
		return nil
	})

	return pageID
}

func runUIControlTests(suite *TestSuite, tools map[string]interface{}, pageID string) {
	// Test element clicking
	suite.runTest("Click simple button", func() error {
		clickTool := tools["click_element"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":  pageID,
			"selector": "#simple-btn",
		}
		
		_, err := clickTool.Execute(args)
		return err
	})

	// Test text typing
	suite.runTest("Type text in input field", func() error {
		typeTool := tools["type_text"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":  pageID,
			"selector": "#text-input",
			"text":     "Hello, RodMCP Testing!",
		}
		
		_, err := typeTool.Execute(args)
		return err
	})

	// Test waiting
	suite.runTest("Wait for specified duration", func() error {
		waitTool := tools["wait"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"seconds": 0.5, // 500ms = 0.5 seconds
		}
		
		_, err := waitTool.Execute(args)
		return err
	})

	// Test waiting for element
	suite.runTest("Wait for element to exist", func() error {
		waitElementTool := tools["wait_for_element"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":  pageID,
			"selector": "#title",
			"timeout":  5000,
		}
		
		_, err := waitElementTool.Execute(args)
		return err
	})

	// Test getting element text
	suite.runTest("Get element text", func() error {
		getTextTool := tools["get_element_text"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":  pageID,
			"selector": "#title",
		}
		
		result, err := getTextTool.Execute(args)
		if err != nil {
			return err
		}
		
		if !strings.Contains(result.Content[0].Text, "RodMCP Test Page") {
			return fmt.Errorf("unexpected element text")
		}
		
		return nil
	})

	// Test getting element attribute
	suite.runTest("Get element attribute", func() error {
		getAttrTool := tools["get_element_attribute"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":   pageID,
			"selector":  "#title",
			"attribute": "data-test",
		}
		
		result, err := getAttrTool.Execute(args)
		if err != nil {
			return err
		}
		
		if !strings.Contains(result.Content[0].Text, "main-title") {
			return fmt.Errorf("unexpected attribute value")
		}
		
		return nil
	})

	// Test scrolling
	suite.runTest("Scroll page", func() error {
		scrollTool := tools["scroll"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id": pageID,
			"x":       0.0,
			"y":       300.0,
		}
		
		_, err := scrollTool.Execute(args)
		return err
	})

	// Test hover element
	suite.runTest("Hover over element", func() error {
		hoverTool := tools["hover_element"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":  pageID,
			"selector": "#counter-btn",
		}
		
		_, err := hoverTool.Execute(args)
		return err
	})

	// Test form interaction
	suite.runTest("Fill and submit form", func() error {
		// Type in name field
		typeTool := tools["type_text"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		nameArgs := map[string]interface{}{
			"page_id":  pageID,
			"selector": "#name",
			"text":     "John Doe",
		}
		
		if _, err := typeTool.Execute(nameArgs); err != nil {
			return fmt.Errorf("failed to type name: %v", err)
		}

		// Type in email field
		emailArgs := map[string]interface{}{
			"page_id":  pageID,
			"selector": "#email",
			"text":     "john@example.com",
		}
		
		if _, err := typeTool.Execute(emailArgs); err != nil {
			return fmt.Errorf("failed to type email: %v", err)
		}

		// Click submit button
		clickTool := tools["click_element"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		submitArgs := map[string]interface{}{
			"page_id":  pageID,
			"selector": "#submit-btn",
		}
		
		if _, err := clickTool.Execute(submitArgs); err != nil {
			return fmt.Errorf("failed to click submit: %v", err)
		}

		return nil
	})
}

func runNetworkTests(suite *TestSuite, tools map[string]interface{}) {
	// Test GET request
	suite.runTest("HTTP GET request", func() error {
		httpTool := tools["http_request"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"url":    "https://httpbin.org/get",
			"method": "GET",
		}
		
		result, err := httpTool.Execute(args)
		if err != nil {
			return err
		}
		
		if !strings.Contains(result.Content[0].Text, "httpbin") {
			return fmt.Errorf("unexpected HTTP response")
		}
		
		return nil
	})

	// Test POST request with JSON data
	suite.runTest("HTTP POST request with JSON", func() error {
		httpTool := tools["http_request"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		jsonData := map[string]interface{}{
			"name":    "RodMCP",
			"version": "1.0.0",
			"test":    true,
		}
		
		args := map[string]interface{}{
			"url":    "https://httpbin.org/post",
			"method": "POST",
			"json":   jsonData,
		}
		
		result, err := httpTool.Execute(args)
		if err != nil {
			return err
		}
		
		// Check if the response contains our POST data in JSON format
		responseText := result.Content[0].Text
		if !strings.Contains(responseText, "RodMCP") && !strings.Contains(responseText, "name") {
			return fmt.Errorf("POST data not found in response: %s", responseText)
		}
		
		return nil
	})

	// Test request with custom headers
	suite.runTest("HTTP request with custom headers", func() error {
		httpTool := tools["http_request"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		headers := map[string]interface{}{
			"User-Agent":      "RodMCP-Test-Suite/1.0",
			"X-Custom-Header": "test-value",
		}
		
		args := map[string]interface{}{
			"url":     "https://httpbin.org/headers",
			"method":  "GET",
			"headers": headers,
		}
		
		result, err := httpTool.Execute(args)
		if err != nil {
			return err
		}
		
		if !strings.Contains(result.Content[0].Text, "RodMCP-Test-Suite") {
			return fmt.Errorf("custom headers not found in response")
		}
		
		return nil
	})
}

func runAdvancedJavaScriptTests(suite *TestSuite, tools map[string]interface{}, pageID string) {
	// Test complex object return
	suite.runTest("Complex JavaScript object return", func() error {
		scriptTool := tools["execute_script"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		script := `
		// Click the counter button multiple times and return detailed info
		for (let i = 0; i < 3; i++) {
			document.getElementById('counter-btn').click();
		}
		
		({
			pageTitle: document.title,
			url: window.location.href,
			counterValue: parseInt(document.getElementById('counter').textContent),
			formElements: {
				nameInput: document.getElementById('name').value,
				emailInput: document.getElementById('email').value
			},
			elementCounts: {
				buttons: document.querySelectorAll('button').length,
				inputs: document.querySelectorAll('input').length,
				divs: document.querySelectorAll('div').length
			},
			timestamp: new Date().toISOString()
		})`
		
		args := map[string]interface{}{
			"script": script,
		}
		
		result, err := scriptTool.Execute(args)
		if err != nil {
			return err
		}
		
		// Verify the result contains expected data
		resultText := result.Content[0].Text
		if !strings.Contains(resultText, "counterValue") || !strings.Contains(resultText, "elementCounts") {
			return fmt.Errorf("complex object not properly returned")
		}
		
		return nil
	})

	// Test DOM manipulation
	suite.runTest("DOM manipulation script", func() error {
		scriptTool := tools["execute_script"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		script := `
		// Create a new element and add it to the page
		const newDiv = document.createElement('div');
		newDiv.id = 'test-created-element';
		newDiv.innerHTML = '<p style="color: red; font-weight: bold;">✨ This element was created by JavaScript!</p>';
		newDiv.style.padding = '20px';
		newDiv.style.border = '2px solid red';
		newDiv.style.margin = '10px 0';
		
		// Insert it after the result div
		const resultDiv = document.getElementById('result');
		resultDiv.parentNode.insertBefore(newDiv, resultDiv.nextSibling);
		
		// Return confirmation
		'DOM element created successfully'`
		
		args := map[string]interface{}{
			"script": script,
		}
		
		result, err := scriptTool.Execute(args)
		if err != nil {
			return err
		}
		
		if !strings.Contains(result.Content[0].Text, "successfully") {
			return fmt.Errorf("DOM manipulation script failed")
		}
		
		return nil
	})

	// Test async operations simulation
	suite.runTest("Async operations simulation", func() error {
		scriptTool := tools["execute_script"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		script := `
		// Simulate async operation with Promise
		new Promise((resolve) => {
			// Show hidden content
			const hiddenContent = document.getElementById('hidden-content');
			const showButton = document.getElementById('show-hidden');
			
			if (hiddenContent.classList.contains('hidden')) {
				showButton.click();
			}
			
			// Wait a bit and resolve
			setTimeout(() => {
				resolve({
					action: 'async_operation_completed',
					hiddenContentVisible: !hiddenContent.classList.contains('hidden'),
					timestamp: new Date().toISOString()
				});
			}, 100);
		})`
		
		args := map[string]interface{}{
			"script": script,
		}
		
		_, err := scriptTool.Execute(args)
		// Note: This test may fail because go-rod's Eval might not handle Promises the same way
		// But we're testing to see if it handles the Promise syntax gracefully
		return err
	})

	// Test error handling
	suite.runTest("JavaScript error handling", func() error {
		scriptTool := tools["execute_script"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		script := `
		try {
			// This will cause an error
			nonExistentFunction();
		} catch (error) {
			// Return error information instead of throwing
			({
				errorCaught: true,
				errorMessage: error.message,
				errorType: error.name
			})
		}`
		
		args := map[string]interface{}{
			"script": script,
		}
		
		result, err := scriptTool.Execute(args)
		if err != nil {
			return err
		}
		
		if !strings.Contains(result.Content[0].Text, "errorCaught") {
			return fmt.Errorf("error handling not working properly")
		}
		
		return nil
	})
}

func runScreenScrapingTests(suite *TestSuite, tools map[string]interface{}, pageID string) {
	// Test basic element scraping from current comprehensive test page
	suite.runTest("Basic element scraping", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id": pageID,
			"selectors": map[string]interface{}{
				"title":       "#title",
				"description": "#description",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		// Use string representation approach like standalone tests
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"].(map[string]interface{})
		
		// Check title extraction
		titleData, exists := scrapedData["title"]
		if !exists {
			return fmt.Errorf("title not found in scraped data")
		}
		
		// Convert to string for validation like standalone tests
		titleStr := fmt.Sprintf("%v", titleData)
		if !strings.Contains(titleStr, "RodMCP Test Page") {
			return fmt.Errorf("expected title content not found")
		}
		
		return nil
	})

	// Test multiple item extraction - Product Cards
	suite.runTest("Multiple item extraction - Product cards", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":            pageID,
			"extract_type":       "multiple",
			"container_selector": ".test-card",
			"selectors": map[string]interface{}{
				"title":       ".card-title",
				"description": ".card-description",
				"price":       ".card-price",
				"link":        ".card-link",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"]
		
		items, ok := scrapedData.([]map[string]interface{})
		if !ok {
			return fmt.Errorf("scraped data is not an array of maps")
		}
		
		if len(items) != 2 {
			return fmt.Errorf("expected 2 product cards, got %d", len(items))
		}
		
		// Check first product
		firstItem := items[0]
		titleData := firstItem["title"]
		titleStr := fmt.Sprintf("%v", titleData)
		if !strings.Contains(titleStr, "Wireless Headphones") {
			return fmt.Errorf("first product title not found")
		}
		
		return nil
	})

	// Test image element extraction
	suite.runTest("Image element extraction", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id": pageID,
			"selectors": map[string]interface{}{
				"hero_image": "#hero-image",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"].(map[string]interface{})
		
		imageData, exists := scrapedData["hero_image"]
		if !exists {
			return fmt.Errorf("hero_image not found in scraped data")
		}
		
		imageStr := fmt.Sprintf("%v", imageData)
		if !strings.Contains(imageStr, "src") || !strings.Contains(imageStr, "alt") {
			return fmt.Errorf("image attributes not found")
		}
		
		if !strings.Contains(imageStr, "Hero Image") {
			return fmt.Errorf("expected alt text not found")
		}
		
		return nil
	})

	// Test link element extraction
	suite.runTest("Link element extraction", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id": pageID,
			"selectors": map[string]interface{}{
				"main_link":     "#main-link",
				"internal_link": "#internal-link",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"].(map[string]interface{})
		
		linkData, exists := scrapedData["main_link"]
		if !exists {
			return fmt.Errorf("main_link not found in scraped data")
		}
		
		linkStr := fmt.Sprintf("%v", linkData)
		if !strings.Contains(linkStr, "href") || !strings.Contains(linkStr, "title") {
			return fmt.Errorf("link attributes not found")
		}
		
		if !strings.Contains(linkStr, "example.com") {
			return fmt.Errorf("expected href not found")
		}
		
		return nil
	})

	// Test wait_for functionality
	suite.runTest("Wait for element functionality", func() error {
		// First trigger the delayed content
		clickTool := tools["click_element"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		clickArgs := map[string]interface{}{
			"page_id":  pageID,
			"selector": "#load-delayed",
		}
		
		if _, err := clickTool.Execute(clickArgs); err != nil {
			return fmt.Errorf("failed to click load button: %w", err)
		}
		
		// Now test wait_for functionality
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":      pageID,
			"wait_for":     "#delayed-content",
			"wait_timeout": 3,
			"selectors": map[string]interface{}{
				"delayed_text": ".delayed-text",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"].(map[string]interface{})
		
		delayedData, exists := scrapedData["delayed_text"]
		if !exists {
			return fmt.Errorf("delayed_text not found after wait")
		}
		
		delayedStr := fmt.Sprintf("%v", delayedData)
		if !strings.Contains(delayedStr, "loaded after delay") {
			return fmt.Errorf("expected delayed content not found")
		}
		
		return nil
	})

	// Test custom script functionality
	suite.runTest("Custom script execution", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		customScript := `
			// Add custom data for testing
			const testDiv = document.createElement('div');
			testDiv.id = 'custom-test-element';
			testDiv.textContent = 'Custom script executed successfully';
			testDiv.className = 'custom-element';
			document.body.appendChild(testDiv);
		`
		
		args := map[string]interface{}{
			"page_id":       pageID,
			"custom_script": customScript,
			"selectors": map[string]interface{}{
				"custom_element": "#custom-test-element",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"].(map[string]interface{})
		
		customData, exists := scrapedData["custom_element"]
		if !exists {
			return fmt.Errorf("custom_element not found after script execution")
		}
		
		customStr := fmt.Sprintf("%v", customData)
		if !strings.Contains(customStr, "Custom script executed successfully") {
			return fmt.Errorf("custom script content not found")
		}
		
		return nil
	})

	// Test scroll to load functionality
	suite.runTest("Scroll to load functionality", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":        pageID,
			"scroll_to_load": true,
			"selectors": map[string]interface{}{
				"lazy_item": "#lazy-item",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"].(map[string]interface{})
		
		lazyData, exists := scrapedData["lazy_item"]
		if !exists {
			return fmt.Errorf("lazy_item not found after scroll")
		}
		
		lazyStr := fmt.Sprintf("%v", lazyData)
		if !strings.Contains(lazyStr, "appears after scrolling") {
			return fmt.Errorf("lazy loaded content not found")
		}
		
		return nil
	})

	// Test input element scraping with values
	suite.runTest("Input element scraping", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id": pageID,
			"selectors": map[string]interface{}{
				"text_input": "#text-input",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"].(map[string]interface{})
		
		inputData, exists := scrapedData["text_input"]
		if !exists {
			return fmt.Errorf("text_input not found in scraped data")
		}
		
		inputStr := fmt.Sprintf("%v", inputData)
		if !strings.Contains(inputStr, "Hello, RodMCP Testing!") {
			return fmt.Errorf("expected input value not found")
		}
		
		return nil
	})

	// Test metadata inclusion
	suite.runTest("Scraping with metadata", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id": pageID,
			"selectors": map[string]interface{}{
				"title": "#title",
			},
			"include_metadata": true,
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		
		// Check for metadata
		if _, exists := data["metadata"]; !exists {
			return fmt.Errorf("metadata not found in result")
		}
		
		// Check for timestamp
		if _, exists := data["timestamp"]; !exists {
			return fmt.Errorf("timestamp not found in result")
		}
		
		return nil
	})

	// Test metadata disabled
	suite.runTest("Scraping without metadata", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id": pageID,
			"selectors": map[string]interface{}{
				"title": "#title",
			},
			"include_metadata": false,
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		
		// Should not have metadata
		if _, exists := data["metadata"]; exists {
			return fmt.Errorf("metadata should not be included when disabled")
		}
		
		// Should still have data
		if _, exists := data["data"]; !exists {
			return fmt.Errorf("scraped data should still be present")
		}
		
		return nil
	})

	// Test invalid container selector
	suite.runTest("Invalid container selector handling", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id":            pageID,
			"extract_type":       "multiple",
			"container_selector": ".does-not-exist",
			"selectors": map[string]interface{}{
				"title": "h1",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"]
		
		items, ok := scrapedData.([]map[string]interface{})
		if !ok {
			return fmt.Errorf("scraped data is not an array of maps")
		}
		
		// Should return empty array for non-existent container
		if len(items) != 0 {
			return fmt.Errorf("expected empty array for invalid container, got %d items", len(items))
		}
		
		return nil
	})

	// Test error handling with non-existent elements
	suite.runTest("Non-existent element handling", func() error {
		scrapeTool := tools["screen_scrape"].(interface {
			Execute(map[string]interface{}) (*types.CallToolResponse, error)
		})
		
		args := map[string]interface{}{
			"page_id": pageID,
			"selectors": map[string]interface{}{
				"existing":     "#title",
				"non_existent": "#does-not-exist",
			},
		}
		
		result, err := scrapeTool.Execute(args)
		if err != nil {
			return err
		}
		
		data := result.Content[0].Data.(map[string]interface{})
		scrapedData := data["data"].(map[string]interface{})
		
		// Should have existing element
		titleData, exists := scrapedData["existing"]
		if !exists {
			return fmt.Errorf("existing element not found")
		}
		
		titleStr := fmt.Sprintf("%v", titleData)
		if !strings.Contains(titleStr, "RodMCP Test Page") {
			return fmt.Errorf("expected title content not found")
		}
		
		// Non-existent should be null or missing - this is acceptable behavior
		return nil
	})
}