package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// DiffSource specifies what to diff against.
type DiffSource struct {
	Commit string // specific commit SHA or ref
	Branch string // branch to compare
	Base   string // base branch for comparison
}

// GetDiff retrieves the git diff for the given repo directory.
// If source is empty, it auto-detects: unstaged → staged → last commit.
func GetDiff(repoDir string, source DiffSource) (string, error) {
	if source.Commit != "" {
		return runGit(repoDir, "show", source.Commit)
	}
	if source.Branch != "" {
		base := source.Base
		if base == "" {
			// Detect default branch
			defaultBranch, err := runGit(repoDir, "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
			if err != nil {
				// Fallback to "main" if detection fails
				base = "main"
			} else {
				// refs/remotes/origin/main → main
				base = strings.TrimPrefix(strings.TrimSpace(defaultBranch), "refs/remotes/origin/")
			}
		}
		return runGit(repoDir, "diff", base+"..."+source.Branch)
	}

	// Auto-detect: unstaged → staged → last commit with code changes
	diff, err := runGit(repoDir, "diff")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(diff) != "" {
		return diff, nil
	}

	diff, err = runGit(repoDir, "diff", "--cached")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(diff) != "" {
		return diff, nil
	}

	// Find most recent commit with actual diff content (skip binary-only commits)
	logOut, err := runGit(repoDir, "log", "--format=%H", "-20")
	if err != nil {
		return "", err
	}
	for _, sha := range strings.Split(strings.TrimSpace(logOut), "\n") {
		sha = strings.TrimSpace(sha)
		if sha == "" {
			continue
		}
		diff, err = runGit(repoDir, "show", sha)
		if err != nil {
			return "", err
		}
		if hasCodeChanges(diff) {
			return diff, nil
		}
	}

	return "", nil
}

// GetRemoteProject extracts the project path (e.g. "group/project") from the
// origin remote URL. Works with both HTTPS and SSH URLs.
func GetRemoteProject(repoDir string) (string, error) {
	out, err := runGit(repoDir, "remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	u := strings.TrimSpace(out)
	u = strings.TrimSuffix(u, ".git")
	parts := strings.Split(u, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("cannot parse project from remote URL: %s", u)
	}
	return strings.Join(parts[len(parts)-2:], "/"), nil
}

// hasCodeChanges checks if a diff contains actual code changes (not binary-only).
func hasCodeChanges(diff string) bool {
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "Binary files") {
			continue
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			return true
		}
	}
	return false
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), string(exitErr.Stderr))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}
