package webtools

import (
	"fmt"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
	"strings"
	"time"
)

// HelpTool provides interactive help and usage guidance
type HelpTool struct {
	logger     *logger.Logger
	helpSystem *HelpSystem
}

func NewHelpTool(log *logger.Logger) *HelpTool {
	return &HelpTool{
		logger:     log,
		helpSystem: NewHelpSystem(),
	}
}

func (t *HelpTool) Name() string {
	return "help"
}

func (t *HelpTool) Description() string {
	return "Get interactive help, usage examples, and workflow suggestions for rodmcp tools. Discover what tools can do and how to use them effectively."
}

func (t *HelpTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"topic": map[string]interface{}{
				"type":        "string", 
				"description": "Help topic: 'overview', 'workflows', 'examples', or specific tool name (e.g., 'create_page')",
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Tool category: 'browser_automation', 'ui_control', 'file_system', 'network'",
			},
		},
	}
}

func (t *HelpTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	topic, _ := args["topic"].(string)
	category, _ := args["category"].(string)

	var helpContent string

	switch topic {
	case "overview", "":
		helpContent = t.getOverview()
	case "workflows":
		helpContent = t.getWorkflows()
	case "examples":
		helpContent = t.getExamples()
	default:
		// Check if it's a specific tool
		if hint, exists := t.helpSystem.GetHint(topic); exists {
			helpContent = t.getToolHelp(hint)
		} else {
			helpContent = t.getUnknownTopic(topic)
		}
	}

	if category != "" {
		helpContent += "\n\n" + t.getCategoryHelp(ToolCategory(category))
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: helpContent,
			Data: map[string]interface{}{
				"topic":    topic,
				"category": category,
			},
		}},
	}, nil
}

func (t *HelpTool) getOverview() string {
	return `# 🛠️ RodMCP Tools Overview

RodMCP provides 22 comprehensive web development tools organized into 7 categories:

## 🌐 Browser Automation (6 tools)
• **create_page** - Generate HTML pages with CSS/JavaScript  
• **navigate_page** - Open URLs and local files
• **execute_script** - Run JavaScript in browser pages
• **take_screenshot** - Capture visual snapshots
• **live_preview** - Start local development server
• **set_browser_visibility** - Switch visible/headless modes

## 🎯 UI Control (8 tools)  
• **click_element** - Click buttons and links
• **type_text** - Fill forms and input fields
• **wait** / **wait_for_element** - Handle timing and loading
• **get_element_text** / **get_element_attribute** - Extract page data
• **scroll** - Navigate long pages
• **hover_element** - Trigger hover effects

## 🕷️ Screen Scraping (1 tool)
• **screen_scrape** - Extract structured data from web pages

## 📝 Form Automation (1 tool)
• **form_fill** - Complete form automation with validation and submission

## ⚡ Advanced Waiting (1 tool)
• **wait_for_condition** - Wait for custom JavaScript conditions (animations, APIs, state changes)

## 🧪 Testing & Assertions (1 tool)
• **assert_element** - Comprehensive element testing (15+ assertion types)

## 📁 File System (3 tools)
• **read_file** / **write_file** - File operations
• **list_directory** - Browse project structure

## 🌍 Network (1 tool)
• **http_request** - Test APIs and web services

## 💡 Quick Start Tips:
1. Use **help** with tool name for detailed examples: help form_fill
2. Use **help workflows** for common usage patterns
3. Use **help examples** for ready-to-use code snippets

🔥 **New Power Tools**: form_fill, wait_for_condition, and assert_element provide professional-grade automation and testing capabilities!

Try: help workflows to see common development patterns!`
}

func (t *HelpTool) getWorkflows() string {
	return t.helpSystem.GetWorkflowSuggestion([]string{})
}

func (t *HelpTool) getExamples() string {
	return `# 📚 Usage Examples

## 🎨 Create Interactive Landing Page
` + "```" + `
create_page:
  filename: "coffee-shop"  
  title: "Mountain View Coffee"
  html: "<header><h1>Premium Coffee</h1></header><main>...</main>"
  css: "body{font-family:Arial} header{background:#8B4513;color:white}"
` + "```" + `

## 📝 Complete Form Automation
` + "```" + `
form_fill:
  fields:
    "#email": "user@example.com"
    "#password": "securepass123"  
    "select[name='country']": "US"
    "input[name='newsletter']": true
  submit: true
  validate_required: true
` + "```" + `

## ⚡ Smart Waiting for Dynamic Content
` + "```" + `
wait_for_condition:
  condition: "document.querySelectorAll('.product-card').length >= 10"
  description: "Wait for product grid to load completely"
  timeout: 15
  interval: 500
` + "```" + `

## 🧪 Comprehensive Element Testing
` + "```" + `
assert_element:
  selector: ".success-alert"
  assertion: "contains_text"
  expected_value: "Form submitted successfully"
  timeout: 5
  case_sensitive: false
` + "```" + `

## 🧪 Advanced Test Form Workflow  
` + "```" + `
1. navigate_page: "contact-form.html"
2. form_fill: Fill entire form with structured data
3. wait_for_condition: "!!window.formValidated"
4. assert_element: selector=".success-message", assertion="visible"
5. take_screenshot: filename="form-test.png"
` + "```" + `

## 🚀 API Testing Workflow
` + "```" + `
1. http_request: url="https://api.example.com/users", method="GET"
2. create_page: Build test interface showing API data
3. execute_script: Make API calls and display results
4. assert_element: Check API response display
5. take_screenshot: Document API response
` + "```" + `

## 📊 Development Server Setup
` + "```" + `
1. create_page: Build your HTML pages
2. live_preview: port=8080, directory="."
3. navigate_page: "localhost:8080"
4. wait_for_condition: "document.readyState === 'complete'"
5. Test and iterate with take_screenshot
` + "```" + `

🔥 **Pro Tip**: Combine form_fill + wait_for_condition + assert_element for robust automation workflows!

Use help [tool_name] for detailed tool-specific examples!`
}

func (t *HelpTool) getToolHelp(hint UsageHint) string {
	var content strings.Builder
	
	content.WriteString(fmt.Sprintf("# 🔧 %s Help\n\n", hint.Tool))
	content.WriteString(fmt.Sprintf("**Category:** %s\n\n", hint.Category))
	content.WriteString(fmt.Sprintf("**Description:**\n%s\n\n", hint.Description))
	
	content.WriteString(fmt.Sprintf("**Example Use Case:**\n%s\n\n", hint.Example))
	
	if len(hint.CommonUse) > 0 {
		content.WriteString("**Common Uses:**\n")
		for _, use := range hint.CommonUse {
			content.WriteString(fmt.Sprintf("• %s\n", use))
		}
		content.WriteString("\n")
	}
	
	if len(hint.WorksWith) > 0 {
		content.WriteString("**Works Well With:**\n")
		for _, tool := range hint.WorksWith {
			content.WriteString(fmt.Sprintf("• %s\n", tool))
		}
		content.WriteString("\n")
	}

	// Add specific examples from enhanced tool system
	examples := GetToolExamples(hint.Tool)
	if len(examples) > 0 {
		content.WriteString("**Concrete Examples:**\n")
		for _, ex := range examples {
			content.WriteString(fmt.Sprintf("**%s:** %s\n", ex.Name, ex.Description))
			content.WriteString("```json\n")
			for key, value := range ex.Parameters {
				content.WriteString(fmt.Sprintf("  \"%s\": %v\n", key, value))
			}
			content.WriteString("```\n")
			content.WriteString(fmt.Sprintf("*Expected: %s*\n\n", ex.Expected))
		}
	}
	
	return content.String()
}

func (t *HelpTool) getCategoryHelp(category ToolCategory) string {
	tools := t.helpSystem.GetToolsByCategory(category)
	if len(tools) == 0 {
		return fmt.Sprintf("No tools found in category: %s", category)
	}
	
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# 📂 %s Tools\n\n", category))
	
	for _, tool := range tools {
		content.WriteString(fmt.Sprintf("## %s\n%s\n\n", tool.Tool, tool.Description))
	}
	
	return content.String()
}

func (t *HelpTool) getUnknownTopic(topic string) string {
	return fmt.Sprintf(`# ❓ Unknown Topic: "%s"

Available help topics:
• **overview** - General tool overview and categories
• **workflows** - Common development workflows  
• **examples** - Ready-to-use code examples
• **[tool_name]** - Specific tool help (e.g., "create_page", "execute_script")

Available categories:
• **browser_automation** - Page creation, navigation, screenshots
• **ui_control** - Clicking, typing, waiting, data extraction  
• **file_system** - File and directory operations
• **network** - HTTP requests and API testing

Try: help overview to get started!`, topic)
}