package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"rodmcp/internal/logger"
	debugpkg "runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"go.uber.org/zap"
)

const (
	// Navigation timeout - how long to wait for page navigation
	NavigationTimeout = 10 * time.Second
	// Connection timeout - how long to wait when checking if a URL is reachable
	ConnectionTimeout = 5 * time.Second
)

type Manager struct {
	logger         *logger.Logger
	browser        *rod.Browser
	pages          map[string]*rod.Page
	mutex          sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	config         Config
	
	// Browser process lifecycle management
	browserPID     int
	controlURL     string
	launcher       *launcher.Launcher
	healthTicker   *time.Ticker
	lastHealthy    time.Time
	restartCount   int
	maxRestarts    int
	lastRestart    time.Time  // Track when last restart occurred
	
	// Connection monitoring
	wsConnections  map[string]bool  // Track WebSocket connections
	connMutex      sync.RWMutex
}

type Config struct {
	Headless     bool
	Debug        bool
	SlowMotion   time.Duration
	WindowWidth  int
	WindowHeight int
}

func NewManager(log *logger.Logger, config Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		logger:        log,
		pages:         make(map[string]*rod.Page),
		ctx:           ctx,
		cancel:        cancel,
		maxRestarts:   3,
		wsConnections: make(map[string]bool),
		lastHealthy:   time.Now(),
	}
}

func (m *Manager) Start(config Config) error {
	m.logger.LogBrowserAction("starting", "", 0)
	start := time.Now()

	// Store config for potential restarts
	m.config = config

	// Find a working browser binary
	browserPath, err := m.findWorkingBrowser()
	if err != nil {
		return fmt.Errorf("no working browser found: %w", err)
	}
	
	m.logger.WithComponent("browser").Info("Using browser binary", zap.String("path", browserPath))

	// Configure launcher
	l := launcher.New().
		Bin(browserPath).
		Headless(config.Headless).
		Set("window-size", fmt.Sprintf("%d,%d", config.WindowWidth, config.WindowHeight))

	// When not headless, ensure the window is visible
	if !config.Headless {
		l = l.Delete("no-startup-window")
	}

	if config.Debug {
		l = l.Devtools(true)
	}

	// Store launcher for process management
	m.launcher = l
	
	// Launch browser with timeout
	launchCtx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()
	
	urlChan := make(chan string, 1)
	errChan := make(chan error, 1)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := debugpkg.Stack()
				m.logger.Error("Browser launch panic", zap.Any("panic", r), zap.String("stack", string(stackTrace)))
				errChan <- fmt.Errorf("browser launch panicked: %v", r)
			}
		}()
		url, err := l.Launch()
		if err != nil {
			errChan <- err
		} else {
			// Store control URL and try to extract PID
			m.controlURL = url
			if pid := m.extractBrowserPID(url); pid > 0 {
				m.browserPID = pid
				m.logger.WithComponent("browser").Info("Browser process started", 
					zap.Int("pid", pid), zap.String("control_url", url))
			}
			urlChan <- url
		}
	}()
	
	var url string
	var launchErr error
	select {
	case url = <-urlChan:
		// Browser launched successfully
	case launchErr = <-errChan:
		// Handle launch error below
	case <-launchCtx.Done():
		return fmt.Errorf("browser launch timed out after 30 seconds - check browser binary and system dependencies")
	}
	
	if launchErr != nil {
		// If browser launch failed and we have a specific binary, try Rod's fallback
		if browserPath != "" {
			m.logger.WithComponent("browser").Warn("System browser failed, trying Rod's browser download", 
				zap.String("failed_path", browserPath), zap.Error(launchErr))
			
			// Try again with Rod's browser download
			l = launcher.New().
				Headless(config.Headless).
				Set("window-size", fmt.Sprintf("%d,%d", config.WindowWidth, config.WindowHeight))
			
			if !config.Headless {
				l = l.Delete("no-startup-window")
			}
			
			if config.Debug {
				l = l.Devtools(true)
			}
			
			// Try fallback launch with timeout
			urlChan2 := make(chan string, 1)
			errChan2 := make(chan error, 1)
			
			go func() {
				defer func() {
					if r := recover(); r != nil {
						stackTrace := debugpkg.Stack()
						m.logger.Error("Fallback browser launch panic", zap.Any("panic", r), zap.String("stack", string(stackTrace)))
						errChan2 <- fmt.Errorf("fallback browser launch panicked: %v", r)
					}
				}()
				url, err := l.Launch()
				if err != nil {
					errChan2 <- err
				} else {
					urlChan2 <- url
				}
			}()
			
			select {
			case url = <-urlChan2:
				// Fallback browser launched successfully
			case launchErr = <-errChan2:
				return fmt.Errorf("failed to launch browser (system: %s failed, Rod download also failed): %w", browserPath, launchErr)
			case <-launchCtx.Done():
				return fmt.Errorf("fallback browser launch timed out after 30 seconds")
			}
			
			m.logger.WithComponent("browser").Info("Successfully using Rod's browser download as fallback")
		} else {
			// Provide more helpful error message for dependency issues
			errStr := launchErr.Error()
			if strings.Contains(errStr, "cannot open shared object file") || strings.Contains(errStr, "not found") {
				return fmt.Errorf("browser launch failed due to missing system dependencies. Please install required libraries or ensure a compatible browser is available: %w", launchErr)
			}
			return fmt.Errorf("failed to launch browser: %w", launchErr)
		}
	}

	// Connect to browser with timeout
	browser := rod.New().ControlURL(url).Context(m.ctx)
	if config.SlowMotion > 0 {
		browser = browser.SlowMotion(config.SlowMotion)
	}

	// Add connection timeout context
	connectCtx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()
	
	if err := browser.Context(connectCtx).Connect(); err != nil {
		if connectCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("browser connection timed out after 30 seconds - check if browser process is responsive")
		}
		return fmt.Errorf("failed to connect to browser: %w", err)
	}
	
	// Small delay to ensure browser is fully initialized
	time.Sleep(100 * time.Millisecond)

	m.mutex.Lock()
	m.browser = browser
	m.lastHealthy = time.Now()
	m.mutex.Unlock()
	
	// Start health monitoring
	m.startHealthMonitoring()
	
	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("started", url, duration)

	return nil
}

func (m *Manager) Stop() error {
	m.logger.LogBrowserAction("stopping", "", 0)
	start := time.Now()

	// Stop health monitoring first
	if m.healthTicker != nil {
		m.healthTicker.Stop()
		m.healthTicker = nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Close all pages safely
	for id, page := range m.pages {
		if page != nil {
			if err := page.Close(); err != nil {
				m.logger.WithComponent("browser").Error("Failed to close page",
					zap.String("page_id", id),
					zap.Error(err))
			}
		}
	}
	m.pages = make(map[string]*rod.Page)

	// Close browser safely with multiple nil checks and panic recovery
	if m.browser != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					stackTrace := debugpkg.Stack()
					m.logger.WithComponent("browser").Error("Recovered from browser close panic",
						zap.Any("panic", r), zap.String("stack", string(stackTrace)))
					// Continue execution - the browser reference will be set to nil below
				}
			}()
			
			// Try to close the browser - any panic will be caught by the defer above
			if err := m.browser.Close(); err != nil {
				m.logger.WithComponent("browser").Error("Failed to close browser",
					zap.Error(err))
			}
		}()
		m.browser = nil // Ensure it's marked as nil after close attempt
	}

	// Cancel context safely
	if m.cancel != nil {
		m.cancel()
	}
	
	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("stopped", "", duration)

	return nil
}

func (m *Manager) NewPage(url string) (*rod.Page, string, error) {
	start := time.Now()

	m.mutex.RLock()
	browser := m.browser
	m.mutex.RUnlock()
	
	if browser == nil {
		return nil, "", fmt.Errorf("browser not started")
	}

	// Test browser health before creating page
	if err := m.testBrowserConnection(browser); err != nil {
		m.logger.WithComponent("browser").Warn("Browser connection unhealthy, attempting restart", zap.Error(err))
		
		// Attempt to restart browser
		if restartErr := m.restartBrowser(); restartErr != nil {
			return nil, "", fmt.Errorf("browser connection unhealthy and restart failed: %w", restartErr)
		}
		
		// Get the new browser reference
		m.mutex.RLock()
		browser = m.browser
		m.mutex.RUnlock()
		
		if browser == nil {
			return nil, "", fmt.Errorf("browser restart succeeded but browser is nil")
		}
	}

	// Use Page() instead of MustPage() to handle connection errors gracefully
	// Add timeout and panic recovery for Page creation
	var page *rod.Page
	var err error
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("page creation panicked: %v", r)
			}
		}()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		page, err = browser.Context(ctx).Page(proto.TargetCreateTarget{})
	}()
	
	if err != nil {
		return nil, "", fmt.Errorf("failed to create new page: %w", err)
	}

	pageID := fmt.Sprintf("page_%d", time.Now().UnixNano())

	m.mutex.Lock()
	m.pages[pageID] = page
	m.mutex.Unlock()

	if url != "" {
		// Check if URL is reachable first
		if err := m.isURLReachable(url); err != nil {
			m.closePage(pageID)
			return nil, "", fmt.Errorf("URL not reachable: %w", err)
		}

		// Navigate with timeout
		ctx, cancel := context.WithTimeout(context.Background(), NavigationTimeout)
		defer cancel()
		
		if err := page.Context(ctx).Navigate(url); err != nil {
			m.closePage(pageID)
			return nil, "", fmt.Errorf("failed to navigate to %s: %w", url, err)
		}

		// Wait for page load with timeout
		if err := page.Context(ctx).WaitLoad(); err != nil {
			m.closePage(pageID)
			return nil, "", fmt.Errorf("failed to wait for page load: %w", err)
		}
	}

	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("page_created", url, duration)

	return page, pageID, nil
}

func (m *Manager) GetPage(pageID string) (*rod.Page, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	page, exists := m.pages[pageID]
	if !exists {
		return nil, fmt.Errorf("page not found: %s", pageID)
	}

	return page, nil
}

func (m *Manager) ClosePage(pageID string) error {
	return m.closePage(pageID)
}

func (m *Manager) closePage(pageID string) error {
	start := time.Now()

	m.mutex.Lock()
	page, exists := m.pages[pageID]
	if exists {
		delete(m.pages, pageID)
	}
	m.mutex.Unlock()

	if !exists {
		return fmt.Errorf("page not found: %s", pageID)
	}

	if err := page.Close(); err != nil {
		return fmt.Errorf("failed to close page: %w", err)
	}

	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("page_closed", pageID, duration)

	return nil
}

func (m *Manager) ListPages() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var pageIDs []string
	for id := range m.pages {
		pageIDs = append(pageIDs, id)
	}

	return pageIDs
}

func (m *Manager) Screenshot(pageID string) ([]byte, error) {
	start := time.Now()

	page, err := m.GetPage(pageID)
	if err != nil {
		return nil, err
	}

	screenshot, err := page.Screenshot(true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("screenshot", pageID, duration)

	return screenshot, nil
}

func (m *Manager) ExecuteScript(pageID string, script string) (interface{}, error) {
	start := time.Now()

	page, err := m.GetPage(pageID)
	if err != nil {
		return nil, err
	}

	// Clean up the script
	script = strings.TrimSpace(script)
	
	// go-rod's page.Eval expects JavaScript wrapped as arrow functions
	// Key insight: page.Eval works with "() => expression" or "() => { statements; return value; }"
	
	lines := strings.Split(script, "\n")
	hasObjectLiteral := false
	
	// Check if script contains object literal expressions that should be returned
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "({") {
			hasObjectLiteral = true
			break
		}
	}
	
	var wrappedScript string
	
	if hasObjectLiteral {
		// Script has object literal - wrap in arrow function with return
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "({") {
				lines[i] = strings.Replace(line, "({", "return ({", 1)
				break
			}
		}
		wrappedScript = fmt.Sprintf("() => {\n%s\n}", strings.Join(lines, "\n"))
	} else {
		// No object literal - check if it's a simple expression or needs statement wrapper
		if len(lines) == 1 && !strings.Contains(script, "=") && !strings.Contains(script, ";") {
			// Single expression, wrap as arrow function expression
			wrappedScript = fmt.Sprintf("() => %s", script)
		} else {
			// Multiple statements, wrap in arrow function block
			wrappedScript = fmt.Sprintf("() => {\n%s\n}", script)
		}
	}

	// Execute the script using page.Eval
	result, err := page.Eval(wrappedScript)
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("script_executed", pageID, duration)

	return result.Value, nil
}

func (m *Manager) NavigateExistingPage(pageID string, url string) error {
	start := time.Now()

	page, err := m.GetPage(pageID)
	if err != nil {
		return err
	}

	// Check if URL is reachable first
	if err := m.isURLReachable(url); err != nil {
		return fmt.Errorf("URL not reachable: %w", err)
	}

	// Navigate with timeout
	ctx, cancel := context.WithTimeout(context.Background(), NavigationTimeout)
	defer cancel()

	if err := page.Context(ctx).Navigate(url); err != nil {
		return fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	// Wait for page load with timeout
	if err := page.Context(ctx).WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("page_navigated", url, duration)

	return nil
}

func (m *Manager) GetPageInfo(pageID string) (map[string]interface{}, error) {
	page, err := m.GetPage(pageID)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"id": pageID,
	}

	// Safely get page info without panic
	if pageInfo, err := page.Info(); err == nil && pageInfo != nil {
		info["url"] = pageInfo.URL
	} else {
		info["url"] = ""
	}

	title, err := page.Element("title")
	if err == nil {
		if titleText, err := title.Text(); err == nil {
			info["title"] = titleText
		}
	}

	return info, nil
}

func (m *Manager) SetVisibility(visible bool) error {
	m.logger.LogBrowserAction("set_visibility", "", 0)
	start := time.Now()

	m.mutex.RLock()
	browser := m.browser
	m.mutex.RUnlock()
	
	if browser == nil {
		return fmt.Errorf("browser not started")
	}

	// Check if visibility is already as requested
	if m.config.Headless == !visible {
		mode := "headless"
		if visible {
			mode = "visible"
		}
		duration := time.Since(start).Milliseconds()
		m.logger.LogBrowserAction("visibility_already_set", mode, duration)
		return nil
	}

	// Store current page URLs to restore after restart
	pageURLs := make(map[string]string)
	m.mutex.RLock()
	for id, page := range m.pages {
		// Safely get page info without panic
		if pageInfo, err := page.Info(); err == nil && pageInfo != nil {
			pageURLs[id] = pageInfo.URL
		}
	}
	m.mutex.RUnlock()

	// Update config
	m.config.Headless = !visible

	// Stop current browser
	if err := m.Stop(); err != nil {
		return fmt.Errorf("failed to stop browser for visibility change: %w", err)
	}

	// Create new context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Start browser with new visibility setting
	if err := m.Start(m.config); err != nil {
		return fmt.Errorf("failed to restart browser with new visibility: %w", err)
	}

	// Restore pages
	for oldID, url := range pageURLs {
		if url != "" {
			_, newID, err := m.NewPage(url)
			if err != nil {
				m.logger.WithComponent("browser").Warn("Failed to restore page after visibility change",
					zap.String("old_page_id", oldID),
					zap.String("url", url),
					zap.Error(err))
			} else {
				m.logger.WithComponent("browser").Info("Restored page after visibility change",
					zap.String("old_page_id", oldID),
					zap.String("new_page_id", newID),
					zap.String("url", url))
			}
		}
	}

	mode := "headless"
	if visible {
		mode = "visible"
	}

	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("visibility_changed", mode, duration)

	m.logger.WithComponent("browser").Info("Browser visibility changed successfully",
		zap.String("mode", mode),
		zap.Int("pages_restored", len(pageURLs)))

	return nil
}

// findWorkingBrowser attempts to find a working browser binary with proper fallbacks
func (m *Manager) findWorkingBrowser() (string, error) {
	// Check for environment variable override first
	if envBrowser := os.Getenv("RODMCP_BROWSER_PATH"); envBrowser != "" {
		if m.isBrowserWorking(envBrowser) {
			m.logger.WithComponent("browser").Info("Using browser from environment variable", 
				zap.String("path", envBrowser))
			return envBrowser, nil
		} else {
			m.logger.WithComponent("browser").Warn("Environment browser path not working, falling back to defaults", 
				zap.String("path", envBrowser))
		}
	}

	// List of browser binaries to try in order of preference
	candidates := []string{
		// User-specified or system browsers
		"/home/darrell/.nix-profile/bin/chromium-browser",
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/snap/bin/chromium",
		// Let Rod download its own if needed (last resort)
		"",
	}
	
	for _, candidate := range candidates {
		if candidate == "" {
			// Empty string means let Rod handle browser download
			m.logger.WithComponent("browser").Info("Using Rod's browser download as fallback")
			return candidate, nil
		}
		
		if m.isBrowserWorking(candidate) {
			return candidate, nil
		}
	}
	
	return "", fmt.Errorf("no working browser binary found after checking all candidates")
}

// isBrowserWorking checks if a browser binary exists and has required dependencies
func (m *Manager) isBrowserWorking(browserPath string) bool {
	// Check if file exists
	if _, err := os.Stat(browserPath); err != nil {
		m.logger.WithComponent("browser").Debug("Browser binary not found", 
			zap.String("path", browserPath), zap.Error(err))
		return false
	}
	
	// Try to run browser with --version to check if dependencies are available (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, browserPath, "--version")
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			m.logger.WithComponent("browser").Debug("Browser binary version check timed out", 
				zap.String("path", browserPath))
		} else {
			m.logger.WithComponent("browser").Debug("Browser binary failed version check", 
				zap.String("path", browserPath), zap.Error(err))
		}
		return false
	}
	
	m.logger.WithComponent("browser").Debug("Browser binary is working", zap.String("path", browserPath))
	return true
}

// isURLReachable checks if a URL is reachable before attempting navigation
func (m *Manager) isURLReachable(targetURL string) error {
	// Skip check for file:// URLs
	if strings.HasPrefix(targetURL, "file://") {
		return nil
	}
	
	// Parse the URL to ensure it's valid
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	// For http/https URLs, do a quick connectivity check
	if parsedURL.Scheme == "http" || parsedURL.Scheme == "https" {
		client := &http.Client{
			Timeout: ConnectionTimeout,
		}
		
		// Use HEAD request for faster check
		ctx, cancel := context.WithTimeout(context.Background(), ConnectionTimeout)
		defer cancel()
		
		req, err := http.NewRequestWithContext(ctx, "HEAD", targetURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("URL not reachable: %w", err)
		}
		resp.Body.Close()
		
		// Accept any status code - even errors like 404 mean the server is reachable
		m.logger.WithComponent("browser").Debug("URL reachability check",
			zap.String("url", targetURL),
			zap.Int("status", resp.StatusCode))
	}
	
	return nil
}

// testBrowserConnection quickly tests if browser connection is healthy
func (m *Manager) testBrowserConnection(browser *rod.Browser) error {
	if browser == nil {
		return fmt.Errorf("browser is nil")
	}
	
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				m.logger.WithComponent("browser").Debug("Test connection panic details",
					zap.Any("panic", r),
					zap.String("stack", string(debugpkg.Stack())))
				err = fmt.Errorf("browser connection test panicked: %v", r)
			}
		}()
		
		// Try to get browser version as a quick health check
		// Use the browser directly without creating a new context
		version, versionErr := browser.Version()
		if versionErr != nil {
			err = fmt.Errorf("failed to get browser version: %w", versionErr)
			return
		}
		m.logger.WithComponent("browser").Debug("Browser version retrieved successfully",
			zap.Any("version", version))
	}()
	
	return err
}

// restartBrowser safely restarts the browser with improved error handling
func (m *Manager) restartBrowser() error {
	m.mutex.Lock()
	// Check restart count to prevent infinite loops
	if m.restartCount >= m.maxRestarts {
		m.mutex.Unlock()
		return fmt.Errorf("browser restart limit exceeded (%d/%d)", m.restartCount, m.maxRestarts)
	}
	m.restartCount++
	currentRestartCount := m.restartCount
	m.lastRestart = time.Now()
	m.mutex.Unlock()
	
	m.logger.WithComponent("browser").Info("Attempting to restart browser",
		zap.Int("restart_attempt", currentRestartCount),
		zap.Int("max_restarts", m.maxRestarts))
	
	// Stop browser with extra safety (ignore panics)
	func() {
		defer func() {
			if r := recover(); r != nil {
				m.logger.WithComponent("browser").Warn("Panic during browser stop, continuing", zap.Any("panic", r))
			}
		}()
		m.Stop()
	}()
	
	// Wait a bit before restarting to avoid rapid restart loops
	time.Sleep(2 * time.Second)
	
	// Create new context
	m.ctx, m.cancel = context.WithCancel(context.Background())
	
	// Start browser
	if err := m.Start(m.config); err != nil {
		return fmt.Errorf("failed to restart browser: %w", err)
	}
	
	m.logger.WithComponent("browser").Info("Browser restarted successfully",
		zap.Int("restart_count", currentRestartCount))
	return nil
}

// CheckHealth verifies the browser connection is still active
func (m *Manager) CheckHealth() error {
	m.mutex.RLock()
	browser := m.browser
	m.mutex.RUnlock()
	
	if browser == nil {
		// This is normal - browser may not be started yet or may have been stopped
		// Don't treat this as an error that needs logging
		return fmt.Errorf("browser not started")
	}

	// Try to get browser version as a simple health check with panic recovery
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Panics during health check are expected when browser is shutting down
				// Log at debug level to avoid noise
				m.logger.WithComponent("browser").Debug("Browser health check recovered from panic",
					zap.Any("panic", r))
				err = fmt.Errorf("browser health check panicked: %v", r)
			}
		}()
		
		// Re-check browser under lock to ensure it's still valid
		m.mutex.RLock()
		currentBrowser := m.browser
		m.mutex.RUnlock()
		
		if currentBrowser == nil {
			err = fmt.Errorf("browser stopped during health check")
			return
		}
		
		// Use Version() instead of Pages() as it's simpler and less likely to panic
		_, err = currentBrowser.Context(ctx).Version()
	}()
	
	if err != nil {
		// Only log at debug level - health check failures are handled by the circuit breaker
		m.logger.WithComponent("browser").Debug("Browser health check failed",
			zap.Error(err))
		return fmt.Errorf("browser connection unhealthy: %w", err)
	}

	return nil
}

// EnsureHealthy checks browser health and restarts if needed
func (m *Manager) EnsureHealthy() error {
	// First, check if we should reset the restart counter
	// Reset if it's been more than 5 minutes since last restart
	m.mutex.Lock()
	if m.restartCount > 0 && time.Since(m.lastRestart) > 5*time.Minute {
		m.logger.WithComponent("browser").Debug("Resetting restart counter after stable operation",
			zap.Int("previous_count", m.restartCount))
		m.restartCount = 0
	}
	m.mutex.Unlock()
	
	if err := m.CheckHealth(); err != nil {
		m.logger.WithComponent("browser").Info("Browser unhealthy, attempting automatic restart",
			zap.Error(err))
		
		// Attempt to restart the browser
		if restartErr := m.restartBrowser(); restartErr != nil {
			m.logger.WithComponent("browser").Error("Failed to restart browser automatically",
				zap.Error(restartErr))
			// Return the original error combined with restart error
			return fmt.Errorf("browser unhealthy and restart failed: %v (restart error: %w)", err, restartErr)
		}
		
		// Wait a moment for browser to stabilize after restart
		time.Sleep(1 * time.Second)
		
		// Browser restarted successfully, verify it's now healthy with retries
		var verifyErr error
		for i := 0; i < 3; i++ {
			verifyErr = m.CheckHealth()
			if verifyErr == nil {
				m.logger.WithComponent("browser").Info("Browser restarted successfully and is now healthy",
					zap.Int("verification_attempt", i+1))
				return nil
			}
			if i < 2 {
				m.logger.WithComponent("browser").Debug("Browser health check failed after restart, retrying",
					zap.Int("attempt", i+1),
					zap.Error(verifyErr))
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		}
		
		m.logger.WithComponent("browser").Error("Browser still unhealthy after restart and retries",
			zap.Error(verifyErr))
		return fmt.Errorf("browser still unhealthy after restart: %w", verifyErr)
	}

	return nil
}

// PageInfo represents information about a browser page/tab
type PageInfo struct {
	PageID string `json:"page_id"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

// GetAllPages returns information about all open pages/tabs
func (m *Manager) GetAllPages() []PageInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var pages []PageInfo
	for pageID, page := range m.pages {
		title := ""
		url := ""
		
		// Try to get page info, but don't fail if it's not available
		if info, err := page.Info(); err == nil {
			title = info.Title
			url = info.URL
		}
		
		// Fallback to basic URL if available
		if url == "" {
			if pageInfo, err := page.Info(); err == nil && pageInfo != nil {
				if pageInfo.URL != "" {
					url = pageInfo.URL
				}
			}
		}
		
		pages = append(pages, PageInfo{
			PageID: pageID,
			Title:  title,
			URL:    url,
		})
	}

	return pages
}

// GetCurrentPageID returns the ID of the currently active page
func (m *Manager) GetCurrentPageID() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// For now, return the first page ID as current
	// This is a simplification - in a real implementation we'd track the active page
	for pageID := range m.pages {
		return pageID
	}

	return ""
}

// SwitchToPage switches to the specified page/tab
func (m *Manager) SwitchToPage(pageID string) error {
	m.mutex.RLock()
	page, exists := m.pages[pageID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("page %s not found", pageID)
	}

	// Activate the page (bring it to front)
	_, err := page.Activate()
	if err != nil {
		return fmt.Errorf("failed to activate page %s: %w", pageID, err)
	}

	m.logger.LogBrowserAction("page_switched", pageID, 0)
	return nil
}

// extractBrowserPID attempts to extract the browser PID from control URL
func (m *Manager) extractBrowserPID(controlURL string) int {
	// Try to find browser process by looking at running processes
	// This is a heuristic approach since Rod doesn't expose the PID directly
	cmd := exec.Command("pgrep", "-f", "chrome.*--remote-debugging-port")
	output, err := cmd.Output()
	if err != nil {
		m.logger.WithComponent("browser").Debug("Could not find browser PID", zap.Error(err))
		return 0
	}
	
	if len(output) > 0 {
		pidStr := strings.TrimSpace(string(output))
		lines := strings.Split(pidStr, "\n")
		if len(lines) > 0 {
			if pid, err := strconv.Atoi(lines[0]); err == nil {
				return pid
			}
		}
	}
	
	return 0
}

// startHealthMonitoring starts periodic health checks of the browser process
func (m *Manager) startHealthMonitoring() {
	if m.healthTicker != nil {
		m.healthTicker.Stop()
	}
	
	m.healthTicker = time.NewTicker(10 * time.Second)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				m.logger.WithComponent("browser").Error("Health monitoring panic", 
					zap.Any("panic", r))
			}
		}()
		
		for {
			select {
			case <-m.ctx.Done():
				return
			case <-m.healthTicker.C:
				m.performHealthCheck()
			}
		}
	}()
}

// performHealthCheck checks if the browser process is still alive and responsive
func (m *Manager) performHealthCheck() {
	m.mutex.RLock()
	browser := m.browser
	pid := m.browserPID
	m.mutex.RUnlock()
	
	if browser == nil {
		return
	}
	
	// Check if browser process is still running
	if pid > 0 && !m.isProcessRunning(pid) {
		m.logger.WithComponent("browser").Warn("Browser process died", 
			zap.Int("pid", pid))
		m.handleBrowserDeath()
		return
	}
	
	// Check browser responsiveness
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	err := func() error {
		defer func() {
			if r := recover(); r != nil {
				m.logger.WithComponent("browser").Debug("Health check panic", 
					zap.Any("panic", r))
			}
		}()
		
		_, err := browser.Context(ctx).Version()
		return err
	}()
	
	if err != nil {
		m.logger.WithComponent("browser").Warn("Browser health check failed", 
			zap.Error(err))
		// Don't immediately restart - wait for multiple failures
		m.mutex.Lock()
		timeSinceHealthy := time.Since(m.lastHealthy)
		m.mutex.Unlock()
		
		if timeSinceHealthy > 30*time.Second {
			m.logger.WithComponent("browser").Warn("Browser unresponsive for too long, marking for restart")
			m.handleBrowserDeath()
		}
	} else {
		m.mutex.Lock()
		m.lastHealthy = time.Now()
		m.mutex.Unlock()
	}
}

// isProcessRunning checks if a process with the given PID is still running
func (m *Manager) isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	
	// Send signal 0 to check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// handleBrowserDeath handles when the browser process dies unexpectedly
func (m *Manager) handleBrowserDeath() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.restartCount >= m.maxRestarts {
		m.logger.WithComponent("browser").Error("Browser restart limit exceeded", 
			zap.Int("restart_count", m.restartCount),
			zap.Int("max_restarts", m.maxRestarts))
		return
	}
	
	m.logger.WithComponent("browser").Info("Attempting automatic browser restart", 
		zap.Int("restart_attempt", m.restartCount+1))
	
	// Stop health monitoring during restart
	if m.healthTicker != nil {
		m.healthTicker.Stop()
		m.healthTicker = nil
	}
	
	// Clean up current browser
	m.browser = nil
	m.browserPID = 0
	
	// Clear pages
	for id := range m.pages {
		delete(m.pages, id)
	}
	
	// Increment restart count
	m.restartCount++
	
	// Restart browser in background
	go func() {
		defer func() {
			if r := recover(); r != nil {
				m.logger.WithComponent("browser").Error("Browser restart panic", 
					zap.Any("panic", r))
			}
		}()
		
		// Wait a bit before restart
		time.Sleep(2 * time.Second)
		
		if err := m.Start(m.config); err != nil {
			m.logger.WithComponent("browser").Error("Failed to restart browser", 
				zap.Error(err))
		} else {
			m.logger.WithComponent("browser").Info("Browser restarted successfully")
		}
	}()
}
