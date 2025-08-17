package mcp

import (
	"fmt"
	"rodmcp/internal/webtools"
	"rodmcp/internal/logger"
	"rodmcp/pkg/types"
)

// Simple test tool that doesn't require external dependencies
type SimpleTestTool struct {
	name        string
	description string
	schema      types.ToolSchema
	result      string
}

func NewSimpleTestTool(name, description, result string) *SimpleTestTool {
	return &SimpleTestTool{
		name:        name,
		description: description,
		result:      result,
		schema: types.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "Test message parameter",
				},
			},
			Required: []string{"message"},
		},
	}
}

func (t *SimpleTestTool) Name() string {
	return t.name
}

func (t *SimpleTestTool) Description() string {
	return t.description
}

func (t *SimpleTestTool) InputSchema() types.ToolSchema {
	return t.schema
}

func (t *SimpleTestTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	message, _ := args["message"].(string)
	
	return &types.CallToolResponse{
		Content: []types.ToolContent{
			{
				Type: "text",
				Text: t.result + ": " + message,
			},
		},
	}, nil
}

// Error test tool that returns an error for testing error handling
type ErrorTestTool struct {
	name        string
	description string
	errorMsg    string
}

func NewErrorTestTool(name, description, errorMsg string) *ErrorTestTool {
	return &ErrorTestTool{
		name:        name,
		description: description,
		errorMsg:    errorMsg,
	}
}

func (t *ErrorTestTool) Name() string {
	return t.name
}

func (t *ErrorTestTool) Description() string {
	return t.description
}

func (t *ErrorTestTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"param": map[string]interface{}{
				"type":        "string",
				"description": "Parameter that will be ignored",
			},
		},
	}
}

func (t *ErrorTestTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
	return nil, fmt.Errorf("%s", t.errorMsg)
}

// Real help tool wrapper for testing
func NewTestHelpTool(log *logger.Logger) Tool {
	return webtools.NewHelpTool(log)
}