package gitops

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// GitOpsSync handles synchronization of resources from Git repositories
type GitOpsSync struct {
	workDir string
}

// NewGitOpsSync creates a new GitOps synchronizer
func NewGitOpsSync(workDir string) *GitOpsSync {
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "colonies-gitops")
	}
	return &GitOpsSync{
		workDir: workDir,
	}
}

// SyncResources syncs resources from a Git repository based on ResourceDefinition GitOps spec
func (g *GitOpsSync) SyncResources(rd *core.ResourceDefinition) ([]*core.Resource, error) {
	if rd.Spec.GitOps == nil {
		return nil, fmt.Errorf("ResourceDefinition does not have GitOps configuration")
	}

	gitOps := rd.Spec.GitOps

	// Set defaults
	branch := gitOps.Branch
	if branch == "" {
		branch = "main"
	}

	repoPath := gitOps.Path
	if repoPath == "" {
		repoPath = "/"
	}

	// Create a unique directory for this repository
	repoHash := hashString(gitOps.RepoURL)
	repoDir := filepath.Join(g.workDir, repoHash)

	// Clone or update the repository
	commitSHA, err := g.cloneOrPull(gitOps.RepoURL, branch, repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to sync git repository: %w", err)
	}

	// Find and parse resource files
	resourcesPath := filepath.Join(repoDir, strings.TrimPrefix(repoPath, "/"))
	resources, err := g.loadResourcesFromPath(resourcesPath, rd.Spec.Names.Kind)
	if err != nil {
		return nil, fmt.Errorf("failed to load resources from path: %w", err)
	}

	// Update GitSync status for each resource
	for _, res := range resources {
		res.GitSync = &core.GitSyncStatus{
			LastSyncTime:  time.Now(),
			LastCommitSHA: commitSHA,
		}
	}

	return resources, nil
}

// cloneOrPull clones a repository if it doesn't exist, or pulls latest changes
func (g *GitOpsSync) cloneOrPull(repoURL, branch, repoDir string) (string, error) {
	// Ensure work directory exists
	if err := os.MkdirAll(g.workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create work directory: %w", err)
	}

	// Check if repository already exists
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		// Clone the repository
		cmd := exec.Command("git", "clone", "--branch", branch, "--depth", "1", repoURL, repoDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
		}
	} else {
		// Repository exists, pull latest changes
		cmd := exec.Command("git", "-C", repoDir, "fetch", "origin", branch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git fetch failed: %w, output: %s", err, string(output))
		}

		cmd = exec.Command("git", "-C", repoDir, "reset", "--hard", fmt.Sprintf("origin/%s", branch))
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git reset failed: %w, output: %s", err, string(output))
		}
	}

	// Get the current commit SHA
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %w", err)
	}

	commitSHA := strings.TrimSpace(string(output))
	return commitSHA, nil
}

// loadResourcesFromPath loads all resource files from a directory
func (g *GitOpsSync) loadResourcesFromPath(path, kind string) ([]*core.Resource, error) {
	resources := []*core.Resource{}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return resources, nil // Return empty list if path doesn't exist
	}

	// Walk through the directory
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-JSON/YAML files
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(filePath)
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return nil
		}

		// Read and parse the file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Try to parse as Resource
		var resource core.Resource
		if err := json.Unmarshal(data, &resource); err != nil {
			// Skip files that aren't valid resources
			return nil
		}

		// Filter by kind if specified
		if kind != "" && resource.Kind != kind {
			return nil
		}

		// Validate the resource
		if err := resource.Validate(); err != nil {
			return fmt.Errorf("invalid resource in file %s: %w", filePath, err)
		}

		resources = append(resources, &resource)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resources, nil
}

// hashString creates a simple hash from a string for directory naming
func hashString(s string) string {
	// Simple hash using the first 16 chars of the repo URL
	// In production, use a proper hash function
	sanitized := strings.ReplaceAll(s, "://", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}
	return sanitized
}

// CleanupWorkDir removes the work directory
func (g *GitOpsSync) CleanupWorkDir() error {
	return os.RemoveAll(g.workDir)
}
