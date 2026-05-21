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
// If source is empty, it auto-detects: unstaged -> staged.
func GetDiff(repoDir string, source DiffSource) (string, error) {
	if source.Commit != "" {
		return runGit(repoDir, "show", source.Commit)
	}
	if source.Branch != "" {
		base := source.Base
		if base == "" {
			base = "main"
		}
		return runGit(repoDir, "diff", base+"..."+source.Branch)
	}

	// Auto-detect: unstaged -> staged
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
	return diff, nil
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
