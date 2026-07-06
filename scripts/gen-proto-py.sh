#!/usr/bin/env bash
# gen-proto-py.sh — sinh Python stub từ proto/ (CÙNG nguồn chân lý với Go).
# Cần: python3 + grpcio-tools (pip install grpcio-tools). Output: proto/gen/python/
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/proto/gen/python"
mkdir -p "$OUT"
python3 -m grpc_tools.protoc \
  -I "$ROOT/proto" \
  --python_out="$OUT" \
  --grpc_python_out="$OUT" \
  "$ROOT/proto/quickwin/plugin/v1/plugin.proto"
# tạo __init__.py để import package quickwin.plugin.v1
find "$OUT/quickwin" -type d -exec sh -c 'touch "$1/__init__.py"' _ {} \;
echo "OK: Python stubs → $OUT"
