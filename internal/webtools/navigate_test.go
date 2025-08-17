package webtools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rodmcp/internal/browser"
)

func TestNavigatePageTool_NewNavigatePageTool(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	if tool == nil {
		t.Fatal("NewNavigatePageTool returned nil")
	}
	
	if tool.logger != log {
		t.Error("Logger not set correctly")
	}
	
	if tool.browser != browserMgr {
		t.Error("Browser manager not set correctly")
	}
}

func TestNavigatePageTool_Name(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	expected := "navigate_page"
	if tool.Name() != expected {
		t.Errorf("Expected name %s, got %s", expected, tool.Name())
	}
}

func TestNavigatePageTool_Description(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	description := tool.Description()
	if description == "" {
		t.Error("Description should not be empty")
	}
	
	if !strings.Contains(description, "Navigate") {
		t.Error("Description should mention Navigate")
	}
	
	if !strings.Contains(description, "URL") {
		t.Error("Description should mention URL")
	}
}

func TestNavigatePageTool_InputSchema(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	schema := tool.InputSchema()
	
	// Check that schema has required properties
	if schema.Type != "object" {
		t.Error("Schema type should be object")
	}
	
	if schema.Properties == nil {
		t.Fatal("Schema properties should not be nil")
	}
	
	// Check required fields
	expectedRequired := []string{"url"}
	if len(schema.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(schema.Required))
	}
	
	for _, field := range expectedRequired {
		found := false
		for _, req := range schema.Required {
			if req == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required field %s not found in schema", field)
		}
	}
	
	// Check that URL property exists
	if _, exists := schema.Properties["url"]; !exists {
		t.Error("Property 'url' not found in schema")
	}
}

func TestNavigatePageTool_Execute_MissingURL(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	args := map[string]interface{}{
		// No URL provided
	}
	
	_, err := tool.Execute(args)
	
	if err == nil {
		t.Error("Execute should fail when URL is missing")
	}
	
	if !strings.Contains(err.Error(), "url") {
		t.Error("Error should mention missing URL")
	}
}

func TestNavigatePageTool_Execute_InvalidURLType(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	args := map[string]interface{}{
		"url": 123, // Should be string
	}
	
	_, err := tool.Execute(args)
	
	if err == nil {
		t.Error("Execute should fail when URL is not a string")
	}
	
	if !strings.Contains(err.Error(), "url parameter must be a string") {
		t.Error("Error should mention URL type validation")
	}
}

func TestNavigatePageTool_Execute_EmptyURL(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	args := map[string]interface{}{
		"url": "", // Empty URL
	}
	
	_, err := tool.Execute(args)
	
	if err == nil {
		t.Error("Execute should fail when URL is empty")
	}
	
	if !strings.Contains(err.Error(), "url cannot be empty") {
		t.Error("Error should mention empty URL")
	}
}

func TestNavigatePageTool_Execute_BrowserNotStarted(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	args := map[string]interface{}{
		"url": "https://example.com",
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation for valid URL: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestNavigatePageTool_Execute_LocalFileHandling(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	// Create a temporary HTML file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.html")
	content := `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body><h1>Test</h1></body>
</html>`
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Use file:// protocol for absolute path
	args := map[string]interface{}{
		"url": "file://" + testFile,
	}
	
	// This will fail because browser manager is not started, but we're testing parameter validation
	response, err := tool.Execute(args)
	
	// Should not return error from parameter validation
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation for valid local file: %v", err)
	}
	
	// If we get a response, it should handle the error gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		// Should mention browser not started or similar operational error
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestNavigatePageTool_Execute_RelativePathHandling(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	// Create test file in current directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	
	content := `<!DOCTYPE html><html><body><h1>Test</h1></body></html>`
	err := os.WriteFile("relative-test.html", []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	args := map[string]interface{}{
		"url": "./relative-test.html", // Proper relative path format
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation for valid relative file: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestNavigatePageTool_Execute_HTTPSURLHandling(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	args := map[string]interface{}{
		"url": "https://example.com",
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation for valid HTTPS URL: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestNavigatePageTool_Execute_HTTPURLHandling(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	args := map[string]interface{}{
		"url": "http://example.com",
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation for valid HTTP URL: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestNavigatePageTool_Execute_FileURL(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
	args := map[string]interface{}{
		"url": "file:///tmp/test.html",
	}
	
	response, err := tool.Execute(args)
	
	// Should not fail parameter validation
	if err != nil && strings.Contains(err.Error(), "parameter") {
		t.Errorf("Should not fail parameter validation for valid file URL: %v", err)
	}
	
	// Should handle browser operation gracefully
	if response != nil && response.IsError {
		responseText := response.Content[0].Text
		if !strings.Contains(responseText, "browser") && !strings.Contains(responseText, "not started") {
			t.Logf("Expected browser-related error, got: %s", responseText)
		}
	}
}

func TestNavigatePageTool_Execute_PanicRecovery(t *testing.T) {
	log := createTestLogger(t)
	browserMgr := &browser.Manager{}
	tool := NewNavigatePageTool(log, browserMgr)
	
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

// Integration test with real browser - tests actual navigation functionality
func TestNavigatePageTool_Integration_RealBrowser(t *testing.T) {
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
	
	tool := NewNavigatePageTool(log, browserMgr)
	
	t.Run("NavigateToWebsite", func(t *testing.T) {
		args := map[string]interface{}{
			"url": "https://example.com",
		}
		
		response, err := tool.Execute(args)
		if err != nil {
			t.Fatalf("Navigate failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("Navigate returned error: %v", response.Content[0].Text)
		}
		
		// Give browser time to load
		time.Sleep(2 * time.Second)
		
		// Verify page is in browser
		pages := browserMgr.GetAllPages()
		if len(pages) == 0 {
			t.Error("No pages found in browser after navigation")
		} else {
			found := false
			for _, page := range pages {
				if strings.Contains(page.URL, "example.com") {
					found = true
					break
				}
			}
			if !found {
				t.Error("Example.com page not found in browser pages")
			}
		}
	})
	
	t.Run("NavigateToLocalFile", func(t *testing.T) {
		// Create a temporary HTML file
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "nav-test.html")
		content := `<!DOCTYPE html>
<html>
<head><title>Navigation Test Page</title></head>
<body><h1>Navigation Test</h1><p>This is a test page for navigation.</p></body>
</html>`
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		args := map[string]interface{}{
			"url": testFile,
		}
		
		response, err := tool.Execute(args)
		if err != nil {
			t.Fatalf("Navigate to local file failed: %v", err)
		}
		
		if response.IsError {
			t.Errorf("Navigate returned error: %v", response.Content[0].Text)
		}
		
		// Give browser time to load
		time.Sleep(1 * time.Second)
		
		// Verify page is in browser
		pages := browserMgr.GetAllPages()
		if len(pages) == 0 {
			t.Error("No pages found in browser after navigation")
		} else {
			found := false
			for _, page := range pages {
				if strings.Contains(page.URL, "nav-test.html") {
					found = true
					break
				}
			}
			if !found {
				t.Error("Local test page not found in browser pages")
			}
		}
	})
}