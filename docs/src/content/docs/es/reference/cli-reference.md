---
title: Referencia de CLI
description: Cada comando, flag y código de salida.
---

Esta página es la referencia autoritativa de la CLI. Para ejemplos orientados a tareas, mirá la [guía de CLI](/velo-deploy/es/guide/cli/).

## Synopsis

```bash
velo-deploy [flags globales] <comando> [flags del comando] [args]
```

## Flags globales

| Flag | Descripción |
| --- | --- |
| `-h`, `--help` | Muestra la ayuda del comando actual. |
| `-V`, `--version` | Imprime la versión y sale. |
| `--no-color` | Deshabilita colores ANSI. |
| `--config <path>` | Override del path del config. Default: `/etc/velo-deploy/config.json`. |

## Comandos

### `deploy`

```bash
velo-deploy deploy <repo-url> [flags]
```

| Flag | Tipo | Default | Descripción |
| --- | --- | --- | --- |
| `--name` | string | derivado | Override del nombre de la app. |
| `--domain` | string | _vacío_ | Dominio público. |
| `--alias` | string | `<name>.local` | Alias local. |
| `--port` | int | auto | Puerto interno (solo Node.js). |
| `--branch` | string | default del repo | Branch a trackear. |
| `--output-dir` | string | auto | Dir de output del build (solo estáticos). |
| `--build-command` | string | `npm run build` | Comando de build. Alias: `--build-cmd`. |
| `--start-command` | string | `npm run start` | Comando de start para systemd. Alias: `--start-cmd`. |

### `add`

```bash
velo-deploy add <name> <path> [flags]
```

| Flag | Tipo | Default | Descripción |
| --- | --- | --- | --- |
| `--type` | `node` \| `static` | auto-detect | Tipo de app. |
| `--domain` | string | _vacío_ | Dominio público. |
| `--alias` | string | `<name>.local` | Alias local. |
| `--port` | int | auto | Puerto interno (solo Node.js). |
| `--output-dir` | string | auto | Dir de output del build (solo estáticos). |

### `list`

```bash
velo-deploy list [flags]
```

| Flag | Descripción |
| --- | --- |
| `--json` | Salida en JSON para scripts. |
| `--quiet` | Imprime solo los nombres, uno por línea. |

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

| Flag | Tipo | Default | Descripción |
| --- | --- | --- | --- |
| `-f`, `--follow` | bool | `false` | Sigue la salida del log. |
| `-n`, `--lines` | int | `200` | Líneas a mostrar. |

### `remove`

```bash
velo-deploy remove <name> [flags]
```

| Flag | Descripción |
| --- | --- |
| `--purge` | Borra también el dir de código de la app. |

### `daemon`

```bash
velo-deploy daemon [flags]
```

| Flag | Tipo | Default | Descripción |
| --- | --- | --- | --- |
| `--port` | int | `9999` | Puerto en el que escucha. |
| `--host` | string | `0.0.0.0` | Address al que se bindea. |

## Variables de entorno

| Variable | Default | Descripción |
| --- | --- | --- |
| `VELO_CONFIG` | `/etc/velo-deploy/config.json` | Override del path del config. |
| `VELO_LOGS_DIR` | `/var/log/velo-deploy` | Override del dir de logs. |
| `VELO_APPS_DIR` | `/opt/deploy/apps` | Override del dir de apps. |
| `NO_COLOR` | _unset_ | Deshabilita colores ANSI. |

## Códigos de salida

| Código | Significado |
| --- | --- |
| `0` | Éxito. |
| `1` | Error genérico. |
| `2` | Uso inválido. |
| `3` | App no encontrada. |
| `4` | Permiso denegado. |
| `5` | Puerto ya en uso. |
| `6` | Operación de Git falló. |
| `7` | Build falló. |
| `8` | El servicio no se volvió healthy a tiempo. |
