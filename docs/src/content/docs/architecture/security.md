---
title: Security
description: How Velo isolates apps and what you can rely on.
---

import { Aside } from '@astrojs/starlight/components';

Velo Deploy inherits its security model from systemd. Each app runs in a hardened unit with no access to the rest of the system.

## Per-app Linux user

When an app is deployed, Velo creates a dedicated Linux user with no password, no login shell, and no home directory:

```
velo-<appname>:x:1001:1001::/nonexistent:/usr/sbin/nologin
```

The app runs under this user. It cannot read other users' files, and the OS will never present a login prompt for it.

## Hardened systemd unit

Each unit file includes the following directives:

| Directive | Effect |
| --- | --- |
| `ProtectSystem=full` | `/usr`, `/boot`, `/etc` are mounted read-only. |
| `ProtectHome=true` | The app cannot read `/home`, `/root`, or `/run/user`. |
| `PrivateTmp=true` | The app gets its own `/tmp` namespace. |
| `NoNewPrivileges=true` | The app cannot escalate privileges with setuid binaries. |
| `ReadWritePaths=` | Writable paths are limited to the app directory and `/tmp`. |
| `PrivateDevices=true` | The app cannot talk to raw devices. |
| `ProtectKernelTunables=true` | `/proc` and `/sys` writes are blocked. |
| `RestrictNamespaces=true` | The app cannot create new namespaces. |
| `MemoryDenyWriteExecute=true` | W^X memory policy (where the kernel supports it). |

A full example:

```ini
[Unit]
Description=Velo Deploy app: my-api
After=network-online.target

[Service]
Type=simple
User=velo-my-api
Group=velo-my-api
WorkingDirectory=/opt/deploy/apps/my-api
ExecStart=/opt/nvm/versions/node/v20.10.0/bin/node index.js
Restart=on-failure
RestartSec=5

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=true
ReadWritePaths=/opt/deploy/apps/my-api /tmp
PrivateDevices=true
ProtectKernelTunables=true
RestrictNamespaces=true
MemoryDenyWriteExecute=true

[Install]
WantedBy=multi-user.target
```

## Filesystem isolation

The app can:

- Read and write `/opt/deploy/apps/<app>/`.
- Read and write `/tmp` (private namespace).
- Read everything in `/usr`, `/etc`, `/var` (mounted read-only).
- Make outbound network connections.

The app **cannot**:

- Read other users' home directories.
- Write anywhere outside its own directory (except its own `/tmp`).
- Spawn setuid processes.
- Mount filesystems.
- Talk to raw block devices.

## Network isolation

Caddy is the only process listening on a public port. Node.js apps bind to `localhost` only — they are not reachable directly from the network. All traffic flows through the Caddy reverse proxy, which terminates TLS and forwards to the right upstream.

## Secrets

<Aside type="warning" title="Do not commit secrets to Git">
	Velo Deploy does not manage secrets. Pass them as environment variables in the systemd unit, never as part of the repository.
</Aside>

To inject environment variables, edit the unit file and add an `Environment=` line, or use an `EnvironmentFile=` directive:

```ini
[Service]
Environment="DATABASE_URL=postgres://..."
EnvironmentFile=/etc/velo-deploy/my-api.env
```

Then reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart velo-my-api
```

## Webhook security

See [Webhooks](/velo-deploy/guide/webhooks/#security) for the recommended HMAC validation and IP allow-listing pattern.

## Supply chain

Velo's Go dependencies are pinned in `go.sum` and verified on every build. The release pipeline signs binaries with [cosign](https://github.com/sigstore/cosign) and the install script checks the signature before installing. See the [release-please configuration](https://github.com/antojsh/velo-deploy/blob/master/release-please-config.json) for the exact pipeline.

## Reporting a vulnerability

See [SECURITY.md](https://github.com/antojsh/velo-deploy/blob/master/SECURITY.md) on GitHub.
