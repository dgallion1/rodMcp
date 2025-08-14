package webtools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultFileAccessConfig(t *testing.T) {
	config := DefaultFileAccessConfig()
	
	if !config.RestrictToWorkingDir {
		t.Error("Expected RestrictToWorkingDir to be true by default")
	}
	
	if config.AllowTempFiles {
		t.Error("Expected AllowTempFiles to be false by default")
	}
	
	if config.MaxFileSize != 10*1024*1024 {
		t.Errorf("Expected MaxFileSize to be 10MB, got %d", config.MaxFileSize)
	}
	
	workingDir, _ := os.Getwd()
	if len(config.AllowedPaths) != 1 || config.AllowedPaths[0] != workingDir {
		t.Errorf("Expected AllowedPaths to contain working directory %s", workingDir)
	}
}

func TestPathValidatorValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "rodmcp_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Test cases
	tests := []struct {
		name        string
		config      *FileAccessConfig
		path        string
		operation   string
		shouldAllow bool
	}{
		{
			name: "Allow path in allowed list",
			config: &FileAccessConfig{
				AllowedPaths:         []string{tempDir},
				RestrictToWorkingDir: false,
			},
			path:        filepath.Join(tempDir, "test.txt"),
			operation:   "read",
			shouldAllow: true,
		},
		{
			name: "Deny path not in allowed list",
			config: &FileAccessConfig{
				AllowedPaths:         []string{tempDir},
				RestrictToWorkingDir: false,
			},
			path:        "/etc/passwd",
			operation:   "read",
			shouldAllow: false,
		},
		{
			name: "Deny path in deny list",
			config: &FileAccessConfig{
				AllowedPaths: []string{tempDir},
				DenyPaths:    []string{tempDir},
			},
			path:        filepath.Join(tempDir, "test.txt"),
			operation:   "read",
			shouldAllow: false,
		},
		{
			name: "Allow working directory when restricted",
			config: &FileAccessConfig{
				RestrictToWorkingDir: true,
			},
			path:        "test.txt", // relative to working dir
			operation:   "read",
			shouldAllow: true,
		},
		{
			name: "Deny outside working directory when restricted",
			config: &FileAccessConfig{
				RestrictToWorkingDir: true,
			},
			path:        "/etc/passwd",
			operation:   "read",
			shouldAllow: false,
		},
		{
			name: "Allow temp files when enabled",
			config: &FileAccessConfig{
				AllowTempFiles: true,
			},
			path:        filepath.Join(os.TempDir(), "test.txt"),
			operation:   "write",
			shouldAllow: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewPathValidator(tt.config)
			err := validator.ValidatePath(tt.path, tt.operation)
			
			if tt.shouldAllow && err != nil {
				t.Errorf("Expected path %s to be allowed, but got error: %v", tt.path, err)
			}
			
			if !tt.shouldAllow && err == nil {
				t.Errorf("Expected path %s to be denied, but it was allowed", tt.path)
			}
		})
	}
}

func TestPathValidatorValidateFileSize(t *testing.T) {
	tests := []struct {
		name      string
		config    *FileAccessConfig
		size      int64
		shouldErr bool
	}{
		{
			name: "Allow small file",
			config: &FileAccessConfig{
				MaxFileSize: 1024,
			},
			size:      512,
			shouldErr: false,
		},
		{
			name: "Deny large file",
			config: &FileAccessConfig{
				MaxFileSize: 1024,
			},
			size:      2048,
			shouldErr: true,
		},
		{
			name: "No limit when MaxFileSize is 0",
			config: &FileAccessConfig{
				MaxFileSize: 0,
			},
			size:      1024 * 1024 * 1024, // 1GB
			shouldErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewPathValidator(tt.config)
			err := validator.ValidateFileSize(tt.size)
			
			if tt.shouldErr && err == nil {
				t.Errorf("Expected file size %d to be rejected", tt.size)
			}
			
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected file size %d to be allowed, but got error: %v", tt.size, err)
			}
		})
	}
}

func TestPathValidatorIsPathUnder(t *testing.T) {
	validator := NewPathValidator(nil)
	
	tests := []struct {
		name       string
		targetPath string
		basePath   string
		expected   bool
	}{
		{
			name:       "Path directly under base",
			targetPath: "/home/user/documents/file.txt",
			basePath:   "/home/user",
			expected:   true,
		},
		{
			name:       "Path not under base",
			targetPath: "/etc/passwd",
			basePath:   "/home/user",
			expected:   false,
		},
		{
			name:       "Path equals base",
			targetPath: "/home/user",
			basePath:   "/home/user",
			expected:   true,
		},
		{
			name:       "Path with similar prefix but different directory",
			targetPath: "/home/user2/file.txt",
			basePath:   "/home/user",
			expected:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isPathUnder(tt.targetPath, tt.basePath)
			if result != tt.expected {
				t.Errorf("isPathUnder(%s, %s) = %v, expected %v", 
					tt.targetPath, tt.basePath, result, tt.expected)
			}
		})
	}
}

func TestEmptyPathValidation(t *testing.T) {
	validator := NewPathValidator(DefaultFileAccessConfig())
	
	err := validator.ValidatePath("", "read")
	if err == nil {
		t.Error("Expected empty path to be rejected")
	}
	
	if err.Error() != "path cannot be empty" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}
}

func TestGetAllowedPaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rodmcp_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	config := &FileAccessConfig{
		AllowedPaths:         []string{tempDir, "/home/user"},
		RestrictToWorkingDir: true,
		AllowTempFiles:      true,
	}
	
	validator := NewPathValidator(config)
	allowedPaths := validator.GetAllowedPaths()
	
	// Should include working dir, temp dir, and the specified allowed paths
	expectedMinLength := 4 // working dir + temp dir + 2 allowed paths
	if len(allowedPaths) < expectedMinLength {
		t.Errorf("Expected at least %d allowed paths, got %d: %v", 
			expectedMinLength, len(allowedPaths), allowedPaths)
	}
	
	// Check that temp dir is included
	tempDirIncluded := false
	for _, path := range allowedPaths {
		if path == os.TempDir() {
			tempDirIncluded = true
			break
		}
	}
	
	if !tempDirIncluded {
		t.Error("Expected temp directory to be in allowed paths list")
	}
}