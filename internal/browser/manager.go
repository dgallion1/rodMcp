package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"rodmcp/internal/logger"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"go.uber.org/zap"
)

const (
	// Navigation timeout - how long to wait for page navigation
	NavigationTimeout = 10 * time.Second
	// Connection timeout - how long to wait when checking if a URL is reachable
	ConnectionTimeout = 5 * time.Second
)

type Manager struct {
	logger  *logger.Logger
	browser *rod.Browser
	pages   map[string]*rod.Page
	mutex   sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	config  Config
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
		logger: log,
		pages:  make(map[string]*rod.Page),
		ctx:    ctx,
		cancel: cancel,
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

	// Launch browser
	url, err := l.Launch()
	if err != nil {
		// If browser launch failed and we have a specific binary, try Rod's fallback
		if browserPath != "" {
			m.logger.WithComponent("browser").Warn("System browser failed, trying Rod's browser download", 
				zap.String("failed_path", browserPath), zap.Error(err))
			
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
			
			url, err = l.Launch()
			if err != nil {
				return fmt.Errorf("failed to launch browser (system: %s failed, Rod download also failed): %w", browserPath, err)
			}
			
			m.logger.WithComponent("browser").Info("Successfully using Rod's browser download as fallback")
		} else {
			// Provide more helpful error message for dependency issues
			errStr := err.Error()
			if strings.Contains(errStr, "cannot open shared object file") || strings.Contains(errStr, "not found") {
				return fmt.Errorf("browser launch failed due to missing system dependencies. Please install required libraries or ensure a compatible browser is available: %w", err)
			}
			return fmt.Errorf("failed to launch browser: %w", err)
		}
	}

	// Connect to browser
	browser := rod.New().ControlURL(url).Context(m.ctx)
	if config.SlowMotion > 0 {
		browser = browser.SlowMotion(config.SlowMotion)
	}

	if err := browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}

	m.browser = browser
	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("started", url, duration)

	return nil
}

func (m *Manager) Stop() error {
	m.logger.LogBrowserAction("stopping", "", 0)
	start := time.Now()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Close all pages
	for id, page := range m.pages {
		if err := page.Close(); err != nil {
			m.logger.WithComponent("browser").Error("Failed to close page",
				zap.String("page_id", id),
				zap.Error(err))
		}
	}
	m.pages = make(map[string]*rod.Page)

	// Close browser
	if m.browser != nil {
		if err := m.browser.Close(); err != nil {
			m.logger.WithComponent("browser").Error("Failed to close browser",
				zap.Error(err))
		}
	}

	m.cancel()
	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("stopped", "", duration)

	return nil
}

func (m *Manager) NewPage(url string) (*rod.Page, string, error) {
	start := time.Now()

	if m.browser == nil {
		return nil, "", fmt.Errorf("browser not started")
	}

	page := m.browser.MustPage()

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
		"id":  pageID,
		"url": page.MustInfo().URL,
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

	if m.browser == nil {
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
		if pageInfo := page.MustInfo(); pageInfo != nil {
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
	
	// Try to run browser with --version to check if dependencies are available
	cmd := exec.Command(browserPath, "--version")
	if err := cmd.Run(); err != nil {
		m.logger.WithComponent("browser").Debug("Browser binary failed version check", 
			zap.String("path", browserPath), zap.Error(err))
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

// CheckHealth verifies the browser connection is still active
func (m *Manager) CheckHealth() error {
	if m.browser == nil {
		return fmt.Errorf("browser not started")
	}

	// Try to get browser version as a simple health check
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	// Use a simple evaluation to check if browser is responsive
	_, err := m.browser.Context(ctx).Pages()
	if err != nil {
		m.logger.WithComponent("browser").Warn("Browser health check failed",
			zap.Error(err))
		return fmt.Errorf("browser connection unhealthy: %w", err)
	}

	return nil
}

// EnsureHealthy checks browser health and restarts if needed
func (m *Manager) EnsureHealthy() error {
	if err := m.CheckHealth(); err != nil {
		m.logger.WithComponent("browser").Info("Browser unhealthy, attempting restart",
			zap.Error(err))
		
		// Store current page URLs before restart
		pageURLs := make(map[string]string)
		m.mutex.RLock()
		for id, page := range m.pages {
			if pageInfo := page.MustInfo(); pageInfo != nil {
				pageURLs[id] = pageInfo.URL
			}
		}
		m.mutex.RUnlock()

		// Stop and restart browser
		if stopErr := m.Stop(); stopErr != nil {
			m.logger.WithComponent("browser").Error("Failed to stop unhealthy browser",
				zap.Error(stopErr))
		}

		// Create new context
		m.ctx, m.cancel = context.WithCancel(context.Background())

		if startErr := m.Start(m.config); startErr != nil {
			return fmt.Errorf("failed to restart browser: %w", startErr)
		}

		// Restore pages
		for oldID, url := range pageURLs {
			if url != "" {
				if _, _, restoreErr := m.NewPage(url); restoreErr != nil {
					m.logger.WithComponent("browser").Warn("Failed to restore page after restart",
						zap.String("page_id", oldID),
						zap.String("url", url),
						zap.Error(restoreErr))
				}
			}
		}

		m.logger.WithComponent("browser").Info("Browser restarted successfully")
	}

	return nil
}
