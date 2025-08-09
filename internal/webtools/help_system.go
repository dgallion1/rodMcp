package webtools

import "fmt"

// ToolCategory represents different categories of tools
type ToolCategory string

const (
	BrowserAutomation ToolCategory = "browser_automation"
	UIControl        ToolCategory = "ui_control"  
	FileSystem       ToolCategory = "file_system"
	Network          ToolCategory = "network"
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
	// Simple workflow suggestion based on common patterns
	suggestions := []string{
		"🌐 **Web Development Workflow:**",
		"1. Use `create_page` to build your HTML with CSS and JavaScript",
		"2. Start `live_preview` server to serve your files locally", 
		"3. Use `navigate_page` to open your page in the browser",
		"4. Take `take_screenshot` to document your progress",
		"5. Use `execute_script` to test JavaScript functionality",
		"",
		"🧪 **Testing Workflow:**",
		"1. Use `navigate_page` to load the page you want to test",
		"2. Use `click_element` and `type_text` to simulate user actions",
		"3. Use `wait_for_element` to handle dynamic content loading",
		"4. Use `get_element_text` to verify expected results",
		"5. Use `take_screenshot` to capture test results",
		"",
		"📊 **API Testing Workflow:**", 
		"1. Use `http_request` to test your API endpoints",
		"2. Use `create_page` to build a test interface",
		"3. Use `execute_script` to make API calls from the browser",
		"4. Use `get_element_text` to verify response data display",
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