#!/usr/bin/env bash
# check-banned-words.sh — chặn commit/push chứa từ khóa nhận diện cá nhân/công ty.
# DANH SÁCH TỪ CẤM KHÔNG NẰM TRONG REPO (bản thân nó là thông tin nhạy cảm):
#   1. env BANNED_WORDS_REGEX (CI: GitHub secret)
#   2. file .git/banned-words   (local, không được track)
#   3. file ~/.quickwin-banned-words
# Không cấu hình → skip kèm cảnh báo (không fail CI của contributor ngoài).
#
# Mode:
#   --staged  : quét diff đang staged + identity (dùng cho pre-commit hook)
#   --all     : quét toàn bộ tree HEAD (trừ upstream/) + toàn bộ commit message (CI)
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MODE="${1:---staged}"

REGEX="${BANNED_WORDS_REGEX:-}"
[ -z "$REGEX" ] && [ -f "$ROOT/.git/banned-words" ] && REGEX="$(head -1 "$ROOT/.git/banned-words")"
[ -z "$REGEX" ] && [ -f "$HOME/.quickwin-banned-words" ] && REGEX="$(head -1 "$HOME/.quickwin-banned-words")"
if [ -z "$REGEX" ]; then
  echo "WARN: banned-words chưa cấu hình (BANNED_WORDS_REGEX / .git/banned-words) — bỏ qua check." >&2
  exit 0
fi

fail=0

# identity đang cấu hình (CI runner có thể chưa set — không được làm chết script)
ident="$( (git -C "$ROOT" config user.name || true; git -C "$ROOT" config user.email || true) )"
if echo "$ident" | grep -qiE "$REGEX"; then
  echo "ERROR: git user.name/user.email chứa từ cấm — sửa: git config user.name/user.email (repo-local)" >&2
  fail=1
fi

if [ "$MODE" = "--staged" ]; then
  if git -C "$ROOT" diff --cached | grep -qiE "$REGEX"; then
    echo "ERROR: nội dung STAGED chứa từ cấm:" >&2
    git -C "$ROOT" diff --cached | grep -inE "$REGEX" | head -5 >&2
    fail=1
  fi
else # --all
  if git -C "$ROOT" grep -ilE "$REGEX" HEAD -- ':(exclude)upstream' >/dev/null 2>&1; then
    echo "ERROR: file trong tree chứa từ cấm:" >&2
    git -C "$ROOT" grep -ilE "$REGEX" HEAD -- ':(exclude)upstream' | head >&2
    fail=1
  fi
  if git -C "$ROOT" log --all --pretty=full | grep -qiE "$REGEX"; then
    echo "ERROR: commit message/author trong lịch sử chứa từ cấm (cần rewrite trước khi push)." >&2
    fail=1
  fi
fi

[ $fail = 0 ] && echo "OK: banned-words sạch ($MODE)"
exit $fail
