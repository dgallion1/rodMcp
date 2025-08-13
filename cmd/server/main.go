package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/mcp"
	"rodmcp/internal/webtools"
	"sort"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Version information (set during build)
var (
	Version   = "1.0.0-dev"  // Default version, can be overridden at build time
	Commit    = "unknown"    // Git commit hash
	BuildDate = "unknown"    // Build timestamp
)

func main() {
	// Check for subcommands first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			showVersion()
			return
		case "list-tools", "tools":
			listTools()
			return
		case "describe-tool":
			if len(os.Args) < 3 {
				fmt.Fprintf(os.Stderr, "Usage: %s describe-tool <tool_name>\n", os.Args[0])
				os.Exit(1)
			}
			describeTool(os.Args[2])
			return
		case "schema":
			exportSchema()
			return
		case "http":
			startHTTPServer()
			return
		case "help", "-h", "--help":
			showHelp()
			return
		}
	}

	// Parse command line flags for server mode
	var (
		logLevel     = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		logDir       = flag.String("log-dir", "logs", "Log directory")
		headless     = flag.Bool("headless", false, "Run browser in headless mode")
		debug        = flag.Bool("debug", false, "Enable browser debug mode")
		slowMotion   = flag.Duration("slow-motion", 0, "Slow motion delay between actions")
		windowWidth  = flag.Int("window-width", 1920, "Browser window width")
		windowHeight = flag.Int("window-height", 1080, "Browser window height")
	)
	flag.Parse()

	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    *logLevel,
		LogDir:      *logDir,
		MaxSize:     100, // 100MB
		MaxBackups:  5,
		MaxAge:      30, // 30 days
		Compress:    true,
		Development: *debug,
	}

	log, err := logger.New(logConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting RodMCP server",
		zap.String("version", Version),
		zap.String("commit", Commit),
		zap.String("log_level", *logLevel),
		zap.Bool("headless", *headless))

	// Initialize browser manager
	browserConfig := browser.Config{
		Headless:     *headless,
		Debug:        *debug,
		SlowMotion:   *slowMotion,
		WindowWidth:  *windowWidth,
		WindowHeight: *windowHeight,
	}

	browserMgr := browser.NewManager(log, browserConfig)
	if err := browserMgr.Start(browserConfig); err != nil {
		log.Fatal("Failed to start browser manager", zap.Error(err))
	}
	defer browserMgr.Stop()

	// Initialize MCP server
	mcpServer := mcp.NewServer(log)

	// Set browser manager for health monitoring
	mcpServer.SetBrowserManager(browserMgr)

	// Register web development tools
	mcpServer.RegisterTool(webtools.NewCreatePageTool(log))
	mcpServer.RegisterTool(webtools.NewNavigatePageTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewScreenshotTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewExecuteScriptTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewBrowserVisibilityTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewLivePreviewTool(log))
	
	// Browser UI control tools
	mcpServer.RegisterTool(webtools.NewClickElementTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewTypeTextTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewWaitTool(log))
	mcpServer.RegisterTool(webtools.NewWaitForElementTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewGetElementTextTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewGetElementAttributeTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewScrollTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewHoverElementTool(log, browserMgr))
	
	// Screen scraping tools
	mcpServer.RegisterTool(webtools.NewScreenScrapeTool(log, browserMgr))
	
	// Form automation tools
	mcpServer.RegisterTool(webtools.NewFormFillTool(log, browserMgr))
	
	// Advanced waiting tools
	mcpServer.RegisterTool(webtools.NewWaitForConditionTool(log, browserMgr))
	
	// Testing and assertion tools
	mcpServer.RegisterTool(webtools.NewAssertElementTool(log, browserMgr))
	
	// File system tools
	mcpServer.RegisterTool(webtools.NewReadFileTool(log))
	mcpServer.RegisterTool(webtools.NewWriteFileTool(log))
	mcpServer.RegisterTool(webtools.NewListDirectoryTool(log))
	
	// Network tools
	mcpServer.RegisterTool(webtools.NewHTTPRequestTool(log))
	
	// Help system
	mcpServer.RegisterTool(webtools.NewHelpTool(log))

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start MCP server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := mcpServer.Start(); err != nil {
			errChan <- err
		}
	}()

	log.Info("RodMCP server started successfully")

	// Send a log message to MCP client
	mcpServer.SendLogMessage("info", "RodMCP server is ready for connections", map[string]interface{}{
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"tools_registered": 19,
		"browser_config": map[string]interface{}{
			"headless":      *headless,
			"debug":         *debug,
			"window_width":  *windowWidth,
			"window_height": *windowHeight,
		},
	})

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Info("Received shutdown signal", zap.String("signal", sig.String()))
	case err := <-errChan:
		log.Error("MCP server error", zap.Error(err))
	}

	log.Info("Shutting down RodMCP server")
	
	// Gracefully stop the MCP server
	if err := mcpServer.Stop(); err != nil {
		log.Error("Error stopping MCP server", zap.Error(err))
	}
}

func startHTTPServer() {
	// Parse HTTP-specific flags
	var (
		port         = flag.Int("port", 8080, "HTTP server port")
		logLevel     = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		logDir       = flag.String("log-dir", "logs", "Log directory")
		headless     = flag.Bool("headless", true, "Run browser in headless mode (default for HTTP)")
		debug        = flag.Bool("debug", false, "Enable browser debug mode")
		slowMotion   = flag.Duration("slow-motion", 0, "Slow motion delay between actions")
		windowWidth  = flag.Int("window-width", 1920, "Browser window width")
		windowHeight = flag.Int("window-height", 1080, "Browser window height")
	)
	flag.CommandLine.Parse(os.Args[2:]) // Skip "rodmcp http"

	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    *logLevel,
		LogDir:      *logDir,
		MaxSize:     100, // 100MB
		MaxBackups:  5,
		MaxAge:      30, // 30 days
		Compress:    true,
		Development: *debug,
	}

	log, err := logger.New(logConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting RodMCP HTTP server",
		zap.String("version", Version),
		zap.String("commit", Commit),
		zap.Int("port", *port),
		zap.String("log_level", *logLevel),
		zap.Bool("headless", *headless))

	// Initialize browser manager
	browserConfig := browser.Config{
		Headless:     *headless,
		Debug:        *debug,
		SlowMotion:   *slowMotion,
		WindowWidth:  *windowWidth,
		WindowHeight: *windowHeight,
	}

	browserMgr := browser.NewManager(log, browserConfig)
	if err := browserMgr.Start(browserConfig); err != nil {
		log.Fatal("Failed to start browser manager", zap.Error(err))
	}
	defer browserMgr.Stop()

	// Initialize HTTP MCP server
	httpServer := mcp.NewHTTPServer(log, *port)

	// Register web development tools
	httpServer.RegisterTool(webtools.NewCreatePageTool(log))
	httpServer.RegisterTool(webtools.NewNavigatePageTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewScreenshotTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewExecuteScriptTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewBrowserVisibilityTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewLivePreviewTool(log))
	
	// Browser UI control tools
	httpServer.RegisterTool(webtools.NewClickElementTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewTypeTextTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewWaitTool(log))
	httpServer.RegisterTool(webtools.NewWaitForElementTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewGetElementTextTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewGetElementAttributeTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewScrollTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewHoverElementTool(log, browserMgr))
	
	// Screen scraping tools
	httpServer.RegisterTool(webtools.NewScreenScrapeTool(log, browserMgr))
	
	// Form automation tools
	httpServer.RegisterTool(webtools.NewFormFillTool(log, browserMgr))
	
	// Advanced waiting tools
	httpServer.RegisterTool(webtools.NewWaitForConditionTool(log, browserMgr))
	
	// Testing and assertion tools
	httpServer.RegisterTool(webtools.NewAssertElementTool(log, browserMgr))
	
	// File system tools
	httpServer.RegisterTool(webtools.NewReadFileTool(log))
	httpServer.RegisterTool(webtools.NewWriteFileTool(log))
	httpServer.RegisterTool(webtools.NewListDirectoryTool(log))
	
	// Network tools
	httpServer.RegisterTool(webtools.NewHTTPRequestTool(log))
	
	// Help system
	httpServer.RegisterTool(webtools.NewHelpTool(log))

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	log.Info("RodMCP HTTP server started successfully",
		zap.String("url", fmt.Sprintf("http://localhost:%d", *port)))

	// Send a log message
	httpServer.SendLogMessage("info", "RodMCP HTTP server is ready for connections", map[string]interface{}{
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"port":            *port,
		"tools_registered": 19,
		"browser_config": map[string]interface{}{
			"headless":      *headless,
			"debug":         *debug,
			"window_width":  *windowWidth,
			"window_height": *windowHeight,
		},
	})

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Info("Received shutdown signal", zap.String("signal", sig.String()))
	case err := <-errChan:
		log.Error("HTTP server error", zap.Error(err))
	}

	log.Info("Shutting down RodMCP HTTP server")
	if err := httpServer.Stop(); err != nil {
		log.Error("Error stopping HTTP server", zap.Error(err))
	}
}

// Helper function to get all registered tools
func getAllTools() map[string]mcp.Tool {
	// Create a temporary logger just for tool registration
	logConfig := logger.Config{
		LogLevel:    "error", // Minimize logging for CLI commands
		LogDir:      "logs",
		MaxSize:     10,
		MaxBackups:  3,
		MaxAge:      28,
		Compress:    true,
		Development: false,
	}
	
	log, err := logger.New(logConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	
	// Create minimal browser manager (won't actually start browser for CLI)
	browserConfig := browser.Config{
		Headless:     true,
		Debug:        false,
		SlowMotion:   0,
		WindowWidth:  1920,
		WindowHeight: 1080,
	}
	browserMgr := browser.NewManager(log, browserConfig)
	
	// Register all tools
	tools := make(map[string]mcp.Tool)
	
	// Browser automation tools
	tools["create_page"] = webtools.NewCreatePageTool(log)
	tools["navigate_page"] = webtools.NewNavigatePageTool(log, browserMgr)
	tools["take_screenshot"] = webtools.NewScreenshotTool(log, browserMgr)
	tools["execute_script"] = webtools.NewExecuteScriptTool(log, browserMgr)
	tools["set_browser_visibility"] = webtools.NewBrowserVisibilityTool(log, browserMgr)
	tools["live_preview"] = webtools.NewLivePreviewTool(log)
	
	// Browser UI control tools
	tools["click_element"] = webtools.NewClickElementTool(log, browserMgr)
	tools["type_text"] = webtools.NewTypeTextTool(log, browserMgr)
	tools["wait"] = webtools.NewWaitTool(log)
	tools["wait_for_element"] = webtools.NewWaitForElementTool(log, browserMgr)
	tools["get_element_text"] = webtools.NewGetElementTextTool(log, browserMgr)
	tools["get_element_attribute"] = webtools.NewGetElementAttributeTool(log, browserMgr)
	tools["scroll"] = webtools.NewScrollTool(log, browserMgr)
	tools["hover_element"] = webtools.NewHoverElementTool(log, browserMgr)
	
	// Screen scraping tools
	tools["screen_scrape"] = webtools.NewScreenScrapeTool(log, browserMgr)
	
	// Form automation tools
	tools["form_fill"] = webtools.NewFormFillTool(log, browserMgr)
	
	// Advanced waiting tools
	tools["wait_for_condition"] = webtools.NewWaitForConditionTool(log, browserMgr)
	
	// Testing and assertion tools
	tools["assert_element"] = webtools.NewAssertElementTool(log, browserMgr)
	
	// File system tools
	tools["read_file"] = webtools.NewReadFileTool(log)
	tools["write_file"] = webtools.NewWriteFileTool(log)
	tools["list_directory"] = webtools.NewListDirectoryTool(log)
	
	// Network tools
	tools["http_request"] = webtools.NewHTTPRequestTool(log)
	
	// Help system
	tools["help"] = webtools.NewHelpTool(log)
	
	return tools
}

func showVersion() {
	fmt.Printf("RodMCP %s\n", Version)
	if Commit != "unknown" {
		fmt.Printf("Git commit: %s\n", Commit)
	}
	if BuildDate != "unknown" {
		fmt.Printf("Build date: %s\n", BuildDate)
	}
	fmt.Printf("Go version: %s\n", "1.24.5+")
	fmt.Printf("MCP protocol version: 2024-11-05\n")
}

func showHelp() {
	fmt.Printf(`RodMCP - Model Context Protocol Server for Web Development

USAGE:
    %s [COMMAND] [FLAGS]

COMMANDS:
    (no command)       Start MCP server (default)
    version           Show version information
    http              Start HTTP-based MCP server
    list-tools, tools  List all available MCP tools
    describe-tool NAME Show detailed documentation for a specific tool
    schema            Export complete MCP tool schema as JSON
    help              Show this help message

SERVER FLAGS (for default MCP server):
    --headless            Run browser in headless mode (default: false)
    --debug               Enable browser debug mode (default: false)
    --log-level LEVEL     Log level: debug, info, warn, error (default: info)
    --log-dir DIR         Log directory (default: logs)
    --slow-motion DURATION Slow motion delay between actions
    --window-width WIDTH  Browser window width (default: 1920)
    --window-height HEIGHT Browser window height (default: 1080)

ENVIRONMENT VARIABLES:
    RODMCP_BROWSER_PATH   Override browser binary path (optional)

HTTP SERVER FLAGS (for 'rodmcp http'):
    --port PORT           HTTP server port (default: 8080)
    --headless            Run browser in headless mode (default: true for HTTP)
    --debug               Enable browser debug mode (default: false)
    --log-level LEVEL     Log level: debug, info, warn, error (default: info)
    --log-dir DIR         Log directory (default: logs)
    --slow-motion DURATION Slow motion delay between actions
    --window-width WIDTH  Browser window width (default: 1920)
    --window-height HEIGHT Browser window height (default: 1080)
    (Also supports RODMCP_BROWSER_PATH environment variable)

EXAMPLES:
    %s                    # Start stdio MCP server
    %s http              # Start HTTP MCP server on port 8080
    %s http --port 3000  # Start HTTP MCP server on port 3000
    %s list-tools        # Show all available tools
    %s describe-tool click_element  # Show tool documentation
    %s schema            # Export tool definitions
    %s --headless        # Start stdio server in headless mode

For more information, see: https://github.com/your-org/rodmcp
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func listTools() {
	fmt.Println("🛠️  RodMCP Available Tools")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("Total: 19 comprehensive web development tools\n\n")
	
	tools := getAllTools()
	
	// Group tools by category
	categories := map[string][]string{
		"🌐 Browser Automation": {
			"create_page", "navigate_page", "take_screenshot", 
			"execute_script", "set_browser_visibility", "live_preview",
		},
		"🎯 Browser UI Control": {
			"click_element", "type_text", "wait", "wait_for_element",
			"get_element_text", "get_element_attribute", "scroll", "hover_element",
		},
		"🕷️ Screen Scraping": {
			"screen_scrape",
		},
		"📁 File System": {
			"read_file", "write_file", "list_directory",
		},
		"🌐 Network": {
			"http_request",
		},
	}
	
	for category, toolNames := range categories {
		fmt.Printf("%s (%d tools)\n", category, len(toolNames))
		fmt.Println(strings.Repeat("-", 40))
		
		for _, name := range toolNames {
			if tool, exists := tools[name]; exists {
				fmt.Printf("  %-20s %s\n", name, tool.Description())
			}
		}
		fmt.Println()
	}
	
	fmt.Printf("📋 Usage Examples:\n")
	fmt.Printf("  %s describe-tool click_element  # Get detailed docs\n", os.Args[0])
	fmt.Printf("  %s schema                      # Export JSON schema\n", os.Args[0])
}

func describeTool(toolName string) {
	tools := getAllTools()
	
	tool, exists := tools[toolName]
	if !exists {
		fmt.Fprintf(os.Stderr, "❌ Tool '%s' not found.\n\n", toolName)
		fmt.Fprintf(os.Stderr, "Available tools:\n")
		
		var names []string
		for name := range tools {
			names = append(names, name)
		}
		sort.Strings(names)
		
		for _, name := range names {
			fmt.Fprintf(os.Stderr, "  - %s\n", name)
		}
		os.Exit(1)
	}
	
	schema := tool.InputSchema()
	
	fmt.Printf("🛠️  Tool: %s\n", tool.Name())
	fmt.Printf("=" + strings.Repeat("=", len(tool.Name())+10) + "\n")
	fmt.Printf("📖 Description: %s\n\n", tool.Description())
	
	fmt.Printf("📋 Parameters:\n")
	if schema.Required != nil && len(schema.Required) > 0 {
		fmt.Printf("  Required: %s\n", strings.Join(schema.Required, ", "))
	} else {
		fmt.Printf("  Required: (none)\n")
	}
	
	if props := schema.Properties; props != nil {
		fmt.Println()
		for paramName, paramDef := range props {
			if paramMap, ok := paramDef.(map[string]interface{}); ok {
				paramType := "unknown"
				if t, ok := paramMap["type"].(string); ok {
					paramType = t
				}
				
				description := ""
				if d, ok := paramMap["description"].(string); ok {
					description = d
				}
				
				required := ""
				if schema.Required != nil {
					for _, req := range schema.Required {
						if req == paramName {
							required = " (required)"
							break
						}
					}
				}
				
				fmt.Printf("  %-15s [%s]%s\n", paramName, paramType, required)
				if description != "" {
					fmt.Printf("                  %s\n", description)
				}
				
				// Show default value if present
				if def, ok := paramMap["default"]; ok {
					fmt.Printf("                  Default: %v\n", def)
				}
				
				// Show constraints
				if min, ok := paramMap["minimum"]; ok {
					fmt.Printf("                  Minimum: %v\n", min)
				}
				if max, ok := paramMap["maximum"]; ok {
					fmt.Printf("                  Maximum: %v\n", max)
				}
				
				fmt.Println()
			}
		}
	}
	
	fmt.Printf("💡 Example Usage:\n")
	switch tool.Name() {
	case "click_element":
		fmt.Printf(`  {"selector": "#submit-button"}
  {"selector": ".menu-item", "timeout": 5}`)
	case "type_text":
		fmt.Printf(`  {"selector": "#email", "text": "user@example.com"}
  {"selector": "input[name='password']", "text": "secret", "clear": false}`)
	case "wait":
		fmt.Printf(`  {"seconds": 3}
  {"seconds": 0.5}`)
	case "http_request":
		fmt.Printf(`  {"url": "https://api.example.com/users", "method": "GET"}
  {"url": "https://api.example.com/users", "method": "POST", "json": {"name": "John"}}`)
	case "read_file":
		fmt.Printf(`  {"path": "index.html"}
  {"path": "./src/components/header.js"}`)
	default:
		fmt.Printf("  (Use 'rodmcp schema' to see complete parameter specifications)")
	}
	
	fmt.Println()
}

func exportSchema() {
	tools := getAllTools()
	
	// Create MCP-compatible schema
	schema := map[string]interface{}{
		"tools": make([]map[string]interface{}, 0, len(tools)),
	}
	
	// Sort tools by name for consistent output
	var names []string
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)
	
	for _, name := range names {
		tool := tools[name]
		toolSchema := map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"inputSchema": tool.InputSchema(),
		}
		schema["tools"] = append(schema["tools"].([]map[string]interface{}), toolSchema)
	}
	
	// Output JSON
	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating schema: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(string(output))
}
