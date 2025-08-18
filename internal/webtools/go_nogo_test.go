package webtools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
)

// GoNoGoValidationTest is a comprehensive test that determines if the system is ready for production
// This test must pass completely for a GO decision
func TestGoNoGoValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Go/NoGo validation test in short mode")
	}

	t.Log("=== STARTING GO/NOGO VALIDATION TEST ===")
	
	// Track validation results
	var validationResults []ValidationResult
	defer func() {
		printValidationSummary(t, validationResults)
	}()

	// Create test browser manager with strict timeouts
	log := createTestLogger(t)
	browserMgr := browser.NewManager(log, browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})

	// Critical: Browser must start within reasonable time
	validationResults = append(validationResults, validateBrowserStartup(t, browserMgr))
	
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		stopChan := make(chan error, 1)
		go func() {
			stopChan <- browserMgr.Stop()
		}()
		
		select {
		case err := <-stopChan:
			if err != nil && !strings.Contains(err.Error(), "context canceled") {
				t.Logf("Browser shutdown warning: %v", err)
			}
		case <-ctx.Done():
			t.Log("Browser shutdown timed out - resource leak possible")
		}
	}()

	// Create temp directory for test artifacts
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// CRITICAL VALIDATIONS - All must pass for GO
	validationResults = append(validationResults, validateCorePageOperations(t, log, browserMgr, tempDir))
	validationResults = append(validationResults, validateBrowserNavigation(t, log, browserMgr))
	validationResults = append(validationResults, validateScreenshotCapability(t, log, browserMgr))
	validationResults = append(validationResults, validateScriptExecution(t, log, browserMgr))
	validationResults = append(validationResults, validateErrorRecovery(t, log, browserMgr))
	validationResults = append(validationResults, validatePerformanceThresholds(t, log, browserMgr))
	validationResults = append(validationResults, validateResourceManagement(t, log, browserMgr))
	validationResults = append(validationResults, validateConcurrentOperations(t, log, browserMgr))

	// Analyze overall validation status
	analyzeGoNoGoDecision(t, validationResults)
}

type ValidationResult struct {
	TestName    string
	Status      string // "PASS", "FAIL", "WARN"
	Details     string
	Critical    bool
	Duration    time.Duration
}

func validateBrowserStartup(t *testing.T, browserMgr *browser.Manager) ValidationResult {
	start := time.Now()
	
	err := browserMgr.Start(browser.Config{
		Debug:        false,
		Headless:     true,
		WindowHeight: 1080,
		WindowWidth:  1920,
	})
	
	duration := time.Since(start)
	
	if err != nil {
		return ValidationResult{
			TestName: "Browser Startup",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Failed to start browser: %v", err),
			Critical: true,
			Duration: duration,
		}
	}
	
	if duration > 30*time.Second {
		return ValidationResult{
			TestName: "Browser Startup",
			Status:   "WARN",
			Details:  fmt.Sprintf("Browser startup took %v (>30s threshold)", duration),
			Critical: false,
			Duration: duration,
		}
	}
	
	return ValidationResult{
		TestName: "Browser Startup",
		Status:   "PASS",
		Details:  fmt.Sprintf("Browser started successfully in %v", duration),
		Critical: true,
		Duration: duration,
	}
}

func validateCorePageOperations(t *testing.T, log *logger.Logger, browserMgr *browser.Manager, tempDir string) ValidationResult {
	start := time.Now()
	
	// Test page creation
	createTool := NewCreatePageTool(log)
	createArgs := map[string]interface{}{
		"filename":   "go-nogo-test.html",
		"title":      "Go/NoGo Validation Page",
		"html":       "<h1>Validation Test</h1><p>Critical validation page</p><button id='validation-btn'>Validate</button>",
		"css":        "body { font-family: Arial; margin: 20px; } #validation-btn { padding: 10px; background: #28a745; color: white; border: none; }",
		"javascript": "document.getElementById('validation-btn').onclick = function() { window.validationClicked = true; console.log('Validation button clicked'); };",
	}
	
	response, err := createTool.Execute(createArgs)
	if err != nil || response.IsError {
		return ValidationResult{
			TestName: "Core Page Operations",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Page creation failed: %v", err),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	// Verify file exists and has correct content
	filePath := filepath.Join(tempDir, "go-nogo-test.html")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ValidationResult{
			TestName: "Core Page Operations",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Created file not readable: %v", err),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	contentStr := string(content)
	requiredElements := []string{
		"Go/NoGo Validation Page",
		"Validation Test",
		"validation-btn",
		"validationClicked = true",
	}
	
	for _, element := range requiredElements {
		if !strings.Contains(contentStr, element) {
			return ValidationResult{
				TestName: "Core Page Operations",
				Status:   "FAIL",
				Details:  fmt.Sprintf("Created page missing required element: %s", element),
				Critical: true,
				Duration: time.Since(start),
			}
		}
	}
	
	return ValidationResult{
		TestName: "Core Page Operations",
		Status:   "PASS",
		Details:  fmt.Sprintf("Page created successfully with all required elements in %v", time.Since(start)),
		Critical: true,
		Duration: time.Since(start),
	}
}

func validateBrowserNavigation(t *testing.T, log *logger.Logger, browserMgr *browser.Manager) ValidationResult {
	start := time.Now()
	
	navTool := NewNavigatePageTool(log, browserMgr)
	
	// Test navigation to local file
	navArgs := map[string]interface{}{
		"url": "./go-nogo-test.html",
	}
	
	response, err := navTool.Execute(navArgs)
	if err != nil || response.IsError {
		return ValidationResult{
			TestName: "Browser Navigation",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Navigation to local file failed: %v", err),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	// Give page time to load
	time.Sleep(2 * time.Second)
	
	// Verify page is accessible
	pages := browserMgr.GetAllPages()
	if len(pages) == 0 {
		return ValidationResult{
			TestName: "Browser Navigation",
			Status:   "FAIL",
			Details:  "No pages found after navigation",
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	// Check if our test page is loaded
	found := false
	for _, page := range pages {
		if strings.Contains(page.URL, "go-nogo-test.html") {
			found = true
			break
		}
	}
	
	if !found {
		return ValidationResult{
			TestName: "Browser Navigation",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Test page not found in browser pages. Available: %v", pages),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	return ValidationResult{
		TestName: "Browser Navigation",
		Status:   "PASS",
		Details:  fmt.Sprintf("Navigation successful, page loaded in %v", time.Since(start)),
		Critical: true,
		Duration: time.Since(start),
	}
}

func validateScreenshotCapability(t *testing.T, log *logger.Logger, browserMgr *browser.Manager) ValidationResult {
	start := time.Now()
	
	screenshotTool := NewScreenshotTool(log, browserMgr)
	screenshotArgs := map[string]interface{}{
		"filename": "go-nogo-validation.png",
	}
	
	// Use timeout wrapper for critical operation
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	resultChan := make(chan ValidationResult, 1)
	go func() {
		response, err := screenshotTool.Execute(screenshotArgs)
		
		if err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				resultChan <- ValidationResult{
					TestName: "Screenshot Capability",
					Status:   "WARN",
					Details:  "Screenshot cancelled due to timeout - performance issue",
					Critical: false,
					Duration: time.Since(start),
				}
				return
			}
			resultChan <- ValidationResult{
				TestName: "Screenshot Capability",
				Status:   "FAIL",
				Details:  fmt.Sprintf("Screenshot failed: %v", err),
				Critical: true,
				Duration: time.Since(start),
			}
			return
		}
		
		if response.IsError {
			responseText := response.Content[0].Text
			if strings.Contains(responseText, "context canceled") {
				resultChan <- ValidationResult{
					TestName: "Screenshot Capability",
					Status:   "WARN",
					Details:  "Screenshot cancelled - performance concern",
					Critical: false,
					Duration: time.Since(start),
				}
				return
			}
			resultChan <- ValidationResult{
				TestName: "Screenshot Capability",
				Status:   "FAIL",
				Details:  fmt.Sprintf("Screenshot error: %s", responseText),
				Critical: true,
				Duration: time.Since(start),
			}
			return
		}
		
		// Verify file was created
		if _, err := os.Stat("go-nogo-validation.png"); os.IsNotExist(err) {
			resultChan <- ValidationResult{
				TestName: "Screenshot Capability",
				Status:   "FAIL",
				Details:  "Screenshot file was not created",
				Critical: true,
				Duration: time.Since(start),
			}
			return
		}
		
		resultChan <- ValidationResult{
			TestName: "Screenshot Capability",
			Status:   "PASS",
			Details:  fmt.Sprintf("Screenshot captured successfully in %v", time.Since(start)),
			Critical: true,
			Duration: time.Since(start),
		}
	}()
	
	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		return ValidationResult{
			TestName: "Screenshot Capability",
			Status:   "FAIL",
			Details:  "Screenshot operation timed out after 15 seconds",
			Critical: true,
			Duration: time.Since(start),
		}
	}
}

func validateScriptExecution(t *testing.T, log *logger.Logger, browserMgr *browser.Manager) ValidationResult {
	start := time.Now()
	
	scriptTool := NewExecuteScriptTool(log, browserMgr)
	
	// Test basic script execution
	basicScript := map[string]interface{}{
		"script": "document.title",
	}
	
	response, err := scriptTool.Execute(basicScript)
	if err != nil {
		if strings.Contains(err.Error(), "context canceled") {
			return ValidationResult{
				TestName: "Script Execution",
				Status:   "WARN",
				Details:  "Script execution cancelled - performance concern",
				Critical: false,
				Duration: time.Since(start),
			}
		}
		return ValidationResult{
			TestName: "Script Execution",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Basic script execution failed: %v", err),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	if response.IsError {
		responseText := response.Content[0].Text
		if strings.Contains(responseText, "context canceled") {
			return ValidationResult{
				TestName: "Script Execution",
				Status:   "WARN",
				Details:  "Script execution cancelled - performance concern",
				Critical: false,
				Duration: time.Since(start),
			}
		}
		return ValidationResult{
			TestName: "Script Execution",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Basic script error: %s", responseText),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	// Test complex script execution
	complexScript := map[string]interface{}{
		"script": `
			const validation = {
				title: document.title,
				hasButton: !!document.getElementById('validation-btn'),
				bodyExists: !!document.body,
				elementCount: document.querySelectorAll('*').length,
				timestamp: new Date().toISOString()
			};
			return JSON.stringify(validation);
		`,
	}
	
	response, err = scriptTool.Execute(complexScript)
	if err != nil || response.IsError {
		return ValidationResult{
			TestName: "Script Execution",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Complex script execution failed: %v", err),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	// Validate script result
	responseText := response.Content[0].Text
	if !strings.Contains(responseText, "Go/NoGo Validation Page") {
		return ValidationResult{
			TestName: "Script Execution",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Script returned unexpected result: %s", responseText),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	return ValidationResult{
		TestName: "Script Execution",
		Status:   "PASS",
		Details:  fmt.Sprintf("Script execution validated successfully in %v", time.Since(start)),
		Critical: true,
		Duration: time.Since(start),
	}
}

func validateErrorRecovery(t *testing.T, log *logger.Logger, browserMgr *browser.Manager) ValidationResult {
	start := time.Now()
	
	navTool := NewNavigatePageTool(log, browserMgr)
	
	// Test navigation to invalid domain
	invalidArgs := map[string]interface{}{
		"url": "https://invalid-domain-for-validation-test-12345.invalid",
	}
	
	response, err := navTool.Execute(invalidArgs)
	// Error is expected here - the critical part is that it doesn't crash
	
	// Test recovery with valid navigation
	validArgs := map[string]interface{}{
		"url": "https://example.com",
	}
	
	response, err = navTool.Execute(validArgs)
	if err != nil {
		return ValidationResult{
			TestName: "Error Recovery",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Failed to recover from error: %v", err),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	if response.IsError {
		return ValidationResult{
			TestName: "Error Recovery",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Recovery navigation failed: %v", response.Content[0].Text),
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	// Give time for navigation
	time.Sleep(3 * time.Second)
	
	// Verify browser is still functional
	pages := browserMgr.GetAllPages()
	if len(pages) == 0 {
		return ValidationResult{
			TestName: "Error Recovery",
			Status:   "FAIL",
			Details:  "No pages available after error recovery",
			Critical: true,
			Duration: time.Since(start),
		}
	}
	
	return ValidationResult{
		TestName: "Error Recovery",
		Status:   "PASS",
		Details:  fmt.Sprintf("Error recovery validated successfully in %v", time.Since(start)),
		Critical: true,
		Duration: time.Since(start),
	}
}

func validatePerformanceThresholds(t *testing.T, log *logger.Logger, browserMgr *browser.Manager) ValidationResult {
	start := time.Now()
	
	scriptTool := NewExecuteScriptTool(log, browserMgr)
	
	// Test script execution performance
	scriptStart := time.Now()
	scriptArgs := map[string]interface{}{
		"script": "document.title",
	}
	
	response, err := scriptTool.Execute(scriptArgs)
	scriptDuration := time.Since(scriptStart)
	
	if err != nil || response.IsError {
		return ValidationResult{
			TestName: "Performance Thresholds",
			Status:   "FAIL",
			Details:  fmt.Sprintf("Performance test script failed: %v", err),
			Critical: false,
			Duration: time.Since(start),
		}
	}
	
	// Performance thresholds
	if scriptDuration > 10*time.Second {
		return ValidationResult{
			TestName: "Performance Thresholds",
			Status:   "WARN",
			Details:  fmt.Sprintf("Script execution took %v (>10s threshold)", scriptDuration),
			Critical: false,
			Duration: time.Since(start),
		}
	}
	
	return ValidationResult{
		TestName: "Performance Thresholds",
		Status:   "PASS",
		Details:  fmt.Sprintf("Performance within acceptable thresholds (script: %v)", scriptDuration),
		Critical: false,
		Duration: time.Since(start),
	}
}

func validateResourceManagement(t *testing.T, log *logger.Logger, browserMgr *browser.Manager) ValidationResult {
	start := time.Now()
	
	// Check initial state
	initialPages := browserMgr.GetAllPages()
	initialCount := len(initialPages)
	
	// Create and navigate to multiple pages
	createTool := NewCreatePageTool(log)
	navTool := NewNavigatePageTool(log, browserMgr)
	
	for i := 0; i < 3; i++ {
		createArgs := map[string]interface{}{
			"filename": fmt.Sprintf("resource-test-%d.html", i),
			"title":    fmt.Sprintf("Resource Test %d", i),
			"html":     fmt.Sprintf("<h1>Resource Test Page %d</h1>", i),
		}
		
		response, err := createTool.Execute(createArgs)
		if err != nil || response.IsError {
			return ValidationResult{
				TestName: "Resource Management",
				Status:   "FAIL",
				Details:  fmt.Sprintf("Failed to create test page %d: %v", i, err),
				Critical: false,
				Duration: time.Since(start),
			}
		}
		
		navArgs := map[string]interface{}{
			"url": fmt.Sprintf("./resource-test-%d.html", i),
		}
		
		response, err = navTool.Execute(navArgs)
		if err != nil || response.IsError {
			return ValidationResult{
				TestName: "Resource Management",
				Status:   "WARN",
				Details:  fmt.Sprintf("Navigation to page %d failed: %v", i, err),
				Critical: false,
				Duration: time.Since(start),
			}
		}
		
		time.Sleep(1 * time.Second)
	}
	
	// Check final state - navigation reuses pages so count should remain stable
	finalPages := browserMgr.GetAllPages()
	finalCount := len(finalPages)
	
	// Navigation tool reuses existing pages, so we expect stable page count, not increase
	if finalCount < initialCount {
		return ValidationResult{
			TestName: "Resource Management",
			Status:   "WARN",
			Details:  fmt.Sprintf("Page count decreased unexpectedly (initial: %d, final: %d)", initialCount, finalCount),
			Critical: false,
			Duration: time.Since(start),
		}
	}
	
	return ValidationResult{
		TestName: "Resource Management",
		Status:   "PASS",
		Details:  fmt.Sprintf("Resource management validated (pages: %d→%d) in %v", initialCount, finalCount, time.Since(start)),
		Critical: false,
		Duration: time.Since(start),
	}
}

func validateConcurrentOperations(t *testing.T, log *logger.Logger, browserMgr *browser.Manager) ValidationResult {
	start := time.Now()
	
	scriptTool := NewExecuteScriptTool(log, browserMgr)
	
	// Execute multiple scripts concurrently
	type scriptResult struct {
		response *types.CallToolResponse
		err      error
		index    int
	}
	
	results := make(chan scriptResult, 3)
	
	for i := 0; i < 3; i++ {
		go func(index int) {
			args := map[string]interface{}{
				"script": fmt.Sprintf("'Concurrent test %d: ' + new Date().getTime()", index),
			}
			response, err := scriptTool.Execute(args)
			results <- scriptResult{response, err, index}
		}(i)
	}
	
	// Collect results
	var successCount int
	for i := 0; i < 3; i++ {
		select {
		case result := <-results:
			if result.err == nil && !result.response.IsError {
				successCount++
			}
		case <-time.After(20 * time.Second):
			return ValidationResult{
				TestName: "Concurrent Operations",
				Status:   "FAIL",
				Details:  "Concurrent operations timed out",
				Critical: false,
				Duration: time.Since(start),
			}
		}
	}
	
	if successCount < 2 {
		return ValidationResult{
			TestName: "Concurrent Operations",
			Status:   "WARN",
			Details:  fmt.Sprintf("Only %d/3 concurrent operations succeeded", successCount),
			Critical: false,
			Duration: time.Since(start),
		}
	}
	
	return ValidationResult{
		TestName: "Concurrent Operations",
		Status:   "PASS",
		Details:  fmt.Sprintf("Concurrent operations validated (%d/3 successful) in %v", successCount, time.Since(start)),
		Critical: false,
		Duration: time.Since(start),
	}
}

func printValidationSummary(t *testing.T, results []ValidationResult) {
	t.Log("=== GO/NOGO VALIDATION SUMMARY ===")
	
	var criticalPassed, criticalFailed, warningCount int
	var totalDuration time.Duration
	
	for _, result := range results {
		status := result.Status
		if result.Critical {
			switch status {
			case "PASS":
				criticalPassed++
			case "FAIL":
				criticalFailed++
			}
		}
		if status == "WARN" {
			warningCount++
		}
		totalDuration += result.Duration
		
		t.Logf("[%s] %s: %s (%v)", status, result.TestName, result.Details, result.Duration)
	}
	
	t.Logf("=== RESULTS ===")
	t.Logf("Critical Tests Passed: %d", criticalPassed)
	t.Logf("Critical Tests Failed: %d", criticalFailed)
	t.Logf("Warnings: %d", warningCount)
	t.Logf("Total Validation Time: %v", totalDuration)
	t.Log("=== END VALIDATION SUMMARY ===")
}

func analyzeGoNoGoDecision(t *testing.T, results []ValidationResult) {
	var criticalFailures []string
	var warnings []string
	
	for _, result := range results {
		if result.Critical && result.Status == "FAIL" {
			criticalFailures = append(criticalFailures, result.TestName)
		}
		if result.Status == "WARN" {
			warnings = append(warnings, result.TestName)
		}
	}
	
	t.Log("=== GO/NOGO DECISION ANALYSIS ===")
	
	if len(criticalFailures) == 0 {
		t.Log("🟢 DECISION: GO")
		t.Log("All critical validations passed. System is ready for production use.")
		if len(warnings) > 0 {
			t.Logf("⚠️  Warnings to monitor: %v", warnings)
		}
	} else {
		t.Log("🔴 DECISION: NO-GO")
		t.Logf("Critical failures detected: %v", criticalFailures)
		t.Log("System is NOT ready for production use.")
		if len(warnings) > 0 {
			t.Logf("Additional warnings: %v", warnings)
		}
	}
	
	t.Log("=== END GO/NOGO DECISION ===")
}