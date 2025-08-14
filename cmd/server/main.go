package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/mcp"
	"rodmcp/internal/webtools"
	"sort"
	"strconv"
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

// daemonize forks the process and runs in the background
func daemonize(pidFile string) error {
	// Check if already running as daemon (child process)
	if os.Getenv("_RODMCP_DAEMON") == "1" {
		return nil // Already in daemon mode
	}

	// Fork the process
	args := append([]string{}, os.Args...)
	cmd := exec.Command(args[0], args[1:]...)
	
	// Set environment variable to identify daemon process
	cmd.Env = append(os.Environ(), "_RODMCP_DAEMON=1")
	
	// Detach from parent terminal
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	
	// Start the child process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	// Write PID file if specified
	if pidFile != "" {
		if err := writePidFile(pidFile, cmd.Process.Pid); err != nil {
			return fmt.Errorf("failed to write PID file: %w", err)
		}
		fmt.Printf("RodMCP daemon started with PID %d (PID file: %s)\n", cmd.Process.Pid, pidFile)
	} else {
		fmt.Printf("RodMCP daemon started with PID %d\n", cmd.Process.Pid)
	}

	// Exit parent process
	os.Exit(0)
	return nil
}

// writePidFile writes the process ID to a file
func writePidFile(pidFile string, pid int) error {
	file, err := os.OpenFile(pidFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))
	return err
}

// removePidFile removes the PID file
func removePidFile(pidFile string) {
	if pidFile != "" {
		os.Remove(pidFile)
	}
}

// loadFileAccessConfig creates file access configuration from command line flags and config file
func loadFileAccessConfig(configFile, allowedPaths, denyPaths string, allowTemp, restrictToWorkDir bool, maxFileSize int64) (*webtools.FileAccessConfig, error) {
	var config *webtools.FileAccessConfig

	// Start with default configuration
	config = webtools.DefaultFileAccessConfig()

	// Load from config file if specified
	if configFile != "" {
		fileData, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}

		if err := json.Unmarshal(fileData, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
		}
	}

	// Override with command line flags if provided
	if allowedPaths != "" {
		config.AllowedPaths = strings.Split(allowedPaths, ",")
		// Trim whitespace from paths
		for i, path := range config.AllowedPaths {
			config.AllowedPaths[i] = strings.TrimSpace(path)
		}
	}

	if denyPaths != "" {
		config.DenyPaths = strings.Split(denyPaths, ",")
		// Trim whitespace from paths
		for i, path := range config.DenyPaths {
			config.DenyPaths[i] = strings.TrimSpace(path)
		}
	}

	// Apply command line overrides
	config.AllowTempFiles = allowTemp
	config.RestrictToWorkingDir = restrictToWorkDir
	config.MaxFileSize = maxFileSize

	// If custom allowed paths are specified, disable working directory restriction
	if allowedPaths != "" {
		config.RestrictToWorkingDir = false
	}

	return config, nil
}

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
		daemon       = flag.Bool("daemon", false, "Run in daemon mode (background process)")
		pidFile      = flag.String("pid-file", "", "Path to PID file for daemon mode")
		
		// File access configuration flags
		configFile        = flag.String("config", "", "Path to configuration file (JSON format)")
		allowedPaths      = flag.String("allowed-paths", "", "Comma-separated list of allowed file paths")
		denyPaths         = flag.String("deny-paths", "", "Comma-separated list of denied file paths")
		allowTemp         = flag.Bool("allow-temp", false, "Allow access to temporary files")
		restrictToWorkDir = flag.Bool("restrict-to-workdir", true, "Restrict file access to working directory only")
		maxFileSize       = flag.Int64("max-file-size", 10485760, "Maximum file size in bytes (default: 10MB)")
	)
	flag.Parse()

	// Handle daemon mode
	if *daemon {
		if err := daemonize(*pidFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
			os.Exit(1)
		}
		// If we reach here, we're in the child process
	}

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
	mcpServer.RegisterTool(webtools.NewTakeElementScreenshotTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewExecuteScriptTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewBrowserVisibilityTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewLivePreviewTool(log))
	
	// Browser UI control tools
	mcpServer.RegisterTool(webtools.NewClickElementTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewTypeTextTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewKeyboardShortcutTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewSwitchTabTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewWaitTool(log))
	mcpServer.RegisterTool(webtools.NewWaitForElementTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewGetElementTextTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewGetElementAttributeTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewScrollTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewHoverElementTool(log, browserMgr))
	
	// Screen scraping tools
	mcpServer.RegisterTool(webtools.NewScreenScrapeTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewExtractTableTool(log, browserMgr))
	
	// Form automation tools
	mcpServer.RegisterTool(webtools.NewFormFillTool(log, browserMgr))
	
	// Advanced waiting tools
	mcpServer.RegisterTool(webtools.NewWaitForConditionTool(log, browserMgr))
	
	// Testing and assertion tools
	mcpServer.RegisterTool(webtools.NewAssertElementTool(log, browserMgr))
	
	// Load file access configuration
	fileConfig, err := loadFileAccessConfig(*configFile, *allowedPaths, *denyPaths, *allowTemp, *restrictToWorkDir, *maxFileSize)
	if err != nil {
		log.Fatal("Failed to load file access configuration", zap.Error(err))
	}

	log.Info("File access configuration loaded",
		zap.Strings("allowed_paths", fileConfig.AllowedPaths),
		zap.Strings("deny_paths", fileConfig.DenyPaths),
		zap.Bool("restrict_to_workdir", fileConfig.RestrictToWorkingDir),
		zap.Bool("allow_temp_files", fileConfig.AllowTempFiles),
		zap.Int64("max_file_size", fileConfig.MaxFileSize))

	// File system tools with path validation
	fileValidator := webtools.NewPathValidator(fileConfig)
	mcpServer.RegisterTool(webtools.NewReadFileTool(log, fileValidator))
	mcpServer.RegisterTool(webtools.NewWriteFileTool(log, fileValidator))
	mcpServer.RegisterTool(webtools.NewListDirectoryTool(log, fileValidator))
	
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
		"tools_registered": 26,
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
	
	// Remove PID file if in daemon mode
	if *daemon {
		removePidFile(*pidFile)
	}
	
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
		daemon       = flag.Bool("daemon", false, "Run in daemon mode (background process)")
		pidFile      = flag.String("pid-file", "", "Path to PID file for daemon mode")
		
		// File access configuration flags
		configFile        = flag.String("config", "", "Path to configuration file (JSON format)")
		allowedPaths      = flag.String("allowed-paths", "", "Comma-separated list of allowed file paths")
		denyPaths         = flag.String("deny-paths", "", "Comma-separated list of denied file paths")
		allowTemp         = flag.Bool("allow-temp", false, "Allow access to temporary files")
		restrictToWorkDir = flag.Bool("restrict-to-workdir", true, "Restrict file access to working directory only")
		maxFileSize       = flag.Int64("max-file-size", 10485760, "Maximum file size in bytes (default: 10MB)")
	)
	flag.CommandLine.Parse(os.Args[2:]) // Skip "rodmcp http"

	// Handle daemon mode
	if *daemon {
		if err := daemonize(*pidFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
			os.Exit(1)
		}
		// If we reach here, we're in the child process
	}

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
	httpServer.RegisterTool(webtools.NewTakeElementScreenshotTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewExecuteScriptTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewBrowserVisibilityTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewLivePreviewTool(log))
	
	// Browser UI control tools
	httpServer.RegisterTool(webtools.NewClickElementTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewTypeTextTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewKeyboardShortcutTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewSwitchTabTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewWaitTool(log))
	httpServer.RegisterTool(webtools.NewWaitForElementTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewGetElementTextTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewGetElementAttributeTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewScrollTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewHoverElementTool(log, browserMgr))
	
	// Screen scraping tools
	httpServer.RegisterTool(webtools.NewScreenScrapeTool(log, browserMgr))
	httpServer.RegisterTool(webtools.NewExtractTableTool(log, browserMgr))
	
	// Form automation tools
	httpServer.RegisterTool(webtools.NewFormFillTool(log, browserMgr))
	
	// Advanced waiting tools
	httpServer.RegisterTool(webtools.NewWaitForConditionTool(log, browserMgr))
	
	// Testing and assertion tools
	httpServer.RegisterTool(webtools.NewAssertElementTool(log, browserMgr))
	
	// Load file access configuration for HTTP server
	fileConfigHTTP, err := loadFileAccessConfig(*configFile, *allowedPaths, *denyPaths, *allowTemp, *restrictToWorkDir, *maxFileSize)
	if err != nil {
		log.Fatal("Failed to load file access configuration", zap.Error(err))
	}

	log.Info("HTTP server file access configuration loaded",
		zap.Strings("allowed_paths", fileConfigHTTP.AllowedPaths),
		zap.Strings("deny_paths", fileConfigHTTP.DenyPaths),
		zap.Bool("restrict_to_workdir", fileConfigHTTP.RestrictToWorkingDir),
		zap.Bool("allow_temp_files", fileConfigHTTP.AllowTempFiles),
		zap.Int64("max_file_size", fileConfigHTTP.MaxFileSize))

	// File system tools with path validation
	fileValidator2 := webtools.NewPathValidator(fileConfigHTTP)
	httpServer.RegisterTool(webtools.NewReadFileTool(log, fileValidator2))
	httpServer.RegisterTool(webtools.NewWriteFileTool(log, fileValidator2))
	httpServer.RegisterTool(webtools.NewListDirectoryTool(log, fileValidator2))
	
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
		"tools_registered": 26,
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
	
	// Remove PID file if in daemon mode
	if *daemon {
		removePidFile(*pidFile)
	}
	
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
	tools["take_element_screenshot"] = webtools.NewTakeElementScreenshotTool(log, browserMgr)
	tools["execute_script"] = webtools.NewExecuteScriptTool(log, browserMgr)
	tools["set_browser_visibility"] = webtools.NewBrowserVisibilityTool(log, browserMgr)
	tools["live_preview"] = webtools.NewLivePreviewTool(log)
	
	// Browser UI control tools
	tools["click_element"] = webtools.NewClickElementTool(log, browserMgr)
	tools["type_text"] = webtools.NewTypeTextTool(log, browserMgr)
	tools["keyboard_shortcuts"] = webtools.NewKeyboardShortcutTool(log, browserMgr)
	tools["switch_tab"] = webtools.NewSwitchTabTool(log, browserMgr)
	tools["wait"] = webtools.NewWaitTool(log)
	tools["wait_for_element"] = webtools.NewWaitForElementTool(log, browserMgr)
	tools["get_element_text"] = webtools.NewGetElementTextTool(log, browserMgr)
	tools["get_element_attribute"] = webtools.NewGetElementAttributeTool(log, browserMgr)
	tools["scroll"] = webtools.NewScrollTool(log, browserMgr)
	tools["hover_element"] = webtools.NewHoverElementTool(log, browserMgr)
	
	// Screen scraping tools
	tools["screen_scrape"] = webtools.NewScreenScrapeTool(log, browserMgr)
	tools["extract_table"] = webtools.NewExtractTableTool(log, browserMgr)
	
	// Form automation tools
	tools["form_fill"] = webtools.NewFormFillTool(log, browserMgr)
	
	// Advanced waiting tools
	tools["wait_for_condition"] = webtools.NewWaitForConditionTool(log, browserMgr)
	
	// Testing and assertion tools
	tools["assert_element"] = webtools.NewAssertElementTool(log, browserMgr)
	
	// File system tools with path validation (use default config for CLI tools)
	fileValidator3 := webtools.NewPathValidator(webtools.DefaultFileAccessConfig())
	tools["read_file"] = webtools.NewReadFileTool(log, fileValidator3)
	tools["write_file"] = webtools.NewWriteFileTool(log, fileValidator3)
	tools["list_directory"] = webtools.NewListDirectoryTool(log, fileValidator3)
	
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
	fmt.Printf(`🤖 RodMCP - Model Context Protocol Server for Web Development

OVERVIEW:
    RodMCP provides comprehensive browser automation and file system access through
    the Model Context Protocol (MCP). It offers 26+ tools for web development,
    testing, and automation with robust security controls.

USAGE:
    %s [COMMAND] [FLAGS]

COMMANDS:
    (default)          Start stdio MCP server for Claude Desktop integration
    version           Show version information and build details  
    http              Start HTTP-based MCP server for API access
    list-tools        List all 26 available tools with descriptions
    describe-tool     Show detailed documentation for a specific tool
    schema            Export complete MCP tool schema as JSON
    help              Show this comprehensive help message

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🖥️  BROWSER CONFIGURATION FLAGS:
    --headless            Run browser in headless mode
                          Default: false (stdio), true (http)
    --debug               Enable browser debug mode and verbose logging
    --slow-motion DURATION Add delay between browser actions (e.g. 100ms)
    --window-width WIDTH  Browser window width in pixels (default: 1920)
    --window-height HEIGHT Browser window height in pixels (default: 1080)

⚙️  PROCESS MANAGEMENT FLAGS:
    --daemon              Run server in daemon mode (background process)
    --pid-file FILE       Path to PID file for daemon mode (optional)

📁 FILE ACCESS SECURITY FLAGS:
    --config FILE         Path to JSON configuration file for advanced settings
    --allowed-paths PATHS Comma-separated list of allowed directory paths
    --deny-paths PATHS    Comma-separated list of explicitly denied paths
    --allow-temp          Allow access to system temporary directory
    --restrict-to-workdir Restrict all file access to current directory only
                          (default: true - automatically disabled if --allowed-paths set)
    --max-file-size BYTES Maximum file size for operations (default: 10485760 = 10MB)

📋 LOGGING & DEBUGGING FLAGS:
    --log-level LEVEL     Set logging verbosity: debug, info, warn, error (default: info)
    --log-dir DIR         Directory for log files (default: logs/)

🌐 HTTP SERVER SPECIFIC FLAGS (for 'rodmcp http'):
    --port PORT           HTTP server port (default: 8080)
    (All browser and file access flags above also apply to HTTP mode)

ENVIRONMENT VARIABLES:
    RODMCP_BROWSER_PATH   Override browser binary path (auto-detected if not set)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔒 FILE ACCESS SECURITY MODEL:

    RodMCP implements defense-in-depth file access controls to prevent unauthorized
    access while maintaining functionality for legitimate automation tasks.

    DEFAULT SECURITY POSTURE:
    ✅ Restricted to current working directory only
    ✅ 10MB file size limit to prevent resource exhaustion  
    ✅ Temporary file access disabled
    ✅ Path traversal attack prevention
    ✅ Symlink resolution with security validation

    CONFIGURATION METHODS:
    1. Command Line Flags (quick setup)
    2. JSON Configuration File (advanced, persistent settings)
    3. Programmatic (modify DefaultFileAccessConfig() in code)

    JSON CONFIG FILE FORMAT:
    {
      "allowed_paths": ["/home/user/projects", "/var/www"],
      "deny_paths": ["/etc", "/root", "/var/log"], 
      "restrict_to_working_dir": false,
      "allow_temp_files": true,
      "max_file_size": 52428800
    }

    SECURITY PRECEDENCE (highest to lowest):
    1. Deny paths (always block, overrides everything)
    2. Command line flags (override config file)
    3. Config file settings (override defaults)
    4. Secure defaults (working directory only)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📖 COMMON USAGE EXAMPLES:

    Basic Server Startup:
    %s                                    # Stdio MCP (Claude Desktop)
    %s http                              # HTTP MCP server on port 8080
    %s http --port 3000                  # HTTP server on custom port
    %s --daemon --pid-file /var/run/rodmcp.pid  # Run as background daemon

    Tool Discovery & Documentation:
    %s list-tools                        # Show all 26 available tools
    %s describe-tool click_element       # Detailed docs for specific tool
    %s schema                            # Export JSON schema for integration

    Browser Configuration:
    %s --headless --debug               # Headless mode with debug logging
    %s --window-width 1280 --window-height 720  # Custom window size
    %s --slow-motion 200ms              # Add delays for debugging

    File Access Examples:
    %s --allowed-paths "/home/user/web,/tmp"     # Allow specific paths
    %s --config security.json                   # Use JSON config file
    %s --allow-temp --max-file-size 50MB        # Allow temp + larger files
    %s --deny-paths "/etc,/root" --allowed-paths "/home"  # Mixed allow/deny

    Development & Debugging:
    %s --log-level debug --log-dir ./logs      # Verbose logging
    %s http --debug --port 8080 --allow-temp   # HTTP debug mode

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🛠️  TOOL CATEGORIES (26 tools total):

    🌐 Browser Automation (7): create_page, navigate_page, take_screenshot,
                               execute_script, set_browser_visibility, live_preview
    🖱️  UI Interaction (4):     click_element, type_text, hover_element, keyboard_shortcuts  
    📑 Tab Management (1):      switch_tab
    ⏳ Timing & Waiting (3):    wait, wait_for_element, wait_for_condition
    📖 Data Extraction (3):     get_element_text, get_element_attribute, scroll
    🕷️  Screen Scraping (2):    screen_scrape, extract_table
    📝 Form Automation (1):     form_fill
    🧪 Testing & Assertions (1): assert_element
    📁 File System (3):         read_file, write_file, list_directory
    🌐 Network (1):             http_request

    Use '%s list-tools' for detailed descriptions of each tool.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔗 Integration & Support:
    
    GitHub: https://github.com/your-org/rodmcp
    MCP Protocol: https://modelcontextprotocol.org
    Claude Desktop Integration: Add to your MCP settings for seamless usage
    
    Version: %s | Build: %s | Go: 1.24.5+ | MCP: 2024-11-05
`, 
		os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], 
		os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], 
		os.Args[0], os.Args[0], os.Args[0], Version, Commit)
}

func listTools() {
	fmt.Println("🛠️  RodMCP Available Tools")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("Total: 26 comprehensive web development tools\n\n")
	
	tools := getAllTools()
	
	// Group tools by category (optimized for LLM clarity)
	categories := map[string][]string{
		"🌐 Browser Automation": {
			"create_page", "navigate_page", "take_screenshot", "take_element_screenshot",
			"execute_script", "set_browser_visibility", "live_preview",
		},
		"🖱️ Browser Interaction": {
			"click_element", "type_text", "hover_element", "keyboard_shortcuts",
		},
		"📑 Tab Management": {
			"switch_tab",
		},
		"⏳ Timing & Waiting": {
			"wait", "wait_for_element", "wait_for_condition",
		},
		"📖 Data Extraction": {
			"get_element_text", "get_element_attribute", "scroll",
		},
		"🕷️ Screen Scraping": {
			"screen_scrape", "extract_table",
		},
		"📝 Form Automation": {
			"form_fill",
		},
		"🧪 Testing & Assertions": {
			"assert_element",
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
	fmt.Println("=" + strings.Repeat("=", len(tool.Name())+10))
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
