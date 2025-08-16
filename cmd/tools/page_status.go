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

// PageStatusTool provides page health and status information
type PageStatusTool struct {
	browserMgr   *browser.EnhancedManager
	retryWrapper *webtools.RetryWrapper
}

// NewPageStatusTool creates a new page status tool
func NewPageStatusTool(browserMgr *browser.EnhancedManager, logger *logger.Logger) *PageStatusTool {
	return &PageStatusTool{
		browserMgr:   browserMgr,
		retryWrapper: webtools.NewRetryWrapper(browserMgr, logger),
	}
}

func (t *PageStatusTool) Name() string {
	return "get_page_status"
}

func (t *PageStatusTool) Description() string {
	return "Get the current status and health of a browser page"
}

func (t *PageStatusTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the page to check status for",
			},
		},
		Required: []string{"page_id"},
	}
}

func (t *PageStatusTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pageID, ok := args["page_id"].(string)
	if !ok {
		return nil, fmt.Errorf("page_id must be a string")
	}

	// Use retry wrapper for getting page status
	status, err := t.retryWrapper.GetPageStatusWithRetry(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get page status: %w", err)
	}

	content := []types.ToolContent{
		{
			Type: "text",
			Text: fmt.Sprintf("Page Status for %s:\n"+
				"URL: %s\n"+
				"Title: %s\n"+
				"Healthy: %v\n"+
				"Last Active: %s\n"+
				"Recovery Count: %d",
				status.PageID,
				status.URL,
				status.Title,
				status.IsHealthy,
				status.LastActive.Format("2006-01-02 15:04:05"),
				status.RecoveryCount),
		},
	}

	if status.Error != "" {
		content[0].Text += fmt.Sprintf("\nError: %s", status.Error)
	}

	return &types.CallToolResponse{
		Content: content,
	}, nil
}