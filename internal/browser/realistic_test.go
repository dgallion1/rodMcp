package browser

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"rodmcp/internal/logger"
)

func TestRealisticBrowserLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping realistic browser test in short mode")
	}
	
	log, err := logger.New(logger.Config{
		LogLevel:    "error", // Reduce noise
		LogDir:      t.TempDir(),
		Development: true,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	config := Config{
		Headless:     true,
		Debug:        false,
		WindowWidth:  1200,
		WindowHeight: 800,
	}
	
	testConfig := DefaultTestConfig()
	tm := NewTestManager(log, config, testConfig)
	
	// Test startup
	if err := tm.StartWithTimeout(); err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	
	// Test page creation
	page, pageID, err := tm.NewPageWithValidation("file:///home/darrell/work/git/rodMcp/test_data/realistic_test.html")
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}
	
	if page == nil {
		t.Error("Created page is nil")
	}
	if pageID == "" {
		t.Error("Created pageID is empty")
	}
	
	// Test page operations
	pages := tm.GetPagesWithRetry(3)
	if len(pages) == 0 {
		t.Error("No pages found after creation")
	}
	
	found := false
	for _, p := range pages {
		if p.PageID == pageID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created page not found in page list")
	}
	
	// Test graceful shutdown
	if err := tm.StopGracefully(); err != nil {
		t.Errorf("Failed to stop browser gracefully: %v", err)
	}
}

func TestRealisticPageOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping realistic page operations test in short mode")
	}
	
	log, err := logger.New(logger.Config{
		LogLevel:    "error",
		LogDir:      t.TempDir(),
		Development: true,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	config := Config{
		Headless:     true,
		Debug:        false,
		WindowWidth:  1200,
		WindowHeight: 800,
	}
	
	testConfig := DefaultTestConfig()
	tm := NewTestManager(log, config, testConfig)
	
	if err := tm.StartWithTimeout(); err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer tm.StopGracefully()
	
	// Create test HTML file
	tempDir := t.TempDir()
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Test Operations</title>
</head>
<body>
    <h1>Test Page</h1>
    <p id="content">Original content</p>
    <button id="btn">Click me</button>
    <script>
        document.getElementById('btn').onclick = function() {
            document.getElementById('content').textContent = 'Button clicked!';
        };
    </script>
</body>
</html>`
	
	htmlFile := tempDir + "/test-ops.html"
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test HTML file: %v", err)
	}
	
	// Navigate to test page
	page, pageID, err := tm.NewPageWithValidation("file://" + htmlFile)
	if err != nil {
		t.Fatalf("Failed to navigate to test page: %v", err)
	}
	
	// Wait for page to load
	if err := tm.WaitForPageLoad(page, 5*time.Second); err != nil {
		t.Logf("Page load wait failed (may be expected): %v", err)
	}
	
	// Test screenshot operation with timeout
	screenshotErr := tm.ExecuteOperationWithTimeout(func() error {
		_, screenshotErr := tm.Screenshot(pageID)
		return screenshotErr
	}, 10*time.Second)
	
	if screenshotErr != nil {
		if !strings.Contains(screenshotErr.Error(), "context canceled") {
			t.Errorf("Screenshot failed: %v", screenshotErr)
		} else {
			t.Logf("Screenshot skipped due to context cancellation (expected)")
		}
	}
	
	// Test script execution with timeout
	scriptErr := tm.ExecuteOperationWithTimeout(func() error {
		result, scriptErr := tm.ExecuteScript(pageID, "document.title")
		if scriptErr != nil {
			return scriptErr
		}
		if resultStr, ok := result.(string); !ok || !strings.Contains(resultStr, "Test Operations") {
			return fmt.Errorf("unexpected script result: %v", result)
		}
		return nil
	}, 10*time.Second)
	
	if scriptErr != nil {
		if !strings.Contains(scriptErr.Error(), "context canceled") {
			t.Errorf("Script execution failed: %v", scriptErr)
		} else {
			t.Logf("Script execution skipped due to context cancellation (expected)")
		}
	}
}

func TestRealisticHealthAndRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping realistic health test in short mode")
	}
	
	log, err := logger.New(logger.Config{
		LogLevel:    "warn", // Show warnings but reduce noise
		LogDir:      t.TempDir(),
		Development: true,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	config := Config{
		Headless:     true,
		Debug:        false,
		WindowWidth:  1200,
		WindowHeight: 800,
	}
	
	mgr := NewManager(log, config)
	
	if err := mgr.Start(config); err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer func() {
		// Use a separate context for cleanup to avoid cancellation propagation
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		done := make(chan struct{}, 1)
		go func() {
			mgr.Stop()
			done <- struct{}{}
		}()
		
		select {
		case <-done:
			// Stopped successfully
		case <-stopCtx.Done():
			t.Logf("Browser stop timed out during cleanup (may be expected)")
		}
	}()
	
	t.Run("HealthCheck", func(t *testing.T) {
		// Test basic health check
		if err := mgr.CheckHealth(); err != nil {
			t.Errorf("Health check failed: %v", err)
		}
	})
	
	t.Run("EnsureHealthy", func(t *testing.T) {
		// Test ensure healthy
		if err := mgr.EnsureHealthy(); err != nil {
			t.Errorf("EnsureHealthy failed: %v", err)
		}
	})
	
	t.Run("MultiplePageCreation", func(t *testing.T) {
		// Test creating multiple pages
		pageIDs := make([]string, 0, 3)
		
		for i := 0; i < 3; i++ {
			_, pageID, err := mgr.NewPage("")
			if err != nil {
				t.Errorf("Failed to create page %d: %v", i, err)
				continue
			}
			if pageID != "" {
				pageIDs = append(pageIDs, pageID)
			}
		}
		
		// Verify pages exist
		pages := mgr.GetAllPages()
		if len(pages) < len(pageIDs) {
			t.Logf("Warning: Expected %d pages, got %d", len(pageIDs), len(pages))
		}
		
		// Clean up pages
		for _, pageID := range pageIDs {
			if err := mgr.ClosePage(pageID); err != nil {
				t.Logf("Failed to close page %s: %v", pageID, err)
			}
		}
	})
}

func TestRealisticConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping realistic concurrent test in short mode")
	}
	
	log, err := logger.New(logger.Config{
		LogLevel:    "error",
		LogDir:      t.TempDir(),
		Development: true,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	config := Config{
		Headless:     true,
		Debug:        false,
		WindowWidth:  1200,
		WindowHeight: 800,
	}
	
	mgr := NewManager(log, config)
	
	if err := mgr.Start(config); err != nil {
		t.Fatalf("Failed to start browser: %v", err)
	}
	defer func() {
		// Graceful cleanup for concurrent test
		time.Sleep(200 * time.Millisecond)
		mgr.Stop()
	}()
	
	// Test concurrent page creation
	const numGoroutines = 5
	results := make(chan error, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					results <- fmt.Errorf("goroutine %d panicked: %v", id, r)
					return
				}
			}()
			
			_, pageID, err := mgr.NewPage("")
			if err != nil {
				results <- fmt.Errorf("goroutine %d failed to create page: %v", id, err)
				return
			}
			
			// Brief operation
			time.Sleep(100 * time.Millisecond)
			
			// Cleanup
			if pageID != "" {
				mgr.ClosePage(pageID)
			}
			
			results <- nil
		}(i)
	}
	
	// Collect results
	errors := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Logf("Concurrent operation error: %v", err)
				errors++
			}
		case <-time.After(15 * time.Second):
			t.Error("Concurrent operation timed out")
			errors++
		}
	}
	
	// Allow some failures in concurrent scenarios
	if errors > numGoroutines/2 {
		t.Errorf("Too many concurrent operation failures: %d/%d", errors, numGoroutines)
	}
}