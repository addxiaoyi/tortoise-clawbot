#!/usr/bin/env bash
# Cài Go trên Linux/macOS/WSL: tải archive chính thức, giải nén, (tuỳ chọn) thêm PATH.
set -euo pipefail

VERSION="${VERSION:-}"
INSTALL_DIR="${INSTALL_DIR:-}"
FORCE="${FORCE:-0}"
SKIP_PATH="${SKIP_PATH:-0}"

usage() {
  cat <<'EOF'
Dùng: curl -fsSL .../install-go.sh | bash
Hoặc: ./scripts/install-go.sh [--version 1.23.4] [--prefix "$HOME/.local/go"] [--force]

Biến môi trường:
  VERSION     Ví dụ 1.23.4 (mặc định: stable mới nhất từ go.dev)
  INSTALL_DIR Thư mục đích (mặc định: $HOME/.local/go)
  FORCE=1     Cài lại dù đã đủ mới
  SKIP_PATH=1 Không ghi vào ~/.profile hoặc ~/.zshrc
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --prefix) INSTALL_DIR="$2"; shift 2 ;;
    --force) FORCE=1; shift ;;
    --skip-path) SKIP_PATH=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Tham số không rõ: $1" >&2; usage; exit 1 ;;
  esac
done

normalize_ver() {
  local v="$1"
  if [[ "$v" =~ ^go([0-9]+\.[0-9]+\.[0-9]+) ]]; then echo "${BASH_REMATCH[1]}"; return; fi
  if [[ "$v" =~ ^([0-9]+\.[0-9]+\.[0-9]+) ]]; then echo "${BASH_REMATCH[1]}"; return; fi
  echo ""
}

compare_ver() {
  local IFS=.
  local -a a=($1) b=($2)
  local i x y
  for i in 0 1 2; do
    x="${a[$i]:-0}"
    y="${b[$i]:-0}"
    if (( x < y )); then echo -1; return; fi
    if (( x > y )); then echo 1; return; fi
  done
  echo 0
}

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) GOARCH=amd64 ;;
  aarch64|arm64) GOARCH=arm64 ;;
  *) echo "Kiến trúc không hỗ trợ: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux) GOOS=linux ;;
  darwin) GOOS=darwin ;;
  *) echo "Hệ điều hành không hỗ trợ: $OS" >&2; exit 1 ;;
esac

if [[ -z "$INSTALL_DIR" ]]; then
  INSTALL_DIR="${HOME}/.local/go"
fi

json="$(curl -fsSL "https://go.dev/dl/?mode=json")" || { echo "Không tải được JSON từ go.dev" >&2; exit 1; }

if [[ -z "$VERSION" ]]; then
  VERSION="$(echo "$json" | python3 -c '
import json, sys
data = json.load(sys.stdin)
for row in data:
    if row.get("stable") and row.get("version", "").startswith("go"):
        print(row["version"][2:])
        break
else:
    for row in data:
        v = row.get("version", "")
        if v.startswith("go"):
            print(v[2:])
            break
')"
fi

VERSION="$(normalize_ver "$VERSION")"
[[ -n "$VERSION" ]] || { echo "Không xác định được phiên bản." >&2; exit 1; }

if command -v go >/dev/null 2>&1 && [[ "$FORCE" != "1" ]]; then
  cur="$(go env GOVERSION 2>/dev/null || true)"
  cur="$(normalize_ver "$cur")"
  if [[ -n "$cur" ]]; then
    cmp="$(compare_ver "$cur" "$VERSION")"
    if [[ "$cmp" -ge 0 ]]; then
      echo "[install-go] Đã có Go $cur (>= $VERSION). Bỏ qua. FORCE=1 để cài lại."
      exit 0
    fi
    echo "[install-go] Go hiện tại: $cur — nâng lên $VERSION"
  fi
fi

NAME="go${VERSION}.${GOOS}-${GOARCH}.tar.gz"
URL="https://go.dev/dl/${NAME}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "[install-go] Đang tải $URL"
curl -fsSL "$URL" -o "$TMP/$NAME"

PARENT="$(dirname "$INSTALL_DIR")"
mkdir -p "$PARENT"
if [[ -d "$INSTALL_DIR" ]]; then
  echo "[install-go] Xoá bản cũ: $INSTALL_DIR"
  rm -rf "$INSTALL_DIR"
fi

tar -xzf "$TMP/$NAME" -C "$PARENT"
mv "$PARENT/go" "$INSTALL_DIR"

GOBIN="$INSTALL_DIR/bin"
if [[ "$SKIP_PATH" != "1" ]]; then
  LINE="export PATH=\"$GOBIN:\$PATH\""
  RC="${HOME}/.profile"
  [[ -f "${HOME}/.zshrc" ]] && RC="${HOME}/.zshrc"
  if [[ -f "$RC" ]] && grep -Fq "$GOBIN" "$RC" 2>/dev/null; then
    echo "[install-go] PATH đã có trong $RC"
  else
    touch "$RC"
    echo "" >> "$RC"
    echo "# Go (install-go.sh)" >> "$RC"
    echo "$LINE" >> "$RC"
    echo "[install-go] Đã thêm vào $RC"
  fi
  export PATH="$GOBIN:$PATH"
fi

echo "[install-go] $($GOBIN/go version)"
echo "[install-go] Xong. Chạy: source $RC (hoặc mở terminal mới)"
