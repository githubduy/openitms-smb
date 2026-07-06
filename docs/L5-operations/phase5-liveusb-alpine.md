---
level: L5
status: draft
owners: [maintainer]
updated: 2026-07-06
related-code: [installer/]
---

# Phase 5-B: Live USB Alpine minimal (v2.x)

Lộ trình #3 (ADR-0004 #5 chốt Alpine). Boot USB là có OpenITMS-SMB chạy sẵn — cứu hộ/triển khai tại chỗ.

## Việc
- [ ] Package list **tường minh trong git** — mỗi gói thêm phải có lý do + review (yêu cầu user).
- [ ] Build image Alpine (mkimage/apkovl) preseed: core + MariaDB + pwsh + plugins + auto-start.
- [ ] **Rủi ro musl (verify sớm):** Alpine dùng musl libc.
      - Core Go: build static (CGO_ENABLED=0) → OK musl.
      - MariaDB: dùng gói apk `mariadb` của Alpine (musl-native) thay Linux glibc tarball.
      - **pwsh: PowerShell hỗ trợ musl HẠN CHẾ** → kiểm chứng sớm; fallback: .NET self-contained
        hoặc đưa pwsh thành optional trên bản Live USB (kênh SSH vẫn đủ cho Linux target).
- [ ] E2E: boot image trong QEMU → UI lên → login.

## Chốt kỹ thuật cần làm đầu Phase 5-B
Test `pwsh` trên Alpine (musl) NGAY để biết trước rủi ro (đã ghi nhận từ 2026-07-05).
