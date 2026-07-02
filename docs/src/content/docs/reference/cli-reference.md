---
title: CLI reference
description: Every command, flag, and exit code.
---

This page is the authoritative CLI reference. For task-oriented examples, see [CLI guide](/velo-deploy/guide/cli/).

## Synopsis

```bash
velo-deploy [global flags] <command> [command flags] [args]
```

## Global flags

| Flag | Description |
| --- | --- |
| `-h`, `--help` | Show help for the current command. |
| `-V`, `--version` | Print version and exit. |
| `--no-color` | Disable ANSI colors. |
| `--config <path>` | Override the config file path. Defaults to `/etc/velo-deploy/config.json`. |

## Commands

### `deploy`

```bash
velo-deploy deploy <repo-url> [flags]
```

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--name` | string | derived | Override the app name. |
| `--domain` | string | _empty_ | Public domain. |
| `--alias` | string | `<name>.local` | Local alias. |
| `--port` | int | auto | Internal port (Node.js only). |
| `--branch` | string | repo default | Branch to track. |
| `--output-dir` | string | auto | Build output dir (static only). |
| `--build-command` | string | `npm run build` | Build command. Alias: `--build-cmd`. |
| `--start-command` | string | `npm run start` | systemd start command. Alias: `--start-cmd`. |

### `add`

```bash
velo-deploy add <name> <path> [flags]
```

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--type` | `node` \| `static` | auto-detect | App type. |
| `--domain` | string | _empty_ | Public domain. |
| `--alias` | string | `<name>.local` | Local alias. |
| `--port` | int | auto | Internal port (Node.js only). |
| `--output-dir` | string | auto | Build output dir (static only). |

### `list`

```bash
velo-deploy list [flags]
```

| Flag | Description |
| --- | --- |
| `--json` | Output as JSON for scripting. |
| `--quiet` | Print names only, one per line. |

### `restart`

```bash
velo-deploy restart <name>
```

### `stop`

```bash
velo-deploy stop <name>
```

### `logs`

```bash
velo-deploy logs <name> [flags]
```

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-f`, `--follow` | bool | `false` | Follow log output. |
| `-n`, `--lines` | int | `200` | Lines to show. |

### `remove`

```bash
velo-deploy remove <name> [flags]
```

| Flag | Description |
| --- | --- |
| `--purge` | Also delete the app source directory. |

### `daemon`

```bash
velo-deploy daemon [flags]
```

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `--port` | int | `9999` | Port to listen on. |
| `--host` | string | `0.0.0.0` | Address to bind to. |

## Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `VELO_CONFIG` | `/etc/velo-deploy/config.json` | Override config path. |
| `VELO_LOGS_DIR` | `/var/log/velo-deploy` | Override logs dir. |
| `VELO_APPS_DIR` | `/opt/deploy/apps` | Override apps dir. |
| `NO_COLOR` | _unset_ | Disable ANSI colors. |

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success. |
| `1` | Generic error. |
| `2` | Invalid usage. |
| `3` | App not found. |
| `4` | Permission denied. |
| `5` | Port already in use. |
| `6` | Git operation failed. |
| `7` | Build failed. |
| `8` | Service did not become healthy in time. |
