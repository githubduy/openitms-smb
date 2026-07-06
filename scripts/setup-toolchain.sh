#!/usr/bin/env bash
# setup-toolchain.sh — tải Go (version pin trong go-version.txt) vào <repo>/Go/
# (yêu cầu dự án: toolchain nằm trong project, KHÔNG dùng Go hệ thống).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GOVER="$(head -1 "$ROOT/go-version.txt" | tr -d '[:space:]')"
GODIR="$ROOT/Go"

if [ -x "$GODIR/go/bin/go" ] || [ -x "$GODIR/go/bin/go.exe" ]; then
  cur="$("$GODIR"/go/bin/go version 2>/dev/null | awk '{print $3}' | sed 's/^go//')" || cur=""
  if [ "$cur" = "$GOVER" ]; then echo "OK: Go $GOVER đã có ở $GODIR"; exit 0; fi
  echo "Go hiện tại ($cur) != pin ($GOVER) — tải lại."
  rm -rf "$GODIR/go"
fi

case "$(uname -s)" in
  Linux)  OS=linux;  EXT=tar.gz ;;
  Darwin) OS=darwin; EXT=tar.gz ;;
  MINGW*|MSYS*|CYGWIN*) OS=windows; EXT=zip ;;
  *) echo "ERROR: OS không hỗ trợ: $(uname -s)" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "ERROR: arch không hỗ trợ: $(uname -m)" >&2; exit 1 ;;
esac

URL="https://go.dev/dl/go${GOVER}.${OS}-${ARCH}.${EXT}"
mkdir -p "$GODIR"
echo "==> tải $URL"
curl -fsSL -o "$GODIR/go-dist.$EXT" "$URL"
echo "==> giải nén vào $GODIR"
if [ "$EXT" = "zip" ]; then
  (cd "$GODIR" && unzip -q go-dist.zip)
else
  tar -C "$GODIR" -xzf "$GODIR/go-dist.$EXT"
fi
rm -f "$GODIR/go-dist.$EXT"
"$GODIR/go/bin/go" version
echo "OK: Go $GOVER tại $GODIR/go/bin/go"
