#!/usr/bin/env bash
# build-all.sh — build core (upstream + patch đã apply) bằng toolchain trong <repo>/Go/.
# Output: <repo>/dist/bin/semaphore(.exe)
# Lưu ý: bản build đầy đủ UI cần web assets (task build trong upstream cần node/npm).
#   - CI/dev đủ tooling: chạy với FULL_UI=1 để build cả frontend qua Taskfile upstream.
#   - Mặc định: build backend Go (đủ cho kiểm tra patch chain + --version).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
UP="$ROOT/upstream"
GO="$ROOT/Go/go/bin/go"
[ -x "$GO" ] || GO="$ROOT/Go/go/bin/go.exe"
[ -x "$GO" ] || { echo "ERROR: chưa có toolchain — chạy scripts/setup-toolchain.sh" >&2; exit 1; }

OUT="$ROOT/dist/bin"
mkdir -p "$OUT"
BIN="semaphore"
case "$(uname -s)" in MINGW*|MSYS*|CYGWIN*) BIN="semaphore.exe";; esac

if [ "${FULL_UI:-0}" = "1" ]; then
  command -v task >/dev/null || { echo "ERROR: FULL_UI=1 cần go-task (taskfile.dev)" >&2; exit 1; }
  (cd "$UP" && PATH="$(dirname "$GO"):$PATH" task build)
  cp "$UP/bin/semaphore"* "$OUT/" 2>/dev/null || true
else
  echo "==> go build backend (không UI — dùng FULL_UI=1 cho bản đầy đủ)"
  (cd "$UP" && "$GO" build -o "$OUT/$BIN" ./cli)
fi

echo "OK: $OUT/$BIN"
"$OUT/$BIN" version 2>/dev/null || "$OUT/$BIN" --version 2>/dev/null || true
