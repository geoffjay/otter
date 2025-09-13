package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Layer represents a single layer definition from the Otterfile
type Layer struct {
	Repository string
	Target     string // Optional target directory, defaults to root
}

// OtterfileConfig holds the parsed configuration from Otterfile/Envfile
type OtterfileConfig struct {
	Layers []Layer
}

// ParseOtterfile reads and parses an Otterfile or Envfile
func ParseOtterfile(filename string) (*OtterfileConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filename, err)
	}
	defer file.Close()

	config := &OtterfileConfig{
		Layers: make([]Layer, 0),
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
	case "LAYER":
		return parseLayerCommand(parts[1:], config)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// parseLayerCommand parses a LAYER command
func parseLayerCommand(args []string, config *OtterfileConfig) error {
	if len(args) == 0 {
		return fmt.Errorf("LAYER command requires a repository URL")
	}

	layer := Layer{
		Repository: args[0],
		Target:     ".", // Default to current directory
	}

	// Parse optional TARGET argument
	for i := 1; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		switch arg {
		case "TARGET":
			if i+1 >= len(args) {
				return fmt.Errorf("TARGET requires a path argument")
			}
			layer.Target = args[i+1]
			i++ // Skip the next argument as it's the target path
		default:
			return fmt.Errorf("unknown LAYER argument: %s", args[i])
		}
	}

	config.Layers = append(config.Layers, layer)
	return nil
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
