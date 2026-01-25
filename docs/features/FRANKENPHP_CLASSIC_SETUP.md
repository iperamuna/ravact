# FrankenPHP Classic Mode Setup

[← Back to Documentation](../README.md)

Below is a full, copy-paste “article” (markdown) you can feed into Claude Code (or your Go TUI help/docs), plus an interactive Bash script that:
•	Prompts for site key (unique name) → used in .service filename + description + socket name
•	Prompts for site root (replaces /var/www/siteA/current everywhere)
•	Prompts for docroot (defaults to <root>/public)
•	Prompts for user/group (defaults www-data)
•	Writes a systemd service to /etc/systemd/system/frankenphp-<site>.service
•	Enables + starts it
•	Prints an Nginx vhost to stdout (and optionally writes it to /etc/nginx/sites-available/<site>.conf and enables it)

⸻

Article: Option A — One FrankenPHP Classic Mode Process per Site (systemd + Nginx)

Why this is the “best default”

Running one FrankenPHP classic-mode server per site gives you clean isolation:
•	Independent restart/deploy per site
•	Separate environment variables per app
•	Dedicated unix socket per site (/run/frankenphp/<site>.sock)
•	Failures don’t cascade across tenants/apps
•	Easy to monitor per-service (systemctl status frankenphp-…, journalctl -u …)

This pattern is especially good for Laravel and other PHP apps with different .env files, queue configs, or release cycles.

⸻

Target architecture

For each site:
•	systemd service: frankenphp-<site>.service
•	socket: /run/frankenphp/<site>.sock
•	app root: /var/www/<site>/current (example)
•	web root: /var/www/<site>/current/public (Laravel default)
•	Nginx vhost proxies all requests to that socket

Internet
|
Nginx (TLS, HTTP/2, vhosts)
|
+--> unix:/run/frankenphp/siteA.sock  -> frankenphp-siteA.service -> /var/www/siteA/current/public
|
+--> unix:/run/frankenphp/siteB.sock  -> frankenphp-siteB.service -> /var/www/siteB/current/public


⸻

Prerequisites
•	FrankenPHP installed somewhere like:
•	/usr/local/bin/frankenphp (recommended for custom installs)
•	or /usr/bin/frankenphp (distro packages)
•	Nginx installed and running
•	systemd (Ubuntu/Debian typical)
•	Your app files exist on disk and permissions are correct

Quick check:

which frankenphp && frankenphp --version
nginx -v
systemctl --version | head -n 1


⸻

Recommended folder layout (works great for deployments)

A common deploy layout:
•	/var/www/siteA/releases/<timestamp> (new release)
•	/var/www/siteA/current -> releases/<timestamp> (symlink)

This way you can deploy by swapping symlink, then restarting the specific service:

sudo systemctl restart frankenphp-siteA


⸻

systemd unit file (per site)

Create: /etc/systemd/system/frankenphp-siteA.service

[Unit]
Description=FrankenPHP classic mode (siteA)
After=network.target
Wants=network.target

[Service]
Type=simple

User=www-data
Group=www-data

# App root (NOT public). Useful if you run artisan commands etc.
WorkingDirectory=/var/www/siteA/current

# Optional: environment variables (add more as needed)
Environment=APP_ENV=production
Environment=APP_BASE_PATH=/var/www/siteA/current

# Runtime socket lives in /run (tmpfs). This is the key isolation piece.
RuntimeDirectory=frankenphp
RuntimeDirectoryMode=0755

# Start FrankenPHP in classic mode, serving the docroot.
# For Laravel, docroot is typically /public.
ExecStart=/usr/local/bin/frankenphp php-server \
--listen unix:/run/frankenphp/siteA.sock \
--root /var/www/siteA/current/public

Restart=always
RestartSec=2
TimeoutStopSec=10

# Hardening (safe defaults; remove if they break your use case)
NoNewPrivileges=true
PrivateTmp=true

# Logging goes to journald:
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target

Apply it:

sudo systemctl daemon-reload
sudo systemctl enable --now frankenphp-siteA
sudo systemctl status frankenphp-siteA --no-pager

View logs:

journalctl -u frankenphp-siteA -f


⸻

Nginx vhost (per site, unix socket)

Example: /etc/nginx/sites-available/siteA.conf

server {
listen 80;
server_name siteA.com www.siteA.com;

    # If you're using certbot, you can keep 80 for ACME and redirect to HTTPS.
    # return 301 https://$host$request_uri;

    location / {
        proxy_pass http://unix:/run/frankenphp/siteA.sock;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP $remote_addr;
    }

    client_max_body_size 50m;
}

Enable & reload:

sudo ln -s /etc/nginx/sites-available/siteA.conf /etc/nginx/sites-enabled/siteA.conf
sudo nginx -t && sudo systemctl reload nginx


⸻

HTTPS (recommended)

For production, terminate TLS at Nginx (LetsEncrypt / certbot). Your upstream remains the unix socket.

⸻

Common pitfalls & fixes

1) 502 Bad Gateway
   •	Check service is running: systemctl status frankenphp-siteA
   •	Check socket exists: ls -l /run/frankenphp/siteA.sock
   •	Check permissions: Nginx user must be able to connect to socket
   Usually both run as www-data → easiest.

2) Wrong docroot
   •	For Laravel, docroot should be <root>/public
   •	For WordPress it’s typically the project root (where index.php is)

3) Deploy with zero downtime
   •	This pattern isolates sites; you can restart per site quickly.
   •	For truly seamless reloads, you’ll need a more advanced strategy (dual socket + switch, or run two instances and swap upstream). Most teams accept per-site fast restarts.

⸻

Interactive Bash generator (systemd + prints Nginx vhost)

Save as frankenphp-site-gen.sh, then: chmod +x frankenphp-site-gen.sh and run with sudo ./frankenphp-site-gen.sh

#!/usr/bin/env bash
set -euo pipefail

# FrankenPHP per-site systemd + nginx vhost generator (classic mode)
# - Writes: /etc/systemd/system/frankenphp-<site>.service
# - Enables + starts service
# - Prints nginx vhost to stdout
# - Optional: writes nginx conf to /etc/nginx/sites-available/<site>.conf and enables it

color() { printf "\033[%sm%s\033[0m" "$1" "$2"; }
info()  { echo "$(color 36 "[INFO]") $*"; }
warn()  { echo "$(color 33 "[WARN]") $*"; }
err()   { echo "$(color 31 "[ERR ]") $*" >&2; }
die()   { err "$*"; exit 1; }

need_root() {
if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
die "Run as root (sudo)."
fi
}

has_cmd() { command -v "$1" >/dev/null 2>&1; }

suggest_site_key() {
local root="$1"
# Try to derive a decent default key from root path
# /var/www/siteA/current -> siteA
# /var/www/my-app -> my-app
local base
base="$(basename "$(dirname "$root")" 2>/dev/null || true)"
if [[ -z "${base}" || "${base}" == "/" || "${base}" == "www" ]]; then
base="$(basename "$root")"
fi
# sanitize
base="${base,,}"                 # lowercase
base="${base// /-}"              # spaces to dash
base="$(echo "$base" | tr -cd 'a-z0-9._-')"  # safe chars
echo "${base:-site}"
}

need_root

has_cmd systemctl || die "systemctl not found (systemd required)."
has_cmd nginx || warn "nginx not found in PATH. (Vhost will still be generated.)"

DEFAULT_USER="www-data"
DEFAULT_GROUP="www-data"

# Detect frankenphp binary
FRANKENPHP_BIN="$(command -v frankenphp || true)"
if [[ -z "$FRANKENPHP_BIN" ]]; then
# common fallbacks
for p in /usr/local/bin/frankenphp /usr/bin/frankenphp; do
if [[ -x "$p" ]]; then FRANKENPHP_BIN="$p"; break; fi
done
fi
[[ -n "$FRANKENPHP_BIN" ]] || die "frankenphp binary not found. Install it and ensure it's in PATH."

echo
info "FrankenPHP binary: $FRANKENPHP_BIN"

echo
read -rp "Enter site root (e.g. /var/www/siteA/current): " SITE_ROOT
[[ -n "$SITE_ROOT" ]] || die "Site root is required."
[[ -d "$SITE_ROOT" ]] || warn "Directory does not exist yet: $SITE_ROOT (continuing anyway)."

DEFAULT_KEY="$(suggest_site_key "$SITE_ROOT")"
read -rp "Enter unique site key [${DEFAULT_KEY}] (used for service/socket filenames): " SITE_KEY
SITE_KEY="${SITE_KEY:-$DEFAULT_KEY}"
# sanitize key
SITE_KEY="${SITE_KEY,,}"
SITE_KEY="${SITE_KEY// /-}"
SITE_KEY="$(echo "$SITE_KEY" | tr -cd 'a-z0-9._-')"
[[ -n "$SITE_KEY" ]] || die "Site key is required."

DEFAULT_DOCROOT="${SITE_ROOT%/}/public"
read -rp "Enter docroot [${DEFAULT_DOCROOT}]: " DOCROOT
DOCROOT="${DOCROOT:-$DEFAULT_DOCROOT}"

read -rp "Run as user [${DEFAULT_USER}]: " RUN_USER
RUN_USER="${RUN_USER:-$DEFAULT_USER}"

read -rp "Run as group [${DEFAULT_GROUP}]: " RUN_GROUP
RUN_GROUP="${RUN_GROUP:-$DEFAULT_GROUP}"

SOCK="/run/frankenphp/${SITE_KEY}.sock"
SERVICE_NAME="frankenphp-${SITE_KEY}"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

echo
read -rp "Enter domain names for nginx server_name (space-separated, e.g. siteA.com www.siteA.com) [${SITE_KEY}.test]: " DOMAINS
DOMAINS="${DOMAINS:-${SITE_KEY}.test}"

info "Generating systemd unit: $SERVICE_FILE"

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=FrankenPHP classic mode (${SITE_KEY})
After=network.target
Wants=network.target

[Service]
Type=simple
User=${RUN_USER}
Group=${RUN_GROUP}
WorkingDirectory=${SITE_ROOT}

Environment=APP_ENV=production
Environment=APP_BASE_PATH=${SITE_ROOT}

RuntimeDirectory=frankenphp
RuntimeDirectoryMode=0755

ExecStart=${FRANKENPHP_BIN} php-server --listen unix:${SOCK} --root ${DOCROOT}

Restart=always
RestartSec=2
TimeoutStopSec=10

NoNewPrivileges=true
PrivateTmp=true

StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

info "Reloading systemd + enabling service..."
systemctl daemon-reload
systemctl enable --now "$SERVICE_NAME"

echo
info "Service status (brief):"
systemctl --no-pager --full status "$SERVICE_NAME" | sed -n '1,18p' || true

echo
info "Nginx vhost (copy/paste):"
echo "------------------------------------------------------------"
cat <<NGINX
server {
listen 80;
server_name ${DOMAINS};

    location / {
        proxy_pass http://unix:${SOCK};
        proxy_set_header Host \$host;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    client_max_body_size 50m;
}
NGINX
echo "------------------------------------------------------------"

echo
read -rp "Write nginx conf to /etc/nginx/sites-available/${SITE_KEY}.conf and enable it? [y/N]: " WRITE_NGINX
WRITE_NGINX="${WRITE_NGINX:-N}"

if [[ "$WRITE_NGINX" =~ ^[Yy]$ ]]; then
NGINX_AVAIL="/etc/nginx/sites-available/${SITE_KEY}.conf"
NGINX_ENABLED="/etc/nginx/sites-enabled/${SITE_KEY}.conf"

info "Writing: $NGINX_AVAIL"
cat > "$NGINX_AVAIL" <<NGINX
server {
listen 80;
server_name ${DOMAINS};

    location / {
        proxy_pass http://unix:${SOCK};
        proxy_set_header Host \$host;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    client_max_body_size 50m;
}
NGINX

if [[ ! -e "$NGINX_ENABLED" ]]; then
info "Enabling: $NGINX_ENABLED"
ln -s "$NGINX_AVAIL" "$NGINX_ENABLED"
else
warn "Already enabled: $NGINX_ENABLED"
fi

if has_cmd nginx; then
info "Testing nginx config..."
nginx -t
info "Reloading nginx..."
systemctl reload nginx
else
warn "nginx command not found; skipping nginx -t and reload."
fi
fi

echo
info "Done."
info "Service:   systemctl status ${SERVICE_NAME}"
info "Logs:      journalctl -u ${SERVICE_NAME} -f"
info "Socket:    ${SOCK}"


⸻

Notes you can bake into your Go TUI UX
•	Default docroot should be:
•	Laravel: <root>/public
•	WordPress: <root> (where index.php is)
•	Site key should be a safe slug (a-z0-9._-)
•	Prefer running both Nginx + FrankenPHP as the same user/group (www-data) to avoid socket permission headaches
•	Use journald for logs; show quick commands for tailing logs inside the TUI

If you want, I can also give you a second script mode that generates everything but does not write any files (dry-run), which is often perfect for a TUI “preview then apply” flow.