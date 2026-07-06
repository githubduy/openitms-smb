---
level: L2
status: draft
owners: [maintainer]
updated: 2026-07-06
related-code: [installer/linux/]
---

# Spec: Installer Linux (P2-01/02)

> status: **draft** — script viết xong + syntax-check; approve sau khi E2E VM offline pass (P2-02 AC).

## Layout bundle (tar.gz → giải nén → `sudo ./install.sh`)
```
openitms-smb-<ver>-linux-amd64/
├── install.sh / uninstall.sh / my.cnf / systemd/
├── bin/            # core binary + pwsh/
├── mariadb/        # MariaDB LTS binary tarball (nguyên cây basedir)
├── plugins/        # winrs-cert, hardening (kèm plugin.yaml + checksum)
├── templates/      # playbook preload
├── certs/          # RỖNG — người dùng ném .pfx/.pem vào
├── config/         # sinh lúc install
└── licenses/       # MIT gốc + GPL MariaDB + THIRD_PARTY_LICENSES.md
```

## Hành vi install.sh (idempotent — chạy lại không phá dữ liệu)
1. Copy bundle → `/opt/openitms` (đổi qua env `OPENITMS_PREFIX`).
2. User hệ thống `openitms-db` (nologin) own datadir.
3. MariaDB: `mariadb-install-db` CHỈ khi datadir chưa có; `my.cnf` **skip-networking**
   (socket-only — ADR-0002); root auth = unix socket.
4. Config lần đầu: `config/config.json` hardcode (mysql qua socket, port 3000, keys random);
   DB password random lưu `.db-pass` (600). Đã có config → giữ nguyên.
5. Systemd: `openitms-db.service` → `openitms.service` (After/Requires; env
   `QUICKWIN_PLUGINS_DIR` trỏ plugins bundle). Hardening cơ bản (NoNewPrivileges, PrivateTmp).
6. Admin mặc định `admin/quickwin123` (yêu cầu gốc; UI ép đổi lần đầu — patch 0003/P1-08).
   In banner URL + hướng dẫn certs.

## uninstall.sh
Mặc định giữ data/certs/config; `--purge` hỏi xác nhận rồi mới xóa sạch.

## Còn thiếu để đóng P2-01/02 (AC)
- [ ] `fetch-mariadb.sh` — tải + pin checksum MariaDB LTS tarball vào bundle (build time).
- [ ] Tính `innodb_buffer_pool_size` theo %RAM thật lúc install.
- [ ] `installer/package.sh` — lắp bundle từ dist/ + upstream FULL_UI build.
- [ ] E2E VM Debian + RHEL sạch không internet: cài < 5 phút, login OK, chạy lại không mất
      dữ liệu, reboot 2 service tự lên đúng thứ tự (tests/e2e/installer/).
