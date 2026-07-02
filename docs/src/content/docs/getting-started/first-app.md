---
title: First app
description: Build and deploy a real Node.js API from scratch.
---

import { Steps, Aside } from '@astrojs/starlight/components';

This walkthrough deploys a minimal Node.js API to your VPS and exposes it under a local alias.

## 1. Create the project locally

<Steps>

1. Create a new folder and initialize a Node project.

	```bash
	mkny hello-velo && cd hello-velo
	npm init -y
	```

2. Add a minimal HTTP server. Create `index.js`:

	```js
	import { createServer } from 'node:http';

	const port = process.env.PORT || 3000;

	createServer((req, res) => {
	  res.writeHead(200, { 'content-type': 'application/json' });
	  res.end(JSON.stringify({ hello: 'velo', path: req.url }));
	}).listen(port, () => {
	  console.log(`listening on ${port}`);
	});
	```

3. Pin the Node version with `engines` and add a start script.

	```json
	{
	  "name": "hello-velo",
	  "type": "module",
	  "engines": { "node": ">=20" },
	  "scripts": { "start": "node index.js" }
	}
	```

4. Push the project to a Git repository (GitHub, Gitea, self-hosted — anything Velo can `git clone`).

	```bash
	git init
	git add .
	git commit -m "feat: minimal hello world"
	gh repo create hello-velo --public --source=. --push
	```

</Steps>

## 2. Deploy from the VPS

SSH into your VPS and run:

```bash
velo-deploy deploy https://github.com/your-username/hello-velo
```

Velo clones, installs (no dependencies in this case, but it will run `npm install` if you have any), picks Node 24 by default unless `engines.node` defines another supported version, registers a systemd unit, and writes a Caddy vhost.

## 3. Verify

```bash
velo-deploy list
```

You should see `hello-velo` with a local alias and an assigned port. Add the alias to your laptop's `/etc/hosts` and curl it:

```bash
curl http://hello-velo.local
# {"hello":"velo","path":"/"}
```

## 4. Add a domain (optional)

When you are ready to share the API:

```bash
velo-deploy add hello-velo /opt/deploy/apps/hello-velo --domain api.example.com
```

Caddy will request a Let's Encrypt certificate the first time the domain resolves to your server. See [Custom domains](/velo-deploy/guide/domains/) for the DNS steps.

<Aside type="tip" title="Watch the logs while you iterate">
	Open a second SSH session and tail the logs:
	```bash
	velo-deploy logs hello-velo -f
	```
	Every restart from `velo-deploy restart` or a webhook deploy shows up here.
</Aside>

## 5. Enable auto-deploy

On the server:

```bash
sudo systemctl enable --now velo-deploy-watcher
```

In your GitHub repo, add a webhook with:

- **Payload URL**: `http://your-server:9999/webhook`
- **Content type**: `application/json`
- **Events**: Just the push event

Push to `main` and Velo will pull, rebuild, and restart automatically. See [Webhooks](/velo-deploy/guide/webhooks/) for the full configuration.

## Next steps

- [CLI reference →](/velo-deploy/reference/cli-reference/)
- [Architecture →](/velo-deploy/architecture/overview/)
