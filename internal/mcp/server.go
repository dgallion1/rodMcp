package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"rodmcp/internal/circuitbreaker"
	"rodmcp/internal/connection"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)


type Server struct {
	logger           *logger.Logger
	tools            map[string]Tool
	toolsMutex       sync.RWMutex
	initialized      bool
	version          types.MCPVersion
	info             types.ServerInfo
	ctx              context.Context
	cancel           context.CancelFunc
	connectionMgr    *connection.ConnectionManager
	circuitBreaker   *circuitbreaker.MultiLevelCircuitBreaker
	browserManager   BrowserHealthChecker // Interface for browser health checking
	lastActivity     time.Time            // Last activity timestamp for heartbeat monitoring
}

type Tool interface {
	Name() string
	Description() string
	InputSchema() types.ToolSchema
	Execute(args map[string]interface{}) (*types.CallToolResponse, error)
}

type BrowserHealthChecker interface {
	CheckHealth() error
	EnsureHealthy() error
}

func NewServer(log *logger.Logger) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Initialize connection manager with robust configuration
	connConfig := connection.DefaultConfig()
	connManager := connection.NewConnectionManager(log, connConfig)
	
	// Initialize circuit breakers for different operation types
	circuitBreaker := circuitbreaker.NewMultiLevelCircuitBreaker()
	
	server := &Server{
		logger:         log,
		tools:          make(map[string]Tool),
		version:        types.CurrentMCPVersion,
		info: types.ServerInfo{
			Name:    "rodmcp",
			Version: "1.0.0",
		},
		ctx:            ctx,
		cancel:         cancel,
		connectionMgr:  connManager,
		circuitBreaker: circuitBreaker,
		lastActivity:   time.Now(),
	}
	
	// Set up circuit breaker callbacks
	circuitBreaker.BrowserCircuitBreaker.CircuitBreaker.OnStateChange(func(from, to circuitbreaker.State) {
		log.WithComponent("circuit-breaker").Warn("Browser circuit breaker state changed",
			zap.String("from", from.String()),
			zap.String("to", to.String()))
	})
	
	circuitBreaker.NetworkCircuitBreaker.CircuitBreaker.OnStateChange(func(from, to circuitbreaker.State) {
		log.WithComponent("circuit-breaker").Warn("Network circuit breaker state changed",
			zap.String("from", from.String()),
			zap.String("to", to.String()))
	})
	
	return server
}


func (s *Server) RegisterTool(tool Tool) {
	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()
	s.tools[tool.Name()] = tool
	s.logger.WithComponent("mcp").Info("Tool registered",
		zap.String("tool", tool.Name()))
}

func (s *Server) SetBrowserManager(browserMgr BrowserHealthChecker) {
	s.browserManager = browserMgr
	s.logger.WithComponent("mcp").Info("Browser manager registered for health monitoring")
}

func (s *Server) Start() error {
	s.logger.WithComponent("mcp").Info("Starting MCP server with enhanced connection management",
		zap.String("version", string(s.version)))

	// Start connection manager
	if err := s.connectionMgr.Start(); err != nil {
		return fmt.Errorf("failed to start connection manager: %w", err)
	}

	// Start message reading with robust connection handling
	return s.startMessageLoop()
}

func (s *Server) startMessageLoop() error {
	s.logger.WithComponent("mcp").Info("Starting robust message loop with connection management")

	// Start health monitoring in background
	go s.startHealthMonitor()

	// Process messages with robust connection handling
	for {
		select {
		case <-s.ctx.Done():
			s.logger.WithComponent("mcp").Info("Server shutting down")
			return nil

		default:
			// Check if we're still connected before trying to read
			if !s.connectionMgr.IsConnected() {
				s.logger.WithComponent("mcp").Debug("Not connected - waiting for reconnection")
				time.Sleep(1 * time.Second)
				continue
			}

			// Read message using connection manager
			line, err := s.connectionMgr.ReadMessage()
			if err != nil {
				if err == io.EOF {
					s.logger.WithComponent("mcp").Info("Input stream closed (EOF) - shutting down server")
					// EOF means the client has disconnected, so we should exit gracefully
					return nil
				}
				
				// Check for "not connected" errors - pause processing
				if strings.Contains(err.Error(), "not connected") {
					s.logger.WithComponent("mcp").Debug("Connection lost - pausing message processing")
					time.Sleep(1 * time.Second)
					continue
				}
				
				// Check if this is a recoverable error
				if strings.Contains(err.Error(), "recoverable") {
					s.logger.WithComponent("mcp").Debug("Recoverable error - continuing operation", zap.Error(err))
					// Brief pause to prevent busy loop, but don't wait too long
					time.Sleep(50 * time.Millisecond)
					continue
				}
				
				// Check for timeout errors - these are also recoverable
				if strings.Contains(err.Error(), "timeout") {
					s.logger.WithComponent("mcp").Debug("Read timeout - continuing", zap.Error(err))
					time.Sleep(10 * time.Millisecond)
					continue
				}
				
				// Log other errors but don't exit - let connection manager handle recovery
				s.logger.WithComponent("mcp").Warn("Read message error - continuing with recovery",
					zap.Error(err))
				
				// Brief pause before retry to prevent busy loop
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if line == "" {
				continue
			}

			s.logger.WithComponent("mcp").Debug("Received message",
				zap.String("message", line))

			// Handle message with error recovery
			if err := s.handleMessage([]byte(line)); err != nil {
				s.logger.WithComponent("mcp").Error("Failed to handle message",
					zap.Error(err))
				// Don't exit on message handling errors, continue processing
			}
		}
	}
}

func (s *Server) startHealthMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Check browser health if we have a browser manager
			if s.browserManager != nil {
				err := s.circuitBreaker.ExecuteBrowserOperation(func() error {
					return s.browserManager.EnsureHealthy()
				})
				
				if err != nil {
					// Health check failures are handled by the circuit breaker
					// Log at debug level to avoid noise
					s.logger.WithComponent("mcp").Debug("Browser health check failed",
						zap.Error(err))
				}
			}
			
			// Log connection stats
			stats := s.connectionMgr.GetStats()
			s.logger.WithComponent("mcp").Debug("Connection health check",
				zap.Any("connection_stats", stats))
			
			// Log circuit breaker stats
			cbStats := s.circuitBreaker.GetOverallStats()
			s.logger.WithComponent("mcp").Debug("Circuit breaker status",
				zap.Any("circuit_breaker_stats", cbStats))
		}
	}
}

func (s *Server) handleMessage(data []byte) error {
	// Don't process messages if we're not connected
	if !s.connectionMgr.IsConnected() {
		s.logger.WithComponent("mcp").Debug("Ignoring message - not connected")
		return nil
	}

	var req types.JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return s.sendError(nil, -32700, "Parse error", nil)
	}

	s.logger.LogMCPRequest(req.Method, req.Params)

	switch req.Method {
	case "initialize":
		return s.handleInitialize(&req)
	case "tools/list":
		return s.handleToolsList(&req)
	case "tools/call":
		return s.handleToolsCall(&req)
	case "notifications/initialized":
		s.initialized = true
		s.logger.WithComponent("mcp").Info("Server initialized")
		return nil
	default:
		return s.sendError(req.ID, -32601, "Method not found", nil)
	}
}

func (s *Server) handleInitialize(req *types.JSONRPCRequest) error {
	var initReq types.InitializeRequest
	if req.Params != nil {
		params, _ := json.Marshal(req.Params)
		if err := json.Unmarshal(params, &initReq); err != nil {
			return s.sendError(req.ID, -32602, "Invalid params", nil)
		}
	}

	// Version negotiation
	if initReq.ProtocolVersion != s.version {
		s.logger.WithComponent("mcp").Warn("Protocol version mismatch",
			zap.String("client_version", string(initReq.ProtocolVersion)),
			zap.String("server_version", string(s.version)))
	}

	response := types.InitializeResponse{
		ProtocolVersion: s.version,
		Capabilities: types.ServerCapabilities{
			Tools:   &types.ToolsCapability{},
			Logging: &types.LoggingCapability{},
		},
		ServerInfo: s.info,
	}

	return s.sendResponse(req.ID, response)
}

func (s *Server) handleToolsList(req *types.JSONRPCRequest) error {
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

	return s.sendResponse(req.ID, result)
}

func (s *Server) handleToolsCall(req *types.JSONRPCRequest) error {
	// Validate connection before executing tools
	if !s.connectionMgr.IsConnected() {
		s.logger.WithComponent("mcp").Warn("Tool call attempted while disconnected", 
			zap.String("tool", "unknown"))
		return s.sendError(req.ID, -32001, "Server not connected", nil)
	}

	var callReq types.CallToolRequest
	if req.Params != nil {
		params, _ := json.Marshal(req.Params)
		if err := json.Unmarshal(params, &callReq); err != nil {
			return s.sendError(req.ID, -32602, "Invalid params", nil)
		}
	}

	// Validate connection again with tool name for better logging
	if !s.connectionMgr.IsConnected() {
		s.logger.WithComponent("mcp").Warn("Tool call attempted while disconnected", 
			zap.String("tool", callReq.Name))
		return s.sendError(req.ID, -32001, "Server not connected", nil)
	}

	s.toolsMutex.RLock()
	tool, exists := s.tools[callReq.Name]
	s.toolsMutex.RUnlock()

	if !exists {
		return s.sendError(req.ID, -32601, "Tool not found", nil)
	}

	s.logger.WithComponent("mcp").Debug("Executing tool", 
		zap.String("tool", callReq.Name))

	result, err := tool.Execute(callReq.Arguments)
	if err != nil {
		s.logger.LogMCPResponse(req.Method, nil, err)
		return s.sendError(req.ID, -32000, "Tool execution failed", err.Error())
	}

	s.logger.LogMCPResponse(req.Method, result, nil)
	return s.sendResponse(req.ID, result)
}

func (s *Server) sendResponse(id interface{}, result interface{}) error {
	response := types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	return s.writeMessage(response)
}

func (s *Server) sendError(id interface{}, code int, message string, data interface{}) error {
	response := types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &types.JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	return s.writeMessage(response)
}

func (s *Server) writeMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Use connection manager for robust message writing
	err = s.connectionMgr.WriteMessage(string(data))
	if err != nil {
		return err
	}

	s.logger.WithComponent("mcp").Debug("Sent message",
		zap.String("message", string(data)))

	return nil
}

func (s *Server) SendLogMessage(level string, message string, data map[string]interface{}) error {
	logData, _ := json.Marshal(data)

	logMsg := types.LoggingMessage{
		Level:  level,
		Data:   json.RawMessage(logData),
		Logger: "rodmcp",
	}

	notification := types.JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/message",
		Params:  logMsg,
	}

	return s.writeMessage(notification)
}

// updateActivity updates the last activity timestamp
func (s *Server) updateActivity() {
	s.lastActivity = time.Now()
}

// sendHeartbeat sends a heartbeat notification to the client
func (s *Server) sendHeartbeat() error {
	s.updateActivity()
	
	// Send a heartbeat notification
	heartbeat := types.JSONRPCRequest{
		JSONRPC: "2.0",
		Method: "notifications/heartbeat",
		Params: map[string]interface{}{
			"timestamp": s.lastActivity.Unix(),
		},
	}
	
	return s.writeMessage(heartbeat)
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	s.logger.WithComponent("mcp").Info("Stopping MCP server")
	
	// Stop connection manager first
	if err := s.connectionMgr.Stop(); err != nil {
		s.logger.WithComponent("mcp").Error("Error stopping connection manager", zap.Error(err))
	}
	
	s.cancel()
	return nil
}
