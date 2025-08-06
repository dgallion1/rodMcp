package main

import (
	"encoding/json"
	"fmt"
	"log"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/webtools"
	"strings"
	"time"
)

func main() {
	fmt.Println("🤖 LLM Browser Capabilities Demo")
	fmt.Println("Let's see what the LLM can and cannot do...")
	
	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    "info",
		LogDir:      "llm_demo_logs",
		MaxSize:     10,
		MaxBackups:  3,
		MaxAge:      1,
		Compress:    false,
		Development: true,
	}

	logr, err := logger.New(logConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logr.Sync()

	// Initialize browser manager with VISIBLE window
	browserConfig := browser.Config{
		Headless:     false,
		Debug:        false,
		SlowMotion:   1 * time.Second,
		WindowWidth:  1200,
		WindowHeight: 800,
	}

	browserMgr := browser.NewManager(logr, browserConfig)
	if err := browserMgr.Start(browserConfig); err != nil {
		log.Fatalf("Failed to start browser: %v", err)
	}
	defer browserMgr.Stop()

	// Initialize tools
	createTool := webtools.NewCreatePageTool(logr)
	navigateTool := webtools.NewNavigatePageTool(logr, browserMgr)
	scriptTool := webtools.NewExecuteScriptTool(logr, browserMgr)
	screenshotTool := webtools.NewScreenshotTool(logr, browserMgr)

	fmt.Println("📝 Creating test page with various elements...")

	// Create test page
	createArgs := map[string]interface{}{
		"filename": "llm_capabilities.html",
		"title":    "🤖 What Can the LLM See and Do?",
		"html": `
			<div class="container">
				<h1>🤖 LLM Browser Capabilities</h1>
				<p>Testing what the AI can detect and control</p>
				
				<div class="test-section">
					<h2>🎯 Click Targets</h2>
					<button id="btn1" onclick="handleClick('btn1')">Button 1</button>
					<button id="btn2" onclick="handleClick('btn2')">Button 2</button>
					<button id="btn3" onclick="handleClick('btn3')">Hidden Button</button>
				</div>
				
				<div class="test-section">
					<h2>📊 State Tracking</h2>
					<p>Clicks detected: <span id="clickCount">0</span></p>
					<p>Last clicked: <span id="lastClicked">None</span></p>
					<p>Timestamp: <span id="timestamp">Never</span></p>
				</div>
				
				<div class="test-section">
					<h2>🔤 Form Elements</h2>
					<input type="text" id="textInput" placeholder="Type something..." />
					<select id="dropdown">
						<option value="option1">Option 1</option>
						<option value="option2">Option 2</option>
						<option value="option3">Option 3</option>
					</select>
					<textarea id="textArea" placeholder="Large text area..."></textarea>
				</div>
				
				<div class="test-section">
					<h2>📋 Activity Log</h2>
					<div id="activityLog" class="log-area">
						<p>🚀 Page loaded - Ready for testing!</p>
					</div>
				</div>
				
				<div class="test-section">
					<h2>🎮 Manual Testing Area</h2>
					<p>👤 <strong>You can:</strong> Click any button, type in fields, change dropdown</p>
					<p>🤖 <strong>LLM can:</strong> Click buttons, read values, take screenshots, run JavaScript</p>
				</div>
			</div>
		`,
		"css": `
			* {
				margin: 0;
				padding: 0;
				box-sizing: border-box;
			}
			
			body {
				font-family: 'Segoe UI', system-ui, sans-serif;
				background: linear-gradient(135deg, #1e3c72 0%, #2a5298 100%);
				color: white;
				padding: 20px;
				line-height: 1.6;
			}
			
			.container {
				max-width: 1000px;
				margin: 0 auto;
			}
			
			h1 {
				text-align: center;
				font-size: 2.5rem;
				margin-bottom: 1rem;
				text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
			}
			
			.test-section {
				background: rgba(255, 255, 255, 0.1);
				margin: 20px 0;
				padding: 20px;
				border-radius: 10px;
				backdrop-filter: blur(10px);
			}
			
			.test-section h2 {
				margin-bottom: 15px;
				color: #ffd700;
			}
			
			button {
				background: linear-gradient(45deg, #ff6b6b, #ee5a24);
				color: white;
				border: none;
				padding: 12px 24px;
				margin: 5px;
				border-radius: 25px;
				cursor: pointer;
				font-size: 1rem;
				transition: all 0.3s ease;
			}
			
			button:hover {
				transform: translateY(-2px);
				box-shadow: 0 4px 15px rgba(255, 107, 107, 0.4);
			}
			
			#btn3 {
				background: linear-gradient(45deg, #6c5ce7, #a29bfe);
				display: none; /* Hidden by default */
			}
			
			input, select, textarea {
				padding: 10px;
				margin: 5px;
				border: none;
				border-radius: 5px;
				font-size: 1rem;
				background: rgba(255, 255, 255, 0.9);
				color: #333;
			}
			
			textarea {
				width: 300px;
				height: 80px;
				resize: vertical;
			}
			
			.log-area {
				background: rgba(0, 0, 0, 0.3);
				padding: 15px;
				border-radius: 5px;
				font-family: monospace;
				font-size: 0.9rem;
				max-height: 150px;
				overflow-y: auto;
			}
			
			.highlight {
				background: rgba(255, 215, 0, 0.3);
				padding: 2px 5px;
				border-radius: 3px;
			}
		`,
		"javascript": `
			let clickCount = 0;
			let activityLog = [];
			
			function updateDisplay() {
				document.getElementById('clickCount').textContent = clickCount;
				document.getElementById('timestamp').textContent = new Date().toLocaleTimeString();
			}
			
			function addToLog(message) {
				activityLog.unshift(new Date().toLocaleTimeString() + ' - ' + message);
				if (activityLog.length > 10) activityLog.pop();
				
				document.getElementById('activityLog').innerHTML = 
					activityLog.map(entry => '<p>' + entry + '</p>').join('');
			}
			
			function handleClick(buttonId) {
				clickCount++;
				document.getElementById('lastClicked').textContent = buttonId;
				updateDisplay();
				addToLog('🖱️ ' + buttonId + ' clicked (Total: ' + clickCount + ')');
				console.log('Button clicked:', buttonId, 'Total clicks:', clickCount);
				
				// Show hidden button after 3 clicks
				if (clickCount === 3) {
					document.getElementById('btn3').style.display = 'inline-block';
					addToLog('🎉 Hidden button revealed!');
				}
			}
			
			// Track form interactions
			document.getElementById('textInput').addEventListener('input', function(e) {
				addToLog('⌨️ Text input: "' + e.target.value + '"');
			});
			
			document.getElementById('dropdown').addEventListener('change', function(e) {
				addToLog('🔽 Dropdown changed to: ' + e.target.value);
			});
			
			document.getElementById('textArea').addEventListener('input', function(e) {
				if (e.target.value.length > 0) {
					addToLog('📝 Textarea updated (' + e.target.value.length + ' chars)');
				}
			});
			
			// Add some dynamic content
			setTimeout(() => {
				addToLog('⏱️ 2 seconds have passed');
			}, 2000);
			
			setTimeout(() => {
				addToLog('🔍 LLM can read this message that appeared after 5 seconds!');
			}, 5000);
			
			// Global state for LLM to access
			window.getPageState = function() {
				return {
					clickCount: clickCount,
					lastClicked: document.getElementById('lastClicked').textContent,
					textInputValue: document.getElementById('textInput').value,
					dropdownValue: document.getElementById('dropdown').value,
					textAreaValue: document.getElementById('textArea').value,
					isHiddenButtonVisible: document.getElementById('btn3').style.display !== 'none',
					activityLogCount: activityLog.length,
					pageLoadTime: document.readyState
				};
			};
			
			console.log('🤖 LLM Capabilities Demo loaded - Ready for testing!');
		`,
	}

	result, err := createTool.Execute(createArgs)
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}
	fmt.Println("✅", result.Content[0].Text)

	fmt.Println("🌐 Opening test page...")
	time.Sleep(1 * time.Second)

	// Navigate to the page
	navigateArgs := map[string]interface{}{
		"url": "llm_capabilities.html",
	}

	result, err = navigateTool.Execute(navigateArgs)
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}
	fmt.Println("✅", result.Content[0].Text)

	fmt.Println("\n🎯 BROWSER IS NOW OPEN!")
	fmt.Println("👤 Go ahead and interact with the page - click buttons, type text, change dropdown")
	fmt.Println("🤖 I'll demonstrate what I can see and do...")

	time.Sleep(3 * time.Second)

	fmt.Println("\n🔍 LLM Action #1: Reading initial page state...")
	scriptArgs := map[string]interface{}{
		"script": "window.getPageState()",
	}

	result, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		fmt.Printf("❌ Failed to get state: %v\n", err)
	} else {
		if result.Content[0].Data != nil {
			dataJSON, _ := json.MarshalIndent(result.Content[0].Data, "", "  ")
			fmt.Printf("📊 Current page state:\n%s\n", string(dataJSON))
		}
	}

	time.Sleep(2 * time.Second)

	fmt.Println("\n🤖 LLM Action #2: Clicking Button 1...")
	scriptArgs = map[string]interface{}{
		"script": "document.getElementById('btn1').click(); 'Button 1 clicked by LLM'",
	}

	result, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		fmt.Printf("❌ Click failed: %v\n", err)
	} else {
		fmt.Printf("✅ %s\n", result.Content[0].Text)
	}

	time.Sleep(2 * time.Second)

	fmt.Println("\n🤖 LLM Action #3: Filling out form fields...")
	scriptArgs = map[string]interface{}{
		"script": `
			document.getElementById('textInput').value = 'Hello from the LLM!';
			document.getElementById('textInput').dispatchEvent(new Event('input'));
			document.getElementById('dropdown').value = 'option2';
			document.getElementById('dropdown').dispatchEvent(new Event('change'));
			document.getElementById('textArea').value = 'The LLM can read and write to all form fields automatically.';
			document.getElementById('textArea').dispatchEvent(new Event('input'));
			'Form fields updated by LLM'
		`,
	}

	result, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		fmt.Printf("❌ Form update failed: %v\n", err)
	} else {
		fmt.Printf("✅ %s\n", result.Content[0].Text)
	}

	time.Sleep(2 * time.Second)

	fmt.Println("\n🤖 LLM Action #4: Clicking more buttons to reveal hidden content...")
	scriptArgs = map[string]interface{}{
		"script": `
			document.getElementById('btn2').click();
			setTimeout(() => document.getElementById('btn1').click(), 100);
			'Clicked multiple buttons to reveal hidden content'
		`,
	}

	result, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		fmt.Printf("❌ Multiple clicks failed: %v\n", err)
	} else {
		fmt.Printf("✅ %s\n", result.Content[0].Text)
	}

	time.Sleep(2 * time.Second)

	fmt.Println("\n🔍 LLM Action #5: Reading updated state after interactions...")
	scriptArgs = map[string]interface{}{
		"script": "window.getPageState()",
	}

	result, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		fmt.Printf("❌ Failed to get updated state: %v\n", err)
	} else {
		if result.Content[0].Data != nil {
			dataJSON, _ := json.MarshalIndent(result.Content[0].Data, "", "  ")
			fmt.Printf("📊 Updated page state:\n%s\n", string(dataJSON))
		}
	}

	time.Sleep(1 * time.Second)

	fmt.Println("\n📸 LLM Action #6: Taking a screenshot to 'see' the page...")
	screenshotArgs := map[string]interface{}{
		"filename": "llm_capabilities_screenshot.png",
	}

	result, err = screenshotTool.Execute(screenshotArgs)
	if err != nil {
		fmt.Printf("❌ Screenshot failed: %v\n", err)
	} else {
		fmt.Printf("✅ %s\n", result.Content[0].Text)
	}

	time.Sleep(1 * time.Second)

	fmt.Println("\n🤖 LLM Action #7: Reading activity log to see what happened...")
	scriptArgs = map[string]interface{}{
		"script": `
			const logElement = document.getElementById('activityLog');
			const logText = logElement.innerText;
			({
				activityLogContent: logText,
				totalLogLines: logText.split('\\n').length,
				containsLLMActions: logText.includes('LLM') || logText.includes('input') || logText.includes('clicked')
			})
		`,
	}

	result, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		fmt.Printf("❌ Failed to read log: %v\n", err)
	} else {
		if result.Content[0].Data != nil {
			dataJSON, _ := json.MarshalIndent(result.Content[0].Data, "", "  ")
			fmt.Printf("📋 Activity log analysis:\n%s\n", string(dataJSON))
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🤖 WHAT THE LLM CAN DO:")
	fmt.Println("✅ Click any element programmatically")
	fmt.Println("✅ Read all text content and form values")
	fmt.Println("✅ Fill forms and interact with inputs")
	fmt.Println("✅ Execute JavaScript code")
	fmt.Println("✅ Take screenshots to 'see' the page")
	fmt.Println("✅ Read dynamic content and state changes")
	fmt.Println("✅ Detect changes made by human users")
	
	fmt.Println("\n🤖 WHAT THE LLM CANNOT DO:")
	fmt.Println("❌ See your mouse cursor position")
	fmt.Println("❌ Detect mouse movements without clicks")
	fmt.Println("❌ React in real-time to your actions")
	fmt.Println("❌ See content outside the browser window")
	fmt.Println("❌ Access your keyboard input outside the browser")
	
	fmt.Println("\n📊 INTERACTION SUMMARY:")
	fmt.Println("• The LLM reads page state through JavaScript")
	fmt.Println("• Screenshots provide visual 'snapshots'")  
	fmt.Println("• All interactions are programmatic via browser automation")
	fmt.Println("• Both human and LLM actions are logged by the page")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n⏱️  Browser will stay open for 10 more seconds for final testing...")
	time.Sleep(10 * time.Second)
}