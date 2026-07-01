# Releasing

This project uses [release-please](https://github.com/googleapis/release-please) to automate versioning and GitHub releases. The pipeline ships two independent packages from this single repository:

| Package | Path | Release type | What ships |
| --- | --- | --- | --- |
| `velo-deploy` | `.` (Go module) | `go` | Cross-platform binaries + checksums |
| `velo-deploy-docs` | `docs/` (Astro) | `node` | Static site as a downloadable zip |

The CLI and the docs are **versioned independently** but share a single release artifact per push. If a push touches only `docs/`, only the docs version moves; if it touches only Go code, only the CLI version moves; if it touches both, both move.

## Conventional Commits

release-please parses commit messages using the [Conventional Commits](https://www.conventionalcommits.org/) spec. PR titles are also validated by `commitlint` (see `.commitlintrc.json`).

| Prefix | Effect |
| --- | --- |
| `feat:` | Minor bump |
| `fix:` | Patch bump |
| `perf:` | Patch bump |
| `feat!:` / `fix!:` (or `BREAKING CHANGE:` footer) | Major bump |
| `chore:`, `docs:`, `refactor:`, `test:`, `ci:`, `build:` | No release on their own; ignored by release-please |

For the docs package, commits under `docs/**` count toward `velo-deploy-docs` regardless of prefix.

## Release flow

1. **You push a commit (or merge a PR) to `master`.**
2. **`Release Please` workflow runs.**
   - `release-please` job: scans commits since the last release, opens a release PR that bumps versions, updates `CHANGELOG.md` / `docs/CHANGELOG.md`, and updates `.release-please-manifest.json` + `docs/package.json` + `go.mod`.
3. **You merge the release PR** (e.g., `chore: release 0.2.0`).
4. **The push to master triggers the workflow again.**
   - `release-please` action detects a release was just merged, creates a GitHub release with the changelog as the body, and outputs the new tag.
   - `check` job: reads `.release-please-manifest.json` to figure out which packages were actually bumped in this release (CLI, docs, or both).
   - `binaries` job: only runs if the CLI was bumped. Builds `velo-deploy` for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`. Packages each as `.tar.gz` (nix) or `.zip` (windows) containing the binary, `README.md`, and `LICENSE`.
   - `docs` job: only runs if the docs were bumped. Runs `pnpm build` and zips the resulting `docs/dist/` as `velo-deploy-docs-site_<version>.zip`.
   - `attach` job: generates `SHA256SUMS` for the binaries, then uploads all artifacts to the GitHub release via `softprops/action-gh-release`.
5. **The release is live on GitHub** with the binaries, checksums, and docs zip attached.

## Release assets

A release with a CLI bump attaches:

```
velo-deploy_0.2.0_linux_amd64.tar.gz
velo-deploy_0.2.0_linux_arm64.tar.gz
velo-deploy_0.2.0_darwin_amd64.tar.gz
velo-deploy_0.2.0_darwin_arm64.tar.gz
velo-deploy_0.2.0_windows_amd64.zip
velo-deploy_0.2.0_windows_arm64.zip
SHA256SUMS
```

If the docs were also bumped:

```
velo-deploy-docs-site_0.2.0.zip
```

A docs-only release attaches just the docs zip (no binaries, no SHA256SUMS).

## How the oneliner installer finds the binary

`install.sh` is the user-facing installer (`curl -sS https://get.velo-deploy.sh | bash`).

When invoked without a specific version:

1. It calls `https://api.github.com/repos/antojsh/velo-deploy/releases?per_page=20` and walks the list.
2. It picks the **most recent release that has a `velo-deploy_<os>_<arch>.tar.gz` asset**. This automatically skips docs-only releases.
3. It downloads the matching asset + `SHA256SUMS`, verifies the checksum, and extracts the binary to `/usr/local/bin/velo-deploy`.

Users can pin a specific version:

```bash
VELO_DEPLOY_VERSION=v0.2.0 bash install.sh
```

Or install from a fork:

```bash
VELO_DEPLOY_REPO=myuser/myfork bash install.sh
```

## One-time bootstrap (first release)

If this is the first time the pipeline runs against an empty release history, complete these steps manually before pushing:

1. **Fix `go.sum`.** Some upstream Go modules (notably `testify` v1.9.0) have had their `go.mod` re-hashed, which makes `go mod download` fail with `checksum mismatch`. Run locally and commit:

   ```bash
   go mod tidy
   git add go.sum go.mod
   git commit -m "chore: refresh go.sum"
   ```

   The workflows also run `go mod tidy` defensively, but committing the fix avoids a broken CI on the first push.

2. **Anchor release-please with the initial tag.** release-please needs a real `v0.1.0` git tag to compute the next version. The manifest pins the version but release-please also calls the GitHub API to generate release notes, which requires the tag to exist:

   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

   This is a one-time operation. Subsequent releases will tag themselves via the release-please workflow.

3. **Enable GitHub Pages.** The docs workflow deploys to Pages on every push. If Pages is not enabled, the deploy step fails with `404 Not Found`. Enable it at **Settings → Pages → Build and deployment → GitHub Actions** for `https://github.com/<owner>/velo-deploy/settings/pages`.

After these three steps, the pipeline runs end-to-end without manual intervention.

## Pre-1.0 versioning

This project is pre-1.0 (`bump-minor-pre-major: false`). Conventional commits behave as follows until 1.0.0:

- `feat:` → minor bump (e.g., 0.1.0 → 0.2.0)
- `fix:` → patch bump (e.g., 0.1.0 → 0.1.1)
- `feat!:` / `BREAKING CHANGE:` → major bump (e.g., 0.1.0 → 1.0.0)

This means a `feat:` is still a meaningful release even before 1.0.

## Validating the pipeline locally

`build-binaries.yml` is a separate workflow that runs on every push and PR to `master`. It builds the same matrix as the release pipeline (minus packaging checksums and attaching to a release) to catch cross-platform compile errors before merge.

You can also trigger it manually via the GitHub Actions UI (`workflow_dispatch`).

## Troubleshooting

### "Could not find a release with a velo-deploy binary for linux/amd64."

The 20 most recent releases are docs-only or were created before the cross-platform build pipeline was wired up. Pin a specific CLI version:

```bash
VELO_DEPLOY_VERSION=v0.2.0 bash install.sh
```

### `release-please` opens a PR but `binaries`/`docs` jobs are skipped.

Check the `check` job's output. It reads `.release-please-manifest.json` and compares each package's version to the new tag. If the CLI version doesn't match the tag, the `binaries` job is skipped (and vice versa for docs). This is the expected behavior for single-package releases.

### The release is created but no binaries are attached.

`release_created` was `true` but the `check` job determined neither package was bumped. This usually means the release-please PR was merged without any conventional commits since the previous release. Revert the merge and verify the commit history since the last tag has `feat:` or `fix:` entries.

### `Error: release-please failed: Invalid previous_tag parameter` on the very first run.

This means there is no `v0.1.0` git tag in the repo yet. release-please's release-notes API call needs a real previous tag. Create it manually (see the "One-time bootstrap" section above) and re-run.

### `verifying ... checksum mismatch` from `go mod download`.

An upstream Go module's `go.mod` was re-hashed. The workflows already run `go mod tidy` to self-heal, but if you need a clean local run:

```bash
go mod tidy
```

Commit the updated `go.sum` (and `go.mod` if it changed). The CI workflows will tolerate the mismatch going forward.

### `Creating Pages deployment failed (status: 404)` from `actions/deploy-pages`.

GitHub Pages is not enabled for this repository. Enable it at **Settings → Pages → GitHub Actions**. Once enabled, the docs workflow's deploy job will succeed.
