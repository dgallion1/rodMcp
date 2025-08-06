package main

import (
	"encoding/json"
	"fmt"
	"log"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/webtools"
	"time"
)

func main() {
	fmt.Println("🎬 Live Browser Demo - Watch Claude Work!")
	fmt.Println("=========================================")
	
	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    "info",
		LogDir:      "demo_logs",
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

	// VISIBLE browser configuration
	browserConfig := browser.Config{
		Headless:     false,  // 🖥️ BROWSER WILL BE VISIBLE
		Debug:        false,
		SlowMotion:   1 * time.Second,  // Slow for demonstration
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
	_ = webtools.NewLivePreviewTool(logr)

	fmt.Println("👀 Watch the browser window that just opened!")
	fmt.Println("📝 Step 1: Creating a stunning demo webpage...")

	// Create an impressive demo page
	createArgs := map[string]interface{}{
		"filename": "claude_demo.html",
		"title":    "🤖 Claude Live Browser Demo",
		"html": `
			<div class="container">
				<header class="hero">
					<h1>🤖 Claude Live Browser Demo</h1>
					<p>Watch me create and interact with web content in real-time!</p>
					<div class="status" id="status">🚀 Demo loaded - Ready for automation!</div>
				</header>
				
				<section class="demo-area">
					<div class="controls">
						<h2>🎮 Interactive Controls</h2>
						<button id="colorBtn" class="btn primary">🎨 Change Theme</button>
						<button id="animateBtn" class="btn secondary">✨ Animate</button>
						<button id="dataBtn" class="btn success">📊 Load Data</button>
					</div>
					
					<div class="display-area">
						<div id="content-box" class="content-box">
							<h3>Dynamic Content Area</h3>
							<p>This area will be updated by Claude's automation...</p>
							<div id="data-display"></div>
						</div>
					</div>
				</section>
				
				<section class="stats">
					<div class="stat-card">
						<h3 id="clickCount">0</h3>
						<p>Button Clicks</p>
					</div>
					<div class="stat-card">
						<h3 id="themeCount">1</h3>
						<p>Theme Changes</p>
					</div>
					<div class="stat-card">
						<h3 id="animCount">0</h3>
						<p>Animations</p>
					</div>
				</section>
				
				<footer class="activity-log">
					<h3>📋 Live Activity Log</h3>
					<div id="log" class="log-container">
						<p>🚀 Demo initialized - Claude will start automation shortly...</p>
					</div>
				</footer>
			</div>
		`,
		"css": `
			* {
				margin: 0;
				padding: 0;
				box-sizing: border-box;
			}
			
			body {
				font-family: 'Segoe UI', -apple-system, sans-serif;
				background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
				color: white;
				min-height: 100vh;
				transition: all 0.5s ease;
			}
			
			.container {
				max-width: 1200px;
				margin: 0 auto;
				padding: 20px;
			}
			
			.hero {
				text-align: center;
				padding: 40px 0;
				margin-bottom: 30px;
			}
			
			.hero h1 {
				font-size: 3rem;
				margin-bottom: 1rem;
				text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
				animation: glow 2s ease-in-out infinite alternate;
			}
			
			@keyframes glow {
				from { text-shadow: 2px 2px 4px rgba(0,0,0,0.3); }
				to { text-shadow: 2px 2px 20px rgba(255,255,255,0.5); }
			}
			
			.status {
				background: rgba(255,255,255,0.1);
				padding: 15px 30px;
				border-radius: 50px;
				backdrop-filter: blur(10px);
				display: inline-block;
				margin-top: 20px;
				border: 2px solid rgba(255,255,255,0.2);
			}
			
			.demo-area {
				display: grid;
				grid-template-columns: 1fr 2fr;
				gap: 30px;
				margin-bottom: 40px;
			}
			
			.controls, .display-area {
				background: rgba(255,255,255,0.1);
				padding: 30px;
				border-radius: 15px;
				backdrop-filter: blur(10px);
				border: 1px solid rgba(255,255,255,0.2);
			}
			
			.btn {
				display: block;
				width: 100%;
				padding: 15px 25px;
				margin: 10px 0;
				border: none;
				border-radius: 50px;
				font-size: 1.1rem;
				font-weight: bold;
				cursor: pointer;
				transition: all 0.3s ease;
				box-shadow: 0 4px 15px rgba(0,0,0,0.2);
			}
			
			.btn:hover {
				transform: translateY(-2px);
				box-shadow: 0 6px 25px rgba(0,0,0,0.3);
			}
			
			.primary { background: linear-gradient(45deg, #ff6b6b, #ee5a24); color: white; }
			.secondary { background: linear-gradient(45deg, #a29bfe, #6c5ce7); color: white; }
			.success { background: linear-gradient(45deg, #00b894, #00cec9); color: white; }
			
			.content-box {
				background: rgba(0,0,0,0.2);
				padding: 25px;
				border-radius: 10px;
				min-height: 200px;
				transition: all 0.5s ease;
			}
			
			.content-box.animate {
				transform: scale(1.05) rotate(1deg);
				background: rgba(255,215,0,0.2);
				box-shadow: 0 10px 30px rgba(255,215,0,0.3);
			}
			
			.stats {
				display: grid;
				grid-template-columns: repeat(3, 1fr);
				gap: 20px;
				margin-bottom: 30px;
			}
			
			.stat-card {
				background: rgba(255,255,255,0.1);
				padding: 25px;
				border-radius: 15px;
				text-align: center;
				backdrop-filter: blur(10px);
			}
			
			.stat-card h3 {
				font-size: 2.5rem;
				margin-bottom: 10px;
				color: #ffd700;
			}
			
			.activity-log {
				background: rgba(0,0,0,0.2);
				padding: 25px;
				border-radius: 15px;
			}
			
			.log-container {
				background: rgba(0,0,0,0.3);
				padding: 20px;
				border-radius: 10px;
				font-family: 'Courier New', monospace;
				font-size: 0.9rem;
				max-height: 200px;
				overflow-y: auto;
			}
			
			/* Theme variations */
			.theme-ocean { background: linear-gradient(135deg, #2196F3 0%, #21CBF3 100%); }
			.theme-sunset { background: linear-gradient(135deg, #FF5722 0%, #FF9800 100%); }
			.theme-forest { background: linear-gradient(135deg, #4CAF50 0%, #8BC34A 100%); }
			.theme-purple { background: linear-gradient(135deg, #9C27B0 0%, #E91E63 100%); }
		`,
		"javascript": `
			console.log('🤖 Claude Demo Page Loading...');
			
			let stats = {
				clicks: 0,
				themes: 1,
				animations: 0
			};
			
			const themes = [
				{ name: 'default', class: '', gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' },
				{ name: 'ocean', class: 'theme-ocean', gradient: 'linear-gradient(135deg, #2196F3 0%, #21CBF3 100%)' },
				{ name: 'sunset', class: 'theme-sunset', gradient: 'linear-gradient(135deg, #FF5722 0%, #FF9800 100%)' },
				{ name: 'forest', class: 'theme-forest', gradient: 'linear-gradient(135deg, #4CAF50 0%, #8BC34A 100%)' },
				{ name: 'purple', class: 'theme-purple', gradient: 'linear-gradient(135deg, #9C27B0 0%, #E91E63 100%)' }
			];
			
			let currentTheme = 0;
			
			function updateStats() {
				document.getElementById('clickCount').textContent = stats.clicks;
				document.getElementById('themeCount').textContent = stats.themes;
				document.getElementById('animCount').textContent = stats.animations;
			}
			
			function addLog(message) {
				const log = document.getElementById('log');
				const time = new Date().toLocaleTimeString();
				const entry = document.createElement('p');
				entry.textContent = time + ' - ' + message;
				log.insertBefore(entry, log.firstChild);
				
				// Keep only last 8 entries
				while (log.children.length > 8) {
					log.removeChild(log.lastChild);
				}
			}
			
			function updateStatus(message) {
				document.getElementById('status').textContent = message;
			}
			
			document.getElementById('colorBtn').addEventListener('click', function() {
				stats.clicks++;
				stats.themes++;
				currentTheme = (currentTheme + 1) % themes.length;
				
				document.body.className = themes[currentTheme].class;
				document.body.style.background = themes[currentTheme].gradient;
				
				updateStats();
				addLog('🎨 Theme changed to: ' + themes[currentTheme].name);
				updateStatus('🎨 Theme: ' + themes[currentTheme].name);
			});
			
			document.getElementById('animateBtn').addEventListener('click', function() {
				stats.clicks++;
				stats.animations++;
				
				const box = document.getElementById('content-box');
				box.classList.toggle('animate');
				
				updateStats();
				addLog('✨ Content box animation toggled');
				updateStatus('✨ Animation triggered!');
			});
			
			document.getElementById('dataBtn').addEventListener('click', function() {
				stats.clicks++;
				
				const display = document.getElementById('data-display');
				const data = {
					timestamp: new Date().toISOString(),
					userAgent: navigator.userAgent.substring(0, 50) + '...',
					screenSize: screen.width + 'x' + screen.height,
					clicks: stats.clicks,
					random: Math.floor(Math.random() * 1000)
				};
				
				display.innerHTML = '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
				
				updateStats();
				addLog('📊 Data loaded and displayed');
				updateStatus('📊 Data loaded successfully');
			});
			
			// Initial setup
			setTimeout(() => {
				addLog('🚀 All event handlers registered');
				updateStatus('🚀 Ready for Claude automation!');
			}, 1000);
			
			// Global function for Claude to call
			window.getPageState = function() {
				return {
					stats: stats,
					currentTheme: themes[currentTheme].name,
					isContentAnimated: document.getElementById('content-box').classList.contains('animate'),
					logEntries: document.getElementById('log').children.length
				};
			};
			
			console.log('🎉 Claude Demo Page Ready!');
		`,
	}

	result, err := createTool.Execute(createArgs)
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}
	fmt.Println("✅", result.Content[0].Text)

	fmt.Println("🌐 Step 2: Opening the page in visible browser...")
	time.Sleep(2 * time.Second)

	navigateArgs := map[string]interface{}{
		"url": "claude_demo.html",
	}

	result, err = navigateTool.Execute(navigateArgs)
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}
	fmt.Printf("✅ %s\n", result.Content[0].Text)

	fmt.Println("\n🎬 Now watch the browser as I automate interactions!")
	fmt.Println("👀 You should see a beautiful gradient webpage with buttons")
	time.Sleep(3 * time.Second)

	fmt.Println("🎨 Step 3: Clicking 'Change Theme' button...")
	scriptArgs := map[string]interface{}{
		"script": "document.getElementById('colorBtn').click();",
	}

	_, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		fmt.Printf("Click error: %v\n", err)
	}

	time.Sleep(3 * time.Second)

	fmt.Println("✨ Step 4: Triggering animation...")
	scriptArgs = map[string]interface{}{
		"script": "document.getElementById('animateBtn').click();",
	}

	_, err = scriptTool.Execute(scriptArgs)
	time.Sleep(3 * time.Second)

	fmt.Println("📊 Step 5: Loading data display...")
	scriptArgs = map[string]interface{}{
		"script": "document.getElementById('dataBtn').click();",
	}

	_, err = scriptTool.Execute(scriptArgs)
	time.Sleep(3 * time.Second)

	fmt.Println("🔄 Step 6: Cycling through more themes...")
	for i := 1; i <= 3; i++ {
		fmt.Printf("   Theme change %d/3...\n", i)
		scriptArgs = map[string]interface{}{
			"script": "document.getElementById('colorBtn').click();",
		}
		scriptTool.Execute(scriptArgs)
		time.Sleep(2 * time.Second)
	}

	fmt.Println("📊 Step 7: Reading final page state...")
	scriptArgs = map[string]interface{}{
		"script": "window.getPageState()",
	}

	result, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		fmt.Printf("State read error: %v\n", err)
	} else {
		if result.Content[0].Data != nil {
			dataJSON, _ := json.MarshalIndent(result.Content[0].Data, "", "  ")
			fmt.Printf("📈 Final page state:\n%s\n", string(dataJSON))
		}
	}

	fmt.Println("📸 Step 8: Taking a screenshot of the final result...")
	screenshotArgs := map[string]interface{}{
		"filename": "claude_demo_final.png",
	}

	result, err = screenshotTool.Execute(screenshotArgs)
	if err != nil {
		fmt.Printf("Screenshot error: %v\n", err)
	} else {
		fmt.Printf("✅ %s\n", result.Content[0].Text)
	}

	fmt.Println("\n" + "🎉 DEMO COMPLETE! 🎉")
	fmt.Println("====================")
	fmt.Println("👀 You just watched Claude:")
	fmt.Println("   • Create a beautiful interactive webpage")
	fmt.Println("   • Open it in a visible browser")
	fmt.Println("   • Click buttons and trigger animations")
	fmt.Println("   • Change themes dynamically")
	fmt.Println("   • Load and display data")
	fmt.Println("   • Read the page state")
	fmt.Println("   • Take a final screenshot")
	fmt.Println("")
	fmt.Println("📁 Generated files:")
	fmt.Println("   • claude_demo.html - The interactive webpage")
	fmt.Println("   • claude_demo_final.png - Screenshot of the result")
	fmt.Println("")
	fmt.Println("🚀 This is what Claude can do for your web development projects!")
	
	fmt.Println("\n⏱️  Keeping browser open for 15 more seconds so you can explore...")
	time.Sleep(15 * time.Second)
}