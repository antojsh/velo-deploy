---
title: Arquitectura
description: Cómo encajan systemd, Caddy y el binario de Go.
---

Velo Deploy es intencionalmente chico. Tres componentes móviles, un solo archivo de config, sin bases de datos.

```
┌─────────────────────────────────────────────────────────────┐
│                        Velo Deploy                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   GitHub ──webhook──▶ velo-deploy-watcher (puerto 9999)    │
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
│   │   Apps Node.js      │      │   Sitios estáticos  │      │
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

## Componentes

### `velo-deploy` (binario Go)

Un único binario estático que trae la CLI, la TUI y los scripts de install/uninstall embebidos. **No** corre como daemon de larga duración — cada comando es de vida corta e idempotente.

La CLI habla con systemd a través de `systemctl` y con Caddy escribiendo archivos en `/etc/caddy/conf.d/`.

### `velo-deploy-watcher` (binario Go, larga duración)

Un segundo binario, también escrito en Go, que escucha en el puerto `9999` entregas de webhooks de GitHub. Es el único proceso de larga duración que Velo Deploy agrega al sistema. Está registrado como la unidad systemd `velo-deploy-watcher.service`.

### systemd

El process manager. Cada app desplegada obtiene una unidad dedicada en `/etc/systemd/system/velo-<app>.service`. Las unidades están endurecidas con `ProtectSystem`, `ProtectHome`, `PrivateTmp` y `NoNewPrivileges` — mirá [Seguridad](/velo-deploy/es/architecture/security/).

### Caddy

El servidor web y terminador de TLS. Velo escribe un vhost por app en `/etc/caddy/conf.d/<app>.conf` y dispara `caddy reload` después de cada cambio. Caddy maneja ACME (Let's Encrypt) automáticamente.

### nvm

Gestor de versiones de Node.js instalado en `/opt/nvm`. Velo instala la versión pedida por `engines.node` y pinea el binario de `node` resultante en la config de la app.

## Flujo de datos

1. **Deploy** — la CLI clona el repo, instala dependencias, corre el build, escribe una unidad systemd, escribe un vhost de Caddy y registra la app en `/etc/velo-deploy/config.json`.
2. **Runtime** — Caddy termina TLS, hace reverse-proxy a los requests Node.js, o sirve archivos estáticos directo.
3. **Webhook** — `velo-deploy-watcher` recibe un evento push, identifica la app que matchea y re-corre los pasos de deploy para esa app.
4. **Tear-down** — `velo-deploy remove` borra la unidad systemd, el vhost de Caddy, la entrada en config y (opcionalmente) el directorio de código.

## ¿Por qué no Docker?

Los contenedores son geniales para portabilidad, pero la mayoría de los workloads Node.js y estáticos no necesitan esa portabilidad. Saltarse la capa de containers significa:

- Menor overhead de memoria (sin `containerd` ni `runc`).
- Cold starts más rápidos (sin pull de imagen).
- Debugging más simple (`systemctl status` en vez de `docker inspect`).
- Menor superficie de ataque (sin daemon privilegiado).

El trade-off es que no podés mover una app de Velo a otro host copiando una imagen. Para eso, mirá el flujo de [upgrade](/velo-deploy/es/operations/upgrade/).

## Stack técnico

| Componente | Tecnología | Propósito |
| --- | --- | --- |
| Lenguaje | Go 1.22+ | Un único binario estático, compilación cruzada amigable. |
| TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Framework de UI terminal. |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) | Estilado de TUI. |
| Process mgmt | systemd | Lifecycle y aislamiento de apps. |
| Web server | [Caddy](https://caddyserver.com/) | Reverse proxy, HTTPS automático. |
| Node.js | [nvm](https://github.com/nvm-sh/nvm) | Gestión de versiones. |
