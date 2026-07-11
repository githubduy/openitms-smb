#!/usr/bin/env bash
# postgres-standalone — PostgreSQL qua docker compose trên host đích.
# Input inject qua env bởi OpenITMS task runner. Docker cài LÊN máy đích (không dùng của OpenITMS).
set -euo pipefail

VERSION="${VERSION:-16}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:?cần POSTGRES_PASSWORD}"
DB_NAME="${DB_NAME:-appdb}"
DIR="${DEPLOY_DIR:-/opt/postgres}"

mkdir -p "$DIR"
cd "$DIR"
cat > docker-compose.yml <<EOF
services:
  postgres:
    image: postgres:${VERSION}
    environment:
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD}"
      POSTGRES_DB: "${DB_NAME}"
    ports: ["5432:5432"]
    volumes: [postgres-data:/var/lib/postgresql/data]
    restart: unless-stopped
volumes: { postgres-data: {} }
EOF

docker compose up -d
echo "OK: PostgreSQL ${VERSION} :5432 (db=${DB_NAME}) trên $(hostname)."
