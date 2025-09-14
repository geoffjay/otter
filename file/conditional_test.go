package file

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseCondition(t *testing.T) {
	tests := []struct {
		name          string
		conditionStr  string
		expectedKey   string
		expectedValue string
		expectError   bool
	}{
		{
			name:          "Valid env condition",
			conditionStr:  "env=development",
			expectedKey:   "env",
			expectedValue: "development",
			expectError:   false,
		},
		{
			name:          "Valid os condition",
			conditionStr:  "os=darwin",
			expectedKey:   "os",
			expectedValue: "darwin",
			expectError:   false,
		},
		{
			name:          "Valid editor condition with spaces",
			conditionStr:  "editor = vscode",
			expectedKey:   "editor",
			expectedValue: "vscode",
			expectError:   false,
		},
		{
			name:         "Empty condition",
			conditionStr: "",
			expectError:  true,
		},
		{
			name:         "Missing equals",
			conditionStr: "env development",
			expectError:  true,
		},
		{
			name:          "Missing value",
			conditionStr:  "env=",
			expectedKey:   "env",
			expectedValue: "",
			expectError:   false,
		},
		{
			name:          "Multiple equals signs",
			conditionStr:  "custom=value=with=equals",
			expectedKey:   "custom",
			expectedValue: "value=with=equals",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition, err := parseCondition(tt.conditionStr)

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

			if condition.Key != tt.expectedKey {
				t.Errorf("Expected key %s, got %s", tt.expectedKey, condition.Key)
			}

			if condition.Value != tt.expectedValue {
				t.Errorf("Expected value %s, got %s", tt.expectedValue, condition.Value)
			}
		})
	}
}

func TestEvaluateCondition_OS(t *testing.T) {
	condition := &Condition{
		Key:   "os",
		Value: runtime.GOOS,
	}

	result, err := evaluateCondition(condition)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !result {
		t.Errorf("Expected condition to be true for current OS %s", runtime.GOOS)
	}

	// Test with wrong OS
	wrongCondition := &Condition{
		Key:   "os",
		Value: "nonexistent-os",
	}

	result, err = evaluateCondition(wrongCondition)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result {
		t.Errorf("Expected condition to be false for wrong OS")
	}
}

func TestEvaluateCondition_Environment(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		conditionValue string
		expected       bool
	}{
		{
			name:           "OTTER_ENV set",
			envVars:        map[string]string{"OTTER_ENV": "production"},
			conditionValue: "production",
			expected:       true,
		},
		{
			name:           "ENV fallback",
			envVars:        map[string]string{"ENV": "staging"},
			conditionValue: "staging",
			expected:       true,
		},
		{
			name:           "NODE_ENV fallback",
			envVars:        map[string]string{"NODE_ENV": "test"},
			conditionValue: "test",
			expected:       true,
		},
		{
			name:           "Default to development",
			envVars:        map[string]string{},
			conditionValue: "development",
			expected:       true,
		},
		{
			name:           "Wrong environment",
			envVars:        map[string]string{"OTTER_ENV": "production"},
			conditionValue: "development",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing environment variables
			os.Unsetenv("OTTER_ENV")
			os.Unsetenv("ENV")
			os.Unsetenv("NODE_ENV")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			condition := &Condition{
				Key:   "env",
				Value: tt.conditionValue,
			}

			result, err := evaluateCondition(condition)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}

			// Clean up
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestEvaluateCondition_Editor(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	vscodeDir := filepath.Join(tempDir, ".vscode")
	cursorDir := filepath.Join(tempDir, ".cursor")

	// Change to temp directory for testing
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name           string
		envVars        map[string]string
		createDirs     []string
		conditionValue string
		expected       bool
	}{
		{
			name:           "OTTER_EDITOR set",
			envVars:        map[string]string{"OTTER_EDITOR": "vim"},
			conditionValue: "vim",
			expected:       true,
		},
		{
			name:           "EDITOR fallback",
			envVars:        map[string]string{"EDITOR": "nano"},
			conditionValue: "nano",
			expected:       true,
		},
		{
			name:           "Auto-detect vscode",
			envVars:        map[string]string{},
			createDirs:     []string{vscodeDir},
			conditionValue: "vscode",
			expected:       true,
		},
		{
			name:           "Auto-detect cursor",
			envVars:        map[string]string{},
			createDirs:     []string{cursorDir},
			conditionValue: "cursor",
			expected:       true,
		},
		{
			name:           "No editor detected",
			envVars:        map[string]string{},
			conditionValue: "vscode",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing environment variables
			os.Unsetenv("OTTER_EDITOR")
			os.Unsetenv("EDITOR")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Create test directories
			for _, dir := range tt.createDirs {
				os.MkdirAll(dir, 0755)
			}

			condition := &Condition{
				Key:   "editor",
				Value: tt.conditionValue,
			}

			result, err := evaluateCondition(condition)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}

			// Clean up
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
			for _, dir := range tt.createDirs {
				os.RemoveAll(dir)
			}
		})
	}
}

func TestEvaluateCondition_CustomVariable(t *testing.T) {
	// Test custom environment variables
	os.Setenv("OTTER_CUSTOM", "myvalue")
	defer os.Unsetenv("OTTER_CUSTOM")

	condition := &Condition{
		Key:   "custom",
		Value: "myvalue",
	}

	result, err := evaluateCondition(condition)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !result {
		t.Errorf("Expected custom condition to be true")
	}

	// Test with wrong value
	wrongCondition := &Condition{
		Key:   "custom",
		Value: "wrongvalue",
	}

	result, err = evaluateCondition(wrongCondition)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result {
		t.Errorf("Expected custom condition to be false")
	}
}

func TestLayerShouldApplyLayer(t *testing.T) {
	tests := []struct {
		name      string
		layer     Layer
		envVars   map[string]string
		expected  bool
		expectErr bool
	}{
		{
			name: "No condition - always apply",
			layer: Layer{
				Repository: "test-repo",
				Target:     ".",
				Condition:  "",
			},
			expected: true,
		},
		{
			name: "Valid condition - should apply",
			layer: Layer{
				Repository: "test-repo",
				Target:     ".",
				Condition:  "env=development",
			},
			envVars:  map[string]string{},
			expected: true, // Default environment is development
		},
		{
			name: "Valid condition - should not apply",
			layer: Layer{
				Repository: "test-repo",
				Target:     ".",
				Condition:  "env=production",
			},
			envVars:  map[string]string{},
			expected: false, // Default environment is development
		},
		{
			name: "Invalid condition format",
			layer: Layer{
				Repository: "test-repo",
				Target:     ".",
				Condition:  "invalid-condition",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			result, err := tt.layer.ShouldApplyLayer()

			// Clean up environment
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterApplicableLayers(t *testing.T) {
	config := &OtterfileConfig{
		Layers: []Layer{
			{
				Repository: "base-layer",
				Target:     ".",
				Condition:  "",
			},
			{
				Repository: "dev-layer",
				Target:     ".",
				Condition:  "env=development",
			},
			{
				Repository: "prod-layer",
				Target:     ".",
				Condition:  "env=production",
			},
			{
				Repository: "os-layer",
				Target:     ".",
				Condition:  "os=" + runtime.GOOS,
			},
		},
	}

	// Test with development environment (default)
	applicableLayers, err := config.FilterApplicableLayers()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedCount := 3 // base, dev, and os layers
	if len(applicableLayers) != expectedCount {
		t.Errorf("Expected %d applicable layers, got %d", expectedCount, len(applicableLayers))
	}

	// Verify the right layers are included
	layerNames := make(map[string]bool)
	for _, layer := range applicableLayers {
		layerNames[layer.Repository] = true
	}

	expectedLayers := []string{"base-layer", "dev-layer", "os-layer"}
	for _, expectedLayer := range expectedLayers {
		if !layerNames[expectedLayer] {
			t.Errorf("Expected layer %s to be included", expectedLayer)
		}
	}

	// Test with production environment
	os.Setenv("OTTER_ENV", "production")
	defer os.Unsetenv("OTTER_ENV")

	applicableLayers, err = config.FilterApplicableLayers()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedCount = 3 // base, prod, and os layers
	if len(applicableLayers) != expectedCount {
		t.Errorf("Expected %d applicable layers, got %d", expectedCount, len(applicableLayers))
	}

	// Verify prod layer is included instead of dev
	layerNames = make(map[string]bool)
	for _, layer := range applicableLayers {
		layerNames[layer.Repository] = true
	}

	if !layerNames["prod-layer"] {
		t.Errorf("Expected prod-layer to be included")
	}

	if layerNames["dev-layer"] {
		t.Errorf("Expected dev-layer to be excluded")
	}
}

func TestParseOtterfileWithConditions(t *testing.T) {
	// Create a temporary Otterfile with conditional layers
	tempDir := t.TempDir()
	otterfilePath := filepath.Join(tempDir, "Otterfile")

	content := `# Test Otterfile with conditions
LAYER git@github.com:example/base.git
LAYER git@github.com:example/dev.git IF env=development
LAYER git@github.com:example/prod.git IF env=production TARGET production
LAYER git@github.com:example/vscode.git IF editor=vscode TARGET .vscode
LAYER git@github.com:example/macos.git IF os=darwin
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
	expectedLayers := []struct {
		repository string
		condition  string
		target     string
	}{
		{"git@github.com:example/base.git", "", "."},
		{"git@github.com:example/dev.git", "env=development", "."},
		{"git@github.com:example/prod.git", "env=production", "production"},
		{"git@github.com:example/vscode.git", "editor=vscode", ".vscode"},
		{"git@github.com:example/macos.git", "os=darwin", "."},
	}

	if len(config.Layers) != len(expectedLayers) {
		t.Errorf("Expected %d layers, got %d", len(expectedLayers), len(config.Layers))
	}

	for i, expected := range expectedLayers {
		if i >= len(config.Layers) {
			t.Errorf("Missing layer at index %d", i)
			continue
		}

		layer := config.Layers[i]
		if layer.Repository != expected.repository {
			t.Errorf("Layer %d: expected repository %s, got %s", i, expected.repository, layer.Repository)
		}
		if layer.Condition != expected.condition {
			t.Errorf("Layer %d: expected condition %s, got %s", i, expected.condition, layer.Condition)
		}
		if layer.Target != expected.target {
			t.Errorf("Layer %d: expected target %s, got %s", i, expected.target, layer.Target)
		}
	}
}
