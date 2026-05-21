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

func TestGetDiff_Unstaged(t *testing.T) {
	dir := setupTestRepo(t)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)
	exec.Command("git", "add", ".").Run()

	diff, err := GetDiff(dir, DiffSource{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No unstaged changes yet, should return empty
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
	exec.Command("git", "add", ".").Run()

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
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "init").Run()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "add main").Run()

	diff, err := GetDiff(dir, DiffSource{Commit: "HEAD~1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff) == 0 {
		t.Error("expected non-empty diff for commit")
	}
}
