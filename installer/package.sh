#!/usr/bin/env bash
# package.sh — lắp bộ cài OpenITMS-SMB Linux hoàn chỉnh vào dist/ (toàn binary native, không Docker).
# Chạy sau build-all.sh. Output: dist/openitms-smb-<ver>-linux-amd64.tar.gz + checksum.
#
# Thành phần MariaDB + pwsh: tải bằng fetch-deps.sh (pin checksum) — nếu chưa có thì cảnh báo,
# đóng gói bản "core + plugins + templates" (đủ để chạy với DB ngoài / dev).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GO="$ROOT/Go/go/bin/go"; [ -x "$GO" ] || GO="$ROOT/Go/go/bin/go.exe"; [ -x "$GO" ] || GO=go

UPTAG="$(git -C "$ROOT/upstream" describe --tags --always 2>/dev/null || echo dev)"
VER="${OPENITMS_VERSION:-0.1.0}-sem${UPTAG#v}"
STAGE="$ROOT/dist/openitms-smb-$VER-linux-amd64"
echo "==> Stage $STAGE"
rm -rf "$STAGE"; mkdir -p "$STAGE"/{bin,plugins,templates,certs,config,licenses}

echo "==> [1/6] core binary"
[ -f "$ROOT/dist/bin/semaphore" ] || { echo "ERROR: chạy build-all.sh (Linux) trước"; exit 1; }
cp "$ROOT/dist/bin/semaphore" "$STAGE/bin/"

echo "==> [2/6] plugins (build từng plugin Go)"
for pdir in "$ROOT"/plugins/*/; do
  name="$(basename "$pdir")"
  [ -f "$pdir/go.mod" ] || continue    # bỏ plugin Python (đóng gói riêng khi có runtime)
  mkdir -p "$STAGE/plugins/$name"
  ( cd "$pdir" && "$GO" build -o "$STAGE/plugins/$name/$name" . )
  cp "$pdir/plugin.yaml" "$STAGE/plugins/$name/"
  echo "    plugin: $name"
done

echo "==> [3/6] templates"
cp -a "$ROOT"/templates/* "$STAGE/templates/" 2>/dev/null || true

echo "==> [4/6] installer scripts + systemd + my.cnf"
cp "$ROOT"/installer/linux/{install.sh,uninstall.sh,my.cnf} "$STAGE/"
cp -a "$ROOT"/installer/linux/systemd "$STAGE/"

echo "==> [5/6] licenses (nghĩa vụ MIT + GPL — plan 2.1)"
cp "$ROOT"/{LICENSE,LICENSE-SEMAPHORE,NOTICE.md} "$STAGE/licenses/"
( cd "$ROOT/upstream" && "$GO" run github.com/google/go-licenses@latest report ./... > "$STAGE/licenses/THIRD_PARTY_LICENSES.md" 2>/dev/null ) || \
  echo "    (go-licenses report skip — offline; CI sinh đầy đủ)"

echo "==> [6/6] MariaDB + pwsh (bundled)"
if [ -d "$ROOT/dist/deps/mariadb" ]; then cp -a "$ROOT/dist/deps/mariadb" "$STAGE/"; else echo "    MariaDB chưa fetch (installer/fetch-deps.sh) — bundle core-only"; fi
if [ -d "$ROOT/dist/deps/pwsh" ]; then mkdir -p "$STAGE/bin/pwsh"; cp -a "$ROOT/dist/deps/pwsh/." "$STAGE/bin/pwsh/"; else echo "    pwsh chưa fetch — bundle core-only"; fi

echo "==> tar.gz + checksum"
TAR="$ROOT/dist/openitms-smb-$VER-linux-amd64.tar.gz"
( cd "$ROOT/dist" && tar -czf "$TAR" "$(basename "$STAGE")" )
( cd "$ROOT/dist" && sha256sum "$(basename "$TAR")" > "$(basename "$TAR").sha256" )
echo "OK: $TAR"
