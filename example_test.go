package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOtterfile(t *testing.T) {
	// Create a temporary Otterfile for testing
	tempDir := t.TempDir()
	otterfilePath := filepath.Join(tempDir, "Otterfile")

	content := `# Test Otterfile
LAYER git@github.com:example/repo1.git
LAYER https://github.com/example/repo2.git TARGET custom/path
LAYER git@github.com:example/repo3.git TARGET .config
`

	err := os.WriteFile(otterfilePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test Otterfile: %v", err)
	}

	// Parse the Otterfile
	config, err := ParseOtterfile(otterfilePath)
	if err != nil {
		t.Fatalf("Failed to parse Otterfile: %v", err)
	}

	// Verify the parsed content
	if len(config.Layers) != 3 {
		t.Errorf("Expected 3 layers, got %d", len(config.Layers))
	}

	// Test first layer
	layer1 := config.Layers[0]
	if layer1.Repository != "git@github.com:example/repo1.git" {
		t.Errorf("Expected repo1, got %s", layer1.Repository)
	}
	if layer1.Target != "." {
		t.Errorf("Expected target '.', got %s", layer1.Target)
	}

	// Test second layer with custom target
	layer2 := config.Layers[1]
	if layer2.Repository != "https://github.com/example/repo2.git" {
		t.Errorf("Expected repo2, got %s", layer2.Repository)
	}
	if layer2.Target != "custom/path" {
		t.Errorf("Expected target 'custom/path', got %s", layer2.Target)
	}

	// Test third layer
	layer3 := config.Layers[2]
	if layer3.Repository != "git@github.com:example/repo3.git" {
		t.Errorf("Expected repo3, got %s", layer3.Repository)
	}
	if layer3.Target != ".config" {
		t.Errorf("Expected target '.config', got %s", layer3.Target)
	}
}

func TestFileOperationsIgnore(t *testing.T) {
	fileOps := NewFileOperations()

	// Test various ignore patterns
	fileOps.ignorePatterns = []string{
		".git/",
		"*.log",
		"node_modules/",
		".DS_Store",
	}

	testCases := []struct {
		path    string
		ignored bool
	}{
		{".git/config", true},
		{".git/hooks/pre-commit", true},
		{"src/main.go", false},
		{"debug.log", true},
		{"application.log", true},
		{"node_modules/package/index.js", true},
		{".DS_Store", true},
		{"src/.DS_Store", true},
		{"README.md", false},
	}

	for _, tc := range testCases {
		result := fileOps.IsIgnored(tc.path)
		if result != tc.ignored {
			t.Errorf("Path %s: expected ignored=%v, got %v", tc.path, tc.ignored, result)
		}
	}
}

func TestGitOperationsRepoName(t *testing.T) {
	gitOps := NewGitOperations("/tmp/cache")

	testCases := []struct {
		repoURL  string
		expected string
	}{
		{
			"git@github.com:user/repo.git",
			"repo-",
		},
		{
			"https://github.com/user/repo.git",
			"repo-",
		},
		{
			"git@gitlab.com:group/project.git",
			"project-",
		},
	}

	for _, tc := range testCases {
		result := gitOps.getRepoDirectoryName(tc.repoURL)
		// The result should start with the expected prefix followed by a hash
		if len(result) <= len(tc.expected) || result[:len(tc.expected)] != tc.expected {
			t.Errorf("URL %s: expected prefix %s, got %s", tc.repoURL, tc.expected, result)
		}
	}
}
