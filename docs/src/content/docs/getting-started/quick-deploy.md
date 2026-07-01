---
title: Quick deploy
description: From a Git URL to a running app in 30 seconds.
---

import { Steps, Aside } from '@astrojs/starlight/components';

The fastest way to see Velo Deploy in action is to deploy a public Git repository.

<Steps>

1. Make sure your VPS is set up. See [Installation](/velo-deploy/getting-started/installation/) if you have not done that yet.

2. Deploy a sample repository.

	```bash
	velo-deploy deploy https://github.com/antojsh/velo-deploy-demo
	```

	The CLI will:

	- Clone the repo into `/opt/deploy/apps/velo-deploy-demo`.
	- Detect the type: Node.js or static.
	- Install dependencies and run the build script.
	- Register a systemd unit and a Caddy vhost.
	- Assign a local alias you can hit from your browser.

3. Open the alias in a browser.

	```bash
	velo-deploy list
	```

	The output shows the alias (something like `velo-deploy-demo.local`) and the port (something like `3000`). Add the alias to your laptop's `/etc/hosts` pointing at the server IP, then visit `http://velo-deploy-demo.local` from your browser.

</Steps>

<Aside type="tip" title="Want a public URL?">
	Pass `--domain` to map a real domain to the app. Velo will request a Let's Encrypt certificate through Caddy the first time the domain is hit.
	```bash
	velo-deploy deploy https://github.com/your/repo --domain demo.example.com
	```
</Aside>

## Useful follow-ups

```bash
velo-deploy list             # see all deployed apps
velo-deploy logs velo-deploy-demo   # last 200 lines
velo-deploy logs velo-deploy-demo -f   # follow live
velo-deploy restart velo-deploy-demo
velo-deploy remove velo-deploy-demo
```

## Next steps

- [First app →](/velo-deploy/getting-started/first-app/) — A full walkthrough with a custom Node.js project.
- [Custom domains →](/velo-deploy/guide/domains/) — Map a real domain to your app.
