# Sourced by the scripts beside it. Leaves the shell at the repository root.

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

# bun installs itself in ~/.bun/bin, which non-interactive shells often miss.
[[ -d "$HOME/.bun/bin" ]] && PATH="$HOME/.bun/bin:$PATH"

die() { echo "error: $*" >&2; exit 1; }

# need <command> <hint>
need() { command -v "$1" >/dev/null || die "$1 is missing — $2"; }

# load_env exports .env, the file the server reads. Variables already set win.
load_env() {
	[[ -f .env ]] || return 0
	local key value
	while IFS='=' read -r key value || [[ -n $key ]]; do
		[[ $key =~ ^[A-Za-z_][A-Za-z0-9_]*$ && -z ${!key+x} ]] || continue
		# A .env saved with Windows line endings ends every value with a
		# carriage return. It has to go before the quotes, or the closing one
		# is no longer last — and a DSN carrying it fails to parse with a
		# message about sslmode, which sends you looking in the wrong file.
		value="${value%$'\r'}"
		value="${value%\"}"
		export "$key=${value#\"}"
	done <.env
}
