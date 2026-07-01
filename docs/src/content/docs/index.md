---
title: Velo Deploy
description: Deploy Node.js and static sites to any VPS without containers, Kubernetes, or Docker.
template: splash
hero:
  title: Deploy without the container ship
  tagline: One VPS, one command, one binary. No containers, no Kubernetes — just systemd, Caddy, and a Go runtime that gets out of your way.
  actions:
    - text: Get started
      link: getting-started/installation/
      icon: right-arrow
      variant: primary
    - text: View on GitHub
      link: https://github.com/antojsh/velo-deploy
      icon: external
      variant: minimal
---

## Ship in three commands

From a fresh Ubuntu VPS to a live HTTPS site in under a minute. No `docker-compose.yml`, no Helm chart, no CI pipeline.

```bash
# 1. Install once
curl -sS https://get.velo-deploy.sh | bash

# 2. Deploy any repo
velo-deploy deploy https://github.com/your/repo

# 3. Push to ship again, automatically
git push origin main
```

## Why Velo Deploy

- **Automatic HTTPS** — Caddy issues and renews Let's Encrypt certs for every domain. Point a domain at your VPS and HTTPS just works.
- **systemd isolation** — Each app is its own Linux user with a hardened unit: `ProtectSystem=strict`, `PrivateTmp`, `NoNewPrivileges`, cgroup resource limits.
- **Git push to ship** — Wire the built-in webhook watcher and every push to `main` redeploys automatically. No CI, no GitHub Actions, no third-party service.
- **One Go binary, ~15 MB** — No runtime, no container daemon, no control plane. Velo adds less to your VPS than a single Node process.

## Ready to ship?

- [Install Velo Deploy](getting-started/installation/) — Requirements and the one-liner installer
- [Quick deploy in 30 seconds](getting-started/quick-deploy/) — Skip the reading, deploy a sample app now
- [Understand the architecture](architecture/overview/) — How systemd, Caddy, and the Go binary fit together
- [Browse the CLI reference](reference/cli-reference/) — Every command, flag, and exit code