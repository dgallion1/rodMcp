package types

import "encoding/json"

// MCP Protocol types based on 2025-06-18 specification

type MCPVersion string

const (
	CurrentMCPVersion MCPVersion = "2025-06-18"
)

// JSON-RPC 2.0 base structures
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP specific message types
type InitializeRequest struct {
	ProtocolVersion MCPVersion      `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo      `json:"clientInfo"`
}

type InitializeResponse struct {
	ProtocolVersion MCPVersion        `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo        `json:"serverInfo"`
}

type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Sampling     map[string]interface{} `json:"sampling,omitempty"`
}

type ServerCapabilities struct {
	Logging      *LoggingCapability      `json:"logging,omitempty"`
	Prompts      *PromptsCapability      `json:"prompts,omitempty"`
	Resources    *ResourcesCapability    `json:"resources,omitempty"`
	Tools        *ToolsCapability        `json:"tools,omitempty"`
	Experimental map[string]interface{}  `json:"experimental,omitempty"`
}

type LoggingCapability struct{}
type PromptsCapability struct{}
type ResourcesCapability struct{}
type ToolsCapability struct{}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool related types
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema ToolSchema  `json:"inputSchema"`
}

type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResponse struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type     string      `json:"type"`
	Text     string      `json:"text,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	MimeType string      `json:"mimeType,omitempty"`
}

// Notification types
type LoggingMessage struct {
	Level  string          `json:"level"`
	Data   json.RawMessage `json:"data,omitempty"`
	Logger string          `json:"logger,omitempty"`
}