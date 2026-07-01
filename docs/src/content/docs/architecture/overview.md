---
title: Architecture overview
description: How systemd, Caddy, and the Go binary fit together.
---

Velo Deploy is intentionally small. Three moving parts, one config file, no databases.

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

## Components

### `velo-deploy` (Go binary)

A single static binary that ships the CLI, the TUI, and the embedded install/uninstall scripts. It does **not** run as a long-lived daemon — every command is short-lived and idempotent.

The CLI talks to systemd through `systemctl` and to Caddy by writing files into `/etc/caddy/conf.d/`.

### `velo-deploy-watcher` (Go binary, long-lived)

A second binary, also written in Go, that listens on port `9999` for GitHub webhook deliveries. It is the only long-running process Velo Deploy adds to the system. It is registered as the `velo-deploy-watcher.service` systemd unit.

### systemd

The process manager. Each deployed app gets a dedicated unit file at `/etc/systemd/system/velo-<app>.service`. Units are hardened with `ProtectSystem`, `ProtectHome`, `PrivateTmp`, and `NoNewPrivileges` — see [Security](/velo-deploy/architecture/security/).

### Caddy

The web server and TLS terminator. Velo writes a per-app vhost to `/etc/caddy/conf.d/<app>.conf` and triggers `caddy reload` after every change. Caddy handles ACME (Let's Encrypt) automatically.

### nvm

Node.js version manager installed at `/opt/nvm`. Velo installs the version requested by `engines.node` and pins the resulting `node` binary in the app's config.

## Data flow

1. **Deploy** — the CLI clones the repo, installs dependencies, runs the build, writes a systemd unit, writes a Caddy vhost, and registers the app in `/etc/velo-deploy/config.json`.
2. **Runtime** — Caddy terminates TLS, reverse-proxies Node.js requests, or serves static files directly.
3. **Webhook** — `velo-deploy-watcher` receives a push event, identifies the matching app, and re-runs the deploy steps for that app.
4. **Tear-down** — `velo-deploy remove` deletes the systemd unit, the Caddy vhost, the config entry, and (optionally) the source directory.

## Why no Docker?

Containers are great for portability, but most Node.js and static workloads do not need that portability. Skipping the container layer means:

- Lower memory overhead (no `containerd` or `runc`).
- Faster cold starts (no image pull).
- Simpler debugging (`systemctl status` instead of `docker inspect`).
- Smaller attack surface (no privileged daemon).

The trade-off is that you cannot move a Velo app to another host by copying an image. For that, see the [upgrade](/velo-deploy/operations/upgrade/) flow.

## Tech stack

| Component | Technology | Purpose |
| --- | --- | --- |
| Language | Go 1.22+ | Single static binary, cross-compile friendly. |
| TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Terminal UI framework. |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) | TUI styling. |
| Process mgmt | systemd | App lifecycle, isolation. |
| Web server | [Caddy](https://caddyserver.com/) | Reverse proxy, automatic HTTPS. |
| Node.js | [nvm](https://github.com/nvm-sh/nvm) | Version management. |
