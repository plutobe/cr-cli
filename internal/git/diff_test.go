package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("setup command %v failed: %v", args, err)
		}
	}
	return dir
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
}

func TestGetDiff_Unstaged(t *testing.T) {
	dir := setupTestRepo(t)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)
	runGitCmd(t, dir, "add", ".")
	runGitCmd(t, dir, "commit", "-m", "init")

	diff, err := GetDiff(dir, DiffSource{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No changes yet, should return empty
	if len(diff) != 0 {
		t.Errorf("expected empty diff, got %q", diff)
	}

	// Make unstaged change
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	diff, err = GetDiff(dir, DiffSource{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff) == 0 {
		t.Error("expected non-empty diff for unstaged changes")
	}
}

func TestGetDiff_Staged(t *testing.T) {
	dir := setupTestRepo(t)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)
	runGitCmd(t, dir, "add", ".")

	diff, err := GetDiff(dir, DiffSource{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Staged but no unstaged, should return staged diff
	if len(diff) == 0 {
		t.Error("expected non-empty diff for staged changes")
	}
}

func TestGetDiff_Commit(t *testing.T) {
	dir := setupTestRepo(t)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)
	runGitCmd(t, dir, "add", ".")
	runGitCmd(t, dir, "commit", "-m", "init")
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	runGitCmd(t, dir, "add", ".")
	runGitCmd(t, dir, "commit", "-m", "add main")

	diff, err := GetDiff(dir, DiffSource{Commit: "HEAD~1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff) == 0 {
		t.Error("expected non-empty diff for commit")
	}
}
