#!/usr/bin/env bash
# smoke.sh — tầng-2: chạy binary thật với BoltDB tạm, assert server sống.
# Dùng được trên máy dev (Windows git-bash / Linux) lẫn CI. Không cần MariaDB/UI assets.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BIN="$ROOT/dist/bin/semaphore"
[ -x "$BIN" ] || BIN="$ROOT/dist/bin/semaphore.exe"
[ -x "$BIN" ] || { echo "ERROR: chưa có binary — chạy scripts/build-all.sh trước" >&2; exit 1; }

PORT="${SMOKE_PORT:-3999}"
WORK="$(mktemp -d)"
trap 'kill $SRV_PID 2>/dev/null || true; wait $SRV_PID 2>/dev/null || true; sleep 1; rm -rf "$WORK" 2>/dev/null || true' EXIT
# môi trường công ty hay đặt http_proxy — không được proxy request localhost
export NO_PROXY="127.0.0.1,localhost" no_proxy="127.0.0.1,localhost"
CURL="curl -m 5"
# binary Windows không hiểu path POSIX của git-bash → đổi sang dạng D:/...
WORKW="$WORK"; command -v cygpath >/dev/null && WORKW="$(cygpath -m "$WORK")"

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

echo "==> start server (bolt, port $PORT)"
"$BIN" server --config "$WORK/config.json" > "$WORK/server.log" 2>&1 &
SRV_PID=$!

ok=0
for i in $(seq 1 30); do
  if $CURL -fsS "http://127.0.0.1:$PORT/api/ping" 2>/dev/null | grep -qi pong; then ok=1; break; fi
  kill -0 $SRV_PID 2>/dev/null || { echo "ERROR: server chết sớm — log:"; tail -20 "$WORK/server.log"; exit 1; }
  sleep 1
done
[ $ok = 1 ] || { echo "ERROR: /api/ping không trả pong sau 30s — log:"; tail -20 "$WORK/server.log"; exit 1; }
echo "OK: /api/ping → pong"

code=$($CURL -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:$PORT/")
[ "$code" = "200" ] || { echo "ERROR: GET / trả $code (mong 200)"; exit 1; }
echo "OK: GET / → 200 (UI serve được)"
echo "SMOKE PASS"
