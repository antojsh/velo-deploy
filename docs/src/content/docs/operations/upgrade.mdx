---
title: Upgrade
description: Upgrade the velo-deploy binary and the system-wide services.
---

import { Aside } from '@astrojs/starlight/components';

There are two things to keep up to date:

1. The `velo-deploy` binary (CLI, TUI, embedded assets).
2. The `velo-deploy-watcher` service (same binary, different entry point).

## Upgrade in place

The recommended way to upgrade is to re-run the one-liner installer. It is idempotent — running it on an existing install upgrades the binary and the watcher unit without touching your apps.

```bash
curl -sS https://github.com/antojsh/velo-deploy/releases/latest/download/velo-deploy-install.sh | bash
```

This:

- Backs up the existing binary to `/usr/local/bin/velo-deploy.bak`.
- Installs the new binary to `/usr/local/bin/velo-deploy`.
- Reloads the `velo-deploy-watcher` systemd unit.
- Leaves `/etc/velo-deploy/config.json` and `/opt/deploy/apps/` untouched.

To roll back:

```bash
sudo mv /usr/local/bin/velo-deploy.bak /usr/local/bin/velo-deploy
sudo systemctl restart velo-deploy-watcher
```

## Pin a specific version

If you need to upgrade to a specific version (for example, to test a release candidate):

```bash
VER=v0.4.0
curl -sSLo velo-deploy.tar.gz \
  "https://github.com/antojsh/velo-deploy/releases/download/${VER}/velo-deploy_${VER#v}_linux_amd64.tar.gz"
sudo install -m 0755 velo-deploy /usr/local/bin/velo-deploy
sudo systemctl restart velo-deploy-watcher
velo-deploy version
```

## Upgrade an app

Apps are versioned independently of the platform. To upgrade a single app:

```bash
velo-deploy deploy https://github.com/user/api --force
```

The `--force` flag re-clones, reinstalls, and restarts. Without it, Velo only re-deploys when the upstream HEAD has changed since the last deploy.

## Upgrading Node.js

To upgrade the Node.js version of a single app, bump the `engines.node` field in `package.json` and push. The next deploy installs the new version automatically.

## Upgrading Caddy

Caddy ships its own self-upgrade mechanism. To upgrade:

```bash
sudo caddy upgrade
```

This downloads the latest Caddy binary, verifies the signature, and restarts the service with zero downtime.

<Aside type="caution" title="Major version upgrades may break vhosts">
	Velo's generated vhosts are written for Caddy 2.x. They continue to work on Caddy 2.x but may need regeneration on Caddy 3.x. Run `velo-deploy list` after the upgrade to make sure every vhost validates.
</Aside>

## Verifying an upgrade

```bash
velo-deploy version
systemctl status velo-deploy-watcher
systemctl status caddy
sudo caddy validate --config /etc/caddy/Caddyfile
velo-deploy list
```

If all five commands return clean, the upgrade is complete.
