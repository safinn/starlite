#!/usr/bin/env bash
set -euo pipefail

DEPLOYER_USER="deployer"
APPS_ROOT="/opt/apps_workspace"
APP_NAME="${1:-starlite}"
APP_USER="$APP_NAME"
APP_HOME="$APPS_ROOT/$APP_NAME"

log() { echo -e "\033[1;34m[INFO]\033[0m $*"; }
ok()  { echo -e "\033[1;32m[OK]\033[0m   $*"; }
warn(){ echo -e "\033[1;33m[WARN]\033[0m $*"; }
err() { echo -e "\033[1;31m[ERR]\033[0m  $*" >&2; }

write_if_changed() {
  local path="$1"
  local content="$2"

  if [[ -f "$path" ]] && diff -q <(echo "$content") "$path" >/dev/null; then
    return 1
  fi

  echo "$content" > "$path"
  return 0
}

[[ $EUID -eq 0 ]] || { err "Run as root"; exit 1; }

# -------------------------------
# System
# -------------------------------
log "Updating system..."
apt-get update -qq
DEBIAN_FRONTEND=noninteractive apt-get upgrade -y -qq

apt-get install -y -qq \
  curl wget ufw fail2ban sudo \
  unattended-upgrades gnupg

# -------------------------------
# Firewall
# -------------------------------
if ! ufw status | grep -q "Status: active"; then
  log "Configuring firewall..."
  ufw default deny incoming
  ufw default allow outgoing
  ufw allow 22/tcp
  ufw allow 80/tcp
  ufw allow 443/tcp
  ufw --force enable >/dev/null 2>&1
  ok "Firewall enabled"
else
  log "Firewall already active"
fi

# -------------------------------
# Users
# -------------------------------
id "$DEPLOYER_USER" &>/dev/null || useradd -m -s /bin/bash "$DEPLOYER_USER"
id "$APP_USER" &>/dev/null || useradd -r -s /usr/sbin/nologin "$APP_USER"

# -------------------------------
# SSH keys (append-only)
# -------------------------------
ROOT_KEYS="/root/.ssh/authorized_keys"
DEPLOYER_SSH="/home/$DEPLOYER_USER/.ssh"
DEPLOYER_KEYS="$DEPLOYER_SSH/authorized_keys"

[[ -s "$ROOT_KEYS" ]] || { err "Root has no SSH keys"; exit 1; }

install -d -m 700 -o "$DEPLOYER_USER" -g "$DEPLOYER_USER" "$DEPLOYER_SSH"
touch "$DEPLOYER_KEYS"
chown "$DEPLOYER_USER:$DEPLOYER_USER" "$DEPLOYER_KEYS"
chmod 600 "$DEPLOYER_KEYS"

while read -r key; do
  grep -qxF "$key" "$DEPLOYER_KEYS" || echo "$key" >> "$DEPLOYER_KEYS"
done < "$ROOT_KEYS"

ok "SSH keys ensured"

# -------------------------------
# SSH config
# -------------------------------
mkdir -p /etc/ssh/sshd_config.d

SSH_CONFIG=$(cat <<EOF
PermitRootLogin prohibit-password
PubkeyAuthentication yes
PasswordAuthentication no
ChallengeResponseAuthentication no
UsePAM yes

AllowUsers $DEPLOYER_USER

MaxAuthTries 3
LoginGraceTime 30

X11Forwarding no
AllowAgentForwarding no
EOF
)

if write_if_changed /etc/ssh/sshd_config.d/00-hardening.conf "$SSH_CONFIG"; then
  log "SSH config updated"

  install -d -m 0755 /run/sshd
  sshd -t
  systemctl restart ssh 2>/dev/null || systemctl restart sshd
  ok "SSH restarted"
else
  log "SSH config unchanged"
fi

# -------------------------------
# Sudo
# -------------------------------
SUDO_CONFIG="$DEPLOYER_USER ALL=(ALL) NOPASSWD: ALL"

if write_if_changed "/etc/sudoers.d/$DEPLOYER_USER" "$SUDO_CONFIG"; then
  chmod 440 "/etc/sudoers.d/$DEPLOYER_USER"
  ok "Sudo configured"
fi

# -------------------------------
# Fail2Ban
# -------------------------------
FAIL2BAN_CONFIG=$(cat <<EOF
[sshd]
enabled = true
maxretry = 3
findtime = 600
bantime = 600
EOF
)

if write_if_changed /etc/fail2ban/jail.d/sshd.local "$FAIL2BAN_CONFIG"; then
  systemctl restart fail2ban
  ok "fail2ban restarted"
else
  log "fail2ban unchanged"
fi

systemctl enable fail2ban >/dev/null

# -------------------------------
# Caddy
# -------------------------------
if ! command -v caddy >/dev/null 2>&1; then
  log "Installing Caddy..."

  install -d -m 0755 /etc/apt/keyrings

  if [[ ! -f /etc/apt/keyrings/caddy.gpg ]]; then
    curl -fsSL https://dl.cloudsmith.io/public/caddy/stable/gpg.key \
      | gpg --dearmor -o /etc/apt/keyrings/caddy.gpg
    chmod 644 /etc/apt/keyrings/caddy.gpg
  fi

  if [[ ! -f /etc/apt/sources.list.d/caddy.list ]]; then
    echo "deb [signed-by=/etc/apt/keyrings/caddy.gpg] https://dl.cloudsmith.io/public/caddy/stable/deb/debian any-version main" \
      > /etc/apt/sources.list.d/caddy.list
  fi

  apt-get update -qq
  apt-get install -y -qq caddy
fi

mkdir -p /etc/caddy/sites-enabled

CADDY_MAIN="import /etc/caddy/sites-enabled/*"

if write_if_changed /etc/caddy/Caddyfile "$CADDY_MAIN"; then
  systemctl restart caddy
  ok "Caddy restarted"
else
  log "Caddy unchanged"
fi

systemctl enable caddy >/dev/null

# -------------------------------
# App dirs
# -------------------------------
mkdir -p "$APP_HOME/versions" "$APP_HOME/current"
chown -R "$APP_USER:$APP_USER" "$APP_HOME"

ok "Setup complete"
