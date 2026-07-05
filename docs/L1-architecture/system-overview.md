---
level: L1
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [plugin-manager/, core-patches/, registry/, installer/]
---

# Kiến trúc tổng thể

Sơ đồ + mô tả đầy đủ: `PLAN.md` mục 3. Tóm tắt các khối:

- **Core (fork Semaphore)** — binary native; = `upstream/` @ baseline + `core-patches/` apply lúc build.
- **Plugin Manager** — package Go riêng (`plugin-manager/`), được hook vào core bằng patch 0001
  (vài dòng). Quét `/plugins`, launch qua HashiCorp go-plugin (gRPC, mTLS), API động từ manifest,
  health-check + restart.
- **Plugins** — process riêng, Go hoặc Python, hợp đồng duy nhất `proto/quickwin/plugin/v1/plugin.proto`.
- **MariaDB bundled** — binary độc lập, socket-only (ADR-0002).
- **Registry** — static index ký số, `local` + `public` (ADR-0003).
- **Thư mục `./certs`** — watcher nạp nóng `.pfx/.pem` cho WinRS certificate auth.
- **Kênh thực thi xuống client:** pwsh → SSH → WinRS(cert).

## Nguyên tắc bất biến
1. Core = upstream + patch mỏng (mọi diff phải là bộ-4).
2. Tính năng mới = plugin; ngoại lệ duy nhất được duyệt: Endpoint Script Manager + Registry client.
3. Một giao thức core↔plugin: gRPC/Protobuf qua go-plugin.
4. Không Docker trong sản phẩm/bộ cài.
