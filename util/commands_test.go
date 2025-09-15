package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommandExecutor(t *testing.T) {
	tempDir := t.TempDir()
	executor := NewCommandExecutor(tempDir)

	t.Run("ExecuteCommand - success", func(t *testing.T) {
		err := executor.ExecuteCommand("echo 'test'")
		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
	})

	t.Run("ExecuteCommand - failure", func(t *testing.T) {
		err := executor.ExecuteCommand("nonexistent-command")
		if err == nil {
			t.Errorf("Expected error for nonexistent command, got success")
		}
	})

	t.Run("ExecuteCommands - empty slice", func(t *testing.T) {
		err := executor.ExecuteCommands([]string{}, "test")
		if err != nil {
			t.Errorf("Expected success for empty commands, got error: %v", err)
		}
	})

	t.Run("ExecuteCommands - multiple commands", func(t *testing.T) {
		// Create a test file to verify commands ran
		testFile := filepath.Join(tempDir, "test.txt")

		commands := []string{
			"echo 'first' > " + testFile,
			"echo 'second' >> " + testFile,
		}

		err := executor.ExecuteCommands(commands, "test")
		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}

		// Verify the file was created and has content
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Errorf("Failed to read test file: %v", err)
		}

		expectedContent := "first\nsecond\n"
		if string(content) != expectedContent {
			t.Errorf("Expected file content '%s', got '%s'", expectedContent, string(content))
		}
	})

	t.Run("ExecuteCommands - stop on first failure", func(t *testing.T) {
		commands := []string{
			"echo 'This should run'",
			"nonexistent-command",
			"echo 'This should not run'",
		}

		err := executor.ExecuteCommands(commands, "test")
		if err == nil {
			t.Errorf("Expected error when command fails, got success")
		}
	})

	t.Run("ExecuteCommandsWithCleanup - success case", func(t *testing.T) {
		commands := []string{"echo 'success'"}
		cleanup := []string{"echo 'cleanup should not run'"}

		err := executor.ExecuteCommandsWithCleanup(commands, "test", cleanup)
		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
	})

	t.Run("ExecuteCommandsWithCleanup - failure with cleanup", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "cleanup-test.txt")

		commands := []string{"nonexistent-command"}
		cleanup := []string{"echo 'cleanup ran' > " + testFile}

		err := executor.ExecuteCommandsWithCleanup(commands, "test", cleanup)
		if err == nil {
			t.Errorf("Expected error when command fails, got success")
		}

		// Verify cleanup ran
		if _, err := os.Stat(testFile); err != nil {
			t.Errorf("Cleanup command did not run - test file not created")
		}
	})
}

func TestCommandExecutorWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create executor with subdirectory as working dir
	executor := NewCommandExecutor(subDir)

	// Create a file in the working directory using relative path
	err = executor.ExecuteCommand("touch test-file.txt")
	if err != nil {
		t.Errorf("Failed to execute command in working directory: %v", err)
	}

	// Verify file was created in the subdirectory, not temp root
	expectedFile := filepath.Join(subDir, "test-file.txt")
	if _, err := os.Stat(expectedFile); err != nil {
		t.Errorf("File was not created in working directory: %v", err)
	}

	// Verify file was NOT created in temp root
	unexpectedFile := filepath.Join(tempDir, "test-file.txt")
	if _, err := os.Stat(unexpectedFile); err == nil {
		t.Errorf("File was incorrectly created in wrong directory")
	}
}
