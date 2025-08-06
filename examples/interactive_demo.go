package main

import (
	"fmt"
	"log"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/webtools"
	"time"
)

func main() {
	fmt.Println("🔄 Interactive Test - You can click while automation runs!")

	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    "info",
		LogDir:      "interactive_logs",
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
		Headless:     false, // ← VISIBLE BROWSER
		Debug:        false,
		SlowMotion:   500 * time.Millisecond,
		WindowWidth:  1000,
		WindowHeight: 700,
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

	fmt.Println("📝 Creating interactive test page...")

	// Create interactive page
	createArgs := map[string]interface{}{
		"filename": "interactive_test.html",
		"title":    "🤖 Human vs Bot Clicks",
		"html": `
			<div class="container">
				<h1>🤖 Human vs Bot Test</h1>
				<p>Both you and the bot can click buttons. Watch the counters!</p>
				
				<div class="stats">
					<div class="stat-box">
						<h2 id="humanCount">0</h2>
						<p>👤 Human Clicks</p>
					</div>
					<div class="stat-box">
						<h2 id="botCount">0</h2>
						<p>🤖 Bot Clicks</p>
					</div>
				</div>
				
				<div class="buttons">
					<button id="humanBtn" class="human-btn">👤 Click Me (Human)</button>
					<button id="botBtn" class="bot-btn">🤖 Bot Target</button>
				</div>
				
				<div class="log" id="activityLog">
					<h3>📊 Activity Log</h3>
					<div id="logEntries"></div>
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
				background: linear-gradient(135deg, #2c3e50 0%, #3498db 100%);
				color: white;
				padding: 20px;
				min-height: 100vh;
			}
			
			.container {
				max-width: 800px;
				margin: 0 auto;
				text-align: center;
			}
			
			h1 {
				font-size: 2.5rem;
				margin-bottom: 1rem;
				text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
			}
			
			.stats {
				display: flex;
				justify-content: center;
				gap: 30px;
				margin: 30px 0;
			}
			
			.stat-box {
				background: rgba(255, 255, 255, 0.1);
				padding: 30px;
				border-radius: 15px;
				backdrop-filter: blur(10px);
				min-width: 150px;
			}
			
			.stat-box h2 {
				font-size: 3rem;
				margin-bottom: 10px;
				color: #ecf0f1;
			}
			
			.buttons {
				margin: 30px 0;
			}
			
			.human-btn, .bot-btn {
				font-size: 1.2rem;
				padding: 15px 30px;
				margin: 10px;
				border: none;
				border-radius: 50px;
				cursor: pointer;
				transition: all 0.3s ease;
				box-shadow: 0 4px 15px rgba(0,0,0,0.2);
			}
			
			.human-btn {
				background: linear-gradient(45deg, #e74c3c, #c0392b);
				color: white;
			}
			
			.human-btn:hover {
				transform: translateY(-2px);
				box-shadow: 0 6px 20px rgba(231, 76, 60, 0.4);
			}
			
			.bot-btn {
				background: linear-gradient(45deg, #27ae60, #229954);
				color: white;
			}
			
			.bot-btn:hover {
				transform: translateY(-2px);
				box-shadow: 0 6px 20px rgba(39, 174, 96, 0.4);
			}
			
			.log {
				background: rgba(0, 0, 0, 0.2);
				border-radius: 10px;
				padding: 20px;
				margin-top: 30px;
				text-align: left;
				max-height: 200px;
				overflow-y: auto;
			}
			
			.log-entry {
				padding: 5px 0;
				border-bottom: 1px solid rgba(255,255,255,0.1);
				font-size: 0.9rem;
			}
			
			.log-entry:last-child {
				border-bottom: none;
			}
		`,
		"javascript": `
			let humanClicks = 0;
			let botClicks = 0;
			
			function updateCounters() {
				document.getElementById('humanCount').textContent = humanClicks;
				document.getElementById('botCount').textContent = botClicks;
			}
			
			function addLogEntry(message) {
				const logEntries = document.getElementById('logEntries');
				const entry = document.createElement('div');
				entry.className = 'log-entry';
				const time = new Date().toLocaleTimeString();
				entry.textContent = time + ' - ' + message;
				logEntries.insertBefore(entry, logEntries.firstChild);
				
				// Keep only last 10 entries
				while (logEntries.children.length > 10) {
					logEntries.removeChild(logEntries.lastChild);
				}
			}
			
			document.getElementById('humanBtn').addEventListener('click', function() {
				humanClicks++;
				updateCounters();
				addLogEntry('👤 Human clicked! Total: ' + humanClicks);
				console.log('Human click detected:', humanClicks);
			});
			
			document.getElementById('botBtn').addEventListener('click', function() {
				botClicks++;
				updateCounters();
				addLogEntry('🤖 Bot clicked! Total: ' + botClicks);
				console.log('Bot click detected:', botClicks);
			});
			
			// Initial log
			addLogEntry('🚀 Interactive test started - Click away!');
			console.log('Interactive test page loaded');
		`,
	}

	result, err := createTool.Execute(createArgs)
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}
	fmt.Println("✅", result.Content[0].Text)

	fmt.Println("🌐 Opening interactive page...")
	time.Sleep(1 * time.Second)

	// Navigate to the page
	navigateArgs := map[string]interface{}{
		"url": "interactive_test.html",
	}

	result, err = navigateTool.Execute(navigateArgs)
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}
	fmt.Println("✅", result.Content[0].Text)

	fmt.Println("\n🎯 BROWSER IS NOW OPEN!")
	fmt.Println("👤 You can click the RED button manually")
	fmt.Println("🤖 Bot will click the GREEN button automatically")
	fmt.Println("📊 Watch both counters increase!")

	// Let user see and interact for a moment
	time.Sleep(3 * time.Second)

	fmt.Println("\n🤖 Bot starting automated clicks every 2 seconds...")
	fmt.Println("👤 Keep clicking the red button yourself!")

	// Simulate bot clicks every 2 seconds for 20 seconds
	for i := 1; i <= 10; i++ {
		fmt.Printf("🤖 Bot click #%d\n", i)

		// Bot clicks the green button
		scriptArgs := map[string]interface{}{
			"script": "document.getElementById('botBtn').click();",
		}

		_, err = scriptTool.Execute(scriptArgs)
		if err != nil {
			fmt.Printf("❌ Bot click failed: %v\n", err)
		}

		time.Sleep(2 * time.Second)
	}

	fmt.Println("\n📊 Getting final stats...")

	// Get final stats
	scriptArgs := map[string]interface{}{
		"script": `
			({
				humanClicks: humanClicks,
				botClicks: botClicks,
				totalClicks: humanClicks + botClicks,
				winner: humanClicks > botClicks ? "Human" : (botClicks > humanClicks ? "Bot" : "Tie")
			})
		`,
	}

	time.Sleep(1 * time.Second)
	result, err = scriptTool.Execute(scriptArgs)
	if err != nil {
		fmt.Printf("❌ Failed to get stats: %v\n", err)
	} else {
		fmt.Println("📊 Final Results:")
		fmt.Printf("   Result: %s\n", result.Content[0].Text)
	}

	fmt.Println("\n🏁 Test complete! Browser will stay open for 10 more seconds...")
	fmt.Println("👤 Feel free to keep clicking to see real-time updates!")

	time.Sleep(10 * time.Second)
}
