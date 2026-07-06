#!/usr/bin/env bash
# export-patch.sh <NNNN-ten-patch> — export thay đổi MỚI (so với các patch trước đã apply)
# thành core-patches/<NNNN-ten-patch>.patch với header chuẩn (WHY bắt buộc).
#
# QUY TRÌNH ĐÚNG:
#   1. scripts/reset-upstream.sh
#   2. scripts/apply-patches.sh            # apply + STAGE các patch trước (baseline)
#   3. sửa code / thêm file trong upstream/ cho patch MỚI (KHÔNG git add)
#   4. scripts/export-patch.sh <NNNN-ten>  # diff working-tree vs baseline = chỉ thay đổi mới
#   5. điền WHY/WHAT; thêm series + CHANGELOG + spec L2 (bộ-4)
#
# go.sum LOẠI khỏi patch (regenerate bằng go mod tidy khi apply). File mới hiện qua `add -N`.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
UP="$ROOT/upstream"
NAME="${1:?Usage: export-patch.sh <NNNN-ten-patch> (vd: 0005-registry-client)}"
OUT="$ROOT/core-patches/$NAME.patch"

# intent-to-add file mới (untracked) để chúng xuất hiện trong git diff (không stage nội dung)
git -C "$UP" add -N . >/dev/null 2>&1 || true

if git -C "$UP" diff --quiet -- . ':(exclude)go.sum'; then
  echo "ERROR: không có thay đổi mới trong upstream/ (đã apply-patches + sửa code chưa?)." >&2
  exit 1
fi

{
  echo "# PATCH: $NAME"
  echo "# WHY: <bắt buộc điền — AI dùng dòng này để rebase khi upstream đổi>"
  echo "# WHAT: <tóm tắt thay đổi mức cao>"
  echo "# SPEC: docs/L2-specs/core-patches/$NAME.md"
  echo "# BASELINE: $(git -C "$UP" describe --tags --always)"
  echo "#"
  # diff working-tree vs INDEX (= baseline các patch trước đã staged), loại go.sum
  git -C "$UP" diff -- . ':(exclude)go.sum'
} > "$OUT"

echo "OK: $OUT"
echo "TIẾP THEO (bộ-4): điền WHY/WHAT header; thêm '$NAME.patch' vào core-patches/series;"
echo "entry vào core-patches/CHANGELOG.md; viết spec docs/L2-specs/core-patches/$NAME.md"
