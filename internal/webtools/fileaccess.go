package webtools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileAccessConfig defines file access restrictions
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

// DefaultFileAccessConfig returns a secure default configuration
func DefaultFileAccessConfig() *FileAccessConfig {
	workingDir, _ := os.Getwd()
	return &FileAccessConfig{
		AllowedPaths:         []string{workingDir},
		DenyPaths:           []string{},
		RestrictToWorkingDir: true,
		AllowTempFiles:      false,
		MaxFileSize:         10 * 1024 * 1024, // 10MB default
	}
}

// PathValidator handles file path access validation
type PathValidator struct {
	config *FileAccessConfig
}

// NewPathValidator creates a new path validator with the given configuration
func NewPathValidator(config *FileAccessConfig) *PathValidator {
	if config == nil {
		config = DefaultFileAccessConfig()
	}
	return &PathValidator{config: config}
}

// ValidatePath validates if a given path is allowed for access
func (pv *PathValidator) ValidatePath(inputPath string, operation string) error {
	if inputPath == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Clean and resolve the path to prevent traversal attacks
	cleanPath := filepath.Clean(inputPath)
	
	// Convert to absolute path for consistent comparison
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for %s: %w", cleanPath, err)
	}

	// Resolve any symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If symlink resolution fails, use the absolute path
		// This handles cases where the target doesn't exist yet (for writes)
		realPath = absPath
	}

	// Check against deny list first (takes precedence)
	if pv.isDenied(realPath) {
		return fmt.Errorf("access denied: path %s is in deny list", realPath)
	}

	// Check against allow list
	if !pv.isAllowed(realPath) {
		return fmt.Errorf("access denied: path %s is not in allowed paths", realPath)
	}

	return nil
}

// ValidateFileSize checks if a file size is within limits for write operations
func (pv *PathValidator) ValidateFileSize(size int64) error {
	if pv.config.MaxFileSize > 0 && size > pv.config.MaxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes", 
			size, pv.config.MaxFileSize)
	}
	return nil
}

// isAllowed checks if the path is in the allowed paths list
func (pv *PathValidator) isAllowed(path string) bool {
	// If restricting to working directory only, check that
	if pv.config.RestrictToWorkingDir {
		workingDir, err := os.Getwd()
		if err == nil {
			absWorkingDir, err := filepath.Abs(workingDir)
			if err == nil {
				// Return true if path is under working directory, false otherwise
				return pv.isPathUnder(path, absWorkingDir)
			}
		}
		// If we can't determine working directory, deny access for security
		return false
	}

	// Check temp files access
	if pv.config.AllowTempFiles {
		tempDir := os.TempDir()
		if absTempDir, err := filepath.Abs(tempDir); err == nil {
			if pv.isPathUnder(path, absTempDir) {
				return true
			}
		}
	}

	// Check allowed paths list
	for _, allowedPath := range pv.config.AllowedPaths {
		absAllowedPath, err := filepath.Abs(allowedPath)
		if err != nil {
			continue
		}
		
		if pv.isPathUnder(path, absAllowedPath) {
			return true
		}
	}

	// If no allowed paths specified and not restricting to working dir, allow all
	if len(pv.config.AllowedPaths) == 0 && !pv.config.RestrictToWorkingDir {
		return true
	}

	return false
}

// isDenied checks if the path is in the denied paths list
func (pv *PathValidator) isDenied(path string) bool {
	for _, denyPath := range pv.config.DenyPaths {
		absDenyPath, err := filepath.Abs(denyPath)
		if err != nil {
			continue
		}
		
		if pv.isPathUnder(path, absDenyPath) {
			return true
		}
	}
	return false
}

// isPathUnder checks if targetPath is under or equal to basePath
func (pv *PathValidator) isPathUnder(targetPath, basePath string) bool {
	// Ensure both paths end with separator for consistent comparison
	basePath = strings.TrimSuffix(basePath, string(filepath.Separator)) + string(filepath.Separator)
	targetPath = strings.TrimSuffix(targetPath, string(filepath.Separator)) + string(filepath.Separator)
	
	// Check if target path starts with base path
	return strings.HasPrefix(targetPath, basePath) || targetPath == basePath
}

// GetAllowedPaths returns the list of allowed paths for informational purposes
func (pv *PathValidator) GetAllowedPaths() []string {
	var paths []string
	
	if pv.config.RestrictToWorkingDir {
		if workingDir, err := os.Getwd(); err == nil {
			paths = append(paths, workingDir)
		}
	}
	
	if pv.config.AllowTempFiles {
		paths = append(paths, os.TempDir())
	}
	
	paths = append(paths, pv.config.AllowedPaths...)
	
	return paths
}