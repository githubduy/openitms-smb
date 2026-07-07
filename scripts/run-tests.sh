#!/usr/bin/env bash
# run-tests.sh — chạy test các package CỦA TA (plugin-manager, plugins, registry).
# Test upstream thuộc trách nhiệm upstream; ta chỉ chạy khi sync (SYNC=1).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GO="$ROOT/Go/go/bin/go"
[ -x "$GO" ] || GO="$ROOT/Go/go/bin/go.exe"
[ -x "$GO" ] || { echo "ERROR: chưa có toolchain — chạy scripts/setup-toolchain.sh" >&2; exit 1; }

for mod in plugin-manager sdk/go registry gitea-manager service-manager plugins/*/; do
  mod="${mod%/}"
  if [ -f "$ROOT/$mod/go.mod" ]; then
    echo "==> test $mod"
    (cd "$ROOT/$mod" && "$GO" test ./...)
  fi
done

if [ "${SYNC:-0}" = "1" ]; then
  echo "==> (sync mode) test upstream sau khi apply patch"
  (cd "$ROOT/upstream" && "$GO" test ./... ) || { echo "WARN: upstream test fail — xem có phải do patch không." >&2; exit 1; }
fi
echo "OK: tests pass"
