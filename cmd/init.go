package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the current directory for otter",
	Long:  `Initialize the current directory by creating the .otter directory structure.`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	otterDir := filepath.Join(currentDir, ".otter")
	cacheDir := filepath.Join(otterDir, "cache")

	// Create .otter directory
	if err := os.MkdirAll(otterDir, 0755); err != nil {
		return fmt.Errorf("failed to create .otter directory: %w", err)
	}

	// Create .otter/cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create .otter/cache directory: %w", err)
	}

	// Create a basic .otterignore file if it doesn't exist
	ignorePath := filepath.Join(currentDir, ".otterignore")
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		defaultIgnore := `# Otter ignore file - specify files and patterns to ignore when merging layers
.git/
.otter/
node_modules/
*.log
*.tmp
.DS_Store
`
		if err := os.WriteFile(ignorePath, []byte(defaultIgnore), 0644); err != nil {
			return fmt.Errorf("failed to create .otterignore file: %w", err)
		}
		fmt.Println("Created .otterignore file")
	}

	// Create a sample Otterfile if it doesn't exist
	otterfilePath := filepath.Join(currentDir, "Otterfile")
	if _, err := os.Stat(otterfilePath); os.IsNotExist(err) {
		sampleOtterfile := `# Otterfile - define layers to pull from git repositories
# Syntax: LAYER <git-repo-url> [TARGET <target-path>]
# Example:
# LAYER git@github.com:otter-layers/go-cobra-cli.git
# LAYER git@github.com:otter-layers/cursor-go-rules.git TARGET .cursor/rules
`
		if err := os.WriteFile(otterfilePath, []byte(sampleOtterfile), 0644); err != nil {
			return fmt.Errorf("failed to create sample Otterfile: %w", err)
		}
		fmt.Println("Created sample Otterfile")
	}

	fmt.Printf("Otter initialized successfully in %s\n", currentDir)
	fmt.Println("Created directories:")
	fmt.Printf("  %s\n", otterDir)
	fmt.Printf("  %s\n", cacheDir)

	return nil
}
