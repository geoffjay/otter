package util

import (
	"crypto/sha256"
	"fmt"
	"net/url"
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

// CloneOrUpdateLayer clones a git repository to the cache directory, updates it if it already exists,
// or returns the path directly for local layers
func (g *GitOperations) CloneOrUpdateLayer(repoURL string) (string, error) {
	// Check if this is a local layer
	if g.isLocalLayer(repoURL) {
		return g.handleLocalLayer(repoURL)
	}

	// Handle remote git repository
	return g.handleRemoteRepository(repoURL)
}

// isLocalLayer checks if the repository URL refers to a local directory
func (g *GitOperations) isLocalLayer(repoURL string) bool {
	// Check for relative paths
	if strings.HasPrefix(repoURL, "./") || strings.HasPrefix(repoURL, "../") {
		return true
	}

	// Check for absolute paths
	if strings.HasPrefix(repoURL, "/") {
		return true
	}

	// Check for file:// URI scheme
	if strings.HasPrefix(repoURL, "file://") {
		return true
	}

	// Check for Windows paths (C:\ etc.)
	if len(repoURL) >= 3 && repoURL[1] == ':' && (repoURL[2] == '\\' || repoURL[2] == '/') {
		return true
	}

	return false
}

// handleLocalLayer processes a local directory layer
func (g *GitOperations) handleLocalLayer(repoURL string) (string, error) {
	var localPath string

	// Handle file:// URI scheme
	if strings.HasPrefix(repoURL, "file://") {
		parsedURL, err := url.Parse(repoURL)
		if err != nil {
			return "", fmt.Errorf("failed to parse file:// URL %s: %w", repoURL, err)
		}
		localPath = parsedURL.Path
	} else {
		localPath = repoURL
	}

	// Convert to absolute path if it's relative
	if !filepath.IsAbs(localPath) {
		absPath, err := filepath.Abs(localPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve absolute path for %s: %w", localPath, err)
		}
		localPath = absPath
	}

	// Verify the directory exists
	if _, err := os.Stat(localPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("local layer directory does not exist: %s", localPath)
		}
		return "", fmt.Errorf("failed to access local layer directory %s: %w", localPath, err)
	}

	// Verify it's actually a directory
	if stat, err := os.Stat(localPath); err != nil {
		return "", fmt.Errorf("failed to stat local layer path %s: %w", localPath, err)
	} else if !stat.IsDir() {
		return "", fmt.Errorf("local layer path is not a directory: %s", localPath)
	}

	fmt.Printf("Using local layer: %s\n", localPath)
	return localPath, nil
}

// handleRemoteRepository processes a remote git repository (existing logic)
func (g *GitOperations) handleRemoteRepository(repoURL string) (string, error) {
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

// GetRepositoryCommit gets the current commit hash of a repository, or returns info for local layers
func (g *GitOperations) GetRepositoryCommit(localPath string) (string, error) {
	// Check if this is a git repository
	if _, err := os.Stat(filepath.Join(localPath, ".git")); err != nil {
		if os.IsNotExist(err) {
			// Not a git repository, return directory info
			return "local-dir", nil
		}
		return "", err
	}

	// It's a git repository, get commit info
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
