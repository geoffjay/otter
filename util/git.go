package util

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

// GitOperations handles all git-related operations
type GitOperations struct {
	cacheDir string
}

// NewGitOperations creates a new GitOperations instance
func NewGitOperations(cacheDir string) *GitOperations {
	return &GitOperations{
		cacheDir: cacheDir,
	}
}

// CloneOrUpdateLayer clones a git repository to the cache directory or updates it if it already exists
func (g *GitOperations) CloneOrUpdateLayer(repoURL string) (string, error) {
	// Create a unique directory name based on the repository URL
	repoName := g.GetRepoDirectoryName(repoURL)
	localPath := filepath.Join(g.cacheDir, repoName)

	// Check if repository already exists
	if _, err := os.Stat(filepath.Join(localPath, ".git")); err == nil {
		// Repository exists, try to update it
		fmt.Printf("Updating layer: %s\n", repoURL)
		return localPath, g.updateRepository(localPath)
	}

	// Repository doesn't exist, clone it
	fmt.Printf("Cloning layer: %s\n", repoURL)
	return localPath, g.cloneRepository(repoURL, localPath)
}

// cloneRepository clones a git repository to the specified path
func (g *GitOperations) cloneRepository(repoURL, localPath string) error {
	// Ensure the cache directory exists
	if err := os.MkdirAll(g.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Clone the repository
	_, err := git.PlainClone(localPath, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", repoURL, err)
	}

	return nil
}

// updateRepository updates an existing git repository
func (g *GitOperations) updateRepository(localPath string) error {
	// Open the existing repository
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return fmt.Errorf("failed to open repository at %s: %w", localPath, err)
	}

	// Get the working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Pull the latest changes
	err = worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})

	// If the error is "already up-to-date", that's fine
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull updates: %w", err)
	}

	if err == git.NoErrAlreadyUpToDate {
		fmt.Println("  Already up-to-date")
	}

	return nil
}

// getRepoDirectoryName creates a unique directory name for a repository URL
func (g *GitOperations) GetRepoDirectoryName(repoURL string) string {
	// Remove common prefixes and suffixes
	name := strings.TrimSuffix(repoURL, ".git")

	// Extract the repository name from different URL formats
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}

	if strings.Contains(name, ":") {
		parts := strings.Split(name, ":")
		name = parts[len(parts)-1]
		// Handle case like git@github.com:user/repo
		if strings.Contains(name, "/") {
			parts = strings.Split(name, "/")
			name = parts[len(parts)-1]
		}
	}

	// Create a hash of the full URL to ensure uniqueness while keeping a readable name
	hash := sha256.Sum256([]byte(repoURL))
	hashStr := fmt.Sprintf("%x", hash[:4]) // Use first 4 bytes of hash

	return fmt.Sprintf("%s-%s", name, hashStr)
}

// GetRepositoryCommit gets the current commit hash of a repository
func (g *GitOperations) GetRepositoryCommit(localPath string) (string, error) {
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return ref.Hash().String(), nil
}
