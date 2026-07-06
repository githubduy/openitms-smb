#!/usr/bin/env bash
# export-patch.sh <NNNN-ten-patch> — export thay đổi ĐANG STAGED trong upstream/
# thành core-patches/<NNNN-ten-patch>.patch với header chuẩn (WHY bắt buộc).
# Quy trình: sửa code trong upstream/ → git -C upstream add -A → chạy script này
# → điền WHY trong file patch → thêm vào series + CHANGELOG + spec L2 (bộ-4).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
UP="$ROOT/upstream"
NAME="${1:?Usage: export-patch.sh <NNNN-ten-patch> (vd: 0001-plugin-manager-hook)}"
OUT="$ROOT/core-patches/$NAME.patch"

if git -C "$UP" diff --cached --quiet; then
  echo "ERROR: không có gì staged trong upstream/. git -C upstream add -A trước." >&2
  exit 1
fi

{
  echo "# PATCH: $NAME"
  echo "# WHY: <bắt buộc điền — AI dùng dòng này để rebase khi upstream đổi>"
  echo "# WHAT: <tóm tắt thay đổi mức cao>"
  echo "# SPEC: docs/L2-specs/core-patches/$NAME.md"
  echo "# BASELINE: $(git -C "$UP" describe --tags --always)"
  echo "#"
  git -C "$UP" diff --cached
} > "$OUT"

echo "OK: $OUT"
echo "TIẾP THEO (bộ-4): điền WHY/WHAT header; thêm '$NAME.patch' vào core-patches/series;"
echo "entry vào core-patches/CHANGELOG.md; viết spec docs/L2-specs/core-patches/$NAME.md"
