package util

import (
	"fmt"
	"os"
	"os/exec"
)

// CommandExecutor handles executing shell commands for hooks
type CommandExecutor struct {
	WorkingDir string
}

// NewCommandExecutor creates a new CommandExecutor
func NewCommandExecutor(workingDir string) *CommandExecutor {
	return &CommandExecutor{
		WorkingDir: workingDir,
	}
}

// ExecuteCommands executes a list of shell commands in sequence
func (c *CommandExecutor) ExecuteCommands(commands []string, context string) error {
	if len(commands) == 0 {
		return nil
	}

	fmt.Printf("  Executing %s commands:\n", context)

	for i, command := range commands {
		fmt.Printf("    [%d/%d] %s\n", i+1, len(commands), command)

		if err := c.ExecuteCommand(command); err != nil {
			return fmt.Errorf("failed to execute %s command '%s': %w", context, command, err)
		}
	}

	return nil
}

// ExecuteCommand executes a single shell command
func (c *CommandExecutor) ExecuteCommand(command string) error {
	if command == "" {
		return fmt.Errorf("empty command")
	}

	// Use shell to execute the command to support shell features like redirection, pipes, etc.
	var cmd *exec.Cmd

	// Detect shell based on OS
	if os.Getenv("SHELL") != "" {
		cmd = exec.Command(os.Getenv("SHELL"), "-c", command)
	} else {
		// Default to /bin/sh on Unix-like systems, cmd.exe on Windows
		cmd = exec.Command("/bin/sh", "-c", command)
	}

	cmd.Dir = c.WorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ExecuteCommandsWithCleanup executes commands and runs cleanup on error
func (c *CommandExecutor) ExecuteCommandsWithCleanup(commands []string, context string, onError []string) error {
	err := c.ExecuteCommands(commands, context)
	if err != nil && len(onError) > 0 {
		fmt.Printf("  Error occurred, running cleanup commands:\n")
		// Execute cleanup commands but don't return their errors (just log them)
		cleanupErr := c.ExecuteCommands(onError, "cleanup")
		if cleanupErr != nil {
			fmt.Printf("  Warning: Cleanup commands failed: %v\n", cleanupErr)
		}
	}
	return err
}
