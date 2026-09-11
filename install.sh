#!/bin/bash
# shellcheck disable=SC1091  # sourced files (/etc/os-release) are not in shellcheck's source path in CI
set -e

# ==========================================
# velo-deploy installer
# curl -fsSL https://github.com/antojsh/velo-deploy/releases/latest/download/velo-deploy-install.sh | sudo bash
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
  # shellcheck source=/etc/os-release
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
apt-get install -y -qq curl git build-essential ca-certificates openssl python3 >/dev/null 2>&1

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

# --- 4. Node.js runtime dir (official tarballs under /opt/deploy/node/<major>) ---
# nvm at /opt/nvm is an optional legacy fallback; Velo does not install it.

# --- 5. Prepare directories ---
log_info "Creating directory structure..."
mkdir -p /etc/velo-deploy/apps
mkdir -p /etc/velo-deploy/caddy
mkdir -p /opt/deploy/apps
mkdir -p /opt/deploy/node
mkdir -p /var/log/velo-deploy
mkdir -p /etc/caddy/conf.d

if [ -f /etc/deploy/config.json ] && [ ! -f /etc/velo-deploy/config.json ]; then
  log_info "Migrating /etc/deploy/config.json to /etc/velo-deploy/config.json"
  cp /etc/deploy/config.json /etc/velo-deploy/config.json
fi

if [ ! -s /etc/velo-deploy/webhook.secret ]; then
  openssl rand -hex 32 > /etc/velo-deploy/webhook.secret
  chmod 600 /etc/velo-deploy/webhook.secret
  log_info "Wrote GitHub webhook secret to /etc/velo-deploy/webhook.secret"
fi

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
    log_error "The latest releases do not include a Linux binary for ${OS}/${ARCH}."
    log_error "Check the release assets or pin a known-good version with: VELO_DEPLOY_VERSION=vX.Y.Z sudo bash install.sh"
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

  if curl -fsSLo "$TMP_DIR/SHA256SUMS.bundle" "$BASE_URL/SHA256SUMS.bundle"; then
    if command -v cosign >/dev/null 2>&1; then
      log_info "Verifying cosign signature..."
      if ! cosign verify-blob \
        --bundle "$TMP_DIR/SHA256SUMS.bundle" \
        --certificate-identity-regexp 'https://github.com/.+/velo-deploy/.github/workflows/release-please.yml@.*' \
        --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
        "$TMP_DIR/SHA256SUMS"; then
        log_error "cosign verification failed. Aborting install."
        exit 1
      fi
    else
      log_warn "cosign is not installed; skipped signature check (SHA256 still verified)."
    fi
  else
    log_warn "No SHA256SUMS.bundle on this release; skipped cosign verification."
  fi

  log_info "Extracting binary..."
  tar -xzf "$TMP_DIR/$ASSET" -C "$TMP_DIR"
  if [ ! -f "$TMP_DIR/velo-deploy" ]; then
    log_error "Archive did not contain a 'velo-deploy' binary."
    exit 1
  fi
  if [ -f /usr/local/bin/velo-deploy ]; then
    cp /usr/local/bin/velo-deploy /usr/local/bin/velo-deploy.bak
    log_info "Backed up existing binary to /usr/local/bin/velo-deploy.bak"
  fi
  install -m 0755 "$TMP_DIR/velo-deploy" /usr/local/bin/velo-deploy
fi

log_info "velo-deploy ${TAG} installed at /usr/local/bin/velo-deploy"

# --- 8. Create systemd service for git-watcher ---
cat > /etc/systemd/system/velo-deploy-watcher.service << 'EOF'
[Unit]
Description=Velo Deploy Git-Watcher Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/velo-deploy daemon --port 9999
Restart=always
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ReadWritePaths=/etc/velo-deploy /opt/deploy /var/log/velo-deploy /etc/caddy /etc/hosts /etc/systemd/system /tmp

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
log_info "  Webhook secret: /etc/velo-deploy/webhook.secret"
log_info "  Use that value as the GitHub webhook secret."
log_info ""
log_info "  Start the auto-deploy daemon:"
log_info "    systemctl enable --now velo-deploy-watcher"
log_info "==========================================="
