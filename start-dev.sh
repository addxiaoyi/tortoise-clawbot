#!/usr/bin/env bash
# Tohelp: start OpenClaw gateway (Linux / macOS / Git Bash)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export OPENCLAW_CONFIG_PATH="$ROOT/.openclaw-dev/openclaw.json"

if [[ -f "$ROOT/.env" ]]; then
  echo "Loading environment from .env..." >&2
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

if [[ -z "${OPENCLAW_GATEWAY_TOKEN:-}" ]]; then
  echo "Warning: OPENCLAW_GATEWAY_TOKEN not set. Using dev-only placeholder; set in .env for real use." >&2
  export OPENCLAW_GATEWAY_TOKEN="dev-only-tohelp-gateway-token-change-me"
fi

export OPENCLAW_SKIP_CHANNELS="${OPENCLAW_SKIP_CHANNELS:-1}"
export CLAWDBOT_SKIP_CHANNELS="${CLAWDBOT_SKIP_CHANNELS:-1}"
export OPENCLAW_NO_VERSION_CHECK="${OPENCLAW_NO_VERSION_CHECK:-1}"

if [[ ! -f "$ROOT/openclaw-main/package.json" ]]; then
  echo "Error: openclaw-main/ missing. Run: npm run doctor" >&2
  exit 1
fi

echo "Starting OpenClaw Gateway..." >&2
echo "Config: $OPENCLAW_CONFIG_PATH" >&2
cd "$ROOT/openclaw-main"
exec node scripts/run-node.mjs --dev gateway
