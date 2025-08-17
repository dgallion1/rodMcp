package webtools

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError provides enhanced error context for LLMs
type ValidationError struct {
	Field       string
	Value       interface{}
	Issue       string
	Context     string
	Suggestions []string
	Examples    []string
	HelpTopic   string
}

func (e *ValidationError) Error() string {
	msg := fmt.Sprintf("%s parameter error: %s", e.Field, e.Issue)
	
	if e.Context != "" {
		msg += fmt.Sprintf(" Context: %s", e.Context)
	}
	
	if len(e.Suggestions) > 0 {
		msg += fmt.Sprintf(" Suggestions: %s", strings.Join(e.Suggestions, ", "))
	}
	
	if len(e.Examples) > 0 {
		msg += fmt.Sprintf(" Examples: %s", strings.Join(e.Examples, ", "))
	}
	
	if e.HelpTopic != "" {
		msg += fmt.Sprintf(" Use 'help %s' for more guidance", e.HelpTopic)
	}
	
	return msg
}

// ValidateSelector provides comprehensive CSS selector validation
func ValidateSelector(selector string, toolName string) error {
	if selector == "" {
		return &ValidationError{
			Field:   "selector",
			Value:   selector,
			Issue:   "selector cannot be empty",
			Context: "CSS selectors are required for element targeting",
			Suggestions: []string{
				"Use #id for unique elements",
				"Use .class for styled elements", 
				"Use tag[attribute] for semantic elements",
				"Use //text() for XPath text matching",
			},
			Examples: []string{
				"#submit-button",
				".btn-primary", 
				"input[name='email']",
				"//button[text()='Login']",
			},
			HelpTopic: toolName,
		}
	}
	
	// Check for common selector issues
	if strings.Contains(selector, "  ") {
		return &ValidationError{
			Field:   "selector",
			Value:   selector,
			Issue:   "selector contains extra spaces",
			Context: "CSS selectors should not have multiple consecutive spaces",
			Suggestions: []string{"Remove extra spaces", "Use single space for descendant selectors"},
			Examples:    []string{"'.parent .child' not '.parent  .child'"},
			HelpTopic:   toolName,
		}
	}
	
	// Validate common patterns
	if strings.HasPrefix(selector, ".") && strings.Contains(selector, " ") {
		// Class selector with descendants - this is fine
	} else if strings.HasPrefix(selector, "#") && strings.Contains(selector, " ") {
		// ID with descendants - this is fine  
	} else if strings.HasPrefix(selector, "//") {
		// XPath - validate basic structure
		if !strings.Contains(selector, "[") && !strings.Contains(selector, "text()") {
			return &ValidationError{
				Field:   "selector", 
				Value:   selector,
				Issue:   "XPath selector may be incomplete",
				Context: "XPath selectors should include attributes or text matching",
				Suggestions: []string{
					"Add attribute matching: [@attr='value']",
					"Add text matching: [text()='content']",
					"Add contains: [contains(text(), 'partial')]",
				},
				Examples: []string{
					"//button[@id='submit']",
					"//span[text()='Click me']", 
					"//div[contains(@class, 'error')]",
				},
				HelpTopic: toolName,
			}
		}
	}
	
	return nil
}

// ValidateURL validates URL formats and provides helpful suggestions
func ValidateURL(url string, toolName string) error {
	if url == "" {
		return &ValidationError{
			Field:   "url",
			Value:   url,
			Issue:   "url cannot be empty",
			Context: "A valid URL or file path is required for navigation",
			Suggestions: []string{
				"Use https:// for web URLs",
				"Use localhost:PORT for development servers",
				"Use file:// for local HTML files",
				"Use relative paths like './index.html'",
			},
			Examples: []string{
				"https://example.com",
				"localhost:3000",
				"file:///path/to/file.html",
				"./page.html",
			},
			HelpTopic: toolName,
		}
	}
	
	// Check for common URL issues
	if strings.Contains(url, " ") {
		return &ValidationError{
			Field:     "url",
			Value:     url,
			Issue:     "url contains spaces",
			Context:   "URLs should not contain spaces",
			Suggestions: []string{"Replace spaces with %20", "Use quotes if needed", "Check for copy-paste errors"},
			Examples:  []string{"'https://example.com/page' not 'https://example.com/my page'"},
			HelpTopic: toolName,
		}
	}
	
	// Validate protocol
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") && 
	   !strings.HasPrefix(url, "file://") && !strings.HasPrefix(url, "./") && 
	   !strings.HasPrefix(url, "../") && !strings.Contains(url, "localhost") &&
	   !strings.HasPrefix(url, "/") {
		return &ValidationError{
			Field:   "url",
			Value:   url,
			Issue:   "url may be missing protocol or be invalid",
			Context: "URLs should include protocol or be valid file paths",
			Suggestions: []string{
				"Add https:// for web URLs",
				"Use localhost:PORT for local servers",
				"Use ./ for relative paths",
				"Use file:// for absolute file paths",
			},
			Examples: []string{
				"https://example.com (not example.com)",
				"localhost:8080 (not :8080)",
				"./index.html (not index.html)",
			},
			HelpTopic: toolName,
		}
	}
	
	return nil
}

// ValidateText validates text input and provides suggestions
func ValidateText(text string, toolName string, allowEmpty bool) error {
	if text == "" && !allowEmpty {
		return &ValidationError{
			Field:   "text",
			Value:   text,
			Issue:   "text cannot be empty",
			Context: "Text content is required for typing operations",
			Suggestions: []string{
				"Provide the text content to type",
				"Use \\n for newlines in textarea fields",
				"Include special characters if needed",
			},
			Examples: []string{
				"user@example.com",
				"Hello\\nWorld (for multiline)",
				"Special chars: !@#$%",
			},
			HelpTopic: toolName,
		}
	}
	
	return nil
}

// ValidateTimeout validates timeout values and provides guidance
func ValidateTimeout(timeout interface{}, toolName string) (int, error) {
	var timeoutVal int
	
	switch v := timeout.(type) {
	case int:
		timeoutVal = v
	case float64:
		timeoutVal = int(v)
	case string:
		return 0, &ValidationError{
			Field:       "timeout",
			Value:       v,
			Issue:       "timeout must be a number, not a string",
			Context:     "Timeout values should be integers representing seconds",
			Suggestions: []string{"Use numbers like 5, 10, 30", "Don't use quotes around timeout values"},
			Examples:    []string{"5 (not '5')", "10 (not 'ten')"},
			HelpTopic:   toolName,
		}
	default:
		return 0, &ValidationError{
			Field:       "timeout",
			Value:       v,
			Issue:       fmt.Sprintf("timeout must be a number, got %T", v),
			Context:     "Timeout values should be integers representing seconds",
			Suggestions: []string{"Use positive integers", "Choose appropriate timeouts based on content type"},
			Examples:    []string{"5 (basic elements)", "10 (forms/dynamic content)", "30 (heavy AJAX)"},
			HelpTopic:   toolName,
		}
	}
	
	if timeoutVal < 1 {
		return 0, &ValidationError{
			Field:       "timeout",
			Value:       timeoutVal,
			Issue:       "timeout must be at least 1 second",
			Context:     "Very short timeouts may cause elements to not be found",
			Suggestions: []string{"Use minimum 1 second", "Increase timeout for dynamic content"},
			Examples:    []string{"1 (minimum)", "5 (typical)", "10 (dynamic content)"},
			HelpTopic:   toolName,
		}
	}
	
	if timeoutVal > 300 {
		return 0, &ValidationError{
			Field:       "timeout",
			Value:       timeoutVal,
			Issue:       "timeout seems unusually long (>5 minutes)",
			Context:     "Very long timeouts may indicate an issue with the approach",
			Suggestions: []string{
				"Consider using wait_for_condition instead",
				"Check if element selector is correct",
				"Verify page is loading properly",
			},
			Examples: []string{"30 (typical maximum)", "60 (very slow loading)", "Use wait_for_condition for complex conditions"},
			HelpTopic: toolName,
		}
	}
	
	return timeoutVal, nil
}

// ValidateFilename validates file names and paths
func ValidateFilename(filename string, toolName string) error {
	if filename == "" {
		return &ValidationError{
			Field:   "filename",
			Value:   filename,
			Issue:   "filename cannot be empty",
			Context: "A filename is required for file operations",
			Suggestions: []string{
				"Use descriptive names",
				"Include .html extension for web pages",
				"Use hyphens instead of spaces",
			},
			Examples: []string{
				"landing-page.html",
				"contact-form",
				"dashboard.html",
			},
			HelpTopic: toolName,
		}
	}
	
	// Check for invalid characters (basic validation)
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	if invalidChars.MatchString(filename) {
		return &ValidationError{
			Field:   "filename",
			Value:   filename,
			Issue:   "filename contains invalid characters",
			Context: "Filenames should not contain special characters that are invalid in file systems",
			Suggestions: []string{
				"Use only letters, numbers, hyphens, and underscores",
				"Replace spaces with hyphens",
				"Avoid: < > : \" / \\ | ? *",
			},
			Examples: []string{
				"my-page.html (not my page.html)",
				"contact_form (not contact/form)",
			},
			HelpTopic: toolName,
		}
	}
	
	return nil
}