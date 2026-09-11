---
title: Layout
description: Layout del filesystem y qué datos guarda cada directorio.
---

Una instalación fresca de Velo Deploy produce el siguiente layout en el servidor.

## Paths del sistema

```
/etc/velo-deploy/
├── config.json          # Metadata y configuración de las apps
├── webhook.secret       # Secreto HMAC de webhooks de GitHub
├── apps/                # Secretos EnvironmentFile por app
│   └── <app-name>.env
└── caddy/               # Snippets de Caddy que sobreviven un redeploy
    └── <app-name>.snippet

/var/log/velo-deploy/    # Logs a nivel sistema (velo-deploy-watcher, velo-deploy)

/etc/systemd/system/
├── velo-deploy-watcher.service   # Daemon watcher de webhooks
└── velo-<app>.service     # Unidades systemd por app

/etc/caddy/
├── Caddyfile              # Caddyfile raíz (carga conf.d)
└── conf.d/
    ├── <app>.conf         # Configs de Caddy por app (basadas en dominio)
    └── _shared.conf       # Config compartida catch-all (basada en path)

/opt/nvm/                  # fallback de nvm para installs viejos
/opt/deploy/
├── apps/                  # Código de las apps (default)
└── node/<major>/          # Tarballs oficiales de Node
```

## Paths de runtime

| Path | Owner | Propósito |
| --- | --- | --- |
| `/etc/velo-deploy/config.json` | `root:root`, `0644` | Única fuente de verdad. |
| `/opt/deploy/apps/<name>/` | `velo-<name>:velo-<name>` | Código fuente de la app. Writable por el user de la app. |
| `/var/log/velo-deploy/<name>.log` | `velo-<name>:velo-<name>` | stdout/stderr de la app. |
| `/etc/systemd/system/velo-<name>.service` | `root:root` | Unidad de la app. |
| `/etc/caddy/conf.d/<name>.conf` | `root:root` | Vhost de la app. |

## Puertos de red

| Puerto | Proceso | Propósito |
| --- | --- | --- |
| `80` | Caddy | HTTP, redirige a HTTPS. |
| `443` | Caddy | HTTPS, ACME, reverse proxy. |
| `9999` | velo-deploy-watcher | Receptor de webhooks de GitHub. |
| `3000-3999` | velo-<app> | Puertos internos de Node.js. No expuestos externamente. |

## Árbol de procesos

```
systemd
├── caddy.service
├── velo-deploy-watcher.service
└── velo-<app>.service
    └── node /opt/deploy/apps/<app>/index.js
```

## Integración con Caddy

El Caddyfile raíz es un one-liner que carga todo en `conf.d`:

```nginx
import /etc/caddy/conf.d/*.conf
```

Esto permite que Velo escriba y remueva vhosts por app sin tocar nunca el Caddyfile raíz.

## Estrategia de backup

Las dos cosas que vale la pena backupear son:

1. `/etc/velo-deploy/config.json` — tu registro de apps.
2. `/opt/deploy/apps/` — el código fuente de tus apps (solo si no lo tenés en Git).

Los logs bajo `/var/log/velo-deploy/` se pueden borrar sin problema. Las unidades systemd y los vhosts de Caddy son reproducibles desde `config.json`, así que no necesitan backup.

## Uso de disco

Una instalación típica usa:

- 80 MB para el binario de Go y Caddy.
- 200 MB para Node 24 en `/opt/deploy/node`.
- 100 MB por app desplegada (mayormente `node_modules`).

Planeá 5 GB de disco libre en un servidor chico.
