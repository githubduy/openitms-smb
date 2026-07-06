#!/usr/bin/env bash
# sonar-scan.sh — quét OpenITMS-SMB bằng SonarQube self-host.
# Config public (project key/sources/exclusions): sonar-project.properties.
# Host + token là THÔNG TIN NHẠY CẢM — KHÔNG commit; truyền qua env:
#   export SONAR_HOST_URL=http://<sonarqube-host>:9000
#   export SONAR_TOKEN=<project analysis token>
#   [SONAR_SCANNER=/path/to/sonar-scanner]   # mặc định: sonar-scanner trong PATH
# Sinh coverage Go trước rồi nạp vào scan (gate coverage mới có ý nghĩa).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

: "${SONAR_HOST_URL:?export SONAR_HOST_URL=http://host:9000}"
: "${SONAR_TOKEN:?export SONAR_TOKEN=<token>}"
SCANNER="${SONAR_SCANNER:-sonar-scanner}"
GO="$ROOT/Go/go/bin/go"; [ -x "$GO" ] || GO="$ROOT/Go/go/bin/go.exe"; [ -x "$GO" ] || GO=go

echo "==> Sinh coverage cho các module của ta"
COVERPATHS=""
for mod in plugin-manager sdk plugins/hello plugins/winrs-cert registry; do
  if [ -f "$ROOT/$mod/go.mod" ]; then
    out="$ROOT/$mod/coverage.out"
    ( cd "$ROOT/$mod" && "$GO" test -coverprofile="coverage.out" ./... >/dev/null 2>&1 ) || true
    [ -f "$out" ] && COVERPATHS="${COVERPATHS:+$COVERPATHS,}$mod/coverage.out"
  fi
done
echo "    coverage: ${COVERPATHS:-<none>}"

echo "==> Scan"
"$SCANNER" \
  "-Dsonar.host.url=$SONAR_HOST_URL" \
  "-Dsonar.token=$SONAR_TOKEN" \
  "-Dsonar.scanner.skipJreProvisioning=true" \
  ${COVERPATHS:+-Dsonar.go.coverage.reportPaths=$COVERPATHS}

echo "==> Xong. Dashboard: $SONAR_HOST_URL/dashboard?id=openitms-smb"
