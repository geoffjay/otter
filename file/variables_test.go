package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseVarCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedKey string
		expectedVal string
		expectError bool
	}{
		{
			name:        "Simple variable",
			args:        []string{"PROJECT_NAME=my-project"},
			expectedKey: "PROJECT_NAME",
			expectedVal: "my-project",
			expectError: false,
		},
		{
			name:        "Variable with spaces in value",
			args:        []string{"DESCRIPTION=My", "awesome", "project"},
			expectedKey: "DESCRIPTION",
			expectedVal: "My awesome project",
			expectError: false,
		},
		{
			name:        "Variable with equals in value",
			args:        []string{"DATABASE_URL=postgres://user:pass@host/db?ssl=require"},
			expectedKey: "DATABASE_URL",
			expectedVal: "postgres://user:pass@host/db?ssl=require",
			expectError: false,
		},
		{
			name:        "Variable with spaces around equals",
			args:        []string{"VERSION", "=", "1.0.0"},
			expectedKey: "VERSION",
			expectedVal: "1.0.0",
			expectError: false,
		},
		{
			name:        "Empty value",
			args:        []string{"EMPTY="},
			expectedKey: "EMPTY",
			expectedVal: "",
			expectError: false,
		},
		{
			name:        "No arguments",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "Missing equals",
			args:        []string{"INVALID", "VALUE"},
			expectError: true,
		},
		{
			name:        "Empty key",
			args:        []string{"=value"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OtterfileConfig{
				Variables: make(map[string]string),
				Layers:    make([]Layer, 0),
			}

			err := parseVarCommand(tt.args, config)

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

			if val, exists := config.Variables[tt.expectedKey]; !exists {
				t.Errorf("Expected variable %s to be set", tt.expectedKey)
			} else if val != tt.expectedVal {
				t.Errorf("Expected value %s, got %s", tt.expectedVal, val)
			}
		})
	}
}

func TestSubstituteVariables(t *testing.T) {
	variables := map[string]string{
		"PROJECT_NAME": "my-api",
		"VERSION":      "1.21",
		"DATABASE":     "postgres",
		"EMPTY":        "",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No substitution",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "Single variable",
			input:    "git@github.com:otter-layers/${PROJECT_NAME}.git",
			expected: "git@github.com:otter-layers/my-api.git",
		},
		{
			name:     "Multiple variables",
			input:    "src/${PROJECT_NAME}/v${VERSION}",
			expected: "src/my-api/v1.21",
		},
		{
			name:     "Variable in middle",
			input:    "git@github.com:otter-layers/${DATABASE}-setup.git",
			expected: "git@github.com:otter-layers/postgres-setup.git",
		},
		{
			name:     "Same variable multiple times",
			input:    "${PROJECT_NAME}/${PROJECT_NAME}/config",
			expected: "my-api/my-api/config",
		},
		{
			name:     "Variable with empty value",
			input:    "prefix-${EMPTY}-suffix",
			expected: "prefix--suffix",
		},
		{
			name:     "Non-existent variable",
			input:    "${NON_EXISTENT}",
			expected: "${NON_EXISTENT}",
		},
		{
			name:     "Malformed variable reference",
			input:    "${MALFORMED",
			expected: "${MALFORMED",
		},
		{
			name:     "Mixed variables and non-existent",
			input:    "${PROJECT_NAME}/${NON_EXISTENT}/${VERSION}",
			expected: "my-api/${NON_EXISTENT}/1.21",
		},
		{
			name:     "Variable at start",
			input:    "${PROJECT_NAME}/path",
			expected: "my-api/path",
		},
		{
			name:     "Variable at end",
			input:    "path/${PROJECT_NAME}",
			expected: "path/my-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteVariables(tt.input, variables)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSubstituteVariables_EnvironmentFallback(t *testing.T) {
	// Test environment variable fallback
	os.Setenv("OTTER_FRAMEWORK", "react")
	os.Setenv("DIRECT_VAR", "direct-value")
	defer func() {
		os.Unsetenv("OTTER_FRAMEWORK")
		os.Unsetenv("DIRECT_VAR")
	}()

	variables := map[string]string{
		"PROJECT_NAME": "my-app",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Otterfile variable takes precedence",
			input:    "${PROJECT_NAME}",
			expected: "my-app",
		},
		{
			name:     "Environment variable with OTTER_ prefix",
			input:    "${FRAMEWORK}",
			expected: "react",
		},
		{
			name:     "Direct environment variable",
			input:    "${DIRECT_VAR}",
			expected: "direct-value",
		},
		{
			name:     "Mixed sources",
			input:    "${PROJECT_NAME}-${FRAMEWORK}-${DIRECT_VAR}",
			expected: "my-app-react-direct-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteVariables(tt.input, variables)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseLayerCommand_WithTemplate(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedRepo     string
		expectedTarget   string
		expectedTemplate map[string]string
		expectError      bool
	}{
		{
			name:             "Layer with single template variable",
			args:             []string{"git@github.com:example/dockerfile.git", "TEMPLATE", "go_version=1.21"},
			expectedRepo:     "git@github.com:example/dockerfile.git",
			expectedTarget:   ".",
			expectedTemplate: map[string]string{"go_version": "1.21"},
			expectError:      false,
		},
		{
			name:             "Layer with multiple template variables",
			args:             []string{"git@github.com:example/template.git", "TEMPLATE", "name=myapp", "version=1.0", "env=prod"},
			expectedRepo:     "git@github.com:example/template.git",
			expectedTarget:   ".",
			expectedTemplate: map[string]string{"name": "myapp", "version": "1.0", "env": "prod"},
			expectError:      false,
		},
		{
			name:             "Layer with TARGET and TEMPLATE",
			args:             []string{"git@github.com:example/config.git", "TARGET", "configs", "TEMPLATE", "database=postgres"},
			expectedRepo:     "git@github.com:example/config.git",
			expectedTarget:   "configs",
			expectedTemplate: map[string]string{"database": "postgres"},
			expectError:      false,
		},
		{
			name:        "Layer with TEMPLATE but no variables",
			args:        []string{"git@github.com:example/template.git", "TEMPLATE"},
			expectError: true,
		},
		{
			name:             "Layer with TEMPLATE and IF condition",
			args:             []string{"git@github.com:example/template.git", "TEMPLATE", "env=dev", "IF", "env=development"},
			expectedRepo:     "git@github.com:example/template.git",
			expectedTarget:   ".",
			expectedTemplate: map[string]string{"env": "dev"},
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OtterfileConfig{
				Variables: make(map[string]string),
				Layers:    make([]Layer, 0),
			}

			err := parseLayerCommand(tt.args, config)

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

			if len(config.Layers) != 1 {
				t.Errorf("Expected 1 layer, got %d", len(config.Layers))
				return
			}

			layer := config.Layers[0]
			if layer.Repository != tt.expectedRepo {
				t.Errorf("Expected repository %s, got %s", tt.expectedRepo, layer.Repository)
			}

			if layer.Target != tt.expectedTarget {
				t.Errorf("Expected target %s, got %s", tt.expectedTarget, layer.Target)
			}

			if len(layer.Template) != len(tt.expectedTemplate) {
				t.Errorf("Expected %d template variables, got %d", len(tt.expectedTemplate), len(layer.Template))
			}

			for key, expectedVal := range tt.expectedTemplate {
				if actualVal, exists := layer.Template[key]; !exists {
					t.Errorf("Expected template variable %s to exist", key)
				} else if actualVal != expectedVal {
					t.Errorf("Expected template variable %s=%s, got %s", key, expectedVal, actualVal)
				}
			}
		})
	}
}

func TestParseOtterfileWithVariables(t *testing.T) {
	// Create a temporary Otterfile with variables and templating
	tempDir := t.TempDir()
	otterfilePath := filepath.Join(tempDir, "Otterfile")

	content := `# Test Otterfile with variables and templating
VAR PROJECT_NAME=my-api
VAR GO_VERSION=1.21
VAR DATABASE=postgres

# Use variables in layers
LAYER git@github.com:otter-layers/go-project.git TARGET src/${PROJECT_NAME}
LAYER git@github.com:otter-layers/${DATABASE}-setup.git
LAYER git@github.com:otter-layers/dockerfile.git TEMPLATE go_version=${GO_VERSION} project=${PROJECT_NAME}
LAYER git@github.com:otter-layers/config.git TARGET config IF env=production TEMPLATE db=${DATABASE}
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

	// Test variables
	expectedVars := map[string]string{
		"PROJECT_NAME": "my-api",
		"GO_VERSION":   "1.21",
		"DATABASE":     "postgres",
	}

	if len(config.Variables) != len(expectedVars) {
		t.Errorf("Expected %d variables, got %d", len(expectedVars), len(config.Variables))
	}

	for key, expectedVal := range expectedVars {
		if actualVal, exists := config.Variables[key]; !exists {
			t.Errorf("Expected variable %s to exist", key)
		} else if actualVal != expectedVal {
			t.Errorf("Expected variable %s=%s, got %s", key, expectedVal, actualVal)
		}
	}

	// Test layers with variable substitution
	expectedLayers := []struct {
		repository string
		target     string
		condition  string
		template   map[string]string
	}{
		{
			repository: "git@github.com:otter-layers/go-project.git",
			target:     "src/my-api",
			condition:  "",
			template:   map[string]string{},
		},
		{
			repository: "git@github.com:otter-layers/postgres-setup.git",
			target:     ".",
			condition:  "",
			template:   map[string]string{},
		},
		{
			repository: "git@github.com:otter-layers/dockerfile.git",
			target:     ".",
			condition:  "",
			template:   map[string]string{"go_version": "1.21", "project": "my-api"},
		},
		{
			repository: "git@github.com:otter-layers/config.git",
			target:     "config",
			condition:  "env=production",
			template:   map[string]string{"db": "postgres"},
		},
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
		if layer.Target != expected.target {
			t.Errorf("Layer %d: expected target %s, got %s", i, expected.target, layer.Target)
		}
		if layer.Condition != expected.condition {
			t.Errorf("Layer %d: expected condition %s, got %s", i, expected.condition, layer.Condition)
		}

		if len(layer.Template) != len(expected.template) {
			t.Errorf("Layer %d: expected %d template variables, got %d", i, len(expected.template), len(layer.Template))
		}

		for key, expectedVal := range expected.template {
			if actualVal, exists := layer.Template[key]; !exists {
				t.Errorf("Layer %d: expected template variable %s to exist", i, key)
			} else if actualVal != expectedVal {
				t.Errorf("Layer %d: expected template variable %s=%s, got %s", i, key, expectedVal, actualVal)
			}
		}
	}
}

func TestComplexVariableSubstitution(t *testing.T) {
	// Test complex scenarios with variable substitution
	tempDir := t.TempDir()
	otterfilePath := filepath.Join(tempDir, "Otterfile")

	// Set environment variables for testing
	os.Setenv("OTTER_TEAM", "backend")
	os.Setenv("GITHUB_ORG", "mycompany")
	defer func() {
		os.Unsetenv("OTTER_TEAM")
		os.Unsetenv("GITHUB_ORG")
	}()

	content := `# Complex variable substitution test
VAR SERVICE_NAME=auth-service
VAR VERSION=v2.1.0
VAR BASE_PATH=services/${SERVICE_NAME}

# Mix of Otterfile variables and environment variables
LAYER git@github.com:${GITHUB_ORG}/${SERVICE_NAME}.git@${VERSION} TARGET ${BASE_PATH}
LAYER git@github.com:${GITHUB_ORG}/common-${TEAM}-tools.git TARGET tools
LAYER git@github.com:templates/config.git TARGET ${BASE_PATH}/config TEMPLATE service=${SERVICE_NAME} version=${VERSION} team=${TEAM}
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

	// Expected results after variable substitution
	expectedLayers := []struct {
		repository string
		target     string
		template   map[string]string
	}{
		{
			repository: "git@github.com:mycompany/auth-service.git@v2.1.0",
			target:     "services/auth-service",
			template:   map[string]string{},
		},
		{
			repository: "git@github.com:mycompany/common-backend-tools.git",
			target:     "tools",
			template:   map[string]string{},
		},
		{
			repository: "git@github.com:templates/config.git",
			target:     "services/auth-service/config",
			template:   map[string]string{"service": "auth-service", "version": "v2.1.0", "team": "backend"},
		},
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
		if layer.Target != expected.target {
			t.Errorf("Layer %d: expected target %s, got %s", i, expected.target, layer.Target)
		}

		if len(layer.Template) != len(expected.template) {
			t.Errorf("Layer %d: expected %d template variables, got %d", i, len(expected.template), len(layer.Template))
		}

		for key, expectedVal := range expected.template {
			if actualVal, exists := layer.Template[key]; !exists {
				t.Errorf("Layer %d: expected template variable %s to exist", i, key)
			} else if actualVal != expectedVal {
				t.Errorf("Layer %d: expected template variable %s=%s, got %s", i, key, expectedVal, actualVal)
			}
		}
	}
}
