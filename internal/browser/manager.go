package browser

import (
	"context"
	"fmt"
	"rodmcp/internal/logger"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"go.uber.org/zap"
)

type Manager struct {
	logger   *logger.Logger
	browser  *rod.Browser
	pages    map[string]*rod.Page
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

type Config struct {
	Headless    bool
	Debug       bool
	SlowMotion  time.Duration
	WindowWidth int
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

	// Configure launcher
	l := launcher.New().
		Headless(config.Headless).
		Set("window-size", fmt.Sprintf("%d,%d", config.WindowWidth, config.WindowHeight))

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

	result, err := page.Eval(script)
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	duration := time.Since(start).Milliseconds()
	m.logger.LogBrowserAction("script_executed", pageID, duration)

	return result.Value, nil
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