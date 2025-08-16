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

// RecoverPageTool recovers an unhealthy page
type RecoverPageTool struct {
	browserMgr   *browser.EnhancedManager
	retryWrapper *webtools.RetryWrapper
}

// NewRecoverPageTool creates a new page recovery tool
func NewRecoverPageTool(browserMgr *browser.EnhancedManager, logger *logger.Logger) *RecoverPageTool {
	return &RecoverPageTool{
		browserMgr:   browserMgr,
		retryWrapper: webtools.NewRetryWrapper(browserMgr, logger),
	}
}

func (t *RecoverPageTool) Name() string {
	return "recover_page"
}

func (t *RecoverPageTool) Description() string {
	return "Attempt to recover an unhealthy or unresponsive browser page"
}

func (t *RecoverPageTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"page_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the page to recover",
			},
		},
		Required: []string{"page_id"},
	}
}

func (t *RecoverPageTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pageID, ok := args["page_id"].(string)
	if !ok {
		return nil, fmt.Errorf("page_id must be a string")
	}

	// Get current status before recovery with retry
	statusBefore, _ := t.retryWrapper.GetPageStatusWithRetry(ctx, pageID)

	// Attempt recovery with retry logic
	err := t.retryWrapper.RecoverPageWithRetry(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to recover page: %w", err)
	}

	// Get status after recovery with retry
	statusAfter, _ := t.retryWrapper.GetPageStatusWithRetry(ctx, pageID)

	content := []types.ToolContent{
		{
			Type: "text",
			Text: fmt.Sprintf("Page Recovery Successful:\n"+
				"Original Page ID: %s\n"+
				"Original Health: %v\n"+
				"New Health: %v\n"+
				"URL: %s",
				pageID,
				statusBefore != nil && statusBefore.IsHealthy,
				statusAfter != nil && statusAfter.IsHealthy,
				func() string {
					if statusAfter != nil {
						return statusAfter.URL
					}
					if statusBefore != nil {
						return statusBefore.URL
					}
					return "unknown"
				}()),
		},
	}

	return &types.CallToolResponse{
		Content: content,
	}, nil
}