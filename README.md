# Velo Deploy

> **Bare Metal PaaS** — Deploy Node.js applications and static sites to any VPS without Docker.

Velo Deploy uses **systemd** for process management and **Caddy** for automatic HTTPS, providing a lightweight, secure alternative to container-based deployments.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](go.mod)
[![Tests](https://img.shields.io/github/actions/workflow/status/antojsh/velo-deploy/test.yml?branch=master&label=tests)](https://github.com/antojsh/velo-deploy/actions/workflows/test.yml)
[![Docs](https://img.shields.io/github/actions/workflow/status/antojsh/velo-deploy/docs.yml?branch=master&label=docs)](https://github.com/antojsh/velo-deploy/actions/workflows/docs.yml)
[![Release](https://img.shields.io/github/v/release/antojsh/velo-deploy)](https://github.com/antojsh/velo-deploy/releases/latest)
[![Documentation](https://img.shields.io/badge/docs-antojsh.github.io%2Fvelo--deploy-blue)](https://antojsh.github.io/velo-deploy/)

[📖 **Read the documentation**](https://antojsh.github.io/velo-deploy/) · [🇪🇸 **En español**](https://antojsh.github.io/velo-deploy/es/) · [🐛 **Report a bug**](https://github.com/antojsh/velo-deploy/issues/new?template=bug.yml) · [💡 **Request a feature**](https://github.com/antojsh/velo-deploy/issues/new?template=feature.yml)

---

## One-liner install

```bash
curl -sS https://get.velo-deploy.sh | bash
```

Then deploy your first app:

```bash
velo-deploy deploy https://github.com/antojsh/velo-deploy-demo
```

**That's it.** Velo clones, builds, and serves the app on a local alias. Point a real domain at it and HTTPS is automatic.

## Features

- ⚡ **One-liner install** — up and running in seconds
- 🔒 **Automatic HTTPS** — Let's Encrypt via Caddy
- 🚀 **Node.js + static sites** — APIs, frontends, and pre-rendered assets
- 🛡️ **Systemd sandboxing** — every app is its own Linux user with a hardened unit
- 🔁 **GitHub webhooks** — push to `main`, get a redeploy
- 🖥️ **TUI dashboard** — keyboard-driven, no web UI to maintain
- 📦 **Single static binary** — no runtime, no Docker, minimal footprint

## How it works

```
┌─────────────────────────────────────────────────────────────┐
│                        Velo Deploy                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   GitHub ──webhook──▶ velo-deploy-watcher (port 9999)      │
│                              │                              │
│                              ▼                              │
│                    ┌─────────────────┐                      │
│                    │  git pull       │                      │
│                    │  npm install    │                      │
│                    │  (build)        │                      │
│                    └────────┬────────┘                      │
│                             │                              │
│              ┌──────────────┴──────────────┐                │
│              ▼                              ▼              │
│   ┌─────────────────────┐      ┌─────────────────────┐      │
│   │   Node.js Apps      │      │   Static Sites      │      │
│   │   (systemd)         │      │   (Caddy)           │      │
│   │   :3000-3999        │      │   file_server       │      │
│   └─────────┬───────────┘      └──────────┬─────────┘      │
│             │                             │                │
│             └──────────┬──────────────────┘                │
│                        ▼                                   │
│               ┌─────────────────┐                          │
│               │     Caddy       │                          │
│               │   :80  :443     │                          │
│               │  HTTPS + Proxy  │                          │
│               └─────────────────┘                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

Read the [architecture overview](https://antojsh.github.io/velo-deploy/architecture/overview/) for the full picture.

## Documentation

The complete documentation lives in [`docs/`](docs/) and is deployed to **https://antojsh.github.io/velo-deploy/**.

| | English | Español |
| --- | --- | --- |
| Getting started | [Install](https://antojsh.github.io/velo-deploy/getting-started/installation/) · [Quick deploy](https://antojsh.github.io/velo-deploy/getting-started/quick-deploy/) · [First app](https://antojsh.github.io/velo-deploy/getting-started/first-app/) | [Instalación](https://antojsh.github.io/velo-deploy/es/getting-started/installation/) · [Deploy rápido](https://antojsh.github.io/velo-deploy/es/getting-started/quick-deploy/) · [Primera app](https://antojsh.github.io/velo-deploy/es/getting-started/first-app/) |
| Guides | [CLI](https://antojsh.github.io/velo-deploy/guide/cli/) · [TUI](https://antojsh.github.io/velo-deploy/guide/tui/) · [Domains](https://antojsh.github.io/velo-deploy/guide/domains/) · [Webhooks](https://antojsh.github.io/velo-deploy/guide/webhooks/) | [CLI](https://antojsh.github.io/velo-deploy/es/guide/cli/) · [TUI](https://antojsh.github.io/velo-deploy/es/guide/tui/) · [Dominios](https://antojsh.github.io/velo-deploy/es/guide/domains/) · [Webhooks](https://antojsh.github.io/velo-deploy/es/guide/webhooks/) |
| Reference | [Config schema](https://antojsh.github.io/velo-deploy/reference/config-schema/) · [CLI reference](https://antojsh.github.io/velo-deploy/reference/cli-reference/) | [Schema de config](https://antojsh.github.io/velo-deploy/es/reference/config-schema/) · [Referencia de CLI](https://antojsh.github.io/velo-deploy/es/reference/cli-reference/) |
| Operations | [Upgrade](https://antojsh.github.io/velo-deploy/operations/upgrade/) · [Uninstall](https://antojsh.github.io/velo-deploy/operations/uninstall/) · [Troubleshooting](https://antojsh.github.io/velo-deploy/operations/troubleshooting/) | [Upgrade](https://antojsh.github.io/velo-deploy/es/operations/upgrade/) · [Desinstalar](https://antojsh.github.io/velo-deploy/es/operations/uninstall/) · [Troubleshooting](https://antojsh.github.io/velo-deploy/es/operations/troubleshooting/) |

## Requirements

- **OS:** Linux (Ubuntu 20.04+, Debian 11+, or similar)
- **Architecture:** x86_64 (amd64)
- **Privileges:** `root` access for installation
- **Ports:** `80`, `443`, and optionally `9999` (webhooks)

## Build from source

Requires Go 1.22 or later.

```bash
git clone https://github.com/antojsh/velo-deploy.git
cd velo-deploy
go build -o velo-deploy ./cmd/velo-deploy
sudo install -m 0755 velo-deploy /usr/local/bin/velo-deploy
sudo ./install.sh
```

## Project structure

```
.
├── cmd/velo-deploy/   # Main entry point (CLI / TUI / daemon)
├── internal/          # Go packages
│   ├── caddy/         # Caddy vhost generation
│   ├── config/        # /etc/velo-deploy/config.json
│   ├── deploy/        # Git pull, build, systemd wiring
│   ├── hosts/         # /etc/hosts management
│   ├── node/          # nvm + Node version management
│   ├── systemd/       # Unit file generation
│   └── tui/           # Bubble Tea dashboard
├── docs/              # Astro + Starlight documentation site
├── .github/           # Issue templates, PR template, workflows
├── install.sh         # System installer
└── uninstall.sh       # System uninstaller
```

## Tech stack

| Component | Technology | Purpose |
| --- | --- | --- |
| Language | Go 1.22+ | Single static binary, fast startup, easy cross-compile |
| TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Terminal UI framework |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) | TUI styling |
| Process mgmt | systemd | App lifecycle, isolation, hardening |
| Web server | [Caddy](https://caddyserver.com/) | Reverse proxy, automatic HTTPS via ACME |
| Node.js | [nvm](https://github.com/nvm-sh/nvm) | Per-app version management |
| Docs | [Astro](https://astro.build/) + [Starlight](https://starlight.astro.build/) | Static documentation site |

## Contributing

We welcome contributions of all sizes — bug fixes, docs, feature proposals, and new generators.

- Read [CONTRIBUTING.md](CONTRIBUTING.md) for the workflow and commit conventions.
- Browse [good first issues](https://github.com/antojsh/velo-deploy/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).
- Join the conversation in [GitHub Discussions](https://github.com/antojsh/velo-deploy/discussions).
- Read our [Code of Conduct](CODE_OF_CONDUCT.md) before participating.

### Development

```bash
# Run the tests
go test ./...

# Run with coverage
go test -cover ./...

# Lint
go vet ./...
gofmt -l .

# Run the docs locally
cd docs && pnpm install && pnpm dev
```

## Releases

This project uses [release-please](https://github.com/googleapis/release-please) to automate versioning, CHANGELOG generation, and GitHub releases. Commits that follow [Conventional Commits](https://www.conventionalcommits.org/) are automatically rolled into the next release. See [CONTRIBUTING.md](CONTRIBUTING.md#commit-conventions) for details.

## Security

Found a vulnerability? Please **do not** open a public issue. Follow the disclosure process in [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE) © 2024 Velo Deploy contributors
