# File Access Security

RodMCP implements comprehensive file access control to restrict file operations to authorized paths only, providing defense against directory traversal attacks and unauthorized file access.

## Overview

The file access security system uses an allowlist-based approach with configurable path restrictions for the following file tools:
- `read_file` - Read file contents  
- `write_file` - Write content to files
- `list_directory` - List directory contents

## Default Security Configuration

By default, RodMCP restricts file operations to the **current working directory only**:

```go
config := DefaultFileAccessConfig()
// Results in:
// - RestrictToWorkingDir: true  
// - AllowedPaths: [current working directory]
// - MaxFileSize: 10MB
// - AllowTempFiles: false
// - DenyPaths: []
```

## Configuration Options

### FileAccessConfig Structure

```go
type FileAccessConfig struct {
    // AllowedPaths lists directory prefixes that are allowed for file operations
    AllowedPaths []string `json:"allowed_paths"`
    
    // DenyPaths lists directory prefixes that are explicitly denied (overrides AllowedPaths)
    DenyPaths []string `json:"deny_paths"`
    
    // RestrictToWorkingDir restricts access to current working directory only
    RestrictToWorkingDir bool `json:"restrict_to_working_dir"`
    
    // AllowTempFiles allows access to system temporary directory
    AllowTempFiles bool `json:"allow_temp_files"`
    
    // MaxFileSize limits file operations to files under this size (bytes, 0 = no limit)
    MaxFileSize int64 `json:"max_file_size"`
}
```

### Configuration Examples

#### Restrict to Working Directory Only (Default)
```go
config := &FileAccessConfig{
    RestrictToWorkingDir: true,
}
```

#### Allow Specific Directories
```go
config := &FileAccessConfig{
    AllowedPaths: []string{
        "/home/user/projects",
        "/tmp/workspace",
        "/var/app/data",
    },
    RestrictToWorkingDir: false,
}
```

#### Allow Temp Files for Processing
```go
config := &FileAccessConfig{
    RestrictToWorkingDir: true,
    AllowTempFiles: true,  // Also allows /tmp directory
}
```

#### Deny Sensitive Directories
```go
config := &FileAccessConfig{
    AllowedPaths: []string{"/home/user"},
    DenyPaths: []string{
        "/home/user/.ssh",
        "/home/user/.config",
    },
}
```

#### Custom File Size Limits
```go
config := &FileAccessConfig{
    RestrictToWorkingDir: true,
    MaxFileSize: 50 * 1024 * 1024,  // 50MB limit
}
```

## Security Features

### Path Validation
- **Absolute Path Resolution**: All paths converted to absolute paths before validation
- **Symlink Resolution**: Follows symlinks to prevent bypass attempts  
- **Directory Traversal Protection**: Uses `filepath.Clean()` and prefix matching
- **Allowlist Enforcement**: Only explicitly allowed paths are accessible

### Access Control Hierarchy
1. **Empty Path Rejection**: Empty paths are always denied
2. **Deny List Priority**: Paths in deny list are rejected regardless of allow list
3. **Working Directory Restriction**: When enabled, only working directory subtree allowed
4. **Allow List Check**: Path must match an allowed directory prefix
5. **Temp Directory**: When enabled, system temp directory is allowed

### File Size Limits
- **Write Operations**: File size validated before writing
- **Configurable Limits**: Set maximum file size per configuration
- **Zero = No Limit**: Setting `MaxFileSize` to 0 disables size checking

## Error Handling

File access violations return descriptive error messages:

```
file access denied: access denied: path /etc/passwd is not in allowed paths
directory access denied: access denied: path /home/other is in deny list  
file size validation failed: file size 52428800 bytes exceeds maximum allowed size 10485760 bytes
```

## Integration

### Tool Creation with Path Validation

```go
// Create path validator with custom configuration
config := &FileAccessConfig{
    AllowedPaths: []string{"/safe/directory"},
    MaxFileSize: 10 * 1024 * 1024,
}
validator := NewPathValidator(config)

// Create file tools with validation
readTool := NewReadFileTool(logger, validator)
writeTool := NewWriteFileTool(logger, validator)  
listTool := NewListDirectoryTool(logger, validator)
```

### Default Secure Configuration

When `validator` is `nil`, tools automatically use the default secure configuration:

```go
// These are equivalent:
tool1 := NewReadFileTool(logger, nil)
tool2 := NewReadFileTool(logger, NewPathValidator(DefaultFileAccessConfig()))
```

## Testing

The security implementation includes comprehensive tests covering:
- Path validation logic
- File size limits  
- Configuration edge cases
- Working directory restrictions
- Temp file access
- Deny list precedence

Run tests:
```bash
go test ./internal/webtools -v -run "TestPath|TestDefault|TestEmpty|TestGet"
```

## Security Best Practices

1. **Use Default Configuration**: Start with the secure default (working directory only)
2. **Minimize Allowed Paths**: Only grant access to directories that are absolutely necessary
3. **Use Deny Lists**: Explicitly block sensitive directories even within allowed paths
4. **Set File Size Limits**: Prevent resource exhaustion with appropriate size limits
5. **Regular Auditing**: Review and audit allowed paths regularly
6. **Monitor Access**: Enable logging to track file access patterns

## Migration from Unprotected Access

Existing deployments will automatically gain protection when upgrading:
- File tools now require a `PathValidator` parameter
- Default configuration restricts to working directory only
- No breaking API changes for end users (MCP protocol unchanged)
- Existing file operations outside working directory will be blocked

## Limitations

- **Performance**: Path resolution and validation adds minimal overhead
- **Symlink Handling**: Symlinks are followed, which could potentially bypass restrictions in complex setups
- **Race Conditions**: TOCTOU races between validation and file operation (inherent filesystem limitation)
- **Case Sensitivity**: Path matching is case-sensitive on case-sensitive filesystems

## Future Enhancements

Potential future security improvements:
- Configuration file loading for dynamic path management
- User-based access controls  
- Audit logging for file operations
- Integration with system access controls (ACLs)
- Pattern-based path matching (glob support)