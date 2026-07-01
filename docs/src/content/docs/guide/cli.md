---
title: CLI
description: Every command, flag, and example for the velo-deploy CLI.
---

import { Tabs, TabItem, Aside } from '@astrojs/starlight/components';

The `velo-deploy` binary is the primary user interface. The TUI is a thin layer on top of these commands.

## Global flags

| Flag | Description |
| --- | --- |
| `-h`, `--help` | Show command-specific help. |
| `-V`, `--version` | Print the version and exit. |

## Commands

### `velo-deploy deploy`

Clone a Git repository, install dependencies, build, and register a new app.

```bash
velo-deploy deploy <repo-url> [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--name` | derived from repo | Override the app name. |
| `--domain` | _empty_ | Map a real domain to the app. Triggers Let's Encrypt via Caddy. |
| `--alias` | `<name>.local` | Local alias for path-based access. |
| `--port` | auto (3000-3999) | Internal Node.js port. Ignored for static sites. |
| `--branch` | repo default | Branch to track. |

**Examples**

```bash
velo-deploy deploy https://github.com/user/api
velo-deploy deploy https://github.com/user/api --domain api.example.com
velo-deploy deploy https://github.com/user/api --alias myapi.local --port 3100
velo-deploy deploy https://github.com/user/site --name marketing-site
```

### `velo-deploy add`

Register an application that already lives on disk. Useful for repos you cloned manually, or for restoring from a backup.

```bash
velo-deploy add <name> <path> [flags]
```

| Flag | Description |
| --- | --- |
| `--type` | `node` or `static`. Auto-detected when omitted. |
| `--domain` | Real domain to map. |
| `--alias` | Local alias (defaults to `<name>.local`). |
| `--port` | Node.js port (3000-3999 range). |

**Examples**

```bash
velo-deploy add my-site /opt/deploy/apps/my-site
velo-deploy add my-site /opt/deploy/apps/my-site --type static
velo-deploy add my-api  /opt/deploy/apps/my-api  --type node --port 3200
velo-deploy add my-site /opt/deploy/apps/my-site --domain mysite.com
```

### `velo-deploy list`

List all registered apps with their type, domain, alias, and status.

```bash
velo-deploy list
```

Output:

```text
NAME          TYPE     DOMAIN              ALIAS                STATUS
my-api        node     api.example.com     my-api.local         running
my-site       static   mysite.com          my-site.local        running
hello-velo    node     -                   hello-velo.local     stopped
```

### `velo-deploy restart`

Restart a Node.js app by reloading its systemd unit. No-op for static sites (Caddy serves them straight from disk).

```bash
velo-deploy restart <name>
```

### `velo-deploy stop`

Stop a Node.js app. The systemd unit stays installed; the process is just terminated.

```bash
velo-deploy stop <name>
```

### `velo-deploy logs`

View the last 200 lines of an app's log, or follow it in real time.

```bash
velo-deploy logs <name> [flags]
```

| Flag | Description |
| --- | --- |
| `-f`, `--follow` | Follow log output (like `tail -f`). |
| `-n`, `--lines` | Number of lines to show (default `200`). |

**Examples**

```bash
velo-deploy logs my-api
velo-deploy logs my-api -f
velo-deploy logs my-api -n 1000
```

### `velo-deploy remove`

Unregister an app, remove its systemd unit, and delete its Caddy vhost. Source code under `/opt/deploy/apps/<name>` is kept by default.

```bash
velo-deploy remove <name> [flags]
```

| Flag | Description |
| --- | --- |
| `--purge` | Also delete the app source directory. |

### `velo-deploy daemon`

Start the webhook daemon in the foreground. In production, use the `velo-deploy-watcher` systemd unit installed by `install.sh`.

```bash
velo-deploy daemon --port 9999
```

<Aside type="note" title="What the daemon is for">
	The daemon listens for GitHub webhook deliveries and triggers a redeploy when a push arrives on the configured branch. See [Webhooks](/velo-deploy/guide/webhooks/).
</Aside>

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success. |
| `1` | Generic error. Inspect stderr. |
| `2` | Invalid usage (missing flag, unknown command). |
| `3` | App not found. |
| `4` | Permission denied. Re-run with `sudo` or as `root`. |
| `5` | Port already in use. |

## Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `VELO_CONFIG` | `/etc/velo-deploy/config.json` | Override the config file path. |
| `VELO_LOGS_DIR` | `/var/log/velo-deploy` | Override the logs directory. |
| `VELO_APPS_DIR` | `/opt/deploy/apps` | Override the apps directory. |
| `NO_COLOR` | _unset_ | Disable ANSI colors (also `velo-deploy --no-color`). |
