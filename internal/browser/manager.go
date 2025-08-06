package browser

import (
	"context"
	"fmt"
	"rodmcp/internal/logger"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"go.uber.org/zap"
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

	// Configure launcher
	l := launcher.New().
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
		return fmt.Errorf("failed to launch browser: %w", err)
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
		if err := page.Navigate(url); err != nil {
			m.closePage(pageID)
			return nil, "", fmt.Errorf("failed to navigate to %s: %w", url, err)
		}

		if err := page.WaitLoad(); err != nil {
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

	if err := page.Navigate(url); err != nil {
		return fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	if err := page.WaitLoad(); err != nil {
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
