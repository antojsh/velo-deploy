---
title: Velo Deploy
description: Deploy Node.js apps and static sites to any VPS with systemd isolation, automatic HTTPS, and no Docker daemon.
template: splash
hero:
  title: Ship Node.js like a serious Linux service
  tagline: >-
    Turn a plain VPS into a small, production-minded platform: one Go binary,
    Caddy HTTPS, systemd units, Linux-user isolation, webhook deploys, and no
    Docker daemon in the critical path.
  actions:
    - text: Deploy in 30 seconds
      link: getting-started/quick-deploy/
      icon: right-arrow
      variant: primary
    - text: Understand the architecture
      link: architecture/overview/
      icon: external
      variant: minimal
---

<div class="velo-pill-row">
  <span class="velo-pill">No Docker daemon</span>
  <span class="velo-pill">systemd-native deploys</span>
  <span class="velo-pill">Caddy HTTPS</span>
  <span class="velo-pill">Linux-user isolation</span>
  <span class="velo-pill">Push-to-deploy</span>
</div>

<div class="velo-metric-row">
  <span class="velo-metric"><strong>1</strong> Go binary</span>
  <span class="velo-metric"><strong>0</strong> runtime control plane</span>
  <span class="velo-metric"><strong>24</strong> Node.js default LTS</span>
</div>

<span class="velo-section-kicker">Production foundation</span>

## The VPS deploy layer for people who understand Linux

You do not need a container ship to move a bicycle. For many Node.js APIs, Astro builds, dashboards, admin panels, and static sites, the cleanest deploy target is still Linux: a process supervised by `systemd`, behind Caddy, running as its own user.

Velo Deploy packages that boring, powerful foundation into a workflow developers can actually operate.

<div class="velo-command-panel">
  <div class="velo-command-header">
    <span class="velo-dot"></span>
    <span class="velo-dot"></span>
    <span class="velo-dot"></span>
    <span>fresh-vps → live HTTPS app</span>
  </div>

```bash
curl -sS https://github.com/antojsh/velo-deploy/releases/latest/download/velo-deploy-install.sh | bash
velo-deploy deploy https://github.com/your/node-or-static-site
git push origin main
```

</div>

<span class="velo-section-kicker">Choose your path</span>

## Start with the right mental model

<div class="velo-doc-grid">
  <a class="velo-doc-card" href="getting-started/installation/">
    <span>01 · Prepare</span>
    <strong>Install once on a VPS</strong>
    <p>Check requirements, install the binary, and verify Caddy plus the webhook watcher.</p>
  </a>
  <a class="velo-doc-card" href="getting-started/quick-deploy/">
    <span>02 · Ship</span>
    <strong>Deploy a repo quickly</strong>
    <p>Go from a Git URL to a running Node.js or static app with HTTPS.</p>
  </a>
  <a class="velo-doc-card" href="architecture/security/">
    <span>03 · Trust</span>
    <strong>Understand isolation</strong>
    <p>See how Linux users, permissions, and hardened systemd units contain each app.</p>
  </a>
</div>

<span class="velo-section-kicker">Why Velo exists</span>

## A smaller platform, built from boring pieces

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
    <span class="velo-icon run">04</span>
    <strong>Node.js and static sites</strong>
    <p>Deploy long-running Node.js services or static output directories like <code>dist</code>, <code>build</code>, and <code>public</code>.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon run">05</span>
    <strong>Push-to-deploy</strong>
    <p>The webhook watcher can redeploy on every push. Simple Git workflow, no external platform required.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon run">06</span>
    <strong>One small Go binary</strong>
    <p>No runtime control plane. Velo adds less operational weight than the average Node.js process.</p>
  </div>
</div>

<span class="velo-section-kicker">Architectural boundaries</span>

## PM2 is not a deployment architecture

<div class="velo-quote">
PM2 is useful, but putting a Node.js process manager in charge of other Node.js processes is like hiring a thief to guard another thief. Sometimes it works. It is not the foundation you want for a VPS platform.
</div>

<div class="velo-compare">
  <div class="velo-compare-card bad">
    <span class="velo-icon warn">!</span>
    <strong>PM2-first deploys</strong>
    <p>Great for quick starts, weaker as an operating model. You still need to solve Linux users, HTTPS, service boot, logs, restarts, resource boundaries, and host-level security.</p>
  </div>
  <div class="velo-compare-card good">
    <span class="velo-icon run">✓</span>
    <strong>Velo Deploy</strong>
    <p>Uses the operating system as the supervisor: systemd for lifecycle, Caddy for HTTPS, Linux users for isolation, and a Go CLI to wire the pieces together.</p>
  </div>
</div>

## Docker is powerful. It is not always necessary.

Docker shines when you need image portability, multi-service development environments, or orchestration. But if your target is one VPS and your workload is a Node.js app or a static site, containers can become accidental complexity.

<div class="velo-principles">
  <div class="velo-principle">
    <strong>Use the OS as the platform</strong>
    <p><code>systemd</code> supervises processes, recovers at boot, and exposes logs through standard Linux tooling.</p>
  </div>
  <div class="velo-principle">
    <strong>Keep delivery Git-shaped</strong>
    <p>Deploy from a repository and optionally let webhooks redeploy when <code>main</code> changes.</p>
  </div>
  <div class="velo-principle">
    <strong>Make HTTPS boring</strong>
    <p>Caddy owns reverse proxying and certificate renewal, so apps stay focused on application code.</p>
  </div>
  <div class="velo-principle">
    <strong>Prefer explicit versions</strong>
    <p>Node 24 LTS is the default; <code>engines.node</code> lets each app pin another supported major.</p>
  </div>
</div>

<div class="velo-footer-cta">

## Start where the value is

<div class="velo-roadmap">
  <a href="getting-started/installation/"><span class="velo-step-number">1</span><span><b>Install Velo Deploy</b><br><small>Prepare the VPS once.</small></span></a>
  <a href="getting-started/quick-deploy/"><span class="velo-step-number">2</span><span><b>Quick deploy in 30 seconds</b><br><small>Ship a repo now.</small></span></a>
  <a href="architecture/security/"><span class="velo-step-number">3</span><span><b>Security model</b><br><small>See how isolation works.</small></span></a>
  <a href="guide/node-versions/"><span class="velo-step-number">4</span><span><b>Node.js versions</b><br><small>Use the default or pin your app.</small></span></a>
</div>

</div>
