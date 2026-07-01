package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/openscript-ch/undercov/internal/coverage"
	"github.com/openscript-ch/undercov/internal/gitrepo"
	"github.com/openscript-ch/undercov/internal/lcov"
)

type Config struct {
	WorkDir            string
	Files              string
	Branch             string
	Push               bool
	Remote             string
	PushForceWithLease bool
	Threshold          float64
	CheckRegression    bool
}

func Run(_ context.Context, config Config) error {
	if config.WorkDir == "" {
		var err error
		config.WorkDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	if config.Files == "" {
		config.Files = "**/coverage/lcov.info"
	}
	if config.Branch == "" {
		config.Branch = "coverage"
	}
	if config.Remote == "" {
		config.Remote = "origin"
	}

	repoRoot, err := gitrepo.RepoRoot(config.WorkDir)
	if err != nil {
		return err
	}

	matchedFiles, err := discoverFiles(repoRoot, splitPatterns(config.Files))
	if err != nil {
		return err
	}
	if len(matchedFiles) == 0 {
		return fmt.Errorf("no coverage files matched %q", config.Files)
	}

	runner := gitrepo.New(repoRoot)
	coverageFiles := make(map[string][]byte, len(matchedFiles))
	var failures []string

	for _, absPath := range matchedFiles {
		relPath, err := filepath.Rel(repoRoot, absPath)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		content, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}
		storagePath := coverage.StoragePath(config.Branch, relPath)
		coverageFiles[storagePath] = content

		currentCoverage, err := lcov.Aggregate(content)
		if err != nil {
			return fmt.Errorf("parse %s: %w", relPath, err)
		}

		fmt.Printf("%s: %.2f%%\n", relPath, currentCoverage)

		if config.Threshold > 0 {
			if err := coverage.CheckThreshold(relPath, currentCoverage, config.Threshold); err != nil {
				failures = append(failures, err.Error())
			}
		}

		if config.CheckRegression {
			previousCoverage, ok, err := loadPreviousCoverage(runner, config.Branch, relPath)
			if err != nil {
				return err
			}
			if ok {
				if err := coverage.CheckRegression(relPath, previousCoverage, currentCoverage); err != nil {
					failures = append(failures, err.Error())
				}
			}
		}
	}

	if err := runner.UpdateBranch(config.Branch, coverageFiles); err != nil {
		return err
	}

	if config.Push {
		if config.PushForceWithLease {
			if err := runner.PushBranchForceWithLease(config.Remote, config.Branch); err != nil {
				return err
			}
		} else {
			if err := runner.PushBranch(config.Remote, config.Branch); err != nil {
				return err
			}
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		return errors.New(strings.Join(failures, "\n"))
	}

	return nil
}

func splitPatterns(value string) []string {
	parts := strings.Split(value, ",")
	patterns := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			patterns = append(patterns, filepath.ToSlash(part))
		}
	}
	return patterns
}

func discoverFiles(repoRoot string, patterns []string) ([]string, error) {
	matched := make(map[string]struct{})
	for _, pattern := range patterns {
		found, err := doublestar.Glob(os.DirFS(repoRoot), pattern)
		if err != nil {
			return nil, err
		}
		for _, relPath := range found {
			absPath := filepath.Join(repoRoot, filepath.FromSlash(relPath))
			if info, err := os.Stat(absPath); err == nil && !info.IsDir() {
				matched[absPath] = struct{}{}
			}
		}
	}

	files := make([]string, 0, len(matched))
	for absPath := range matched {
		files = append(files, absPath)
	}
	sort.Strings(files)
	return files, nil
}

func loadPreviousCoverage(runner *gitrepo.Runner, branch, relPath string) (float64, bool, error) {
	content, err := runner.ReadFile(branch, coverage.StoragePath(branch, relPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || gitrepo.IsNotFound(err) {
			return 0, false, nil
		}
		return 0, false, err
	}

	parsed, err := lcov.Aggregate(content)
	if err != nil {
		return 0, false, err
	}

	return parsed, true, nil
}
