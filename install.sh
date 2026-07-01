#!/bin/bash
set -e

# ==========================================
# velo-deploy installer
# curl -sS https://get.velo-deploy.sh | bash
# ==========================================
#
# Env vars:
#   VELO_DEPLOY_REPO    Override GitHub repo (default: antojsh/velo-deploy)
#   VELO_DEPLOY_VERSION Override release tag (default: latest)
#                        Examples: v0.2.0, latest

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[velo-deploy]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[velo-deploy]${NC} $1"; }
log_error() { echo -e "${RED}[velo-deploy]${NC} $1"; }

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
apt-get install -y -qq curl git build-essential ufw ca-certificates >/dev/null 2>&1

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
  [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
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

# --- 7. Download velo-deploy binary from GitHub release ---
REPO="${VELO_DEPLOY_REPO:-antojsh/velo-deploy}"
TAG="${VELO_DEPLOY_VERSION:-latest}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *) log_error "Unsupported OS: $OS"; exit 1 ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Walk recent releases and pick the first one that has a velo-deploy binary
# for this OS/arch. release-please can create docs-only releases that don't
# ship a CLI asset; we want the most recent release that actually does.
find_latest_cli_release() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases?per_page=20" 2>/dev/null | \
    python3 -c "
import json, sys
target = 'velo-deploy_${OS}_${ARCH}.tar.gz'
for r in json.load(sys.stdin):
    for a in r.get('assets', []):
        if a['name'].endswith(target):
            print(r['tag_name'])
            sys.exit(0)
" 2>/dev/null
}

if [ "$TAG" = "latest" ]; then
  log_info "Resolving latest CLI release for ${OS}/${ARCH}..."
  TAG=$(find_latest_cli_release)
  if [ -z "$TAG" ]; then
    log_error "Could not find a release with a velo-deploy binary for ${OS}/${ARCH}."
    log_error "Pin a specific version: VELO_DEPLOY_VERSION=v0.1.0 bash install.sh"
    exit 1
  fi
fi

# release-please tags are v-prefixed (e.g., v0.1.0); asset names use the bare
# version (e.g., velo-deploy_0.1.0_linux_amd64.tar.gz).
ASSET_VERSION="${TAG#v}"
ASSET="velo-deploy_${ASSET_VERSION}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

if [ -f "./velo-deploy" ]; then
  log_info "Using local ./velo-deploy binary (dev install)..."
  install -m 0755 ./velo-deploy /usr/local/bin/velo-deploy
else
  log_info "Downloading ${ASSET} from ${REPO}@${TAG}..."
  if ! curl -fsSLo "$TMP_DIR/$ASSET" "$BASE_URL/$ASSET"; then
    log_error "Failed to download $BASE_URL/$ASSET"
    log_error "Check that the release exists and your platform is supported."
    exit 1
  fi

  log_info "Downloading SHA256SUMS..."
  if ! curl -fsSLo "$TMP_DIR/SHA256SUMS" "$BASE_URL/SHA256SUMS"; then
    log_error "Failed to download SHA256SUMS from $BASE_URL"
    exit 1
  fi

  log_info "Verifying checksum..."
  if ! (cd "$TMP_DIR" && sha256sum -c --ignore-missing SHA256SUMS); then
    log_error "Checksum verification failed. Aborting install."
    exit 1
  fi

  log_info "Extracting binary..."
  tar -xzf "$TMP_DIR/$ASSET" -C "$TMP_DIR"
  if [ ! -f "$TMP_DIR/velo-deploy" ]; then
    log_error "Archive did not contain a 'velo-deploy' binary."
    exit 1
  fi
  install -m 0755 "$TMP_DIR/velo-deploy" /usr/local/bin/velo-deploy
fi

log_info "velo-deploy ${TAG} installed at /usr/local/bin/velo-deploy"

# --- 8. Create systemd service for git-watcher ---
cat > /etc/systemd/system/velo-deploy-watcher.service << 'EOF'
[Unit]
Description=Velo Deploy Git-Watcher Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/velo-deploy daemon --port 9999
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
log_info "velo-deploy-watcher.service installed. Enable with: systemctl enable --now velo-deploy-watcher"

# --- Done ---
echo ""
log_info "==========================================="
log_info "  velo-deploy ${TAG} installed successfully!"
log_info ""
log_info "  Usage:"
log_info "    velo-deploy              # Launch TUI"
log_info "    velo-deploy deploy <repo>  # Deploy a repo"
log_info "    velo-deploy list          # List all apps"
log_info ""
log_info "  Start the auto-deploy daemon:"
log_info "    systemctl enable --now velo-deploy-watcher"
log_info "==========================================="
