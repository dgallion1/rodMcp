package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type HTTPServer struct {
	logger      *logger.Logger
	tools       map[string]Tool
	toolsMutex  sync.RWMutex
	server      *http.Server
	initialized bool
	version     types.MCPVersion
	info        types.ServerInfo
	port        int
}

// NewHTTPServer creates a new HTTP-based MCP server
func NewHTTPServer(log *logger.Logger, port int) *HTTPServer {
	return &HTTPServer{
		logger: log,
		tools:  make(map[string]Tool),
		version: types.CurrentMCPVersion,
		info: types.ServerInfo{
			Name:    "rodmcp-http",
			Version: "1.0.0",
		},
		port: port,
	}
}

func (s *HTTPServer) RegisterTool(tool Tool) {
	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()
	s.tools[tool.Name()] = tool
	s.logger.WithComponent("http-mcp").Info("Tool registered",
		zap.String("tool", tool.Name()))
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()
	
	// CORS middleware
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
	
	// MCP endpoints
	mux.HandleFunc("/mcp/initialize", corsHandler(s.handleInitialize))
	mux.HandleFunc("/mcp/tools/list", corsHandler(s.handleToolsList))
	mux.HandleFunc("/mcp/tools/call", corsHandler(s.handleToolsCall))
	mux.HandleFunc("/health", corsHandler(s.handleHealth))
	
	// Server info endpoint
	mux.HandleFunc("/", corsHandler(s.handleRoot))

	s.server = &http.Server{
		Addr:         ":" + strconv.Itoa(s.port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.WithComponent("http-mcp").Info("Starting HTTP MCP server",
		zap.Int("port", s.port),
		zap.String("version", string(s.version)))

	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop() error {
	if s.server == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	s.logger.WithComponent("http-mcp").Info("Shutting down HTTP MCP server")
	return s.server.Shutdown(ctx)
}

func (s *HTTPServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.toolsMutex.RLock()
	toolCount := len(s.tools)
	s.toolsMutex.RUnlock()
	
	response := map[string]interface{}{
		"service":     "RodMCP HTTP Server",
		"version":     s.info.Version,
		"protocol":    s.version,
		"tools":       toolCount,
		"initialized": s.initialized,
		"endpoints": map[string]string{
			"initialize":  "/mcp/initialize",
			"tools_list":  "/mcp/tools/list", 
			"tools_call":  "/mcp/tools/call",
			"health":      "/health",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.toolsMutex.RLock()
	toolCount := len(s.tools)
	s.toolsMutex.RUnlock()
	
	health := map[string]interface{}{
		"status":      "healthy",
		"tools":       toolCount,
		"initialized": s.initialized,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *HTTPServer) handleInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var initReq types.InitializeRequest
	if err := json.NewDecoder(r.Body).Decode(&initReq); err != nil {
		s.sendHTTPError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}
	
	// Version negotiation
	if initReq.ProtocolVersion != s.version {
		s.logger.WithComponent("http-mcp").Warn("Protocol version mismatch",
			zap.String("client_version", string(initReq.ProtocolVersion)),
			zap.String("server_version", string(s.version)))
	}
	
	s.initialized = true
	
	response := types.InitializeResponse{
		ProtocolVersion: s.version,
		Capabilities: types.ServerCapabilities{
			Tools:   &types.ToolsCapability{},
			Logging: &types.LoggingCapability{},
		},
		ServerInfo: s.info,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.toolsMutex.RLock()
	defer s.toolsMutex.RUnlock()
	
	var tools []types.Tool
	for _, tool := range s.tools {
		tools = append(tools, types.Tool{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		})
	}
	
	result := map[string]interface{}{
		"tools": tools,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) handleToolsCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var callReq types.CallToolRequest
	if err := json.NewDecoder(r.Body).Decode(&callReq); err != nil {
		s.sendHTTPError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}
	
	s.toolsMutex.RLock()
	tool, exists := s.tools[callReq.Name]
	s.toolsMutex.RUnlock()
	
	if !exists {
		s.sendHTTPError(w, http.StatusNotFound, "Tool not found", fmt.Sprintf("Tool '%s' is not available", callReq.Name))
		return
	}
	
	// Log the tool execution attempt
	s.logger.WithComponent("http-mcp").Info("Executing tool",
		zap.String("tool", callReq.Name),
		zap.Any("args", callReq.Arguments))
	
	result, err := tool.Execute(callReq.Arguments)
	if err != nil {
		s.logger.WithComponent("http-mcp").Error("Tool execution failed",
			zap.String("tool", callReq.Name),
			zap.Error(err))
		s.sendHTTPError(w, http.StatusInternalServerError, "Tool execution failed", err.Error())
		return
	}
	
	s.logger.WithComponent("http-mcp").Info("Tool executed successfully",
		zap.String("tool", callReq.Name))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) sendHTTPError(w http.ResponseWriter, statusCode int, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    statusCode,
			"message": message,
			"details": details,
		},
	}
	
	json.NewEncoder(w).Encode(errorResponse)
}

// SendLogMessage sends a log message (for HTTP, we just log it internally)
func (s *HTTPServer) SendLogMessage(level string, message string, data map[string]interface{}) error {
	switch level {
	case "error":
		s.logger.WithComponent("http-mcp").Error(message, zap.Any("data", data))
	case "warn":
		s.logger.WithComponent("http-mcp").Warn(message, zap.Any("data", data))
	case "debug":
		s.logger.WithComponent("http-mcp").Debug(message, zap.Any("data", data))
	default:
		s.logger.WithComponent("http-mcp").Info(message, zap.Any("data", data))
	}
	return nil
}