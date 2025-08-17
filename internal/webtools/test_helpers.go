package webtools

import (
	"context"
	"testing"
	"time"

	"rodmcp/internal/logger"
)

// Helper function to create test logger
func createTestLogger(t *testing.T) *logger.Logger {
	log, err := logger.New(logger.Config{
		LogLevel:    "info",
		LogDir:      t.TempDir(),
		Development: true,
	})
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}
	return log
}

// Helper function to create test context with timeout
func createTestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// Helper function to check if error is due to context cancellation
func isContextCancelledError(err error) bool {
	if err == nil {
		return false
	}
	return err == context.Canceled || err == context.DeadlineExceeded
}