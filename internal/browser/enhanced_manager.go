package browser

import (
	"context"
	"fmt"
	"rodmcp/internal/logger"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"go.uber.org/zap"
)

// EnhancedManager extends the base Manager with additional stability features
type EnhancedManager struct {
	*Manager
	
	// Retry configuration
	maxRetries       int
	retryDelay       time.Duration
	
	// Page state tracking
	pageStates       map[string]*PageState
	pageStatesMutex  sync.RWMutex
	
	// Recovery tracking
	recoveryAttempts map[string]int
	recoveryMutex    sync.RWMutex
	
	// Browser restart tracking with backoff
	browserRestartAttempts int
	browserRestartMutex    sync.RWMutex
	lastBrowserRestart     time.Time
	restartBackoffBase     time.Duration
	restartBackoffMax      time.Duration
}

// PageState tracks the state of a browser page for recovery
type PageState struct {
	PageID        string
	URL           string
	Title         string
	LastActive    time.Time
	IsHealthy     bool
	RecoveryCount int
	Context       context.Context
	Cancel        context.CancelFunc
}

// NewEnhancedManager creates a new enhanced browser manager
func NewEnhancedManager(log *logger.Logger, config Config) *EnhancedManager {
	base := NewManager(log, config)
	
	return &EnhancedManager{
		Manager:            base,
		maxRetries:         3,
		retryDelay:         1 * time.Second,
		pageStates:         make(map[string]*PageState),
		recoveryAttempts:   make(map[string]int),
		restartBackoffBase: 1 * time.Second,
		restartBackoffMax:  30 * time.Second,
	}
}

// NewPageWithRetry creates a new page with automatic retry on failure
func (em *EnhancedManager) NewPageWithRetry(url string) (*rod.Page, string, error) {
	var page *rod.Page
	var pageID string
	var lastErr error
	
	for attempt := 0; attempt <= em.maxRetries; attempt++ {
		if attempt > 0 {
			em.logger.WithComponent("browser").Info("Retrying page creation",
				zap.Int("attempt", attempt),
				zap.String("url", url))
			time.Sleep(em.retryDelay * time.Duration(attempt))
		}
		
		// Ensure browser is healthy before creating page
		if err := em.EnsureHealthy(); err != nil {
			lastErr = fmt.Errorf("browser unhealthy: %w", err)
			// Check for context errors and handle appropriately
			if em.isContextError(err) {
				lastErr = em.handleContextError(err, "page_creation")
			}
			continue
		}
		
		page, pageID, lastErr = em.NewPage(url)
		if lastErr == nil {
			// Track page state for recovery
			em.trackPageState(pageID, url, page)
			return page, pageID, nil
		}
		
		// Check for context errors first
		if em.isContextError(lastErr) {
			lastErr = em.handleContextError(lastErr, "page_creation")
			continue // Try again after restart
		}
		
		// Check if error is recoverable
		if !em.isRecoverableError(lastErr) {
			return nil, "", lastErr
		}
		
		em.logger.WithComponent("browser").Warn("Page creation failed, will retry",
			zap.Error(lastErr),
			zap.Int("attempt", attempt))
	}
	
	return nil, "", fmt.Errorf("failed after %d retries: %w", em.maxRetries, lastErr)
}

// NavigateWithRetry navigates to a URL with automatic retry
func (em *EnhancedManager) NavigateWithRetry(pageID string, url string) error {
	var lastErr error
	
	for attempt := 0; attempt <= em.maxRetries; attempt++ {
		if attempt > 0 {
			em.logger.WithComponent("browser").Info("Retrying navigation",
				zap.Int("attempt", attempt),
				zap.String("url", url))
			time.Sleep(em.retryDelay * time.Duration(attempt))
		}
		
		// Try to recover page if unhealthy
		if err := em.ensurePageHealthy(pageID); err != nil {
			lastErr = err
			// Check for context errors
			if em.isContextError(err) {
				lastErr = em.handleContextError(err, "page_health_check")
			}
			continue
		}
		
		lastErr = em.NavigateExistingPage(pageID, url)
		if lastErr == nil {
			// Update page state
			em.updatePageState(pageID, url)
			return nil
		}
		
		// Check for context errors first
		if em.isContextError(lastErr) {
			lastErr = em.handleContextError(lastErr, "navigation")
			continue // Try again after restart
		}
		
		// Check if error is recoverable
		if !em.isRecoverableError(lastErr) {
			return lastErr
		}
		
		em.logger.WithComponent("browser").Warn("Navigation failed, will retry",
			zap.Error(lastErr),
			zap.Int("attempt", attempt))
	}
	
	return fmt.Errorf("navigation failed after %d retries: %w", em.maxRetries, lastErr)
}

// ScreenshotWithRetry takes a screenshot with automatic retry
func (em *EnhancedManager) ScreenshotWithRetry(pageID string) ([]byte, error) {
	var lastErr error
	var screenshot []byte
	
	for attempt := 0; attempt <= em.maxRetries; attempt++ {
		if attempt > 0 {
			em.logger.WithComponent("browser").Info("Retrying screenshot",
				zap.Int("attempt", attempt),
				zap.String("page_id", pageID))
			time.Sleep(em.retryDelay * time.Duration(attempt))
		}
		
		// Ensure page is healthy
		if err := em.ensurePageHealthy(pageID); err != nil {
			lastErr = err
			// Check for context errors
			if em.isContextError(err) {
				lastErr = em.handleContextError(err, "screenshot_health_check")
			}
			continue
		}
		
		screenshot, lastErr = em.Screenshot(pageID)
		if lastErr == nil {
			return screenshot, nil
		}
		
		// Check for context errors first
		if em.isContextError(lastErr) {
			lastErr = em.handleContextError(lastErr, "screenshot")
			continue // Try again after restart
		}
		
		// Check if error is recoverable
		if !em.isRecoverableError(lastErr) {
			return nil, lastErr
		}
		
		em.logger.WithComponent("browser").Warn("Screenshot failed, will retry",
			zap.Error(lastErr),
			zap.Int("attempt", attempt))
	}
	
	return nil, fmt.Errorf("screenshot failed after %d retries: %w", em.maxRetries, lastErr)
}

// ExecuteScriptWithRetry executes JavaScript with automatic retry
func (em *EnhancedManager) ExecuteScriptWithRetry(pageID string, script string) (interface{}, error) {
	var lastErr error
	var result interface{}
	
	for attempt := 0; attempt <= em.maxRetries; attempt++ {
		if attempt > 0 {
			em.logger.WithComponent("browser").Info("Retrying script execution",
				zap.Int("attempt", attempt),
				zap.String("page_id", pageID))
			time.Sleep(em.retryDelay * time.Duration(attempt))
		}
		
		// Ensure page is healthy
		if err := em.ensurePageHealthy(pageID); err != nil {
			lastErr = err
			// Check for context errors
			if em.isContextError(err) {
				lastErr = em.handleContextError(err, "script_health_check")
			}
			continue
		}
		
		result, lastErr = em.ExecuteScript(pageID, script)
		if lastErr == nil {
			return result, nil
		}
		
		// Check for context errors first
		if em.isContextError(lastErr) {
			lastErr = em.handleContextError(lastErr, "script_execution")
			continue // Try again after restart
		}
		
		// Check if error is recoverable
		if !em.isRecoverableError(lastErr) {
			return nil, lastErr
		}
		
		em.logger.WithComponent("browser").Warn("Script execution failed, will retry",
			zap.Error(lastErr),
			zap.Int("attempt", attempt))
	}
	
	return nil, fmt.Errorf("script execution failed after %d retries: %w", em.maxRetries, lastErr)
}

// GetPageStatus returns the current status of a page
func (em *EnhancedManager) GetPageStatus(pageID string) (*PageStatus, error) {
	em.pageStatesMutex.RLock()
	state, exists := em.pageStates[pageID]
	em.pageStatesMutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("page %s not found", pageID)
	}
	
	page, err := em.GetPage(pageID)
	if err != nil {
		return &PageStatus{
			PageID:    pageID,
			IsHealthy: false,
			Error:     err.Error(),
		}, nil
	}
	
	// Check page health
	isHealthy := em.testPageHealth(page)
	
	return &PageStatus{
		PageID:        pageID,
		URL:           state.URL,
		Title:         state.Title,
		IsHealthy:     isHealthy,
		LastActive:    state.LastActive,
		RecoveryCount: state.RecoveryCount,
	}, nil
}

// PageStatus represents the current status of a page
type PageStatus struct {
	PageID        string    `json:"page_id"`
	URL           string    `json:"url"`
	Title         string    `json:"title"`
	IsHealthy     bool      `json:"is_healthy"`
	LastActive    time.Time `json:"last_active"`
	RecoveryCount int       `json:"recovery_count"`
	Error         string    `json:"error,omitempty"`
}

// RecoverPage attempts to recover an unhealthy page
func (em *EnhancedManager) RecoverPage(pageID string) error {
	em.pageStatesMutex.RLock()
	state, exists := em.pageStates[pageID]
	em.pageStatesMutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("page %s not found", pageID)
	}
	
	em.logger.WithComponent("browser").Info("Attempting page recovery",
		zap.String("page_id", pageID),
		zap.String("url", state.URL))
	
	// Close the old page
	if err := em.ClosePage(pageID); err != nil {
		em.logger.WithComponent("browser").Warn("Failed to close page during recovery",
			zap.String("page_id", pageID),
			zap.Error(err))
	}
	
	// Create a new page with the same URL
	page, newPageID, err := em.NewPageWithRetry(state.URL)
	if err != nil {
		return fmt.Errorf("failed to recover page: %w", err)
	}
	
	// Update page tracking
	em.mutex.Lock()
	delete(em.pages, pageID)
	em.pages[newPageID] = page
	em.mutex.Unlock()
	
	// Update page state
	em.pageStatesMutex.Lock()
	delete(em.pageStates, pageID)
	newState := &PageState{
		PageID:        newPageID,
		URL:           state.URL,
		Title:         state.Title,
		LastActive:    time.Now(),
		IsHealthy:     true,
		RecoveryCount: state.RecoveryCount + 1,
	}
	newState.Context, newState.Cancel = context.WithCancel(context.Background())
	em.pageStates[newPageID] = newState
	em.pageStatesMutex.Unlock()
	
	em.logger.WithComponent("browser").Info("Page recovered successfully",
		zap.String("old_page_id", pageID),
		zap.String("new_page_id", newPageID))
	
	return nil
}

// trackPageState tracks the state of a page for recovery
func (em *EnhancedManager) trackPageState(pageID, url string, page *rod.Page) {
	ctx, cancel := context.WithCancel(context.Background())
	
	state := &PageState{
		PageID:     pageID,
		URL:        url,
		LastActive: time.Now(),
		IsHealthy:  true,
		Context:    ctx,
		Cancel:     cancel,
	}
	
	// Try to get title
	if page != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					em.logger.WithComponent("browser").Debug("Failed to get page title", zap.Any("panic", r))
				}
			}()
			
			if info, err := page.Info(); err == nil && info != nil {
				state.Title = info.Title
			}
		}()
	}
	
	em.pageStatesMutex.Lock()
	em.pageStates[pageID] = state
	em.pageStatesMutex.Unlock()
}

// updatePageState updates the state of a page
func (em *EnhancedManager) updatePageState(pageID, url string) {
	em.pageStatesMutex.Lock()
	defer em.pageStatesMutex.Unlock()
	
	if state, exists := em.pageStates[pageID]; exists {
		state.URL = url
		state.LastActive = time.Now()
	}
}

// ensurePageHealthy ensures a page is healthy, attempting recovery if needed
func (em *EnhancedManager) ensurePageHealthy(pageID string) error {
	page, err := em.GetPage(pageID)
	if err != nil {
		// Page doesn't exist, try to recover
		return em.RecoverPage(pageID)
	}
	
	// Test page health
	if !em.testPageHealth(page) {
		em.logger.WithComponent("browser").Warn("Page unhealthy, attempting recovery",
			zap.String("page_id", pageID))
		return em.RecoverPage(pageID)
	}
	
	return nil
}

// testPageHealth tests if a page is healthy
func (em *EnhancedManager) testPageHealth(page *rod.Page) bool {
	if page == nil {
		return false
	}
	
	// Try to execute a simple script as health check
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	var healthy bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				healthy = false
			}
		}()
		
		result, err := page.Context(ctx).Eval("() => true")
		healthy = err == nil && result != nil
	}()
	
	return healthy
}

// isRecoverableError determines if an error is recoverable
func (em *EnhancedManager) isRecoverableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	recoverableErrors := []string{
		"context canceled",
		"context cancelled", // British spelling
		"context deadline exceeded",
		"timeout",
		"connection reset",
		"broken pipe",
		"target closed",
		"browser not started",
		"browser connection unhealthy",
		"page not found",
		"websocket: close",
		"connection refused",
		"network unreachable",
		"no such host",
	}
	
	for _, recoverable := range recoverableErrors {
		if strings.Contains(errStr, recoverable) {
			return true
		}
	}
	
	return false
}

// isContextError checks if the error is related to context cancellation or timeout
func (em *EnhancedManager) isContextError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check if it's a context error directly
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}
	
	// Check error message for context-related errors
	errStr := strings.ToLower(err.Error())
	contextErrors := []string{
		"context canceled",
		"context cancelled", // British spelling
		"context deadline exceeded",
		"context timeout",
		"operation was canceled",
		"operation was cancelled",
	}
	
	for _, contextErr := range contextErrors {
		if strings.Contains(errStr, contextErr) {
			return true
		}
	}
	
	return false
}

// WaitForElement waits for an element with retry logic
func (em *EnhancedManager) WaitForElement(pageID, selector string, timeout time.Duration) error {
	page, err := em.GetPage(pageID)
	if err != nil {
		return err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Use Rod's built-in wait with our timeout context
	_, err = page.Context(ctx).Element(selector)
	if err != nil {
		// Check if it's a timeout - retry with page recovery
		if ctx.Err() == context.DeadlineExceeded {
			em.logger.WithComponent("browser").Warn("Element wait timeout, attempting page recovery",
				zap.String("selector", selector))
			
			// Try to recover the page
			if recoverErr := em.RecoverPage(pageID); recoverErr != nil {
				return fmt.Errorf("element not found and recovery failed: %w", recoverErr)
			}
			
			// Try once more after recovery
			page, err = em.GetPage(pageID)
			if err != nil {
				return err
			}
			
			newCtx, newCancel := context.WithTimeout(context.Background(), timeout/2)
			defer newCancel()
			
			_, err = page.Context(newCtx).Element(selector)
		}
	}
	
	return err
}

// ClickElement clicks an element with retry logic
func (em *EnhancedManager) ClickElement(pageID, selector string) error {
	var lastErr error
	
	for attempt := 0; attempt <= em.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(em.retryDelay * time.Duration(attempt))
		}
		
		page, err := em.GetPage(pageID)
		if err != nil {
			lastErr = err
			continue
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		element, err := page.Context(ctx).Element(selector)
		if err != nil {
			lastErr = err
			continue
		}
		
		err = element.Click(proto.InputMouseButtonLeft, 1)
		if err == nil {
			return nil
		}
		
		lastErr = err
	}
	
	return fmt.Errorf("click failed after %d retries: %w", em.maxRetries, lastErr)
}

// GetElementText gets text from an element with retry logic
func (em *EnhancedManager) GetElementText(pageID, selector string) (string, error) {
	var lastErr error
	
	for attempt := 0; attempt <= em.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(em.retryDelay * time.Duration(attempt))
		}
		
		page, err := em.GetPage(pageID)
		if err != nil {
			lastErr = err
			continue
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		element, err := page.Context(ctx).Element(selector)
		if err != nil {
			lastErr = err
			continue
		}
		
		text, err := element.Text()
		if err == nil {
			return text, nil
		}
		
		lastErr = err
	}
	
	return "", fmt.Errorf("get text failed after %d retries: %w", em.maxRetries, lastErr)
}

// RestartBrowser performs an enhanced browser restart with exponential backoff
func (em *EnhancedManager) RestartBrowser() error {
	em.browserRestartMutex.Lock()
	defer em.browserRestartMutex.Unlock()
	
	// Calculate backoff delay based on restart attempts
	backoffDelay := em.calculateRestartBackoff()
	
	// Check if we should wait before restarting
	if time.Since(em.lastBrowserRestart) < backoffDelay {
		remainingWait := backoffDelay - time.Since(em.lastBrowserRestart)
		em.logger.WithComponent("browser").Info("Waiting before browser restart due to backoff",
			zap.Duration("wait_time", remainingWait),
			zap.Int("restart_attempts", em.browserRestartAttempts))
		time.Sleep(remainingWait)
	}
	
	em.browserRestartAttempts++
	em.lastBrowserRestart = time.Now()
	
	em.logger.WithComponent("browser").Info("Performing enhanced browser restart",
		zap.Int("attempt", em.browserRestartAttempts),
		zap.Duration("backoff_delay", backoffDelay))
	
	// Store current page URLs for restoration
	pageURLs := em.storePageURLs()
	
	// Perform the restart using base manager's restart functionality
	if err := em.Manager.EnsureHealthy(); err != nil {
		// EnsureHealthy will trigger a restart if needed
		return fmt.Errorf("browser restart failed: %w", err)
	}
	
	// Restore pages if any were open
	if len(pageURLs) > 0 {
		em.restorePages(pageURLs)
	}
	
	// Reset restart attempts on successful restart after a grace period
	go func() {
		time.Sleep(5 * time.Minute)
		em.browserRestartMutex.Lock()
		if time.Since(em.lastBrowserRestart) >= 5*time.Minute {
			em.browserRestartAttempts = 0
			em.logger.WithComponent("browser").Debug("Reset browser restart attempts after stable operation")
		}
		em.browserRestartMutex.Unlock()
	}()
	
	return nil
}

// calculateRestartBackoff calculates the backoff delay for browser restarts
func (em *EnhancedManager) calculateRestartBackoff() time.Duration {
	if em.browserRestartAttempts == 0 {
		return 0
	}
	
	// Exponential backoff: 1s, 2s, 4s, 8s, 16s, max 30s
	backoff := em.restartBackoffBase
	for i := 1; i < em.browserRestartAttempts && backoff < em.restartBackoffMax; i++ {
		backoff *= 2
	}
	
	if backoff > em.restartBackoffMax {
		backoff = em.restartBackoffMax
	}
	
	return backoff
}

// storePageURLs stores current page URLs for restoration after restart
func (em *EnhancedManager) storePageURLs() map[string]string {
	pageURLs := make(map[string]string)
	
	em.pageStatesMutex.RLock()
	for pageID, state := range em.pageStates {
		if state.URL != "" {
			pageURLs[pageID] = state.URL
		}
	}
	em.pageStatesMutex.RUnlock()
	
	return pageURLs
}

// restorePages attempts to restore pages after browser restart
func (em *EnhancedManager) restorePages(pageURLs map[string]string) {
	for oldPageID, url := range pageURLs {
		_, newPageID, err := em.NewPageWithRetry(url)
		if err != nil {
			em.logger.WithComponent("browser").Warn("Failed to restore page after restart",
				zap.String("old_page_id", oldPageID),
				zap.String("url", url),
				zap.Error(err))
		} else {
			em.logger.WithComponent("browser").Info("Restored page after restart",
				zap.String("old_page_id", oldPageID),
				zap.String("new_page_id", newPageID),
				zap.String("url", url))
		}
	}
}

// handleContextError detects context errors and triggers automatic restart
func (em *EnhancedManager) handleContextError(err error, operation string) error {
	if !em.isContextError(err) {
		return err // Not a context error, return as-is
	}
	
	em.logger.WithComponent("browser").Warn("Context error detected, triggering browser restart",
		zap.String("operation", operation),
		zap.Error(err))
	
	// Attempt automatic restart
	if restartErr := em.RestartBrowser(); restartErr != nil {
		return fmt.Errorf("context error in %s and restart failed: %v (restart error: %w)", operation, err, restartErr)
	}
	
	return fmt.Errorf("context error in %s, browser restarted successfully: %w", operation, err)
}