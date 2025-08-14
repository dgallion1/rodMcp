package webtools

import "fmt"

// ToolCategory represents different categories of tools
type ToolCategory string

const (
	BrowserAutomation ToolCategory = "browser_automation"
	UIControl        ToolCategory = "ui_control"  
	FileSystem       ToolCategory = "file_system"
	Network          ToolCategory = "network"
	FormAutomation   ToolCategory = "form_automation"
	AdvancedWaiting  ToolCategory = "advanced_waiting"
	Testing          ToolCategory = "testing"
)

// UsageHint provides contextual information about tool usage
type UsageHint struct {
	Tool        string
	Category    ToolCategory
	Description string
	Example     string
	CommonUse   []string
	WorksWith   []string
}

// HelpSystem provides enhanced tool discoverability
type HelpSystem struct {
	hints map[string]UsageHint
}

func NewHelpSystem() *HelpSystem {
	h := &HelpSystem{
		hints: make(map[string]UsageHint),
	}
	h.initializeHints()
	return h
}

func (h *HelpSystem) initializeHints() {
	h.hints["create_page"] = UsageHint{
		Tool:        "create_page",
		Category:    BrowserAutomation,
		Description: "Generate complete HTML pages with embedded CSS and JavaScript. Ideal for rapid prototyping, landing pages, interactive demos, and testing UI components.",
		Example:     "Create a responsive coffee shop landing page with contact form, image gallery, and smooth scrolling navigation",
		CommonUse: []string{
			"Build landing pages with forms and animations",
			"Create interactive dashboards with real-time data",
			"Prototype responsive designs with CSS Grid/Flexbox",
			"Generate test pages for automated testing",
		},
		WorksWith: []string{"navigate_page", "take_screenshot", "live_preview", "execute_script"},
	}
	
	h.hints["navigate_page"] = UsageHint{
		Tool:        "navigate_page",
		Category:    BrowserAutomation,
		Description: "Open URLs or local files in the browser. Essential for testing websites, loading content for interaction, and switching between pages.",
		Example:     "Navigate to localhost:3000 to test your React application, then take screenshots of different pages",
		CommonUse: []string{
			"Load local HTML files for testing",
			"Navigate to development servers (localhost:3000, :8080)",
			"Open live websites for analysis and interaction", 
			"Switch between different pages in your application",
		},
		WorksWith: []string{"click_element", "type_text", "take_screenshot", "execute_script"},
	}

	h.hints["execute_script"] = UsageHint{
		Tool:        "execute_script",
		Category:    BrowserAutomation,
		Description: "Run JavaScript code directly in browser pages. Powerful for DOM manipulation, form validation, API testing, and custom interactions.",
		Example:     "Execute JavaScript to validate all forms on the page, simulate user interactions, and test API endpoints",
		CommonUse: []string{
			"Test form validation and user interactions",
			"Manipulate DOM elements dynamically",
			"Make API calls and handle responses",
			"Extract data from pages programmatically",
			"Simulate complex user workflows",
		},
		WorksWith: []string{"navigate_page", "click_element", "get_element_text", "wait_for_element"},
	}

	h.hints["click_element"] = UsageHint{
		Tool:        "click_element", 
		Category:    UIControl,
		Description: "Click buttons, links, and interactive elements using CSS selectors. Essential for automated testing and user interaction simulation.",
		Example:     "Click the 'Submit' button after filling out a contact form, then wait for success message",
		CommonUse: []string{
			"Submit forms and trigger form validation",
			"Navigate through multi-step workflows",
			"Test button interactions and state changes",
			"Trigger dropdown menus and modal dialogs",
		},
		WorksWith: []string{"type_text", "wait_for_element", "get_element_text", "take_screenshot"},
	}

	h.hints["live_preview"] = UsageHint{
		Tool:        "live_preview",
		Category:    BrowserAutomation,
		Description: "Start a local HTTP server for live development with auto-reload. Perfect for testing static sites and serving multiple HTML files.",
		Example:     "Start preview server at localhost:8080, then navigate to test multiple pages in your website",
		CommonUse: []string{
			"Serve static HTML/CSS/JS files for testing",
			"Create multi-page website demonstrations",
			"Test responsive designs across different pages",
			"Share local development with others",
		},
		WorksWith: []string{"navigate_page", "create_page", "take_screenshot", "http_request"},
	}

	h.hints["form_fill"] = UsageHint{
		Tool:        "form_fill",
		Category:    FormAutomation,
		Description: "Complete form automation solution that handles all input types (text, select, radio, checkbox, textarea) with validation and optional submission. The ultimate tool for login forms, contact forms, checkout processes, and registration workflows.",
		Example:     "Fill out entire contact form with name, email, message, country selection, and newsletter subscription, then submit with validation",
		CommonUse: []string{
			"Automate login and registration workflows",
			"Fill complex checkout forms with multiple fields",
			"Test form validation and error handling",
			"Submit contact forms and feedback forms",
			"Handle multi-step form wizards efficiently",
		},
		WorksWith: []string{"navigate_page", "wait_for_condition", "assert_element", "take_screenshot"},
	}

	h.hints["wait_for_condition"] = UsageHint{
		Tool:        "wait_for_condition",
		Category:    AdvancedWaiting,
		Description: "Wait for custom JavaScript conditions to become true. Much more powerful than waiting for elements - perfect for animations, API responses, state changes, and complex application logic. Essential for modern single-page applications.",
		Example:     "Wait for React component state to update, API data to load, or animation to complete before proceeding with automation",
		CommonUse: []string{
			"Wait for API responses and data loading",
			"Handle animation and transition completion",
			"Wait for React/Vue component state changes", 
			"Monitor application loading states",
			"Wait for dynamic content and lazy loading",
		},
		WorksWith: []string{"execute_script", "assert_element", "form_fill", "screen_scrape"},
	}

	h.hints["assert_element"] = UsageHint{
		Tool:        "assert_element",
		Category:    Testing,
		Description: "Comprehensive element testing framework with 15+ assertion types. Essential for automated testing, validation workflows, and ensuring UI correctness. Provides detailed pass/fail reporting with diagnostics.",
		Example:     "Assert that login button is visible and enabled, success message contains correct text, and form fields have expected values",
		CommonUse: []string{
			"Verify element existence and visibility states",
			"Test form field values and states",
			"Validate text content and attributes",
			"Check CSS classes and styling",
			"Count elements and verify quantities",
		},
		WorksWith: []string{"form_fill", "wait_for_condition", "click_element", "navigate_page"},
	}

	h.hints["extract_table"] = UsageHint{
		Tool:        "extract_table",
		Category:    BrowserAutomation,
		Description: "Extract structured data from HTML tables with support for multiple output formats (JSON objects, arrays, CSV), column filtering, and header management. Perfect for data scraping and analysis workflows.",
		Example:     "Extract product catalog table to JSON objects, filter specific columns, and export pricing data to CSV for analysis",
		CommonUse: []string{
			"Convert HTML tables to structured JSON data",
			"Export table data to CSV for spreadsheet analysis",
			"Extract specific columns from large data tables",
			"Scrape pricing and product information",
			"Generate reports from web-based dashboards",
		},
		WorksWith: []string{"navigate_page", "wait_for_condition", "screen_scrape", "http_request"},
	}

	h.hints["take_element_screenshot"] = UsageHint{
		Tool:        "take_element_screenshot",
		Category:    BrowserAutomation,
		Description: "Capture screenshots of specific elements with smart positioning, padding control, and visibility waiting. Perfect for focused testing, bug reporting, and component documentation.",
		Example:     "Screenshot error messages for bug reports, capture navigation menus for testing, or document form field states for validation workflows",
		CommonUse: []string{
			"Document UI components and element states",
			"Capture error messages and validation states",
			"Test element positioning and styling",
			"Generate visual evidence for bug reports",
			"Create focused screenshots for documentation",
		},
		WorksWith: []string{"navigate_page", "wait_for_element", "assert_element", "click_element"},
	}

	h.hints["keyboard_shortcuts"] = UsageHint{
		Tool:        "keyboard_shortcuts",
		Category:    UIControl,
		Description: "Send keyboard combinations and special keys (Ctrl+C/V, F5, Tab, Enter, arrow keys, function keys). Essential for form navigation, shortcuts, copy/paste operations, and testing keyboard interactions.",
		Example:     "Navigate form fields with Tab, trigger browser refresh with F5, copy/paste text with Ctrl+C/V, or test keyboard shortcuts in web applications",
		CommonUse: []string{
			"Navigate forms with Tab/Shift+Tab key combinations",
			"Copy/paste operations with Ctrl+C, Ctrl+V, Ctrl+A",
			"Trigger browser functions with F5 refresh, F12 DevTools",
			"Submit forms using Enter key",
			"Test application keyboard shortcuts and hotkeys",
			"Navigate menus and dropdowns with arrow keys",
		},
		WorksWith: []string{"type_text", "click_element", "wait_for_element", "form_fill"},
	}

	h.hints["switch_tab"] = UsageHint{
		Tool:        "switch_tab",
		Category:    UIControl,
		Description: "Switch between browser tabs for multi-tab workflow automation. Create new tabs, switch between existing tabs, close tabs, and manage multi-tab workflows efficiently.",
		Example:     "Open multiple sites in different tabs, switch between them for comparison, or manage complex workflows across multiple pages simultaneously",
		CommonUse: []string{
			"Create new tabs and navigate to different URLs",
			"Switch between tabs using directional navigation (next, previous, first, last)",
			"Close specific tabs or all tabs except current",
			"List all open tabs with titles and URLs",
			"Manage multi-tab testing workflows and comparisons",
			"Automate workflows requiring multiple open pages",
		},
		WorksWith: []string{"navigate_page", "create_page", "take_screenshot", "screen_scrape"},
	}

	// Add more hints for other tools...
}

func (h *HelpSystem) GetHint(toolName string) (UsageHint, bool) {
	hint, exists := h.hints[toolName]
	return hint, exists
}

func (h *HelpSystem) GetToolsByCategory(category ToolCategory) []UsageHint {
	var tools []UsageHint
	for _, hint := range h.hints {
		if hint.Category == category {
			tools = append(tools, hint)
		}
	}
	return tools
}

func (h *HelpSystem) GetWorkflowSuggestion(goals []string) string {
	// Enhanced workflow suggestions with new power tools
	suggestions := []string{
		"🌐 **Web Development Workflow:**",
		"1. Use `create_page` to build your HTML with CSS and JavaScript",
		"2. Start `live_preview` server to serve your files locally", 
		"3. Use `navigate_page` to open your page in the browser",
		"4. Use `wait_for_condition` to ensure page is fully loaded",
		"5. Take `take_screenshot` to document your progress",
		"6. Use `execute_script` to test JavaScript functionality",
		"",
		"📝 **Form Automation Workflow (🔥 NEW):**",
		"1. Use `navigate_page` to load your form page",
		"2. Use `form_fill` to complete entire form with structured data",
		"3. Use `assert_element` to verify form submission success",
		"4. Use `take_screenshot` to document the result",
		"",
		"🧪 **Advanced Testing Workflow (🔥 ENHANCED):**",
		"1. Use `navigate_page` to load the page you want to test",
		"2. Use `wait_for_condition` to wait for dynamic content/APIs",
		"3. Use `form_fill` for complex form interactions",
		"4. Use `assert_element` with comprehensive validation",
		"5. Use `take_screenshot` to capture test evidence",
		"",
		"⚡ **Modern SPA Testing (🔥 NEW):**",
		"1. Use `navigate_page` to load your React/Vue application",
		"2. Use `wait_for_condition` to wait for component state changes",
		"3. Use `form_fill` for user registration/login flows",
		"4. Use `assert_element` to verify UI state and content",
		"5. Use `take_element_screenshot` to document component states",
		"6. Use `screen_scrape` to extract and validate data",
		"",
		"📊 **Table Data Extraction Workflow (🔥 NEW):**",
		"1. Use `navigate_page` to load page with data tables",
		"2. Use `wait_for_condition` to ensure table data is loaded",
		"3. Use `extract_table` to convert HTML tables to structured data",
		"4. Use `assert_element` to verify extraction success",
		"5. Use `write_file` to save extracted data for analysis",
		"",
		"📸 **Visual Testing & Bug Reporting (🔥 NEW):**",
		"1. Use `navigate_page` to load the problematic page",
		"2. Use `click_element` or `form_fill` to reproduce the issue",
		"3. Use `take_element_screenshot` to capture error states",
		"4. Use `assert_element` to verify expected vs actual behavior",
		"5. Use `take_screenshot` for full page context documentation",
		"",
		"🚀 **API Testing Workflow:**", 
		"1. Use `http_request` to test your API endpoints",
		"2. Use `create_page` to build a test interface",
		"3. Use `execute_script` to make API calls from the browser",
		"4. Use `assert_element` to verify response data display",
	}
	
	return fmt.Sprintf("%s", joinStrings(suggestions, "\n"))
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}