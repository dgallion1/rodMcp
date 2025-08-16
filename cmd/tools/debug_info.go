package tools

import (
	"fmt"
	"rodmcp/internal/browser"
	"rodmcp/internal/circuitbreaker"
	"rodmcp/internal/connection"
	"rodmcp/pkg/types"
	"runtime"
	"time"
)

// DebugInfoTool provides detailed debugging information
type DebugInfoTool struct {
	browserMgr     *browser.EnhancedManager
	connectionMgr  *connection.ConnectionManager
	circuitBreaker *circuitbreaker.MultiLevelCircuitBreaker
}

// NewDebugInfoTool creates a new debug info tool
func NewDebugInfoTool(
	browserMgr *browser.EnhancedManager,
	connectionMgr *connection.ConnectionManager,
	circuitBreaker *circuitbreaker.MultiLevelCircuitBreaker,
) *DebugInfoTool {
	return &DebugInfoTool{
		browserMgr:     browserMgr,
		connectionMgr:  connectionMgr,
		circuitBreaker: circuitBreaker,
	}
}

func (t *DebugInfoTool) Name() string {
	return "debug_info"
}

func (t *DebugInfoTool) Description() string {
	return "Get detailed debugging information about the current state of RodMCP"
}

func (t *DebugInfoTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"verbose": map[string]interface{}{
				"type":        "boolean",
				"description": "Include verbose details (optional, default: false)",
			},
		},
	}
}

func (t *DebugInfoTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	verbose := false
	if v, ok := args["verbose"].(bool); ok {
		verbose = v
	}

	report := "=== RodMCP Debug Information ===\n"
	report += fmt.Sprintf("Timestamp: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// System information
	report += "=== System Info ===\n"
	report += fmt.Sprintf("Go Version: %s\n", runtime.Version())
	report += fmt.Sprintf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	report += fmt.Sprintf("Goroutines: %d\n", runtime.NumGoroutine())
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	report += fmt.Sprintf("Memory: Alloc=%v MB, Sys=%v MB\n\n",
		m.Alloc/1024/1024, m.Sys/1024/1024)

	// Browser information
	report += "=== Browser Status ===\n"
	browserErr := t.browserMgr.CheckHealth()
	if browserErr == nil {
		report += "Status: ✅ Healthy\n"
	} else {
		report += fmt.Sprintf("Status: ❌ Unhealthy - %s\n", browserErr.Error())
	}
	
	pages := t.browserMgr.GetAllPages()
	report += fmt.Sprintf("Open Pages: %d\n", len(pages))
	
	if verbose && len(pages) > 0 {
		for _, page := range pages {
			status, _ := t.browserMgr.GetPageStatus(page.PageID)
			if status != nil {
				report += fmt.Sprintf("  - %s: %s (Healthy: %v, Recoveries: %d)\n",
					page.PageID, page.URL, status.IsHealthy, status.RecoveryCount)
			}
		}
	}
	report += "\n"

	// Connection information
	if t.connectionMgr != nil {
		report += "=== Connection Status ===\n"
		stats := t.connectionMgr.GetStats()
		report += fmt.Sprintf("Connected: %v\n", stats["connected"])
		report += fmt.Sprintf("Reconnect Count: %v\n", stats["reconnect_count"])
		if idleTime, ok := stats["idle_time"].(time.Duration); ok {
			report += fmt.Sprintf("Idle Time: %v\n", idleTime)
		}
		if verbose {
			report += fmt.Sprintf("Connection Attempts: %v\n", stats["connection_attempts"])
			report += fmt.Sprintf("Input Buffer Size: %v\n", stats["input_buffer_size"])
			report += fmt.Sprintf("Output Buffer Size: %v\n", stats["output_buffer_size"])
		}
		report += "\n"
	}

	// Circuit breaker information
	if t.circuitBreaker != nil {
		report += "=== Circuit Breaker Status ===\n"
		cbStats := t.circuitBreaker.GetOverallStats()
		
		if browserState, ok := cbStats["BrowserState"]; ok {
			report += fmt.Sprintf("Browser Circuit: %v", browserState)
			if failures, ok := cbStats["BrowserFailures"]; ok {
				if requests, ok := cbStats["BrowserRequests"]; ok {
					report += fmt.Sprintf(" (Failures: %v/%v)", failures, requests)
				}
			}
			report += "\n"
		}
		
		if networkState, ok := cbStats["NetworkState"]; ok {
			report += fmt.Sprintf("Network Circuit: %v", networkState)
			if failures, ok := cbStats["NetworkFailures"]; ok {
				if requests, ok := cbStats["NetworkRequests"]; ok {
					report += fmt.Sprintf(" (Failures: %v/%v)", failures, requests)
				}
			}
			report += "\n"
		}
		
		if toolState, ok := cbStats["ToolState"]; ok {
			report += fmt.Sprintf("Tool Circuit: %v", toolState)
			if failures, ok := cbStats["ToolFailures"]; ok {
				if requests, ok := cbStats["ToolRequests"]; ok {
					report += fmt.Sprintf(" (Failures: %v/%v)", failures, requests)
				}
			}
			report += "\n"
		}
		
		if verbose {
			report += fmt.Sprintf("\nCircuit Breaker Thresholds:\n")
			report += fmt.Sprintf("  Failure Threshold: 5 failures\n")
			report += fmt.Sprintf("  Success Threshold: 2 successes\n")
			report += fmt.Sprintf("  Timeout: 30 seconds\n")
			report += fmt.Sprintf("  Half-Open Max Requests: 3\n")
		}
		report += "\n"
	}

	// Performance metrics
	if verbose {
		report += "=== Performance Metrics ===\n"
		report += fmt.Sprintf("CPU Cores: %d\n", runtime.NumCPU())
		report += fmt.Sprintf("CGO Calls: %d\n", runtime.NumCgoCall())
		
		// Force GC and get stats
		runtime.GC()
		runtime.ReadMemStats(&m)
		report += fmt.Sprintf("GC Runs: %d\n", m.NumGC)
		report += fmt.Sprintf("GC Pause Total: %v ms\n", m.PauseTotalNs/1000000)
		report += "\n"
	}

	// Recommendations
	report += "=== Recommendations ===\n"
	recommendations := []string{}
	
	if browserErr != nil {
		recommendations = append(recommendations, "⚠️ Browser is unhealthy - consider using 'recover_page' or restarting")
	}
	
	if len(pages) > 10 {
		recommendations = append(recommendations, "⚠️ Many pages open - consider closing unused pages to free resources")
	}
	
	if t.connectionMgr != nil {
		stats := t.connectionMgr.GetStats()
		if reconnects, ok := stats["reconnect_count"].(int64); ok && reconnects > 5 {
			recommendations = append(recommendations, "⚠️ High reconnection count - check network stability")
		}
	}
	
	if cbStats := t.circuitBreaker.GetOverallStats(); cbStats != nil {
		if browserState, ok := cbStats["BrowserState"]; ok && browserState != "closed" {
			recommendations = append(recommendations, "⚠️ Browser circuit breaker is open - operations may be failing")
		}
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "✅ System appears healthy")
	}
	
	for _, rec := range recommendations {
		report += fmt.Sprintf("  %s\n", rec)
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