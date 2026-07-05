---
level: L0
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: []
---

# Thuật ngữ

| Thuật ngữ | Nghĩa trong dự án này |
|---|---|
| **upstream** | Repo Semaphore UI gốc, nằm ở submodule `upstream/`, pin 1 tag, CẤM sửa tay |
| **core-patch** | 1 file diff trong `core-patches/` thay đổi upstream; apply bằng script; phải có đủ **bộ-4** |
| **bộ-4** | patch + dòng trong `series` + entry `CHANGELOG.md` + spec `docs/L2-specs/core-patches/` |
| **series** | File liệt kê thứ tự apply patch (`core-patches/series`) |
| **Plugin Manager** | Module Go (ngoài cây upstream) quét `/plugins`, chạy plugin qua go-plugin, sinh API động |
| **API động** | Endpoint `/api/plugins/<name>/<route>` sinh tự động từ manifest plugin |
| **manifest** | `plugin.yaml` của plugin: name, version, license, routes, permissions, checksum |
| **registry** | Kho phân phối artifact (plugin / template / endpoint-script) dạng static index có ký số; có `local` và `public` |
| **template** | Kịch bản playbook chạy 1-click từ UI (JEA/WinRS, Odoo, MariaDB, ClickHouse…) |
| **endpoint-script** | Script chuẩn bị máy client (bật WinRS, cấu hình SSH, enroll cert…) |
| **baseline** | Tag upstream mà toàn bộ patch được viết trên đó (hiện: v2.18.16) |
| **sync-upstream** | Quy trình nâng baseline lên tag upstream mới (script + AI rebase patch) |
| **dist** | Thư mục output đóng gói cuối (`dist/`), toàn binary native, không Docker |
