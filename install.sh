#!/bin/bash
set -e

# ==========================================
# deploy.sh - One-liner installer
# curl -sS https://get.deploy.sh | bash
# ==========================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[deploy]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[deploy]${NC} $1"; }
log_error() { echo -e "${RED}[deploy]${NC} $1"; }

# --- 0. Require root ---
if [ "$(id -u)" -ne 0 ]; then
  log_error "This script must be run as root (sudo)."
  exit 1
fi

# --- 1. Detect distro ---
if [ -f /etc/os-release ]; then
  . /etc/os-release
  DISTRO="${ID,,}"
else
  log_error "Cannot detect OS. /etc/os-release not found."
  exit 1
fi

if [[ "$DISTRO" != "ubuntu" && "$DISTRO" != "debian" ]]; then
  log_warn "Unsupported distro: $DISTRO. This installer only supports Ubuntu/Debian."
  log_warn "Exiting."
  exit 1
fi

log_info "Detected: $PRETTY_NAME"

# --- 2. Install dependencies ---
log_info "Installing dependencies..."
apt-get update -qq
apt-get install -y -qq curl git build-essential ufw >/dev/null 2>&1

# --- 3. Install Caddy ---
if ! command -v caddy &>/dev/null; then
  log_info "Installing Caddy..."
  install -d -m 0755 /etc/apt/keyrings
  curl -fsSL https://dl.cloudsmith.io/public/caddy/stable/gpg.key \
    | gpg --dearmor -o /etc/apt/keyrings/caddy-stable-archive-keyring.gpg
  echo "deb [signed-by=/etc/apt/keyrings/caddy-stable-archive-keyring.gpg] https://dl.cloudsmith.io/public/caddy/stable/deb/debian any-version main" \
    > /etc/apt/sources.list.d/caddy-stable.list
  apt-get update -qq
  apt-get install -y -qq caddy >/dev/null 2>&1
  log_info "Caddy installed."
else
  log_info "Caddy already installed."
fi

# --- 4. Install nvm to /opt/nvm (accessible by all system users) ---
export NVM_DIR="/opt/nvm"
if [ ! -s "$NVM_DIR/nvm.sh" ]; then
  log_info "Installing nvm to /opt/nvm..."
  mkdir -p "$NVM_DIR"
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | NVM_DIR="$NVM_DIR" bash >/dev/null 2>&1
  # Make nvm available in current shell
  [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
  # Make node accessible system-wide
  chmod -R a+rX "$NVM_DIR"
  log_info "nvm installed at /opt/nvm."
else
  log_info "nvm already installed at /opt/nvm."
  [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
fi

# --- 5. Prepare directories ---
log_info "Creating directory structure..."
mkdir -p /etc/deploy
mkdir -p /opt/deploy/apps
mkdir -p /var/log/deploy
mkdir -p /etc/caddy/conf.d

# --- 6. Ensure caddy.conf.d is imported ---
CADDYFILE="/etc/caddy/Caddyfile"
if ! grep -q "import /etc/caddy/conf.d/*.conf" "$CADDYFILE" 2>/dev/null; then
  echo "" >> "$CADDYFILE"
  echo "import /etc/caddy/conf.d/*.conf" >> "$CADDYFILE"
  log_info "Added import directive to Caddyfile."
fi

# --- 7. Download deploy binary ---
log_info "Downloading deploy binary..."
# TODO: Replace with real download URL
# curl -sSLo /usr/local/bin/deploy https://github.com/youruser/deploy/releases/latest/download/deploy-linux-amd64
# chmod +x /usr/local/bin/deploy

# For now, just place the binary if built locally
if [ -f "./deploy" ]; then
  cp ./deploy /usr/local/bin/deploy
  chmod +x /usr/local/bin/deploy
  log_info "Deploy binary installed at /usr/local/bin/deploy"
else
  log_warn "Deploy binary not found. Build it first with: go build -o deploy ./cmd/deploy"
fi

# --- 8. Create systemd service for git-watcher ---
cat > /etc/systemd/system/deploy-watcher.service << 'EOF'
[Unit]
Description=Deploy Git-Watcher Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/deploy daemon --port 9999
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
log_info "deploy-watcher.service installed. Enable with: systemctl enable --now deploy-watcher"

# --- Done ---
echo ""
log_info "==========================================="
log_info "  deploy installed successfully!"
log_info ""
log_info "  Usage:"
log_info "    deploy              # Launch TUI"
log_info "    deploy deploy <repo>  # Deploy a repo"
log_info "    deploy list          # List all apps"
log_info ""
log_info "  Start the auto-deploy daemon:"
log_info "    systemctl enable --now deploy-watcher"
log_info "==========================================="
