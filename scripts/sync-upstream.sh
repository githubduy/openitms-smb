#!/usr/bin/env bash
# sync-upstream.sh <tag> — quy trình đồng bộ khi Semaphore ra release mới (plan mục 4.2).
# 1. fetch + checkout tag mới trong upstream/   2. apply toàn bộ series
# 3. build + test                                4. FAIL → dừng, in patch conflict
# Sau khi PASS: commit gitlink submodule mới + mở PR (người/AI skill sync-upstream làm).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
UP="$ROOT/upstream"
TAG="${1:?Usage: sync-upstream.sh <upstream-tag> (vd: v2.18.17)}"

echo "==> [1/4] fetch tag $TAG"
git -C "$UP" fetch --depth 1 origin tag "$TAG"
"$ROOT/scripts/reset-upstream.sh"
git -C "$UP" checkout -q "$TAG"
echo "    upstream @ $(git -C "$UP" describe --tags)"

echo "==> [2/4] apply core-patches/series"
"$ROOT/scripts/apply-patches.sh"

echo "==> [3/4] build"
"$ROOT/scripts/build-all.sh"

echo "==> [4/4] test"
"$ROOT/scripts/run-tests.sh" || { echo "ERROR: test fail sau sync — xem log trên." >&2; exit 1; }

echo "OK: sync $TAG thành công. Tiếp: commit gitlink upstream mới + PR (liệt kê patch giữ nguyên/rebase)."
