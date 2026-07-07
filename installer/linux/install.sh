#!/usr/bin/env bash
# OpenITMS-SMB — cài đặt 1 lệnh trên Linux (chạy từ thư mục giải nén tar.gz).
# Idempotent: chạy lại KHÔNG phá dữ liệu. Không cần internet.
# Layout bundle: bin/ mariadb/ plugins/ templates/ certs/ config/ licenses/ install.sh
set -euo pipefail

PREFIX="${OPENITMS_PREFIX:-/opt/openitms}"
DATA="$PREFIX/data"
BIN_NAME="semaphore"          # đổi 1 dòng khi rename binary (ADR-0004 #1)
SVC_APP="openitms"
SVC_DB="openitms-db"
DB_USER="openitms-db"
ADMIN_LOGIN="admin"
ADMIN_PASS="quickwin123"      # yêu cầu gốc — banner UI ép đổi ở lần login đầu
# Mật khẩu MariaDB mặc định — ĐỦ AN TOÀN (dài, mixed-case + số + ký hiệu), KNOWN default
# tiện quản trị. DB chỉ listen socket/localhost (không expose). Override qua env
# OPENITMS_DB_PASSWORD. Hardening plugin cảnh báo nếu còn default.
DB_PASS_DEFAULT="${OPENITMS_DB_PASSWORD:-OpenITMS@MariaDB#2026}"
HERE="$(cd "$(dirname "$0")" && pwd)"

[ "$(id -u)" = 0 ] || { echo "ERROR: cần chạy bằng root (sudo ./install.sh)"; exit 1; }
command -v systemctl >/dev/null || { echo "ERROR: cần systemd"; exit 1; }

echo "==> [1/6] Copy bundle vào $PREFIX"
mkdir -p "$PREFIX" "$DATA/db" "$DATA/tmp" "$PREFIX/certs"
for d in bin mariadb plugins templates licenses config; do
  [ -d "$HERE/$d" ] && cp -a "$HERE/$d" "$PREFIX/"
done

echo "==> [2/6] User hệ thống"
id -u "$DB_USER" >/dev/null 2>&1 || useradd -r -s /usr/sbin/nologin -d "$DATA" "$DB_USER"
chown -R "$DB_USER:$DB_USER" "$DATA"
chmod 750 "$PREFIX/certs"

echo "==> [3/6] MariaDB (bundled, socket-only)"
SOCK="$DATA/db/mysql.sock"
if [ ! -d "$DATA/db/mysql" ]; then
  "$PREFIX/mariadb/scripts/mariadb-install-db" \
    --basedir="$PREFIX/mariadb" --datadir="$DATA/db" --user="$DB_USER" \
    --auth-root-authentication-method=socket >/dev/null
  echo "    datadir khởi tạo xong"
else
  echo "    datadir đã có — giữ nguyên (idempotent)"
fi
install -m 640 -o "$DB_USER" "$HERE/my.cnf" "$PREFIX/my.cnf"
sed -i "s|@PREFIX@|$PREFIX|g; s|@DATA@|$DATA|g; s|@DBUSER@|$DB_USER|g" "$PREFIX/my.cnf"

echo "==> [4/6] Config lần đầu (hardcode — sửa sau qua Settings UI)"
DB_PASS_FILE="$PREFIX/.db-pass"
if [ ! -f "$DB_PASS_FILE" ]; then
  printf '%s' "$DB_PASS_DEFAULT" > "$DB_PASS_FILE"   # default đủ an toàn, known (đổi qua env)
  chmod 600 "$DB_PASS_FILE"
fi
DB_PASS="$(cat "$DB_PASS_FILE")"
if [ ! -f "$PREFIX/config/config.json" ]; then
  mkdir -p "$PREFIX/config"
  rk(){ head -c32 /dev/urandom | base64 | tr -d '=+/' | head -c32; }
  cat > "$PREFIX/config/config.json" <<EOF
{
  "dialect": "mysql",
  "mysql": { "host": "$SOCK", "user": "openitms", "pass": "$DB_PASS", "name": "openitms" },
  "port": "3000",
  "tmp_path": "$DATA/tmp",
  "cookie_hash": "$(rk)",
  "cookie_encryption": "$(rk)",
  "access_key_encryption": "$(rk)"
}
EOF
  chmod 600 "$PREFIX/config/config.json"
fi

echo "==> [5/6] Systemd units"
for u in "$SVC_DB" "$SVC_APP"; do
  install -m 644 "$HERE/systemd/$u.service" "/etc/systemd/system/$u.service"
  sed -i "s|@PREFIX@|$PREFIX|g; s|@DATA@|$DATA|g; s|@DBUSER@|$DB_USER|g; s|@BIN@|$BIN_NAME|g" \
    "/etc/systemd/system/$u.service"
done
systemctl daemon-reload
systemctl enable --now "$SVC_DB"

echo "    chờ DB socket..."
for i in $(seq 1 30); do [ -S "$SOCK" ] && break; sleep 1; done
[ -S "$SOCK" ] || { echo "ERROR: MariaDB không lên — journalctl -u $SVC_DB"; exit 1; }

echo "    tạo database + user app (idempotent)"
"$PREFIX/mariadb/bin/mariadb" --socket="$SOCK" -u root <<EOF
CREATE DATABASE IF NOT EXISTS openitms CHARACTER SET utf8mb4;
CREATE USER IF NOT EXISTS 'openitms'@'localhost' IDENTIFIED BY '$DB_PASS';
ALTER USER 'openitms'@'localhost' IDENTIFIED BY '$DB_PASS';
GRANT ALL PRIVILEGES ON openitms.* TO 'openitms'@'localhost';
FLUSH PRIVILEGES;
EOF

echo "==> [6/6] Admin mặc định + start app"
"$PREFIX/bin/$BIN_NAME" user add --admin --login "$ADMIN_LOGIN" --name Admin \
  --email admin@localhost --password "$ADMIN_PASS" --config "$PREFIX/config/config.json" \
  2>/dev/null || echo "    admin đã tồn tại — bỏ qua (idempotent)"
systemctl enable --now "$SVC_APP"

IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
cat <<EOF

╔══════════════════════════════════════════════════════════════╗
  OpenITMS-SMB cài đặt XONG.
  URL:        http://${IP:-127.0.0.1}:3000
  Đăng nhập:  $ADMIN_LOGIN / $ADMIN_PASS   (ĐỔI NGAY ở lần login đầu)
  Certs:      ném .pfx/.pem vào $PREFIX/certs là dùng được ngay
  Service:    systemctl status $SVC_APP $SVC_DB
╚══════════════════════════════════════════════════════════════╝
EOF
