#!/usr/bin/env bash
# plugin-through-core.sh — E2E patch 0001: core (đã patch) load plugin hello,
# gọi API động QUA auth middleware của Semaphore bằng session thật.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BIN="$ROOT/dist/bin/semaphore"
[ -x "$BIN" ] || BIN="$ROOT/dist/bin/semaphore.exe"
[ -x "$BIN" ] || { echo "ERROR: chưa có binary — scripts/build-all.sh trước" >&2; exit 1; }
GO="$ROOT/Go/go/bin/go"; [ -x "$GO" ] || GO="$ROOT/Go/go/bin/go.exe"

export NO_PROXY="127.0.0.1,localhost" no_proxy="127.0.0.1,localhost"
CURL="curl -m 5 -s"
PORT="${E2E_PORT:-3998}"
WORK="$(mktemp -d)"
WORKW="$WORK"; command -v cygpath >/dev/null && WORKW="$(cygpath -m "$WORK")"
trap 'kill $SRV_PID 2>/dev/null || true; wait $SRV_PID 2>/dev/null || true; sleep 1; rm -rf "$WORK" 2>/dev/null || true' EXIT

echo "==> [1/5] build plugin hello vào $WORK/plugins/hello"
mkdir -p "$WORK/plugins/hello"
EXT=""; case "$(uname -s)" in MINGW*|MSYS*|CYGWIN*) EXT=".exe";; esac
(cd "$ROOT/plugins/hello" && "$GO" build -o "$WORK/plugins/hello/hello$EXT" .)
cp "$ROOT/plugins/hello/plugin.yaml" "$WORK/plugins/hello/"

cat > "$WORK/config.json" <<EOF
{
  "dialect": "bolt",
  "bolt": { "host": "$WORKW/database.boltdb" },
  "port": "$PORT",
  "tmp_path": "$WORKW/tmp",
  "cookie_hash": "$(head -c32 /dev/urandom | base64 | tr -d '=+/' | head -c32)",
  "cookie_encryption": "$(head -c32 /dev/urandom | base64 | tr -d '=+/' | head -c32)",
  "access_key_encryption": "$(head -c32 /dev/urandom | base64 | tr -d '=+/' | head -c32)"
}
EOF

echo "==> [2/5] tạo admin user"
"$BIN" user add --admin --login admin --name Admin --email admin@localhost --password quickwin123 --config "$WORK/config.json" >/dev/null

echo "==> [3/5] start server (plugins dir = $WORK/plugins)"
QUICKWIN_PLUGINS_DIR="$WORKW/plugins" "$BIN" server --config "$WORK/config.json" > "$WORK/server.log" 2>&1 &
SRV_PID=$!
for i in $(seq 1 30); do
  $CURL -f "http://127.0.0.1:$PORT/api/ping" >/dev/null 2>&1 && break
  kill -0 $SRV_PID 2>/dev/null || { echo "ERROR: server chết sớm:"; tail -20 "$WORK/server.log"; exit 1; }
  sleep 1
done

echo "==> [4/5] chưa login gọi plugin API phải bị chặn"
code=$($CURL -o /dev/null -w '%{http_code}' "http://127.0.0.1:$PORT/api/plugins/hello/info")
[ "$code" = "401" ] || [ "$code" = "403" ] || { echo "ERROR: unauthenticated phải 401/403, got $code"; exit 1; }
echo "OK: unauthenticated → $code"

echo "==> [5/5] login + gọi API động qua session thật"
$CURL -c "$WORK/cookies" -H 'Content-Type: application/json' \
  -d '{"auth":"admin","password":"quickwin123"}' \
  "http://127.0.0.1:$PORT/api/auth/login" > /dev/null
body=$($CURL -b "$WORK/cookies" "http://127.0.0.1:$PORT/api/plugins/hello/info")
echo "$body" | grep -q '"name":"hello"' || { echo "ERROR: /api/plugins/hello/info trả: $body"; tail -20 "$WORK/server.log"; exit 1; }
echo "OK: info → $body"

echo_body=$($CURL -b "$WORK/cookies" -X POST -d 'xin chao core' "http://127.0.0.1:$PORT/api/plugins/hello/echo")
echo "$echo_body" | grep -q 'xin chao core' && echo "$echo_body" | grep -q '"caller":"admin"' \
  || { echo "ERROR: echo trả: $echo_body"; exit 1; }
echo "OK: echo → $echo_body"

grep -q "plugin manager mounted" "$WORK/server.log" || { echo "ERROR: log không thấy plugin manager mounted"; exit 1; }
echo "PLUGIN-THROUGH-CORE PASS"
