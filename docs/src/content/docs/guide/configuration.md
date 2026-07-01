---
title: Configuration
description: The config.json schema, defaults, and override patterns.
---

import { Aside } from '@astrojs/starlight/components';

Velo Deploy stores its state in a single JSON file at `/etc/velo-deploy/config.json`. The file is human-editable; changes are picked up on the next CLI invocation.

## File location

```
/etc/velo-deploy/config.json
```

Override the path with the `VELO_CONFIG` environment variable.

## Full schema

```json
{
  "caddy_conf_dir": "/etc/caddy/conf.d",
  "apps_dir": "/opt/deploy/apps",
  "logs_dir": "/var/log/velo-deploy",
  "daemon_port": "9999",
  "nvm_dir": "/opt/nvm",
  "apps": {
    "my-api": {
      "name": "my-api",
      "type": "node",
      "repo_url": "https://github.com/user/my-api",
      "branch": "main",
      "node_version": "20",
      "port": 3000,
      "domain": "",
      "alias": "my-api.local",
      "node_path": "/opt/nvm/versions/node/v20.0.0/bin/node",
      "entry_point": "index.js"
    },
    "my-site": {
      "name": "my-site",
      "type": "static",
      "domain": "",
      "alias": "my-site.local",
      "output_dir": "dist"
    }
  }
}
```

## Top-level fields

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `caddy_conf_dir` | string | `/etc/caddy/conf.d` | Where per-app Caddy vhosts are written. |
| `apps_dir` | string | `/opt/deploy/apps` | Where app source code lives. |
| `logs_dir` | string | `/var/log/velo-deploy` | Where deploy logs are written. |
| `daemon_port` | string | `9999` | Port the webhook daemon binds to. |
| `nvm_dir` | string | `/opt/nvm` | nvm install location. |
| `apps` | object | `{}` | Map of app name to app config. |

## App fields

### Common

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | string | yes | Unique app identifier. |
| `type` | `"node"` \| `"static"` | yes | Runtime type. |
| `domain` | string | no | Public domain. Triggers Let's Encrypt. |
| `alias` | string | no | Local alias for path-based access. |

### Node.js only

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `repo_url` | string | yes | Git URL. |
| `branch` | string | no | Branch to track. Default: repo default. |
| `node_version` | string | no | Major version to use. Default: `20`. |
| `port` | number | yes | Internal port. Default: first free in 3000-3999. |
| `node_path` | string | yes | Absolute path to the Node binary. |
| `entry_point` | string | yes | File to start. |

### Static only

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `repo_url` | string | no | Git URL. Optional for `add`. |
| `output_dir` | string | yes | Build output directory. |

## Editing the file

<Aside type="caution" title="Validate before reloading">
	After editing `config.json` by hand, run `velo-deploy list` to make sure the JSON parses. A broken file will make every command fail with a parse error.
</Aside>

```bash
sudo jq . /etc/velo-deploy/config.json
velo-deploy list
```

For Node.js apps, restart the affected unit after editing:

```bash
sudo systemctl restart velo-myapp
```

For static sites, no restart is required — Caddy picks up vhost changes on `caddy reload` (which Velo triggers automatically).

## Versioning

The config format follows [Semantic Versioning](https://semver.org/). New fields are additive and ship with sensible defaults. Renames or removals ship in a major release and are announced in the [CHANGELOG](https://github.com/antojsh/velo-deploy/blob/master/CHANGELOG.md).
