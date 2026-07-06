#!/usr/bin/env bash
# gen-proto.sh — regenerate Go stubs từ proto/ (nguồn chân lý duy nhất, cấm sửa gen/).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="$ROOT/Go/bin:$ROOT/Go/go/bin:$PATH"
command -v buf >/dev/null || { echo "ERROR: thiếu buf — GOBIN=$ROOT/Go/bin go install github.com/bufbuild/buf/cmd/buf@v1.50.0" >&2; exit 1; }
cd "$ROOT/proto"
buf lint
buf generate
echo "OK: stubs → proto/gen/go"
