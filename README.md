# Velo Deploy

**Bare Metal PaaS** — Deploy Node.js applications and static sites to any VPS without Docker.

Velo Deploy uses **systemd** for process management and **Caddy** for automatic HTTPS, providing a lightweight, secure alternative to container-based deployments.

```
curl -sS https://get.velo-deploy.sh | bash
```

## Features

- **One-liner Installation** — Get running in seconds
- **Automatic HTTPS** — Let's Encrypt certificates via Caddy
- **Node.js + Static Sites** — Deploy APIs and frontend apps
- **Security Sandboxing** — systemd isolation per app
- **Auto-Deploy** — GitHub webhooks for continuous deployment
- **TUI Dashboard** — Terminal-based management interface
- **CLI Tools** — Full command-line control
- **Minimal Resources** — No Docker, minimal RAM/CPU footprint

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Velo Deploy                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   GitHub ──webhook──▶ deploy-watcher (port 9999)           │
│                              │                              │
│                              ▼                              │
│                    ┌─────────────────┐                      │
│                    │  git pull       │                      │
│                    │  npm install    │                      │
│                    │  (build)        │                      │
│                    └────────┬────────┘                      │
│                             │                              │
│              ┌──────────────┴──────────────┐                │
│              ▼                              ▼              │
│   ┌─────────────────────┐      ┌─────────────────────┐      │
│   │   Node.js Apps      │      │   Static Sites      │      │
│   │   (systemd)        │      │   (Caddy)          │      │
│   │   :3000-3999       │      │   file_server      │      │
│   └─────────┬───────────┘      └──────────┬─────────┘      │
│             │                             │                │
│             └──────────┬──────────────────┘                │
│                        ▼                                   │
│               ┌─────────────────┐                          │
│               │     Caddy       │                          │
│               │   :80  :443     │                          │
│               │  HTTPS + Proxy  │                          │
│               └─────────────────┘                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Requirements

- **OS:** Linux (Ubuntu 20.04+, Debian 11+, or similar)
- **Architecture:** x86_64 (amd64)
- **Privileges:** Root access for installation
- **Ports:** 80, 443, and optionally 9999 (webhooks)

## Installation

### Quick Install

```bash
curl -sS https://get.velo-deploy.sh | bash
```

### Manual Install

```bash
# Clone the repository
git clone https://github.com/antojsh/velo-deploy.git
cd velo-deploy

# Build the binary
go build -o velo-deploy ./cmd/deploy

# Install (requires root)
sudo ./install.sh
```

### Post-Installation

After installation, start the TUI:

```bash
velo-deploy
```

## Usage

### TUI Dashboard

Launch the terminal-based dashboard for visual management:

```bash
velo-deploy
```

**Keyboard Shortcuts:**

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate app list |
| `Tab` | Switch panel |
| `Enter` | Select app |
| `N` | New deploy (from git) |
| `A` | Add existing app |
| `R` | Restart app (Node.js only) |
| `S` | Stop app (Node.js only) |
| `L` | View logs |
| `D` | Delete app |
| `Q` | Quit |

### CLI Commands

#### Deploy from Git

```bash
# Basic deployment
velo-deploy deploy https://github.com/user/my-api

# With custom domain
velo-deploy deploy https://github.com/user/my-api --domain api.example.com

# With custom alias
velo-deploy deploy https://github.com/user/my-api --alias myapi.local
```

#### Add Existing App

Register an already-cloned application:

```bash
# Auto-detect type (Node.js or static)
velo-deploy add my-site /opt/deploy/apps/my-site

# Explicitly set type
velo-deploy add my-site /opt/deploy/apps/my-site --type static
velo-deploy add my-api /opt/deploy/apps/my-api --type node

# With domain
velo-deploy add my-site /opt/deploy/apps/my-site --domain mysite.com
```

#### List Apps

```bash
velo-deploy list
```

#### Restart / Stop (Node.js apps)

```bash
velo-deploy restart my-api
velo-deploy stop my-api
```

#### View Logs

```bash
# View last 200 lines
velo-deploy logs my-api

# Follow log output
velo-deploy logs my-api -f
```

#### Remove App

```bash
velo-deploy remove my-api
```

### Auto-Deploy via Webhooks

1. **Start the webhook daemon:**

   ```bash
   # On a specific port
   velo-deploy daemon --port 9999

   # Or via systemd (installed automatically)
   sudo systemctl enable --now velo-watcher
   ```

2. **Configure GitHub webhook:**

   - Go to your repository → Settings → Webhooks → Add webhook
   - Payload URL: `http://your-server:9999/webhook`
   - Content type: `application/json`
   - Events: Push events (select Just the push event)

3. **How it works:**

   - Only pushes to `main` or `master` trigger deployments
   - Velo automatically pulls, rebuilds, and restarts the app

## Project Structure

```
/etc/velo-deploy/
├── config.json          # App metadata and configuration
├── apps/                # Deployed application source code
│   └── <app-name>/
├── logs/                # Deploy logs
│   └── <app-name>.log
/var/log/velo-deploy/    # System logs

/etc/systemd/system/
├── velo-watcher.service  # Webhook watcher daemon
├── velo-<app>.service   # Per-app systemd units

/etc/caddy/conf.d/
├── <app>.conf           # Per-app Caddy configs (domain-based)
├── _shared.conf         # Shared catch-all config (path-based)
```

## Configuration

Configuration is stored at `/etc/velo-deploy/config.json`:

```json
{
  "caddy_conf_dir": "/etc/caddy/conf.d",
  "apps_dir": "/opt/deploy/apps",
  "logs_dir": "/var/log/velo-deploy",
  "daemon_port": "9999",
  "nvm_dir": "/opt/nvm",
  "apps": {
    "my-api": {
      "name": "my-api",
      "type": "node",
      "repo_url": "https://github.com/user/my-api",
      "branch": "main",
      "node_version": "20",
      "port": 3000,
      "domain": "",
      "alias": "my-api.local",
      "node_path": "/opt/nvm/versions/node/v20.0.0/bin/node",
      "entry_point": "index.js"
    },
    "my-site": {
      "name": "my-site",
      "type": "static",
      "domain": "",
      "alias": "my-site.local",
      "output_dir": "dist"
    }
  }
}
```

## Security

Each application is isolated using Linux security features:

| Setting | Purpose |
|---------|---------|
| `ProtectSystem=full` | /usr, /boot, /etc are read-only |
| `ProtectHome=true` | Cannot access other users' home directories |
| `PrivateTmp=true` | Isolated /tmp filesystem |
| `NoNewPrivileges=true` | Prevents privilege escalation |
| `ReadWritePaths=` | Limited to app directory only |

**Per-App Users:** Each app runs under its own Linux user (`velo-<appname>`) with no login shell and no password.

**Filesystem Isolation:** Apps can only write to their own directory and the system tmp.

## Node.js Version Management

Velo Deploy uses nvm to manage Node.js versions:

- Versions are installed to `/opt/nvm/versions/node/`
- Detected from `package.json` `engines.node` field
- Default version: **Node.js 20**
- Supported: Node.js 16, 18, 20, 22

```json
{
  "name": "my-app",
  "engines": {
    "node": ">=18.0.0"
  }
}
```

## Static Site Support

Velo Deploy automatically detects and serves static sites:

**Detection:**
- Contains `index.html` in root
- Contains `package.json` with build script

**Build Process:**
- Runs `npm run build` automatically
- Serves from output directory (auto-detected or specified)

**Supported Output Directories** (in order of priority):
1. `dist`
2. `build`
3. `_site`
4. `public`
5. `output`

**Example Caddy config for static sites:**

```caddy
mysite.local {
    root * /opt/deploy/apps/mysite/dist
    file_server
}
```

## Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.22+ | Single static binary |
| TUI | Bubble Tea | Terminal UI framework |
| Styling | Lipgloss | TUI styling |
| Process Mgmt | systemd | App lifecycle, isolation |
| Web Server | Caddy | Reverse proxy, automatic HTTPS |
| Node.js | nvm | Version management |

## Troubleshooting

### App not starting

```bash
# Check systemd status
sudo systemctl status velo-myapp

# View logs
sudo journalctl -u velo-myapp -f

# Check Caddy config
sudo caddy validate --config /etc/caddy/conf.d/myapp.conf
```

### Webhook not working

```bash
# Verify daemon is running
sudo systemctl status velo-watcher

# Test webhook endpoint
curl -X POST http://localhost:9999/webhook
```

### Port conflicts

```bash
# Check which apps are using ports
ss -tlnp | grep -E '3000|4000'

# Remove stale configs
sudo rm /etc/caddy/conf.d/myapp.conf
sudo caddy reload
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/antojsh/velo-deploy.git
cd velo-deploy

# Install Go (1.22+)
# Ubuntu/Debian:
sudo apt install golang-go

# Build
go build -o velo-deploy ./cmd/deploy

# Run tests
go test ./...

# Run with verbose output
go run ./cmd/deploy
```

### Code Style

- Follow Go standard conventions
- Run `go fmt` before committing
- Ensure `go vet` passes

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style library
- [Caddy](https://caddyserver.com/) - Web server
- [nvm](https://github.com/nvm-sh/nvm) - Node version management
