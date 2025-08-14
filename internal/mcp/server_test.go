package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
	"sync"
	"testing"
	"time"
)

// Mock tool for testing
type mockTool struct {
	name        string
	description string
	schema      types.ToolSchema
	executeFunc func(args map[string]interface{}) (*types.CallToolResponse, error)
}

func (m *mockTool) Name() string { return m.name }
func (m *mockTool) Description() string { return m.description }
func (m *mockTool) InputSchema() types.ToolSchema { return m.schema }
func (m *mockTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	if m.executeFunc != nil {
		return m.executeFunc(args)
	}
	return &types.CallToolResponse{
		Content: []types.ToolContent{
			{
				Type: "text",
				Text: "Mock execution successful",
			},
		},
	}, nil
}

// Mock browser manager for testing
type mockBrowserManager struct {
	healthError      error
	ensureHealthFunc func() error
	callCount        int
	mutex            sync.Mutex
}

func (m *mockBrowserManager) CheckHealth() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.callCount++
	return m.healthError
}

func (m *mockBrowserManager) EnsureHealthy() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.callCount++
	if m.ensureHealthFunc != nil {
		return m.ensureHealthFunc()
	}
	return m.healthError
}

func (m *mockBrowserManager) GetCallCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.callCount
}

func TestNewServer(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	
	if server.logger == nil {
		t.Error("Server logger is nil")
	}
	
	if server.tools == nil {
		t.Error("Server tools map is nil")
	}
	
	if server.version != types.CurrentMCPVersion {
		t.Errorf("Expected version %s, got %s", types.CurrentMCPVersion, server.version)
	}
	
	if server.info.Name != "rodmcp" {
		t.Errorf("Expected server name 'rodmcp', got %s", server.info.Name)
	}
	
	if server.ctx == nil {
		t.Error("Server context is nil")
	}
}

func TestRegisterTool(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	tool := &mockTool{
		name:        "test_tool",
		description: "A test tool",
		schema: types.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"message": map[string]interface{}{"type": "string", "description": "Test message"},
			},
			Required: []string{"message"},
		},
	}
	
	server.RegisterTool(tool)
	
	server.toolsMutex.RLock()
	registeredTool, exists := server.tools["test_tool"]
	server.toolsMutex.RUnlock()
	
	if !exists {
		t.Error("Tool was not registered")
	}
	
	if registeredTool.Name() != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got %s", registeredTool.Name())
	}
}

func TestSetBrowserManager(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	browserMgr := &mockBrowserManager{}
	server.SetBrowserManager(browserMgr)
	
	if server.browserManager == nil {
		t.Error("Browser manager was not set")
	}
}

func TestHandleInitializeMessage(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// Create initialize request
	initReq := types.InitializeRequest{
		ProtocolVersion: types.CurrentMCPVersion,
		ClientInfo: types.ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}
	
	reqData := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  initReq,
	}
	
	err := server.handleInitialize(&reqData)
	if err != nil {
		t.Errorf("handleInitialize failed: %v", err)
	}
}

func TestHandleToolsList(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// Register a test tool
	tool := &mockTool{
		name:        "list_test_tool",
		description: "Tool for testing list functionality",
		schema: types.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"param": map[string]interface{}{"type": "string", "description": "Test parameter"},
			},
		},
	}
	server.RegisterTool(tool)
	
	// Create tools/list request
	reqData := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	
	err := server.handleToolsList(&reqData)
	if err != nil {
		t.Errorf("handleToolsList failed: %v", err)
	}
}

func TestHandleToolsCall(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// Register a test tool with custom execution
	executed := false
	tool := &mockTool{
		name:        "call_test_tool",
		description: "Tool for testing call functionality",
		schema: types.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"message": map[string]interface{}{"type": "string", "description": "Message parameter"},
			},
		},
		executeFunc: func(args map[string]interface{}) (*types.CallToolResponse, error) {
			executed = true
			return &types.CallToolResponse{
				Content: []types.ToolContent{
					{
						Type: "text",
						Text: "Custom execution result",
					},
				},
			}, nil
		},
	}
	server.RegisterTool(tool)
	
	// Create tools/call request
	callReq := types.CallToolRequest{
		Name: "call_test_tool",
		Arguments: map[string]interface{}{
			"message": "test message",
		},
	}
	
	reqData := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  callReq,
	}
	
	err := server.handleToolsCall(&reqData)
	if err != nil {
		t.Errorf("handleToolsCall failed: %v", err)
	}
	
	if !executed {
		t.Error("Tool execute function was not called")
	}
}

func TestHandleToolsCallNotFound(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// Create tools/call request for non-existent tool
	callReq := types.CallToolRequest{
		Name: "nonexistent_tool",
		Arguments: map[string]interface{}{
			"param": "value",
		},
	}
	
	reqData := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  callReq,
	}
	
	err := server.handleToolsCall(&reqData)
	// Should not return error (error is sent as JSON-RPC error response)
	if err != nil {
		t.Errorf("handleToolsCall should not return error for tool not found: %v", err)
	}
}

func TestHandleToolsCallExecutionError(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// Register a tool that returns an error
	tool := &mockTool{
		name:        "error_tool",
		description: "Tool that returns an error",
		executeFunc: func(args map[string]interface{}) (*types.CallToolResponse, error) {
			return nil, errors.New("execution failed")
		},
	}
	server.RegisterTool(tool)
	
	// Create tools/call request
	callReq := types.CallToolRequest{
		Name:      "error_tool",
		Arguments: map[string]interface{}{},
	}
	
	reqData := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "tools/call",
		Params:  callReq,
	}
	
	err := server.handleToolsCall(&reqData)
	// Should not return error (error is sent as JSON-RPC error response)
	if err != nil {
		t.Errorf("handleToolsCall should not return error for execution failure: %v", err)
	}
}

func TestHandleMessage(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// Test initialize message
	initMsg := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: types.InitializeRequest{
			ProtocolVersion: types.CurrentMCPVersion,
		},
	}
	
	data, err := json.Marshal(initMsg)
	if err != nil {
		t.Fatalf("Failed to marshal test message: %v", err)
	}
	
	err = server.handleMessage(data)
	if err != nil {
		t.Errorf("handleMessage failed: %v", err)
	}
}

func TestHandleMessageInvalidJSON(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// Test with invalid JSON
	invalidJSON := []byte(`{"invalid": json}`)
	
	err := server.handleMessage(invalidJSON)
	// Should not return error (error is sent as JSON-RPC error response)
	if err != nil {
		t.Errorf("handleMessage should not return error for invalid JSON: %v", err)
	}
}

func TestHandleMessageUnknownMethod(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// Test with unknown method
	unknownMsg := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
	}
	
	data, err := json.Marshal(unknownMsg)
	if err != nil {
		t.Fatalf("Failed to marshal test message: %v", err)
	}
	
	err = server.handleMessage(data)
	// Should not return error (error is sent as JSON-RPC error response)
	if err != nil {
		t.Errorf("handleMessage should not return error for unknown method: %v", err)
	}
}

func TestHandleNotificationsInitialized(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	if server.initialized {
		t.Error("Server should not be initialized initially")
	}
	
	// Test notifications/initialized message
	initNotification := types.JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	
	data, err := json.Marshal(initNotification)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}
	
	err = server.handleMessage(data)
	if err != nil {
		t.Errorf("handleMessage failed for initialized notification: %v", err)
	}
	
	if !server.initialized {
		t.Error("Server should be initialized after notification")
	}
}

func TestSendResponse(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	result := map[string]string{"status": "success"}
	
	// This would normally write to stdout, but we can't easily capture that in tests
	// We're just testing that it doesn't panic or return an error
	err := server.sendResponse(1, result)
	if err != nil {
		t.Errorf("sendResponse failed: %v", err)
	}
}

func TestSendError(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	err := server.sendError(1, -32000, "Test error", "Additional data")
	if err != nil {
		t.Errorf("sendError failed: %v", err)
	}
}

func TestSendLogMessage(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	logData := map[string]interface{}{
		"component": "test",
		"action":    "testing",
	}
	
	err := server.SendLogMessage("info", "Test log message", logData)
	if err != nil {
		t.Errorf("SendLogMessage failed: %v", err)
	}
}

func TestUpdateActivity(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	initialTime := server.lastActivity
	
	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)
	
	server.updateActivity()
	
	if !server.lastActivity.After(initialTime) {
		t.Error("Activity timestamp should be updated")
	}
}

func TestSendHeartbeat(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// This would normally write to stdout
	err := server.sendHeartbeat()
	if err != nil {
		t.Errorf("sendHeartbeat failed: %v", err)
	}
}

func TestStop(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	err := server.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
	
	// Check that context is cancelled
	select {
	case <-server.ctx.Done():
		// Good, context was cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after Stop()")
	}
}

// Test browser health checking integration  
func TestBrowserHealthChecking(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	browserMgr := &mockBrowserManager{
		healthError: nil,
	}
	server.SetBrowserManager(browserMgr)
	
	// Just test that the browser manager is set
	if server.browserManager == nil {
		t.Error("Browser manager should be set")
	}
	
	// Stop server immediately
	server.Stop()
}

func TestBrowserHealthCheckingWithError(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	expectedError := errors.New("browser unhealthy")
	browserMgr := &mockBrowserManager{
		healthError: expectedError,
		ensureHealthFunc: func() error {
			return expectedError
		},
	}
	server.SetBrowserManager(browserMgr)
	
	// The connection monitor would log the error but continue running
	// We can't easily test the periodic behavior without advanced time control
	if server.browserManager == nil {
		t.Error("Browser manager should be set")
	}
}

// Benchmark tests
func BenchmarkNewServer(b *testing.B) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server := NewServer(log)
		_ = server
	}
}

func BenchmarkRegisterTool(b *testing.B) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	tool := &mockTool{
		name:        "benchmark_tool",
		description: "Tool for benchmarking",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.RegisterTool(tool)
	}
}

func BenchmarkHandleToolsList(b *testing.B) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	// Register multiple tools
	for i := 0; i < 10; i++ {
		tool := &mockTool{
			name:        fmt.Sprintf("tool_%d", i),
			description: fmt.Sprintf("Tool number %d", i),
		}
		server.RegisterTool(tool)
	}
	
	reqData := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.handleToolsList(&reqData)
	}
}

func BenchmarkHandleToolsCall(b *testing.B) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewServer(log)
	
	tool := &mockTool{
		name:        "benchmark_call_tool",
		description: "Tool for benchmarking calls",
		executeFunc: func(args map[string]interface{}) (*types.CallToolResponse, error) {
			return &types.CallToolResponse{
				Content: []types.ToolContent{
					{Type: "text", Text: "benchmark result"},
				},
			}, nil
		},
	}
	server.RegisterTool(tool)
	
	callReq := types.CallToolRequest{
		Name:      "benchmark_call_tool",
		Arguments: map[string]interface{}{"test": "value"},
	}
	
	reqData := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  callReq,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.handleToolsCall(&reqData)
	}
}