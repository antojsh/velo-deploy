---
title: Troubleshooting
description: Diagnose the most common production issues.
---

import { Aside } from '@astrojs/starlight/components';

## App not starting

```bash
sudo systemctl status velo-myapp
sudo journalctl -u velo-myapp -n 200 --no-pager
velo-deploy logs myapp
```

Common causes:

- **Wrong `node_path`** — the `node_version` changed but the path did not. Edit `/etc/velo-deploy/config.json` and update `node_path`.
- **Missing dependency** — `node_modules` is empty after a `git pull`. Run `velo-deploy deploy <repo> --force`.
- **Port already in use** — exit code `5`. Find the process: `sudo ss -tlnp | grep ':3000'`.
- **Permission denied** — the app directory is not owned by `velo-<app>`. Fix with `sudo chown -R velo-myapp:velo-myapp /opt/deploy/apps/myapp`.

## App crashes immediately

```bash
sudo journalctl -u velo-myapp -e --no-pager
```

The journal usually shows a stack trace. Most common causes:

- A top-level `await` in an `index.js` with `type: "module"` not set.
- A missing environment variable — see [Security → Secrets](/velo-deploy/architecture/security/#secrets).
- A `node_modules` built for a different Node version after an upgrade.

## Webhook not firing

```bash
sudo systemctl status velo-watcher
sudo journalctl -u velo-watcher -f
```

Test the endpoint locally:

```bash
curl -X POST http://localhost:9999/webhook
```

If you get a connection refused, the daemon is not running. If you get a `404`, the route is wrong (it should be `/webhook`, not `/`).

If GitHub is sending events but nothing is happening on the server, check:

- The `repo_url` in `config.json` matches the URL GitHub sends.
- The branch is `main` or `master`.
- The deploy user has write access to `/opt/deploy/apps/<name>/`.

## HTTPS not working

```bash
sudo caddy validate --config /etc/caddy/Caddyfile
sudo caddy reload --config /etc/caddy/Caddyfile
sudo journalctl -u caddy -f
```

Common causes:

- **DNS not pointing at the server** — `dig +short <domain>` should return the server IP.
- **Port 80 blocked** — Let's Encrypt needs port 80 for HTTP-01 validation. Check with `curl -I http://<domain>/.well-known/acme-challenge/test`.
- **Rate limit** — Let's Encrypt allows 50 certs per week per domain. If you hit it, wait a week or use a different validation method.

## Port conflicts

```bash
sudo ss -tlnp | grep -E ':3[0-9]{3}'
```

If two apps try to use the same port, the second deploy fails with exit code `5`. Re-deploy with an explicit `--port`:

```bash
velo-deploy deploy https://github.com/user/api --port 3100
```

## Stale Caddy config after `remove`

```bash
sudo ls /etc/caddy/conf.d/
sudo rm /etc/caddy/conf.d/<app>.conf
sudo caddy reload --config /etc/caddy/Caddyfile
```

This should not happen on a normal `velo-deploy remove`, but if it does, the manual clean-up is safe.

## Disk full

```bash
du -sh /opt/deploy/apps/* | sort -h
du -sh /var/log/velo-deploy/* | sort -h
```

Most disk usage comes from `node_modules`. For Node.js apps, prune dev dependencies in production:

```bash
velo-deploy deploy https://github.com/user/api
sudo -u velo-user npm prune --production
```

For static sites, the build output is usually small.

## Logs eating disk

```bash
sudo journalctl --vacuum-size=100M
```

This caps the systemd journal at 100 MB. Add the same line to `/etc/systemd/journald.conf`:

```ini
[Journal]
SystemMaxUse=100M
```

Then restart:

```bash
sudo systemctl restart systemd-journald
```

## Getting more help

<Aside type="tip" title="Before opening an issue">
	Run `velo-deploy version` and `velo-deploy list --json` and include the output in your bug report. That alone will tell us most of what we need to know.
</Aside>

- [GitHub Issues](https://github.com/antojsh/velo-deploy/issues) for bugs and feature requests.
- [GitHub Discussions](https://github.com/antojsh/velo-deploy/discussions) for questions and ideas.
- [SECURITY.md](https://github.com/antojsh/velo-deploy/blob/master/SECURITY.md) for vulnerability reports.
