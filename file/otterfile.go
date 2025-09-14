package file

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
)

// Layer represents a single layer definition from the Otterfile
type Layer struct {
	Repository string
	Target     string            // Optional target directory, defaults to root
	Condition  string            // Optional condition for applying the layer (e.g., "env=development")
	Template   map[string]string // Optional template variables to pass to the layer
}

// Condition represents a parsed condition for layer application
type Condition struct {
	Key   string
	Value string
}

// OtterfileConfig holds the parsed configuration from Otterfile/Envfile
type OtterfileConfig struct {
	Variables map[string]string // Variables defined with VAR command
	Layers    []Layer
}

// ParseOtterfile reads and parses an Otterfile or Envfile
func ParseOtterfile(filename string) (*OtterfileConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filename, err)
	}
	defer file.Close()

	config := &OtterfileConfig{
		Variables: make(map[string]string),
		Layers:    make([]Layer, 0),
	}

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if err := parseLine(line, config, lineNumber); err != nil {
			return nil, fmt.Errorf("error on line %d: %w", lineNumber, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}

	return config, nil
}

// parseLine parses a single line from the Otterfile
func parseLine(line string, config *OtterfileConfig, lineNumber int) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToUpper(parts[0])

	switch command {
	case "VAR":
		return parseVarCommand(parts[1:], config)
	case "LAYER":
		return parseLayerCommand(parts[1:], config)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// parseVarCommand parses a VAR command
func parseVarCommand(args []string, config *OtterfileConfig) error {
	if len(args) == 0 {
		return fmt.Errorf("VAR command requires a variable definition")
	}

	// Join all args back into a single string in case the value contains spaces
	varDef := strings.Join(args, " ")

	// Split on the first '=' to separate key and value
	parts := strings.SplitN(varDef, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("VAR command must be in format 'KEY=VALUE', got: %s", varDef)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if key == "" {
		return fmt.Errorf("variable name cannot be empty")
	}

	// Apply variable substitution to the value using previously defined variables
	resolvedValue := substituteVariables(value, config.Variables)
	config.Variables[key] = resolvedValue
	return nil
}

// parseLayerCommand parses a LAYER command
func parseLayerCommand(args []string, config *OtterfileConfig) error {
	if len(args) == 0 {
		return fmt.Errorf("LAYER command requires a repository URL")
	}

	layer := Layer{
		Repository: args[0],
		Target:     ".", // Default to current directory
		Template:   make(map[string]string),
	}

	// Parse optional TARGET, IF, and TEMPLATE arguments
	for i := 1; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		switch arg {
		case "TARGET":
			if i+1 >= len(args) {
				return fmt.Errorf("TARGET requires a path argument")
			}
			layer.Target = args[i+1]
			i++ // Skip the next argument as it's the target path
		case "IF":
			if i+1 >= len(args) {
				return fmt.Errorf("IF requires a condition argument")
			}
			layer.Condition = args[i+1]
			i++ // Skip the next argument as it's the condition
		case "TEMPLATE":
			if i+1 >= len(args) {
				return fmt.Errorf("TEMPLATE requires template variable assignments")
			}
			// Parse template variables (key=value format, possibly multiple)
			for j := i + 1; j < len(args); j++ {
				if strings.Contains(args[j], "=") {
					parts := strings.SplitN(args[j], "=", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						layer.Template[key] = value
					}
				} else {
					// This argument doesn't contain '=', so it's likely a different argument type
					i = j - 1 // Back up one step so the outer loop processes this argument
					break
				}
				i = j // Move the outer loop index forward
			}
		default:
			return fmt.Errorf("unknown LAYER argument: %s", args[i])
		}
	}

	// Apply variable substitution to repository URL and target
	layer.Repository = substituteVariables(layer.Repository, config.Variables)
	layer.Target = substituteVariables(layer.Target, config.Variables)

	// Apply variable substitution to template values
	for key, value := range layer.Template {
		layer.Template[key] = substituteVariables(value, config.Variables)
	}

	config.Layers = append(config.Layers, layer)
	return nil
}

// substituteVariables replaces ${VAR_NAME} placeholders with actual variable values
func substituteVariables(text string, variables map[string]string) string {
	// Regular expression to match ${VAR_NAME} patterns
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(text, func(match string) string {
		// Extract the variable name from ${VAR_NAME}
		varName := match[2 : len(match)-1] // Remove ${ and }

		// First check custom variables defined in Otterfile
		if value, exists := variables[varName]; exists {
			return value
		}

		// Then check environment variables (with OTTER_ prefix)
		envVarName := "OTTER_" + strings.ToUpper(varName)
		if value := os.Getenv(envVarName); value != "" {
			return value
		}

		// Finally check direct environment variables
		if value := os.Getenv(varName); value != "" {
			return value
		}

		// If variable is not found, return the original placeholder
		return match
	})
}

// FindOtterfile looks for Otterfile or Envfile in the current directory
func FindOtterfile() (string, error) {
	candidates := []string{"Otterfile", "Envfile"}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no Otterfile or Envfile found in current directory")
}

// parseCondition parses a condition string (e.g., "env=development")
func parseCondition(conditionStr string) (*Condition, error) {
	if conditionStr == "" {
		return nil, fmt.Errorf("condition cannot be empty")
	}

	parts := strings.SplitN(conditionStr, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("condition must be in format 'key=value', got: %s", conditionStr)
	}

	return &Condition{
		Key:   strings.TrimSpace(parts[0]),
		Value: strings.TrimSpace(parts[1]),
	}, nil
}

// evaluateCondition evaluates a condition against the current environment
func evaluateCondition(condition *Condition) (bool, error) {
	if condition == nil {
		return true, nil
	}

	switch condition.Key {
	case "os":
		return condition.Value == runtime.GOOS, nil
	case "arch":
		return condition.Value == runtime.GOARCH, nil
	case "env", "environment":
		envValue := os.Getenv("OTTER_ENV")
		if envValue == "" {
			envValue = os.Getenv("ENV")
		}
		if envValue == "" {
			envValue = os.Getenv("NODE_ENV")
		}
		if envValue == "" {
			envValue = "development" // Default to development
		}
		return condition.Value == envValue, nil
	case "editor":
		editorValue := os.Getenv("OTTER_EDITOR")
		if editorValue == "" {
			editorValue = os.Getenv("EDITOR")
		}
		if editorValue == "" {
			// Try to detect common editors
			if _, err := os.Stat(".vscode"); err == nil {
				editorValue = "vscode"
			} else if _, err := os.Stat(".cursor"); err == nil {
				editorValue = "cursor"
			}
		}
		return condition.Value == editorValue, nil
	default:
		// Check for custom environment variables
		envVarName := "OTTER_" + strings.ToUpper(condition.Key)
		envValue := os.Getenv(envVarName)
		return condition.Value == envValue, nil
	}
}

// ShouldApplyLayer determines if a layer should be applied based on its condition
func (l *Layer) ShouldApplyLayer() (bool, error) {
	if l.Condition == "" {
		return true, nil // No condition means always apply
	}

	condition, err := parseCondition(l.Condition)
	if err != nil {
		return false, fmt.Errorf("failed to parse condition '%s': %w", l.Condition, err)
	}

	return evaluateCondition(condition)
}

// FilterApplicableLayers filters layers based on their conditions
func (config *OtterfileConfig) FilterApplicableLayers() ([]Layer, error) {
	var applicableLayers []Layer

	for _, layer := range config.Layers {
		shouldApply, err := layer.ShouldApplyLayer()
		if err != nil {
			return nil, fmt.Errorf("error evaluating condition for layer %s: %w", layer.Repository, err)
		}

		if shouldApply {
			applicableLayers = append(applicableLayers, layer)
		}
	}

	return applicableLayers, nil
}
