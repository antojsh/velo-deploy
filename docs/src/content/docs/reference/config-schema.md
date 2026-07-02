---
title: Config schema reference
description: The full /etc/velo-deploy/config.json schema, with types and defaults.
---

This page is the authoritative schema reference. For a high-level walkthrough, see [Configuration](/velo-deploy/guide/configuration/).

## Top-level

```ts
type Config = {
  caddy_conf_dir?: string;   // default: /etc/caddy/conf.d
  apps_dir?: string;         // default: /opt/deploy/apps
  logs_dir?: string;         // default: /var/log/velo-deploy
  daemon_port?: string;      // default: "9999"
  nvm_dir?: string;          // default: /opt/nvm
  apps: Record<string, AppConfig>;
};
```

## App config

```ts
type AppConfig = {
  name: string;
  type: 'node' | 'static';
  domain?: string;
  path?: string;
  alias?: string;
} & ({
  type: 'node';
  repo_url: string;
  branch?: string;
  node_version: string;
  port: number;
  node_path: string;
  entry_point: string;
} | {
  type: 'static';
  repo_url?: string;
  output_dir: string;
});
```

## Defaults

| Field | Default |
| --- | --- |
| `caddy_conf_dir` | `/etc/caddy/conf.d` |
| `apps_dir` | `/opt/deploy/apps` |
| `logs_dir` | `/var/log/velo-deploy` |
| `daemon_port` | `9999` |
| `nvm_dir` | `/opt/nvm` |
| `apps.<name>.alias` | `<name>.local` |
| `apps.<name>.path` | `/` with domain, `/<name>` without domain |
| `apps.<name>.branch` | repo default branch |
| `apps.<name>.node_version` | `24` |
| `apps.<name>.port` | first free in `3000-3999` |
| `apps.<name>.output_dir` | first existing of `dist`, `build`, `_site`, `public`, `output` |

## Validation rules

- `name` must match `^[a-z0-9][a-z0-9-]{0,62}$` (lowercase, digits, dashes).
- `name` must be unique across `apps`.
- `port` must be an integer in `3000-3999` for Node.js apps.
- `domain` must be a valid DNS name (no scheme, no path).
- `path` must be a URL path such as `/api` or `/admin`.
- `alias` must be a valid DNS name.
- `node_version` must be one of `16`, `18`, `20`, `22`, `24`.
- `entry_point` must be a relative path inside the app directory.
- `output_dir` must be a relative path inside the app directory.

## Example

```json
{
  "caddy_conf_dir": "/etc/caddy/conf.d",
  "apps_dir": "/opt/deploy/apps",
  "logs_dir": "/var/log/velo-deploy",
  "daemon_port": "9999",
  "nvm_dir": "/opt/nvm",
  "apps": {
    "api": {
      "name": "api",
      "type": "node",
      "repo_url": "https://github.com/acme/api",
      "branch": "main",
      "node_version": "24",
      "port": 3000,
      "domain": "api.acme.com",
      "alias": "api.local",
      "node_path": "/opt/nvm/versions/node/v24.18.0/bin/node",
      "entry_point": "dist/main.js"
    },
    "marketing": {
      "name": "marketing",
      "type": "static",
      "repo_url": "https://github.com/acme/marketing",
      "branch": "main",
      "domain": "acme.com",
      "alias": "marketing.local",
      "output_dir": "dist"
    }
  }
}
```

## Versioning

The config schema is versioned with [SemVer](https://semver.org/). The current major version is `1`. New fields are additive — old configs continue to work. Breaking changes are announced in the [CHANGELOG](https://github.com/antojsh/velo-deploy/blob/master/CHANGELOG.md) at least one minor release before they ship.
