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
	Tool         string
	Category     ToolCategory
	Description  string
	Example      string
	CommonUse    []string
	WorksWith    []string
	Complexity   string   // "basic", "intermediate", "advanced"
	Prerequisites []string // Tools that should be learned first
	LearningTips []string // Tips for LLM usage
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
		Complexity: "basic",
		Prerequisites: []string{},
		LearningTips: []string{
			"Start with simple HTML structure, then add CSS styling",
			"Use semantic HTML5 elements for better accessibility",
			"Test responsive design with different viewport sizes",
			"Combine with live_preview for instant feedback",
		},
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
		Complexity: "basic",
		Prerequisites: []string{},
		LearningTips: []string{
			"Use file:// protocol for local HTML files",
			"Always wait for page to load before interacting with elements",
			"Use localhost:PORT for development servers",
			"Check if page loads successfully before proceeding",
		},
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
		Complexity: "basic",
		Prerequisites: []string{"navigate_page"},
		LearningTips: []string{
			"Use specific selectors like #id or unique classes for reliability",
			"Wait for elements to be visible before clicking",
			"Use browser dev tools to test selectors first", 
			"Consider using wait_for_element before clicking dynamic elements",
			"If you get 'No pages available' error, use create_page or navigate_page first",
		},
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
		Complexity: "intermediate",
		Prerequisites: []string{"navigate_page", "click_element"},
		LearningTips: []string{
			"Use structured data with field selectors as keys",
			"Test individual form fields before automating the full form",
			"Enable validation checking to catch form errors",
			"Use wait_for_condition after submission to verify success",
		},
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
		Complexity: "advanced",
		Prerequisites: []string{"navigate_page", "execute_script"},
		LearningTips: []string{
			"Write JavaScript conditions that return true/false",
			"Use browser dev tools to test conditions first",
			"Start with simple conditions before complex state checking",
			"Use descriptive condition descriptions for debugging",
		},
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
		Complexity: "intermediate",
		Prerequisites: []string{"navigate_page", "click_element"},
		LearningTips: []string{
			"Start with basic assertions like 'exists' and 'visible'",
			"Use 'contains_text' for partial text matching",
			"Chain assertions in a logical test sequence",
			"Use case-insensitive matching when text case varies",
		},
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
		Complexity: "intermediate",
		Prerequisites: []string{"navigate_page"},
		LearningTips: []string{
			"Start with 'list' action to see all available tabs",
			"Use 'create' action with URL to open new tabs",
			"Use directional navigation for systematic tab switching",
			"Close tabs when done to keep workspace organized",
		},
	}
	
	// File system tools with timeout and size limit information
	h.hints["read_file"] = UsageHint{
		Tool:        "read_file",
		Category:    FileSystem,
		Description: "Read file contents with security controls and size limits. Features 30-second timeout protection and configurable size limits (10MB default).",
		Example:     "Read configuration files, source code, or data files for analysis and processing",
		CommonUse: []string{
			"Read configuration files and settings",
			"Analyze source code and documentation",
			"Process data files and logs",
			"Load content for web page generation",
		},
		WorksWith: []string{"write_file", "create_page", "http_request"},
		Complexity: "basic",
		Prerequisites: []string{},
		LearningTips: []string{
			"Respects configured file access security settings",
			"Files larger than limit (default 10MB) will be rejected with clear message",
			"30-second timeout prevents hanging on slow storage",
			"Check file path permissions before reading",
		},
	}
	
	h.hints["write_file"] = UsageHint{
		Tool:        "write_file",
		Category:    FileSystem,
		Description: "Write content to files with size validation and timeout protection. Includes directory creation and configurable size limits for safety.",
		Example:     "Save generated HTML, configuration files, or processed data with automatic directory creation",
		CommonUse: []string{
			"Save generated HTML, CSS, and JavaScript files",
			"Write configuration and settings files",
			"Export processed data and reports",
			"Create documentation and README files",
		},
		WorksWith: []string{"read_file", "create_page", "list_directory"},
		Complexity: "basic",
		Prerequisites: []string{},
		LearningTips: []string{
			"Content size is checked before writing (default 10MB limit)",
			"Automatically creates parent directories if needed",
			"30-second timeout prevents blocking on slow storage",
			"Respects file access security configuration",
		},
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

// GetLLMGuidance provides specific guidance for LLM users
func (h *HelpSystem) GetLLMGuidance() string {
	guidance := []string{
		"# 🤖 LLM-Optimized Usage Patterns",
		"",
		"## 🎯 **Basic Tool Progression**",
		"**Start Here (🟢 Basic):**",
		"1. **create_page** → Build HTML pages for testing",
		"2. **navigate_page** → Load pages in browser", 
		"3. **take_screenshot** → Visual confirmation",
		"4. **click_element** → Basic interactions",
		"5. **type_text** → Form field input",
		"",
		"## 🔧 **Intermediate Workflows (🟡 Intermediate)**", 
		"**Form Automation:**",
		"• **form_fill** → Complete entire forms efficiently",
		"• **assert_element** → Verify form submission success",
		"• **switch_tab** → Multi-tab form comparisons",
		"",
		"**Testing & Validation:**",
		"• **assert_element** → Comprehensive UI testing",
		"• **take_element_screenshot** → Document test results",
		"• **extract_table** → Data validation workflows",
		"",
		"## ⚡ **Advanced Automation (🔴 Advanced)**",
		"**Dynamic Content Handling:**",
		"• **wait_for_condition** → Handle SPAs and AJAX",
		"• **screen_scrape** → Complex data extraction",
		"• **execute_script** → Custom JavaScript logic",
		"",
		"## 💡 **LLM Best Practices**",
		"",
		"### ✅ **Error Prevention Patterns**",
		"```",
		"❌ Avoid: Clicking elements immediately after navigation",
		"✅ Better: navigate_page → wait_for_element → click_element",
		"",
		"❌ Avoid: Complex selectors without testing",
		"✅ Better: Start with simple selectors (#id, .class)",
		"",
		"❌ Avoid: Hardcoded timeouts",
		"✅ Better: Use wait_for_condition for dynamic content",
		"```",
		"",
		"### ⏰ **Timeout & Non-Blocking Design**",
		"**🎉 Good News: RodMCP won't block your conversation!**",
		"",
		"• **All operations have timeouts** (30s browser ops, 30s file I/O, 10s command execution)",
		"• **Helpful error messages** guide you to the correct next steps",
		"• **No infinite waiting** - tools fail fast with clear explanations",
		"• **Memory protection** - file operations respect size limits (10MB default)",
		"",
		"**Example Error Flow:**",
		"```",
		"You: Use click_element to click #button",
		"❌ Error: No browser pages currently open. To use click_element:",
		"   1. Create a page: use create_page to make a new HTML page, or",
		"   2. Navigate to a URL: use navigate_page to load an existing website",
		"   Then you can interact with elements on the page.",
		"```",
		"",
		"### 🎯 **Selector Strategy**",
		"**Reliability Priority:**",
		"1. **#id** (most reliable) - unique identifiers",
		"2. **[name='field']** (forms) - stable form field names", 
		"3. **.unique-class** (styling) - specific CSS classes",
		"4. **tag[attribute]** (semantic) - HTML5 semantic elements",
		"5. **//text()** (XPath) - when content is stable",
		"",
		"### 🔄 **Progressive Complexity**",
		"**Start Simple, Build Up:**",
		"```",
		"Level 1: navigate_page + take_screenshot (validation)",
		"Level 2: + click_element + type_text (basic interaction)", 
		"Level 3: + wait_for_element + assert_element (robust testing)",
		"Level 4: + form_fill + wait_for_condition (complex workflows)",
		"Level 5: + screen_scrape + execute_script (advanced automation)",
		"```",
		"",
		"### 🎬 **Debugging Workflow**",
		"**When Things Go Wrong:**",
		"1. **take_screenshot** → See current page state",
		"2. **execute_script: 'document.querySelector(\"selector\")'** → Test selector",
		"3. **wait_for_element** → Ensure element exists", 
		"4. **assert_element: 'exists'** → Verify element presence",
		"5. **take_element_screenshot** → Focus on problematic element",
		"",
		"### ⚠️ **Common LLM Pitfalls**",
		"",
		"**❌ Don't Do This:**",
		"• Clicking invisible or disabled elements",
		"• Using fragile selectors (nth-child without context)",
		"• Skipping page load verification",
		"• Hardcoding delays instead of waiting for conditions",
		"• Ignoring error messages from tools",
		"",
		"**✅ Do This Instead:**",
		"• Verify elements are visible/enabled before interaction",
		"• Use semantic selectors with business meaning",
		"• Always verify page navigation succeeded",
		"• Use wait_for_condition for dynamic states",
		"• Read error messages - they contain helpful examples",
		"",
		"## 🚀 **Pro Tips for LLMs**",
		"",
		"1. **Parameter Examples**: Every parameter description includes examples - use them!",
		"2. **Error Context**: Error messages provide specific examples and suggestions",
		"3. **Tool Complexity**: 🟢 Basic → 🟡 Intermediate → 🔴 Advanced progression",
		"4. **Prerequisites**: Check tool prerequisites before attempting complex workflows",
		"5. **Learning Tips**: Each tool has specific LLM learning tips",
		"6. **Non-Blocking Design**: Tools timeout gracefully - no conversation freezing!",
		"7. **Progressive Workflows**: Start with create_page/navigate_page before interaction tools",
		"",
		"### 🛡️ **Reliability Features**",
		"",
		"• **Automatic Timeouts**: No tool will hang your conversation indefinitely",
		"• **Memory Limits**: File operations protect against excessive memory usage", 
		"• **Clear Error Messages**: Each error explains exactly what to do next",
		"• **Size Validation**: File operations check limits before processing",
		"• **Graceful Degradation**: Tools fail fast with helpful suggestions",
		"",
		"**Use `help [tool_name]` for detailed guidance on any tool!**",
	}
	
	return fmt.Sprintf("%s", joinStrings(guidance, "\n"))
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