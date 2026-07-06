#!/usr/bin/env bash
# apply-patches.sh — apply toàn bộ core-patches/series vào cây upstream/ (working tree).
# KHÔNG commit vào submodule; dùng reset-upstream.sh để về trạng thái sạch.
# Exit != 0 ngay khi 1 patch fail, in rõ patch + hunk lỗi.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
UP="$ROOT/upstream"
SERIES="$ROOT/core-patches/series"

if ! git -C "$UP" diff --quiet || ! git -C "$UP" diff --cached --quiet; then
  echo "ERROR: upstream/ không sạch. Chạy scripts/reset-upstream.sh trước." >&2
  exit 1
fi

applied=0
while IFS= read -r line; do
  patch="$(echo "$line" | sed 's/#.*//' | xargs)"
  [ -z "$patch" ] && continue
  file="$ROOT/core-patches/$patch"
  [ -f "$file" ] || { echo "ERROR: không thấy $file (khai trong series)" >&2; exit 1; }
  echo "==> apply $patch"
  if ! git -C "$UP" apply --index --verbose "$file"; then
    echo "ERROR: patch FAIL: $patch — đọc header VÌ SAO trong patch + spec docs/L2-specs/core-patches/ để rebase." >&2
    exit 1
  fi
  applied=$((applied+1))
done < "$SERIES"

# go.sum không nằm trong patch — regenerate nếu go.mod bị patch đổi (deps mới cần sum).
# so với HEAD (patch được stage nên phải dùng HEAD, không phải unstaged diff).
if ! git -C "$UP" diff HEAD --quiet -- go.mod 2>/dev/null; then
  GO="$ROOT/Go/go/bin/go"; [ -x "$GO" ] || GO="$ROOT/Go/go/bin/go.exe"; [ -x "$GO" ] || GO=go
  echo "==> go.mod đã đổi → go mod tidy (regenerate go.sum)"
  ( cd "$UP" && "$GO" mod tidy ) || { echo "ERROR: go mod tidy fail sau apply" >&2; exit 1; }
fi

echo "OK: đã apply $applied patch vào upstream/ (working tree, chưa commit)."
