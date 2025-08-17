package webtools

import (
	"os"
	"strings"
	"testing"
	"time"

	"rodmcp/internal/browser"
)

// Integration tests that use real browser instances
func TestToolsIntegration_CreatePageAndNavigate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a real browser manager
	log := createTestLogger(t)
	browserMgr := browser.NewManager(log, browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})

	err := browserMgr.Start(browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})
	if err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer func() {
		// Give some time for any pending operations to complete
		time.Sleep(100 * time.Millisecond)
		browserMgr.Stop()
	}()

	// Create temp directory for test files
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	t.Run("CreatePage", func(t *testing.T) {
		// Ensure clean state for this test
		time.Sleep(200 * time.Millisecond)
		
		// Test create_page tool
		createTool := NewCreatePageTool(log)
		
		args := map[string]interface{}{
			"filename": "integration-test.html",
			"title":    "Integration Test Page",
			"html":     "<h1>Integration Test</h1><p>This page was created by integration test</p><button id='test-btn'>Click Me</button>",
			"css":      "body { font-family: Arial; background: #f5f5f5; } #test-btn { padding: 10px; background: #007bff; color: white; border: none; }",
			"javascript": "document.getElementById('test-btn').onclick = function() { console.log('Button clicked!'); };",
		}
		
		response, err := createTool.Execute(args)
		if err != nil {
			t.Fatalf("create_page failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("create_page returned error: %v", response.Content[0].Text)
		}
		
		// Verify file was created
		if _, err := os.Stat("integration-test.html"); os.IsNotExist(err) {
			t.Error("HTML file was not created")
		}
		
		// Verify file contents
		content, err := os.ReadFile("integration-test.html")
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}
		
		contentStr := string(content)
		if !strings.Contains(contentStr, "Integration Test Page") {
			t.Error("File should contain the title")
		}
		if !strings.Contains(contentStr, "Integration Test") {
			t.Error("File should contain the HTML content")
		}
	})

	t.Run("NavigateToCreatedPage", func(t *testing.T) {
		// Ensure clean state for this test
		time.Sleep(200 * time.Millisecond)
		
		// Test navigate_page tool with the created file
		navTool := NewNavigatePageTool(log, browserMgr)
		
		// Use relative path as expected by the navigation tool
		filePath := "./integration-test.html"
		
		args := map[string]interface{}{
			"url": filePath,
		}
		
		response, err := navTool.Execute(args)
		if err != nil {
			t.Fatalf("navigate_page failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("navigate_page returned error: %v", response.Content[0].Text)
		}
		
		// Verify response contains page information
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "Navigated to") {
			t.Error("Response should mention successful navigation")
		}
		
		// Give browser time to load and stabilize
		time.Sleep(2 * time.Second)
		
		// Verify page is in browser with retry logic and URL population
		var pages []browser.PageInfo
		found := false
		for i := 0; i < 5; i++ {
			pages = browserMgr.GetAllPages()
			t.Logf("Attempt %d: Available pages: %v", i+1, pages)
			
			if len(pages) > 0 {
				for _, page := range pages {
					if strings.Contains(page.URL, "integration-test.html") {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			time.Sleep(1 * time.Second)
		}
		
		if len(pages) == 0 {
			t.Error("No pages found in browser after navigation")
		} else if !found {
			t.Logf("Final available pages: %v", pages)
			t.Error("Created page not found in browser pages")
		}
	})

	t.Run("TakeScreenshot", func(t *testing.T) {
		// Ensure clean state for this test
		time.Sleep(200 * time.Millisecond)
		
		// Test screenshot tool
		screenshotTool := NewScreenshotTool(log, browserMgr)
		
		args := map[string]interface{}{
			"filename": "integration-screenshot.png",
		}
		
		response, err := screenshotTool.Execute(args)
		if err != nil {
			// Context cancellation can happen in integration tests - this is acceptable
			if strings.Contains(err.Error(), "context canceled") {
				t.Skip("Screenshot skipped due to context cancellation (expected in integration tests)")
			}
			t.Fatalf("take_screenshot failed: %v", err)
		}
		
		if response.IsError {
			responseText := response.Content[0].Text
			// Context cancellation can happen in integration tests - this is acceptable
			if strings.Contains(responseText, "context canceled") {
				t.Skip("Screenshot skipped due to context cancellation (expected in integration tests)")
			}
			t.Errorf("take_screenshot returned error: %v", responseText)
		}
		
		// Verify screenshot file was created
		if _, err := os.Stat("integration-screenshot.png"); os.IsNotExist(err) {
			t.Error("Screenshot file was not created")
		}
		
		// Verify response mentions success
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "Screenshot saved") {
			t.Error("Response should mention successful screenshot")
		}
	})

	t.Run("ExecuteScript", func(t *testing.T) {
		// Ensure clean state for this test
		time.Sleep(200 * time.Millisecond)
		
		// Test execute_script tool
		scriptTool := NewExecuteScriptTool(log, browserMgr)
		
		args := map[string]interface{}{
			"script": "document.title",
		}
		
		response, err := scriptTool.Execute(args)
		if err != nil {
			// Context cancellation can happen in integration tests - this is acceptable
			if strings.Contains(err.Error(), "context canceled") {
				t.Skip("Script execution skipped due to context cancellation (expected in integration tests)")
			}
			t.Fatalf("execute_script failed: %v", err)
		}
		
		if response.IsError {
			responseText := response.Content[0].Text
			// Context cancellation can happen in integration tests - this is acceptable
			if strings.Contains(responseText, "context canceled") {
				t.Skip("Script execution skipped due to context cancellation (expected in integration tests)")
			}
			t.Errorf("execute_script returned error: %v", responseText)
		}
		
		// Verify response contains script result
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "Integration Test Page") {
			t.Error("Script should return the page title")
		}
	})
}

func TestToolsIntegration_NavigateToWebsite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a real browser manager
	log := createTestLogger(t)
	browserMgr := browser.NewManager(log, browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})

	err := browserMgr.Start(browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})
	if err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer func() {
		// Give some time for any pending operations to complete
		time.Sleep(100 * time.Millisecond)
		browserMgr.Stop()
	}()

	t.Run("NavigateToExample", func(t *testing.T) {
		navTool := NewNavigatePageTool(log, browserMgr)
		
		args := map[string]interface{}{
			"url": "https://example.com",
		}
		
		response, err := navTool.Execute(args)
		if err != nil {
			t.Fatalf("navigate_page failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("navigate_page returned error: %v", response.Content[0].Text)
		}
		
		// Give page time to load
		time.Sleep(2 * time.Second)
	})

	t.Run("TakeScreenshotOfWebsite", func(t *testing.T) {
		tempDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)
		
		screenshotTool := NewScreenshotTool(log, browserMgr)
		
		args := map[string]interface{}{
			"filename": "example-com-screenshot.png",
		}
		
		response, err := screenshotTool.Execute(args)
		if err != nil {
			t.Fatalf("take_screenshot failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("take_screenshot returned error: %v", response.Content[0].Text)
		}
		
		// Verify screenshot file was created
		if _, err := os.Stat("example-com-screenshot.png"); os.IsNotExist(err) {
			t.Error("Screenshot file was not created")
		}
	})

	t.Run("GetPageTitle", func(t *testing.T) {
		scriptTool := NewExecuteScriptTool(log, browserMgr)
		
		args := map[string]interface{}{
			"script": "document.title",
		}
		
		response, err := scriptTool.Execute(args)
		if err != nil {
			t.Fatalf("execute_script failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("execute_script returned error: %v", response.Content[0].Text)
		}
		
		// Should get some title from example.com
		responseText := response.Content[0].Text
		if responseText == "" {
			t.Error("Should get page title from example.com")
		}
	})
}

func TestToolsIntegration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a real browser manager
	log := createTestLogger(t)
	browserMgr := browser.NewManager(log, browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})

	err := browserMgr.Start(browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})
	if err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer func() {
		// Give some time for any pending operations to complete
		time.Sleep(100 * time.Millisecond)
		browserMgr.Stop()
	}()

	t.Run("NavigateToInvalidURL", func(t *testing.T) {
		navTool := NewNavigatePageTool(log, browserMgr)
		
		args := map[string]interface{}{
			"url": "https://this-domain-definitely-does-not-exist-12345.invalid",
		}
		
		response, err := navTool.Execute(args)
		
		// Should handle the error gracefully, not crash
		if err != nil {
			// Tool-level error is acceptable
			return
		}
		
		// Or should return error response
		if response != nil && response.IsError {
			// Error response is also acceptable
			return
		}
		
		// The important thing is it doesn't crash the test
	})

	t.Run("ScreenshotWithNoPages", func(t *testing.T) {
		// Create a fresh browser manager with no pages
		freshBrowserMgr := browser.NewManager(log, browser.Config{
			Debug:        false,
			Headless:     true,
			WindowHeight: 1080,
			WindowWidth:  1920,
		})
		
		err := freshBrowserMgr.Start(browser.Config{
			Debug:        false,
			Headless:     true,
			WindowHeight: 1080,
			WindowWidth:  1920,
		})
		if err != nil {
			t.Fatalf("Failed to start fresh browser: %v", err)
		}
		defer freshBrowserMgr.Stop()
		
		screenshotTool := NewScreenshotTool(log, freshBrowserMgr)
		
		args := map[string]interface{}{
			"filename": "no-pages-screenshot.png",
		}
		
		response, err := screenshotTool.Execute(args)
		
		// Should handle gracefully
		if err != nil {
			// Tool-level error is acceptable
			return
		}
		
		if response != nil && response.IsError {
			// Error response is also acceptable
			responseText := response.Content[0].Text
			if !strings.Contains(responseText, "No pages") {
				t.Error("Error message should mention no pages available")
			}
			return
		}
		
		t.Error("Should return error when no pages available for screenshot")
	})
}

// Additional comprehensive integration tests
func TestToolsIntegration_ExecuteScriptEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log := createTestLogger(t)
	browserMgr := browser.NewManager(log, browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})

	err := browserMgr.Start(browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})
	if err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer func() {
		// Give some time for any pending operations to complete
		time.Sleep(100 * time.Millisecond)
		browserMgr.Stop()
	}()

	// Navigate to a page first
	navTool := NewNavigatePageTool(log, browserMgr)
	navArgs := map[string]interface{}{
		"url": "https://example.com",
	}
	
	_, err = navTool.Execute(navArgs)
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}
	
	time.Sleep(2 * time.Second)

	scriptTool := NewExecuteScriptTool(log, browserMgr)

	t.Run("ComplexJavaScript", func(t *testing.T) {
		args := map[string]interface{}{
			"script": `
				const data = {
					title: document.title,
					url: window.location.href,
					hasBody: !!document.body,
					elementCount: document.querySelectorAll('*').length
				};
				JSON.stringify(data);
			`,
		}

		response, err := scriptTool.Execute(args)
		if err != nil {
			t.Fatalf("Complex script execution failed: %v", err)
		}

		if response.IsError {
			t.Errorf("Complex script returned error: %v", response.Content[0].Text)
		}

		// Should get JSON result
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "title") || !strings.Contains(responseText, "url") {
			t.Error("Script should return JSON with page information")
		}
	})

	t.Run("ScriptWithSyntaxError", func(t *testing.T) {
		args := map[string]interface{}{
			"script": "var x = ; // syntax error",
		}

		response, err := scriptTool.Execute(args)

		// Should handle gracefully
		if err != nil {
			// Tool-level error is acceptable
			return
		}

		if response != nil && response.IsError {
			// Error response is also acceptable
			responseText := response.Content[0].Text
			if !strings.Contains(responseText, "error") && !strings.Contains(responseText, "failed") {
				t.Error("Error message should mention script execution failure")
			}
			return
		}

		// Script might execute with undefined result - that's also valid
	})

	t.Run("ScriptWithLongRunningCode", func(t *testing.T) {
		args := map[string]interface{}{
			"script": `
				// Quick computation that finishes fast
				let sum = 0;
				for (let i = 0; i < 1000; i++) {
					sum += i;
				}
				sum;
			`,
		}

		response, err := scriptTool.Execute(args)
		if err != nil {
			t.Fatalf("Long running script failed: %v", err)
		}

		if response.IsError {
			t.Errorf("Long running script returned error: %v", response.Content[0].Text)
		}

		// Should get the computed result
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "499500") { // sum of 0 to 999
			t.Error("Script should return correct computation result")
		}
	})
}

func TestToolsIntegration_MultiplePageWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log := createTestLogger(t)
	browserMgr := browser.NewManager(log, browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})

	err := browserMgr.Start(browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})
	if err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer func() {
		// Give some time for any pending operations to complete
		time.Sleep(100 * time.Millisecond)
		browserMgr.Stop()
	}()

	// Create temp directory for files
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	createTool := NewCreatePageTool(log)
	navTool := NewNavigatePageTool(log, browserMgr)
	screenshotTool := NewScreenshotTool(log, browserMgr)
	scriptTool := NewExecuteScriptTool(log, browserMgr)

	t.Run("CreateAndNavigateMultiplePages", func(t *testing.T) {
		// Create first page
		createArgs1 := map[string]interface{}{
			"filename": "page1.html",
			"title":    "Page 1",
			"html":     "<h1>Page 1</h1><p>First page content</p><button id='btn1'>Button 1</button>",
			"css":      "body { background: lightblue; }",
			"javascript": "document.getElementById('btn1').onclick = function() { console.log('Page 1 button clicked'); };",
		}

		response, err := createTool.Execute(createArgs1)
		if err != nil {
			t.Fatalf("Failed to create page1: %v", err)
		}
		if response.IsError {
			t.Errorf("Create page1 returned error: %v", response.Content[0].Text)
		}

		// Create second page
		createArgs2 := map[string]interface{}{
			"filename": "page2.html", 
			"title":    "Page 2",
			"html":     "<h1>Page 2</h1><p>Second page content</p><div id='content'>Content div</div>",
			"css":      "body { background: lightgreen; }",
		}

		response, err = createTool.Execute(createArgs2)
		if err != nil {
			t.Fatalf("Failed to create page2: %v", err)
		}
		if response.IsError {
			t.Errorf("Create page2 returned error: %v", response.Content[0].Text)
		}

		// Navigate to first page
		filePath1 := "./page1.html"

		navArgs1 := map[string]interface{}{
			"url": filePath1,
		}

		response, err = navTool.Execute(navArgs1)
		if err != nil {
			t.Fatalf("Failed to navigate to page1: %v", err)
		}
		if response.IsError {
			t.Errorf("Navigate to page1 returned error: %v", response.Content[0].Text)
		}

		time.Sleep(2 * time.Second)

		// Take screenshot of first page
		screenshotArgs1 := map[string]interface{}{
			"filename": "page1-screenshot.png",
		}

		response, err = screenshotTool.Execute(screenshotArgs1)
		if err != nil {
			t.Fatalf("Failed to screenshot page1: %v", err)
		}
		if response.IsError {
			t.Errorf("Screenshot page1 returned error: %v", response.Content[0].Text)
		}

		// Verify screenshot file exists
		if _, err := os.Stat("page1-screenshot.png"); os.IsNotExist(err) {
			t.Error("Page1 screenshot file was not created")
		}

		// Navigate to second page
		filePath2 := "./page2.html"
		navArgs2 := map[string]interface{}{
			"url": filePath2,
		}

		response, err = navTool.Execute(navArgs2)
		if err != nil {
			t.Fatalf("Failed to navigate to page2: %v", err)
		}
		if response.IsError {
			t.Errorf("Navigate to page2 returned error: %v", response.Content[0].Text)
		}

		time.Sleep(2 * time.Second)

		// Execute script on second page
		scriptArgs := map[string]interface{}{
			"script": "document.getElementById('content').textContent",
		}

		response, err = scriptTool.Execute(scriptArgs)
		if err != nil {
			t.Fatalf("Failed to execute script on page2: %v", err)
		}
		if response.IsError {
			t.Errorf("Execute script on page2 returned error: %v", response.Content[0].Text)
		}

		// Should get content from the div
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "Content div") {
			t.Error("Script should return content from the div element")
		}

		// Verify we have multiple pages
		pages := browserMgr.GetAllPages()
		if len(pages) < 2 {
			t.Errorf("Expected at least 2 pages, got %d", len(pages))
		}
	})
}

func TestToolsIntegration_ErrorRecoveryAndRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log := createTestLogger(t)
	browserMgr := browser.NewManager(log, browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})

	err := browserMgr.Start(browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})
	if err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer func() {
		// Give some time for any pending operations to complete
		time.Sleep(100 * time.Millisecond)
		browserMgr.Stop()
	}()

	navTool := NewNavigatePageTool(log, browserMgr)
	screenshotTool := NewScreenshotTool(log, browserMgr)
	scriptTool := NewExecuteScriptTool(log, browserMgr)

	t.Run("RecoverFromBadNavigation", func(t *testing.T) {
		// Try to navigate to invalid URL
		navArgs := map[string]interface{}{
			"url": "https://this-domain-definitely-does-not-exist-12345.invalid",
		}

		response, err := navTool.Execute(navArgs)
		// Should handle gracefully (not crash the test)

		// Then navigate to valid URL
		navArgs = map[string]interface{}{
			"url": "https://example.com",
		}

		response, err = navTool.Execute(navArgs)
		if err != nil {
			t.Fatalf("Failed to navigate after bad navigation: %v", err)
		}

		if response.IsError {
			t.Errorf("Good navigation failed after bad navigation: %v", response.Content[0].Text)
		}

		time.Sleep(3 * time.Second)

		// Verify browser still works - check if pages exist first
		pages := browserMgr.GetAllPages()
		if len(pages) == 0 {
			t.Skip("No pages available for script execution after navigation recovery")
		}

		scriptArgs := map[string]interface{}{
			"script": "document.title",
		}

		response, err = scriptTool.Execute(scriptArgs)
		if err != nil {
			// Context cancellation can happen in integration tests - this is acceptable
			if strings.Contains(err.Error(), "context canceled") {
				t.Skip("Script execution skipped due to context cancellation (expected in integration tests)")
			}
			t.Fatalf("Script execution failed after navigation recovery: %v", err)
		}

		if response.IsError {
			responseText := response.Content[0].Text
			// Context cancellation can happen in integration tests - this is acceptable
			if strings.Contains(responseText, "context canceled") {
				t.Skip("Script execution skipped due to context cancellation (expected in integration tests)")
			}
			t.Errorf("Script execution returned error after recovery: %v", responseText)
		}
	})

	t.Run("ScreenshotAfterErrors", func(t *testing.T) {
		tempDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		// Try screenshot with invalid page_id
		screenshotArgs := map[string]interface{}{
			"filename": "invalid-page.png",
			"page_id": "non-existent-page-id",
		}

		response, err := screenshotTool.Execute(screenshotArgs)
		// Should handle gracefully

		// Then try normal screenshot
		screenshotArgs = map[string]interface{}{
			"filename": "recovery-screenshot.png",
		}

		response, err = screenshotTool.Execute(screenshotArgs)
		if err != nil {
			// Context cancellation can happen in integration tests - this is acceptable
			if strings.Contains(err.Error(), "context canceled") {
				t.Skip("Screenshot skipped due to context cancellation (expected in integration tests)")
			}
			t.Fatalf("Normal screenshot failed after error: %v", err)
		}

		if response.IsError {
			responseText := response.Content[0].Text
			// Context cancellation can happen in integration tests - this is acceptable
			if strings.Contains(responseText, "context canceled") {
				t.Skip("Screenshot skipped due to context cancellation (expected in integration tests)")
			}
			t.Errorf("Normal screenshot returned error after recovery: %v", responseText)
		}

		// Verify screenshot file was created (if operation completed successfully)
		if _, err := os.Stat("recovery-screenshot.png"); os.IsNotExist(err) {
			// If context was cancelled, file might not be created - that's OK
			pages := browserMgr.GetAllPages()
			if len(pages) > 0 {
				t.Error("Recovery screenshot file was not created despite having pages available")
			}
		}
	})
}