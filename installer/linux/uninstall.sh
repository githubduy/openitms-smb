#!/usr/bin/env bash
# Gỡ OpenITMS-SMB. Mặc định GIỮ dữ liệu (data + certs + config); --purge mới xóa.
set -euo pipefail
PREFIX="${OPENITMS_PREFIX:-/opt/openitms}"
[ "$(id -u)" = 0 ] || { echo "ERROR: cần root"; exit 1; }

systemctl disable --now openitms openitms-db 2>/dev/null || true
rm -f /etc/systemd/system/openitms.service /etc/systemd/system/openitms-db.service
systemctl daemon-reload

if [ "${1:-}" = "--purge" ]; then
  read -r -p "XÓA TOÀN BỘ dữ liệu tại $PREFIX? (gõ 'yes'): " ans
  [ "$ans" = "yes" ] && rm -rf "$PREFIX" && echo "Đã purge." || echo "Bỏ qua purge."
else
  for d in bin mariadb plugins templates licenses; do rm -rf "${PREFIX:?}/$d"; done
  echo "Đã gỡ binary. Dữ liệu giữ tại $PREFIX (data/, certs/, config/) — dùng --purge để xóa."
fi
