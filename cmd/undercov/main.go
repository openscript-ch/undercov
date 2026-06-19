package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/openscript-ch/undercov/internal/app"
)

var version = "dev"

func main() {
	files := flag.String("files", "**/coverage/lcov.info", "Glob pattern for LCOV files")
	branch := flag.String("branch", "coverage", "Branch used to store coverage data")
	push := flag.Bool("push", false, "Push the coverage branch to the configured remote")
	remote := flag.String("remote", "origin", "Remote used when pushing coverage data")
	pushForceWithLease := flag.Bool("push-force-with-lease", false, "Use --force-with-lease when pushing coverage branch")
	threshold := flag.Float64("threshold", 0, "Minimum coverage percentage")
	checkRegression := flag.Bool("check-regression", false, "Fail if coverage regresses compared to previously stored data")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("undercov version %s\n", version)
		os.Exit(0)
	}

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	config := app.Config{
		WorkDir:            workDir,
		Files:              *files,
		Branch:             *branch,
		Push:               *push,
		Remote:             *remote,
		PushForceWithLease: *pushForceWithLease,
		Threshold:          *threshold,
		CheckRegression:    *checkRegression,
	}
	if err := app.Run(context.Background(), config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
