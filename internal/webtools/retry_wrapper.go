package webtools

import (
	"context"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/retry"
	"time"
)

// RetryWrapper provides retry functionality for browser operations
type RetryWrapper struct {
	browser        *browser.EnhancedManager
	logger         *logger.Logger
	strategyMgr    *retry.StrategyManager
	defaultTimeout time.Duration
}

// NewRetryWrapper creates a new retry wrapper for browser operations
func NewRetryWrapper(browser *browser.EnhancedManager, logger *logger.Logger) *RetryWrapper {
	return &RetryWrapper{
		browser:        browser,
		logger:         logger,
		strategyMgr:    retry.WithLogger(logger.Logger),
		defaultTimeout: 30 * time.Second,
	}
}

// NavigateWithRetry navigates to a URL with retry logic
func (rw *RetryWrapper) NavigateWithRetry(ctx context.Context, url string) (pageID string, err error) {
	err = rw.strategyMgr.RetryWithStrategy(ctx, "tool_operation", "navigate", func() error {
		// Check if there are existing pages, if so navigate the first one instead of creating new
		pages := rw.browser.GetAllPages()
		var currentPageID string
		
		if len(pages) > 0 {
			// Use existing page
			currentPageID = pages[0].PageID
			if navErr := rw.browser.NavigateWithRetry(currentPageID, url); navErr != nil {
				return navErr
			}
			pageID = currentPageID
		} else {
			// Create new page
			_, newPageID, createErr := rw.browser.NewPageWithRetry(url)
			if createErr != nil {
				return createErr
			}
			pageID = newPageID
		}
		
		return nil
	})
	
	return pageID, err
}

// ScreenshotWithRetry takes a screenshot with retry logic
func (rw *RetryWrapper) ScreenshotWithRetry(ctx context.Context, pageID string) ([]byte, error) {
	var screenshot []byte
	err := rw.strategyMgr.RetryWithStrategy(ctx, "tool_operation", "screenshot", func() error {
		data, screenshotErr := rw.browser.ScreenshotWithRetry(pageID)
		if screenshotErr != nil {
			return screenshotErr
		}
		screenshot = data
		return nil
	})
	
	return screenshot, err
}

// ExecuteScriptWithRetry executes JavaScript with retry logic
func (rw *RetryWrapper) ExecuteScriptWithRetry(ctx context.Context, pageID string, script string) (interface{}, error) {
	var result interface{}
	err := rw.strategyMgr.RetryWithStrategy(ctx, "tool_operation", "execute_script", func() error {
		data, execErr := rw.browser.ExecuteScriptWithRetry(pageID, script)
		if execErr != nil {
			return execErr
		}
		result = data
		return nil
	})
	
	return result, err
}

// ClickElementWithRetry clicks an element with retry logic
func (rw *RetryWrapper) ClickElementWithRetry(ctx context.Context, pageID string, selector string) error {
	return rw.strategyMgr.RetryWithStrategy(ctx, "tool_operation", "click_element", func() error {
		return rw.browser.ClickElement(pageID, selector)
	})
}

// GetElementTextWithRetry gets element text with retry logic
func (rw *RetryWrapper) GetElementTextWithRetry(ctx context.Context, pageID string, selector string) (string, error) {
	var text string
	err := rw.strategyMgr.RetryWithStrategy(ctx, "tool_operation", "get_element_text", func() error {
		result, getErr := rw.browser.GetElementText(pageID, selector)
		if getErr != nil {
			return getErr
		}
		text = result
		return nil
	})
	
	return text, err
}

// WaitForElementWithRetry waits for an element with retry logic
func (rw *RetryWrapper) WaitForElementWithRetry(ctx context.Context, pageID string, selector string, timeout time.Duration) error {
	return rw.strategyMgr.RetryWithStrategy(ctx, "tool_operation", "wait_for_element", func() error {
		return rw.browser.WaitForElement(pageID, selector, timeout)
	})
}

// CreatePageWithRetry creates a new page with retry logic
func (rw *RetryWrapper) CreatePageWithRetry(ctx context.Context, url string) (pageID string, err error) {
	err = rw.strategyMgr.RetryWithStrategy(ctx, "tool_operation", "create_page", func() error {
		_, newPageID, createErr := rw.browser.NewPageWithRetry(url)
		if createErr != nil {
			return createErr
		}
		pageID = newPageID
		return nil
	})
	
	return pageID, err
}

// GetPageStatusWithRetry gets page status with retry logic
func (rw *RetryWrapper) GetPageStatusWithRetry(ctx context.Context, pageID string) (*browser.PageStatus, error) {
	var status *browser.PageStatus
	err := rw.strategyMgr.RetryWithStrategy(ctx, "tool_operation", "get_page_status", func() error {
		result, statusErr := rw.browser.GetPageStatus(pageID)
		if statusErr != nil {
			return statusErr
		}
		status = result
		return nil
	})
	
	return status, err
}

// RecoverPageWithRetry recovers a page with retry logic
func (rw *RetryWrapper) RecoverPageWithRetry(ctx context.Context, pageID string) error {
	return rw.strategyMgr.RetryWithStrategy(ctx, "browser_operation", "recover_page", func() error {
		return rw.browser.RecoverPage(pageID)
	})
}

// EnsureHealthyWithRetry ensures browser is healthy with retry logic
func (rw *RetryWrapper) EnsureHealthyWithRetry(ctx context.Context) error {
	return rw.strategyMgr.RetryWithStrategy(ctx, "browser_operation", "ensure_healthy", func() error {
		return rw.browser.EnsureHealthy()
	})
}

// RestartBrowserWithRetry restarts browser with retry logic
func (rw *RetryWrapper) RestartBrowserWithRetry(ctx context.Context) error {
	return rw.strategyMgr.RetryWithStrategy(ctx, "critical_operation", "restart_browser", func() error {
		return rw.browser.RestartBrowser()
	})
}

// WithTimeout sets a custom timeout for operations
func (rw *RetryWrapper) WithTimeout(timeout time.Duration) *RetryWrapper {
	return &RetryWrapper{
		browser:        rw.browser,
		logger:         rw.logger,
		strategyMgr:    rw.strategyMgr,
		defaultTimeout: timeout,
	}
}

// GetStrategyInfo returns information about available retry strategies
func (rw *RetryWrapper) GetStrategyInfo() map[string]interface{} {
	strategies := rw.strategyMgr.ListStrategies()
	result := make(map[string]interface{})
	
	for _, strategy := range strategies {
		info, _ := rw.strategyMgr.GetStrategyInfo(strategy.Name)
		result[strategy.Name] = info
	}
	
	return result
}

// ExecuteWithRetry executes a generic operation with retry logic
func (rw *RetryWrapper) ExecuteWithRetry(ctx context.Context, strategyName string, operationName string, fn retry.RetryableFunc) error {
	return rw.strategyMgr.RetryWithStrategy(ctx, strategyName, operationName, fn)
}

// ExecuteWithRetryAndResult executes a generic operation with retry logic and returns a result
func (rw *RetryWrapper) ExecuteWithRetryAndResult(ctx context.Context, strategyName string, operationName string, fn retry.RetryableWithResultFunc) (interface{}, error) {
	return rw.strategyMgr.RetryWithStrategyAndResult(ctx, strategyName, operationName, fn)
}