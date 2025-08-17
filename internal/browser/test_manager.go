package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"go.uber.org/zap"
	"rodmcp/internal/logger"
)

// TestManager is a browser manager optimized for testing scenarios
// It handles timeouts and cleanup more gracefully to avoid test flakiness
type TestManager struct {
	*Manager
	testConfig TestConfig
}

type TestConfig struct {
	// Faster timeouts for testing
	StartupTimeout  time.Duration
	OperationTimeout time.Duration
	ShutdownTimeout time.Duration
	// More lenient error handling
	IgnoreShutdownErrors bool
	ReusePages          bool
}

// DefaultTestConfig returns sensible defaults for testing
func DefaultTestConfig() TestConfig {
	return TestConfig{
		StartupTimeout:       15 * time.Second,
		OperationTimeout:     10 * time.Second,
		ShutdownTimeout:      5 * time.Second,
		IgnoreShutdownErrors: true,
		ReusePages:          true,
	}
}

// NewTestManager creates a browser manager optimized for testing
func NewTestManager(log *logger.Logger, config Config, testConfig TestConfig) *TestManager {
	mgr := NewManager(log, config)
	return &TestManager{
		Manager:    mgr,
		testConfig: testConfig,
	}
}

// StartWithTimeout starts the browser with a test-appropriate timeout
func (tm *TestManager) StartWithTimeout() error {
	ctx, cancel := context.WithTimeout(context.Background(), tm.testConfig.StartupTimeout)
	defer cancel()
	
	startChan := make(chan error, 1)
	go func() {
		startChan <- tm.Manager.Start(tm.config)
	}()
	
	select {
	case err := <-startChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("browser startup timed out after %v", tm.testConfig.StartupTimeout)
	}
}

// StopGracefully stops the browser with test-appropriate timeout and error handling
func (tm *TestManager) StopGracefully() error {
	if tm.testConfig.IgnoreShutdownErrors {
		// Create a separate context for shutdown to avoid cancellation issues
		ctx, cancel := context.WithTimeout(context.Background(), tm.testConfig.ShutdownTimeout)
		defer cancel()
		
		stopChan := make(chan error, 1)
		go func() {
			stopChan <- tm.Manager.Stop()
		}()
		
		select {
		case err := <-stopChan:
			// Log but don't fail tests on shutdown errors
			if err != nil {
				tm.logger.WithComponent("browser").Debug("Browser shutdown warning (expected in tests)", 
					zap.Error(err))
			}
			return nil
		case <-ctx.Done():
			tm.logger.WithComponent("browser").Debug("Browser shutdown timed out (expected in tests)")
			return nil
		}
	}
	
	return tm.Manager.Stop()
}

// NewPageWithValidation creates a new page with proper validation and timeout
func (tm *TestManager) NewPageWithValidation(url string) (*rod.Page, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), tm.testConfig.OperationTimeout)
	defer cancel()
	
	pageChan := make(chan struct {
		page   *rod.Page
		pageID string
		err    error
	}, 1)
	
	go func() {
		page, pageID, err := tm.Manager.NewPage(url)
		pageChan <- struct {
			page   *rod.Page
			pageID string
			err    error
		}{page, pageID, err}
	}()
	
	select {
	case result := <-pageChan:
		if result.err != nil {
			return nil, "", result.err
		}
		
		// Give page time to load
		time.Sleep(500 * time.Millisecond)
		return result.page, result.pageID, nil
		
	case <-ctx.Done():
		return nil, "", fmt.Errorf("page creation timed out after %v", tm.testConfig.OperationTimeout)
	}
}

// ExecuteOperationWithTimeout executes any browser operation with timeout
func (tm *TestManager) ExecuteOperationWithTimeout(operation func() error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	errChan := make(chan error, 1)
	go func() {
		errChan <- operation()
	}()
	
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}

// WaitForPageLoad waits for a page to be fully loaded and ready
func (tm *TestManager) WaitForPageLoad(page *rod.Page, timeout time.Duration) error {
	return tm.ExecuteOperationWithTimeout(func() error {
		return page.WaitLoad()
	}, timeout)
}

// GetPagesWithRetry gets all pages with retry logic for flaky connections
func (tm *TestManager) GetPagesWithRetry(maxAttempts int) []PageInfo {
	var pages []PageInfo
	var lastErr error
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					lastErr = fmt.Errorf("panic during GetAllPages: %v", r)
				}
			}()
			
			pages = tm.Manager.GetAllPages()
			lastErr = nil
		}()
		
		if lastErr == nil {
			return pages
		}
		
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}
	}
	
	// Return empty slice if all attempts failed
	tm.logger.WithComponent("browser").Warn("Failed to get pages after retries", 
		zap.Error(lastErr))
	return []PageInfo{}
}