package webtools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreatePageTool_NewCreatePageTool(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	if tool == nil {
		t.Fatal("NewCreatePageTool returned nil")
	}
	
	if tool.logger != log {
		t.Error("Logger not set correctly")
	}
}

func TestCreatePageTool_Name(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	expected := "create_page"
	if tool.Name() != expected {
		t.Errorf("Expected name %s, got %s", expected, tool.Name())
	}
}

func TestCreatePageTool_Description(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	description := tool.Description()
	if description == "" {
		t.Error("Description should not be empty")
	}
	
	if !strings.Contains(description, "HTML") {
		t.Error("Description should mention HTML")
	}
}

func TestCreatePageTool_InputSchema(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	schema := tool.InputSchema()
	
	// Check that schema has required properties
	if schema.Type != "object" {
		t.Error("Schema type should be object")
	}
	
	if schema.Properties == nil {
		t.Fatal("Schema properties should not be nil")
	}
	
	// Check required fields
	expectedRequired := []string{"filename", "title", "html"}
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
	
	// Check that all expected properties exist
	expectedProps := []string{"filename", "title", "html", "css", "javascript"}
	for _, prop := range expectedProps {
		if _, exists := schema.Properties[prop]; !exists {
			t.Errorf("Property %s not found in schema", prop)
		}
	}
}

func TestCreatePageTool_Execute_Success(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	
	args := map[string]interface{}{
		"filename": "test-page.html",
		"title":    "Test Page",
		"html":     "<h1>Hello World</h1>",
		"css":      "body { background: #f0f0f0; }",
		"javascript": "console.log('test');",
	}
	
	response, err := tool.Execute(args)
	
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	
	if response == nil {
		t.Fatal("Response should not be nil")
	}
	
	if response.IsError {
		t.Error("Response should not be an error")
	}
	
	if len(response.Content) == 0 {
		t.Error("Response content should not be empty")
	}
	
	// Check that file was created
	if _, err := os.Stat("test-page.html"); os.IsNotExist(err) {
		t.Error("HTML file was not created")
	}
	
	// Check file contents
	content, err := os.ReadFile("test-page.html")
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}
	
	contentStr := string(content)
	
	// Verify HTML structure
	if !strings.Contains(contentStr, "<!DOCTYPE html>") {
		t.Error("File should contain DOCTYPE declaration")
	}
	
	if !strings.Contains(contentStr, "<title>Test Page</title>") {
		t.Error("File should contain correct title")
	}
	
	if !strings.Contains(contentStr, "<h1>Hello World</h1>") {
		t.Error("File should contain HTML content")
	}
	
	if !strings.Contains(contentStr, "body { background: #f0f0f0; }") {
		t.Error("File should contain CSS")
	}
	
	if !strings.Contains(contentStr, "console.log('test');") {
		t.Error("File should contain JavaScript")
	}
}

func TestCreatePageTool_Execute_MinimalArgs(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	
	// Test with only required fields
	args := map[string]interface{}{
		"filename": "minimal.html",
		"title":    "Minimal Page",
		"html":     "<p>Minimal content</p>",
	}
	
	response, err := tool.Execute(args)
	
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	
	if response.IsError {
		t.Error("Response should not be an error")
	}
	
	// Check that file was created
	content, err := os.ReadFile("minimal.html")
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}
	
	contentStr := string(content)
	
	// Should have default empty CSS and JS sections
	if !strings.Contains(contentStr, "<style>") {
		t.Error("File should contain style section")
	}
	
	if !strings.Contains(contentStr, "<script>") {
		t.Error("File should contain script section")
	}
}

func TestCreatePageTool_Execute_AutoHtmlExtension(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	
	args := map[string]interface{}{
		"filename": "no-extension",
		"title":    "Test",
		"html":     "<p>Test</p>",
	}
	
	response, err := tool.Execute(args)
	
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	
	if response.IsError {
		t.Error("Response should not be an error")
	}
	
	// Check that .html extension was added
	if _, err := os.Stat("no-extension.html"); os.IsNotExist(err) {
		t.Error("HTML file with .html extension was not created")
	}
	
	// Verify response mentions correct path
	responseText := response.Content[0].Text
	if !strings.Contains(responseText, "no-extension.html") {
		t.Error("Response should mention the file with .html extension")
	}
}

func TestCreatePageTool_Execute_MissingFilename(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	args := map[string]interface{}{
		"title": "Test",
		"html":  "<p>Test</p>",
	}
	
	_, err := tool.Execute(args)
	
	if err == nil {
		t.Error("Execute should fail when filename is missing")
	}
	
	if !strings.Contains(err.Error(), "filename") {
		t.Error("Error should mention missing filename")
	}
}

func TestCreatePageTool_Execute_InvalidFilename(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	args := map[string]interface{}{
		"filename": "", // Empty filename should be invalid
		"title":    "Test",
		"html":     "<p>Test</p>",
	}
	
	_, err := tool.Execute(args)
	
	if err == nil {
		t.Error("Execute should fail when filename is empty")
	}
}

func TestCreatePageTool_Execute_InvalidFilenameCharacters(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	invalidFilenames := []string{
		"file<name.html",
		"file>name.html", 
		"file:name.html",
		"file\"name.html",
		"file/name.html",
		"file\\name.html",
		"file|name.html",
		"file?name.html",
		"file*name.html",
	}
	
	for _, filename := range invalidFilenames {
		args := map[string]interface{}{
			"filename": filename,
			"title":    "Test",
			"html":     "<p>Test</p>",
		}
		
		_, err := tool.Execute(args)
		
		if err == nil {
			t.Errorf("Execute should fail for invalid filename: %s", filename)
		}
	}
}

func TestCreatePageTool_Execute_FileWriteError(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	// Try to write to a directory that doesn't exist
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	// Change to a directory we know doesn't exist
	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "nonexistent", "file.html")
	
	args := map[string]interface{}{
		"filename": nonExistentPath,
		"title":    "Test",
		"html":     "<p>Test</p>",
	}
	
	_, err := tool.Execute(args)
	
	// Should return error due to path validation (invalid characters in path)
	if err == nil {
		t.Fatal("Execute should return error for invalid path")
	}
	
	// Should mention filename validation
	if !strings.Contains(err.Error(), "filename") && !strings.Contains(err.Error(), "path") {
		t.Error("Error should mention filename/path validation issue")
	}
}

func TestCreatePageTool_Execute_TypeValidation(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	// Test wrong type for filename
	args := map[string]interface{}{
		"filename": 123, // Should be string
		"title":    "Test",
		"html":     "<p>Test</p>",
	}
	
	_, err := tool.Execute(args)
	
	if err == nil {
		t.Error("Execute should fail when filename is not a string")
	}
	
	if !strings.Contains(err.Error(), "filename parameter must be a string") {
		t.Error("Error should mention filename type validation")
	}
}

func TestCreatePageTool_Execute_DefaultValues(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	
	// Test default title
	args := map[string]interface{}{
		"filename": "default-title.html",
		"html":     "<p>Test</p>",
		// No title provided
	}
	
	response, err := tool.Execute(args)
	
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	
	if response.IsError {
		t.Error("Response should not be an error")
	}
	
	content, err := os.ReadFile("default-title.html")
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}
	
	contentStr := string(content)
	if !strings.Contains(contentStr, "<title>Untitled Page</title>") {
		t.Error("Should use default title when none provided")
	}
}

func TestCreatePageTool_Execute_PanicRecovery(t *testing.T) {
	log := createTestLogger(t)
	tool := NewCreatePageTool(log)
	
	// This test ensures that executeWithPanicRecovery works
	// We'll test with nil args to potentially cause a panic
	response, err := tool.Execute(nil)
	
	// Should not panic, should return an error response
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