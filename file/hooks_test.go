package file

import (
	"os"
	"testing"
)

func TestParseGlobalHooks(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedBefore []string
		expectedAfter  []string
		expectedError  []string
		expectError    bool
	}{
		{
			name: "Global hooks with single commands",
			content: `ON_BEFORE_BUILD: ["echo 'Starting build'"]
ON_AFTER_BUILD: ["echo 'Build completed'"]
ON_ERROR: ["echo 'Build failed'"]`,
			expectedBefore: []string{"echo 'Starting build'"},
			expectedAfter:  []string{"echo 'Build completed'"},
			expectedError:  []string{"echo 'Build failed'"},
			expectError:    false,
		},
		{
			name: "Global hooks with multiple commands",
			content: `ON_BEFORE_BUILD: ["echo 'Starting'", "make clean"]
ON_AFTER_BUILD: ["make test", "make package", "echo 'Done'"]
ON_ERROR: ["make clean", "echo 'Error cleanup'"]`,
			expectedBefore: []string{"echo 'Starting'", "make clean"},
			expectedAfter:  []string{"make test", "make package", "echo 'Done'"},
			expectedError:  []string{"make clean", "echo 'Error cleanup'"},
			expectError:    false,
		},
		{
			name: "Only some global hooks defined",
			content: `ON_BEFORE_BUILD: ["echo 'Starting'"]
LAYER ./test-layer`,
			expectedBefore: []string{"echo 'Starting'"},
			expectedAfter:  nil,
			expectedError:  nil,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "test-otterfile-*.txt")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write content
			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tmpFile.Close()

			// Parse the file
			config, err := ParseOtterfile(tmpFile.Name())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check OnBeforeBuild
			if !stringSlicesEqual(config.OnBeforeBuild, tt.expectedBefore) {
				t.Errorf("OnBeforeBuild: expected %v, got %v", tt.expectedBefore, config.OnBeforeBuild)
			}

			// Check OnAfterBuild
			if !stringSlicesEqual(config.OnAfterBuild, tt.expectedAfter) {
				t.Errorf("OnAfterBuild: expected %v, got %v", tt.expectedAfter, config.OnAfterBuild)
			}

			// Check OnError
			if !stringSlicesEqual(config.OnError, tt.expectedError) {
				t.Errorf("OnError: expected %v, got %v", tt.expectedError, config.OnError)
			}
		})
	}
}

func TestParseLayerHooks(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedBefore []string
		expectedAfter  []string
		expectError    bool
	}{
		{
			name:           "Layer with before and after hooks",
			content:        `LAYER ./test-layer BEFORE ["echo 'Before layer'"] AFTER ["echo 'After layer'"]`,
			expectedBefore: []string{"echo 'Before layer'"},
			expectedAfter:  []string{"echo 'After layer'"},
			expectError:    false,
		},
		{
			name:           "Layer with multiple before hooks",
			content:        `LAYER ./test-layer BEFORE ["chmod +x scripts/setup.sh", "./scripts/setup.sh"]`,
			expectedBefore: []string{"chmod +x scripts/setup.sh", "./scripts/setup.sh"},
			expectedAfter:  nil,
			expectError:    false,
		},
		{
			name:           "Layer with template and hooks",
			content:        `LAYER ./test-layer TEMPLATE name=test BEFORE ["echo 'Setup'"] AFTER ["echo 'Cleanup'"]`,
			expectedBefore: []string{"echo 'Setup'"},
			expectedAfter:  []string{"echo 'Cleanup'"},
			expectError:    false,
		},
		{
			name:        "Layer with invalid hook syntax",
			content:     `LAYER ./test-layer BEFORE invalid-syntax`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "test-otterfile-*.txt")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write content
			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tmpFile.Close()

			// Parse the file
			config, err := ParseOtterfile(tmpFile.Name())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(config.Layers) == 0 {
				t.Errorf("Expected at least one layer")
				return
			}

			layer := config.Layers[0]

			// Check Before hooks
			if !stringSlicesEqual(layer.Before, tt.expectedBefore) {
				t.Errorf("Before hooks: expected %v, got %v", tt.expectedBefore, layer.Before)
			}

			// Check After hooks
			if !stringSlicesEqual(layer.After, tt.expectedAfter) {
				t.Errorf("After hooks: expected %v, got %v", tt.expectedAfter, layer.After)
			}
		})
	}
}

// Helper function to compare string slices
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
