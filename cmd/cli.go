package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cliCmd = &cobra.Command{
	Use:   "otter",
	Short: "Otter simplifies development environment setup through layered templates",
	Long: `Otter is a tool that simplifies development environment setup through a layer concept 
that pulls other templates containing files into the project it's run inside of.`,
}

// Execute runs the root command.
func Execute() {
	if err := cliCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cliCmd.AddCommand(initCmd)
	cliCmd.AddCommand(buildCmd)
}
