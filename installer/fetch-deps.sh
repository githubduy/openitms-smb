#!/usr/bin/env bash
# fetch-deps.sh — chuẩn bị binary dependency (MariaDB + PowerShell Core) vào dist/deps/
# để package.sh lắp vào bundle. Chạy trên Linux (release CI hoặc build máy).
#
# 2 NGUỒN (ưu tiên local trước — hợp air-gapped / mạng chặn external):
#   1. LOCAL COPY: nếu có binary sẵn ở installer/vendor/<name>/ HOẶC env DEPS_<NAME>_DIR
#      → COPY thẳng (không tải mạng). Đây là cách khuyến nghị: maintainer stage sẵn MariaDB binary.
#   2. DOWNLOAD: nếu không có local → tải từ URL trong deps.lock + verify sha256.
#
# Version + sha256 pin trong deps.lock (KHÔNG tự bump — supply-chain guard).
# GPL: MariaDB (GPLv2) đóng gói binary độc lập giao tiếp socket — mere aggregation (plan 2.2).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOCK="$ROOT/installer/deps.lock"
VENDOR="$ROOT/installer/vendor"   # nơi stage binary local (gitignored — không commit binary lớn)
OUT="$ROOT/dist/deps"
mkdir -p "$OUT"
[ -f "$LOCK" ] || { echo "ERROR: thiếu $LOCK"; exit 1; }

verify_sha() { echo "$2  $1" | sha256sum -c - >/dev/null; }

extract_into() {  # file dest
  local f="$1" dest="$2"
  mkdir -p "$dest"
  case "$f" in
    *.tar.gz|*.tgz) tar -C "$dest" --strip-components=1 -xzf "$f" ;;
    *.tar.xz)       tar -C "$dest" --strip-components=1 -xJf "$f" ;;
    *.zip)
      # strip 1 top-level dir nếu zip bọc trong 1 thư mục (vd mariadb-11.4.4-winx64/)
      local tmp; tmp="$(mktemp -d)"
      ( cd "$tmp" && unzip -oq "$f" )
      local top; top=( "$tmp"/* )
      if [ "${#top[@]}" -eq 1 ] && [ -d "${top[0]}" ]; then
        cp -a "${top[0]}/." "$dest/"
      else
        cp -a "$tmp/." "$dest/"
      fi
      rm -rf "$tmp"
      ;;
    *)              cp "$f" "$dest/" ;;
  esac
}

while IFS='|' read -r name url sha dest; do
  case "$name" in ''|'#'*) continue;; esac
  name="$(echo "$name" | xargs)"; dest="$(echo "$dest" | xargs)"

  # (1) LOCAL COPY — env DEPS_<NAME>_DIR (thư mục đã giải nén) hoặc installer/vendor/<name>/
  envvar="DEPS_$(echo "$name" | tr 'a-z-' 'A-Z_')_DIR"
  localdir="${!envvar:-}"
  [ -z "$localdir" ] && [ -d "$VENDOR/$name" ] && localdir="$VENDOR/$name"
  if [ -n "$localdir" ] && [ -d "$localdir" ]; then
    echo "==> [$name] COPY từ local: $localdir (không tải mạng)"
    mkdir -p "$OUT/$dest"; cp -a "$localdir/." "$OUT/$dest/"
    echo "    → $OUT/$dest"
    continue
  fi
  # (1b) LOCAL ARCHIVE — installer/vendor/<name>.tar.gz (đã pin sha256)
  for ext in tar.gz tgz tar.xz zip; do
    if [ -f "$VENDOR/$name.$ext" ]; then
      echo "==> [$name] giải nén archive local: $VENDOR/$name.$ext"
      verify_sha "$VENDOR/$name.$ext" "$sha" || { echo "ERROR: checksum LỆCH $name (local archive)"; exit 1; }
      extract_into "$VENDOR/$name.$ext" "$OUT/$dest"; echo "    → $OUT/$dest"; continue 2
    fi
  done

  # (2) DOWNLOAD — fallback
  [ -n "$url" ] || { echo "WARN: [$name] không có local copy + không có URL → bỏ qua (bundle thiếu $name)"; continue; }
  echo "==> [$name] tải: $url"
  f="/tmp/$name.dl"
  curl -fsSL -o "$f" "$url" || { echo "WARN: tải $name fail (mạng?) → bỏ qua"; continue; }
  verify_sha "$f" "$sha" || { echo "ERROR: checksum LỆCH $name — DỪNG (nghi supply-chain)"; exit 1; }
  extract_into "$f" "$OUT/$dest"; rm -f "$f"; echo "    → $OUT/$dest"
done < "$LOCK"

echo "OK: deps → $OUT (package.sh sẽ lắp vào bundle)"
