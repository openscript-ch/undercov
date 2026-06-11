package app

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunStoresCoverage(t *testing.T) {
	repoRoot := initGitRepo(t)
	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), []byte("SF:foo.go\nLF:2\nLH:1\nend_of_record\n"))

	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString([]byte("coverage/lcov.info"))
	output, err := gitShow(repoRoot, "coverage-data:.undercov/coverage-data/"+encoded+".lcov")
	if err != nil {
		t.Fatalf("git show error = %v", err)
	}
	if string(output) != "SF:foo.go\nLF:2\nLH:1\nend_of_record\n" {
		t.Fatalf("stored content mismatch: %q", string(output))
	}
}

func TestRunRegressionCheck(t *testing.T) {
	repoRoot := initGitRepo(t)
	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), []byte("SF:foo.go\nLF:2\nLH:2\nend_of_record\n"))
	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data"}); err != nil {
		t.Fatalf("initial Run() error = %v", err)
	}

	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), []byte("SF:foo.go\nLF:2\nLH:1\nend_of_record\n"))
	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data", CheckRegression: true}); err == nil {
		t.Fatal("Run() expected regression error, got nil")
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	mustRun(t, repoRoot, "git", "init")
	mustRun(t, repoRoot, "git", "config", "user.name", "test")
	mustRun(t, repoRoot, "git", "config", "user.email", "test@example.com")
	mustRun(t, repoRoot, "git", "config", "commit.gpgsign", "false")
	writeFile(t, filepath.Join(repoRoot, "README.md"), []byte("repo"))
	mustRun(t, repoRoot, "git", "add", ".")
	mustRun(t, repoRoot, "git", "commit", "-m", "init", "--no-gpg-sign")
	return repoRoot
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, string(out))
	}
}

func gitShow(dir, ref string) ([]byte, error) {
	cmd := exec.Command("git", "show", ref)
	cmd.Dir = dir
	return cmd.Output()
}
