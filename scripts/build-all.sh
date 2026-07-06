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
  # api/router.go dùng go:embed public/* — backend-only cần placeholder assets
  # (chỉ trong working tree upstream; reset-upstream.sh sẽ dọn)
  if [ ! -f "$UP/api/public/index.html" ]; then
    mkdir -p "$UP/api/public"
    echo '<!-- placeholder: backend-only build, no UI assets (FULL_UI=1 for real UI) -->' > "$UP/api/public/index.html"
  fi
  IMPORT="$(head -1 "$UP/go.mod" | awk '{print $2}')"
  UPTAG="$(git -C "$UP" describe --tags --always)"
  QWVER="quickwin-dev-sem${UPTAG#v}"
  (cd "$UP" && "$GO" build -ldflags "-X $IMPORT/util.Ver=$QWVER -X $IMPORT/util.Commit=$(git -C "$UP" rev-parse --short HEAD)" -o "$OUT/$BIN" ./cli)
fi

echo "OK: $OUT/$BIN"
"$OUT/$BIN" version 2>/dev/null || "$OUT/$BIN" --version 2>/dev/null || true
