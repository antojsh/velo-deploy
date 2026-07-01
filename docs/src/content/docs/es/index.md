---
title: Velo Deploy
description: Desplegá Node.js y sitios estáticos en cualquier VPS sin contenedores, Kubernetes o Docker.
template: splash
hero:
  title: Desplegá sin el barco de contenedores
  tagline: Un VPS, un comando, un binario. Sin contenedores, sin Kubernetes — solo systemd, Caddy y un runtime de Go que no se mete en tu camino.
  actions:
    - text: Empezar
      link: getting-started/installation/
      icon: right-arrow
      variant: primary
    - text: Ver en GitHub
      link: https://github.com/antojsh/velo-deploy
      icon: external
      variant: minimal
---

## Deployá en tres comandos

De un VPS Ubuntu nuevo a una web con HTTPS en vivo en menos de un minuto. Sin `docker-compose.yml`, sin Helm chart, sin pipeline de CI.

```bash
# 1. Instalá una vez
curl -sS https://get.velo-deploy.sh | bash

# 2. Deployá cualquier repo
velo-deploy deploy https://github.com/tu/repo

# 3. Push para redesplegar, automático
git push origin main
```

## Por qué Velo Deploy

- **HTTPS automático** — Caddy emite y renueva certificados de Let's Encrypt para cada dominio. Apuntá un dominio a tu VPS y HTTPS funciona solo.
- **Aislamiento con systemd** — Cada app es su propio usuario Linux con unidad endurecida: `ProtectSystem=strict`, `PrivateTmp`, `NoNewPrivileges`, límites de cgroups.
- **Git push para deployar** — Conectá el watcher de webhooks y cada push a `main` redespliega solo. Sin CI, sin GitHub Actions, sin servicios de terceros.
- **Un binario de Go, ~15 MB** — Sin runtime, sin container daemon, sin control plane. Velo suma menos a tu VPS que un solo proceso de Node.

## ¿Listo para deployar?

- [Instalar Velo Deploy](getting-started/installation/) — Requisitos y el instalador de una línea
- [Deploy rápido en 30 segundos](getting-started/quick-deploy/) — Salteate la lectura, deployá una app de ejemplo ahora
- [Entender la arquitectura](architecture/overview/) — Cómo encajan systemd, Caddy y el binario de Go
- [Ver la referencia de CLI](reference/cli-reference/) — Cada comando, flag y código de salida
