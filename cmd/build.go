package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/geoffjay/otter/file"
	"github.com/geoffjay/otter/util"

	"github.com/spf13/cobra"
)

var (
	buildFile string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the development environment by applying layers",
	Long: `Build the development environment by reading the Otterfile/Envfile and applying 
all defined layers to the current project.`,
	RunE: runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&buildFile, "file", "f", "", "Specify the Otterfile/Envfile to use (default: auto-detect)")
}

func runBuild(cmd *cobra.Command, args []string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if .otter directory exists
	otterDir := filepath.Join(currentDir, ".otter")
	if _, err := os.Stat(otterDir); os.IsNotExist(err) {
		return fmt.Errorf(".otter directory not found. Please run 'otter init' first")
	}

	cacheDir := filepath.Join(otterDir, "cache")

	// Find Otterfile if not specified
	var otterfilePath string
	if buildFile != "" {
		otterfilePath = buildFile
	} else {
		otterfilePath, err = file.FindOtterfile()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Using configuration file: %s\n", otterfilePath)

	// Parse the Otterfile
	config, err := file.ParseOtterfile(otterfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", otterfilePath, err)
	}

	if len(config.Layers) == 0 {
		fmt.Println("No layers defined in configuration file.")
		return nil
	}

	// Filter applicable layers based on conditions
	applicableLayers, err := config.FilterApplicableLayers()
	if err != nil {
		return fmt.Errorf("failed to filter applicable layers: %w", err)
	}

	if len(applicableLayers) == 0 {
		fmt.Println("No layers are applicable for current environment.")
		return nil
	}

	if len(applicableLayers) < len(config.Layers) {
		fmt.Printf("Found %d layer(s), applying %d layer(s) based on conditions:\n", len(config.Layers), len(applicableLayers))
	} else {
		fmt.Printf("Found %d layer(s) to process:\n", len(applicableLayers))
	}

	// Initialize git and file operations
	gitOps := util.NewGitOperations(cacheDir)
	fileOps := util.NewFileOperations()

	// Load ignore patterns
	if err := fileOps.LoadIgnorePatterns(currentDir); err != nil {
		return fmt.Errorf("failed to load ignore patterns: %w", err)
	}

	// Process each applicable layer
	for i, layer := range applicableLayers {
		fmt.Printf("\n[%d/%d] Processing layer: %s\n", i+1, len(applicableLayers), layer.Repository)
		if layer.Condition != "" {
			fmt.Printf("  Condition: %s\n", layer.Condition)
		}

		// Clone or update the layer
		layerPath, err := gitOps.CloneOrUpdateLayer(layer.Repository)
		if err != nil {
			return fmt.Errorf("failed to process layer %s: %w", layer.Repository, err)
		}

		// Determine target directory
		var targetPath string
		if layer.Target == "." {
			targetPath = currentDir
		} else {
			targetPath = filepath.Join(currentDir, layer.Target)
		}

		fmt.Printf("  Target directory: %s\n", targetPath)

		// Copy files from layer to target
		if err := fileOps.CopyLayer(layerPath, targetPath, currentDir); err != nil {
			return fmt.Errorf("failed to copy layer files: %w", err)
		}

		// Show commit information
		commit, err := gitOps.GetRepositoryCommit(layerPath)
		if err == nil {
			fmt.Printf("  Layer commit: %s\n", commit[:8])
		}

		fmt.Printf("  ✓ Layer applied successfully\n")
	}

	fmt.Printf("\n🎉 Build completed successfully! Applied %d layer(s).\n", len(config.Layers))

	return nil
}
