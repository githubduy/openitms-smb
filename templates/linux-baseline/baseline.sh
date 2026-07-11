#!/usr/bin/env bash
# linux-baseline — cập nhật hệ thống + gói thiết yếu + unattended-upgrades (Debian/Ubuntu).
# Input inject qua env (UPPER_SNAKE) bởi OpenITMS task runner. Idempotent.
set -euo pipefail

EXTRA_PACKAGES="${EXTRA_PACKAGES:-curl vim htop}"

if ! command -v apt-get >/dev/null 2>&1; then
  echo "template này dành cho Debian/Ubuntu (apt)." >&2
  exit 1
fi

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  SUDO="sudo -n"
fi

export DEBIAN_FRONTEND=noninteractive
$SUDO apt-get update -y
$SUDO apt-get upgrade -y

# shellcheck disable=SC2086  # EXTRA_PACKAGES cố ý tách từ thành nhiều gói
$SUDO apt-get install -y ca-certificates unattended-upgrades $EXTRA_PACKAGES

$SUDO dpkg-reconfigure -f noninteractive unattended-upgrades || true

echo "OK: đã cập nhật + bật unattended-upgrades trên $(hostname)."
