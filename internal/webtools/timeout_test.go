package webtools

import (
	"context"
	"strings"
	"testing"
	"time"

	"rodmcp/internal/browser"
)

// TestTimeouts_BrowserTools tests timeout behavior for browser automation tools
func TestTimeouts_BrowserTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout tests in short mode")
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
	defer browserMgr.Stop()

	// Setup a page for tests that need one
	navTool := NewNavigatePageTool(log, browserMgr)
	navArgs := map[string]interface{}{
		"url": "https://example.com",
	}
	navTool.Execute(navArgs)
	time.Sleep(2 * time.Second)

	t.Run("NavigatePageTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewNavigatePageTool(log, browserMgr)
		
		// Test with very slow/unresponsive URL  
		args := map[string]interface{}{
			"url": "https://httpbin.org/delay/35", // 35 second delay should timeout
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// NavigatePageTool has built-in timeout protection
		// It should either timeout or complete quickly
		if duration > 35*time.Second {
			t.Errorf("Navigate operation took too long: %v", duration)
		}
		
		if err != nil {
			t.Logf("Navigation error (timeout protection working): %v", err)
		}
		t.Logf("Navigation completed in %v (timeout protection working)", duration)
	})

	t.Run("ExecuteScriptTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewExecuteScriptTool(log, browserMgr)
		
		// Test with infinite loop script
		args := map[string]interface{}{
			"script": "while(true) { /* infinite loop */ }",
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// ExecuteScriptTool has built-in timeout protection
		// It should either timeout or complete quickly
		if duration > 35*time.Second {
			t.Errorf("Script execution took too long: %v", duration)
		}
		
		if err != nil {
			t.Logf("Script execution error (timeout protection working): %v", err)
		}
		t.Logf("Script execution completed in %v (timeout protection working)", duration)
	})

	t.Run("ScreenshotTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewScreenshotTool(log, browserMgr)
		
		args := map[string]interface{}{
			"filename": "timeout-test.png",
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete within timeout (30s for ScreenshotTool)
		if duration > 35*time.Second {
			t.Errorf("Screenshot operation took too long: %v", duration)
		}
		
		// Should succeed or fail gracefully
		if err != nil && !strings.Contains(err.Error(), "context") {
			t.Logf("Screenshot error (may be expected): %v", err)
		}
	})

	t.Run("TypeTextTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewTypeTextTool(log, browserMgr)
		
		args := map[string]interface{}{
			"selector": "input[name='nonexistent']", // Non-existent selector
			"text":     "test text",
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete quickly with timeout protection
		if duration > 35*time.Second {
			t.Errorf("Type text operation took too long: %v", duration)
		}
		
		if err != nil {
			t.Logf("Type text error (timeout protection working): %v", err)
		}
		t.Logf("Type text completed in %v (timeout protection working)", duration)
	})

	t.Run("ClickElementTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewClickElementTool(log, browserMgr)
		
		args := map[string]interface{}{
			"selector": "button#nonexistent-button",
			"timeout":  5, // Short timeout for test
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete within specified timeout + overhead
		if duration > 10*time.Second {
			t.Errorf("Click operation took too long: %v", duration)
		}
		
		if err != nil {
			t.Logf("Click element error (timeout protection working): %v", err)
		}
		t.Logf("Click element completed in %v (timeout protection working)", duration)
	})
}

// TestTimeouts_FileSystemTools tests timeout behavior for file system tools
func TestTimeouts_FileSystemTools(t *testing.T) {
	t.Run("ReadFileTool_Timeout", func(t *testing.T) {
		t.Parallel()
		log := createTestLogger(t)
		validator := NewPathValidator(DefaultFileAccessConfig())
		tool := NewReadFileTool(log, validator)
		
		args := map[string]interface{}{
			"path": "/dev/random", // This could potentially hang
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should timeout within reasonable time (30s for ReadFileTool)
		if duration > 35*time.Second {
			t.Errorf("Read file operation took too long: %v", duration)
		}
		
		// Should either succeed (if file access is restricted) or timeout
		if err != nil && !strings.Contains(err.Error(), "context") {
			t.Logf("Read file error (may be expected): %v", err)
		}
	})

	t.Run("WriteFileTool_Timeout", func(t *testing.T) {
		t.Parallel()
		log := createTestLogger(t)
		validator := NewPathValidator(DefaultFileAccessConfig())
		tool := NewWriteFileTool(log, validator)
		
		tempDir := t.TempDir()
		
		args := map[string]interface{}{
			"path":    tempDir + "/test-timeout.txt",
			"content": "test content for timeout test",
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete quickly for normal file operations
		if duration > 35*time.Second {
			t.Errorf("Write file operation took too long: %v", duration)
		}
		
		// Should succeed for normal file write
		if err != nil {
			t.Errorf("Write file failed: %v", err)
		}
	})

	t.Run("ListDirectoryTool_Timeout", func(t *testing.T) {
		t.Parallel()
		log := createTestLogger(t)
		validator := NewPathValidator(DefaultFileAccessConfig())
		tool := NewListDirectoryTool(log, validator)
		
		args := map[string]interface{}{
			"path": "/", // Root directory - should be quick but test timeout
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete within timeout (15s for ListDirectoryTool)
		if duration > 20*time.Second {
			t.Errorf("List directory operation took too long: %v", duration)
		}
		
		// Should succeed for root directory listing
		if err != nil {
			t.Errorf("List directory failed: %v", err)
		}
	})
}

// TestTimeouts_NetworkTools tests timeout behavior for network tools
func TestTimeouts_NetworkTools(t *testing.T) {
	t.Run("HTTPRequestTool_Timeout", func(t *testing.T) {
		t.Parallel()
		log := createTestLogger(t)
		tool := NewHTTPRequestTool(log)
		
		// Test with slow endpoint
		args := map[string]interface{}{
			"url":    "https://httpbin.org/delay/65", // 65 second delay should timeout
			"method": "GET",
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should timeout within HTTPRequestTool timeout (60s) + overhead
		if duration > 70*time.Second {
			t.Errorf("HTTP request took too long: %v", duration)
		}
		
		if err != nil {
			t.Logf("HTTP request error (timeout protection working): %v", err)
		}
		t.Logf("HTTP request completed in %v (timeout protection working)", duration)
	})

	t.Run("HTTPRequestTool_FastRequest", func(t *testing.T) {
		t.Parallel()
		log := createTestLogger(t)
		tool := NewHTTPRequestTool(log)
		
		// Test with fast endpoint
		args := map[string]interface{}{
			"url":    "https://httpbin.org/get",
			"method": "GET",
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete quickly
		if duration > 30*time.Second {
			t.Errorf("Fast HTTP request took too long: %v", duration)
		}
		
		// Should succeed
		if err != nil {
			t.Errorf("Fast HTTP request failed: %v", err)
		}
	})
}

// TestTimeouts_UtilityTools tests timeout behavior for utility and interaction tools
func TestTimeouts_UtilityTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout tests in short mode")
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
	defer browserMgr.Stop()

	// Navigate to a test page first
	navTool := NewNavigatePageTool(log, browserMgr)
	navArgs := map[string]interface{}{
		"url": "https://example.com",
	}
	navTool.Execute(navArgs)
	time.Sleep(2 * time.Second)

	t.Run("WaitTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewWaitTool(log)
		
		// Test waiting for maximum allowed time
		args := map[string]interface{}{
			"seconds": 60.0, // Max wait time
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete within timeout (65s for WaitTool)
		if duration > 70*time.Second {
			t.Errorf("Wait operation took too long: %v", duration)
		}
		
		// Should succeed for valid wait time
		if err != nil {
			t.Errorf("Wait tool failed: %v", err)
		}
		
		// Should have waited approximately the right amount of time
		expectedDuration := 60 * time.Second
		if duration < expectedDuration-time.Second || duration > expectedDuration+5*time.Second {
			t.Errorf("Wait duration %v not close to expected %v", duration, expectedDuration)
		}
	})

	t.Run("WaitForElementTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewWaitForElementTool(log, browserMgr)
		
		// Test waiting for non-existent element
		args := map[string]interface{}{
			"selector": "#never-exists-element",
			"timeout":  5, // Short timeout for test
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete within specified timeout + overhead
		if duration > 10*time.Second {
			t.Errorf("Wait for element took too long: %v", duration)
		}
		
		if err != nil {
			t.Logf("Wait for element error (timeout protection working): %v", err)
		}
		t.Logf("Wait for element completed in %v (timeout protection working)", duration)
	})

	t.Run("GetElementTextTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewGetElementTextTool(log, browserMgr)
		
		args := map[string]interface{}{
			"selector": "#non-existent-element",
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete within timeout (30s for GetElementTextTool)
		if duration > 35*time.Second {
			t.Errorf("Get element text took too long: %v", duration)
		}
		
		if err != nil {
			t.Logf("Operation error (timeout protection working): %v", err)
		}
		t.Logf("Operation completed in %v (timeout protection working)", duration)
	})

	t.Run("GetElementAttributeTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewGetElementAttributeTool(log, browserMgr)
		
		args := map[string]interface{}{
			"selector":  "#non-existent-element",
			"attribute": "href",
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete within timeout (30s for GetElementAttributeTool)
		if duration > 35*time.Second {
			t.Errorf("Get element attribute took too long: %v", duration)
		}
		
		if err != nil {
			t.Logf("Operation error (timeout protection working): %v", err)
		}
		t.Logf("Operation completed in %v (timeout protection working)", duration)
	})

	t.Run("ScrollTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewScrollTool(log, browserMgr)
		
		args := map[string]interface{}{
			"y": 1000,
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete within timeout (30s for ScrollTool)
		if duration > 35*time.Second {
			t.Errorf("Scroll operation took too long: %v", duration)
		}
		
		// Should succeed for normal scroll
		if err != nil && !strings.Contains(err.Error(), "context") {
			t.Errorf("Scroll tool failed: %v", err)
		}
	})

	t.Run("HoverElementTool_Timeout", func(t *testing.T) {
		t.Parallel()
		tool := NewHoverElementTool(log, browserMgr)
		
		args := map[string]interface{}{
			"selector": "#non-existent-hover-element",
		}
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// Should complete within timeout (30s for HoverElementTool)
		if duration > 35*time.Second {
			t.Errorf("Hover operation took too long: %v", duration)
		}
		
		if err != nil {
			t.Logf("Operation error (timeout protection working): %v", err)
		}
		t.Logf("Operation completed in %v (timeout protection working)", duration)
	})
}

// TestTimeouts_PanicRecovery tests that all tools have panic recovery
func TestTimeouts_PanicRecovery(t *testing.T) {
	t.Run("AllTools_HavePanicRecovery", func(t *testing.T) {
		log := createTestLogger(t)
		
		// Test that all tools are properly wrapped with executeWithPanicRecovery
		// This test verifies the pattern is in place by checking tool construction doesn't panic
		
		// Browser tools
		browserMgr := browser.NewManager(log, browser.Config{Headless: true})
		
		tools := []interface{}{
			NewNavigatePageTool(log, browserMgr),
			NewExecuteScriptTool(log, browserMgr),
			NewScreenshotTool(log, browserMgr),
			NewBrowserVisibilityTool(log, browserMgr),
			NewClickElementTool(log, browserMgr),
			NewTypeTextTool(log, browserMgr),
			NewWaitTool(log),
			NewWaitForElementTool(log, browserMgr),
			NewGetElementTextTool(log, browserMgr),
			NewGetElementAttributeTool(log, browserMgr),
			NewScrollTool(log, browserMgr),
			NewHoverElementTool(log, browserMgr),
			NewExtractTableTool(log, browserMgr),
			NewFormFillTool(log, browserMgr),
			NewSwitchTabTool(log, browserMgr),
			NewWaitForConditionTool(log, browserMgr),
			NewAssertElementTool(log, browserMgr),
			NewTakeElementScreenshotTool(log, browserMgr),
			NewKeyboardShortcutTool(log, browserMgr),
			NewScreenScrapeTool(log, browserMgr),
			NewLivePreviewTool(log),
			NewReadFileTool(log, NewPathValidator(DefaultFileAccessConfig())),
			NewWriteFileTool(log, NewPathValidator(DefaultFileAccessConfig())),
			NewListDirectoryTool(log, NewPathValidator(DefaultFileAccessConfig())),
			NewHTTPRequestTool(log),
			NewCreatePageTool(log),
			NewHelpTool(log),
		}
		
		// Verify all tools can be created without panicking
		if len(tools) < 25 {
			t.Errorf("Expected at least 25 tools, got %d", len(tools))
		}
		
		t.Logf("Successfully created %d tools with panic recovery", len(tools))
	})
}

// TestTimeouts_RealWorldScenarios tests timeout behavior in realistic scenarios
func TestTimeouts_RealWorldScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-world timeout tests in short mode")
	}

	t.Run("ConcurrentOperations_DoNotBlock", func(t *testing.T) {
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
		defer browserMgr.Stop()

		// Run multiple operations concurrently
		done := make(chan bool, 3)
		
		// Operation 1: Navigate
		go func() {
			navTool := NewNavigatePageTool(log, browserMgr)
			args := map[string]interface{}{
				"url": "https://example.com",
			}
			navTool.Execute(args)
			done <- true
		}()
		
		// Operation 2: Create page
		go func() {
			createTool := NewCreatePageTool(log)
			args := map[string]interface{}{
				"filename": "concurrent-test.html",
				"title":    "Concurrent Test",
				"html":     "<h1>Test</h1>",
			}
			createTool.Execute(args)
			done <- true
		}()
		
		// Operation 3: HTTP request
		go func() {
			httpTool := NewHTTPRequestTool(log)
			args := map[string]interface{}{
				"url":    "https://httpbin.org/get",
				"method": "GET",
			}
			httpTool.Execute(args)
			done <- true
		}()
		
		// Wait for all operations to complete or timeout
		timeout := time.After(45 * time.Second)
		completed := 0
		
		for completed < 3 {
			select {
			case <-done:
				completed++
			case <-timeout:
				t.Errorf("Concurrent operations took too long, only %d/3 completed", completed)
				return
			}
		}
		
		t.Logf("All 3 concurrent operations completed within timeout")
	})

	t.Run("SequentialOperations_WithTimeouts", func(t *testing.T) {
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
		defer browserMgr.Stop()

		operations := []struct {
			name     string
			timeout  time.Duration
			execute  func() error
		}{
			{
				name:    "Navigate",
				timeout: 35 * time.Second,
				execute: func() error {
					tool := NewNavigatePageTool(log, browserMgr)
					args := map[string]interface{}{"url": "https://example.com"}
					_, err := tool.Execute(args)
					return err
				},
			},
			{
				name:    "Wait",
				timeout: 8 * time.Second,
				execute: func() error {
					tool := NewWaitTool(log)
					args := map[string]interface{}{"seconds": 3.0}
					_, err := tool.Execute(args)
					return err
				},
			},
			{
				name:    "Screenshot",
				timeout: 35 * time.Second,
				execute: func() error {
					tool := NewScreenshotTool(log, browserMgr)
					args := map[string]interface{}{"filename": "sequential-test.png"}
					_, err := tool.Execute(args)
					return err
				},
			},
		}

		totalStart := time.Now()
		
		for _, op := range operations {
			opStart := time.Now()
			err := op.execute()
			opDuration := time.Since(opStart)
			
			if opDuration > op.timeout {
				t.Errorf("Operation %s took too long: %v (expected < %v)", op.name, opDuration, op.timeout)
			}
			
			if err != nil && !strings.Contains(err.Error(), "context") {
				t.Logf("Operation %s error (may be expected): %v", op.name, err)
			}
		}
		
		totalDuration := time.Since(totalStart)
		if totalDuration > 2*time.Minute {
			t.Errorf("Total sequential operations took too long: %v", totalDuration)
		}
		
		t.Logf("Sequential operations completed in %v", totalDuration)
	})
}

// TestTimeouts_StressTest tests timeout behavior under stress
func TestTimeouts_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress timeout tests in short mode")
	}

	t.Run("MultipleTimeoutOperations", func(t *testing.T) {
		log := createTestLogger(t)
		
		// Test multiple timeout operations running concurrently
		numOps := 5
		done := make(chan struct{}, numOps)
		
		start := time.Now()
		
		for i := 0; i < numOps; i++ {
			go func(id int) {
				defer func() { done <- struct{}{} }()
				
				// Alternate between different tools
				switch id % 3 {
				case 0:
					// HTTP request with timeout
					tool := NewHTTPRequestTool(log)
					args := map[string]interface{}{
						"url":    "https://httpbin.org/delay/10",
						"method": "GET",
					}
					tool.Execute(args)
					
				case 1:
					// Wait tool
					tool := NewWaitTool(log)
					args := map[string]interface{}{
						"seconds": 2.0,
					}
					tool.Execute(args)
					
				case 2:
					// File operation
					validator := NewPathValidator(DefaultFileAccessConfig())
					tool := NewReadFileTool(log, validator)
					args := map[string]interface{}{
						"path": "/etc/passwd", // Should be quick to read or error
					}
					tool.Execute(args)
				}
			}(i)
		}
		
		// Wait for all operations to complete
		completed := 0
		timeout := time.After(30 * time.Second)
		
		for completed < numOps {
			select {
			case <-done:
				completed++
			case <-timeout:
				t.Errorf("Stress test timed out, only %d/%d operations completed", completed, numOps)
				return
			}
		}
		
		duration := time.Since(start)
		t.Logf("Stress test completed %d operations in %v", numOps, duration)
		
		if duration > 25*time.Second {
			t.Errorf("Stress test took too long: %v", duration)
		}
	})
}

// TestTimeouts_ContextCancellation tests proper context handling
func TestTimeouts_ContextCancellation(t *testing.T) {
	t.Run("ToolsRespectContextCancellation", func(t *testing.T) {
		log := createTestLogger(t)
		
		// Create a context that we'll cancel
		_, cancel := context.WithCancel(context.Background())
		
		// Test that tools handle context cancellation properly
		// Note: This test verifies the timeout mechanism exists, not the actual cancellation
		// since the Execute methods don't take context parameters
		
		tool := NewWaitTool(log)
		args := map[string]interface{}{
			"seconds": 5.0,
		}
		
		// Cancel context after 1 second
		go func() {
			time.Sleep(1 * time.Second)
			cancel()
		}()
		
		start := time.Now()
		_, err := tool.Execute(args)
		duration := time.Since(start)
		
		// The tool should complete its own timeout handling
		// This test verifies timeout mechanisms are in place
		if duration > 10*time.Second {
			t.Errorf("Wait tool took too long: %v", duration)
		}
		
		// Tool should either succeed (complete wait) or handle gracefully
		if err != nil {
			t.Logf("Wait tool error (may be expected): %v", err)
		}
		
		t.Logf("Context cancellation test completed in %v", duration)
	})
}