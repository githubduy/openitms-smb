#!/usr/bin/env bash
# fetch-deps.sh — tải binary dependency (MariaDB + PowerShell Core) với PIN CHECKSUM,
# giải nén vào dist/deps/ để package.sh lắp vào bundle. Chạy trên Linux (release CI).
#
# Version + sha256 pin trong deps.lock (KHÔNG tự bump — đổi có review, tránh supply-chain).
# GPL: MariaDB (GPLv2) đóng gói binary độc lập giao tiếp socket — mere aggregation (plan 2.2).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOCK="$ROOT/installer/deps.lock"
OUT="$ROOT/dist/deps"
mkdir -p "$OUT"

[ -f "$LOCK" ] || { echo "ERROR: thiếu $LOCK (khai version+sha256 dependency)"; exit 1; }

verify_sha() { echo "$2  $1" | sha256sum -c - >/dev/null; }

fetch() {  # name url sha256 -> tải + verify vào /tmp
  local name="$1" url="$2" want="$3" f="/tmp/$name"
  echo "==> tải $name"
  curl -fsSL -o "$f" "$url"
  if ! verify_sha "$f" "$want"; then
    echo "ERROR: checksum LỆCH cho $name — DỪNG (nghi ngờ supply-chain)"; exit 1
  fi
  echo "$f"
}

# đọc deps.lock: dòng "name|url|sha256|extract-into"
while IFS='|' read -r name url sha dest; do
  case "$name" in ''|'#'*) continue;; esac
  f="$(fetch "$name" "$url" "$sha")"
  mkdir -p "$OUT/$dest"
  case "$f" in
    *.tar.gz|*.tgz) tar -C "$OUT/$dest" --strip-components=1 -xzf "$f" ;;
    *.tar.xz)       tar -C "$OUT/$dest" --strip-components=1 -xJf "$f" ;;
    *.zip)          ( cd "$OUT/$dest" && unzip -q "$f" ) ;;
    *) cp "$f" "$OUT/$dest/" ;;
  esac
  echo "    → $OUT/$dest"
done < "$LOCK"

echo "OK: deps → $OUT (package.sh sẽ lắp vào bundle)"
