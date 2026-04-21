#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[deploy]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[deploy]${NC} $1"; }
log_error() { echo -e "${RED}[deploy]${NC} $1"; }

if [ "$(id -u)" -ne 0 ]; then
  log_error "Must be run as root (sudo)."
  exit 1
fi

log_warn "This will remove ALL deployed apps and the deploy tool itself."
read -p "Are you sure? (y/N) " -r
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  log_info "Aborted."
  exit 0
fi

# 1. Stop and remove all deploy services
for svc in /etc/systemd/system/deploy-*.service; do
  if [ -f "$svc" ]; then
    name=$(basename "$svc" .service)
    log_info "Stopping and removing $name..."
    systemctl stop "$name" 2>/dev/null || true
    systemctl disable "$name" 2>/dev/null || true
    rm -f "$svc"
  fi
done
systemctl daemon-reload

# 2. Remove Caddy configs
rm -rf /etc/caddy/conf.d/*.conf
caddy reload 2>/dev/null || true

# 3. Clean /etc/hosts (remove deploy aliases)
if [ -f /etc/hosts ]; then
  sed -i '/\.local$/d' /etc/hosts
fi

# 4. Remove deploy users
for u in $(cut -d: -f1 /etc/passwd | grep '^deploy-'); do
  log_info "Removing user $u..."
  userdel -r "$u" 2>/dev/null || true
done

# 5. Remove directories
rm -rf /opt/deploy
rm -rf /etc/deploy
rm -rf /var/log/deploy

# 6. Remove binary
rm -f /usr/local/bin/deploy

# 7. Remove daemon service
rm -f /etc/systemd/system/deploy-watcher.service
systemctl daemon-reload

log_info "Deploy has been completely removed."
