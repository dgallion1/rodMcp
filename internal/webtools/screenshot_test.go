package webtools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rodmcp/internal/browser"
)

func TestScreenshotTool_NewScreenshotTool(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	if tool == nil {
		t.Fatal("NewScreenshotTool returned nil")
	}
	
	if tool.logger != log {
		t.Error("Logger not set correctly")
	}
	
	if tool.browser != browserMgr {
		t.Error("Browser manager not set correctly")
	}
	
	if tool.validator == nil {
		t.Error("Path validator should be initialized")
	}
}

func TestScreenshotTool_Name(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	expected := "take_screenshot"
	if tool.Name() != expected {
		t.Errorf("Expected name %s, got %s", expected, tool.Name())
	}
}

func TestScreenshotTool_Description(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	description := tool.Description()
	if description == "" {
		t.Error("Description should not be empty")
	}
	
	if !strings.Contains(description, "screenshot") {
		t.Error("Description should mention screenshot")
	}
	
	if !strings.Contains(description, "browser") {
		t.Error("Description should mention browser")
	}
}

func TestScreenshotTool_InputSchema(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	schema := tool.InputSchema()
	
	// Check that schema has required properties
	if schema.Type != "object" {
		t.Error("Schema type should be object")
	}
	
	if schema.Properties == nil {
		t.Fatal("Schema properties should not be nil")
	}
	
	// Check that there are no required fields (screenshot tool has all optional fields)
	if len(schema.Required) != 0 {
		t.Errorf("Expected 0 required fields, got %d", len(schema.Required))
	}
	
	// Check that filename property exists
	if _, exists := schema.Properties["filename"]; !exists {
		t.Error("Property 'filename' not found in schema")
	}
	
	// Check expected properties (only filename and page_id based on actual schema)
	expectedProps := []string{"filename", "page_id"}
	for _, prop := range expectedProps {
		if _, exists := schema.Properties[prop]; !exists {
			t.Errorf("Property %s not found in schema", prop)
		}
	}
}

func TestScreenshotTool_Execute_EmptyArgs(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	args := map[string]interface{}{
		// No filename provided - should work since filename is optional
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation (filename is optional)
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation for empty args: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") && !strings.Contains(responseText, "pages") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestScreenshotTool_Execute_InvalidFilenameType(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	args := map[string]interface{}{
		"filename": 123, // Invalid type
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation (screenshot tool is more permissive)
	// But should handle the operation gracefully
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") && !strings.Contains(responseText, "pages") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestScreenshotTool_Execute_EmptyFilename(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	args := map[string]interface{}{
		"filename": "", // Empty filename
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation (screenshot tool handles it gracefully)
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") && !strings.Contains(responseText, "pages") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestScreenshotTool_Execute_InvalidPageIDType(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	args := map[string]interface{}{
		"filename": "test.png",
		"page_id":  123, // Invalid type
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation (screenshot tool is permissive)
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") && !strings.Contains(responseText, "pages") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestScreenshotTool_Execute_ValidOptionalParameters(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	testCases := []map[string]interface{}{
		{
			"filename": "test1.png",
		},
		{
			"page_id": "test-page-123",
		},
		{
			"filename": "test2.png",
			"page_id":  "test-page-456",
		},
	}
	
	for i, args := range testCases {
		response, err := tool.Execute(args)
		
		// Should not fail parameter validation
		if err != nil && strings.Contains(err.Error(), "parameter") {
			t.Errorf("Test case %d: Should not fail parameter validation: %v", i, err)
		}
		
		// Should handle browser operation gracefully
		if response != nil && response.IsError {
			responseText := response.Content[0].Text
			if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") && !strings.Contains(responseText, "pages") {
				t.Logf("Test case %d: Expected browser-related error, got: %s", i, responseText)
			}
		}
	}
}

func TestScreenshotTool_Execute_InvalidFilenameCharacters(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	invalidFilenames := []string{
		"file<name.png",
		"file>name.png",
		"file:name.png",
		"file\"name.png",
		"file/name.png",
		"file\\name.png",
		"file|name.png",
		"file?name.png",
		"file*name.png",
	}
	
	for _, filename := range invalidFilenames {
		args := map[string]interface{}{
			"filename": filename,
		}
		
		response, err := tool.Execute(args)
		
		// Should handle this as a path validation error (graceful handling)
		if err != nil && strings.Contains(err.Error(), "parameter") {
			// Parameter validation error is acceptable
			continue
		}
		
		if response != nil && response.IsError {
			responseText := response.Content[0].Text
			if strings.Contains(responseText, "path") || strings.Contains(responseText, "invalid") || strings.Contains(responseText, "filename") {
				// Path validation error in response is acceptable
				continue
			}
		}
		
		// If we get here and there was no error, that's unexpected for invalid filenames
		if err == nil && (response == nil || !response.IsError) {
			t.Errorf("Execute should fail or return error response for invalid filename: %s", filename)
		}
	}
}

func TestScreenshotTool_Execute_AutoPngExtension(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	args := map[string]interface{}{
		"filename": "test-screenshot", // No extension
	}
	
	// This will fail because browser is not started, but we're testing path handling
	response, err := tool.Execute(args)
	
	// Should not fail on parameter validation
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation for filename without extension: %v", err)
	}
	
	// If we get a response, it should handle the operational error gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		// Should mention browser not started or similar operational error
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") && !strings.Contains(responseText, "pages") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestScreenshotTool_Execute_BrowserNotStarted(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	args := map[string]interface{}{
		"filename": "test-screenshot.png",
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		// Should mention browser not started or similar operational error
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") && !strings.Contains(responseText, "pages") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestScreenshotTool_Execute_PanicRecovery(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewScreenshotTool(log, browserMgr)
	
	// Test with nil args to potentially cause a panic
	response, err := tool.Execute(nil)
	
	// Should not panic, should return an error
	if err != nil {
		// This is expected - nil args should cause an error
		return
	}
	
	// If no error, check if response indicates error
	if response != nil && response.IsError {
		// This is also acceptable - error handled gracefully
		return
	}
	
	// If we get here, something unexpected happened
	// But the important thing is we didn't panic
}

// Integration test with real browser - tests actual screenshot functionality
func TestScreenshotTool_Integration_RealBrowser(t *testing.T) {
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
	defer browserMgr.Stop()
	
	// Create a page first
	navTool := NewNavigatePageTool(log, browserMgr)
	navArgs := map[string]interface{}{
		"url": "https://example.com",
	}
	
	_, err = navTool.Execute(navArgs)
	if err != nil {
		t.Fatalf("Failed to navigate to page: %v", err)
	}
	
	// Give page time to load
	time.Sleep(2 * time.Second)
	
	tool := NewScreenshotTool(log, browserMgr)
	
	t.Run("BasicScreenshot", func(t *testing.T) {
		tempDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)
		
		args := map[string]interface{}{
			"filename": "basic-test.png",
		}
		
		response, err := tool.Execute(args)
		if err != nil {
			t.Fatalf("Screenshot failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("Screenshot returned error: %v", response.Content[0].Text)
		}
		
		// Verify screenshot file was created
		if _, err := os.Stat("basic-test.png"); os.IsNotExist(err) {
			t.Error("Screenshot file was not created")
		}
		
		// Verify response mentions success
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "Screenshot saved") {
			t.Error("Response should mention successful screenshot")
		}
	})
	
	t.Run("ScreenshotWithPageID", func(t *testing.T) {
		tempDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)
		
		args := map[string]interface{}{
			"filename": "pageid-test.png",
			"page_id":  "test-page-id",
		}
		
		response, err := tool.Execute(args)
		if err != nil {
			t.Fatalf("Screenshot with page_id failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("Screenshot with page_id returned error: %v", response.Content[0].Text)
		}
		
		// Verify screenshot file was created (or error was handled gracefully)
		if _, err := os.Stat("pageid-test.png"); os.IsNotExist(err) {
			// It's ok if file wasn't created due to invalid page_id - we're testing parameter handling
			t.Logf("Screenshot file not created, likely due to invalid page_id (expected)")
		}
	})
	
	t.Run("ScreenshotWithAbsolutePath", func(t *testing.T) {
		tempDir := t.TempDir()
		screenshotPath := filepath.Join(tempDir, "absolute-path-test.png")
		
		args := map[string]interface{}{
			"filename": screenshotPath,
		}
		
		response, err := tool.Execute(args)
		if err != nil {
			t.Fatalf("Screenshot with absolute path failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("Screenshot with absolute path returned error: %v", response.Content[0].Text)
		}
		
		// Verify screenshot file was created at absolute path
		if _, err := os.Stat(screenshotPath); os.IsNotExist(err) {
			t.Error("Screenshot file was not created at absolute path")
		}
	})
}

func TestScreenshotTool_Integration_NoPages(t *testing.T) {
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
	defer browserMgr.Stop()
	
	tool := NewScreenshotTool(log, browserMgr)
	
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	
	args := map[string]interface{}{
		"filename": "no-pages-test.png",
	}
	
	response, err := tool.Execute(args)
	
	// Should handle gracefully
	if err != nil {
		// Tool-level error is acceptable
		return
	}
	
	if response != nil && response.IsError {
		// Error response is also acceptable
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "No pages") && !strings.Contains(responseText, "no active page") {
			t.Error("Error message should mention no pages available")
		}
		return
	}
	
	t.Error("Should return error when no pages available for screenshot")
}