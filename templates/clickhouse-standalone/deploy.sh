#!/usr/bin/env bash
set -euo pipefail
DEFAULT_PASSWORD="${DEFAULT_PASSWORD:?cần DEFAULT_PASSWORD}"
DIR="${DEPLOY_DIR:-/opt/clickhouse}"; mkdir -p "$DIR"; cd "$DIR"
cat > docker-compose.yml <<EOF
services:
  clickhouse:
    image: clickhouse/clickhouse-server:24.8
    environment: { CLICKHOUSE_PASSWORD: "${DEFAULT_PASSWORD}" }
    ports: ["8123:8123","9000:9000"]
    ulimits: { nofile: { soft: 262144, hard: 262144 } }
    volumes: [ch-data:/var/lib/clickhouse]
volumes: { ch-data: {} }
EOF
docker compose up -d
echo "OK: ClickHouse :8123 (HTTP) :9000 (native) trên $(hostname)"
