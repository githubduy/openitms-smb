#!/usr/bin/env bash
# docker-cluster-setup — cài Docker Engine + compose lên host Linux đích.
# LƯU Ý: Docker được cài LÊN MÁY CLIENT theo yêu cầu người dùng — bản thân OpenITMS-SMB
# KHÔNG dùng Docker (nguyên tắc đóng gói: binary native).
set -euo pipefail
if command -v docker >/dev/null && docker compose version >/dev/null 2>&1; then
  echo "OK: Docker + compose đã có ($(docker --version))"; exit 0
fi
. /etc/os-release
case "$ID" in
  ubuntu|debian)
    curl -fsSL https://get.docker.com | sh
    ;;
  rhel|centos|rocky|almalinux|fedora)
    dnf -y install dnf-plugins-core
    dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    dnf -y install docker-ce docker-ce-cli containerd.io docker-compose-plugin
    systemctl enable --now docker
    ;;
  *) echo "ERROR: distro $ID chưa hỗ trợ tự động — cài Docker thủ công."; exit 1 ;;
esac
docker --version && docker compose version
echo "OK: Docker sẵn sàng trên $(hostname)"
