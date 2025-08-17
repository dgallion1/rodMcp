package mcp

import (
	"context"
	"fmt"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"time"
)

// TestBrowserManager provides a real browser manager for testing
// but with simplified lifecycle management
type TestBrowserManager struct {
	manager *browser.Manager
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewTestBrowserManager(log *logger.Logger) *TestBrowserManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	config := browser.Config{
		Headless:     true,  // Always headless for tests
		Debug:        false,
		SlowMotion:   0,
		WindowWidth:  1280,
		WindowHeight: 720,
	}
	
	manager := browser.NewManager(log, config)
	
	return &TestBrowserManager{
		manager: manager,
		logger:  log,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (m *TestBrowserManager) Start() error {
	config := browser.Config{
		Headless:     true,
		Debug:        false,
		SlowMotion:   0,
		WindowWidth:  1280,
		WindowHeight: 720,
	}
	
	return m.manager.Start(config)
}

func (m *TestBrowserManager) Stop() error {
	m.cancel()
	return m.manager.Stop()
}

func (m *TestBrowserManager) CheckHealth() error {
	if m.manager == nil {
		return fmt.Errorf("browser manager not initialized")
	}
	return m.manager.CheckHealth()
}

func (m *TestBrowserManager) EnsureHealthy() error {
	if m.manager == nil {
		return fmt.Errorf("browser manager not initialized")  
	}
	return m.manager.EnsureHealthy()
}

// GetManager returns the underlying browser manager for advanced operations
func (m *TestBrowserManager) GetManager() *browser.Manager {
	return m.manager
}

// Wait for browser to be ready (useful in tests)
func (m *TestBrowserManager) WaitReady(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(m.ctx, timeout)
	defer cancel()
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for browser to be ready")
		case <-ticker.C:
			if err := m.CheckHealth(); err == nil {
				return nil
			}
		}
	}
}