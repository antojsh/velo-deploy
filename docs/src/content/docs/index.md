---
title: Velo Deploy
description: Bare Metal PaaS — Deploy Node.js applications and static sites to any VPS without Docker.
template: splash
hero:
  tagline: A lightweight, secure PaaS that uses systemd for process isolation and Caddy for automatic HTTPS. No containers, no Kubernetes, no overhead.
  actions:
    - text: Get started
      link: getting-started/installation/
      icon: right-arrow
      variant: primary
    - text: View on GitHub
      link: https://github.com/antojsh/velo-deploy
      icon: external
      variant: minimal
  image:
    file: ../../assets/logo-dark.svg
    alt: Velo Deploy logo
---

import { Card, CardGrid, LinkCard } from '@astrojs/starlight/components';

<CardGrid stagger>
	<Card title="One-liner install" icon="rocket">
		From zero to running in seconds. A single curl command installs the daemon, TUI, CLI, and all required services.

		```bash
		curl -sS https://get.velo-deploy.sh | bash
		```
	</Card>
	<Card title="Automatic HTTPS" icon="seti:lock">
		Caddy issues and renews Let's Encrypt certificates for every domain. No manual cert management, no downtime.
	</Card>
	<Card title="Node.js and static sites" icon="seti:javascript">
		Deploy long-running Node services and pre-built static assets from the same CLI. Auto-detects build output directories.
	</Card>
	<Card title="Systemd isolation" icon="seti:default">
		Each app runs as its own Linux user with a hardened systemd unit: read-only system, isolated tmp, no privilege escalation.
	</Card>
	<Card title="GitHub webhooks" icon="github">
		Push to `main` and Velo pulls, builds, and restarts. No CI pipeline required.
	</Card>
	<Card title="TUI dashboard" icon="seti:terminal">
		Manage every app from a single keyboard-driven terminal interface. No web UI to maintain.
	</Card>
</CardGrid>

## Why Velo Deploy?

Container-based platforms are powerful, but most teams only need a small subset of their features. Velo Deploy ships just enough to deploy a Node.js app or a static site to a single VPS, with the security guarantees you would expect from a PaaS.

<LinkCard
	title="Read the architecture overview →"
	description="Understand how systemd, Caddy, and the Go binary fit together."
	href="architecture/overview/"
/>

## Quick navigation

<CardGrid>
	<LinkCard title="Install" href="getting-started/installation/" description="System requirements and the one-liner installer." />
	<LinkCard title="Quick deploy" href="getting-started/quick-deploy/" description="Go from a Git URL to a running app in 30 seconds." />
	<LinkCard title="CLI reference" href="reference/cli-reference/" description="Every command, flag, and exit code." />
	<LinkCard title="Troubleshooting" href="operations/troubleshooting/" description="Diagnose the most common production issues." />
</CardGrid>
