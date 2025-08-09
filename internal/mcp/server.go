package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	logger         *logger.Logger
	tools          map[string]Tool
	toolsMutex     sync.RWMutex
	initialized    bool
	version        types.MCPVersion
	info           types.ServerInfo
	ctx            context.Context
	cancel         context.CancelFunc
	lastActivity   time.Time
	activityMutex  sync.RWMutex
	heartbeatChan  chan struct{}
	browserManager BrowserHealthChecker // Interface for browser health checking
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
	return &Server{
		logger:        log,
		tools:         make(map[string]Tool),
		version:       types.CurrentMCPVersion,
		info: types.ServerInfo{
			Name:    "rodmcp",
			Version: "1.0.0",
		},
		ctx:           ctx,
		cancel:        cancel,
		lastActivity:  time.Now(),
		heartbeatChan: make(chan struct{}, 1),
	}
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
	s.logger.WithComponent("mcp").Info("Starting MCP server with connection management",
		zap.String("version", string(s.version)))

	// Start connection monitoring in background
	go s.startConnectionMonitor()

	// Start message reading with timeout handling
	return s.startMessageLoop()
}

func (s *Server) startConnectionMonitor() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.activityMutex.RLock()
			lastActivity := s.lastActivity
			s.activityMutex.RUnlock()

			// If no activity for 5 minutes, send heartbeat
			if time.Since(lastActivity) > 5*time.Minute {
				s.logger.WithComponent("mcp").Debug("Sending connection heartbeat")
				if err := s.sendHeartbeat(); err != nil {
					s.logger.WithComponent("mcp").Warn("Heartbeat failed",
						zap.Error(err))
				}
			}
			
			// Check browser health if we have a browser manager
			if s.browserManager != nil {
				if err := s.browserManager.EnsureHealthy(); err != nil {
					s.logger.WithComponent("mcp").Error("Browser health check failed",
						zap.Error(err))
				}
			}
			
			// If no activity for 10 minutes, log warning
			if time.Since(lastActivity) > 10*time.Minute {
				s.logger.WithComponent("mcp").Warn("No client activity detected",
					zap.Duration("idle_time", time.Since(lastActivity)))
			}
		}
	}
}

func (s *Server) startMessageLoop() error {
	scanner := bufio.NewScanner(os.Stdin)
	
	// Set a reasonable buffer size for large messages
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	inputChan := make(chan string, 10)
	errorChan := make(chan error, 1)

	// Read input in a separate goroutine to enable timeout handling
	go func() {
		defer close(inputChan)
		for scanner.Scan() {
			select {
			case inputChan <- scanner.Text():
			case <-s.ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("scanner error: %w", err)
		}
	}()

	// Process messages with timeout handling
	for {
		select {
		case <-s.ctx.Done():
			s.logger.WithComponent("mcp").Info("Server shutting down")
			return nil

		case err := <-errorChan:
			s.logger.WithComponent("mcp").Error("Input error", zap.Error(err))
			return err

		case line, ok := <-inputChan:
			if !ok {
				// Channel closed, likely EOF from stdin
				s.logger.WithComponent("mcp").Info("Input stream closed")
				return nil
			}

			if line == "" {
				continue
			}

			// Update activity timestamp
			s.updateActivity()

			s.logger.WithComponent("mcp").Debug("Received message",
				zap.String("message", line))

			if err := s.handleMessage([]byte(line)); err != nil {
				s.logger.WithComponent("mcp").Error("Failed to handle message",
					zap.Error(err))
				// Don't exit on message handling errors, continue processing
			}

		case <-time.After(1 * time.Minute):
			// Periodic check - not strictly necessary but helps with debugging
			s.logger.WithComponent("mcp").Debug("Server alive check")
		}
	}
}

func (s *Server) handleMessage(data []byte) error {
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
	var callReq types.CallToolRequest
	if req.Params != nil {
		params, _ := json.Marshal(req.Params)
		if err := json.Unmarshal(params, &callReq); err != nil {
			return s.sendError(req.ID, -32602, "Invalid params", nil)
		}
	}

	s.toolsMutex.RLock()
	tool, exists := s.tools[callReq.Name]
	s.toolsMutex.RUnlock()

	if !exists {
		return s.sendError(req.ID, -32601, "Tool not found", nil)
	}

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

	_, err = io.WriteString(os.Stdout, string(data)+"\n")
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
	s.activityMutex.Lock()
	defer s.activityMutex.Unlock()
	s.lastActivity = time.Now()
}

// sendHeartbeat sends a heartbeat ping to check connection health
func (s *Server) sendHeartbeat() error {
	ping := types.JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/ping",
		Params: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}

	return s.writeMessage(ping)
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	s.logger.WithComponent("mcp").Info("Stopping MCP server")
	s.cancel()
	return nil
}
