package browser

import (
	"context"
	"fmt"
	"testing"
	"time"

	"rodmcp/internal/logger"
)

func TestEnhancedManager_NewEnhancedManager(t *testing.T) {
	log := createTestLogger(t)
	config := Config{
		Headless:     true,
		Debug:        false,
		WindowWidth:  1200,
		WindowHeight: 800,
	}
	
	enhanced := NewEnhancedManager(log, config)
	if enhanced == nil {
		t.Fatal("NewEnhancedManager returned nil")
	}
	
	if enhanced.Manager == nil {
		t.Error("Enhanced manager should have underlying manager")
	}
}

func TestEnhancedManager_NewPageWithRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}
	
	log := createTestLogger(t)
	config := Config{
		Headless:     true,
		Debug:        false,
		WindowWidth:  1200,
		WindowHeight: 800,
	}
	
	enhanced := NewEnhancedManager(log, config)
	
	// Start the underlying manager
	if err := enhanced.Manager.Start(config); err != nil {
		t.Fatalf("Failed to start browser manager: %v", err)
	}
	defer enhanced.Manager.Stop()
	
	t.Run("ValidURL", func(t *testing.T) {
		// Test with a local file URL to avoid network dependencies
		page, pageID, err := enhanced.NewPageWithRetry("file:///home/darrell/work/git/rodMcp/test_data/simple_test.html")
		if err != nil {
			t.Errorf("NewPageWithRetry failed: %v", err)
		}
		if page == nil {
			t.Error("NewPageWithRetry returned nil page")
		}
		if pageID == "" {
			t.Error("NewPageWithRetry returned empty pageID")
		}
		
		// Clean up
		if pageID != "" {
			enhanced.Manager.ClosePage(pageID)
		}
	})
	
	t.Run("InvalidURL", func(t *testing.T) {
		page, pageID, err := enhanced.NewPageWithRetry("invalid://url")
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
		if page != nil {
			t.Error("Expected nil page for invalid URL")
			enhanced.Manager.ClosePage(pageID)
		}
	})
}

func TestEnhancedManager_NavigateWithRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}
	
	log := createTestLogger(t)
	config := Config{
		Headless:     true,
		Debug:        false,
		WindowWidth:  1200,
		WindowHeight: 800,
	}
	
	enhanced := NewEnhancedManager(log, config)
	
	if err := enhanced.Manager.Start(config); err != nil {
		t.Fatalf("Failed to start browser manager: %v", err)
	}
	defer enhanced.Manager.Stop()
	
	// Create a page first
	_, pageID, err := enhanced.Manager.NewPage("")
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}
	defer enhanced.Manager.ClosePage(pageID)
	
	t.Run("ValidNavigation", func(t *testing.T) {
		err := enhanced.NavigateWithRetry(pageID, "file:///home/darrell/work/git/rodMcp/test_data/navigate_test.html")
		if err != nil {
			t.Errorf("NavigateWithRetry failed: %v", err)
		}
	})
	
	t.Run("InvalidPageID", func(t *testing.T) {
		err := enhanced.NavigateWithRetry("invalid-page-id", "file:///home/darrell/work/git/rodMcp/test_data/simple_test.html")
		if err == nil {
			t.Error("Expected error for invalid page ID")
		}
	})
}

func TestEnhancedManager_IsRecoverableError(t *testing.T) {
	log := createTestLogger(t)
	config := Config{Headless: true}
	enhanced := NewEnhancedManager(log, config)
	
	testCases := []struct {
		name       string
		err        error
		recoverable bool
	}{
		{
			name:       "ContextCanceled",
			err:        context.Canceled,
			recoverable: true,
		},
		{
			name:       "ContextTimeout",
			err:        context.DeadlineExceeded,
			recoverable: true,
		},
		{
			name:       "ConnectionError",
			err:        fmt.Errorf("connection failed"),
			recoverable: true,
		},
		{
			name:       "BrowserError",
			err:        fmt.Errorf("browser process died"),
			recoverable: true,
		},
		{
			name:       "ValidationError",
			err:        fmt.Errorf("invalid selector"),
			recoverable: false,
		},
		{
			name:       "NilError",
			err:        nil,
			recoverable: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := enhanced.isRecoverableError(tc.err)
			if result != tc.recoverable {
				t.Errorf("isRecoverableError(%v) = %v, want %v", tc.err, result, tc.recoverable)
			}
		})
	}
}

func TestEnhancedManager_IsContextError(t *testing.T) {
	log := createTestLogger(t)
	config := Config{Headless: true}
	enhanced := NewEnhancedManager(log, config)
	
	testCases := []struct {
		name      string
		err       error
		isContext bool
	}{
		{
			name:      "ContextCanceled",
			err:       context.Canceled,
			isContext: true,
		},
		{
			name:      "ContextTimeout",
			err:       context.DeadlineExceeded,
			isContext: true,
		},
		{
			name:      "ContextInMessage",
			err:       fmt.Errorf("operation failed: context canceled"),
			isContext: true,
		},
		{
			name:      "ContextDeadlineInMessage",
			err:       fmt.Errorf("operation failed: context deadline exceeded"),
			isContext: true,
		},
		{
			name:      "RegularError",
			err:       fmt.Errorf("some other error"),
			isContext: false,
		},
		{
			name:      "NilError",
			err:       nil,
			isContext: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := enhanced.isContextError(tc.err)
			if result != tc.isContext {
				t.Errorf("isContextError(%v) = %v, want %v", tc.err, result, tc.isContext)
			}
		})
	}
}

func TestEnhancedManager_CalculateRestartBackoff(t *testing.T) {
	log := createTestLogger(t)
	config := Config{Headless: true}
	enhanced := NewEnhancedManager(log, config)
	
	testCases := []struct {
		name        string
		attempt     int
		expectMin   time.Duration
		expectMax   time.Duration
	}{
		{
			name:      "FirstAttempt",
			attempt:   1,
			expectMin: 1 * time.Second,
			expectMax: 3 * time.Second,
		},
		{
			name:      "SecondAttempt", 
			attempt:   2,
			expectMin: 2 * time.Second,
			expectMax: 6 * time.Second,
		},
		{
			name:      "ThirdAttempt",
			attempt:   3,
			expectMin: 4 * time.Second,
			expectMax: 12 * time.Second,
		},
		{
			name:      "HighAttempt",
			attempt:   10,
			expectMin: 10 * time.Second,
			expectMax: 30 * time.Second, // Updated to match actual restartBackoffMax
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate setting restart attempts and calculate backoff
			enhanced.browserRestartAttempts = tc.attempt
			backoff := enhanced.calculateRestartBackoff()
			if backoff < tc.expectMin {
				t.Errorf("Backoff %v is less than expected minimum %v", backoff, tc.expectMin)
			}
			if backoff > tc.expectMax {
				t.Errorf("Backoff %v is greater than expected maximum %v", backoff, tc.expectMax)
			}
		})
	}
}

// Helper function to create test logger
func createTestLogger(t *testing.T) *logger.Logger {
	log, err := logger.New(logger.Config{
		LogLevel:    "error", // Reduce noise in tests
		LogDir:      t.TempDir(),
		Development: true,
	})
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}
	return log
}