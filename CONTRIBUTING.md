# Contributing to RodMCP

Thank you for your interest in contributing to RodMCP! This document provides guidelines and information for contributors.

## 🤝 Ways to Contribute

- **🐛 Bug Reports** - Report issues and bugs
- **💡 Feature Requests** - Suggest new functionality
- **📝 Documentation** - Improve guides and examples
- **🛠️ Code Contributions** - Add features, fix bugs, optimize performance
- **🧪 Testing** - Test on different platforms and configurations
- **🎨 Examples** - Create usage examples and demos

## 📋 Getting Started

### Prerequisites
- Go 1.24 or later
- Git
- Basic understanding of MCP (Model Context Protocol)
- Familiarity with browser automation (helpful but not required)

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/rodmcp.git
   cd rodmcp
   ```

2. **Build and Test**
   ```bash
   go mod download
   go build -o bin/rodmcp cmd/server/main.go
   ./bin/rodmcp --help
   ```

3. **Run Examples**
   ```bash
   go build -o bin/test-example examples/test_example.go
   ./bin/test-example
   ```

## 🏗️ Project Structure

```
rodmcp/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── browser/         # Browser automation (Rod integration)
│   ├── logger/          # Logging system
│   ├── mcp/            # MCP protocol implementation
│   └── webtools/       # Web development tools
├── pkg/types/          # Shared type definitions
├── examples/           # Usage examples and demos
├── configs/            # Configuration templates
└── docs/              # Documentation files
```

## 🛠️ Adding New Tools

### 1. Implement the Tool Interface

Create a new tool in `internal/webtools/`:

```go
type YourTool struct {
    logger *logger.Logger
    browser *browser.Manager
}

func NewYourTool(log *logger.Logger, browserMgr *browser.Manager) *YourTool {
    return &YourTool{logger: log, browser: browserMgr}
}

func (t *YourTool) Name() string {
    return "your_tool"
}

func (t *YourTool) Description() string {
    return "Description of what your tool does"
}

func (t *YourTool) InputSchema() types.ToolSchema {
    return types.ToolSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "param1": map[string]interface{}{
                "type": "string",
                "description": "Description of parameter",
            },
        },
        Required: []string{"param1"},
    }
}

func (t *YourTool) Execute(args map[string]interface{}) (*types.CallToolResponse, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start).Milliseconds()
        t.logger.LogToolExecution(t.Name(), args, true, duration)
    }()

    // Your implementation here
    
    return &types.CallToolResponse{
        Content: []types.ToolContent{{
            Type: "text",
            Text: "Result of your tool execution",
        }},
    }, nil
}
```

### 2. Register the Tool

Add your tool to `cmd/server/main.go`:

```go
mcpServer.RegisterTool(webtools.NewYourTool(log, browserMgr))
```

### 3. Add Documentation

Create usage examples in the appropriate documentation files.

## 📝 Code Style Guidelines

### Go Code Style
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Add comprehensive error handling
- Include logging for debugging
- Write clear, descriptive variable names

### Logging Standards
- Use structured logging with Zap
- Include component context: `logger.WithComponent("component_name")`
- Log important actions and errors
- Use appropriate log levels (DEBUG, INFO, WARN, ERROR)

### Error Handling
```go
if err != nil {
    return &types.CallToolResponse{
        Content: []types.ToolContent{{
            Type: "text",
            Text: fmt.Sprintf("Operation failed: %v", err),
        }},
        IsError: true,
    }, nil
}
```

## 🧪 Testing Guidelines

### Manual Testing
- Test your tool with the example scripts
- Verify both headless and visible browser modes
- Test error conditions and edge cases
- Check logging output for clarity

### Integration Testing
- Ensure your tool works with Claude
- Test MCP protocol compliance
- Verify JSON schema validation

### Platform Testing
- Test on Linux (primary development platform)
- Verify Windows compatibility if possible
- Check macOS compatibility if possible

## 📚 Documentation Standards

### Code Documentation
- Add comments for complex logic
- Document public functions and types
- Include usage examples in comments

### User Documentation
- Update README.md if adding major features
- Add tool documentation to MCP_USAGE.md
- Update INSTALLATION.md if needed
- Add changelog entries

### Example Format
```markdown
### tool_name
Brief description of what the tool does
- **Purpose**: Main use case
- **Parameters**: List of required/optional parameters  
- **Example**: "Ask Claude to use this tool like this"
```

## 🔄 Pull Request Process

### Before Submitting
1. **Build and Test**
   ```bash
   go build -o bin/rodmcp cmd/server/main.go
   ./bin/test-example  # or your test
   ```

2. **Code Quality**
   ```bash
   go fmt ./...
   go vet ./...
   ```

3. **Documentation**
   - Update relevant documentation
   - Add changelog entry
   - Include usage examples

### PR Description Template
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Other (please describe)

## Testing
- [ ] Manual testing completed
- [ ] Works with Claude Desktop
- [ ] Works with Claude Code
- [ ] Documentation updated

## Screenshots (if applicable)
Include screenshots of browser automation or tool results
```

## 🐛 Bug Reports

### Issue Template
```markdown
**Describe the Bug**
Clear description of what the bug is

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. Ask Claude '...'
3. See error

**Expected Behavior**
What you expected to happen

**Environment**
- OS: [e.g., Ubuntu 22.04]
- Go version: [e.g., 1.24.1]
- RodMCP version: [e.g., 1.0.0]
- Claude client: [Desktop/Code]

**Logs**
Include relevant log entries from rodmcp.log
```

## 💡 Feature Requests

### Feature Template
```markdown
**Feature Description**
Clear description of the proposed feature

**Use Case**
Why is this feature needed? What problem does it solve?

**Proposed Implementation**
How would you like to see this implemented?

**Additional Context**
Any other context, mockups, or examples
```

## 🎯 Development Priorities

### High Priority
- 🔧 **Core Tool Improvements** - Enhance existing tools
- 🐛 **Bug Fixes** - Resolve reported issues  
- 📚 **Documentation** - Improve guides and examples
- 🧪 **Testing** - Better test coverage

### Medium Priority
- ⚡ **Performance** - Optimize browser operations
- 🌐 **Compatibility** - Better cross-platform support
- 🛠️ **New Tools** - Additional web development tools
- 🔧 **Configuration** - More customization options

### Future Considerations
- 🔌 **Plugin System** - External tool integration
- 🌍 **Multi-Browser** - Firefox, Safari support
- 📊 **Analytics** - Usage metrics and insights
- 🚀 **CI/CD** - Automated testing and deployment

## ❓ Questions and Support

- **Documentation**: Check existing docs first
- **Issues**: Search existing issues before creating new ones
- **Discussions**: Use GitHub Discussions for questions
- **Community**: Follow project guidelines and be respectful

## 📜 Code of Conduct

- **Be respectful** to all contributors
- **Be constructive** in feedback and discussions  
- **Focus on the code**, not the person
- **Help others** learn and contribute
- **Follow project standards** and guidelines

Thank you for contributing to RodMCP! 🚀