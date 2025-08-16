package tools

import (
	"fmt"
	"rodmcp/internal/browser"
	"rodmcp/pkg/types"
	"time"
)

// BrowserHealthTool checks and reports browser health
type BrowserHealthTool struct {
	browserMgr *browser.EnhancedManager
}

// NewBrowserHealthTool creates a new browser health tool
func NewBrowserHealthTool(browserMgr *browser.EnhancedManager) *BrowserHealthTool {
	return &BrowserHealthTool{
		browserMgr: browserMgr,
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
	// Check browser health
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

	// Add page information
	if len(pages) > 0 {
		report += "\nOpen Pages:\n"
		for _, page := range pages {
			status, _ := t.browserMgr.GetPageStatus(page.PageID)
			healthStatus := "❓"
			if status != nil {
				if status.IsHealthy {
					healthStatus = "✅"
				} else {
					healthStatus = "❌"
				}
			}
			report += fmt.Sprintf("  %s %s - %s\n", healthStatus, page.PageID, page.URL)
		}
	}

	// Attempt to ensure healthy if not healthy
	if !isHealthy {
		report += "\nAttempting automatic recovery...\n"
		if ensureErr := t.browserMgr.EnsureHealthy(); ensureErr == nil {
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