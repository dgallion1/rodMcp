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

		"form_fill": `📝 Complete form automation with validation and submission (🔥 NEW)
		
Purpose: Ultimate form handling for all input types and workflows
Examples: 
  • "Fill login form with username/password and submit"
  • "Complete checkout with billing/shipping information"
  • "Automate contact form with validation checking"
		
Perfect for: Registration flows, e-commerce, contact forms, multi-step wizards`,

		"wait_for_condition": `⚡ Wait for custom JavaScript conditions (🔥 NEW)
		
Purpose: Smart waiting for modern apps with dynamic content
Examples:
  • "Wait for React component state: window.appReady === true"
  • "Wait for API data: document.querySelectorAll('.item').length >= 10"  
  • "Wait for animation: element.style.opacity === '1'"
		
Perfect for: SPAs, animations, API responses, state changes`,

		"assert_element": `🧪 Comprehensive element testing framework (🔥 NEW)
		
Purpose: Professional testing with 15+ assertion types
Examples:
  • "Assert button is visible and enabled"
  • "Verify form field contains expected text"
  • "Check element has correct CSS class"
		
Perfect for: Automated testing, validation, QA workflows, UI verification`,

		"extract_table": `📊 Extract structured data from HTML tables (🔥 NEW)
		
Purpose: Convert HTML tables to structured data formats (JSON, CSV, arrays)
Examples:
  • "Extract product catalog table to JSON objects"
  • "Convert pricing table to CSV for analysis"
  • "Extract specific columns from data table"
		
Perfect for: Data scraping, report generation, table analysis, data export`,

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

		"take_element_screenshot": `📸 Capture screenshots of specific elements (🔥 NEW)
		
Purpose: Focused visual testing and element documentation
Examples:
  • "Screenshot just the navigation menu for testing"
  • "Capture error message element for bug reporting"
  • "Document specific form field validation states"
		
Perfect for: Element testing, UI debugging, component documentation`,

		"keyboard_shortcuts": `⌨️ Send keyboard combinations and special keys
		
Purpose: Form navigation, keyboard shortcuts, copy/paste operations
Examples:
  • "Navigate form fields with Tab/Shift+Tab"
  • "Copy text with Ctrl+C and paste with Ctrl+V"
  • "Refresh page with F5 or trigger DevTools with F12"
		
Perfect for: Keyboard automation, form navigation, shortcut testing`,

		"switch_tab": `🔄 Multi-tab workflow automation
		
Purpose: Create, switch between, and manage multiple browser tabs efficiently
Examples:
  • "Open comparison sites in multiple tabs and switch between them"
  • "Create new tabs for different test scenarios"
  • "Manage complex workflows across multiple pages"
		
Perfect for: Multi-tab testing, workflow automation, tab management`,
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
		
		"take_element_screenshot": {
			{
				Name: "Button Screenshot",
				Description: "Capture a specific button for testing documentation",
				Parameters: map[string]interface{}{
					"selector": "#submit-button",
					"filename": "submit-button.png",
					"padding": 15,
					"scroll_into_view": true,
				},
				Expected: "Saves screenshot of submit button with 15px padding",
			},
			{
				Name: "Error Message Capture",
				Description: "Screenshot validation error for bug reporting",
				Parameters: map[string]interface{}{
					"selector": ".error-message",
					"wait_for_element": true,
					"timeout": 5,
					"padding": 20,
				},
				Expected: "Captures error message element after waiting for visibility",
			},
			{
				Name: "Form Field Documentation",
				Description: "Document form field state for testing",
				Parameters: map[string]interface{}{
					"selector": "#email-field",
					"filename": "email-field-state.png",
					"scroll_into_view": false,
					"padding": 5,
				},
				Expected: "Screenshots email field without scrolling for documentation",
			},
		},
		
		"form_fill": {
			{
				Name: "Contact Form Automation",
				Description: "Fill out complete contact form with validation",
				Parameters: map[string]interface{}{
					"fields": map[string]interface{}{
						"#name": "John Doe",
						"#email": "john@example.com",
						"#message": "Hello! I'm interested in your services.",
						"select[name='department']": "sales",
						"input[name='newsletter']": true,
					},
					"submit": true,
					"validate_required": true,
				},
				Expected: "Fills all fields, validates required fields, and submits form",
			},
			{
				Name: "E-commerce Checkout",
				Description: "Complete checkout form for online purchase",
				Parameters: map[string]interface{}{
					"fields": map[string]interface{}{
						"#firstName": "Jane",
						"#lastName": "Smith", 
						"#email": "jane.smith@example.com",
						"#address": "123 Main St",
						"#city": "San Francisco",
						"select[name='state']": "CA",
						"#zipCode": "94102",
						"input[name='saveInfo']": false,
					},
					"validate_required": true,
					"trigger_events": true,
				},
				Expected: "Completes checkout form with billing information and validation",
			},
		},
		
		"wait_for_condition": {
			{
				Name: "API Response Waiting",
				Description: "Wait for API data to load in React app",
				Parameters: map[string]interface{}{
					"condition": "window.appState && window.appState.dataLoaded === true",
					"description": "Wait for React app data loading to complete",
					"timeout": 15,
					"interval": 200,
				},
				Expected: "Waits until React app state indicates data is loaded",
			},
			{
				Name: "Animation Completion",
				Description: "Wait for CSS animation to finish",
				Parameters: map[string]interface{}{
					"condition": "document.querySelector('.loading-spinner').style.display === 'none'",
					"description": "Wait for loading animation to complete",
					"timeout": 10,
					"interval": 100,
					"return_value": true,
				},
				Expected: "Waits for loading spinner to disappear, returns final condition value",
			},
		},
		
		"assert_element": {
			{
				Name: "Login Success Validation",
				Description: "Assert successful login with multiple checks",
				Parameters: map[string]interface{}{
					"selector": ".welcome-message",
					"assertion": "contains_text",
					"expected_value": "Welcome back",
					"timeout": 5,
					"case_sensitive": false,
				},
				Expected: "Passes if welcome message contains expected text",
			},
			{
				Name: "Form Field Validation",
				Description: "Assert form field has correct value and attributes",
				Parameters: map[string]interface{}{
					"selector": "#email",
					"assertion": "attribute_equals",
					"attribute_name": "value",
					"expected_value": "test@example.com",
					"timeout": 2,
				},
				Expected: "Passes if email field contains the expected value",
			},
			{
				Name: "Element Visibility Test",
				Description: "Verify element is visible and properly styled",
				Parameters: map[string]interface{}{
					"selector": ".success-alert",
					"assertion": "visible",
					"timeout": 3,
				},
				Expected: "Passes if success alert is visible on screen",
			},
		},
		
		"extract_table": {
			{
				Name: "Product Catalog Extraction",
				Description: "Extract complete product table to structured JSON",
				Parameters: map[string]interface{}{
					"selector": "#products-table",
					"output_format": "objects",
					"include_headers": true,
					"skip_empty_rows": true,
				},
				Expected: "Returns array of product objects with all table data",
			},
			{
				Name: "Financial Data CSV Export",
				Description: "Extract pricing table and convert to CSV format",
				Parameters: map[string]interface{}{
					"selector": ".pricing-table tbody",
					"output_format": "csv",
					"column_filter": []interface{}{"Product", "Price", "Features"},
					"max_rows": 50,
				},
				Expected: "Returns CSV string with filtered columns for analysis",
			},
			{
				Name: "Raw Data Array Extraction",
				Description: "Extract table as raw arrays for processing",
				Parameters: map[string]interface{}{
					"selector": "table.data-grid",
					"output_format": "array",
					"include_headers": false,
					"header_row": 1,
				},
				Expected: "Returns array of arrays with cell values for custom processing",
			},
		},
		
		"keyboard_shortcuts": {
			{
				Name: "Form Navigation",
				Description: "Navigate through form fields using Tab key",
				Parameters: map[string]interface{}{
					"keys": "Tab",
					"selector": "#contact-form",
				},
				Expected: "Moves focus to next form field within the contact form",
			},
			{
				Name: "Copy and Paste Text",
				Description: "Select all text and copy it to clipboard",
				Parameters: map[string]interface{}{
					"keys": "Ctrl+A",
					"selector": "textarea#message",
				},
				Expected: "Selects all text in the message textarea",
			},
			{
				Name: "Browser Refresh",
				Description: "Refresh the current page using F5 key",
				Parameters: map[string]interface{}{
					"keys": "F5",
				},
				Expected: "Refreshes the current page",
			},
			{
				Name: "Form Submission",
				Description: "Submit form using Enter key",
				Parameters: map[string]interface{}{
					"keys": "Enter",
					"selector": "#submit-button",
				},
				Expected: "Submits the form by pressing Enter on submit button",
			},
		},
		
		"switch_tab": {
			{
				Name: "Create New Tab",
				Description: "Open a new tab and navigate to a specific URL",
				Parameters: map[string]interface{}{
					"action": "create",
					"url": "https://example.com",
				},
				Expected: "Creates new tab, navigates to example.com, and switches to it",
			},
			{
				Name: "Switch to Next Tab",
				Description: "Switch to the next tab in sequence",
				Parameters: map[string]interface{}{
					"action": "switch",
					"switch_to": "next",
				},
				Expected: "Switches focus to the next available browser tab",
			},
			{
				Name: "List All Open Tabs",
				Description: "Get information about all currently open tabs",
				Parameters: map[string]interface{}{
					"action": "list",
				},
				Expected: "Returns list of all tabs with titles, URLs, and page IDs",
			},
			{
				Name: "Close Current Tab",
				Description: "Close the currently active tab",
				Parameters: map[string]interface{}{
					"action": "close",
					"target": "current",
				},
				Expected: "Closes current tab and switches to another available tab",
			},
			{
				Name: "Close All Tabs Except Current",
				Description: "Close all tabs while keeping the current tab open",
				Parameters: map[string]interface{}{
					"action": "close_all",
				},
				Expected: "Closes all tabs except the currently active one",
			},
		},
	}
	
	if exs, exists := examples[toolName]; exists {
		return exs
	}
	return []ToolExample{}
}