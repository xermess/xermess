#!/usr/bin/env bash
#
# Start the production stack on this machine: deploy/compose.yaml with
# deploy/compose.local.yaml over it, served on https://id.localhost and
# https://admin-id.localhost. Safe to run again.
#
# What differs from a server is only what cannot work on one machine — see
# deploy/Caddyfile.local. Everything else is the stack that would be deployed.

set -euo pipefail
. "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

need docker "https://docs.docker.com/engine/install"
need openssl "install openssl"

compose=(docker compose -f deploy/compose.yaml -f deploy/compose.local.yaml)

# set_value KEY VALUE — in deploy/.env, whether or not the key is there yet.
# The delimiter is | because a base64 secret can contain / but never that.
set_value() {
	grep -qE "^$1=" deploy/.env || echo "$1=" >>deploy/.env
	sed -i.bak "s|^$1=.*|$1=$2|" deploy/.env && rm -f deploy/.env.bak
}

# An existing deploy/.env is left alone: its XERMESS_SECRET_KEY is what the
# stored signing keys are sealed with, and its POSTGRES_PASSWORD is the one
# the database was initialised with — Postgres reads that only once, on an
# empty data directory, so replacing it here would lock the API out.
if [[ ! -f deploy/.env ]]; then
	umask 077
	cp deploy/.env.example deploy/.env

	set_value PUBLIC_URL https://id.localhost
	set_value ADMIN_URL https://admin-id.localhost
	set_value XERMESS_SECRET_KEY "$(openssl rand -base64 32)"
	set_value POSTGRES_PASSWORD "$(openssl rand -hex 32)"
	# So the first administrator can be created without an authenticator app.
	# A server runs this as required; deploy/compose.yaml asks for that.
	set_value XERMESS_ADMIN_MFA optional

	echo "created deploy/.env for https://id.localhost — it is gitignored"
fi

"${compose[@]}" up -d --build --wait

cat <<'TEXT'

Ready.

  https://admin-id.localhost/admin/login   the panel — makes the first administrator
  https://id.localhost/login               the sign-in pages

The certificate is Caddy's own CA, so the browser will warn once; the foot of
deploy/Caddyfile.local says how to trust it. Chrome and Firefox resolve
*.localhost themselves, so /etc/hosts needs nothing.

  make deploy-local-logs   follow the API
  make deploy-local-down   stop it and drop its data
TEXT
