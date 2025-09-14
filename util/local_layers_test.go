package util

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestIsLocalLayer(t *testing.T) {
	gitOps := NewGitOperations("/tmp/cache")

	tests := []struct {
		name     string
		repoURL  string
		expected bool
	}{
		{
			name:     "Relative path with dot slash",
			repoURL:  "./local-layer",
			expected: true,
		},
		{
			name:     "Relative path with dot dot slash",
			repoURL:  "../parent-layer",
			expected: true,
		},
		{
			name:     "Absolute path",
			repoURL:  "/absolute/path/to/layer",
			expected: true,
		},
		{
			name:     "File URI scheme",
			repoURL:  "file:///path/to/layer",
			expected: true,
		},
		{
			name:     "Windows absolute path with drive letter",
			repoURL:  "C:\\path\\to\\layer",
			expected: true,
		},
		{
			name:     "Windows absolute path with forward slash",
			repoURL:  "C:/path/to/layer",
			expected: true,
		},
		{
			name:     "Git SSH URL",
			repoURL:  "git@github.com:user/repo.git",
			expected: false,
		},
		{
			name:     "Git HTTPS URL",
			repoURL:  "https://github.com/user/repo.git",
			expected: false,
		},
		{
			name:     "Git protocol URL",
			repoURL:  "git://github.com/user/repo.git",
			expected: false,
		},
		{
			name:     "Relative path without dot slash",
			repoURL:  "layers/my-layer",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gitOps.isLocalLayer(tt.repoURL)
			if result != tt.expected {
				t.Errorf("isLocalLayer(%s) = %v, expected %v", tt.repoURL, result, tt.expected)
			}
		})
	}
}

func TestHandleLocalLayer(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	layerDir := filepath.Join(tempDir, "test-layer")

	// Create the layer directory with some content
	err := os.MkdirAll(layerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test layer directory: %v", err)
	}

	// Create a test file in the layer
	testFile := filepath.Join(layerDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	gitOps := NewGitOperations("/tmp/cache")

	tests := []struct {
		name      string
		repoURL   string
		expectErr bool
	}{
		{
			name:      "Valid relative path",
			repoURL:   "./test-layer",
			expectErr: false,
		},
		{
			name:      "Valid absolute path",
			repoURL:   layerDir,
			expectErr: false,
		},
		{
			name:      "Valid file URI",
			repoURL:   "file://" + layerDir,
			expectErr: false,
		},
		{
			name:      "Non-existent directory",
			repoURL:   "./non-existent",
			expectErr: true,
		},
		{
			name:      "Path to file instead of directory",
			repoURL:   testFile,
			expectErr: true,
		},
	}

	// Change to temp directory for relative path tests
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layerPath, err := gitOps.handleLocalLayer(tt.repoURL)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.repoURL)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.repoURL, err)
				return
			}

			// Verify the returned path exists and is a directory
			if stat, err := os.Stat(layerPath); err != nil {
				t.Errorf("Returned path %s does not exist: %v", layerPath, err)
			} else if !stat.IsDir() {
				t.Errorf("Returned path %s is not a directory", layerPath)
			}

			// Verify the path is absolute
			if !filepath.IsAbs(layerPath) {
				t.Errorf("Returned path %s is not absolute", layerPath)
			}
		})
	}
}

func TestCloneOrUpdateLayer_LocalLayers(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	layerDir := filepath.Join(tempDir, "test-layer")

	// Create the layer directory with some content
	err := os.MkdirAll(layerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test layer directory: %v", err)
	}

	// Create test files
	files := map[string]string{
		"config.yaml": "key: value",
		"script.sh":   "#!/bin/bash\necho 'hello'",
		"README.md":   "# Test Layer",
	}

	for filename, content := range files {
		filePath := filepath.Join(layerDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	gitOps := NewGitOperations(filepath.Join(tempDir, "cache"))

	tests := []struct {
		name      string
		repoURL   string
		expectErr bool
	}{
		{
			name:      "Relative path layer",
			repoURL:   "./test-layer",
			expectErr: false,
		},
		{
			name:      "Absolute path layer",
			repoURL:   layerDir,
			expectErr: false,
		},
		{
			name:      "File URI layer",
			repoURL:   "file://" + layerDir,
			expectErr: false,
		},
		{
			name:      "Non-existent local layer",
			repoURL:   "./missing-layer",
			expectErr: true,
		},
	}

	// Change to temp directory for relative path tests
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layerPath, err := gitOps.CloneOrUpdateLayer(tt.repoURL)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.repoURL)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.repoURL, err)
				return
			}

			// Verify all test files exist in the returned layer path
			for filename := range files {
				filePath := filepath.Join(layerPath, filename)
				if _, err := os.Stat(filePath); err != nil {
					t.Errorf("Expected file %s not found in layer path %s", filename, layerPath)
				}
			}
		})
	}
}

func TestGetRepositoryCommit_LocalLayers(t *testing.T) {
	tempDir := t.TempDir()

	// Create a regular directory (not a git repo)
	regularDir := filepath.Join(tempDir, "regular-dir")
	err := os.MkdirAll(regularDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create regular directory: %v", err)
	}

	gitOps := NewGitOperations("/tmp/cache")

	tests := []struct {
		name           string
		path           string
		expectedResult string
		expectErr      bool
	}{
		{
			name:           "Regular directory (not git repo)",
			path:           regularDir,
			expectedResult: "local-dir",
			expectErr:      false,
		},
		{
			name:      "Non-existent directory",
			path:      filepath.Join(tempDir, "missing"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gitOps.GetRepositoryCommit(tt.path)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.path)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.path, err)
				return
			}

			if result != tt.expectedResult {
				t.Errorf("Expected result %s, got %s", tt.expectedResult, result)
			}
		})
	}
}

func TestLocalLayerIntegration(t *testing.T) {
	// Create a complete test setup
	tempDir := t.TempDir()

	// Create multiple layer directories
	layer1Dir := filepath.Join(tempDir, "base-layer")
	layer2Dir := filepath.Join(tempDir, "config-layer")

	// Create layer 1
	err := os.MkdirAll(layer1Dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create layer1 directory: %v", err)
	}

	layer1Files := map[string]string{
		"base.txt":    "base content",
		"shared.conf": "base_config=true",
	}

	for filename, content := range layer1Files {
		filePath := filepath.Join(layer1Dir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	// Create layer 2
	err = os.MkdirAll(layer2Dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create layer2 directory: %v", err)
	}

	layer2Files := map[string]string{
		"config.yaml": "environment: local",
		"app.conf":    "debug=true",
	}

	for filename, content := range layer2Files {
		filePath := filepath.Join(layer2Dir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	// Test the complete flow
	gitOps := NewGitOperations(filepath.Join(tempDir, "cache"))

	// Change to temp directory
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Test processing multiple local layers
	layers := []string{
		"./base-layer",
		"./config-layer",
		"file://" + layer1Dir, // Test file URI with absolute path
	}

	for i, layerURL := range layers {
		t.Run(fmt.Sprintf("Layer_%d_%s", i, layerURL), func(t *testing.T) {
			layerPath, err := gitOps.CloneOrUpdateLayer(layerURL)
			if err != nil {
				t.Errorf("Failed to process layer %s: %v", layerURL, err)
				return
			}

			// Verify the path exists
			if _, err := os.Stat(layerPath); err != nil {
				t.Errorf("Layer path %s does not exist: %v", layerPath, err)
			}

			// Test commit info
			commit, err := gitOps.GetRepositoryCommit(layerPath)
			if err != nil {
				t.Errorf("Failed to get commit info for %s: %v", layerPath, err)
			} else if commit != "local-dir" {
				t.Errorf("Expected commit info 'local-dir', got '%s'", commit)
			}
		})
	}
}
