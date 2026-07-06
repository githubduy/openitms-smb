#!/usr/bin/env bash
set -euo pipefail
VERSION="${VERSION:-11.4}"; ROOT_PASSWORD="${ROOT_PASSWORD:?cần ROOT_PASSWORD}"
DIR="${DEPLOY_DIR:-/opt/mariadb}"; mkdir -p "$DIR"; cd "$DIR"
cat > docker-compose.yml <<EOF
services:
  mariadb:
    image: mariadb:${VERSION}
    environment: { MARIADB_ROOT_PASSWORD: "${ROOT_PASSWORD}" }
    ports: ["3306:3306"]
    volumes: [mariadb-data:/var/lib/mysql]
volumes: { mariadb-data: {} }
EOF
docker compose up -d
echo "OK: MariaDB ${VERSION} :3306 trên $(hostname)"
