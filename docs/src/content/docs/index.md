---
title: Velo Deploy
description: Deploy Node.js apps and static sites to any VPS with systemd isolation, automatic HTTPS, and no Docker daemon.
template: splash
hero:
  title: Ship Node.js like a serious Linux service
  tagline: >-
    Velo Deploy turns a plain VPS into a tiny platform for Node.js apps and
    static sites: one Go binary, Caddy HTTPS, systemd units, Linux users, and
    zero Docker tax.
  actions:
    - text: Deploy in 30 seconds
      link: getting-started/quick-deploy/
      icon: right-arrow
      variant: primary
    - text: See the architecture
      link: architecture/overview/
      icon: external
      variant: minimal
---

<div class="velo-pill-row">
  <span class="velo-pill">No Docker daemon</span>
  <span class="velo-pill">No Kubernetes ceremony</span>
  <span class="velo-pill">systemd-native deploys</span>
  <span class="velo-pill">Automatic HTTPS</span>
  <span class="velo-pill">Linux-user isolation</span>
</div>

## The VPS deploy layer for people who understand production

You do not need a container ship to move a bicycle. For many Node.js APIs, Astro builds, dashboards, admin panels, and static sites, the cleanest deploy target is still Linux: a process supervised by `systemd`, behind Caddy, running as its own user.

Velo Deploy packages that boring, powerful foundation into a workflow developers can actually use.

<div class="velo-command-panel">
  <div class="velo-command-header">
    <span class="velo-dot"></span>
    <span class="velo-dot"></span>
    <span class="velo-dot"></span>
    <span>fresh-vps → live HTTPS app</span>
  </div>

<pre><code class="language-bash">curl -sS https://github.com/antojsh/velo-deploy/releases/latest/download/velo-deploy-install.sh | bash
velo-deploy deploy https://github.com/your/node-or-static-site
git push origin main</code></pre>

</div>

## Why deploy with Velo?

<div class="velo-grid">
  <div class="velo-card">
    <span class="velo-icon">01</span>
    <strong>No Docker tax</strong>
    <p>Skip the daemon, image builds, registries, compose files, and container networking for apps that only need a reliable Linux service.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">02</span>
    <strong>Secure Linux isolation</strong>
    <p>Each app runs as its own Linux user with hardened systemd settings like <code>ProtectSystem</code>, <code>PrivateTmp</code>, and <code>NoNewPrivileges</code>.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">03</span>
    <strong>Automatic HTTPS</strong>
    <p>Caddy handles certificates and renewals. Point DNS to the VPS, deploy the app, and stop babysitting TLS.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">04</span>
    <strong>Node.js and static sites</strong>
    <p>Deploy long-running Node.js services or static output directories like <code>dist</code>, <code>build</code>, and <code>public</code>.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">05</span>
    <strong>Push-to-deploy</strong>
    <p>The webhook watcher can redeploy on every push. Simple Git workflow, no external platform required.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">06</span>
    <strong>One small Go binary</strong>
    <p>No runtime control plane. Velo adds less operational weight than the average Node.js process.</p>
  </div>
</div>

## PM2 is not a deployment architecture

<div class="velo-quote">
PM2 is useful, but putting a Node.js process manager in charge of other Node.js processes is like hiring a thief to guard another thief. Sometimes it works. It is not the foundation you want for a VPS platform.
</div>

<div class="velo-compare">
  <div class="velo-compare-card bad">
    <strong>PM2-first deploys</strong>
    <p>Great for quick starts, weaker as an operating model. You still need to solve Linux users, HTTPS, service boot, logs, restarts, resource boundaries, and host-level security.</p>
  </div>
  <div class="velo-compare-card good">
    <strong>Velo Deploy</strong>
    <p>Uses the operating system as the supervisor: systemd for lifecycle, Caddy for HTTPS, Linux users for isolation, and a Go CLI to wire the pieces together.</p>
  </div>
</div>

## Docker is powerful. It is not always necessary.

Docker shines when you need image portability, multi-service development environments, or orchestration. But if your target is one VPS and your workload is a Node.js app or a static site, containers can become accidental complexity.

Velo chooses the simpler building blocks first:

- `systemd` for process supervision and boot recovery.
- Linux users and permissions for isolation.
- Caddy for reverse proxying and TLS.
- Git for delivery.
- Node 24 LTS by default, with explicit `engines.node` overrides when your app needs another supported major version.

## Start where the value is

- [Install Velo Deploy](getting-started/installation/) — prepare a VPS once.
- [Quick deploy in 30 seconds](getting-started/quick-deploy/) — ship a repo now.
- [Security model](architecture/security/) — see how isolation works.
- [Node.js versions](guide/node-versions/) — use the default or pin your app.
