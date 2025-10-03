package util

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

	// Filename pattern (pattern doesn't contain /, should match filename in any directory)
	if !strings.Contains(pattern, "/") {
		pathParts := strings.Split(path, "/")
		filename := pathParts[len(pathParts)-1]
		if pattern == filename {
			return true
		}
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

// loadLayerIgnorePatterns loads ignore patterns from a layer's .otterignore file
func (f *FileOperations) loadLayerIgnorePatterns(layerPath string) ([]string, error) {
	ignorePath := filepath.Join(layerPath, ".otterignore")

	// If .otterignore doesn't exist in the layer, return empty patterns
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		return []string{}, nil
	}

	file, err := os.Open(ignorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open layer .otterignore: %w", err)
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading layer .otterignore: %w", err)
	}

	return patterns, nil
}

// isIgnoredWithPatterns checks if a file path should be ignored based on given patterns
func (f *FileOperations) isIgnoredWithPatterns(relativePath string, patterns []string) bool {
	for _, pattern := range patterns {
		if f.matchPattern(pattern, relativePath) {
			return true
		}
	}
	return false
}

// CopyLayer copies files from a layer directory to the target directory
func (f *FileOperations) CopyLayer(layerPath, targetPath string, projectRoot string, templateVars map[string]string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetPath, err)
	}

	// Load layer-specific ignore patterns and combine with project patterns
	layerIgnorePatterns, err := f.loadLayerIgnorePatterns(layerPath)
	if err != nil {
		return fmt.Errorf("failed to load layer ignore patterns: %w", err)
	}

	// Combine project-level and layer-level ignore patterns
	combinedPatterns := append(f.IgnorePatterns, layerIgnorePatterns...)

	// CRITICAL: Always ignore these files/directories to prevent dangerous overwrites
	criticalIgnorePatterns := []string{
		".git",         // Never copy .git folder from layers (would overwrite project's git repo)
		".git/",        // Directory pattern for .git
		".otter",       // Never copy .otter cache folder from layers
		".otter/",      // Directory pattern for .otter
		".otterignore", // Never copy .otterignore files from layers
		".gitignore",   // Never copy .gitignore files from layers (would overwrite project's git ignore rules)
	}
	combinedPatterns = append(combinedPatterns, criticalIgnorePatterns...)

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

		// Check if this file should be ignored using combined patterns
		if f.isIgnoredWithPatterns(relativePath, combinedPatterns) {
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
			// Copy file with template processing if variables are provided
			return f.copyFile(srcPath, destPath, info.Mode(), templateVars)
		}
	})
}

// copyFile copies a single file from src to dst with optional template processing
func (f *FileOperations) copyFile(src, dst string, mode os.FileMode, templateVars map[string]string) error {
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

	// Read the source file content
	srcContent, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	var finalContent []byte

	// Check if we have template variables and the file contains template syntax
	if len(templateVars) > 0 && f.containsTemplateSyntax(string(srcContent)) {
		// Process the file as a template
		processedContent, err := f.processTemplate(string(srcContent), templateVars, src)
		if err != nil {
			return fmt.Errorf("failed to process template %s: %w", src, err)
		}
		finalContent = []byte(processedContent)
		fmt.Printf("  Template processed: %s\n", dst)
	} else {
		// Copy file as-is
		finalContent = srcContent
	}

	// Write the final content to destination
	if err := os.WriteFile(dst, finalContent, mode); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// containsTemplateSyntax checks if content contains Go template syntax
func (f *FileOperations) containsTemplateSyntax(content string) bool {
	// Check for Go template syntax: {{ .variable }} or {{ .function }}
	return strings.Contains(content, "{{") && strings.Contains(content, "}}")
}

// processTemplate processes a template string with the provided variables
func (f *FileOperations) processTemplate(content string, templateVars map[string]string, filename string) (string, error) {
	// Create a new template
	tmpl, err := template.New(filepath.Base(filename)).Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template with the variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateVars); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
