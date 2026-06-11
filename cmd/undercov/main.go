package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openscript-ch/undercov/internal/app"
)

func main() {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	config := app.Config{WorkDir: workDir}
	if err := app.Run(context.Background(), config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
