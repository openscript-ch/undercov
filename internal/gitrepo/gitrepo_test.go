package gitrepo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPushBranchReturnsHelpfulError(t *testing.T) {
	repoRoot := initGitRepo(t)
	runner := New(repoRoot)

	err := runner.PushBranch("missing", "coverage-data")
	if err == nil {
		t.Fatal("PushBranch() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "push coverage branch \"coverage-data\" to remote \"missing\"") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestPushBranchForceWithLeasePublishesToRemote(t *testing.T) {
	repoRoot := initGitRepo(t)
	remoteRoot := initBareRemote(t)
	mustRun(t, repoRoot, "git", "remote", "add", "origin", remoteRoot)

	files := map[string][]byte{
		".undercov/coverage-data/Zm9vLm1k.lcov": []byte("SF:foo.go\nLF:1\nLH:1\nend_of_record\n"),
	}
	runner := New(repoRoot)
	if err := runner.UpdateBranch("coverage-data", files); err != nil {
		t.Fatalf("UpdateBranch() error = %v", err)
	}
	if err := runner.PushBranchForceWithLease("origin", "coverage-data"); err != nil {
		t.Fatalf("PushBranchForceWithLease() error = %v", err)
	}

	if !remoteBranchExists(t, remoteRoot, "coverage-data") {
		t.Fatal("expected coverage-data branch in remote")
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

func initBareRemote(t *testing.T) string {
	t.Helper()
	parent := t.TempDir()
	remoteRoot := filepath.Join(parent, "remote.git")
	mustRun(t, parent, "git", "init", "--bare", remoteRoot)
	return remoteRoot
}

func remoteBranchExists(t *testing.T, bareRepo string, branch string) bool {
	t.Helper()
	cmd := exec.Command("git", "--git-dir", bareRepo, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
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
