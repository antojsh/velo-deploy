# Contributing to Velo Deploy

Thank you for your interest in contributing to Velo Deploy. This document covers everything you need to send a pull request that lands quickly.

## Code of conduct

By participating, you agree to uphold our [Code of Conduct](CODE_OF_CONDUCT.md). Please read it before contributing.

## Quick links

- [Good first issues](https://github.com/antojsh/velo-deploy/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)
- [Help wanted](https://github.com/antojsh/velo-deploy/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
- [Discussions](https://github.com/antojsh/velo-deploy/discussions)
- [Security policy](SECURITY.md)

## Ways to contribute

- **Bug reports** — open an issue with the [bug template](.github/ISSUE_TEMPLATE/bug.yml).
- **Feature requests** — open an issue with the [feature template](.github/ISSUE_TEMPLATE/feature.yml) or start a discussion first.
- **Documentation** — typo fixes, missing examples, and new guides are always welcome. The docs live in [`docs/`](docs/) and use Astro + Starlight.
- **Code** — bug fixes, new commands, refactors. Always start with an issue or discussion.
- **Tests** — coverage gaps, integration tests, fuzzing.
- **Triage** — answering questions, reproducing bugs, reviewing PRs.

## Development setup

### Prerequisites

- Go 1.22 or later
- Linux x86_64 for full integration testing (macOS and ARM work for unit tests)
- A clean VPS or VM where you can run the installer end-to-end
- Node 20+ and pnpm 9+ for documentation changes

### Clone and build

```bash
git clone https://github.com/antojsh/velo-deploy.git
cd velo-deploy
go mod download
go build -o velo-deploy ./cmd/velo-deploy
```

### Run the tests

```bash
# All tests
go test ./...

# With coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Per-package
go test -v ./internal/config/...
go test -v ./internal/systemd/...
```

The CI workflow enforces minimum coverage thresholds per package — see [`.github/workflows/test.yml`](.github/workflows/test.yml).

### Run the docs locally

```bash
cd docs
pnpm install
pnpm dev          # http://localhost:4321
pnpm build        # production build
pnpm check        # type-check and validate links
```

The docs use Astro 5 + Starlight with bilingual content (`docs/src/content/docs/en/` and `docs/src/content/docs/es/`). When you change a page, update both locales.

### Linting

```bash
# Go
gofmt -l .
go vet ./...

# Markdown / docs
cd docs && pnpm exec astro check
```

## Workflow

1. **Open an issue** for any non-trivial change. Discuss the approach before writing code.
2. **Fork** the repository.
3. **Create a feature branch** off `master`:
   ```bash
   git checkout -b feat/short-description
   git checkout -b fix/short-description
   git checkout -b docs/short-description
   ```
4. **Make your changes** in small, focused commits.
5. **Add or update tests.** PRs without tests are unlikely to be merged.
6. **Update documentation** if you change user-facing behavior.
7. **Run the linter and the test suite locally** before pushing.
8. **Open a pull request** using the [PR template](.github/PULL_REQUEST_TEMPLATE.md).
9. **Address review feedback** in additional commits — do not force-push unless asked.

## Commit conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/). Commit messages are parsed by [release-please](https://github.com/googleapis/release-please) to generate the [CHANGELOG](CHANGELOG.md) and bump the version, so getting the format right matters.

### Format

```
<type>(<scope>)<!>: <short description>

<longer description>

<footer>
```

- **type** — one of the values below.
- **scope** (optional) — a short identifier of the affected area (`cli`, `tui`, `daemon`, `caddy`, `config`, `systemd`, `node`, `hosts`, `deploy`, `docs`, `ci`).
- **`!`** — breaking change marker. Use it together with a `BREAKING CHANGE:` footer.
- **short description** — imperative, lowercase, no period, max 72 chars.
- **footer** — one or more of:
  - `Refs: #123` — references a related issue or PR.
  - `Closes: #123` — closes an issue when the commit lands.
  - `BREAKING CHANGE: <description>` — required for breaking changes.

### Types

| Type | Description | Version bump |
| --- | --- | --- |
| `feat` | A new user-facing feature | minor |
| `fix` | A bug fix | patch |
| `docs` | Documentation only | none |
| `refactor` | Code change that neither fixes a bug nor adds a feature | none |
| `perf` | Performance improvement | patch |
| `test` | Add or correct tests | none |
| `build` | Build system or external dependencies (Go modules, scripts) | none |
| `ci` | CI configuration files and scripts | none |
| `chore` | Other changes that don't modify `src` or test files | none |
| `revert` | Reverts a previous commit | patch |

### Examples

```text
feat(cli): add --json flag to velo-deploy list

Adds machine-readable output for scripts and dashboards. The output is
stable and versioned.

Closes: #142
```

```text
fix(systemd): correct ReadWritePaths for static sites

Static apps were getting a ReadWritePaths entry pointing at the build
output dir even though the app process never writes there. Drop the
entry to tighten the sandbox.

Fixes: #189
```

```text
feat(api)!: rename velo-deploy config --port to --http-port

The --port flag was ambiguous for static sites. Rename to --http-port
and add an alias for backwards compatibility.

BREAKING CHANGE: --port is no longer accepted. Use --http-port instead.
The --http-port alias preserves the old behaviour for one major release.
```

```text
docs(es): translate the configuration guide
```

### Validating your commit messages

```bash
# Install commitlint (one-off)
npm install -g @commitlint/cli @commitlint/config-conventional

# Check the last commit
commitlint --from=HEAD~1 --to=HEAD --config=.commitlintrc.json
```

## Pull request guidelines

- **One change per PR.** Separate refactors from feature changes.
- **Keep the diff small.** A PR that touches more than ~400 lines should be split — see the [chained PRs](.github/PULL_REQUEST_TEMPLATE.md#checklist) section of the template.
- **Include tests** for any new behavior or bug fix.
- **Update the docs** in both `en/` and `es/` if you change user-facing behavior.
- **Run the linter and the test suite locally** before pushing.
- **Fill out the PR template** completely. The CI checks will fail if the description is empty.
- **Link the issue** with `Closes: #N` in a commit footer or in the PR description.
- **Be patient.** Reviews usually take 1-3 days.

### Review process

1. A maintainer will review within 1-3 days. If you don't hear back, ping in a comment.
2. CI must pass — tests, lint, and coverage thresholds.
3. At least one maintainer approval is required.
4. The PR is squash-merged into `master`.
5. release-please opens a release PR; merging it publishes the release.

## Documentation

The documentation site lives in [`docs/`](docs/) and is built with [Astro](https://astro.build/) + [Starlight](https://starlight.astro.build/). When you change user-facing behavior:

1. Update the English page under `docs/src/content/docs/en/`.
2. Update the matching Spanish page under `docs/src/content/docs/es/`.
3. Add cross-references in both languages.
4. Run `pnpm check` in `docs/` to validate links and frontmatter.

If you only speak one language, mark the PR with the `needs-translation` label and a maintainer will handle the other side.

## Release process

Releases are fully automated via [release-please](https://github.com/googleapis/release-please). The flow is:

1. Conventional commits land on `master`.
2. release-please opens a release PR with version bump, `CHANGELOG.md` update, and GitHub release notes.
3. A maintainer reviews and merges the release PR.
4. release-please creates a Git tag, builds the binaries, signs them with cosign, and publishes a GitHub release.
5. `get.velo-deploy.sh` picks up the new tag automatically.

You do **not** need to bump the version or edit the CHANGELOG manually.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
