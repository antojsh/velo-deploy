---
title: Webhooks
description: Auto-deploy on push to main with GitHub webhooks.
---

import { Steps, Aside } from '@astrojs/starlight/components';

The webhook daemon (`velo-deploy-watcher`) listens for push events on port `9999` and triggers a redeploy of the matching app.

## Architecture

```
GitHub ──webhook──▶ velo-deploy-watcher (port 9999)
                            │
                            ▼
                  git pull → npm ci → npm run build
                            │
                            ▼
                  systemctl restart velo-<app>
```

## 1. Start the daemon

If you used the one-liner installer, the daemon is already running under systemd:

```bash
sudo systemctl status velo-deploy-watcher
```

To start it manually (for testing or running on a non-systemd system):

```bash
velo-deploy daemon --port 9999
```

## 2. Configure the GitHub webhook

In your repository, go to **Settings → Webhooks → Add webhook** and fill in:

| Field | Value |
| --- | --- |
| Payload URL | `http://<server-ip>:9999/webhook` |
| Content type | `application/json` |
| Events | Just the push event |
| SSL verification | Enable (recommended). Disable only for self-signed dev environments. |

If you have a domain pointing at the server, use HTTPS:

```
https://webhook.example.com:9999/webhook
```

## 3. Configure the app

Velo matches incoming webhooks to apps by the **repository URL**. Make sure the `repo_url` in `config.json` matches the Git URL GitHub sends in the payload.

```json
{
  "apps": {
    "my-api": {
      "repo_url": "https://github.com/your-username/my-api",
      "branch": "main"
    }
  }
}
```

Only pushes to `main` or `master` trigger a deploy. Other branches are ignored.

## 4. Test it

From your laptop:

```bash
curl -X POST http://<server-ip>:9999/webhook
```

You should see a `200` response and, in `journalctl -u velo-deploy-watcher -f`, a log line about the request.

Push a small change to `main` and watch it land:

```bash
# On the server
journalctl -u velo-deploy-watcher -f
velo-deploy logs my-api -f
```

## Security

<Aside type="caution" title="The webhook endpoint is unauthenticated">
	By design, the daemon accepts any `POST` to `/webhook` that claims to come from a known repo. For production, put the daemon behind a reverse proxy that:
	- Restricts source IPs to GitHub's published ranges: [GitHub webhook IPs](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries#about-validating-webhook-deliveries).
	- Verifies the `X-Hub-Signature-256` HMAC header against your webhook secret.
	- Exposes the daemon over HTTPS, not plain HTTP.
</Aside>

### Validating the HMAC

A future version of `velo-deploy-watcher` will validate the `X-Hub-Signature-256` header out of the box. Until then, the recommended pattern is to put Caddy in front:

```nginx
webhook.example.com {
    @github header X-Hub-Signature-256 *
    handle_response @github {
        request_body {
            replace "REPLACE_ME" "computed_hmac"
        }
    }
    reverse_proxy localhost:9999
}
```

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| `502` from GitHub | Daemon not running or port `9999` blocked. |
| `200` but no redeploy | Branch is not `main`/`master`, or `repo_url` does not match. |
| `Permission denied` in logs | The daemon is not running as `root`. |
| Slow deploys | `npm install` is rebuilding from scratch. Commit `package-lock.json` and use `npm ci`. |
