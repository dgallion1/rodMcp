package browser

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"rodmcp/internal/logger"
	"strings"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"}) // Quiet for tests
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	
	if manager.logger == nil {
		t.Error("Manager logger is nil")
	}
	
	if manager.pages == nil {
		t.Error("Manager pages map is nil")
	}
	
	if manager.ctx == nil {
		t.Error("Manager context is nil")
	}
}

func TestManagerConfig(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{
		Headless:     true,
		Debug:        false,
		SlowMotion:   100 * time.Millisecond,
		WindowWidth:  1024,
		WindowHeight: 768,
	}
	
	manager := NewManager(log, config)
	
	// Start manager to store config
	if err := manager.Start(config); err != nil {
		t.Skipf("Skipping browser test (no browser available): %v", err)
	}
	defer manager.Stop()
	
	if manager.config.WindowWidth != 1024 {
		t.Errorf("Expected WindowWidth 1024, got %d", manager.config.WindowWidth)
	}
	
	if manager.config.WindowHeight != 768 {
		t.Errorf("Expected WindowHeight 768, got %d", manager.config.WindowHeight)
	}
	
	if manager.config.SlowMotion != 100*time.Millisecond {
		t.Errorf("Expected SlowMotion 100ms, got %v", manager.config.SlowMotion)
	}
}

func TestManagerStartStop(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	
	// Test start
	err := manager.Start(config)
	if err != nil {
		t.Skipf("Skipping browser test (no browser available): %v", err)
	}
	
	// Verify browser is started
	if manager.browser == nil {
		t.Error("Browser should be started but is nil")
	}
	
	// Test stop
	err = manager.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestNewPageWithoutBrowser(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	// Try to create page without starting browser
	_, _, err := manager.NewPage("")
	if err == nil {
		t.Error("Expected error when creating page without browser")
	}
	
	if !strings.Contains(err.Error(), "browser not started") {
		t.Errorf("Expected 'browser not started' error, got: %v", err)
	}
}

func TestNewPageAndClose(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	
	err := manager.Start(config)
	if err != nil {
		t.Skipf("Skipping browser test (no browser available): %v", err)
	}
	defer manager.Stop()
	
	// Create a page without URL
	page, pageID, err := manager.NewPage("")
	if err != nil {
		t.Fatalf("NewPage failed: %v", err)
	}
	
	if page == nil {
		t.Error("Page should not be nil")
	}
	
	if pageID == "" {
		t.Error("PageID should not be empty")
	}
	
	// Verify page is in manager's pages map
	pages := manager.ListPages()
	found := false
	for _, id := range pages {
		if id == pageID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Page not found in manager's pages list")
	}
	
	// Close the page
	err = manager.ClosePage(pageID)
	if err != nil {
		t.Errorf("ClosePage failed: %v", err)
	}
	
	// Verify page is removed
	pages = manager.ListPages()
	for _, id := range pages {
		if id == pageID {
			t.Error("Page should be removed after close")
		}
	}
}

func TestGetPageNotFound(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	_, err := manager.GetPage("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent page")
	}
	
	if !strings.Contains(err.Error(), "page not found") {
		t.Errorf("Expected 'page not found' error, got: %v", err)
	}
}

func TestClosePageNotFound(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	err := manager.ClosePage("nonexistent")
	if err == nil {
		t.Error("Expected error for closing nonexistent page")
	}
	
	if !strings.Contains(err.Error(), "page not found") {
		t.Errorf("Expected 'page not found' error, got: %v", err)
	}
}

func TestListPages(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	
	err := manager.Start(config)
	if err != nil {
		t.Skipf("Skipping browser test (no browser available): %v", err)
	}
	defer manager.Stop()
	
	// Initially no pages
	pages := manager.ListPages()
	if len(pages) != 0 {
		t.Errorf("Expected 0 pages, got %d", len(pages))
	}
	
	// Create some pages
	_, pageID1, err := manager.NewPage("")
	if err != nil {
		t.Fatalf("NewPage failed: %v", err)
	}
	
	_, pageID2, err := manager.NewPage("")
	if err != nil {
		t.Fatalf("NewPage failed: %v", err)
	}
	
	// List pages
	pages = manager.ListPages()
	if len(pages) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(pages))
	}
	
	// Verify page IDs
	pageMap := make(map[string]bool)
	for _, id := range pages {
		pageMap[id] = true
	}
	
	if !pageMap[pageID1] {
		t.Error("Page 1 not found in list")
	}
	
	if !pageMap[pageID2] {
		t.Error("Page 2 not found in list")
	}
	
	// Clean up
	manager.ClosePage(pageID1)
	manager.ClosePage(pageID2)
}

func TestScreenshotWithoutPage(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	_, err := manager.Screenshot("nonexistent")
	if err == nil {
		t.Error("Expected error for screenshot of nonexistent page")
	}
}

func TestExecuteScriptWithoutPage(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	_, err := manager.ExecuteScript("nonexistent", "1 + 1")
	if err == nil {
		t.Error("Expected error for script execution on nonexistent page")
	}
}

func TestFindWorkingBrowser(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	// This should either find a browser or return an empty string for Rod download
	browserPath, err := manager.findWorkingBrowser()
	if err != nil {
		// It's OK if no browser is found - that's what we test for
		if !strings.Contains(err.Error(), "no working browser binary found") {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		// If successful, should be either empty string or valid path
		if browserPath != "" {
			if _, statErr := os.Stat(browserPath); statErr != nil {
				t.Errorf("Browser path %s should exist: %v", browserPath, statErr)
			}
		}
	}
}

func TestIsBrowserWorking(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	// Test with nonexistent path
	if manager.isBrowserWorking("/nonexistent/browser") {
		t.Error("Nonexistent browser should not be working")
	}
	
	// Test with directory instead of file
	if manager.isBrowserWorking("/tmp") {
		t.Error("Directory should not be working as browser")
	}
	
	// Test with invalid executable
	if manager.isBrowserWorking("/bin/false") {
		t.Error("Invalid executable should not be working as browser")
	}
}

func TestIsURLReachable(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	// Test file:// URL (should always pass)
	err := manager.isURLReachable("file:///tmp/test.html")
	if err != nil {
		t.Errorf("file:// URL should be reachable: %v", err)
	}
	
	// Test invalid URL
	err = manager.isURLReachable("invalid://url")
	if err == nil {
		t.Error("Invalid URL should not be reachable")
	}
	
	// Test unreachable HTTP URL
	err = manager.isURLReachable("http://nonexistent.localhost:99999/test")
	if err == nil {
		t.Error("Unreachable URL should not be reachable")
	}
	
	// Test with a local test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	err = manager.isURLReachable(server.URL)
	if err != nil {
		t.Errorf("Test server URL should be reachable: %v", err)
	}
}

func TestCheckHealthWithoutBrowser(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	err := manager.CheckHealth()
	if err == nil {
		t.Error("Health check should fail without browser")
	}
	
	if !strings.Contains(err.Error(), "browser not started") {
		t.Errorf("Expected 'browser not started' error, got: %v", err)
	}
}

func TestGetPageInfo(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	
	err := manager.Start(config)
	if err != nil {
		t.Skipf("Skipping browser test (no browser available): %v", err)
	}
	defer manager.Stop()
	
	// Create a page
	_, pageID, err := manager.NewPage("")
	if err != nil {
		t.Fatalf("NewPage failed: %v", err)
	}
	defer manager.ClosePage(pageID)
	
	// Get page info
	info, err := manager.GetPageInfo(pageID)
	if err != nil {
		t.Errorf("GetPageInfo failed: %v", err)
	}
	
	if info["id"] != pageID {
		t.Errorf("Expected page ID %s, got %v", pageID, info["id"])
	}
	
	if _, hasURL := info["url"]; !hasURL {
		t.Error("Page info should have URL field")
	}
}

func TestGetAllPages(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	
	err := manager.Start(config)
	if err != nil {
		t.Skipf("Skipping browser test (no browser available): %v", err)
	}
	defer manager.Stop()
	
	// Initially no pages
	pages := manager.GetAllPages()
	if len(pages) != 0 {
		t.Errorf("Expected 0 pages, got %d", len(pages))
	}
	
	// Create a page
	_, pageID, err := manager.NewPage("")
	if err != nil {
		t.Fatalf("NewPage failed: %v", err)
	}
	defer manager.ClosePage(pageID)
	
	// Get all pages
	pages = manager.GetAllPages()
	if len(pages) != 1 {
		t.Errorf("Expected 1 page, got %d", len(pages))
	}
	
	if pages[0].PageID != pageID {
		t.Errorf("Expected page ID %s, got %s", pageID, pages[0].PageID)
	}
}

func TestGetCurrentPageID(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	
	// No pages initially
	currentID := manager.GetCurrentPageID()
	if currentID != "" {
		t.Errorf("Expected empty current page ID, got %s", currentID)
	}
	
	err := manager.Start(config)
	if err != nil {
		t.Skipf("Skipping browser test (no browser available): %v", err)
	}
	defer manager.Stop()
	
	// Create a page
	_, pageID, err := manager.NewPage("")
	if err != nil {
		t.Fatalf("NewPage failed: %v", err)
	}
	defer manager.ClosePage(pageID)
	
	// Should return the page ID
	currentID = manager.GetCurrentPageID()
	if currentID != pageID {
		t.Errorf("Expected current page ID %s, got %s", pageID, currentID)
	}
}

func TestSwitchToPage(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	
	// Test switching to nonexistent page
	err := manager.SwitchToPage("nonexistent")
	if err == nil {
		t.Error("Expected error for switching to nonexistent page")
	}
	
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// Benchmark tests
func BenchmarkNewManager(b *testing.B) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewManager(log, config)
		_ = manager
	}
}

func BenchmarkListPages(b *testing.B) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	err := manager.Start(config)
	if err != nil {
		b.Skipf("Skipping browser benchmark (no browser available): %v", err)
	}
	defer manager.Stop()
	
	// Create some pages
	for i := 0; i < 10; i++ {
		_, pageID, err := manager.NewPage("")
		if err != nil {
			b.Fatalf("NewPage failed: %v", err)
		}
		defer manager.ClosePage(pageID)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.ListPages()
	}
}

// Integration test with actual HTTP server
func TestNewPageWithHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true, WindowWidth: 800, WindowHeight: 600}
	
	manager := NewManager(log, config)
	
	err := manager.Start(config)
	if err != nil {
		t.Skipf("Skipping browser test (no browser available): %v", err)
	}
	defer manager.Stop()
	
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body><h1>Hello World</h1></body>
</html>`)
	}))
	defer server.Close()
	
	// Create page with URL
	_, pageID, err := manager.NewPage(server.URL)
	if err != nil {
		t.Fatalf("NewPage with URL failed: %v", err)
	}
	defer manager.ClosePage(pageID)
	
	// Get page info
	info, err := manager.GetPageInfo(pageID)
	if err != nil {
		t.Errorf("GetPageInfo failed: %v", err)
	}
	
	// Verify URL matches (allow for trailing slash normalization)
	actualURL, ok := info["url"].(string)
	if !ok {
		t.Errorf("URL should be a string, got %T", info["url"])
	} else if actualURL != server.URL && actualURL != server.URL+"/" {
		t.Errorf("Expected URL %s or %s/, got %s", server.URL, server.URL, actualURL)
	}
}

// Test environment variable browser override
func TestEnvironmentBrowserPath(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	config := Config{Headless: true}
	
	manager := NewManager(log, config)
	
	// Set environment variable to invalid path
	os.Setenv("RODMCP_BROWSER_PATH", "/nonexistent/browser")
	defer os.Unsetenv("RODMCP_BROWSER_PATH")
	
	_, err := manager.findWorkingBrowser()
	// Should fallback to other options when env var is invalid
	if err != nil {
		// This is expected - the function should try the env var and then fallback
		if !strings.Contains(err.Error(), "no working browser binary found") {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}