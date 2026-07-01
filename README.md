# undercov

Don't go undercov! Track your test coverage and check for under-coverage with this Git standalone tool. 🕵️

## About

undercov is a command-line tool built with Go that helps you track your test coverage. It works standalone and stores the coverage data inside a branch in your Git repository. With undercov, you can easily check on your CI pipeline if your changes meet the required coverage thresholds, ensuring that your code is well-tested and maintainable.

- Supports monorepos with multiple coverage files.
- Checks for coverage regressions in pull requests.

## Usage

undercov is a single multi-platform binary that you can download from the release page.

### Options

- `threshold`: The minimum coverage percentage required to pass the check.
- `files`: The glob pattern to locate the coverage files (e.g., `**/coverage/lcov.info`).
- `branch`: The name of the branch where the coverage data will be stored (default: `coverage`).
- `push`: Push the updated coverage branch to a remote (default: `false`).
- `remote`: The remote used when `push` is enabled (default: `origin`).
- `push-force-with-lease`: Push with `--force-with-lease` when `push` is enabled (default: `false`).

By default, undercov stores coverage snapshots in the local coverage branch only. Enable `push` in CI to publish updates to the remote branch. If your remote branch can diverge (for example with parallel jobs), `push-force-with-lease` allows replacing the remote tip while still protecting against unexpected concurrent updates.

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

## Releases

Releases are managed with git-cliff and Forgejo workflows.

1. Conventional commits are pushed to `main`.
2. The `version.yml` workflow computes the next semantic version, updates `CHANGELOG.md` with git-cliff, and creates or updates a release PR.
3. When the release PR is merged, `tag.yml` creates and pushes the corresponding `v*` tag.
4. The tag triggers `release.yml`, which builds binaries for Linux (`x86_64`, `arm64`, `armv7`) and Windows (`x86_64`, `arm64`) and uploads artifacts plus SHA256 checksums.

To guarantee that a tag pushed by automation can trigger the release workflow in your Forgejo setup, configure the `RELEASE_BOT_TOKEN` repository secret and use a token with permission to push branches and tags.
