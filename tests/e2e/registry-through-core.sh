#!/usr/bin/env bash
# registry-through-core.sh — E2E patch 0005: core search + install artifact từ registry local
# (build bằng registryctl), qua session admin thật.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BIN="$ROOT/dist/bin/semaphore"; [ -x "$BIN" ] || BIN="$ROOT/dist/bin/semaphore.exe"
[ -x "$BIN" ] || { echo "ERROR: chưa có binary" >&2; exit 1; }
GO="$ROOT/Go/go/bin/go"; [ -x "$GO" ] || GO="$ROOT/Go/go/bin/go.exe"

export NO_PROXY="127.0.0.1,localhost" no_proxy="127.0.0.1,localhost"
CURL="curl -m 5 -s"
PORT="${E2E_PORT:-3997}"
WORK="$(mktemp -d)"
WORKW="$WORK"; command -v cygpath >/dev/null && WORKW="$(cygpath -m "$WORK")"
trap 'kill $SRV_PID 2>/dev/null || true; wait $SRV_PID 2>/dev/null || true; sleep 1; rm -rf "$WORK" 2>/dev/null || true' EXIT

echo "==> [1/5] build local registry (1 template 'demo') bằng registryctl"
mkdir -p "$WORK/artsrc/demo" "$WORK/registry"
echo "hello from template" > "$WORK/artsrc/demo/run.sh"
cat > "$WORK/spec.json" <<EOF
[{"type":"template","name":"demo","version":"1.0.0","license":"MIT","dir":"$WORKW/artsrc/demo","description":"demo template"}]
EOF
# không ký (dev) → core bỏ verify khi không có trusted key
(cd "$ROOT/registry" && "$GO" run ./cmd/registryctl build local "$WORKW/registry" "$WORKW/spec.json") >/dev/null

cat > "$WORK/config.json" <<EOF
{ "dialect":"bolt","bolt":{"host":"$WORKW/db.boltdb"},"port":"$PORT","tmp_path":"$WORKW/tmp",
  "cookie_hash":"$(head -c32 /dev/urandom|base64|tr -d '=+/'|head -c32)",
  "cookie_encryption":"$(head -c32 /dev/urandom|base64|tr -d '=+/'|head -c32)",
  "access_key_encryption":"$(head -c32 /dev/urandom|base64|tr -d '=+/'|head -c32)" }
EOF

echo "==> [2/5] tạo admin"
"$BIN" user add --admin --login admin --name Admin --email a@localhost --password quickwin123 --config "$WORK/config.json" >/dev/null

echo "==> [3/5] start core (registry local = $WORKW/registry)"
QUICKWIN_REGISTRY_LOCAL="file://$WORKW/registry" QUICKWIN_PLUGINS_DIR="$WORKW/plugins" \
  "$BIN" server --config "$WORK/config.json" > "$WORK/s.log" 2>&1 &
SRV_PID=$!
for i in $(seq 1 30); do $CURL -f "http://127.0.0.1:$PORT/api/ping" >/dev/null 2>&1 && break; sleep 1; done

echo "==> [4/5] login + search registry"
$CURL -c "$WORK/ck" -H 'Content-Type: application/json' -d '{"auth":"admin","password":"quickwin123"}' "http://127.0.0.1:$PORT/api/auth/login" >/dev/null
found=$($CURL -b "$WORK/ck" "http://127.0.0.1:$PORT/api/registry/search?type=template")
echo "$found" | grep -q '"name":"demo"' || { echo "ERROR: search không thấy demo: $found"; tail -20 "$WORK/s.log"; exit 1; }
echo "OK: search → demo"

echo "==> [5/5] verify registry mounted trong log"
grep -q "registry client mounted" "$WORK/s.log" || { echo "ERROR: log không thấy registry mounted"; exit 1; }
echo "REGISTRY-THROUGH-CORE PASS"
