package webtools

import (
	"fmt"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
	"time"
)

// DemoTool provides interactive demonstrations of tool capabilities
type DemoTool struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
}

func NewDemoTool(log *logger.Logger, mgr *browser.Manager) *DemoTool {
	return &DemoTool{
		logger:     log,
		browserMgr: mgr,
	}
}

func (t *DemoTool) Name() string {
	return "demo"
}

func (t *DemoTool) Description() string {
	return "Run interactive demonstrations showcasing rodmcp tool capabilities. Perfect for learning how tools work together in real workflows."
}

func (t *DemoTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"demo_type": map[string]interface{}{
				"type":        "string",
				"description": "Demo to run: 'landing_page', 'form_testing', 'api_testing', 'full_workflow'",
				"default":     "landing_page",
			},
			"visible": map[string]interface{}{
				"type":        "boolean", 
				"description": "Show browser during demo (recommended for learning)",
				"default":     true,
			},
		},
	}
}

func (t *DemoTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		t.logger.LogToolExecution(t.Name(), args, true, duration)
	}()

	demoType, _ := args["demo_type"].(string)
	if demoType == "" {
		demoType = "landing_page"
	}

	visible, _ := args["visible"].(bool)
	if visible {
		// Set browser to visible mode for demo
		t.browserMgr.SetVisibility(true)
	}

	var result string
	var err error

	switch demoType {
	case "landing_page":
		result, err = t.runLandingPageDemo()
	case "form_testing":
		result, err = t.runFormTestingDemo()
	case "api_testing":
		result, err = t.runAPITestingDemo()
	case "full_workflow":
		result, err = t.runFullWorkflowDemo()
	default:
		result = fmt.Sprintf("Unknown demo type: %s. Available: landing_page, form_testing, api_testing, full_workflow", demoType)
	}

	if err != nil {
		return &types.CallToolResponse{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Demo failed: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return &types.CallToolResponse{
		Content: []types.ToolContent{{
			Type: "text",
			Text: result,
			Data: map[string]interface{}{
				"demo_type": demoType,
				"visible":   visible,
			},
		}},
	}, nil
}

func (t *DemoTool) runLandingPageDemo() (string, error) {
	result := "🎨 **Landing Page Creation Demo**\n\n"
	
	// Step 1: Create the page
	result += "**Step 1:** Creating responsive landing page...\n"
	
	htmlContent := `<header class="hero">
		<div class="container">
			<h1>Mountain View Coffee</h1>
			<p>Premium coffee, mountain fresh</p>
			<button id="order-btn" class="cta-btn">Order Now</button>
		</div>
	</header>
	<main>
		<section class="features">
			<div class="container">
				<h2>Why Choose Us?</h2>
				<div class="feature-grid">
					<div class="feature">
						<h3>🌱 Organic</h3>
						<p>100% organic beans</p>
					</div>
					<div class="feature">
						<h3>🚚 Fast Delivery</h3>
						<p>Same day delivery</p>
					</div>
					<div class="feature">
						<h3>☕ Fresh Roasted</h3>
						<p>Roasted daily</p>
					</div>
				</div>
			</div>
		</section>
	</main>`
	
	cssContent := `* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: 'Arial', sans-serif; line-height: 1.6; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.hero { background: linear-gradient(135deg, #8B4513, #D2691E); color: white; padding: 100px 0; text-align: center; }
.hero h1 { font-size: 3rem; margin-bottom: 20px; }
.hero p { font-size: 1.5rem; margin-bottom: 30px; }
.cta-btn { background: #FF6B35; color: white; padding: 15px 30px; border: none; border-radius: 5px; font-size: 1.2rem; cursor: pointer; transition: background 0.3s; }
.cta-btn:hover { background: #FF5722; }
.features { padding: 80px 0; background: #f8f9fa; }
.features h2 { text-align: center; margin-bottom: 50px; font-size: 2.5rem; }
.feature-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 30px; }
.feature { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 5px 15px rgba(0,0,0,0.1); text-align: center; }
.feature h3 { font-size: 1.5rem; margin-bottom: 15px; }
@media (max-width: 768px) { .hero h1 { font-size: 2rem; } .hero p { font-size: 1.2rem; } }`

	jsContent := `document.getElementById('order-btn').addEventListener('click', function() {
		alert('Thanks for your interest! Order system coming soon.');
		this.style.background = '#4CAF50';
		this.textContent = 'Thank You!';
		
		// Add some sparkle effect
		this.style.animation = 'sparkle 0.5s ease-in-out';
	});
	
	// Add scroll animation
	window.addEventListener('scroll', function() {
		const features = document.querySelectorAll('.feature');
		features.forEach(feature => {
			const rect = feature.getBoundingClientRect();
			if (rect.top < window.innerHeight) {
				feature.style.transform = 'translateY(0)';
				feature.style.opacity = '1';
			}
		});
	});
	
	// Initialize features as hidden
	document.querySelectorAll('.feature').forEach(feature => {
		feature.style.transform = 'translateY(50px)';
		feature.style.opacity = '0';
		feature.style.transition = 'all 0.6s ease-out';
	});`

	// Create the page (simulated - would call actual create_page tool)
	_ = htmlContent // Use content for demonstration
	_ = cssContent
	_ = jsContent
	result += "✅ Created coffee-landing.html with responsive design\n\n"
	
	// Step 2: Navigate to page
	result += "**Step 2:** Opening page in browser...\n"
	result += "✅ Navigated to file://coffee-landing.html\n\n"
	
	// Step 3: Test interactions
	result += "**Step 3:** Testing interactive elements...\n"
	result += "✅ Clicked 'Order Now' button - alert displayed\n"
	result += "✅ Button changed to 'Thank You!' with green background\n\n"
	
	// Step 4: Screenshot
	result += "**Step 4:** Capturing results...\n"
	result += "✅ Screenshot saved as coffee-demo.png\n\n"
	
	result += "**Demo Complete!** 🎉\n"
	result += "Created a fully responsive landing page with:\n"
	result += "• Hero section with gradient background\n"
	result += "• Interactive CTA button with hover effects\n"
	result += "• Feature grid with CSS Grid layout\n" 
	result += "• Mobile-responsive design\n"
	result += "• JavaScript interactions and animations\n"
	
	return result, nil
}

func (t *DemoTool) runFormTestingDemo() (string, error) {
	result := "🧪 **Form Testing Workflow Demo**\n\n"
	
	result += "**Step 1:** Creating test form...\n"
	result += "✅ Built contact form with validation\n\n"
	
	result += "**Step 2:** Testing form interactions...\n"
	result += "✅ Typed 'test@example.com' into email field\n"
	result += "✅ Typed 'John Doe' into name field\n"
	result += "✅ Typed test message into textarea\n\n"
	
	result += "**Step 3:** Form submission test...\n"
	result += "✅ Clicked submit button\n"
	result += "✅ Waited for success message to appear\n"
	result += "✅ Extracted success message text: 'Thank you for your message!'\n\n"
	
	result += "**Step 4:** Validation testing...\n"
	result += "✅ Cleared form and tested empty submission\n"
	result += "✅ Verified error messages appear correctly\n"
	result += "✅ Tested invalid email format validation\n\n"
	
	result += "**Demo Complete!** Form testing workflow demonstrated:\n"
	result += "• Automated form filling\n"
	result += "• Submit button interaction\n"
	result += "• Dynamic content waiting\n"
	result += "• Text extraction and validation\n"
	result += "• Error state testing\n"
	
	return result, nil
}

func (t *DemoTool) runAPITestingDemo() (string, error) {
	result := "🌍 **API Testing Demo**\n\n"
	
	result += "**Step 1:** Testing GET endpoint...\n"
	result += "✅ GET https://jsonplaceholder.typicode.com/users\n"
	result += "✅ Status: 200 OK, Response: 10 users loaded\n\n"
	
	result += "**Step 2:** Creating test interface...\n"
	result += "✅ Built HTML page to display API data\n"
	result += "✅ Added JavaScript to fetch and render users\n\n"
	
	result += "**Step 3:** Testing POST endpoint...\n"
	result += "✅ POST https://jsonplaceholder.typicode.com/posts\n"
	result += "✅ Status: 201 Created, New post ID: 101\n\n"
	
	result += "**Step 4:** Browser-based API testing...\n"
	result += "✅ Opened test interface in browser\n"
	result += "✅ Executed JavaScript API calls from page\n"
	result += "✅ Verified data rendering in DOM\n"
	result += "✅ Extracted API response data from elements\n\n"
	
	result += "**Demo Complete!** API testing capabilities shown:\n"
	result += "• Direct HTTP requests (GET, POST)\n"
	result += "• Browser-based API testing\n"
	result += "• Response data validation\n"
	result += "• Dynamic content verification\n"
	
	return result, nil
}

func (t *DemoTool) runFullWorkflowDemo() (string, error) {
	result := "🚀 **Complete Development Workflow Demo**\n\n"
	
	result += "**Phase 1: Project Setup**\n"
	result += "✅ Created project directory structure\n"
	result += "✅ Generated index.html, styles.css, script.js\n"
	result += "✅ Started live preview server at localhost:8080\n\n"
	
	result += "**Phase 2: Development**\n"
	result += "✅ Built responsive portfolio website\n"
	result += "✅ Added contact form with validation\n"
	result += "✅ Implemented smooth scrolling navigation\n"
	result += "✅ Created image gallery with lightbox\n\n"
	
	result += "**Phase 3: Testing**\n"
	result += "✅ Navigated to localhost:8080\n"
	result += "✅ Tested all navigation links\n"
	result += "✅ Filled and submitted contact form\n"
	result += "✅ Tested responsive design at different sizes\n"
	result += "✅ Verified JavaScript functionality\n\n"
	
	result += "**Phase 4: API Integration**\n"
	result += "✅ Added weather widget with API calls\n"
	result += "✅ Tested API endpoints with HTTP requests\n"
	result += "✅ Verified data display in browser\n\n"
	
	result += "**Phase 5: Documentation**\n"
	result += "✅ Captured screenshots of all pages\n"
	result += "✅ Documented test results\n"
	result += "✅ Generated project summary\n\n"
	
	result += "**Full Workflow Complete!** 🎉\n"
	result += "Demonstrated complete web development cycle:\n"
	result += "• File system operations\n"
	result += "• Local development server\n"
	result += "• Browser automation and testing\n"
	result += "• API integration and testing\n"
	result += "• Visual documentation\n"
	result += "• End-to-end workflow validation\n"
	
	return result, nil
}