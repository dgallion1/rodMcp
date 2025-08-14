package webtools

import (
	"strings"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:       "selector",
		Value:       "",
		Issue:       "selector cannot be empty",
		Context:     "CSS selectors are required",
		Suggestions: []string{"Use #id", "Use .class"},
		Examples:    []string{"#submit-button", ".btn-primary"},
		HelpTopic:   "click_element",
	}
	
	errStr := err.Error()
	
	if !strings.Contains(errStr, "selector parameter error") {
		t.Error("Error should contain field parameter error")
	}
	
	if !strings.Contains(errStr, "selector cannot be empty") {
		t.Error("Error should contain issue description")
	}
	
	if !strings.Contains(errStr, "Context: CSS selectors are required") {
		t.Error("Error should contain context")
	}
	
	if !strings.Contains(errStr, "Suggestions: Use #id, Use .class") {
		t.Error("Error should contain suggestions")
	}
	
	if !strings.Contains(errStr, "Examples: #submit-button, .btn-primary") {
		t.Error("Error should contain examples")
	}
	
	if !strings.Contains(errStr, "Use 'help click_element' for more guidance") {
		t.Error("Error should contain help topic")
	}
}

func TestValidationError_ErrorMinimal(t *testing.T) {
	err := &ValidationError{
		Field: "test",
		Issue: "test issue",
	}
	
	errStr := err.Error()
	expected := "test parameter error: test issue"
	
	if errStr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, errStr)
	}
}

func TestValidateSelector_Empty(t *testing.T) {
	err := ValidateSelector("", "test_tool")
	
	if err == nil {
		t.Error("Empty selector should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if valErr.Field != "selector" {
		t.Errorf("Expected field 'selector', got '%s'", valErr.Field)
	}
	
	if !strings.Contains(valErr.Issue, "cannot be empty") {
		t.Error("Issue should mention empty selector")
	}
	
	if len(valErr.Suggestions) == 0 {
		t.Error("Should provide suggestions for empty selector")
	}
	
	if len(valErr.Examples) == 0 {
		t.Error("Should provide examples for empty selector")
	}
}

func TestValidateSelector_ValidSelectors(t *testing.T) {
	validSelectors := []string{
		"#submit-button",
		".btn-primary",
		"input[name='email']",
		"div > p",
		".parent .child",
		"//button[text()='Login']",
		"//div[@class='content']",
		"button[type='submit']",
		"form .error-message",
	}
	
	for _, selector := range validSelectors {
		err := ValidateSelector(selector, "test_tool")
		if err != nil {
			t.Errorf("Valid selector '%s' should not return error: %v", selector, err)
		}
	}
}

func TestValidateSelector_ExtraSpaces(t *testing.T) {
	err := ValidateSelector(".parent  .child", "test_tool")
	
	if err == nil {
		t.Error("Selector with extra spaces should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if !strings.Contains(valErr.Issue, "extra spaces") {
		t.Error("Issue should mention extra spaces")
	}
}

func TestValidateSelector_IncompleteXPath(t *testing.T) {
	err := ValidateSelector("//button", "test_tool")
	
	if err == nil {
		t.Error("Incomplete XPath should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if !strings.Contains(valErr.Issue, "incomplete") {
		t.Error("Issue should mention incomplete XPath")
	}
	
	if len(valErr.Suggestions) == 0 {
		t.Error("Should provide suggestions for incomplete XPath")
	}
}

func TestValidateSelector_ValidXPath(t *testing.T) {
	validXPaths := []string{
		"//button[@id='submit']",
		"//span[text()='Click me']",
		"//div[contains(@class, 'error')]",
		"//input[@name='email']",
		"//button[text()='Login']",
	}
	
	for _, xpath := range validXPaths {
		err := ValidateSelector(xpath, "test_tool")
		if err != nil {
			t.Errorf("Valid XPath '%s' should not return error: %v", xpath, err)
		}
	}
}

func TestValidateURL_Empty(t *testing.T) {
	err := ValidateURL("", "test_tool")
	
	if err == nil {
		t.Error("Empty URL should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if valErr.Field != "url" {
		t.Errorf("Expected field 'url', got '%s'", valErr.Field)
	}
	
	if !strings.Contains(valErr.Issue, "cannot be empty") {
		t.Error("Issue should mention empty URL")
	}
}

func TestValidateURL_ValidURLs(t *testing.T) {
	validURLs := []string{
		"https://example.com",
		"http://localhost:8080",
		"file:///path/to/file.html",
		"./index.html",
		"../page.html",
		"localhost:3000",
		"https://example.com/path?query=value",
	}
	
	for _, url := range validURLs {
		err := ValidateURL(url, "test_tool")
		if err != nil {
			t.Errorf("Valid URL '%s' should not return error: %v", url, err)
		}
	}
}

func TestValidateURL_Spaces(t *testing.T) {
	err := ValidateURL("https://example.com/my page", "test_tool")
	
	if err == nil {
		t.Error("URL with spaces should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if !strings.Contains(valErr.Issue, "spaces") {
		t.Error("Issue should mention spaces in URL")
	}
}

func TestValidateURL_MissingProtocol(t *testing.T) {
	err := ValidateURL("example.com", "test_tool")
	
	if err == nil {
		t.Error("URL without protocol should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if !strings.Contains(valErr.Issue, "missing protocol") {
		t.Error("Issue should mention missing protocol")
	}
}

func TestValidateText_Empty_NotAllowed(t *testing.T) {
	err := ValidateText("", "test_tool", false)
	
	if err == nil {
		t.Error("Empty text should return error when not allowed")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if valErr.Field != "text" {
		t.Errorf("Expected field 'text', got '%s'", valErr.Field)
	}
}

func TestValidateText_Empty_Allowed(t *testing.T) {
	err := ValidateText("", "test_tool", true)
	
	if err != nil {
		t.Errorf("Empty text should not return error when allowed: %v", err)
	}
}

func TestValidateText_ValidText(t *testing.T) {
	validTexts := []string{
		"Hello World",
		"user@example.com",
		"Line 1\nLine 2",
		"Special chars: !@#$%^&*()",
		"123456",
		"Test text with spaces",
	}
	
	for _, text := range validTexts {
		err := ValidateText(text, "test_tool", false)
		if err != nil {
			t.Errorf("Valid text '%s' should not return error: %v", text, err)
		}
	}
}

func TestValidateTimeout_ValidIntegers(t *testing.T) {
	validTimeouts := []int{1, 5, 10, 30, 60, 300}
	
	for _, timeout := range validTimeouts {
		result, err := ValidateTimeout(timeout, "test_tool")
		if err != nil {
			t.Errorf("Valid timeout %d should not return error: %v", timeout, err)
		}
		
		if result != timeout {
			t.Errorf("Expected timeout %d, got %d", timeout, result)
		}
	}
}

func TestValidateTimeout_ValidFloats(t *testing.T) {
	validTimeouts := []float64{1.0, 5.5, 10.9, 30.0}
	
	for _, timeout := range validTimeouts {
		result, err := ValidateTimeout(timeout, "test_tool")
		if err != nil {
			t.Errorf("Valid timeout %f should not return error: %v", timeout, err)
		}
		
		expected := int(timeout)
		if result != expected {
			t.Errorf("Expected timeout %d, got %d", expected, result)
		}
	}
}

func TestValidateTimeout_String(t *testing.T) {
	_, err := ValidateTimeout("5", "test_tool")
	
	if err == nil {
		t.Error("String timeout should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if !strings.Contains(valErr.Issue, "must be a number, not a string") {
		t.Error("Issue should mention string type error")
	}
}

func TestValidateTimeout_InvalidType(t *testing.T) {
	_, err := ValidateTimeout([]int{5}, "test_tool")
	
	if err == nil {
		t.Error("Invalid type timeout should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if !strings.Contains(valErr.Issue, "must be a number") {
		t.Error("Issue should mention type error")
	}
}

func TestValidateTimeout_TooSmall(t *testing.T) {
	_, err := ValidateTimeout(0, "test_tool")
	
	if err == nil {
		t.Error("Zero timeout should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if !strings.Contains(valErr.Issue, "at least 1 second") {
		t.Error("Issue should mention minimum timeout")
	}
}

func TestValidateTimeout_TooLarge(t *testing.T) {
	_, err := ValidateTimeout(500, "test_tool")
	
	if err == nil {
		t.Error("Very large timeout should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if !strings.Contains(valErr.Issue, "unusually long") {
		t.Error("Issue should mention timeout being too long")
	}
}

func TestValidateFilename_Empty(t *testing.T) {
	err := ValidateFilename("", "test_tool")
	
	if err == nil {
		t.Error("Empty filename should return error")
	}
	
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Error("Error should be ValidationError type")
	}
	
	if valErr.Field != "filename" {
		t.Errorf("Expected field 'filename', got '%s'", valErr.Field)
	}
}

func TestValidateFilename_ValidFilenames(t *testing.T) {
	validFilenames := []string{
		"test.html",
		"landing-page.html",
		"contact_form",
		"dashboard.html",
		"my-app",
		"file123",
		"index.html",
	}
	
	for _, filename := range validFilenames {
		err := ValidateFilename(filename, "test_tool")
		if err != nil {
			t.Errorf("Valid filename '%s' should not return error: %v", filename, err)
		}
	}
}

func TestValidateFilename_InvalidCharacters(t *testing.T) {
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
		err := ValidateFilename(filename, "test_tool")
		if err == nil {
			t.Errorf("Invalid filename '%s' should return error", filename)
		}
		
		valErr, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Error for '%s' should be ValidationError type", filename)
		}
		
		if !strings.Contains(valErr.Issue, "invalid characters") {
			t.Errorf("Error for '%s' should mention invalid characters", filename)
		}
	}
}

// Benchmark tests
func BenchmarkValidateSelector(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateSelector("#test-button", "test_tool")
	}
}

func BenchmarkValidateURL(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateURL("https://example.com", "test_tool")
	}
}

func BenchmarkValidateText(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateText("Test text", "test_tool", false)
	}
}

func BenchmarkValidateTimeout(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateTimeout(10, "test_tool")
	}
}

func BenchmarkValidateFilename(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateFilename("test.html", "test_tool")
	}
}

// Test edge cases
func TestValidateSelector_EdgeCases(t *testing.T) {
	testCases := []struct {
		selector  string
		shouldErr bool
		errPart   string
	}{
		{"#test", false, ""}, // Simple selector OK
		{".class1.class2", false, ""}, // Multiple classes OK
		{"div#id.class", false, ""}, // Combined selectors OK
		{"[data-test='value with spaces']", false, ""}, // Attribute values with spaces OK
		{"div:nth-child(2n+1)", false, ""}, // Pseudo selectors OK
		{".parent   .child", true, "extra spaces"}, // Multiple spaces should error
		{"//div[contains(@class,  'test')]", true, "extra spaces"}, // XPath with extra spaces
	}
	
	for _, tc := range testCases {
		err := ValidateSelector(tc.selector, "test_tool")
		
		if tc.shouldErr && err == nil {
			t.Errorf("Selector '%s' should return error", tc.selector)
		} else if !tc.shouldErr && err != nil {
			t.Errorf("Selector '%s' should not return error: %v", tc.selector, err)
		} else if tc.shouldErr && err != nil {
			if !strings.Contains(err.Error(), tc.errPart) {
				t.Errorf("Error for selector '%s' should contain '%s', got: %v", 
					tc.selector, tc.errPart, err)
			}
		}
	}
}

func TestValidateURL_EdgeCases(t *testing.T) {
	testCases := []struct {
		url       string
		shouldErr bool
		errPart   string
	}{
		{"https://example.com:8080", false, ""}, // Port in URL OK
		{"http://user:pass@example.com", false, ""}, // Auth in URL OK
		{"file:///absolute/path/file.html", false, ""}, // Absolute file path OK
		{"invalidurl", true, "missing protocol"}, // Invalid URL without protocol
		{"://example.com", true, "missing protocol"}, // Missing protocol
		{"data:text/html,<html></html>", true, "missing protocol"}, // Data URL should error
	}
	
	for _, tc := range testCases {
		err := ValidateURL(tc.url, "test_tool")
		
		if tc.shouldErr && err == nil {
			t.Errorf("URL '%s' should return error", tc.url)
		} else if !tc.shouldErr && err != nil {
			t.Errorf("URL '%s' should not return error: %v", tc.url, err)
		} else if tc.shouldErr && err != nil {
			if !strings.Contains(err.Error(), tc.errPart) {
				t.Errorf("Error for URL '%s' should contain '%s', got: %v", 
					tc.url, tc.errPart, err)
			}
		}
	}
}