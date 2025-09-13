package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileOperations handles file copying and ignore patterns
type FileOperations struct {
	IgnorePatterns []string
}

// NewFileOperations creates a new FileOperations instance
func NewFileOperations() *FileOperations {
	return &FileOperations{
		IgnorePatterns: make([]string, 0),
	}
}

// LoadIgnorePatterns loads ignore patterns from .otterignore file
func (f *FileOperations) LoadIgnorePatterns(projectRoot string) error {
	ignorePath := filepath.Join(projectRoot, ".otterignore")

	// If .otterignore doesn't exist, that's fine
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(ignorePath)
	if err != nil {
		return fmt.Errorf("failed to open .otterignore: %w", err)
	}
	defer file.Close()

	f.IgnorePatterns = make([]string, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		f.IgnorePatterns = append(f.IgnorePatterns, line)
	}

	return scanner.Err()
}

// IsIgnored checks if a file path should be ignored based on ignore patterns
func (f *FileOperations) IsIgnored(relativePath string) bool {
	for _, pattern := range f.IgnorePatterns {
		if f.matchPattern(pattern, relativePath) {
			return true
		}
	}
	return false
}

// matchPattern checks if a path matches an ignore pattern
func (f *FileOperations) matchPattern(pattern, path string) bool {
	// Simple pattern matching - can be enhanced with more complex glob patterns later

	// Exact match
	if pattern == path {
		return true
	}

	// Directory pattern (ends with /)
	if strings.HasSuffix(pattern, "/") {
		dirPattern := strings.TrimSuffix(pattern, "/")
		return strings.HasPrefix(path, dirPattern+"/") || path == dirPattern
	}

	// Wildcard pattern (contains *)
	if strings.Contains(pattern, "*") {
		return f.matchWildcard(pattern, path)
	}

	// Prefix match
	return strings.HasPrefix(path, pattern)
}

// matchWildcard performs simple wildcard matching
func (f *FileOperations) matchWildcard(pattern, path string) bool {
	// Simple implementation for basic wildcards
	// This can be enhanced with more sophisticated pattern matching

	if pattern == "*" {
		return true
	}

	if strings.HasPrefix(pattern, "*.") {
		extension := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, extension)
	}

	return false
}

// CopyLayer copies files from a layer directory to the target directory
func (f *FileOperations) CopyLayer(layerPath, targetPath string, projectRoot string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetPath, err)
	}

	return filepath.Walk(layerPath, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from layer root
		relativePath, err := filepath.Rel(layerPath, srcPath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip the root directory itself
		if relativePath == "." {
			return nil
		}

		// Check if this file should be ignored
		if f.IsIgnored(relativePath) {
			fmt.Printf("  Ignoring: %s\n", relativePath)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate destination path
		destPath := filepath.Join(targetPath, relativePath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// Copy file
			return f.copyFile(srcPath, destPath, info.Mode())
		}
	})
}

// copyFile copies a single file from src to dst
func (f *FileOperations) copyFile(src, dst string, mode os.FileMode) error {
	// Check if destination file exists and prompt for overwrite
	if _, err := os.Stat(dst); err == nil {
		fmt.Printf("  Overwriting: %s\n", dst)
	} else {
		fmt.Printf("  Creating: %s\n", dst)
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Set file permissions
	return os.Chmod(dst, mode)
}
