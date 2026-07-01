---
title: Upgrade
description: Upgradear el binario de velo-deploy y los servicios a nivel sistema.
---

import { Aside } from '@astrojs/starlight/components';

Hay dos cosas que mantener al día:

1. El binario `velo-deploy` (CLI, TUI, assets embebidos).
2. El servicio `velo-deploy-watcher` (el mismo binario, distinto entry point).

## Upgrade in-place

La forma recomendada de upgradear es re-correr el instalador de una línea. Es idempotente — correrlo en una instalación existente upgraddea el binario y la unidad del watcher sin tocar tus apps.

```bash
curl -sS https://github.com/antojsh/velo-deploy/releases/latest/download/velo-deploy-install.sh | bash
```

Esto:

- Backupea el binario existente a `/usr/local/bin/velo-deploy.bak`.
- Instala el nuevo binario en `/usr/local/bin/velo-deploy`.
- Recarga la unidad systemd de `velo-deploy-watcher`.
- Deja `/etc/velo-deploy/config.json` y `/opt/deploy/apps/` intactos.

Para hacer rollback:

```bash
sudo mv /usr/local/bin/velo-deploy.bak /usr/local/bin/velo-deploy
sudo systemctl restart velo-deploy-watcher
```

## Pinear una versión específica

Si necesitás upgradear a una versión específica (por ejemplo, para testear un release candidate):

```bash
VER=v0.4.0
curl -sSLo velo-deploy.tar.gz \
  "https://github.com/antojsh/velo-deploy/releases/download/${VER}/velo-deploy_${VER#v}_linux_amd64.tar.gz"
sudo install -m 0755 velo-deploy /usr/local/bin/velo-deploy
sudo systemctl restart velo-deploy-watcher
velo-deploy version
```

## Upgradear una app

Las apps se versionan de forma independiente de la plataforma. Para upgradear una sola app:

```bash
velo-deploy deploy https://github.com/user/api --force
```

El flag `--force` re-clona, reinstala y reinicia. Sin él, Velo solo re-despliega cuando el HEAD upstream haya cambiado desde el último deploy.

## Upgradear Node.js

Para upgradear la versión de Node.js de una sola app, bumpeá el campo `engines.node` en `package.json` y hacé push. El próximo deploy instala la nueva versión automáticamente.

## Upgradear Caddy

Caddy trae su propio mecanismo de self-upgrade. Para upgradear:

```bash
sudo caddy upgrade
```

Esto descarga el último binario de Caddy, verifica la firma y reinicia el servicio con cero downtime.

<Aside type="caution" title="Los upgrades de mayor versión pueden romper vhosts">
	Los vhosts generados por Velo están escritos para Caddy 2.x. Siguen funcionando en Caddy 2.x pero pueden necesitar regeneración en Caddy 3.x. Corré `velo-deploy list` después del upgrade para asegurarte de que cada vhost valida.
</Aside>

## Verificar un upgrade

```bash
velo-deploy version
systemctl status velo-deploy-watcher
systemctl status caddy
sudo caddy validate --config /etc/caddy/Caddyfile
velo-deploy list
```

Si los cinco comandos vuelven limpios, el upgrade está completo.
