#!/usr/bin/env bash
#
# Run the whole stack: the Loginer API and the console and id apps.
#
#   scripts/start.sh --dev     Vite dev servers, hot reload (default)
#   scripts/start.sh --prod    production builds, served by Node
#
# The API logs to this terminal; the apps log to .logs/<app>.log. Ctrl-C stops
# everything. The URLs are the same in both modes, so .env fits both.

set -euo pipefail
. "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

case "${1:---dev}" in
--dev | dev) mode=dev ;;
--prod | prod) mode=prod ;;
*) die "usage: scripts/start.sh [--dev|--prod]" ;;
esac

need go "https://go.dev/dl"
need node "https://nodejs.org"
need bun "https://bun.sh"
load_env

api_port="${LOGINER_ADDR:-:8080}" api_port="${api_port##*:}"
admin_port="${LOGINER_ADMIN_ADDR:-:8081}" admin_port="${admin_port##*:}"

# name  port  API it calls                   paths routed to that API
apps=(
	"id      5173 http://localhost:$api_port   /api/v1/account,/oauth2,/.well-known"
	"console 5174 http://localhost:$admin_port /api/v1/admin"
)

logs="$root/.logs"
mkdir -p "$logs" bin

# A port is taken when anything at all answers on it (curl exit 7: refused).
for port in "$api_port" "$admin_port" 5173 5174; do
	code=0 && curl -s -o /dev/null --max-time 1 "http://localhost:$port" || code=$?
	[[ $code == 7 ]] || die "port $port is already in use"
done

for app in "${apps[@]}"; do
	read -r name _ <<<"$app"
	[[ -d web/$name/node_modules ]] || die "web/$name has no dependencies — run make setup"
done

# Background jobs ignore Ctrl-C, so stopping is done here, for all of them.
pids=()
trap 'trap - EXIT; kill ${pids[@]+"${pids[@]}"} 2>/dev/null; wait' EXIT
trap 'exit 0' INT TERM # stopping on purpose is not a failure

# ready <name> <port> <pid>: report once the app answers, or show why it died.
ready() {
	for _ in $(seq 240); do
		kill -0 "$3" 2>/dev/null || { printf '\n✗ %s stopped — .logs/%s.log:\n' "$1" "$1"; tail -n 20 "$logs/$1.log"; return; }
		curl -s -o /dev/null --max-time 1 "http://localhost:$2" && { echo "✓ $1 on http://localhost:$2"; return; }
		sleep 0.5
	done
}

echo "==> building ($mode)"
if [[ $mode == prod ]]; then
	builds=()
	for app in "${apps[@]}"; do
		read -r name _ <<<"$app"
		(cd "web/$name" && exec bun run build) >"$logs/$name-build.log" 2>&1 &
		builds+=("$name:$!")
		pids+=($!)
	done
	version="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$version" -o bin/loginer ./cmd/loginer
	for build in "${builds[@]}"; do
		wait "${build#*:}" || { tail -n 30 "$logs/${build%%:*}-build.log"; die "web/${build%%:*} did not build"; }
	done
else
	go build -o bin/loginer ./cmd/loginer
fi

for app in "${apps[@]}"; do
	read -r name port api paths <<<"$app"
	(
		cd "web/$name"
		if [[ $mode == prod ]]; then
			PORT=$port ORIGIN=http://localhost:$port API_URL=$api API_PATHS=$paths exec node ../serve.js
		else
			API_URL=$api exec node_modules/.bin/vite dev
		fi
	) >"$logs/$name.log" 2>&1 &
	pids+=($!)
	ready "$name" "$port" $! &
	pids+=($!)
done

echo "==> Loginer ($mode) — app logs in .logs/"
# Gin's route listing is noise in production.
[[ $mode == prod ]] && export GIN_MODE="${GIN_MODE:-release}"
bin/loginer &
pids+=($!)
wait $!
