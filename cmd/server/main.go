package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/mcp"
	"rodmcp/internal/webtools"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
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
		zap.String("version", "1.0.0"),
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

	// Register web development tools
	mcpServer.RegisterTool(webtools.NewCreatePageTool(log))
	mcpServer.RegisterTool(webtools.NewNavigatePageTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewScreenshotTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewExecuteScriptTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewBrowserVisibilityTool(log, browserMgr))
	mcpServer.RegisterTool(webtools.NewLivePreviewTool(log))

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
		"tools_registered": 6,
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
}
