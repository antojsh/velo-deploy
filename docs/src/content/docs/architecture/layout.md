---
title: Layout
description: Filesystem layout and the data each directory holds.
---

A fresh Velo Deploy install produces the following layout on the server.

## System paths

```
/etc/velo-deploy/
├── config.json          # App metadata and configuration
├── apps/                # Deployed application source code
│   └── <app-name>/
└── logs/                # Per-app deploy logs (rolling)

/var/log/velo-deploy/    # System-level logs (velo-deploy-watcher, velo-deploy)

/etc/systemd/system/
├── velo-deploy-watcher.service   # Webhook watcher daemon
└── velo-<app>.service     # Per-app systemd units

/etc/caddy/
├── Caddyfile              # Root Caddyfile (loads conf.d)
└── conf.d/
    ├── <app>.conf         # Per-app Caddy configs (domain-based)
    └── _shared.conf       # Shared catch-all config (path-based)

/opt/nvm/                  # nvm install (Node versions live here)
/opt/deploy/               # Apps live here by default
```

## Runtime paths

| Path | Owner | Purpose |
| --- | --- | --- |
| `/etc/velo-deploy/config.json` | `root:root`, `0644` | Single source of truth. |
| `/opt/deploy/apps/<name>/` | `velo-<name>:velo-<name>` | App source code. Writable by the app user. |
| `/var/log/velo-deploy/<name>.log` | `velo-<name>:velo-<name>` | App stdout/stderr. |
| `/etc/systemd/system/velo-<name>.service` | `root:root` | App unit. |
| `/etc/caddy/conf.d/<name>.conf` | `root:root` | App vhost. |

## Network ports

| Port | Process | Purpose |
| --- | --- | --- |
| `80` | Caddy | HTTP, redirects to HTTPS. |
| `443` | Caddy | HTTPS, ACME, reverse proxy. |
| `9999` | velo-deploy-watcher | GitHub webhook receiver. |
| `3000-3999` | velo-<app> | Internal Node.js ports. Not exposed externally. |

## Process tree

```
systemd
├── caddy.service
├── velo-deploy-watcher.service
└── velo-<app>.service
    └── node /opt/deploy/apps/<app>/index.js
```

## Caddy integration

The root Caddyfile is a one-liner that loads everything in `conf.d`:

```nginx
import /etc/caddy/conf.d/*.conf
```

This lets Velo write and remove per-app vhosts without ever touching the root Caddyfile.

## Backup strategy

The two things worth backing up are:

1. `/etc/velo-deploy/config.json` — your app registry.
2. `/opt/deploy/apps/` — your app source code (only if you do not have it in Git).

Logs under `/var/log/velo-deploy/` are safe to delete. The systemd units and Caddy vhosts are reproducible from `config.json`, so they do not need to be backed up.

## Disk usage

A typical install uses:

- 80 MB for the Go binary and Caddy.
- 200 MB for nvm + Node 20.
- 100 MB per deployed app (mostly `node_modules`).

Plan for 5 GB of free disk on a small server.
