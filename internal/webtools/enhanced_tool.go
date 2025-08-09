package webtools

import "rodmcp/pkg/types"

// EnhancedTool extends the base Tool interface with help capabilities
type EnhancedTool interface {
	Name() string
	Description() string
	InputSchema() types.ToolSchema
	Execute(args map[string]interface{}) (*types.CallToolResponse, error)
	
	// Enhanced help methods
	GetUsageHint() UsageHint
	GetExamples() []ToolExample
	GetCommonWorkflows() []string
}

// ToolExample represents a concrete usage example
type ToolExample struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Expected    string
}

// Enhanced descriptions with context and examples
func GetEnhancedDescription(toolName string) string {
	descriptions := map[string]string{
		"create_page": `🎨 Create interactive HTML pages with embedded CSS and JavaScript
		
Purpose: Rapid prototyping, landing pages, demos, and testing
Examples: 
  • "Create a responsive coffee shop landing page"
  • "Build a contact form with validation" 
  • "Generate a dashboard with charts"
		
Perfect for: UI mockups, interactive demos, test pages`,

		"navigate_page": `🌐 Navigate browser to URLs or local files
		
Purpose: Load pages for testing, interaction, and analysis
Examples:
  • "Open localhost:3000 to test React app"
  • "Load test-form.html for UI testing"
  • "Navigate to production site for comparison"
		
Perfect for: Loading content, switching pages, testing workflows`,

		"execute_script": `⚡ Execute JavaScript code in browser pages
		
Purpose: DOM manipulation, testing, API calls, custom interactions
Examples:
  • "Test form validation on contact page"
  • "Extract all links from navigation menu" 
  • "Simulate button clicks and user workflows"
		
Perfect for: Testing JavaScript, data extraction, automation`,

		"click_element": `🖱️ Click buttons, links, and interactive elements
		
Purpose: User interaction simulation, form submission, navigation
Examples:
  • "Click 'Submit' button after filling form"
  • "Click navigation menu items to test routing"
  • "Trigger dropdown menus and modals"
		
Perfect for: Testing user flows, form interactions, navigation`,

		"live_preview": `🚀 Start local HTTP server for live development
		
Purpose: Serve static files, multi-page testing, development server
Examples:
  • "Serve website files at localhost:8080"
  • "Start preview for multi-page site testing"
  • "Create development server for static assets"
		
Perfect for: Local development, file serving, website testing`,

		"take_screenshot": `📸 Capture visual snapshots of web pages
		
Purpose: Visual validation, documentation, progress tracking
Examples:
  • "Take screenshot after form submission"
  • "Capture responsive design at mobile size"
  • "Document test results visually"
		
Perfect for: Visual testing, documentation, debugging`,
	}
	
	if desc, exists := descriptions[toolName]; exists {
		return desc
	}
	return "Tool description not available"
}

// Generate usage examples for tools
func GetToolExamples(toolName string) []ToolExample {
	examples := map[string][]ToolExample{
		"create_page": {
			{
				Name: "Landing Page",
				Description: "Create a responsive coffee shop landing page",
				Parameters: map[string]interface{}{
					"filename": "coffee-landing",
					"title": "Mountain View Coffee",
					"html": `<header><h1>Welcome to Mountain View Coffee</h1></header>
<main><section class="hero"><p>Premium coffee, mountain fresh</p></section></main>`,
					"css": `body{font-family:Arial;margin:0} .hero{text-align:center;padding:50px;background:#8B4513;color:white}`,
				},
				Expected: "Creates coffee-landing.html with responsive design",
			},
		},
		"execute_script": {
			{
				Name: "Form Validation Test",
				Description: "Test all form validation on the page",
				Parameters: map[string]interface{}{
					"script": `document.querySelectorAll('form').forEach(form => { 
  console.log('Testing form:', form.id || form.className);
  // Trigger validation
  form.checkValidity();
});`,
				},
				Expected: "Validates all forms and logs results to console",
			},
		},
	}
	
	if exs, exists := examples[toolName]; exists {
		return exs
	}
	return []ToolExample{}
}