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
	
	"github.com/go-rod/rod"
)

// TestBrowserManager provides a reusable browser instance for realistic testing
type TestBrowserManager struct {
	*browser.Manager
	tempDir string
	log     *logger.Logger
}

// NewTestBrowserManager creates a browser manager optimized for testing
func NewTestBrowserManager(t *testing.T) *TestBrowserManager {
	log := createTestLogger(t)
	
	// Create temp directory for test files
	tempDir := t.TempDir()
	
	// Change to temp directory for relative path tests
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})
	
	config := browser.Config{
		Headless:     true,
		Debug:        false,
		SlowMotion:   0,
		WindowHeight: 800,  // Smaller for faster startup
		WindowWidth:  1200,
	}
	
	mgr := browser.NewManager(log, config)
	
	// Start browser with proper timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	startChan := make(chan error, 1)
	go func() {
		startChan <- mgr.Start(config)
	}()
	
	select {
	case err := <-startChan:
		if err != nil {
			t.Fatalf("Failed to start browser: %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Browser startup timed out after 30 seconds")
	}
	
	testMgr := &TestBrowserManager{
		Manager: mgr,
		tempDir: tempDir,
		log:     log,
	}
	
	// Proper cleanup that doesn't cause context cancellation
	t.Cleanup(func() {
		// Give time for any pending operations
		time.Sleep(100 * time.Millisecond)
		
		// Stop browser gracefully
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		
		stopChan := make(chan error, 1)
		go func() {
			stopChan <- mgr.Stop()
		}()
		
		select {
		case err := <-stopChan:
			if err != nil && !strings.Contains(err.Error(), "context canceled") {
				t.Logf("Browser stop warning: %v", err)
			}
		case <-stopCtx.Done():
			t.Logf("Browser stop timed out, may have leaked resources")
		}
	})
	
	return testMgr
}

// CreateTestPage creates a test HTML page and returns its file path
func (tm *TestBrowserManager) CreateTestPage(t *testing.T, filename, content string) string {
	if filename == "" {
		filename = fmt.Sprintf("test-page-%d.html", time.Now().UnixNano())
	}
	
	if !strings.HasSuffix(filename, ".html") {
		filename += ".html"
	}
	
	fullPath := filepath.Join(tm.tempDir, filename)
	
	if content == "" {
		content = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <h1>Test Page</h1>
    <p>This is a test page created at %s</p>
    <button id="test-btn">Click Me</button>
    <div id="content">Test content</div>
    <script>
        document.getElementById('test-btn').onclick = function() {
            console.log('Button clicked!');
            document.getElementById('content').textContent = 'Button was clicked!';
        };
    </script>
</body>
</html>`, time.Now().Format("15:04:05"))
	}
	
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test page: %v", err)
	}
	
	return filename // Return relative path for navigation
}

// NavigateToPageWithRetry navigates to a page with realistic retry logic
func (tm *TestBrowserManager) NavigateToPageWithRetry(t *testing.T, url string, maxAttempts int) (*rod.Page, string) {
	var lastErr error
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		page, pageID, err := tm.NewPage(url)
		if err == nil {
			// Give page time to load
			time.Sleep(500 * time.Millisecond)
			return page, pageID
		}
		
		lastErr = err
		
		if attempt < maxAttempts {
			t.Logf("Navigation attempt %d failed: %v, retrying...", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	
	t.Fatalf("Failed to navigate after %d attempts: %v", maxAttempts, lastErr)
	return nil, ""
}

// ExecuteWithTimeout executes a browser operation with a realistic timeout
func (tm *TestBrowserManager) ExecuteWithTimeout(t *testing.T, operation func() error, timeout time.Duration, description string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	done := make(chan error, 1)
	go func() {
		done <- operation()
	}()
	
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("%s failed: %v", description, err)
		}
	case <-ctx.Done():
		t.Errorf("%s timed out after %v", description, timeout)
	}
}