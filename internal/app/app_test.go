package app

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	author, err := gitLog(repoRoot, "coverage-data", "--format=%an|%ae", "-1")
	if err != nil {
		t.Fatalf("git show author error = %v", err)
	}
	if string(author) != "test|test@example.com\n" {
		t.Fatalf("commit author mismatch: %q", string(author))
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

func TestRunSameCoverageTwiceDoesNotFail(t *testing.T) {
	repoRoot := initGitRepo(t)
	content := []byte("SF:foo.go\nLF:2\nLH:2\nend_of_record\n")
	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), content)

	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data"}); err != nil {
		t.Fatalf("first Run() error = %v", err)
	}
	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data"}); err != nil {
		t.Fatalf("second Run() error = %v", err)
	}
}

func TestRunDoesNotPushWhenDisabled(t *testing.T) {
	repoRoot := initGitRepo(t)
	remoteRoot := initBareRemote(t)
	mustRun(t, repoRoot, "git", "remote", "add", "origin", remoteRoot)

	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), []byte("SF:foo.go\nLF:2\nLH:1\nend_of_record\n"))
	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if remoteBranchExists(t, remoteRoot, "coverage-data") {
		t.Fatal("expected remote coverage-data branch to remain absent when push is disabled")
	}
}

func TestRunPushesCoverageToRemote(t *testing.T) {
	repoRoot := initGitRepo(t)
	remoteRoot := initBareRemote(t)
	mustRun(t, repoRoot, "git", "remote", "add", "origin", remoteRoot)

	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), []byte("SF:foo.go\nLF:2\nLH:1\nend_of_record\n"))
	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data", Push: true, Remote: "origin"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !remoteBranchExists(t, remoteRoot, "coverage-data") {
		t.Fatal("expected remote coverage-data branch to be created")
	}

	encoded := base64.RawURLEncoding.EncodeToString([]byte("coverage/lcov.info"))
	output, err := gitShowBare(remoteRoot, "coverage-data:.undercov/coverage-data/"+encoded+".lcov")
	if err != nil {
		t.Fatalf("git show in bare remote error = %v", err)
	}
	if string(output) != "SF:foo.go\nLF:2\nLH:1\nend_of_record\n" {
		t.Fatalf("remote stored content mismatch: %q", string(output))
	}
}

func TestRunPushFailsOnDivergedRemoteWithoutForce(t *testing.T) {
	repoRoot := initGitRepo(t)
	remoteRoot := initBareRemote(t)
	mustRun(t, repoRoot, "git", "remote", "add", "origin", remoteRoot)

	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), []byte("SF:foo.go\nLF:2\nLH:2\nend_of_record\n"))
	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data", Push: true, Remote: "origin"}); err != nil {
		t.Fatalf("initial Run() error = %v", err)
	}

	advanceRemoteBranch(t, remoteRoot, "coverage-data")

	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), []byte("SF:foo.go\nLF:2\nLH:1\nend_of_record\n"))
	err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data", Push: true, Remote: "origin"})
	if err == nil {
		t.Fatal("Run() expected push error on diverged remote, got nil")
	}
	if !strings.Contains(err.Error(), "push coverage branch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPushForceWithLeaseOnDivergedRemote(t *testing.T) {
	repoRoot := initGitRepo(t)
	remoteRoot := initBareRemote(t)
	mustRun(t, repoRoot, "git", "remote", "add", "origin", remoteRoot)

	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), []byte("SF:foo.go\nLF:2\nLH:2\nend_of_record\n"))
	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data", Push: true, Remote: "origin"}); err != nil {
		t.Fatalf("initial Run() error = %v", err)
	}

	advanceRemoteBranch(t, remoteRoot, "coverage-data")
	mustRun(t, repoRoot, "git", "fetch", "origin", "coverage-data:refs/remotes/origin/coverage-data")

	writeFile(t, filepath.Join(repoRoot, "coverage", "lcov.info"), []byte("SF:foo.go\nLF:2\nLH:1\nend_of_record\n"))
	if err := Run(context.Background(), Config{WorkDir: repoRoot, Files: "coverage/lcov.info", Branch: "coverage-data", Push: true, Remote: "origin", PushForceWithLease: true}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString([]byte("coverage/lcov.info"))
	output, err := gitShowBare(remoteRoot, "coverage-data:.undercov/coverage-data/"+encoded+".lcov")
	if err != nil {
		t.Fatalf("git show in bare remote error = %v", err)
	}
	if string(output) != "SF:foo.go\nLF:2\nLH:1\nend_of_record\n" {
		t.Fatalf("remote stored content mismatch after force-with-lease push: %q", string(output))
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

func gitLog(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"log"}, args...)...)
	cmd.Dir = dir
	return cmd.Output()
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

func gitShowBare(bareRepo string, ref string) ([]byte, error) {
	cmd := exec.Command("git", "--git-dir", bareRepo, "show", ref)
	return cmd.Output()
}

func advanceRemoteBranch(t *testing.T, bareRepo string, branch string) {
	t.Helper()
	parent := t.TempDir()
	clonePath := filepath.Join(parent, "clone")
	mustRun(t, parent, "git", "clone", bareRepo, clonePath)
	mustRun(t, clonePath, "git", "config", "user.name", "remote")
	mustRun(t, clonePath, "git", "config", "user.email", "remote@example.com")
	mustRun(t, clonePath, "git", "checkout", "-B", branch, "origin/"+branch)
	writeFile(t, filepath.Join(clonePath, "remote.txt"), []byte("remote"))
	mustRun(t, clonePath, "git", "add", "remote.txt")
	mustRun(t, clonePath, "git", "commit", "-m", "advance remote", "--no-gpg-sign")
	mustRun(t, clonePath, "git", "push", "origin", branch)
}
