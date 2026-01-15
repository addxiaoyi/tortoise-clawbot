#!/usr/bin/env bash
# Replaces openclaw-main/ with https://github.com/openclaw/openclaw as a git submodule.
# DESTRUCTIVE. Read docs/setup-openclaw-submodule.md first.
set -euo pipefail

if [[ "${TOHELP_CONFIRM_SUBMODULE_INIT:-}" != "1" ]]; then
  echo "Refusing: set TOHELP_CONFIRM_SUBMODULE_INIT=1 after backup." >&2
  echo "See docs/setup-openclaw-submodule.md" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -d .git ]]; then
  echo "Error: not a git repository root (missing .git)." >&2
  exit 1
fi

if [[ -d openclaw-main ]]; then
  echo "Removing openclaw-main/ ..." >&2
  rm -rf openclaw-main
fi

echo "Adding submodule openclaw-main ..." >&2
git submodule add https://github.com/openclaw/openclaw.git openclaw-main
git submodule update --init --recursive

echo "" >&2
echo "Done. Next: install deps inside openclaw-main (see upstream README), then npm run doctor" >&2
