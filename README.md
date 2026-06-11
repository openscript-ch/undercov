# undercov

Don't go undercov! Track your test coverage and check for under-coverage with this Git standalone tool. 🕵️

## About

undercov is a command-line tool built with Go that helps you track your test coverage. It works standalone and stores the coverage data inside a branch in your Git repository. With undercov, you can easily check on your CI pipeline if your changes meet the required coverage thresholds, ensuring that your code is well-tested and maintainable.

- Supports monorepos with multiple coverage files.

## Usage

undercov is a single multi-platform binary that you can download from the release page.

### Github Actions

You can use undercov in your Github Actions workflow to check for under-coverage. Here's an example of how to set it up:

```yaml
name: Coverage check

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  coverage:
    steps:
      - name: Check out repository
        uses: actions/checkout@v6

      - name: Run tests and generate lcov file
        run: |
          # Run your tests and generate the lcov file

      - name: Check coverage with undercov
        uses: openscript-ch/undercov@v1
        with:
          threshold: 80
          files: '**/coverage/lcov.info'
          branch: coverage
```

### Options

- `threshold`: The minimum coverage percentage required to pass the check.
- `files`: The glob pattern to locate the coverage files (e.g., `**/coverage/lcov.info`).
- `branch`: The name of the branch where the coverage data will be stored (default: `coverage`).

## Development

1. Clone the repository
2. Download Go dependencies

```bash
go mod download
```

3. Run lint checks with the same golangci-lint version used in CI

```bash
make lint
```

4. Run tests

```bash
make test
```

If you prefer to run golangci-lint directly without `make`, use:

```bash
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run
```
