package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
	"sync"

	"go.uber.org/zap"
)

type Server struct {
	logger      *logger.Logger
	tools       map[string]Tool
	toolsMutex  sync.RWMutex
	initialized bool
	version     types.MCPVersion
	info        types.ServerInfo
}

type Tool interface {
	Name() string
	Description() string
	InputSchema() types.ToolSchema
	Execute(args map[string]interface{}) (*types.CallToolResponse, error)
}

func NewServer(log *logger.Logger) *Server {
	return &Server{
		logger:  log,
		tools:   make(map[string]Tool),
		version: types.CurrentMCPVersion,
		info: types.ServerInfo{
			Name:    "rodmcp",
			Version: "1.0.0",
		},
	}
}

func (s *Server) RegisterTool(tool Tool) {
	s.toolsMutex.Lock()
	defer s.toolsMutex.Unlock()
	s.tools[tool.Name()] = tool
	s.logger.WithComponent("mcp").Info("Tool registered", 
		zap.String("tool", tool.Name()))
}

func (s *Server) Start() error {
	s.logger.WithComponent("mcp").Info("Starting MCP server", 
		zap.String("version", string(s.version)))

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		s.logger.WithComponent("mcp").Debug("Received message", 
			zap.String("message", line))

		if err := s.handleMessage([]byte(line)); err != nil {
			s.logger.WithComponent("mcp").Error("Failed to handle message", 
				zap.Error(err))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from stdin: %w", err)
	}

	return nil
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
			Tools: &types.ToolsCapability{},
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