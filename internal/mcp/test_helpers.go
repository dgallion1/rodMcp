package mcp

import (
	"rodmcp/internal/logger"
	"testing"
	"time"
)

// TestHelper provides utilities for setting up and tearing down test resources
type TestHelper struct {
	logger         *logger.Logger
	browserManager *TestBrowserManager
	server         *Server
	started        bool
}

func NewTestHelper(t *testing.T) *TestHelper {
	log, err := logger.New(logger.Config{
		LogLevel: "error",
		LogDir:   "/tmp",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	return &TestHelper{
		logger: log,
	}
}

// SetupServer creates and configures a test server
func (h *TestHelper) SetupServer(t *testing.T) *Server {
	h.server = NewServer(h.logger)
	
	// Start the connection manager for testing
	if err := h.server.connectionMgr.Start(); err != nil {
		t.Fatalf("Failed to start connection manager: %v", err)
	}
	
	return h.server
}

// SetupBrowserManager creates a test browser manager
func (h *TestHelper) SetupBrowserManager(t *testing.T, startBrowser bool) *TestBrowserManager {
	h.browserManager = NewTestBrowserManager(h.logger)
	
	if startBrowser {
		if err := h.browserManager.Start(); err != nil {
			t.Fatalf("Failed to start test browser: %v", err)
		}
		h.started = true
		
		// Wait for browser to be ready
		if err := h.browserManager.WaitReady(5 * time.Second); err != nil {
			t.Fatalf("Browser did not become ready: %v", err)
		}
	}
	
	return h.browserManager
}

// SetupServerWithBrowser creates both server and browser manager
func (h *TestHelper) SetupServerWithBrowser(t *testing.T, startBrowser bool) (*Server, *TestBrowserManager) {
	server := h.SetupServer(t)
	browserMgr := h.SetupBrowserManager(t, startBrowser)
	
	server.SetBrowserManager(browserMgr)
	
	return server, browserMgr
}

// Cleanup tears down all test resources
func (h *TestHelper) Cleanup(t *testing.T) {
	if h.server != nil {
		if err := h.server.Stop(); err != nil {
			t.Logf("Warning: Failed to stop server: %v", err)
		}
		
		// Stop connection manager
		h.server.connectionMgr.Stop()
	}
	
	if h.browserManager != nil && h.started {
		if err := h.browserManager.Stop(); err != nil {
			t.Logf("Warning: Failed to stop browser: %v", err)
		}
	}
}

// RegisterTestTools adds a set of common test tools to the server
func (h *TestHelper) RegisterTestTools(server *Server) {
	// Register a variety of test tools
	server.RegisterTool(NewSimpleTestTool("test_tool", "A basic test tool", "Test result"))
	server.RegisterTool(NewSimpleTestTool("param_tool", "Tool with parameters", "Parameter result"))
	server.RegisterTool(NewErrorTestTool("error_tool", "Tool that fails", "test error"))
	server.RegisterTool(NewTestHelpTool(h.logger))
}

// CreateTestServer is a convenience function for simple test setups
func CreateTestServer(t *testing.T) (*Server, *TestHelper) {
	helper := NewTestHelper(t)
	server := helper.SetupServer(t)
	helper.RegisterTestTools(server)
	
	t.Cleanup(func() {
		helper.Cleanup(t)
	})
	
	return server, helper
}

// CreateTestServerWithBrowser is a convenience function for browser-enabled tests
func CreateTestServerWithBrowser(t *testing.T, startBrowser bool) (*Server, *TestBrowserManager, *TestHelper) {
	helper := NewTestHelper(t)
	server, browserMgr := helper.SetupServerWithBrowser(t, startBrowser)
	helper.RegisterTestTools(server)
	
	t.Cleanup(func() {
		helper.Cleanup(t)
	})
	
	return server, browserMgr, helper
}