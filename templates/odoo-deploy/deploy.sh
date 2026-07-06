#!/usr/bin/env bash
# odoo-deploy — Odoo + Postgres qua docker compose lên host đích (yêu cầu Docker sẵn).
set -euo pipefail
ODOO_VERSION="${ODOO_VERSION:-17}"; ADMIN_PASSWORD="${ADMIN_PASSWORD:?cần ADMIN_PASSWORD}"
DIR="${DEPLOY_DIR:-/opt/odoo}"; mkdir -p "$DIR"; cd "$DIR"
cat > docker-compose.yml <<EOF
services:
  db:
    image: postgres:16
    environment: { POSTGRES_DB: postgres, POSTGRES_USER: odoo, POSTGRES_PASSWORD: odoo }
    volumes: [odoo-db:/var/lib/postgresql/data]
  odoo:
    image: odoo:${ODOO_VERSION}
    depends_on: [db]
    ports: ["8069:8069"]
    environment: { HOST: db, USER: odoo, PASSWORD: odoo }
    command: ["odoo","--db_host=db","-r","odoo","-w","odoo","--admin-passwd=${ADMIN_PASSWORD}"]
    volumes: [odoo-web:/var/lib/odoo]
volumes: { odoo-db: {}, odoo-web: {} }
EOF
docker compose up -d
echo "OK: Odoo ${ODOO_VERSION} → http://$(hostname -I | awk '{print $1}'):8069"
