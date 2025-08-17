package main

import (
	"fmt"
	"log"
	"os"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"time"
)

func main() {
	fmt.Println("=== Browser Headless Mode Test ===")
	fmt.Println()

	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    "debug",
		LogDir:      "test_logs",
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

	// Test 1: Start in headless mode
	fmt.Println("TEST 1: Starting browser in HEADLESS mode...")
	testHeadlessMode(logr, true)

	// Test 2: Start in visible mode
	fmt.Println("\nTEST 2: Starting browser in VISIBLE mode...")
	testHeadlessMode(logr, false)

	// Test 3: Dynamic visibility switching
	fmt.Println("\nTEST 3: Testing dynamic visibility switching...")
	testDynamicVisibility(logr)

	// Test 4: Check environment variables
	fmt.Println("\nTEST 4: Checking environment variables...")
	checkEnvironment()

	fmt.Println("\n=== All tests completed ===")
}

func testHeadlessMode(logr *logger.Logger, headless bool) {
	mode := "VISIBLE"
	if headless {
		mode = "HEADLESS"
	}

	browserConfig := browser.Config{
		Headless:     headless,
		Debug:        true,
		SlowMotion:   0,
		WindowWidth:  1280,
		WindowHeight: 720,
	}

	fmt.Printf("  Initializing browser manager (headless=%v)...\n", headless)
	browserMgr := browser.NewManager(logr, browserConfig)

	fmt.Printf("  Starting browser in %s mode...\n", mode)
	if err := browserMgr.Start(browserConfig); err != nil {
		fmt.Printf("  ❌ Failed to start browser: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Browser started successfully in %s mode\n", mode)

	// Create a test page
	fmt.Printf("  Creating test page...\n")
	_, pageID, err := browserMgr.NewPage("https://example.com")
	if err != nil {
		fmt.Printf("  ❌ Failed to create page: %v\n", err)
		browserMgr.Stop()
		return
	}

	fmt.Printf("  ✅ Page created (ID: %s)\n", pageID)

	// Take screenshot to verify browser is working
	fmt.Printf("  Taking screenshot...\n")
	screenshot, err := browserMgr.Screenshot(pageID)
	if err != nil {
		fmt.Printf("  ❌ Failed to take screenshot: %v\n", err)
	} else {
		fmt.Printf("  ✅ Screenshot taken (%d bytes)\n", len(screenshot))

		// Save screenshot
		filename := fmt.Sprintf("test_screenshot_%s.png", mode)
		if err := os.WriteFile(filename, screenshot, 0644); err != nil {
			fmt.Printf("  ⚠️  Failed to save screenshot: %v\n", err)
		} else {
			fmt.Printf("  💾 Screenshot saved to %s\n", filename)
		}
	}

	// Get page info
	info, err := browserMgr.GetPageInfo(pageID)
	if err != nil {
		fmt.Printf("  ❌ Failed to get page info: %v\n", err)
	} else {
		fmt.Printf("  ✅ Page info: %v\n", info)
	}

	// Keep browser open for a moment in visible mode
	if !headless {
		fmt.Printf("  🔍 Browser window should be visible for 5 seconds...\n")
		time.Sleep(5 * time.Second)
	} else {
		time.Sleep(1 * time.Second)
	}

	// Close page
	fmt.Printf("  Closing page...\n")
	if err := browserMgr.ClosePage(pageID); err != nil {
		fmt.Printf("  ❌ Failed to close page: %v\n", err)
	} else {
		fmt.Printf("  ✅ Page closed\n")
	}

	// Stop browser
	fmt.Printf("  Stopping browser...\n")
	if err := browserMgr.Stop(); err != nil {
		fmt.Printf("  ❌ Failed to stop browser: %v\n", err)
	} else {
		fmt.Printf("  ✅ Browser stopped\n")
	}
}

func testDynamicVisibility(logr *logger.Logger) {
	// Start in headless mode
	browserConfig := browser.Config{
		Headless:     true,
		Debug:        false,
		SlowMotion:   0,
		WindowWidth:  1280,
		WindowHeight: 720,
	}

	fmt.Printf("  Starting browser in HEADLESS mode...\n")
	browserMgr := browser.NewManager(logr, browserConfig)
	if err := browserMgr.Start(browserConfig); err != nil {
		fmt.Printf("  ❌ Failed to start browser: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Browser started in HEADLESS mode\n")

	// Create a test page
	_, pageID, err := browserMgr.NewPage("https://example.com")
	if err != nil {
		fmt.Printf("  ❌ Failed to create page: %v\n", err)
		browserMgr.Stop()
		return
	}
	fmt.Printf("  ✅ Page created (ID: %s)\n", pageID)

	// Switch to visible mode
	fmt.Printf("  Switching to VISIBLE mode...\n")
	if err := browserMgr.SetVisibility(true); err != nil {
		fmt.Printf("  ❌ Failed to switch to visible mode: %v\n", err)
	} else {
		fmt.Printf("  ✅ Switched to VISIBLE mode\n")
		fmt.Printf("  🔍 Browser window should be visible for 3 seconds...\n")
		time.Sleep(3 * time.Second)
	}

	// Switch back to headless mode
	fmt.Printf("  Switching back to HEADLESS mode...\n")
	if err := browserMgr.SetVisibility(false); err != nil {
		fmt.Printf("  ❌ Failed to switch to headless mode: %v\n", err)
	} else {
		fmt.Printf("  ✅ Switched back to HEADLESS mode\n")
		time.Sleep(1 * time.Second)
	}

	// Stop browser
	fmt.Printf("  Stopping browser...\n")
	if err := browserMgr.Stop(); err != nil {
		fmt.Printf("  ❌ Failed to stop browser: %v\n", err)
	} else {
		fmt.Printf("  ✅ Browser stopped\n")
	}
}

func checkEnvironment() {
	fmt.Println("  Environment variables:")

	// Check DISPLAY
	display := os.Getenv("DISPLAY")
	if display == "" {
		fmt.Printf("  ❌ DISPLAY not set (required for visible browser on Linux)\n")
	} else {
		fmt.Printf("  ✅ DISPLAY=%s\n", display)
	}

	// Check if running in container
	if _, err := os.Stat("/.dockerenv"); err == nil {
		fmt.Printf("  ⚠️  Running in Docker container (may affect browser visibility)\n")
	}

	// Check if running over SSH
	sshClient := os.Getenv("SSH_CLIENT")
	if sshClient != "" {
		fmt.Printf("  ⚠️  Running over SSH (may affect browser visibility)\n")
	}

	// Check if running in WSL
	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
		fmt.Printf("  ℹ️  Running in WSL\n")

		// Check WSL-specific display setup
		wslDisplay := os.Getenv("WSL_DISTRO_NAME")
		if wslDisplay != "" {
			fmt.Printf("  ℹ️  WSL Distribution: %s\n", wslDisplay)
		}
	}

	// Check for Xvfb
	xvfbRun := os.Getenv("XVFB_RUN")
	if xvfbRun != "" {
		fmt.Printf("  ⚠️  XVFB_RUN is set (virtual display)\n")
	}

	// Check for Chrome/Chromium executable
	paths := []string{
		"/usr/bin/google-chrome",
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
		"/snap/bin/chromium",
	}

	foundBrowser := false
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✅ Found browser: %s\n", path)
			foundBrowser = true
			break
		}
	}

	if !foundBrowser {
		fmt.Printf("  ⚠️  No Chrome/Chromium browser found in standard locations\n")
	}
}
