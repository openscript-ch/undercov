package gitrepo

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Runner struct {
	repoRoot string
}

func New(repoRoot string) *Runner {
	return &Runner{repoRoot: repoRoot}
}

func RepoRoot(workDir string) (string, error) {
	cmd := exec.Command("git", "-C", workDir, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve git repository root: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *Runner) BranchExists(branch string) (bool, error) {
	cmd := exec.Command("git", "-C", r.repoRoot, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *Runner) ReadFile(branch, repoPath string) ([]byte, error) {
	cmd := exec.Command("git", "-C", r.repoRoot, "show", branch+":"+repoPath)
	output, err := cmd.Output()
	if err != nil {
		if IsNotFound(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	return output, nil
}

func (r *Runner) UpdateBranch(branch string, files map[string][]byte) error {
	if len(files) == 0 {
		return errors.New("no coverage files to store")
	}

	tmpDir, err := os.MkdirTemp("", "undercov-worktree-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	branchExists, err := r.BranchExists(branch)
	if err != nil {
		return err
	}

	if branchExists {
		if err := runGit(r.repoRoot, "worktree", "add", "--force", tmpDir, branch); err != nil {
			return err
		}
	} else {
		if err := runGit(r.repoRoot, "worktree", "add", "--detach", tmpDir, "HEAD"); err != nil {
			return err
		}
		if err := runGit(tmpDir, "checkout", "--orphan", branch); err != nil {
			return err
		}
		if err := runGit(tmpDir, "rm", "-rf", "--ignore-unmatch", "."); err != nil {
			return err
		}
		if err := runGit(tmpDir, "clean", "-fdx"); err != nil {
			return err
		}
	}

	for repoPath, content := range files {
		fullPath := filepath.Join(tmpDir, filepath.FromSlash(repoPath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, content, 0o644); err != nil {
			return err
		}
	}

	if err := runGit(tmpDir, "add", ".undercov"); err != nil {
		return err
	}

	staged, err := runGitOutput(tmpDir, "status", "--porcelain", ".undercov")
	if err != nil {
		return err
	}
	if strings.TrimSpace(staged) != "" {
		name, email := resolveCommitIdentity(tmpDir)
		if err := runGit(tmpDir, "-c", "user.name="+name, "-c", "user.email="+email, "-c", "commit.gpgsign=false", "commit", "-m", "chore: update coverage", "--no-gpg-sign"); err != nil {
			return err
		}
	}

	if err := runGit(r.repoRoot, "worktree", "remove", "--force", tmpDir); err != nil {
		return err
	}

	return nil
}

func (r *Runner) PushBranch(remote, branch string) error {
	if err := runGit(r.repoRoot, "push", remote, branch+":"+branch); err != nil {
		return fmt.Errorf("push coverage branch %q to remote %q: %w", branch, remote, err)
	}

	return nil
}

func (r *Runner) PushBranchForceWithLease(remote, branch string) error {
	if err := runGit(r.repoRoot, "push", "--force-with-lease", remote, branch+":"+branch); err != nil {
		return fmt.Errorf("push coverage branch %q to remote %q with force-with-lease: %w", branch, remote, err)
	}

	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
		}
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return nil
}

func runGitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return strings.TrimSpace(string(out)), nil
}

func resolveCommitIdentity(dir string) (string, string) {
	name := firstNonEmpty(
		strings.TrimSpace(os.Getenv("GIT_AUTHOR_NAME")),
		strings.TrimSpace(os.Getenv("GIT_COMMITTER_NAME")),
	)
	email := firstNonEmpty(
		strings.TrimSpace(os.Getenv("GIT_AUTHOR_EMAIL")),
		strings.TrimSpace(os.Getenv("GIT_COMMITTER_EMAIL")),
	)

	if name == "" {
		if configuredName, err := runGitOutput(dir, "config", "--get", "user.name"); err == nil {
			name = configuredName
		}
	}
	if email == "" {
		if configuredEmail, err := runGitOutput(dir, "config", "--get", "user.email"); err == nil {
			email = configuredEmail
		}
	}

	if name == "" {
		name = "undercov"
	}
	if email == "" {
		email = "undercov@local"
	}

	return name, email
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func IsNotFound(err error) bool {
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		return false
	}
	text := string(exitError.Stderr)
	return strings.Contains(text, "unknown revision") || strings.Contains(text, "pathspec") || strings.Contains(text, "exists on disk, but not in")
}
