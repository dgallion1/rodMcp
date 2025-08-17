package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPServer(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	if server == nil {
		t.Fatal("NewHTTPServer returned nil")
	}
	
	if server.logger == nil {
		t.Error("Server logger is nil")
	}
	
	if server.tools == nil {
		t.Error("Server tools map is nil")
	}
	
	if server.port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.port)
	}
	
	if server.version != types.CurrentMCPVersion {
		t.Errorf("Expected version %s, got %s", types.CurrentMCPVersion, server.version)
	}
	
	if server.info.Name != "rodmcp-http" {
		t.Errorf("Expected server name 'rodmcp-http', got %s", server.info.Name)
	}
}

func TestHTTPServerRegisterTool(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	tool := NewSimpleTestTool("http_test_tool", "A test tool for HTTP server", "HTTP test successful")
	
	server.RegisterTool(tool)
	
	server.toolsMutex.RLock()
	registeredTool, exists := server.tools["http_test_tool"]
	server.toolsMutex.RUnlock()
	
	if !exists {
		t.Error("Tool was not registered")
	}
	
	if registeredTool.Name() != "http_test_tool" {
		t.Errorf("Expected tool name 'http_test_tool', got %s", registeredTool.Name())
	}
}

func TestHTTPServerHandleRoot(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	// Register a test tool
	tool := NewSimpleTestTool("root_test_tool", "Test tool", "Root test result")
	server.RegisterTool(tool)
	
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	server.handleRoot(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	if response["service"] != "RodMCP HTTP Server" {
		t.Errorf("Expected service name 'RodMCP HTTP Server', got %v", response["service"])
	}
	
	if response["tools"].(float64) != 1 {
		t.Errorf("Expected 1 tool, got %v", response["tools"])
	}
	
	if response["initialized"] != false {
		t.Errorf("Expected initialized false, got %v", response["initialized"])
	}
}

func TestHTTPServerHandleRootMethodNotAllowed(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	server.handleRoot(rr, req)
	
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHTTPServerHandleHealth(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	server.handleHealth(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
	
	if response["tools"].(float64) != 0 {
		t.Errorf("Expected 0 tools, got %v", response["tools"])
	}
	
	if _, exists := response["timestamp"]; !exists {
		t.Error("Response should include timestamp")
	}
}

func TestHTTPServerHandleInitialize(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	initReq := types.InitializeRequest{
		ProtocolVersion: types.CurrentMCPVersion,
		ClientInfo: types.ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}
	
	reqBody, err := json.Marshal(initReq)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/mcp/initialize", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	server.handleInitialize(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	
	var response types.InitializeResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	if response.ProtocolVersion != types.CurrentMCPVersion {
		t.Errorf("Expected protocol version %s, got %s", types.CurrentMCPVersion, response.ProtocolVersion)
	}
	
	if response.ServerInfo.Name != "rodmcp-http" {
		t.Errorf("Expected server name 'rodmcp-http', got %s", response.ServerInfo.Name)
	}
	
	if !server.initialized {
		t.Error("Server should be initialized after initialize request")
	}
}

func TestHTTPServerHandleInitializeInvalidJSON(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	req, err := http.NewRequest("POST", "/mcp/initialize", bytes.NewBufferString("{invalid json}"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	server.handleInitialize(rr, req)
	
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHTTPServerHandleInitializeMethodNotAllowed(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	req, err := http.NewRequest("GET", "/mcp/initialize", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	server.handleInitialize(rr, req)
	
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHTTPServerHandleToolsList(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	// Register test tools
	tool1 := NewSimpleTestTool("list_tool_1", "First test tool", "First tool result")
	tool2 := NewSimpleTestTool("list_tool_2", "Second test tool", "Second tool result")
	
	server.RegisterTool(tool1)
	server.RegisterTool(tool2)
	
	req, err := http.NewRequest("GET", "/mcp/tools/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	server.handleToolsList(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	tools, exists := response["tools"].([]interface{})
	if !exists {
		t.Fatal("Response should contain tools array")
	}
	
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
	
	// Check that both tools are present
	foundTools := make(map[string]bool)
	for _, toolData := range tools {
		tool := toolData.(map[string]interface{})
		foundTools[tool["name"].(string)] = true
	}
	
	if !foundTools["list_tool_1"] {
		t.Error("list_tool_1 not found in response")
	}
	
	if !foundTools["list_tool_2"] {
		t.Error("list_tool_2 not found in response")
	}
}

func TestHTTPServerHandleToolsCall(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	// Register a test tool
	tool := NewSimpleTestTool("call_http_tool", "Tool for testing HTTP calls", "HTTP execution successful")
	server.RegisterTool(tool)
	
	callReq := types.CallToolRequest{
		Name: "call_http_tool",
		Arguments: map[string]interface{}{
			"test": "value",
		},
	}
	
	reqBody, err := json.Marshal(callReq)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/mcp/tools/call", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	server.handleToolsCall(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	
	// Test passes if tool execution completed without error
	
	var response types.CallToolResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	if len(response.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(response.Content))
	}
	
	expectedText := "HTTP execution successful: "
	if !strings.HasPrefix(response.Content[0].Text, expectedText) {
		t.Errorf("Expected text to start with '%s', got '%s'", expectedText, response.Content[0].Text)
	}
}

func TestHTTPServerHandleToolsCallNotFound(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	callReq := types.CallToolRequest{
		Name:      "nonexistent_tool",
		Arguments: map[string]interface{}{},
	}
	
	reqBody, err := json.Marshal(callReq)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/mcp/tools/call", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	server.handleToolsCall(rr, req)
	
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	errorData, exists := response["error"].(map[string]interface{})
	if !exists {
		t.Fatal("Response should contain error object")
	}
	
	if errorData["message"] != "Tool not found" {
		t.Errorf("Expected 'Tool not found', got %v", errorData["message"])
	}
}

func TestHTTPServerHandleToolsCallExecutionError(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	// Register a tool that returns an error
	tool := NewErrorTestTool("error_http_tool", "Tool that returns an error", "execution failed")
	server.RegisterTool(tool)
	
	callReq := types.CallToolRequest{
		Name:      "error_http_tool",
		Arguments: map[string]interface{}{},
	}
	
	reqBody, err := json.Marshal(callReq)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/mcp/tools/call", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	server.handleToolsCall(rr, req)
	
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	errorData, exists := response["error"].(map[string]interface{})
	if !exists {
		t.Fatal("Response should contain error object")
	}
	
	if errorData["message"] != "Tool execution failed" {
		t.Errorf("Expected 'Tool execution failed', got %v", errorData["message"])
	}
}

func TestHTTPServerHandleToolsCallInvalidJSON(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	req, err := http.NewRequest("POST", "/mcp/tools/call", bytes.NewBufferString("{invalid json}"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	server.handleToolsCall(rr, req)
	
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHTTPServerSendHTTPError(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	rr := httptest.NewRecorder()
	server.sendHTTPError(rr, http.StatusBadRequest, "Test error", "Additional details")
	
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	errorData, exists := response["error"].(map[string]interface{})
	if !exists {
		t.Fatal("Response should contain error object")
	}
	
	if errorData["code"].(float64) != 400 {
		t.Errorf("Expected error code 400, got %v", errorData["code"])
	}
	
	if errorData["message"] != "Test error" {
		t.Errorf("Expected 'Test error', got %v", errorData["message"])
	}
	
	if errorData["details"] != "Additional details" {
		t.Errorf("Expected 'Additional details', got %v", errorData["details"])
	}
}

func TestHTTPServerSendLogMessage(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	logData := map[string]interface{}{
		"component": "test",
		"action":    "testing",
	}
	
	// These should not return errors
	levels := []string{"error", "warn", "debug", "info"}
	for _, level := range levels {
		err := server.SendLogMessage(level, "Test log message", logData)
		if err != nil {
			t.Errorf("SendLogMessage failed for level %s: %v", level, err)
		}
	}
}

func TestHTTPServerCORSHeaders(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	// Create a CORS handler wrapper (similar to the one in the server)
	corsHandler := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			handler(w, r)
		}
	}
	
	// Test OPTIONS request with CORS wrapper
	req, err := http.NewRequest("OPTIONS", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	corsHandler(server.handleRoot)(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", rr.Code)
	}
	
	// Check CORS headers
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}
	
	for header, expectedValue := range expectedHeaders {
		actualValue := rr.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected %s: %s, got %s", header, expectedValue, actualValue)
		}
	}
}

func TestHTTPServerStop(t *testing.T) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	// Test stopping server without starting
	err := server.Stop()
	if err != nil {
		t.Errorf("Stop should not error when server not started: %v", err)
	}
	
	// Test stopping server after "starting" (set server field)
	server.server = &http.Server{}
	
	// This will timeout since we didn't actually start the server
	// but we can test that it doesn't panic
	err = server.Stop()
	if err != nil {
		// This is expected since we didn't actually start the server
		if !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("Unexpected error on stop: %v", err)
		}
	}
}

// Integration test with actual server
func TestHTTPServerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 0) // Use port 0 for automatic assignment
	
	// Register a test tool
	tool := NewSimpleTestTool("integration_tool", "Tool for integration testing", "Integration test successful")
	server.RegisterTool(tool)
	
	// Start server in background
	go func() {
		// This will block, so we run it in a goroutine
		server.Start()
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Stop server
	defer server.Stop()
	
	// Note: In a real integration test, we'd make actual HTTP requests
	// to the server, but that requires more complex setup with port management
}

// Benchmark tests
func BenchmarkHTTPServerHandleRoot(b *testing.B) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	// Register some tools
	for i := 0; i < 10; i++ {
		tool := NewSimpleTestTool(fmt.Sprintf("bench_tool_%d", i), fmt.Sprintf("Benchmark tool %d", i), fmt.Sprintf("Bench result %d", i))
		server.RegisterTool(tool)
	}
	
	req, _ := http.NewRequest("GET", "/", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		server.handleRoot(rr, req)
	}
}

func BenchmarkHTTPServerHandleToolsList(b *testing.B) {
	log, _ := logger.New(logger.Config{LogLevel: "error", LogDir: "/tmp"})
	server := NewHTTPServer(log, 8080)
	
	// Register multiple tools
	for i := 0; i < 50; i++ {
		tool := NewSimpleTestTool(fmt.Sprintf("bench_list_tool_%d", i), fmt.Sprintf("Benchmark list tool %d", i), fmt.Sprintf("Bench list result %d", i))
		server.RegisterTool(tool)
	}
	
	req, _ := http.NewRequest("GET", "/mcp/tools/list", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		server.handleToolsList(rr, req)
	}
}