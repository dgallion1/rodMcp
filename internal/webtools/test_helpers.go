package webtools

import (
	"testing"

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