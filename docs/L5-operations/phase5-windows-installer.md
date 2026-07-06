---
level: L5
status: draft
owners: [maintainer]
updated: 2026-07-06
related-code: [installer/]
---

# Phase 5-A: Windows installer (v1.x — sau v1.0 Linux)

Lộ trình phân phối #2 (plan 7.5). Đóng gói OpenITMS-SMB chạy native trên Windows.

## Việc
- [ ] `installer/windows/install.ps1` — giải nén, cấu hình, đăng ký **Windows Service**
      (NSSM hoặc `sc create`) cho core + MariaDB. Idempotent như install.sh Linux.
- [ ] MariaDB **Windows build** bundled (thay vì Linux tarball). Socket → named pipe/localhost TCP.
- [ ] pwsh đã là native Windows (đơn giản hơn Linux).
- [ ] Thư mục `.\certs` + config lần đầu + admin default (như Linux).
- [ ] E2E: VM Windows Server/11 sạch → install.ps1 → login UI → chạy winrs-cert xuống Win11 khác.

## Lưu ý
- Binary Go build sẵn `GOOS=windows` (đã chạy — dev machine Windows build được).
- Không Docker (luật #5). Service quản lý bằng SCM.
