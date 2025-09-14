package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLayerSpecificIgnorePatterns(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create project root with .otterignore
	projectRoot := filepath.Join(tempDir, "project")
	err := os.MkdirAll(projectRoot, 0755)
	if err != nil {
		t.Fatalf("Failed to create project root: %v", err)
	}

	// Create project .otterignore that ignores README.md
	projectIgnore := filepath.Join(projectRoot, ".otterignore")
	err = os.WriteFile(projectIgnore, []byte("README.md\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create project .otterignore: %v", err)
	}

	// Create layer directory with its own .otterignore
	layerDir := filepath.Join(tempDir, "layer")
	err = os.MkdirAll(layerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create layer directory: %v", err)
	}

	// Create layer .otterignore that ignores LICENSE
	layerIgnore := filepath.Join(layerDir, ".otterignore")
	err = os.WriteFile(layerIgnore, []byte("LICENSE\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create layer .otterignore: %v", err)
	}

	// Create layer files
	layerFiles := map[string]string{
		"LICENSE":      "MIT License...",
		"README.md":    "# My Layer",
		"FOO.md":       "# FOO Documentation",
		".otterignore": "LICENSE", // This should be ignored by default
	}

	for filename, content := range layerFiles {
		filePath := filepath.Join(layerDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create layer file %s: %v", filename, err)
		}
	}

	// Create target directory
	targetDir := filepath.Join(tempDir, "target")

	// Initialize FileOperations and load project ignore patterns
	fileOps := NewFileOperations()
	err = fileOps.LoadIgnorePatterns(projectRoot)
	if err != nil {
		t.Fatalf("Failed to load project ignore patterns: %v", err)
	}

	// Copy layer to target
	err = fileOps.CopyLayer(layerDir, targetDir, projectRoot)
	if err != nil {
		t.Fatalf("Failed to copy layer: %v", err)
	}

	// Verify results
	expectedFiles := []string{"FOO.md"}
	ignoredFiles := []string{"LICENSE", "README.md", ".otterignore"}

	// Check that expected files exist
	for _, filename := range expectedFiles {
		filePath := filepath.Join(targetDir, filename)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("Expected file %s was not copied to target", filename)
		}
	}

	// Check that ignored files don't exist
	for _, filename := range ignoredFiles {
		filePath := filepath.Join(targetDir, filename)
		if _, err := os.Stat(filePath); err == nil {
			t.Errorf("Ignored file %s was incorrectly copied to target", filename)
		}
	}
}

func TestLoadLayerIgnorePatterns(t *testing.T) {
	tempDir := t.TempDir()
	fileOps := NewFileOperations()

	tests := []struct {
		name             string
		ignoreContent    string
		expectedPatterns []string
		hasIgnoreFile    bool
	}{
		{
			name:             "Layer with .otterignore",
			ignoreContent:    "*.log\ntemp/\n# comment\n\nsecrets.txt",
			expectedPatterns: []string{"*.log", "temp/", "secrets.txt"},
			hasIgnoreFile:    true,
		},
		{
			name:             "Layer without .otterignore",
			expectedPatterns: []string{},
			hasIgnoreFile:    false,
		},
		{
			name:             "Layer with empty .otterignore",
			ignoreContent:    "# Only comments\n\n",
			expectedPatterns: []string{},
			hasIgnoreFile:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create layer directory
			layerDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(layerDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create layer directory: %v", err)
			}

			// Create .otterignore if needed
			if tt.hasIgnoreFile {
				ignorePath := filepath.Join(layerDir, ".otterignore")
				err = os.WriteFile(ignorePath, []byte(tt.ignoreContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create .otterignore: %v", err)
				}
			}

			// Load patterns
			patterns, err := fileOps.loadLayerIgnorePatterns(layerDir)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify patterns
			if len(patterns) != len(tt.expectedPatterns) {
				t.Errorf("Expected %d patterns, got %d", len(tt.expectedPatterns), len(patterns))
				return
			}

			for i, expected := range tt.expectedPatterns {
				if i >= len(patterns) || patterns[i] != expected {
					t.Errorf("Expected pattern '%s', got '%s'", expected, patterns[i])
				}
			}
		})
	}
}

func TestCombinedIgnorePatterns(t *testing.T) {
	tempDir := t.TempDir()

	// Create project with ignore patterns
	projectRoot := filepath.Join(tempDir, "project")
	err := os.MkdirAll(projectRoot, 0755)
	if err != nil {
		t.Fatalf("Failed to create project root: %v", err)
	}

	projectIgnore := filepath.Join(projectRoot, ".otterignore")
	err = os.WriteFile(projectIgnore, []byte("project-ignore.txt\nshared-ignore.txt\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create project .otterignore: %v", err)
	}

	// Create layer with ignore patterns
	layerDir := filepath.Join(tempDir, "layer")
	err = os.MkdirAll(layerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create layer directory: %v", err)
	}

	layerIgnore := filepath.Join(layerDir, ".otterignore")
	err = os.WriteFile(layerIgnore, []byte("layer-ignore.txt\nshared-ignore.txt\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create layer .otterignore: %v", err)
	}

	// Create test files (excluding .otterignore which was already created above)
	testFiles := []string{
		"project-ignore.txt", // Should be ignored by project patterns
		"layer-ignore.txt",   // Should be ignored by layer patterns
		"shared-ignore.txt",  // Should be ignored by both (duplicate pattern)
		"keep-this.txt",      // Should NOT be ignored
		// Note: .otterignore was already created above and shouldn't be overwritten
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(layerDir, filename)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create target directory
	targetDir := filepath.Join(tempDir, "target")

	// Initialize FileOperations and load project patterns
	fileOps := NewFileOperations()
	err = fileOps.LoadIgnorePatterns(projectRoot)
	if err != nil {
		t.Fatalf("Failed to load project ignore patterns: %v", err)
	}

	// Copy layer
	err = fileOps.CopyLayer(layerDir, targetDir, projectRoot)
	if err != nil {
		t.Fatalf("Failed to copy layer: %v", err)
	}

	// Verify only keep-this.txt was copied
	expectedFiles := []string{"keep-this.txt"}
	ignoredFiles := []string{"project-ignore.txt", "layer-ignore.txt", "shared-ignore.txt", ".otterignore"}

	for _, filename := range expectedFiles {
		filePath := filepath.Join(targetDir, filename)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("Expected file %s was not copied", filename)
		}
	}

	for _, filename := range ignoredFiles {
		filePath := filepath.Join(targetDir, filename)
		if _, err := os.Stat(filePath); err == nil {
			t.Errorf("File %s should have been ignored but was copied", filename)
		}
	}
}

func TestCriticalFileProtection(t *testing.T) {
	// Test that critical files/directories (.git, .otter, .otterignore) are NEVER copied from layers
	tempDir := t.TempDir()

	// Create project root
	projectRoot := filepath.Join(tempDir, "project")
	err := os.MkdirAll(projectRoot, 0755)
	if err != nil {
		t.Fatalf("Failed to create project root: %v", err)
	}

	// Create layer directory with critical files that should NEVER be copied
	layerDir := filepath.Join(tempDir, "layer")
	err = os.MkdirAll(layerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create layer directory: %v", err)
	}

	// Create .git directory in layer (this should NEVER be copied)
	gitDir := filepath.Join(layerDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create .otter directory in layer (this should NEVER be copied)
	otterDir := filepath.Join(layerDir, ".otter")
	err = os.MkdirAll(otterDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .otter directory: %v", err)
	}

	// Create critical files that should NEVER be copied
	criticalFiles := map[string]string{
		".git/config":       "[core]\n  repositoryformatversion = 0",
		".git/HEAD":         "ref: refs/heads/main",
		".otter/cache.json": "{}",
		".otterignore":      "*.tmp",
		".gitignore":        "*.log\n*.tmp\nnode_modules/",
		"safe-file.txt":     "This file should be copied",
	}

	for filename, content := range criticalFiles {
		filePath := filepath.Join(layerDir, filename)
		// Ensure directory exists for nested files
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", filename, err)
		}

		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create critical file %s: %v", filename, err)
		}
	}

	// Create target directory
	targetDir := filepath.Join(tempDir, "target")

	// Initialize FileOperations and copy layer
	fileOps := NewFileOperations()
	err = fileOps.LoadIgnorePatterns(projectRoot)
	if err != nil {
		t.Fatalf("Failed to load project ignore patterns: %v", err)
	}

	// Copy layer to target
	err = fileOps.CopyLayer(layerDir, targetDir, projectRoot)
	if err != nil {
		t.Fatalf("Failed to copy layer: %v", err)
	}

	// Verify that critical files/directories were NOT copied
	criticalItems := []string{
		".git",
		".git/config",
		".git/HEAD",
		".otter",
		".otter/cache.json",
		".otterignore",
		".gitignore",
	}

	for _, item := range criticalItems {
		itemPath := filepath.Join(targetDir, item)
		if _, err := os.Stat(itemPath); err == nil {
			t.Errorf("CRITICAL SECURITY ISSUE: %s was copied from layer to target (this should NEVER happen)", item)
		}
	}

	// Verify that safe files WERE copied
	safeFile := filepath.Join(targetDir, "safe-file.txt")
	if _, err := os.Stat(safeFile); err != nil {
		t.Errorf("Safe file was not copied: %s", safeFile)
	}
}

func TestIsIgnoredWithPatterns(t *testing.T) {
	fileOps := NewFileOperations()

	patterns := []string{
		"*.log",
		"temp/",
		"secrets.txt",
		"node_modules/",
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"file.log", true},
		{"debug.log", true},
		{"temp/file.txt", true},
		{"temp/", true},
		{"secrets.txt", true},
		{"node_modules/package.json", true},
		{"src/main.go", false},
		{"README.md", false},
		{"logs/error.txt", false}, // logs/ is not in patterns, only *.log files
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := fileOps.isIgnoredWithPatterns(tt.path, patterns)
			if result != tt.expected {
				t.Errorf("isIgnoredWithPatterns(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}
