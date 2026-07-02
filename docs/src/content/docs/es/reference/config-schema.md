---
title: Schema de config
description: El schema completo de /etc/velo-deploy/config.json, con tipos y defaults.
---

Esta página es la referencia autoritativa del schema. Para un walkthrough de alto nivel, mirá [Configuración](/velo-deploy/es/guide/configuration/).

## Nivel superior

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

## Config de app

```ts
type AppConfig = {
  name: string;
  type: 'node' | 'static';
  domain?: string;
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

| Campo | Default |
| --- | --- |
| `caddy_conf_dir` | `/etc/caddy/conf.d` |
| `apps_dir` | `/opt/deploy/apps` |
| `logs_dir` | `/var/log/velo-deploy` |
| `daemon_port` | `9999` |
| `nvm_dir` | `/opt/nvm` |
| `apps.<name>.alias` | `<name>.local` |
| `apps.<name>.branch` | branch default del repo |
| `apps.<name>.node_version` | `24` |
| `apps.<name>.port` | primer libre en `3000-3999` |
| `apps.<name>.output_dir` | primer existente de `dist`, `build`, `_site`, `public`, `output` |

## Reglas de validación

- `name` debe matchear `^[a-z0-9][a-z0-9-]{0,62}$` (lowercase, dígitos, guiones).
- `name` debe ser único entre todas las `apps`.
- `port` debe ser un entero en `3000-3999` para apps Node.js.
- `domain` debe ser un nombre DNS válido (sin scheme, sin path).
- `alias` debe ser un nombre DNS válido.
- `node_version` debe ser uno de `16`, `18`, `20`, `22`, `24`.
- `entry_point` debe ser un path relativo adentro del dir de la app.
- `output_dir` debe ser un path relativo adentro del dir de la app.

## Ejemplo

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

## Versionado

El schema del config está versionado con [SemVer](https://semver.org/). La mayor versión actual es `1`. Los campos nuevos son aditivos — los configs viejos siguen funcionando. Los breaking changes se anuncian en el [CHANGELOG](https://github.com/antojsh/velo-deploy/blob/master/CHANGELOG.md) al menos un minor release antes de salir.
