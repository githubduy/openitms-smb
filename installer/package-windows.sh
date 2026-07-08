#!/usr/bin/env bash
# package-windows.sh — lắp bộ cài OpenITMS-SMB Windows (.zip) vào dist/ (cross-compile từ Linux/CI).
# Chạy sau khi frontend đã build vào upstream/api/public (FULL_UI=1 build-all.sh) để binary Windows
# nhúng UI. Output: dist/openitms-smb-<ver>-windows-amd64.zip + checksum.
#
# Layout zip (install.ps1 dùng $Here là ROOT — bin/ mariadb/ gitea/ ... là SIBLING):
#   install.ps1 uninstall.ps1 bin\ mariadb\ gitea\ plugins\ templates\ certs\ config\ licenses\
# MariaDB/pwsh/Gitea Windows: lấy từ dist/deps/*-win nếu có; thiếu -> bundle core-only.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
UP="$ROOT/upstream"
GO="$ROOT/Go/go/bin/go"; [ -x "$GO" ] || GO="$ROOT/Go/go/bin/go.exe"; [ -x "$GO" ] || GO=go

UPTAG="$(git -C "$UP" describe --tags --always 2>/dev/null || echo dev)"
VER="${OPENITMS_VERSION:-0.1.0}-sem${UPTAG#v}"
STAGE="$ROOT/dist/openitms-smb-$VER-windows-amd64"
echo "==> Stage $STAGE"
rm -rf "$STAGE"; mkdir -p "$STAGE"/{bin,plugins,templates,certs,config,licenses}

echo "==> [1/6] core binary (cross-compile GOOS=windows)"
if [ ! -f "$UP/api/public/index.html" ]; then
  mkdir -p "$UP/api/public"
  echo '<!-- placeholder: no UI assets (build frontend for full UI) -->' > "$UP/api/public/index.html"
fi
IMPORT="$(head -1 "$UP/go.mod" | awk '{print $2}')"
QWVER="quickwin-${OPENITMS_VERSION:-0.1.0}-sem${UPTAG#v}"
( cd "$UP" && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 "$GO" build \
    -ldflags "-X $IMPORT/util.Ver=$QWVER -X $IMPORT/util.Commit=$(git -C "$UP" rev-parse --short HEAD)" \
    -o "$STAGE/bin/openitms-app.exe" ./cli )

echo "==> [2/6] plugins (cross-compile .exe)"
for pdir in "$ROOT"/plugins/*/; do
  name="$(basename "$pdir")"
  [ -f "$pdir/go.mod" ] || continue    # bỏ plugin Python
  mkdir -p "$STAGE/plugins/$name"
  ( cd "$pdir" && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 "$GO" build -o "$STAGE/plugins/$name/$name.exe" . )
  cp "$pdir/plugin.yaml" "$STAGE/plugins/$name/"
  echo "    plugin: $name"
done

echo "==> [3/6] templates"
cp -a "$ROOT"/templates/* "$STAGE/templates/" 2>/dev/null || true

echo "==> [4/6] installer scripts (ROOT của zip)"
cp "$ROOT"/installer/windows/install.ps1 "$STAGE/install.ps1"
cp "$ROOT"/installer/windows/uninstall.ps1 "$STAGE/uninstall.ps1"

echo "==> [5/6] licenses"
cp "$ROOT"/{LICENSE,LICENSE-SEMAPHORE,NOTICE.md} "$STAGE/licenses/" 2>/dev/null || true
( cd "$UP" && "$GO" run github.com/google/go-licenses@latest report ./... > "$STAGE/licenses/THIRD_PARTY_LICENSES.md" 2>/dev/null ) || \
  echo "    (go-licenses skip)"

echo "==> [6/6] MariaDB + pwsh + Gitea (Windows, bundled nếu có)"
if [ -d "$ROOT/dist/deps/mariadb-win" ]; then cp -a "$ROOT/dist/deps/mariadb-win" "$STAGE/mariadb"; else echo "    MariaDB Windows chưa fetch — bundle core-only"; fi
if [ -d "$ROOT/dist/deps/pwsh-win" ]; then mkdir -p "$STAGE/bin/pwsh"; cp -a "$ROOT/dist/deps/pwsh-win/." "$STAGE/bin/pwsh/"; else echo "    pwsh Windows chưa fetch — bundle core-only"; fi
if [ -f "$ROOT/dist/deps/gitea/gitea.exe" ]; then mkdir -p "$STAGE/gitea"; cp "$ROOT/dist/deps/gitea/gitea.exe" "$STAGE/gitea/"; else echo "    Gitea Windows chưa fetch — bundle không có git server local"; fi

echo "==> zip + checksum"
ZIP="$ROOT/dist/openitms-smb-$VER-windows-amd64.zip"
rm -f "$ZIP"
if command -v zip >/dev/null; then
  ( cd "$ROOT/dist" && zip -qr "$ZIP" "$(basename "$STAGE")" )
else
  echo "ERROR: cần 'zip' để đóng gói Windows (.zip). CI ubuntu có sẵn." >&2; exit 1
fi
( cd "$ROOT/dist" && sha256sum "$(basename "$ZIP")" > "$(basename "$ZIP").sha256" )
echo "OK: $ZIP"
