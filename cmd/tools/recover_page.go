package tools

import (
	"fmt"
	"rodmcp/internal/browser"
	"rodmcp/pkg/types"
)

// RecoverPageTool recovers an unhealthy page
type RecoverPageTool struct {
	browserMgr *browser.EnhancedManager
}

// NewRecoverPageTool creates a new page recovery tool
func NewRecoverPageTool(browserMgr *browser.EnhancedManager) *RecoverPageTool {
	return &RecoverPageTool{
		browserMgr: browserMgr,
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
		Properties: map[string]types.Property{
			"page_id": {
				Type:        "string",
				Description: "The ID of the page to recover",
			},
		},
		Required: []string{"page_id"},
	}
}

func (t *RecoverPageTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	pageID, ok := args["page_id"].(string)
	if !ok {
		return nil, fmt.Errorf("page_id must be a string")
	}

	// Get current status before recovery
	statusBefore, _ := t.browserMgr.GetPageStatus(pageID)

	// Attempt recovery
	err := t.browserMgr.RecoverPage(pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to recover page: %w", err)
	}

	// Get status after recovery
	statusAfter, _ := t.browserMgr.GetPageStatus(pageID)

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