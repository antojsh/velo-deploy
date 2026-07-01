---
title: Installation
description: System requirements, supported platforms, and the one-liner installer.
---

import { Tabs, TabItem, Steps, Aside } from '@astrojs/starlight/components';

Velo Deploy runs on a Linux VPS and orchestrates three system services: the `velo-deploy` CLI/TUI binary, the `velo-deploy-watcher` webhook daemon, and the Caddy web server.

## Requirements

| Requirement | Minimum | Recommended |
| --- | --- | --- |
| OS | Ubuntu 20.04, Debian 11 | Ubuntu 22.04 LTS, Debian 12 |
| Architecture | x86_64 (amd64) | x86_64 |
| Privileges | `root` for installation | `root` |
| Free RAM | 512 MB | 1 GB + headroom for apps |
| Free disk | 5 GB | 20 GB |
| Open ports | 80, 443 | 80, 443, 9999 (webhooks) |

<Aside type="caution" title="ARM and macOS are not supported">
	The installer targets `linux/amd64`. The Go source compiles on other platforms, but the systemd and Caddy integration has only been validated on Ubuntu and Debian.
</Aside>

## One-liner install

<Steps>

1. SSH into your VPS as `root`.

	```bash
	ssh root@your-server
	```

2. Run the installer. It downloads the latest release, validates checksums, and registers the `velo-deploy-watcher` systemd unit.

	```bash
	curl -sS https://get.velo-deploy.sh | bash
	```

3. Verify the install.

	```bash
	velo-deploy version
	systemctl status velo-deploy-watcher
	```

</Steps>

## Manual install

Use the manual flow when you want to pin a specific version or run from source.

<Tabs syncKey="install-method">
	<TabItem label="From a release tarball" icon="seti:default">

		<!-- markdownlint-disable MD046 -->
		```bash
		VER=$(curl -sS https://api.github.com/repos/antojsh/velo-deploy/releases/latest | grep tag_name | cut -d '"' -f 4)
		curl -sSLo velo-deploy.tar.gz "https://github.com/antojsh/velo-deploy/releases/download/${VER}/velo-deploy_${VER#v}_linux_amd64.tar.gz"
		tar -xzf velo-deploy.tar.gz
		sudo install -m 0755 velo-deploy /usr/local/bin/velo-deploy
		sudo ./install.sh
		```
		<!-- markdownlint-enable MD046 -->

	</TabItem>
	<TabItem label="From source" icon="seti:shell">

		Requires Go 1.22 or later.

		```bash
		git clone https://github.com/antojsh/velo-deploy.git
		cd velo-deploy
		go build -o velo-deploy ./cmd/velo-deploy
		sudo install -m 0755 velo-deploy /usr/local/bin/velo-deploy
		sudo ./install.sh
		```

	</TabItem>
</Tabs>

## What the installer does

1. Installs Caddy and enables it as a systemd service.
2. Installs `nvm` at `/opt/nvm` for Node.js version management.
3. Copies `velo-deploy` to `/usr/local/bin/velo-deploy`.
4. Creates the configuration directory at `/etc/velo-deploy/`.
5. Creates the apps directory at `/opt/deploy/apps/`.
6. Creates the logs directory at `/var/log/velo-deploy/`.
7. Installs and starts the `velo-deploy-watcher` systemd service on port `9999`.

## Post-installation checks

```bash
# Binary on PATH
which velo-deploy

# Caddy is running
systemctl status caddy

# Webhook watcher is running
systemctl status velo-deploy-watcher

# Default config exists
cat /etc/velo-deploy/config.json
```

If all four checks pass, you are ready to deploy your first app.

## Next steps

- [Quick deploy →](/velo-deploy/getting-started/quick-deploy/)
- [First app walkthrough →](/velo-deploy/getting-started/first-app/)
