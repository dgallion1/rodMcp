package tools

import (
	"context"
	"fmt"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/webtools"
	"rodmcp/pkg/types"
	"time"
)

// BrowserHealthTool checks and reports browser health
type BrowserHealthTool struct {
	browserMgr   *browser.EnhancedManager
	retryWrapper *webtools.RetryWrapper
}

// NewBrowserHealthTool creates a new browser health tool
func NewBrowserHealthTool(browserMgr *browser.EnhancedManager, logger *logger.Logger) *BrowserHealthTool {
	return &BrowserHealthTool{
		browserMgr:   browserMgr,
		retryWrapper: webtools.NewRetryWrapper(browserMgr, logger),
	}
}

func (t *BrowserHealthTool) Name() string {
	return "browser_health"
}

func (t *BrowserHealthTool) Description() string {
	return "Check the health and status of the browser instance"
}

func (t *BrowserHealthTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type:       "object",
		Properties: map[string]interface{}{},
	}
}

func (t *BrowserHealthTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Check browser health with retry
	err := t.browserMgr.CheckHealth()
	isHealthy := err == nil

	// Get all pages
	pages := t.browserMgr.GetAllPages()

	// Build health report
	report := fmt.Sprintf("Browser Health Report:\n"+
		"Status: %s\n"+
		"Open Pages: %d\n"+
		"Timestamp: %s\n",
		func() string {
			if isHealthy {
				return "✅ Healthy"
			}
			return "❌ Unhealthy"
		}(),
		len(pages),
		time.Now().Format("2006-01-02 15:04:05"))

	if err != nil {
		report += fmt.Sprintf("Error: %s\n", err.Error())
	}

	// Add page information with retry for page status
	if len(pages) > 0 {
		report += "\nOpen Pages:\n"
		for _, page := range pages {
			status, statusErr := t.retryWrapper.GetPageStatusWithRetry(ctx, page.PageID)
			healthStatus := "❓"
			if statusErr == nil && status != nil {
				if status.IsHealthy {
					healthStatus = "✅"
				} else {
					healthStatus = "❌"
				}
			}
			report += fmt.Sprintf("  %s %s - %s\n", healthStatus, page.PageID, page.URL)
		}
	}

	// Attempt to ensure healthy if not healthy with retry
	if !isHealthy {
		report += "\nAttempting automatic recovery...\n"
		if ensureErr := t.retryWrapper.EnsureHealthyWithRetry(ctx); ensureErr == nil {
			report += "✅ Recovery successful!\n"
		} else {
			report += fmt.Sprintf("❌ Recovery failed: %s\n", ensureErr.Error())
		}
	}

	content := []types.ToolContent{
		{
			Type: "text",
			Text: report,
		},
	}

	return &types.CallToolResponse{
		Content: content,
	}, nil
}