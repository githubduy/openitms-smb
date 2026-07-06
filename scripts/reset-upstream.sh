#!/usr/bin/env bash
# reset-upstream.sh — đưa upstream/ về đúng commit pin (xóa mọi patch đã apply + file lạ).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
UP="$ROOT/upstream"
git -C "$UP" reset --hard HEAD
git -C "$UP" clean -fd
echo "OK: upstream/ sạch tại $(git -C "$UP" describe --tags --always)"
